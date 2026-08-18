package careevents

import (
	"context"

	"github.com/google/uuid"
)

// Service reads the activity timeline.
//
// Reading only: events are written by the domains that cause them, through a
// Recorder, never through this. There is no Create here because there is no
// endpoint that creates one (plans/phase7.md §21).
type Service struct {
	events *Repository
}

// NewService builds the service.
func NewService(events *Repository) *Service {
	return &Service{events: events}
}

// defaultPage and maxPage bound one page of activity, so a client cannot ask
// for a whole timeline in one request. docs/11 puts the expected working set at
// 20–50 recent events, which is what the default is sized for
// (plans/phase7.md §30).
const (
	defaultPage = 30
	maxPage     = 100
)

// Activity returns one page of a senior's timeline, newest first.
func (s *Service) Activity(
	ctx context.Context,
	seniorID uuid.UUID,
	cursor string,
	limit int,
) (Page, error) {
	if limit <= 0 || limit > maxPage {
		limit = defaultPage
	}
	return s.events.List(ctx, seniorID, cursor, limit)
}
