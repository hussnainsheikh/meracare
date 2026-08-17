import type { Appointment, AppointmentKind, AppointmentStatus } from './appointment';
import { dateInTimezone, timeInTimezone } from './datetime';

/**
 * Plain-language wording for appointments.
 *
 * Nothing internal reaches a screen: not a status identifier, not a UTC
 * instant, not `doctor_visit`. Somebody reads "Doctor's visit" and "Booked",
 * never "scheduled" (plans/phase6.md §§7, 31).
 */

/**
 * What a status means, in words.
 *
 * Deliberately not colour names or ticks: appointment status must be legible
 * without relying on colour (plans/phase6.md §31).
 *
 * "Booked" rather than "Scheduled" because that is what a person says about an
 * appointment they have made, and because "scheduled" is the database's word.
 */
const STATUS_LABELS: Record<AppointmentStatus, string> = {
  scheduled: 'Booked',
  completed: 'Went',
  cancelled: 'Cancelled',
};

export function appointmentStatusLabel(status: AppointmentStatus): string {
  return STATUS_LABELS[status];
}

/** The tone a status should be shown in, for callers that also use colour. */
export type AppointmentTone = 'neutral' | 'positive' | 'muted';

const STATUS_TONES: Record<AppointmentStatus, AppointmentTone> = {
  scheduled: 'neutral',
  completed: 'positive',
  cancelled: 'muted',
};

export function appointmentStatusTone(status: AppointmentStatus): AppointmentTone {
  return STATUS_TONES[status];
}

/** How a kind reads on a card: "Doctor's visit", "Blood test or scan". */
const KIND_LABELS: Record<AppointmentKind, string> = {
  doctor_visit: "Doctor's visit",
  hospital_visit: 'Hospital visit',
  therapy: 'Therapy',
  laboratory: 'Blood test or scan',
  care_meeting: 'Care meeting',
  other: 'Something else',
};

export function appointmentKindLabel(kind: AppointmentKind): string {
  return KIND_LABELS[kind];
}

/** The kinds the create screen offers, in the order it offers them. */
export const SELECTABLE_APPOINTMENT_KINDS: { kind: AppointmentKind; label: string }[] = (
  Object.keys(KIND_LABELS) as AppointmentKind[]
).map((kind) => ({ kind, label: KIND_LABELS[kind] }));

/** The time an appointment starts, in the senior's timezone: "09:30". */
export function appointmentTimeLabel(
  appointment: Pick<Appointment, 'scheduledAt'>,
  timezone: string,
): string {
  return timeInTimezone(appointment.scheduledAt, timezone);
}

/** The date an appointment falls on, in the senior's timezone. */
export function appointmentDateLabel(
  appointment: Pick<Appointment, 'scheduledAt'>,
  timezone: string,
): string {
  return dateInTimezone(appointment.scheduledAt, timezone);
}

/**
 * When an appointment runs, in words: "09:30", or "09:30 to 10:15".
 *
 * The end is shown only when somebody recorded one, which is the minority of
 * appointments — a card that always said "09:30 to —" would be noise.
 */
export function appointmentWhenLabel(
  appointment: Pick<Appointment, 'scheduledAt' | 'endsAt'>,
  timezone: string,
): string {
  const start = timeInTimezone(appointment.scheduledAt, timezone);
  if (appointment.endsAt === null) return start;
  return `${start} to ${timeInTimezone(appointment.endsAt, timezone)}`;
}

/**
 * Where an appointment is, in one line: "Dr Ahmed · City Hospital".
 *
 * Either half may be missing, and both often are for a family care meeting.
 */
export function appointmentPlaceLabel(
  appointment: Pick<Appointment, 'providerName' | 'location'>,
): string {
  return [appointment.providerName, appointment.location].filter(Boolean).join(' · ');
}
