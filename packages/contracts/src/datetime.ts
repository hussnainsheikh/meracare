/**
 * Reading an instant in somebody else's day.
 *
 * Care is scheduled where the senior lives. A daughter in London looking at her
 * mother in Karachi must see the time her mother will experience, not the time
 * on her own phone (plans/phase4.md §14, plans/phase5.md §23).
 */

/**
 * Formats an instant in a named timezone.
 *
 * An unknown timezone falls back to the device's rather than throwing: a dose
 * shown at a slightly wrong hour is recoverable, a screen that crashes is not.
 */
export function formatInTimezone(
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

/** The clock time of an instant, in the senior's timezone: "08:00". */
export function timeInTimezone(instant: string, timezone: string): string {
  return formatInTimezone(instant, timezone, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
}

/** The date of an instant, in the senior's timezone: "Monday 17 August". */
export function dateInTimezone(instant: string, timezone: string): string {
  return formatInTimezone(instant, timezone, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  });
}

/**
 * The calendar date an instant falls on in a timezone, as `YYYY-MM-DD`.
 *
 * Used to group a timeline by day. It has to be computed in the senior's zone
 * rather than from the UTC date: an event at 00:30 in Karachi is 19:30 the
 * previous day in UTC, and grouping by the latter would file it under
 * yesterday for everyone (plans/phase7.md §§13, 16).
 */
export function dateKeyInTimezone(instant: string, timezone: string): string {
  const at = new Date(instant);

  try {
    // en-CA renders as YYYY-MM-DD, which sorts and compares as a string.
    return new Intl.DateTimeFormat('en-CA', {
      timeZone: timezone,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    }).format(at);
  } catch {
    return new Intl.DateTimeFormat('en-CA', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    }).format(at);
  }
}
