// Package recurrence expands "every day at 09:00" into the instants it means.
//
// It began as part of internal/tasks in Phase 4. Medication schedules need the
// same rule grammar, the same timezone-correct expansion, and the same wire
// shape, and a second copy of a daylight-saving-aware calendar walk is exactly
// the kind of duplication that goes wrong quietly. So the engine lives here and
// both domains use it; neither owns it (plans/phase5.md §3).
package recurrence

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Errors returned when a client's recurrence cannot be understood. They are
// distinct values so the handler can point at the right field.
var (
	// ErrUnknownFrequency is a repeat setting other than daily or weekly.
	ErrUnknownFrequency = errors.New("recurrence: unknown frequency")
	// ErrUnknownWeekday is a day name that is not a day.
	ErrUnknownWeekday = errors.New("recurrence: unknown weekday")
	// ErrNoWeekdays is a weekly repeat that names no days.
	ErrNoWeekdays = errors.New("recurrence: no weekdays given")
)

// Rule is how often something repeats.
//
// The stored form is a deliberate subset of the iCalendar RRULE grammar:
//
//	FREQ=DAILY
//	FREQ=WEEKLY;BYDAY=MO,WE,FR
//
// A subset of a standard rather than an invention of our own, because the MVP
// needs exactly "every day", "every weekday", and "these days of the week"
// (plans/phase4.md §7, plans/phase5.md §3) — and when a later phase needs
// monthly or intervals, the grammar already has a place to put them.
//
// Nothing here is shown to a user. The mobile client turns a Rule into wording
// like "Every weekday"; the rule string never reaches a screen.
type Rule struct {
	// Frequency is DAILY or WEEKLY.
	Frequency Frequency
	// Weekdays are the days a weekly rule fires on, sorted, Sunday first.
	// Empty for a daily rule.
	Weekdays []time.Weekday
}

// Frequency is how a rule repeats.
type Frequency string

const (
	// FrequencyDaily fires every day.
	FrequencyDaily Frequency = "DAILY"
	// FrequencyWeekly fires on named days of the week.
	FrequencyWeekly Frequency = "WEEKLY"
)

// Daily returns the "every day" rule.
func Daily() Rule { return Rule{Frequency: FrequencyDaily} }

// Weekly returns a rule firing on the given days.
func Weekly(days ...time.Weekday) Rule {
	return Rule{Frequency: FrequencyWeekly, Weekdays: normaliseWeekdays(days)}
}

// weekdayCodes maps the two-letter RRULE codes to Go weekdays.
var weekdayCodes = map[string]time.Weekday{
	"SU": time.Sunday,
	"MO": time.Monday,
	"TU": time.Tuesday,
	"WE": time.Wednesday,
	"TH": time.Thursday,
	"FR": time.Friday,
	"SA": time.Saturday,
}

var weekdayNames = [...]string{"SU", "MO", "TU", "WE", "TH", "FR", "SA"}

// Parse reads a stored rule.
//
// It is strict: anything it does not recognise is an error rather than a
// silently ignored clause. A rule that parsed loosely would schedule care at
// the wrong times, which is worse than refusing to save it.
func Parse(rule string) (Rule, error) {
	parts := strings.Split(strings.TrimSpace(rule), ";")

	var (
		parsed   Rule
		sawFreq  bool
		sawByDay bool
	)

	for _, part := range parts {
		if part == "" {
			continue
		}

		name, value, ok := strings.Cut(part, "=")
		if !ok {
			return Rule{}, fmt.Errorf("recurrence: malformed clause %q", part)
		}

		switch strings.ToUpper(strings.TrimSpace(name)) {
		case "FREQ":
			frequency := Frequency(strings.ToUpper(strings.TrimSpace(value)))
			if frequency != FrequencyDaily && frequency != FrequencyWeekly {
				return Rule{}, fmt.Errorf("recurrence: unsupported frequency %q", value)
			}
			parsed.Frequency = frequency
			sawFreq = true

		case "BYDAY":
			days, err := parseWeekdays(value)
			if err != nil {
				return Rule{}, err
			}
			parsed.Weekdays = days
			sawByDay = true

		default:
			return Rule{}, fmt.Errorf("recurrence: unsupported clause %q", name)
		}
	}

	if !sawFreq {
		return Rule{}, fmt.Errorf("recurrence: missing FREQ")
	}

	switch parsed.Frequency {
	case FrequencyDaily:
		if sawByDay {
			return Rule{}, fmt.Errorf("recurrence: BYDAY is not meaningful with FREQ=DAILY")
		}
	case FrequencyWeekly:
		if len(parsed.Weekdays) == 0 {
			return Rule{}, fmt.Errorf("recurrence: FREQ=WEEKLY needs at least one BYDAY")
		}
	}

	return parsed, nil
}

