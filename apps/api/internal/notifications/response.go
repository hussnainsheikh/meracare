package notifications

import "time"

// PreferencesResponse is the JSON form of one user's settings.
type PreferencesResponse struct {
	TaskReminders        bool   `json:"taskReminders"`
	MedicationReminders  bool   `json:"medicationReminders"`
	AppointmentReminders bool   `json:"appointmentReminders"`
	UpdatedAt            string `json:"updatedAt"`
}

// ToPreferencesResponse converts preferences for the wire.
//
// The user id is not included. The only preferences any request can reach are
// the caller's own, so returning the id would add a field whose only possible
// value the client already knows.
func ToPreferencesResponse(preferences Preferences) PreferencesResponse {
	updatedAt := ""
	if !preferences.UpdatedAt.IsZero() {
		updatedAt = preferences.UpdatedAt.UTC().Format(time.RFC3339)
	}

	return PreferencesResponse{
		TaskReminders:        preferences.TaskReminders,
		MedicationReminders:  preferences.MedicationReminders,
		AppointmentReminders: preferences.AppointmentReminders,
		UpdatedAt:            updatedAt,
	}
}

// DeviceResponse is the JSON form of a registered installation.
//
// There is no push token field, and adding one would be a security regression
// rather than a feature: the token is a credential for making somebody's phone
// buzz, the client that registered it already has it, and no other client has
// any business reading it (plans/phase8.md §8).
type DeviceResponse struct {
	ID         string `json:"id"`
	DeviceID   string `json:"deviceId"`
	Platform   string `json:"platform"`
	AppVersion string `json:"appVersion"`
	Active     bool   `json:"active"`
	LastSeenAt string `json:"lastSeenAt"`
	// PushTokenRegistered says whether we hold a token for this device,
	// without saying what it is — enough for the settings screen to explain
	// that notifications are set up.
	PushTokenRegistered bool `json:"pushTokenRegistered"`
}

// ToDeviceResponse converts a device for the wire.
func ToDeviceResponse(device Device) DeviceResponse {
	return DeviceResponse{
		ID:                  device.ID.String(),
		DeviceID:            device.DeviceID,
		Platform:            string(device.Platform),
		AppVersion:          device.AppVersion,
		Active:              device.Active,
		LastSeenAt:          device.LastSeenAt.UTC().Format(time.RFC3339),
		PushTokenRegistered: device.PushToken != "",
	}
}

// ReminderResponse is one notification the device should schedule.
//
// There is no title and no body. The device composes both from this and the
// shared wording in packages/contracts, which is what keeps a medicine's name
// out of anything that can appear on a lock screen (plans/phase8.md §§17, 47).
type ReminderResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	SeniorID       string `json:"seniorId"`
	SeniorName     string `json:"seniorName"`
	SeniorTimezone string `json:"seniorTimezone"`

	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`

	DueAt  string `json:"dueAt"`
	FireAt string `json:"fireAt"`
}

// ToReminderResponse converts one reminder for the wire.
func ToReminderResponse(reminder Reminder) ReminderResponse {
	return ReminderResponse{
		ID:             reminder.ID.String(),
		Type:           string(reminder.Type),
		SeniorID:       reminder.SeniorID.String(),
		SeniorName:     reminder.SeniorName,
		SeniorTimezone: reminder.SeniorTimezone,
		EntityType:     string(reminder.EntityType),
		EntityID:       reminder.EntityID.String(),
		DueAt:          reminder.DueAt.UTC().Format(time.RFC3339),
		FireAt:         reminder.FireAt.UTC().Format(time.RFC3339),
	}
}

// PlanResponse is a complete reminder plan.
type PlanResponse struct {
	Reminders []ReminderResponse `json:"reminders"`
	// GeneratedAt and HorizonEndsAt let the device reason about what this plan
	// covers: everything scheduled beyond HorizonEndsAt is outside the plan and
	// must not be cancelled just because it is absent from it.
	GeneratedAt   string `json:"generatedAt"`
	HorizonEndsAt string `json:"horizonEndsAt"`
}

// ToPlanResponse converts a plan for the wire.
func ToPlanResponse(reminders []Reminder, now time.Time) PlanResponse {
	responses := make([]ReminderResponse, 0, len(reminders))
	for _, reminder := range reminders {
		responses = append(responses, ToReminderResponse(reminder))
	}

	return PlanResponse{
		Reminders:     responses,
		GeneratedAt:   now.UTC().Format(time.RFC3339),
		HorizonEndsAt: now.Add(horizon).UTC().Format(time.RFC3339),
	}
}
