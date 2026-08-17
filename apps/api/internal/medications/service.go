package medications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/meracare/api/internal/recurrence"
	"github.com/meracare/api/internal/seniors"
)

// SeniorLookup loads the senior a medication belongs to, for their timezone.
// internal/seniors.Repository satisfies it.
type SeniorLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (seniors.Senior, error)
}

// Service coordinates medications, their schedules, and the doses they produce.
type Service struct {
	medications *Repository
	seniors     SeniorLookup
}

// NewService builds the service.
func NewService(medications *Repository, seniorLookup SeniorLookup) *Service {
	return &Service{medications: medications, seniors: seniorLookup}
}

// ErrBadWindow is returned when the requested date range is unusable. It is a
// sentinel so the handler can answer 400 without inspecting the message.
var ErrBadWindow = errors.New("medications: unusable date range")

// horizon is how far ahead doses are written when a schedule is created or
// edited.
//
// Long enough that a family can see the coming weeks, short enough that a daily
// medicine does not write years of rows nobody has looked at. Reading beyond it
// generates on demand, so this is a convenience, not a limit.
const horizon = 60 * 24 * time.Hour

// missedLookback is how far back a missed-dose query generates before it reads.
//
// Doses are written when somebody reads the window they fall in, which leaves a
// gap: if nobody opened the app for a week, last Tuesday's dose was never
// written and so could never be reported missed. Generating the recent past
// before answering closes that gap without a job that has to be running for the
// answer to be true (plans/phase5.md §8).
const missedLookback = 14 * 24 * time.Hour

// maxWindowDays bounds an explicit date range, so one request cannot ask the
// server to expand every schedule for a decade.
const maxWindowDays = 92

// maxMissed bounds the missed list. A senior who is hundreds of doses behind
// needs a phone call, not a longer list.
const maxMissed = 200

// upcomingDays is how far "upcoming" looks ahead.
const upcomingDays = 7

// defaultHistoryPage and maxHistoryPage bound one page of history, so a client
// cannot ask for every dose ever recorded in one request (plans/phase5.md §16).
const (
	defaultHistoryPage = 30
	maxHistoryPage     = 100
)

// --- Medications -----------------------------------------------------------

// ScheduleInput is one time of day a medication is to be taken.
type ScheduleInput struct {
	Recurrence    recurrence.Rule
	ScheduledTime recurrence.TimeOfDay
}

// CreateInput is a request to add a medication.
type CreateInput struct {
	SeniorID        uuid.UUID
	CreatedByUserID uuid.UUID

	Name         string
	Dosage       string
	Form         Form
	Instructions string
	Notes        string

	// Schedules may be empty: a medicine can be recorded before anybody has
	// decided when it is taken, and times added afterwards.
	Schedules []ScheduleInput
}

// Created is a new medication together with the times it is taken and the doses
// already generated for the coming weeks.
type Created struct {
	Medication Medication
	Schedules  []Schedule
	Instances  []Instance
}

// Create adds a medication and its times.
func (s *Service) Create(ctx context.Context, input CreateInput, now time.Time) (Created, error) {
	medication, err := s.medications.CreateMedication(ctx, CreateMedicationParams{
		SeniorID:        input.SeniorID,
		CreatedByUserID: input.CreatedByUserID,
		Name:            input.Name,
		Dosage:          input.Dosage,
		Form:            input.Form,
		Instructions:    input.Instructions,
		Notes:           input.Notes,
	})
	if err != nil {
		return Created{}, err
	}

	schedules := make([]Schedule, 0, len(input.Schedules))
	for _, wanted := range input.Schedules {
		schedule, err := s.medications.CreateSchedule(ctx, CreateScheduleParams{
			MedicationID:  medication.ID,
			SeniorID:      medication.SeniorID,
			Recurrence:    wanted.Recurrence,
			ScheduledTime: wanted.ScheduledTime,
		})
		if err != nil {
			// A repeated time in one request is the client's mistake, not a
			// reason to leave the medicine half-created: the times that did
			// save are kept and the duplicate is reported.
			if errors.Is(err, ErrDuplicateSchedule) {
				continue
			}
			return Created{}, err
		}
		schedules = append(schedules, schedule)
	}

	instances, err := s.generateForMedication(ctx, medication, now, now.Add(horizon))
	if err != nil {
		return Created{}, err
	}

	return Created{Medication: medication, Schedules: schedules, Instances: instances}, nil
}