func parseWeekdays(value string) ([]time.Weekday, error) {
	codes := strings.Split(value, ",")
	days := make([]time.Weekday, 0, len(codes))

	for _, code := range codes {
		weekday, ok := weekdayCodes[strings.ToUpper(strings.TrimSpace(code))]
		if !ok {
			return nil, fmt.Errorf("recurrence: unrecognised day %q", code)
		}
		days = append(days, weekday)
	}
	if len(days) == 0 {
		return nil, fmt.Errorf("recurrence: BYDAY is empty")
	}

	return normaliseWeekdays(days), nil
}

// normaliseWeekdays sorts and de-duplicates, so an equivalent rule always
// stores identically and two rules can be compared as strings.
func normaliseWeekdays(days []time.Weekday) []time.Weekday {
	cleaned := make([]time.Weekday, 0, len(days))
	for _, day := range days {
		if day < time.Sunday || day > time.Saturday {
			continue
		}
		if !slices.Contains(cleaned, day) {
			cleaned = append(cleaned, day)
		}
	}
	slices.Sort(cleaned)
	return cleaned
}

// String renders the rule for storage.
func (r Rule) String() string {
	if r.Frequency == FrequencyWeekly {
		codes := make([]string, len(r.Weekdays))
		for i, day := range r.Weekdays {
			codes[i] = weekdayNames[day]
		}
		return "FREQ=WEEKLY;BYDAY=" + strings.Join(codes, ",")
	}
	return "FREQ=" + string(r.Frequency)
}

// fires reports whether the rule produces an occurrence on the given weekday.
func (r Rule) fires(day time.Weekday) bool {
	if r.Frequency == FrequencyDaily {
		return true
	}
	return slices.Contains(r.Weekdays, day)
}

// maxOccurrences bounds one expansion, so a wide window cannot generate an
// unbounded number of rows. A daily rule over the longest window the API
// accepts stays well inside it.
const maxOccurrences = 512

// Occurrences returns the instants this rule fires on within [from, to).
//
// The walk is over local calendar dates in loc, not over instants: "every day
// at 09:00" means nine in the morning where the senior lives, on each of their
// days. Adding 24 hours repeatedly would drift by an hour across a
// daylight-saving boundary and silently reschedule their care
// (plans/phase4.md §15, plans/phase5.md §23).
//
// A local time that a daylight-saving jump skips entirely is normalised
// forward by the standard library, so an occurrence is never lost.
func (r Rule) Occurrences(due TimeOfDay, loc *time.Location, from, to time.Time) []time.Time {
	if loc == nil || !to.After(from) {
		return nil
	}

	// Start a day early and finish a day late: an occurrence's local date can
	// sit just outside the window's local dates while its instant falls inside,
	// around a timezone offset change. The instant comparison below is what
	// actually decides membership.
	year, month, day := from.In(loc).Date()
	end := to.In(loc).AddDate(0, 0, 1)

	occurrences := make([]time.Time, 0)
	cursor := time.Date(year, month, day-1, 0, 0, 0, 0, loc)

	for cursor.Before(end) && len(occurrences) < maxOccurrences {
		if r.fires(cursor.Weekday()) {
			y, m, d := cursor.Date()
			at := time.Date(y, m, d, due.Hour, due.Minute, 0, 0, loc)

			if !at.Before(from) && at.Before(to) {
				occurrences = append(occurrences, at)
			}
		}

		// Rebuilding from the calendar date, rather than adding 24 hours, is
		// what keeps the walk on real local days.
		y, m, d := cursor.Date()
		cursor = time.Date(y, m, d+1, 0, 0, 0, 0, loc)
	}

	return occurrences
}

