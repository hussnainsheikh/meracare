package notifications

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// horizon is how far ahead a reminder plan reaches.
	//
	// Long enough that somebody who does not open the app for a few days is
	// still reminded — the reminders are scheduled on the device, so nothing
	// fires unless it was planned before the app was last closed. Short enough
	// that the plan reflects reasonably current authorization: a revoked
	// caregiver's already-scheduled reminders survive until their app next
	// reconciles, and this bounds how long that can matter
	// (plans/phase8.md §23).
	horizon = 7 * 24 * time.Hour

	// maxReminders is how many reminders one plan may contain.
	//
	// Not an arbitrary round number: iOS keeps at most 64 pending local
	// notifications per app and silently discards the rest, so a plan larger
	// than that would be partly imaginary. Fifty leaves headroom for anything
	// the OS itself schedules and keeps the plan comfortably inside the limit
	// on both platforms (plans/phase8.md §46).
	maxReminders = 50

	// maxDeviceIDLength and maxAppVersionLength bound free text so a client
	// cannot store arbitrarily large strings against a user.
	maxDeviceIDLength   = 128
	maxAppVersionLength = 64
)

// Service answers the two questions this package exists for: what may we
// remind this user about, and where would we send a push if we sent one.
type Service struct {
	preferences *Repository
	circle      CircleSource
	sources     map[ReminderType]ScheduleSource
}

// NewService wires the service to its sources.
//
// The schedule sources are supplied rather than constructed, and are keyed by
// reminder type so a domain that has no reminder path simply has no entry
// instead of a nil check at every use.
func NewService(
	repository *Repository,
	circle CircleSource,
	tasks, medications, appointments ScheduleSource,
) *Service {
	return &Service{
		preferences: repository,
		circle:      circle,
		sources: map[ReminderType]ScheduleSource{
			ReminderTaskReminder:        tasks,
			ReminderMedicationReminder:  medications,
			ReminderAppointmentReminder: appointments,
		},
	}
}

// Preferences returns one user's settings.
func (s *Service) Preferences(ctx context.Context, userID uuid.UUID) (Preferences, error) {
	return s.preferences.GetPreferences(ctx, userID)
}

// UpdatePreferences applies a change to one user's settings.
//
// The user is always the authenticated caller. There is no path by which one
// user's identifier reaches this from a request body, which is what makes
// "users cannot modify another user's preferences" a property of the shape of
// the code rather than a check that could be forgotten (plans/phase8.md §40).
func (s *Service) UpdatePreferences(
	ctx context.Context,
	userID uuid.UUID,
	update PreferenceUpdate,
) (Preferences, error) {
	return s.preferences.SavePreferences(ctx, userID, update)
}

// ErrInvalidDevice is returned when a registration cannot be understood. The
// handler turns it into a 422 with the offending field.
type ErrInvalidDevice struct {
	Field   string
	Message string
}

func (e ErrInvalidDevice) Error() string {
	return fmt.Sprintf("notifications: %s %s", e.Field, e.Message)
}

// RegisterDevice records the caller's installation.
func (s *Service) RegisterDevice(ctx context.Context, params RegisterParams) (Device, error) {
	params.DeviceID = strings.TrimSpace(params.DeviceID)
	params.AppVersion = strings.TrimSpace(params.AppVersion)
	params.PushToken = strings.TrimSpace(params.PushToken)

	switch {
	case params.DeviceID == "":
		return Device{}, ErrInvalidDevice{Field: "deviceId", Message: "is required"}
	case len(params.DeviceID) > maxDeviceIDLength:
		return Device{}, ErrInvalidDevice{Field: "deviceId", Message: "is too long"}
	case !params.Platform.Valid():
		return Device{}, ErrInvalidDevice{Field: "platform", Message: "is not a platform we recognise"}
	case len(params.AppVersion) > maxAppVersionLength:
		return Device{}, ErrInvalidDevice{Field: "appVersion", Message: "is too long"}
	}

	return s.preferences.Register(ctx, params)
}

// DeactivateDevice stops MeraCare reaching one of the caller's installations.
func (s *Service) DeactivateDevice(ctx context.Context, userID uuid.UUID, deviceID string) error {
	return s.preferences.Deactivate(ctx, userID, deviceID)
}

// Plan returns the reminders this user's device should schedule.
//
// The whole plan, not a delta: the device reconciles what it has against what
// this returns, and a full list is the only version of that which is correct
// after a reinstall, a restore, or a month offline.
func (s *Service) Plan(ctx context.Context, userID uuid.UUID, now time.Time) ([]Reminder, error) {
	preferences, err := s.preferences.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	memberships, err := s.circle.Memberships(ctx, userID)
	if err != nil {
		return nil, err
	}

	return planFor(ctx, s.sources, memberships, userID, preferences, now, now.Add(horizon))
}