// Get loads one medication. The caller's access to its senior is checked by the
// handler, which cannot know the senior until this has run.
func (s *Service) Get(ctx context.Context, medicationID uuid.UUID) (Medication, error) {
	return s.medications.GetMedication(ctx, medicationID)
}

// List returns a senior's medications.
func (s *Service) List(ctx context.Context, seniorID uuid.UUID) ([]Medication, error) {
	return s.medications.ListMedications(ctx, seniorID)
}

// GetSchedule loads one schedule. The handler checks that it belongs to the
// medication the caller was authorized for.
func (s *Service) GetSchedule(ctx context.Context, scheduleID uuid.UUID) (Schedule, error) {
	return s.medications.GetSchedule(ctx, scheduleID)
}

// Schedules returns every time of day one medication is taken.
func (s *Service) Schedules(ctx context.Context, medicationID uuid.UUID) ([]Schedule, error) {
	return s.medications.ListSchedules(ctx, medicationID)
}

// NextDose returns the earliest pending dose from now, or nil when there is
// none. It powers "next dose" on the detail screen (plans/phase5.md §15).
func (s *Service) NextDose(
	ctx context.Context,
	medication Medication,
	now time.Time,
) (*Instance, error) {
	if !medication.Active {
		return nil, nil
	}
	if _, err := s.generateForMedication(ctx, medication, now, now.Add(horizon)); err != nil {
		return nil, err
	}
	return s.medications.NextPendingDose(ctx, medication.ID, now)
}

// UpdateInput edits a medication.
type UpdateInput struct {
	MedicationID uuid.UUID
	Name         *string
	Dosage       *string
	Instructions *string
	Notes        *string
	Active       *bool
	Form         *Form
	ClearForm    bool
}

// Update edits a medication and brings its unstarted doses into line.
//
// Doses that are already due, or that anybody has taken or skipped, are left
// exactly as they were: renaming a medicine or correcting its dosage must not
// rewrite what somebody already swallowed. Only future pending doses are
// discarded and regenerated, so tomorrow's dose says what the label says today
// (plans/phase5.md §12).
func (s *Service) Update(ctx context.Context, input UpdateInput, now time.Time) (Medication, error) {
	medication, err := s.medications.UpdateMedication(ctx, input.MedicationID, UpdateMedicationParams{
		Name:         input.Name,
		Dosage:       input.Dosage,
		Instructions: input.Instructions,
		Notes:        input.Notes,
		Active:       input.Active,
		Form:         input.Form,
		ClearForm:    input.ClearForm,
	})
	if err != nil {
		return Medication{}, err
	}

	if _, err := s.medications.DiscardFuturePendingForMedication(ctx, medication.ID, now); err != nil {
		return Medication{}, err
	}

	// A stopped medicine produces nothing further; its history remains
	// (plans/phase5.md §13).
	if !medication.Active {
		return medication, nil
	}

	if _, err := s.generateForMedication(ctx, medication, now, now.Add(horizon)); err != nil {
		return Medication{}, err
	}
	return medication, nil
}

// --- Schedules -------------------------------------------------------------

// AddSchedule adds a time of day to a medication.
func (s *Service) AddSchedule(
	ctx context.Context,
	medication Medication,
	input ScheduleInput,
	now time.Time,
) (Schedule, error) {
	schedule, err := s.medications.CreateSchedule(ctx, CreateScheduleParams{
		MedicationID:  medication.ID,
		SeniorID:      medication.SeniorID,
		Recurrence:    input.Recurrence,
		ScheduledTime: input.ScheduledTime,
	})
	if err != nil {
		return Schedule{}, err
	}

	if _, err := s.generateForMedication(ctx, medication, now, now.Add(horizon)); err != nil {
		return Schedule{}, err
	}
	return schedule, nil
}

// UpdateScheduleInput edits one time of day.
type UpdateScheduleInput struct {
	ScheduleID    uuid.UUID
	Recurrence    *recurrence.Rule
	ScheduledTime *recurrence.TimeOfDay
	Active        *bool
}

// UpdateSchedule edits a time of day and regenerates what has not happened yet.
func (s *Service) UpdateSchedule(
	ctx context.Context,
	medication Medication,
	input UpdateScheduleInput,
	now time.Time,
) (Schedule, error) {
	schedule, err := s.medications.UpdateSchedule(ctx, input.ScheduleID, UpdateScheduleParams{
		Recurrence:    input.Recurrence,
		ScheduledTime: input.ScheduledTime,
		Active:        input.Active,
	})
	if err != nil {
		return Schedule{}, err
	}

	// Everything from this schedule that has not fallen due yet goes, whether
	// the rule moved or the time did. What is left is regenerated from the new
	// arrangement below.
	if _, err := s.medications.DiscardFuturePending(ctx, schedule.ID, now); err != nil {
		return Schedule{}, err
	}

	if !schedule.Active {
		return schedule, nil
	}

	if _, err := s.generateForMedication(ctx, medication, now, now.Add(horizon)); err != nil {
		return Schedule{}, err
	}
	return schedule, nil
}

