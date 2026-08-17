package appointments

import (
	"time"

	"github.com/google/uuid"
)

// Response is the JSON representation of an appointment.
type Response struct {
	ID       string `json:"id"`
	SeniorID string `json:"seniorId"`

	Title string `json:"title"`
	// Kind is null when nobody said what sort of visit it is.
	Kind         *string `json:"kind"`
	ProviderName *string `json:"providerName"`
	Location     *string `json:"location"`
	Notes        *string `json:"notes"`

	// AssignedUserID is null when nobody in particular is taking them.
	AssignedUserID *string `json:"assignedUserId"`

	// ScheduledAt and EndsAt are instants. The client renders them in the
	// senior's timezone, which the senior resource carries.
	ScheduledAt string  `json:"scheduledAt"`
	EndsAt      *string `json:"endsAt"`

	Status string `json:"status"`

	CompletedAt *string `json:"completedAt"`
	CompletedBy *string `json:"completedBy"`
	CancelledAt *string `json:"cancelledAt"`
	CancelledBy *string `json:"cancelledBy"`

	// CreatedBy is carried so a later phase can say who booked it, and so the
	// detail screen can attribute the entry without a second request.
	CreatedBy string `json:"createdBy"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ToResponse renders one appointment.
func ToResponse(appointment Appointment) Response {
	return Response{
		ID:             appointment.ID.String(),
		SeniorID:       appointment.SeniorID.String(),
		Title:          appointment.Title,
		Kind:           optionalString(string(appointment.Kind)),
		ProviderName:   optionalString(appointment.ProviderName),
		Location:       optionalString(appointment.Location),
		Notes:          optionalString(appointment.Notes),
		AssignedUserID: optionalUUID(appointment.AssignedUserID),
		ScheduledAt:    appointment.ScheduledAt.UTC().Format(time.RFC3339),
		EndsAt:         optionalTime(appointment.EndsAt),
		Status:         string(appointment.Status),
		CompletedAt:    optionalTime(appointment.CompletedAt),
		CompletedBy:    optionalUUID(appointment.CompletedBy),
		CancelledAt:    optionalTime(appointment.CancelledAt),
		CancelledBy:    optionalUUID(appointment.CancelledBy),
		CreatedBy:      appointment.CreatedByUserID.String(),
		CreatedAt:      appointment.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      appointment.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func responses(found []Appointment) []Response {
	items := make([]Response, 0, len(found))
	for _, appointment := range found {
		items = append(items, ToResponse(appointment))
	}
	return items
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func optionalUUID(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	rendered := value.String()
	return &rendered
}

func optionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	rendered := value.UTC().Format(time.RFC3339)
	return &rendered
}
