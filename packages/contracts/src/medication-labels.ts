import { dateInTimezone, timeInTimezone } from './datetime';
import type {
  MedicationDose,
  MedicationForm,
  MedicationSchedule,
  MedicationStatus,
} from './medication';
import { recurrenceLabel, WORKING_WEEK, type Recurrence } from './recurrence';

/**
 * Plain-language wording for medication.
 *
 * Nothing internal reaches a screen: not the RRULE the server stores, not a
 * status identifier, not a UTC instant. Somebody reads "Every day at 08:00" and
 * "Not taken", never "FREQ=DAILY" or "pending" (plans/phase5.md §§11, 31).
 */

/**
 * What a status means, in words.
 *
 * Deliberately not colour names or ticks: medication status must be legible
 * without relying on colour (plans/phase5.md §31).
 */
const STATUS_LABELS: Record<MedicationStatus, string> = {
  pending: 'To take',
  taken: 'Taken',
  skipped: 'Skipped',
  missed: 'Missed',
};

export function doseStatusLabel(status: MedicationStatus): string {
  return STATUS_LABELS[status];
}

/** The tone a status should be shown in, for callers that also use colour. */
export type MedicationTone = 'neutral' | 'positive' | 'attention' | 'muted';

const STATUS_TONES: Record<MedicationStatus, MedicationTone> = {
  pending: 'neutral',
  taken: 'positive',
  missed: 'attention',
  skipped: 'muted',
};

export function doseStatusTone(status: MedicationStatus): MedicationTone {
  return STATUS_TONES[status];
}

/** How a form reads on a card: "Tablet", "Eye drops". */
const FORM_LABELS: Record<MedicationForm, string> = {
  tablet: 'Tablet',
  capsule: 'Capsule',
  liquid: 'Liquid',
  drops: 'Drops',
  inhaler: 'Inhaler',
  injection: 'Injection',
  patch: 'Patch',
  cream: 'Cream',
  other: 'Other',
};

export function formLabel(form: MedicationForm): string {
  return FORM_LABELS[form];
}

/** The forms the create screen offers, in the order it offers them. */
export const SELECTABLE_FORMS: { form: MedicationForm; label: string }[] = (
  Object.keys(FORM_LABELS) as MedicationForm[]
).map((form) => ({ form, label: FORM_LABELS[form] }));

/**
 * A schedule in words: "Every day at 08:00", "Every weekday at 09:00".
 *
 * The time is the schedule's own wall-clock time, which is already in the
 * senior's timezone — no conversion, because it is not an instant.
 */
export function scheduleLabel(
  schedule: Pick<MedicationSchedule, 'recurrence' | 'scheduledTime'>,
): string {
  return `${recurrenceLabel(schedule.recurrence)} at ${schedule.scheduledTime}`;
}

/** Several schedules in one line: "Every day at 08:00 and 20:00". */
export function schedulesLabel(schedules: MedicationSchedule[]): string {
  const live = schedules.filter((schedule) => schedule.active);
  if (live.length === 0) return 'No times set yet';
  if (live.length === 1) return scheduleLabel(live[0] as MedicationSchedule);

  // The common case is one rule at several times, which reads far better as
  // "Every day at 08:00 and 20:00" than as two sentences.
  const [first, ...rest] = live;
  const sameRule = rest.every(
    (schedule) => recurrenceLabel(schedule.recurrence) === recurrenceLabel(first!.recurrence),
  );

  const times = live.map((schedule) => schedule.scheduledTime).sort();
  const spoken = `${times.slice(0, -1).join(', ')} and ${times[times.length - 1]}`;

  if (sameRule) {
    return `${recurrenceLabel(first!.recurrence)} at ${spoken}`;
  }
  return live.map(scheduleLabel).join('; ');
}

/** The time of day a dose falls due, in the senior's timezone. */
export function doseTimeLabel(
  dose: Pick<MedicationDose, 'scheduledFor'>,
  timezone: string,
): string {
  return timeInTimezone(dose.scheduledFor, timezone);
}

/** The date a dose falls due, in the senior's timezone. */
export function doseDateLabel(
  dose: Pick<MedicationDose, 'scheduledFor'>,
  timezone: string,
): string {
  return dateInTimezone(dose.scheduledFor, timezone);
}

/** When the next dose is, in words, for the detail screen. */
export function nextDoseLabel(nextDoseAt: string | null, timezone: string): string {
  if (nextDoseAt === null) return 'Nothing scheduled';
  return `${dateInTimezone(nextDoseAt, timezone)} at ${timeInTimezone(nextDoseAt, timezone)}`;
}

/**
 * The repeat options the medication form offers.
 *
 * A deliberately short list, matching the care-task form: the MVP needs these
 * shapes, and a general recurrence editor would be a lot of interface for a
 * screen an older adult has to understand (plans/phase5.md §3).
 */
export interface MedicationRecurrencePreset {
  key: string;
  label: string;
  description: string;
  recurrence: Recurrence;
}

export const MEDICATION_RECURRENCE_PRESETS: MedicationRecurrencePreset[] = [
  {
    key: 'daily',
    label: 'Every day',
    description: 'The usual pattern for a regular medicine.',
    recurrence: { frequency: 'daily', weekdays: [] },
  },
  {
    key: 'weekdays',
    label: 'Every weekday',
    description: 'Monday to Friday.',
    recurrence: { frequency: 'weekly', weekdays: WORKING_WEEK },
  },
  {
    key: 'custom',
    label: 'Chosen days',
    description: 'Pick which days of the week.',
    recurrence: { frequency: 'weekly', weekdays: [] },
  },
];
