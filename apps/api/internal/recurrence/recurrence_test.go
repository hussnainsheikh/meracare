package recurrence

import (
	"testing"
	"time"
)

func TestParseReadsTheRulesTheMVPSupports(t *testing.T) {
	cases := []struct {
		rule string
		want string
	}{
		{"FREQ=DAILY", "FREQ=DAILY"},
		{"FREQ=WEEKLY;BYDAY=MO", "FREQ=WEEKLY;BYDAY=MO"},
		{"FREQ=WEEKLY;BYDAY=MO,WE,FR", "FREQ=WEEKLY;BYDAY=MO,WE,FR"},
		{"FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR", "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR"},
		// Order and case are normalised, so an equivalent rule always stores
		// identically and two rules can be compared as strings.
		{"FREQ=WEEKLY;BYDAY=FR,MO,we", "FREQ=WEEKLY;BYDAY=MO,WE,FR"},
		{"FREQ=WEEKLY;BYDAY=MO,MO", "FREQ=WEEKLY;BYDAY=MO"},
		{"freq=daily", "FREQ=DAILY"},
	}

	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			rule, err := Parse(tc.rule)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tc.rule, err)
			}
			if got := rule.String(); got != tc.want {
				t.Errorf("round trip = %q, want %q", got, tc.want)
			}
		})
	}
}

// A rule that parsed loosely would schedule care at the wrong times, which is
// worse than refusing to save it.
func TestParseRefusesWhatItCannotHonour(t *testing.T) {
	cases := []string{
		"",
		"FREQ=MONTHLY",
		"FREQ=YEARLY",
		"BYDAY=MO",
		"FREQ=WEEKLY",
		"FREQ=WEEKLY;BYDAY=",
		"FREQ=WEEKLY;BYDAY=XX",
		"FREQ=DAILY;BYDAY=MO",
		"FREQ=DAILY;INTERVAL=2",
		"FREQ=DAILY;COUNT=10",
		"nonsense",
	}

	for _, rule := range cases {
		t.Run(rule, func(t *testing.T) {
			if _, err := Parse(rule); err == nil {
				t.Errorf("Parse(%q) succeeded; want an error", rule)
			}
		})
	}
}

func TestDailyOccurrencesLandAtTheDueTimeEveryDay(t *testing.T) {
	london := mustLoad(t, "Europe/London")
	due := TimeOfDay{Hour: 9}

	from := time.Date(2026, time.August, 17, 0, 0, 0, 0, london)
	to := from.AddDate(0, 0, 4)

	got := Daily().Occurrences(due, london, from, to)

	want := []time.Time{
		time.Date(2026, time.August, 17, 9, 0, 0, 0, london),
		time.Date(2026, time.August, 18, 9, 0, 0, 0, london),
		time.Date(2026, time.August, 19, 9, 0, 0, 0, london),
		time.Date(2026, time.August, 20, 9, 0, 0, 0, london),
	}
	assertOccurrences(t, got, want)
}

func TestWeeklyOccurrencesLandOnlyOnTheChosenDays(t *testing.T) {
	karachi := mustLoad(t, "Asia/Karachi")
	due := TimeOfDay{Hour: 8, Minute: 30}

	// A fortnight from a Monday.
	from := time.Date(2026, time.August, 17, 0, 0, 0, 0, karachi)
	to := from.AddDate(0, 0, 14)

	got := Weekly(time.Monday, time.Wednesday, time.Friday).Occurrences(due, karachi, from, to)

	want := []time.Time{
		time.Date(2026, time.August, 17, 8, 30, 0, 0, karachi), // Mon
		time.Date(2026, time.August, 19, 8, 30, 0, 0, karachi), // Wed
		time.Date(2026, time.August, 21, 8, 30, 0, 0, karachi), // Fri
		time.Date(2026, time.August, 24, 8, 30, 0, 0, karachi), // Mon
		time.Date(2026, time.August, 26, 8, 30, 0, 0, karachi), // Wed
		time.Date(2026, time.August, 28, 8, 30, 0, 0, karachi), // Fri
	}
	assertOccurrences(t, got, want)
}

