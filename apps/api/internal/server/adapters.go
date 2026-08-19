package server

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/meracare/api/internal/appointments"
	"github.com/meracare/api/internal/invitations"
	"github.com/meracare/api/internal/medications"
	"github.com/meracare/api/internal/notifications"
	"github.com/meracare/api/internal/seniors"
	"github.com/meracare/api/internal/tasks"
	"github.com/meracare/api/internal/users"
)

// userLookup adapts the users repository to the narrow interface the invitation
// flow needs.
//
// The adapter lives here, at the composition root, so internal/invitations
// depends on a two-method interface rather than on the whole users package.
type userLookup struct {
	repo *users.Repository
}

func (l userLookup) GetByID(ctx context.Context, id uuid.UUID) (invitations.UserSummary, error) {
	user, err := l.repo.GetByID(ctx, id)
	if err != nil {
		return invitations.UserSummary{}, err
	}
	return invitations.UserSummary{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}, nil
}

func (l userLookup) FindIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	return l.repo.FindIDByEmail(ctx, email)
}

// The four adapters below let internal/notifications ask "what is coming up?"
// without importing tasks, medications, appointments, or seniors.
//
// The alternative — a notifications package that knows about doses — is how a
// reminder system starts making care decisions. Keeping the translation here,
// at the composition root, means the notification code cannot reach a dosage
// even by accident, which is most of what plans/phase8.md §§1, 16, and 17 ask
// for (docs/05-api-and-backend-spec.md).

// circleSource adapts the seniors repository to the reminder plan's view of a
// care circle.
type circleSource struct {
	repo *seniors.Repository
}

// Memberships returns only the caller's active relationships.
//
// This is where a revoked caregiver stops being reminded: they have no active
// membership, so no senior of theirs contributes reminders, and the next plan
// their device fetches is empty of that senior (plans/phase8.md §23).
func (c circleSource) Memberships(
	ctx context.Context,
	userID uuid.UUID,
) ([]notifications.Membership, error) {
	found, err := c.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	memberships := make([]notifications.Membership, 0, len(found))
	for _, membership := range found {
		if !membership.Relationship.IsActive() {
			continue
		}

		memberships = append(memberships, notifications.Membership{
			Senior: notifications.Senior{
				ID:          membership.Senior.ID,
				DisplayName: membership.Senior.DisplayName,
				Timezone:    membership.Senior.Timezone,
			},
			Permissions: membership.Relationship.Permissions,
		})
	}
	return memberships, nil
}

// taskSource adapts the task service to the reminder plan.
type taskSource struct {
	service *tasks.Service
}

// Upcoming returns the senior's outstanding task occurrences in the window.
//
// Going through the service rather than the repository is deliberate: the
// service materialises occurrences for the window before reading them, so a
// recurring task that nobody has opened the app to look at still produces
// reminders. That also keeps the recurrence engine singular — the plan consumes
// the expansion tasks already does, and never expands a rule itself
// (plans/phase8.md §21).
func (t taskSource) Upcoming(
	ctx context.Context,
	seniorID uuid.UUID,
	from, to time.Time,
) ([]notifications.Due, error) {
	instances, err := t.service.List(ctx, tasks.ListInput{
		SeniorID: seniorID,
		Scope:    tasks.ScopeWindow,
		From:     from,
		To:       to,
	}, from)
	if err != nil {
		return nil, err
	}

	due := make([]notifications.Due, 0, len(instances))
	for _, instance := range instances {
		// Anything already dealt with needs no reminder. Overdue is derived
		// rather than stored, so a still-pending occurrence is exactly what is
		// outstanding (plans/phase8.md §22).
		if instance.Status != tasks.StatusPending {
			continue
		}

		due = append(due, notifications.Due{
			EntityID:   instance.ID,
			At:         instance.ScheduledFor,
			AssigneeID: instance.AssignedUserID,
		})
	}
	return due, nil
}

// medicationSource adapts the medication service to the reminder plan.
type medicationSource struct {
	service *medications.Service
}

// Upcoming returns the senior's outstanding doses in the window.
//
// Doses carry no assignee: a medicine is the senior's, and anybody permitted to
// record it may be the one who helps. So every circle member with
// medications.view is reminded, which is the behaviour a family sharing care
// actually needs (plans/phase8.md §4).
func (m medicationSource) Upcoming(
	ctx context.Context,
	seniorID uuid.UUID,
	from, to time.Time,
) ([]notifications.Due, error) {
	instances, err := m.service.ListDoses(ctx, medications.ListDosesInput{
		SeniorID: seniorID,
		Scope:    medications.ScopeWindow,
		From:     from,
		To:       to,
	}, from)
	if err != nil {
		return nil, err
	}

	due := make([]notifications.Due, 0, len(instances))
	for _, instance := range instances {
		if instance.Status != medications.StatusPending {
			continue
		}

		due = append(due, notifications.Due{EntityID: instance.ID, At: instance.ScheduledFor})
	}
	return due, nil
}

// appointmentSource adapts the appointment service to the reminder plan.
type appointmentSource struct {
	service *appointments.Service
}

// Upcoming returns the senior's still-scheduled appointments in the window.
//
// A cancelled appointment is filtered out here, and that is the whole of
// "cancelling an appointment cancels its reminder": the plan stops containing
// it, and the device's next reconciliation removes it. Nothing has to remember
// to clean anything up (plans/phase8.md §22).
func (a appointmentSource) Upcoming(
	ctx context.Context,
	seniorID uuid.UUID,
	from, to time.Time,
) ([]notifications.Due, error) {
	booked, err := a.service.Window(ctx, seniorID, from, to)
	if err != nil {
		return nil, err
	}

	due := make([]notifications.Due, 0, len(booked))
	for _, appointment := range booked {
		if appointment.Status != appointments.StatusScheduled {
			continue
		}

		due = append(due, notifications.Due{
			EntityID:   appointment.ID,
			At:         appointment.ScheduledAt,
			AssigneeID: appointment.AssignedUserID,
		})
	}
	return due, nil
}
