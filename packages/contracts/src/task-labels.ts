import { dateInTimezone, timeInTimezone } from './datetime';
import { WORKING_WEEK } from './recurrence';
import type { CareTask, TaskRecurrence, TaskStatus } from './task';

/**
 * Plain-language wording for care tasks.
 *
 * Nothing internal reaches a screen: not the RRULE the server stores, not a
 * status identifier, not a UTC instant. An older adult reads "Every weekday"
 * and "Not done yet", never "FREQ=WEEKLY;BYDAY=MO" or "pending"
 * (plans/phase4.md §§21, 34).
 *
 * The repeat wording itself lives in `recurrence.ts`, because a medication
 * schedule has to read the same way.
 */

export { recurrenceLabel, SELECTABLE_WEEKDAYS } from './recurrence';

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
  return timeInTimezone(task.scheduledFor, timezone);
}

/** The date a task falls due, in the senior's timezone. */
export function taskDateLabel(task: Pick<CareTask, 'scheduledFor'>, timezone: string): string {
  return dateInTimezone(task.scheduledFor, timezone);
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