func TestWeekdayRecurrenceSkipsTheWeekend(t *testing.T) {
	london := mustLoad(t, "Europe/London")
	weekdays := Weekly(time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday)

	// Friday through Monday.
	from := time.Date(2026, time.August, 21, 0, 0, 0, 0, london)
	to := time.Date(2026, time.August, 25, 0, 0, 0, 0, london)

	got := weekdays.Occurrences(TimeOfDay{Hour: 9}, london, from, to)

	want := []time.Time{
		time.Date(2026, time.August, 21, 9, 0, 0, 0, london), // Fri
		time.Date(2026, time.August, 24, 9, 0, 0, 0, london), // Mon
	}
	assertOccurrences(t, got, want)
}

// The heart of §15. A task due at 09:00 must stay at 09:00 local through a
// daylight-saving change. Adding 24 hours repeatedly would drift it to 08:00 or
// 10:00 and quietly reschedule somebody's care.
func TestDailyOccurrencesHoldTheirLocalTimeAcrossDaylightSaving(t *testing.T) {
	london := mustLoad(t, "Europe/London")
	due := TimeOfDay{Hour: 9}

	// British Summer Time ends on Sunday 25 October 2026: clocks go back an
	// hour, so that day has 25 hours.
	from := time.Date(2026, time.October, 23, 0, 0, 0, 0, london)
	to := time.Date(2026, time.October, 28, 0, 0, 0, 0, london)

	got := Daily().Occurrences(due, london, from, to)

	if len(got) != 5 {
		t.Fatalf("got %d occurrences, want 5", len(got))
	}
	for _, at := range got {
		local := at.In(london)
		if local.Hour() != 9 || local.Minute() != 0 {
			t.Errorf("occurrence %s is at %02d:%02d local, want 09:00",
				at.Format(time.RFC3339), local.Hour(), local.Minute())
		}
	}

	// And the offset really did change underneath them, so the test would
	// notice a naive 24-hour walk.
	if got[0].In(london).Format("-0700") == got[4].In(london).Format("-0700") {
		t.Fatal("expected the UTC offset to change across the transition")
	}
}

// Spring forward: 01:00 to 02:00 does not exist on 29 March 2026 in London. An
// occurrence must still be produced rather than silently lost.
func TestOccurrenceSurvivesATimeThatDoesNotExist(t *testing.T) {
	london := mustLoad(t, "Europe/London")

	from := time.Date(2026, time.March, 29, 0, 0, 0, 0, london)
	to := time.Date(2026, time.March, 30, 0, 0, 0, 0, london)

	got := Daily().Occurrences(TimeOfDay{Hour: 1, Minute: 30}, london, from, to)

	if len(got) != 1 {
		t.Fatalf("got %d occurrences, want 1 — the task must not vanish", len(got))
	}
}

// A senior in Karachi and one in London have different days, so "today" is a
// different span of instants for each.
func TestOccurrencesAreComputedInTheSeniorsOwnTimezone(t *testing.T) {
	karachi := mustLoad(t, "Asia/Karachi")
	london := mustLoad(t, "Europe/London")
	due := TimeOfDay{Hour: 9}

	from := time.Date(2026, time.August, 17, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)

	inKarachi := Daily().Occurrences(due, karachi, from, to)
	inLondon := Daily().Occurrences(due, london, from, to)

	if len(inKarachi) != 1 || len(inLondon) != 1 {
		t.Fatalf("got %d and %d occurrences, want 1 each", len(inKarachi), len(inLondon))
	}
	if inKarachi[0].Equal(inLondon[0]) {
		t.Error("09:00 in Karachi and 09:00 in London are not the same instant")
	}
}

func TestOccurrencesExcludeTheEndOfTheWindow(t *testing.T) {
	london := mustLoad(t, "Europe/London")
	due := TimeOfDay{Hour: 9}

	from := time.Date(2026, time.August, 17, 9, 0, 0, 0, london)
	to := time.Date(2026, time.August, 18, 9, 0, 0, 0, london)

	got := Daily().Occurrences(due, london, from, to)

	// Inclusive at the start, exclusive at the end: adjacent windows tile
	// without producing the same occurrence twice.
	assertOccurrences(t, got, []time.Time{from})
}

