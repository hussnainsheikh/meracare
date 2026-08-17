package tasks

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOverdueIsDerivedFromTheClock(t *testing.T) {
	due := time.Date(2026, time.August, 17, 9, 0, 0, 0, time.UTC)
	before := due.Add(-time.Minute)
	after := due.Add(time.Minute)

	pending := Instance{ScheduledFor: due, Status: StatusPending}

	if got := pending.EffectiveStatus(before); got != StatusPending {
		t.Errorf("before its due time = %q, want %q", got, StatusPending)
	}
	if got := pending.EffectiveStatus(after); got != StatusOverdue {
		t.Errorf("after its due time = %q, want %q", got, StatusOverdue)
	}

	// Nothing was written to reach that conclusion: the stored status is
	// untouched, which is what lets the answer be right without a sweep.
	if pending.Status != StatusPending {
		t.Errorf("stored status changed to %q", pending.Status)
	}
}

// Once somebody has acted, the clock stops mattering. A task completed late is
// completed, not overdue.
func TestASettledTaskIsNeverOverdue(t *testing.T) {
	due := time.Date(2026, time.August, 17, 9, 0, 0, 0, time.UTC)
	longAfter := due.AddDate(0, 0, 30)

	for _, status := range []Status{StatusCompleted, StatusSkipped, StatusCancelled} {
		t.Run(string(status), func(t *testing.T) {
			instance := Instance{ScheduledFor: due, Status: status}

			if got := instance.EffectiveStatus(longAfter); got != status {
				t.Errorf("got %q, want %q", got, status)
			}
			if instance.Overdue(longAfter) {
				t.Error("a settled task reported itself overdue")
			}
		})
	}
}

func TestOverdueBeginsStrictlyAfterTheDueTime(t *testing.T) {
	due := time.Date(2026, time.August, 17, 9, 0, 0, 0, time.UTC)
	instance := Instance{ScheduledFor: due, Status: StatusPending}

	// Exactly at the due time the task is due, not late.
	if instance.Overdue(due) {
		t.Error("a task is not overdue at the moment it falls due")
	}
}

func TestPendingTaskAcceptsAnyAction(t *testing.T) {
	cases := map[Action]Status{
		ActionComplete: StatusCompleted,
		ActionSkip:     StatusSkipped,
		ActionCancel:   StatusCancelled,
	}

	for action, want := range cases {
		t.Run(string(action), func(t *testing.T) {
			next, repeat, err := Transition(StatusPending, action)
			if err != nil {
				t.Fatalf("Transition: %v", err)
			}
			if next != want {
				t.Errorf("next = %q, want %q", next, want)
			}
			if repeat {
				t.Error("a first action reported itself as a repeat")
			}
		})
	}
}

// §27: the mobile client retries queued mutations when it comes back online, so
// the same completion can arrive twice. The second must succeed quietly rather
// than showing the user an error for work that was recorded.
func TestRepeatingAnActionIsASuccessNotAConflict(t *testing.T) {
	cases := []struct {
		current Status
		action  Action
	}{
		{StatusCompleted, ActionComplete},
		{StatusSkipped, ActionSkip},
		{StatusCancelled, ActionCancel},
	}

	for _, tc := range cases {
		t.Run(string(tc.current), func(t *testing.T) {
			next, repeat, err := Transition(tc.current, tc.action)
			if err != nil {
				t.Fatalf("Transition: %v", err)
			}
			if !repeat {
				t.Error("a repeated action was not reported as a repeat")
			}
			if next != tc.current {
				t.Errorf("next = %q, want the state to be unchanged at %q", next, tc.current)
			}
		})
	}
}

// §28: two devices, two different actions. The server is authoritative, and the
// second action must not overwrite a record of care that was actually given.
func TestADifferentOutcomeCannotOverwriteOne(t *testing.T) {
	cases := []struct {
		current Status
		action  Action
	}{
		{StatusCompleted, ActionSkip},
		{StatusCompleted, ActionCancel},
		{StatusSkipped, ActionComplete},
		{StatusSkipped, ActionCancel},
		{StatusCancelled, ActionComplete},
		{StatusCancelled, ActionSkip},
	}

	for _, tc := range cases {
		t.Run(string(tc.current)+"/"+string(tc.action), func(t *testing.T) {
			if _, _, err := Transition(tc.current, tc.action); !errors.Is(err, ErrInvalidTransition) {
				t.Errorf("err = %v, want ErrInvalidTransition", err)
			}
		})
	}
}

func TestUnknownActionIsRefused(t *testing.T) {
	if _, _, err := Transition(StatusPending, Action("delete")); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("err = %v, want ErrInvalidTransition", err)
	}
}

// The derived status must never reach the database: the CHECK constraint on
// care_task_instances.status does not accept it.
func TestOverdueIsNotAStorableStatus(t *testing.T) {
	if StatusOverdue.ValidStored() {
		t.Error("overdue reported itself storable")
	}
	for _, status := range StoredStatuses {
		if !status.ValidStored() {
			t.Errorf("%q reported itself unstorable", status)
		}
	}
}

func TestRecurringDistinguishesAOneTimeTask(t *testing.T) {
	templateID := uuid.New()

	if (Instance{}).Recurring() {
		t.Error("a task with no template reported itself recurring")
	}
	if !(Instance{TemplateID: &templateID}).Recurring() {
		t.Error("a task from a template reported itself one-time")
	}
}

func TestAssignedTo(t *testing.T) {
	me := uuid.New()
	somebodyElse := uuid.New()

	if (Instance{}).AssignedTo(me) {
		t.Error("an unassigned task reported itself assigned")
	}
	if !(Instance{AssignedUserID: &me}).AssignedTo(me) {
		t.Error("a task assigned to me reported otherwise")
	}
	if (Instance{AssignedUserID: &somebodyElse}).AssignedTo(me) {
		t.Error("a task assigned to somebody else reported itself mine")
	}
}
