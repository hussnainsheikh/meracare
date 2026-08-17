/**
 * Medication: what a person takes, when they take it, and whether they did.
 *
 * Mirrors internal/medications in the Go API. The distinction the domain draws
 * (docs/03-domain-model.md) is between the *medication* — "Metformin, 500 mg" —
 * the *schedule* that says when it is taken, and an *instance*: one concrete
 * dose somebody takes or skips.
 *
 * This is care coordination, not clinical software. The app records what a
 * family entered and reminds them of it; nothing here reasons about whether a
 * medicine or a dose is appropriate (plans/phase5.md §32).
 */

import type { CursorPage } from './pagination';
import type { Recurrence } from './recurrence';

/**
 * The state of one dose.
 *
 * `missed` is derived by the server from the clock, never stored: a dose is
 * missed exactly when its time and its grace period have passed and nobody has
 * acted on it. The client renders this value directly.
 */
export const MEDICATION_STATUSES = ['pending', 'taken', 'skipped', 'missed'] as const;
export type MedicationStatus = (typeof MEDICATION_STATUSES)[number];

/** Statuses that mean somebody has already decided the outcome. */
export const SETTLED_MEDICATION_STATUSES = ['taken', 'skipped'] as const;

/** Reports whether a dose still needs taking. */
export function isDoseOpen(dose: Pick<MedicationDose, 'status'>): boolean {
  return dose.status === 'pending' || dose.status === 'missed';
}

/** Reports whether the dose's window has passed with nobody acting on it. */
export function isDoseMissed(dose: Pick<MedicationDose, 'status'>): boolean {
  return dose.status === 'missed';
}

/** The physical shape a medicine comes in. */
export const MEDICATION_FORMS = [
  'tablet',
  'capsule',
  'liquid',
  'drops',
  'inhaler',
  'injection',
  'patch',
  'cream',
  'other',
] as const;
export type MedicationForm = (typeof MEDICATION_FORMS)[number];

/** A medicine a senior takes. */
export interface Medication {
  id: string;
  seniorId: string;

  name: string;
  /** As entered: "500 mg", "1 tablet", "two puffs". Never computed. */
  dosage: string;
  /** Null when nobody said what shape it comes in. */
  form: MedicationForm | null;
  instructions: string | null;
  notes: string | null;

  /** False for a medicine that has been stopped. Its doses remain. */
  active: boolean;

  createdAt: string;
  updatedAt: string;
}

/** A medicine with the times it is taken, as the detail screen needs it. */
export interface MedicationDetail extends Medication {
  schedules: MedicationSchedule[];
  /** ISO-8601 instant, or null when the medicine is stopped or unscheduled. */
  nextDoseAt: string | null;
}

/**
 * One time of day a medication is taken, and on which days.
 *
 * "Every day at 08:00 and 20:00" is two schedules, not one rule with two times.
 * That is what makes it possible to stop the evening dose without touching the
 * morning one (plans/phase5.md §3).
 */
export interface MedicationSchedule {
  id: string;
  medicationId: string;
  recurrence: Recurrence;
  /** Wall-clock `HH:MM` in the senior's timezone, not an instant. */
  scheduledTime: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

/** One concrete dose. */
export interface MedicationDose {
  id: string;
  /** Null for a one-off dose that no schedule produced. */
  scheduleId: string | null;
  medicationId: string;
  seniorId: string;

  /**
   * The name and dosage as they read when the dose was scheduled, so a history
   * entry is not rewritten by a later change to the prescription.
   */
  name: string;
  dosage: string;

  /** ISO-8601 instant. Render it in the senior's timezone, not the device's. */
  scheduledFor: string;
  status: MedicationStatus;
  /** True when this came from a schedule. */
  recurring: boolean;

  takenAt: string | null;
  takenBy: string | null;
  skippedAt: string | null;
  skippedBy: string | null;

  notes: string | null;

  createdAt: string;
  updatedAt: string;
}

/** Which set of doses to fetch. */
export const MEDICATION_DOSE_SCOPES = ['today', 'upcoming', 'missed', 'window'] as const;
export type MedicationDoseScope = (typeof MEDICATION_DOSE_SCOPES)[number];

/** One time of day, as a request sends it. */
export interface MedicationScheduleInput {
  recurrence: Recurrence;
  /** Wall-clock `HH:MM` in the senior's timezone. */
  scheduledTime: string;
}

/** `POST /v1/seniors/{id}/medications` request body. */
export interface CreateMedicationRequest {
  name: string;
  dosage?: string;
  form?: MedicationForm | null;
  instructions?: string;
  notes?: string;
  /** May be empty: a medicine can be recorded before its times are decided. */
  schedules?: MedicationScheduleInput[];
}

/** `POST /v1/seniors/{id}/medications` response. */
export interface CreateMedicationResponse {
  medication: MedicationDetail;
  /** The doses already generated for the coming weeks. */
  doses: MedicationDose[];
}

/** `PATCH /v1/medications/{id}` request body. Absent fields are left unchanged. */
export interface UpdateMedicationRequest {
  name?: string;
  dosage?: string;
  form?: MedicationForm;
  /** Sent to remove the form, which absence alone cannot express. */
  clearForm?: boolean;
  instructions?: string;
  notes?: string;
  /** False stops the medicine. Its doses are kept. */
  active?: boolean;
}

/** `PATCH /v1/medications/{id}/schedules/{id}` request body. */
export interface UpdateMedicationScheduleRequest {
  recurrence?: Recurrence;
  scheduledTime?: string;
  /** False stops this time of day without touching the others. */
  active?: boolean;
}

/** `POST /v1/medications/{id}/doses` request body: one dose, no rule behind it. */
export interface AddMedicationDoseRequest {
  /** ISO-8601 instant. */
  scheduledFor: string;
}

/** An optional note recorded with a dose. */
export interface MedicationActionRequest {
  notes?: string;
}

/** `GET /v1/seniors/{id}/medications` response. */
export interface MedicationListResponse {
  items: Medication[];
}

/** `GET /v1/seniors/{id}/medications/doses` response. */
export interface MedicationDoseListResponse {
  items: MedicationDose[];
}

/** `GET /v1/medications/{id}/schedules` response. */
export interface MedicationScheduleListResponse {
  items: MedicationSchedule[];
}

/** `GET /v1/medications/{id}/instances` response: history, newest first. */
export type MedicationHistoryResponse = CursorPage<MedicationDose>;
