package medications

import (
	"errors"
	"testing"
	"time"
)

// The state machine is the whole of the conflict policy: a replayed action
// succeeds, a contradictory one is refused, and the server's version stands
// (plans/phase5.md §§21–22).
func TestTransitionFromPendingRecordsTheAction(t *testing.T) {
	cases := []struct {
		action Action
		want   Status
	}{
		{ActionTake, StatusTaken},
		{ActionSkip, StatusSkipped},
	}

	for _, tc := range cases {
		t.Run(string(tc.action), func(t *testing.T) {
			next, repeat, err := Transition(StatusPending, tc.action)
			if err != nil {
				t.Fatalf("Transition: %v", err)
			}
			if next != tc.want {
				t.Errorf("next = %q, want %q", next, tc.want)
			}
			if repeat {
				t.Error("a first action is not a repeat")
			}
		})
	}
}

// The offline queue replays what it holds. A retry of an action that already
// landed must succeed quietly, or somebody sees an error for medicine they took.
func TestRepeatingAnActionIsNotAnError(t *testing.T) {
	cases := []struct {
		current Status
		action  Action
	}{
		{StatusTaken, ActionTake},
		{StatusSkipped, ActionSkip},
	}

	for _, tc := range cases {
		t.Run(string(tc.current), func(t *testing.T) {
			next, repeat, err := Transition(tc.current, tc.action)
			if err != nil {
				t.Fatalf("Transition: %v", err)
			}
			if !repeat {
				t.Error("repeat = false, want true")
			}
			if next != tc.current {
				t.Errorf("next = %q, want it unchanged at %q", next, tc.current)
			}
		})
	}
}

// Skipping a dose somebody has already taken would overwrite a record of
// medicine that was actually swallowed.
func TestContradictingASettledDoseIsRefused(t *testing.T) {
	cases := []struct {
		current Status
		action  Action
	}{
		{StatusTaken, ActionSkip},
		{StatusSkipped, ActionTake},
	}

	for _, tc := range cases {
		t.Run(string(tc.current)+"/"+string(tc.action), func(t *testing.T) {
			if _, _, err := Transition(tc.current, tc.action); !errors.Is(err, ErrInvalidTransition) {
				t.Errorf("err = %v, want ErrInvalidTransition", err)
			}
		})
	}
}

func TestTransitionRefusesAnUnknownAction(t *testing.T) {
	if _, _, err := Transition(StatusPending, Action("swallow")); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("err = %v, want ErrInvalidTransition", err)
	}
}

// Missed is derived, not stored: it is a reading of the clock against the dose,
// so nothing has to run overnight for it to be true (plans/phase5.md §8).
func TestADoseIsMissedOnlyOnceItsWindowHasPassed(t *testing.T) {
	due := time.Date(2026, time.August, 17, 8, 0, 0, 0, time.UTC)
	dose := Instance{ScheduledFor: due, Status: StatusPending}

	cases := []struct {
		name string
		now  time.Time
		want Status
	}{
		{"before it is due", due.Add(-time.Hour), StatusPending},
		{"exactly when due", due, StatusPending},
		// Somebody making breakfast is not late.
		{"a few minutes late", due.Add(20 * time.Minute), StatusPending},
		{"just inside the window", due.Add(MissedAfter - time.Minute), StatusPending},
		{"just outside the window", due.Add(MissedAfter + time.Minute), StatusMissed},
		{"the next day", due.AddDate(0, 0, 1), StatusMissed},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := dose.EffectiveStatus(tc.now); got != tc.want {
				t.Errorf("status = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestASettledDoseIsNeverMissed(t *testing.T) {
	due := time.Date(2026, time.August, 17, 8, 0, 0, 0, time.UTC)
	longAfter := due.AddDate(0, 0, 30)

	for _, status := range []Status{StatusTaken, StatusSkipped} {
		t.Run(string(status), func(t *testing.T) {
			dose := Instance{ScheduledFor: due, Status: status}

			if dose.Missed(longAfter) {
				t.Error("a dose somebody acted on cannot become missed")
			}
			if got := dose.EffectiveStatus(longAfter); got != status {
				t.Errorf("status = %q, want %q", got, status)
			}
		})
	}
}

// A missed dose is still pending underneath, so it can be taken late — which is
// what people do, and what the record should say.
func TestAMissedDoseCanStillBeTaken(t *testing.T) {
	next, repeat, err := Transition(StatusPending, ActionTake)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if next != StatusTaken || repeat {
		t.Errorf("next = %q, repeat = %v; want taken, false", next, repeat)
	}
}

func TestDueReportsAWindowThatIsOpen(t *testing.T) {
	due := time.Date(2026, time.August, 17, 8, 0, 0, 0, time.UTC)
	dose := Instance{ScheduledFor: due, Status: StatusPending}

	if dose.Due(due.Add(-time.Minute)) {
		t.Error("a dose is not due before its time")
	}
	if !dose.Due(due.Add(time.Minute)) {
		t.Error("a dose is due inside its window")
	}
	if dose.Due(due.Add(MissedAfter + time.Minute)) {
		t.Error("a missed dose is no longer merely due")
	}
}

// StatusMissed must never reach the database: the CHECK constraint on
// medication_instances.status does not recognise it.
func TestMissedIsNotAStoredStatus(t *testing.T) {
	if StatusMissed.ValidStored() {
		t.Error("missed is derived and must not be storable")
	}
	for _, status := range []Status{StatusPending, StatusTaken, StatusSkipped} {
		if !status.ValidStored() {
			t.Errorf("%q should be storable", status)
		}
	}
}

func TestSettledCoversTheOutcomesSomebodyRecorded(t *testing.T) {
	if StatusPending.Settled() {
		t.Error("pending is not settled")
	}
	if !StatusTaken.Settled() || !StatusSkipped.Settled() {
		t.Error("taken and skipped are settled")
	}
}

func TestFormsAreARecognisedList(t *testing.T) {
	if !FormTablet.Valid() || !FormOther.Valid() {
		t.Error("listed forms should be valid")
	}
	if Form("suppository").Valid() {
		t.Error("an unlisted form should be refused rather than stored")
	}
	if Form("").Valid() {
		t.Error("an empty form is absent, not a value")
	}
}
