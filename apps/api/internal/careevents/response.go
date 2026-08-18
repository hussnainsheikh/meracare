package careevents

import (
	"time"

	"github.com/google/uuid"
)

// Response is the JSON representation of one care event.
//
// It carries what happened, who did it, when, and the small amount of detail a
// sentence needs — and no rendered sentence. The wording is the client's job,
// so the same event reads correctly in whatever language and phrasing the app
// grows into, and so a change of wording is not a data migration
// (plans/phase7.md §§3, 15).
type Response struct {
	ID       string `json:"id"`
	SeniorID string `json:"seniorId"`

	// Type is the documented identifier. The client maps it to words; it must
	// never be shown as it is (plans/phase7.md §14).
	Type string `json:"type"`

	// ActorUserID is null for an event no person performed.
	ActorUserID *string `json:"actorUserId"`

	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`

	Metadata Metadata `json:"metadata"`

	OccurredAt string `json:"occurredAt"`
}

// ToResponse renders one event.
func ToResponse(event Event) Response {
	metadata := event.Metadata
	if metadata == nil {
		metadata = Metadata{}
	}

	return Response{
		ID:          event.ID.String(),
		SeniorID:    event.SeniorID.String(),
		Type:        string(event.Type),
		ActorUserID: optionalUUID(event.ActorUserID),
		EntityType:  string(event.EntityType),
		EntityID:    event.EntityID.String(),
		Metadata:    metadata,
		OccurredAt:  event.OccurredAt.UTC().Format(time.RFC3339Nano),
	}
}

func responses(found []Event) []Response {
	items := make([]Response, 0, len(found))
	for _, event := range found {
		items = append(items, ToResponse(event))
	}
	return items
}

func optionalUUID(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	rendered := value.String()
	return &rendered
}
