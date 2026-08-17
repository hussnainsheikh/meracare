/**
 * How often something repeats, and how to say it out loud.
 *
 * Care tasks and medication schedules share one rule grammar, mirroring
 * internal/recurrence in the Go API. The server stores an RRULE string; it
 * never appears here, and never on a screen (plans/phase4.md §21,
 * plans/phase5.md §3).
 */

/** How often a rule fires. */
export const RECURRENCE_FREQUENCIES = ['daily', 'weekly'] as const;
export type RecurrenceFrequency = (typeof RECURRENCE_FREQUENCIES)[number];

export const WEEKDAYS = [
  'sunday',
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday',
] as const;
export type Weekday = (typeof WEEKDAYS)[number];

/** A repeat rule in structured form. */
export interface Recurrence {
  frequency: RecurrenceFrequency;
  /** Empty for a daily rule. */
  weekdays: Weekday[];
}

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

/** Monday to Friday, the shape most routines take. */
export const WORKING_WEEK: Weekday[] = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'];

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
export function recurrenceLabel(recurrence: Recurrence): string {
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