// --- Doses -----------------------------------------------------------------

// Scope names one of the dose views the app asks for.
type Scope string

const (
	// ScopeToday is the senior's own calendar day.
	ScopeToday Scope = "today"
	// ScopeUpcoming is the week after today.
	ScopeUpcoming Scope = "upcoming"
	// ScopeMissed is everything whose window has passed with nobody acting.
	ScopeMissed Scope = "missed"
	// ScopeWindow is an explicit date range.
	ScopeWindow Scope = "window"
)

// ListDosesInput asks for a senior's doses.
type ListDosesInput struct {
	SeniorID uuid.UUID
	Scope    Scope
	// From and To bound ScopeWindow, as instants.
	From time.Time
	To   time.Time
}

// ListDoses returns a senior's doses for the requested view.
//
// Doses are written for the window before it is read. That is a write on a read
// path, and it is deliberate: the alternative is a scheduler that must be
// running for the data to be correct, and a medication schedule that silently
// stops when a background job dies is worse than one that costs an insert to
// look at. The unique constraint on (schedule_id, scheduled_for) makes the
// write idempotent and safe under concurrency (plans/phase5.md §§8, 27).
func (s *Service) ListDoses(
	ctx context.Context,
	input ListDosesInput,
	now time.Time,
) ([]Instance, error) {
	senior, err := s.seniors.GetByID(ctx, input.SeniorID)
	if err != nil {
		return nil, err
	}

	from, to, err := s.window(input, senior.Location(), now)
	if err != nil {
		return nil, err
	}

	if err := s.generateForSenior(ctx, senior, from, to); err != nil {
		return nil, err
	}

	if input.Scope == ScopeMissed {
		// Only what the grace period has already run out on counts as missed,
		// which is the same reading Instance.EffectiveStatus gives.
		return s.medications.ListMissed(ctx, input.SeniorID, now.Add(-MissedAfter), maxMissed)
	}

	return s.medications.ListWindow(ctx, input.SeniorID, from, to)
}

// window turns a scope into an instant range in the senior's own day.
func (s *Service) window(
	input ListDosesInput,
	location *time.Location,
	now time.Time,
) (from, to time.Time, err error) {
	local := now.In(location)
	year, month, day := local.Date()
	startOfToday := time.Date(year, month, day, 0, 0, 0, 0, location)

	switch input.Scope {
	case ScopeToday:
		return startOfToday, startOfToday.AddDate(0, 0, 1), nil

	case ScopeUpcoming:
		start := startOfToday.AddDate(0, 0, 1)
		return start, start.AddDate(0, 0, upcomingDays), nil

	case ScopeMissed:
		// The range here is what gets generated, not what gets returned: the
		// query below reads every pending dose older than the grace period.
		return now.Add(-missedLookback), now, nil

	case ScopeWindow:
		if !input.To.After(input.From) {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: it must end after it starts", ErrBadWindow)
		}
		if input.To.Sub(input.From) > maxWindowDays*24*time.Hour {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: longer than %d days", ErrBadWindow, maxWindowDays)
		}
		return input.From, input.To, nil

	default:
		return time.Time{}, time.Time{}, fmt.Errorf("%w: unknown scope %q", ErrBadWindow, input.Scope)
	}
}

// History returns one page of a medication's doses, newest first.
//
// The first page generates the recent past and the near future first, so the
// history is complete rather than showing only the days somebody happened to
// open the app on. Later pages read backwards through rows that already exist.
func (s *Service) History(
	ctx context.Context,
	medication Medication,
	cursor string,
	limit int,
	now time.Time,
) (HistoryPage, error) {
	if limit <= 0 || limit > maxHistoryPage {
		limit = defaultHistoryPage
	}

	if cursor == "" {
		if _, err := s.generateForMedication(
			ctx, medication, now.Add(-missedLookback), now.Add(horizon),
		); err != nil {
			return HistoryPage{}, err
		}
	}

	return s.medications.ListHistory(ctx, medication.ID, cursor, limit)
}