func TestOccurrencesRefuseAnEmptyOrBackwardsWindow(t *testing.T) {
	london := mustLoad(t, "Europe/London")
	at := time.Date(2026, time.August, 17, 0, 0, 0, 0, london)

	if got := Daily().Occurrences(TimeOfDay{Hour: 9}, london, at, at); len(got) != 0 {
		t.Errorf("empty window produced %d occurrences", len(got))
	}
	if got := Daily().Occurrences(TimeOfDay{Hour: 9}, london, at, at.AddDate(0, 0, -1)); len(got) != 0 {
		t.Errorf("backwards window produced %d occurrences", len(got))
	}
	if got := Daily().Occurrences(TimeOfDay{Hour: 9}, nil, at, at.AddDate(0, 0, 1)); len(got) != 0 {
		t.Errorf("nil location produced %d occurrences", len(got))
	}
}

// A window wider than the cap must not expand without bound.
func TestOccurrencesAreCapped(t *testing.T) {
	london := mustLoad(t, "Europe/London")
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, london)

	got := Daily().Occurrences(TimeOfDay{Hour: 9}, london, from, from.AddDate(10, 0, 0))

	if len(got) > maxOccurrences {
		t.Errorf("got %d occurrences, want at most %d", len(got), maxOccurrences)
	}
}

func TestParseTimeOfDay(t *testing.T) {
	cases := []struct {
		value string
		want  string
	}{
		{"09:00", "09:00"},
		{"9:00", "09:00"},
		{"23:59", "23:59"},
		{"00:00", "00:00"},
		// PostgreSQL renders a time column with seconds.
		{"08:30:00", "08:30"},
	}

	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			at, err := ParseTimeOfDay(tc.value)
			if err != nil {
				t.Fatalf("ParseTimeOfDay(%q): %v", tc.value, err)
			}
			if got := at.String(); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseTimeOfDayRejectsImpossibleTimes(t *testing.T) {
	for _, value := range []string{"", "24:00", "09:60", "-1:00", "nine", "09", "09:aa"} {
		t.Run(value, func(t *testing.T) {
			if _, err := ParseTimeOfDay(value); err == nil {
				t.Errorf("ParseTimeOfDay(%q) succeeded; want an error", value)
			}
		})
	}
}

func TestParseRequestReadsWhatTheClientSends(t *testing.T) {
	daily, err := ParseRequest("daily", nil)
	if err != nil {
		t.Fatalf("daily: %v", err)
	}
	if daily.String() != "FREQ=DAILY" {
		t.Errorf("daily = %q", daily.String())
	}

	weekly, err := ParseRequest("weekly", []string{"monday", "friday"})
	if err != nil {
		t.Fatalf("weekly: %v", err)
	}
	if weekly.String() != "FREQ=WEEKLY;BYDAY=MO,FR" {
		t.Errorf("weekly = %q", weekly.String())
	}

	if _, err := ParseRequest("weekly", nil); err == nil {
		t.Error("a weekly rule with no days should be refused")
	}
	if _, err := ParseRequest("weekly", []string{"someday"}); err == nil {
		t.Error("an unrecognised day should be refused")
	}
	if _, err := ParseRequest("hourly", nil); err == nil {
		t.Error("an unsupported frequency should be refused")
	}
}

func mustLoad(t *testing.T, name string) *time.Location {
	t.Helper()

	location, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	return location
}

func assertOccurrences(t *testing.T, got, want []time.Time) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d occurrences, want %d:\n got: %v\nwant: %v",
			len(got), len(want), format(got), format(want))
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("occurrence %d = %s, want %s",
				i, got[i].Format(time.RFC3339), want[i].Format(time.RFC3339))
		}
	}
}

func format(times []time.Time) []string {
	rendered := make([]string, len(times))
	for i, at := range times {
		rendered[i] = at.Format(time.RFC3339)
	}
	return rendered
}
