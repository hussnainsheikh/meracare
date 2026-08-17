/**
 * Turning a date and time somebody typed into an instant.
 *
 * A care task set for "tomorrow at 09:00", or an extra dose "tonight at 22:00",
 * means that wall-clock time where the senior lives. The device's own zone is
 * irrelevant and using it would put the entry on the wrong hour for anybody
 * caring from another country (plans/phase4.md §14, plans/phase5.md §23).
 */

/**
 * Turns a local date and time in a named timezone into an ISO-8601 instant.
 *
 * The offset is discovered by formatting a candidate instant in that zone and
 * measuring how far it lands from the wall-clock time asked for. That avoids a
 * date library for the one calculation the MVP needs, and it is correct across
 * a daylight-saving boundary because the offset is read at the date in question
 * rather than assumed.
 */
export function instantFor(date: string, time: string, timezone: string): string {
  const naive = new Date(`${date}T${time}:00Z`);
  if (Number.isNaN(naive.getTime())) return new Date().toISOString();

  const offset = offsetAt(naive, timezone);
  return new Date(naive.getTime() - offset).toISOString();
}

/** The timezone's offset from UTC, in milliseconds, at a given instant. */
export function offsetAt(instant: Date, timezone: string): number {
  try {
    const formatter = new Intl.DateTimeFormat('en-GB', {
      timeZone: timezone,
      hour12: false,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });

    const parts = new Map(formatter.formatToParts(instant).map((part) => [part.type, part.value]));
    const at = (type: Intl.DateTimeFormatPartTypes) => Number(parts.get(type) ?? '0');

    const asUtc = Date.UTC(
      at('year'),
      at('month') - 1,
      at('day'),
      // Some locales render midnight as hour 24; both mean the same instant.
      parts.get('hour') === '24' ? 0 : at('hour'),
      at('minute'),
      at('second'),
    );

    return asUtc - instant.getTime();
  } catch {
    return 0;
  }
}

/** Today's date as `YYYY-MM-DD`, for a date field's starting value. */
export function todayInIso(): string {
  return new Date().toISOString().slice(0, 10);
}
