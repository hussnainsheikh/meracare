// Package tasks implements the daily care routine: the templates that describe
// recurring care, and the concrete occurrences a care circle completes or skips
// (docs/03-domain-model.md, CareTaskTemplate and CareTaskInstance).
//
// Keep the vocabulary in sync with packages/contracts/src/tasks.ts.
package tasks

import (
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
)

// Status is the state of one task occurrence.
type Status string

const (
	// StatusPending is waiting to be done.
	StatusPending Status = "pending"
	// StatusCompleted was done, by somebody, at some time.
	StatusCompleted Status = "completed"
	// StatusSkipped was deliberately not done. It stays in the record: a
	// skipped task is care history, not an absence of one (plans/phase4.md §11).
	StatusSkipped Status = "skipped"
	// StatusCancelled is an occurrence that should no longer happen.
	StatusCancelled Status = "cancelled"

	// StatusOverdue is derived from the clock and never stored.
	//
	// A task does not become overdue because somebody opened a screen or
	// because a job ran; it is overdue precisely when its due time has passed
	// and nobody has acted on it. Deriving it means the answer is right the
	// first time it is asked, with no sweep to fall behind (plans/phase4.md §12).
	StatusOverdue Status = "overdue"
)

// StoredStatuses are the values the database accepts. The CHECK constraint on
// care_task_instances.status mirrors this list — StatusOverdue is absent from
// both.
var StoredStatuses = []Status{StatusPending, StatusCompleted, StatusSkipped, StatusCancelled}

// ValidStored reports whether the status is one that can be persisted.
func (s Status) ValidStored() bool { return slices.Contains(StoredStatuses, s) }

// Settled reports whether the task has reached a final state.
func (s Status) Settled() bool {
	return s == StatusCompleted || s == StatusSkipped || s == StatusCancelled
}

// Template is a recurring care task: the rule, not an occurrence of it.
type Template struct {
	ID              uuid.UUID
	SeniorID        uuid.UUID
	CreatedByUserID uuid.UUID

	Title       string
	Description string

	// AssignedUserID is nil when the task belongs to whoever is available.
	AssignedUserID *uuid.UUID

	Recurrence Recurrence
	DueTime    TimeOfDay

	Active bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Instance is one concrete occurrence of a care task.
type Instance struct {
	ID uuid.UUID
	// TemplateID is nil for a one-time task.
	TemplateID      *uuid.UUID
	SeniorID        uuid.UUID
	CreatedByUserID uuid.UUID

	Title       string
	Description string

	AssignedUserID *uuid.UUID

	ScheduledFor time.Time
	Status       Status

	CompletedAt *time.Time
	CompletedBy *uuid.UUID

	SkippedAt *time.Time
	SkippedBy *uuid.UUID

	Notes string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Recurring reports whether the occurrence came from a template.
func (i Instance) Recurring() bool { return i.TemplateID != nil }

// EffectiveStatus is the status as of now, including the derived overdue state.
//
// Only a pending task can be overdue: once somebody has completed, skipped, or
// cancelled it, the clock stops mattering.
func (i Instance) EffectiveStatus(now time.Time) Status {
	if i.Status == StatusPending && now.After(i.ScheduledFor) {
		return StatusOverdue
	}
	return i.Status
}

// Overdue reports whether the task is past due and still waiting.
func (i Instance) Overdue(now time.Time) bool {
	return i.EffectiveStatus(now) == StatusOverdue
}

// AssignedTo reports whether the task is this user's to do.
func (i Instance) AssignedTo(userID uuid.UUID) bool {
	return i.AssignedUserID != nil && *i.AssignedUserID == userID
}

// ErrInvalidTransition is returned when the requested action contradicts what
// has already happened to the task.
//
// It is deliberately distinct from "already in that state": completing a task
// twice is a retry and succeeds, whereas skipping a task somebody has already
// completed would overwrite a record of care that was actually given
// (plans/phase4.md §28).
var ErrInvalidTransition = errors.New("task is not in a state that allows this")

// Action is something a user does to a task occurrence.
type Action string

const (
	// ActionComplete marks the task done.
	ActionComplete Action = "complete"
	// ActionSkip marks the task deliberately not done.
	ActionSkip Action = "skip"
	// ActionCancel withdraws the occurrence.
	ActionCancel Action = "cancel"
)

// resultOf is the status each action produces.
var resultOf = map[Action]Status{
	ActionComplete: StatusCompleted,
	ActionSkip:     StatusSkipped,
	ActionCancel:   StatusCancelled,
}

// Transition decides what an action means for a task in its current state.
//
// It returns the resulting status, and whether the action is a repeat of one
// that already happened. A repeat is not an error: the mobile client retries
// queued mutations when it comes back online, and a retry that failed loudly
// would leave the user staring at an error for work that was recorded
// (plans/phase4.md §27).
func Transition(current Status, action Action) (next Status, repeat bool, err error) {
	target, ok := resultOf[action]
	if !ok {
		return "", false, ErrInvalidTransition
	}

	switch current {
	case StatusPending:
		return target, false, nil

	// The same action again: nothing changes, and the caller is told so it can
	// leave the original actor and timestamp alone.
	case target:
		return current, true, nil

	// Any other settled state: the task already has a different outcome, and
	// the server is authoritative over which one stands.
	default:
		return "", false, ErrInvalidTransition
	}
}