// AddDose records a single dose that no rule produced — an extra tablet
// tonight, or a course short enough that a repeat rule would be ceremony.
//
// This is the one-time schedule case: the dose exists on its own and no
// schedule edit will ever touch it (plans/phase5.md §35).
func (s *Service) AddDose(
	ctx context.Context,
	medication Medication,
	createdByUserID uuid.UUID,
	scheduledFor time.Time,
) (Instance, error) {
	return s.medications.CreateInstance(ctx, CreateInstanceParams{
		MedicationID:    medication.ID,
		SeniorID:        medication.SeniorID,
		CreatedByUserID: createdByUserID,
		Name:            medication.Name,
		Dosage:          medication.Dosage,
		ScheduledFor:    scheduledFor,
	})
}

// GetInstance loads one dose.
func (s *Service) GetInstance(ctx context.Context, instanceID uuid.UUID) (Instance, error) {
	return s.medications.GetInstance(ctx, instanceID)
}

// ActInput records an action on one dose.
type ActInput struct {
	InstanceID uuid.UUID
	Action     Action
	// ActorID is the authenticated caller. It is never read from the request
	// body: a client that could name the actor could attribute somebody else's
	// medicine to them (plans/phase5.md §6).
	ActorID uuid.UUID
	Notes   *string
}

// ActResult reports what happened, including whether the action had already
// been recorded.
type ActResult struct {
	Instance Instance
	// Repeat is true when the dose was already in the requested state, which
	// makes a replayed mutation a success rather than a conflict.
	Repeat bool
}

// Act records a dose as taken or skipped.
//
// The conditional UPDATE in the repository is the real guard. When it matches
// no row the dose is no longer pending, and this re-reads it to decide between
// two very different answers: the same action arriving twice, which succeeds
// quietly, and a different outcome already recorded, which is refused so the
// server's version stands (plans/phase5.md §§21–22).
func (s *Service) Act(ctx context.Context, input ActInput) (ActResult, error) {
	instance, err := s.medications.Act(ctx, ActParams{
		InstanceID: input.InstanceID,
		Action:     input.Action,
		ActorID:    input.ActorID,
		Notes:      input.Notes,
	})
	if err == nil {
		return ActResult{Instance: instance}, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return ActResult{}, err
	}

	current, err := s.medications.GetInstance(ctx, input.InstanceID)
	if err != nil {
		return ActResult{}, err
	}

	_, repeat, err := Transition(current.Status, input.Action)
	if err != nil {
		return ActResult{}, err
	}
	return ActResult{Instance: current, Repeat: repeat}, nil
}

// --- Generation ------------------------------------------------------------

// generateForSenior writes every live schedule's doses for the window.
func (s *Service) generateForSenior(
	ctx context.Context,
	senior seniors.Senior,
	from, to time.Time,
) error {
	scheduled, err := s.medications.ListDueSchedulesForSenior(ctx, senior.ID)
	if err != nil {
		return err
	}
	return s.generate(ctx, senior.Location(), scheduled, from, to)
}

// generateForMedication writes one medication's doses for the window and
// returns them.
func (s *Service) generateForMedication(
	ctx context.Context,
	medication Medication,
	from, to time.Time,
) ([]Instance, error) {
	senior, err := s.seniors.GetByID(ctx, medication.SeniorID)
	if err != nil {
		return nil, err
	}

	scheduled, err := s.medications.ListDueSchedulesForMedication(ctx, medication.ID)
	if err != nil {
		return nil, err
	}
	if err := s.generate(ctx, senior.Location(), scheduled, from, to); err != nil {
		return nil, err
	}

	return s.medications.ListMedicationWindow(ctx, medication.ID, from, to)
}

// generate expands each schedule over the window and writes what it produces.
//
// Generation starts no earlier than the schedule's own creation: asking for
// last month must not invent a month of doses nobody was ever expected to take,
// and then report them all missed.
func (s *Service) generate(
	ctx context.Context,
	location *time.Location,
	scheduled []Scheduled,
	from, to time.Time,
) error {
	for _, entry := range scheduled {
		if !entry.Schedule.Active {
			continue
		}

		start := from
		if entry.Schedule.CreatedAt.After(start) {
			start = entry.Schedule.CreatedAt
		}
		if !to.After(start) {
			continue
		}

		occurrences := entry.Schedule.Recurrence.Occurrences(
			entry.Schedule.ScheduledTime, location, start, to,
		)
		if err := s.medications.Materialise(
			ctx, entry, entry.CreatedByUserID, occurrences,
		); err != nil {
			return err
		}
	}
	return nil
}