// TimeOfDay is a wall-clock time in the senior's timezone.
//
// Deliberately not a time.Time: "09:00" is not an instant, and storing it as
// one would attach a date and an offset that mean nothing.
type TimeOfDay struct {
	Hour   int
	Minute int
}

// ParseTimeOfDay reads "HH:MM", and tolerates the "HH:MM:SS" PostgreSQL
// returns for a time column.
func ParseTimeOfDay(value string) (TimeOfDay, error) {
	trimmed := strings.TrimSpace(value)

	hourText, rest, ok := strings.Cut(trimmed, ":")
	if !ok {
		return TimeOfDay{}, fmt.Errorf("time of day: expected HH:MM, got %q", value)
	}
	minuteText, _, _ := strings.Cut(rest, ":")

	hour, err := strconv.Atoi(hourText)
	if err != nil {
		return TimeOfDay{}, fmt.Errorf("time of day: %q is not a number", hourText)
	}
	minute, err := strconv.Atoi(minuteText)
	if err != nil {
		return TimeOfDay{}, fmt.Errorf("time of day: %q is not a number", minuteText)
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return TimeOfDay{}, fmt.Errorf("time of day: %q is not a valid time", value)
	}

	return TimeOfDay{Hour: hour, Minute: minute}, nil
}

// String renders "HH:MM".
func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

// Response describes a repeat rule in structured form, for the client.
//
// The stored RRULE string never leaves the server. The client turns this into
// wording like "Every weekday", which is what a user should read
// (plans/phase4.md §21).
type Response struct {
	Frequency string   `json:"frequency"`
	Weekdays  []string `json:"weekdays"`
}

// weekdayNamesLong are the wire names for days, lowercase and unambiguous.
var weekdayNamesLong = [...]string{
	"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
}

// ToResponse renders a rule for the client.
func ToResponse(rule Rule) Response {
	days := make([]string, 0, len(rule.Weekdays))
	for _, day := range rule.Weekdays {
		days = append(days, weekdayNamesLong[day])
	}

	frequency := "daily"
	if rule.Frequency == FrequencyWeekly {
		frequency = "weekly"
	}

	return Response{Frequency: frequency, Weekdays: days}
}

// ParseRequest reads the structured form a client sends.
func ParseRequest(frequency string, weekdays []string) (Rule, error) {
	switch frequency {
	case "daily":
		return Daily(), nil

	case "weekly":
		days := make([]time.Weekday, 0, len(weekdays))
		for _, name := range weekdays {
			day, ok := weekdayFromName(name)
			if !ok {
				return Rule{}, ErrUnknownWeekday
			}
			days = append(days, day)
		}
		if len(days) == 0 {
			return Rule{}, ErrNoWeekdays
		}
		return Weekly(days...), nil

	default:
		return Rule{}, ErrUnknownFrequency
	}
}

func weekdayFromName(name string) (time.Weekday, bool) {
	for index, candidate := range weekdayNamesLong {
		if candidate == name {
			return time.Weekday(index), true
		}
	}
	return 0, false
}

// Message turns a parse failure into wording a person can act on.
//
// Shared by every domain that accepts a schedule, so "choose at least one day"
// reads identically whether the user was editing a care task or a medication.
func Message(err error) string {
	switch {
	case errors.Is(err, ErrNoWeekdays):
		return "Choose at least one day of the week."
	case errors.Is(err, ErrUnknownWeekday):
		return "We do not recognise one of those days."
	default:
		return "Choose whether this repeats daily or weekly."
	}
}
