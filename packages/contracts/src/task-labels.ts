import type { CareTask, TaskRecurrence, TaskStatus, Weekday } from './task';

/**
 * Plain-language wording for care tasks.
 *
 * Nothing internal reaches a screen: not the RRULE the server stores, not a
 * status identifier, not a UTC instant. An older adult reads "Every weekday"
 * and "Not done yet", never "FREQ=WEEKLY;BYDAY=MO" or "pending"
 * (plans/phase4.md §§21, 34).
 */

const WEEKDAY_NAMES: Record<Weekday, string> = {
  sunday: 'Sunday',
  monday: 'Monday',
  tuesday: 'Tuesday',
  wednesday: 'Wednesday',
  thursday: 'Thursday',
  friday: 'Friday',
  saturday: 'Saturday',
};

const WEEKDAY_ORDER: Weekday[] = [
  'sunday',
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday',
];

const WORKING_WEEK: Weekday[] = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'];

/** Joins a list the way a person would say it: "Monday, Wednesday and Friday". */
function sentenceList(parts: string[]): string {
  if (parts.length === 0) return '';
  if (parts.length === 1) return parts[0] as string;

  const last = parts[parts.length - 1] as string;
  return `${parts.slice(0, -1).join(', ')} and ${last}`;
}

function sameDays(a: Weekday[], b: Weekday[]): boolean {
  return a.length === b.length && b.every((day) => a.includes(day));
}

/**
 * Describes a repeat rule in words.
 *
 * The common shapes get the wording somebody would actually use — "Every
 * weekday" rather than "Monday, Tuesday, Wednesday, Thursday and Friday".
 */
export function recurrenceLabel(recurrence: TaskRecurrence): string {
  if (recurrence.frequency === 'daily') {
    return 'Every day';
  }

  const days = WEEKDAY_ORDER.filter((day) => recurrence.weekdays.includes(day));

  if (days.length === 0) return 'Every week';
  if (sameDays(days, WORKING_WEEK)) return 'Every weekday';
  if (sameDays(days, ['saturday', 'sunday'])) return 'Every weekend';
  if (days.length === 7) return 'Every day';
  if (days.length === 1) return `Every ${WEEKDAY_NAMES[days[0] as Weekday]}`;

  return `Every ${sentenceList(days.map((day) => WEEKDAY_NAMES[day]))}`;
}

/** A one-time task has no rule to describe. */
export const ONE_TIME_LABEL = 'Once';

/**
 * What a status means, in words.
 *
 * Deliberately not colour names or ticks: status must be legible without
 * relying on colour (plans/phase4.md §34).
 */
const STATUS_LABELS: Record<TaskStatus, string> = {
  pending: 'To do',
  overdue: 'Overdue',
  completed: 'Done',
  skipped: 'Skipped',
  cancelled: 'Cancelled',
};

export function statusLabel(status: TaskStatus): string {
  return STATUS_LABELS[status];
}

/** The tone a status should be shown in, for callers that also use colour. */
export type TaskTone = 'neutral' | 'positive' | 'attention' | 'muted';

const STATUS_TONES: Record<TaskStatus, TaskTone> = {
  pending: 'neutral',
  overdue: 'attention',
  completed: 'positive',
  skipped: 'muted',
  cancelled: 'muted',
};

export function statusTone(status: TaskStatus): TaskTone {
  return STATUS_TONES[status];
}

/**
 * The time of day a task falls due, in the senior's timezone.
 *
 * A family member in London looking at a parent in Karachi must see the time
 * their parent will experience, not the time on their own phone
 * (plans/phase4.md §14).
 */
export function taskTimeLabel(task: Pick<CareTask, 'scheduledFor'>, timezone: string): string {
  return formatInTimezone(task.scheduledFor, timezone, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
}

/** The date a task falls due, in the senior's timezone. */
export function taskDateLabel(task: Pick<CareTask, 'scheduledFor'>, timezone: string): string {
  return formatInTimezone(task.scheduledFor, timezone, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  });
}

/**
 * Formats an instant in a named timezone.
 *
 * An unknown timezone falls back to the device's rather than throwing: a task
 * shown at a slightly wrong hour is recoverable, a screen that crashes is not.
 */
function formatInTimezone(
  instant: string,
  timezone: string,
  options: Intl.DateTimeFormatOptions,
): string {
  const at = new Date(instant);

  try {
    return new Intl.DateTimeFormat('en-GB', { ...options, timeZone: timezone }).format(at);
  } catch {
    return new Intl.DateTimeFormat('en-GB', options).format(at);
  }
}

/**
 * The repeat options the create form offers.
 *
 * A deliberately short list: the MVP needs these shapes, and a general
 * recurrence editor would be a lot of interface for a screen an older adult has
 * to understand (plans/phase4.md §7).
 */
export interface RecurrencePreset {
  key: string;
  label: string;
  description: string;
  /** Null means the task happens once. */
  recurrence: TaskRecurrence | null;
}

export const RECURRENCE_PRESETS: RecurrencePreset[] = [
  {
    key: 'once',
    label: 'Once',
    description: 'A single task on one day.',
    recurrence: null,
  },
  {
    key: 'daily',
    label: 'Every day',
    description: 'Repeats every day at the same time.',
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

/** The day names a custom weekly rule chooses between, Monday first. */
export const SELECTABLE_WEEKDAYS: { weekday: Weekday; label: string; short: string }[] = [
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday',
  'sunday',
].map((day) => ({
  weekday: day as Weekday,
  label: WEEKDAY_NAMES[day as Weekday],
  short: WEEKDAY_NAMES[day as Weekday].slice(0, 3),
}));
