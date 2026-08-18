/**
 * Care events: one chronological record of what has happened in a senior's
 * care, across every domain.
 *
 * Mirrors internal/careevents in the Go API and docs/03-domain-model.md's
 * CareEvent. One timeline, not four — a family member asking "what happened
 * yesterday?" is asking about their relative, not about tasks
 * (plans/phase7.md, objective).
 *
 * Events are written by the server as a side effect of domain actions. Nothing
 * here creates one, and there is no endpoint that would
 * (plans/phase7.md §21).
 */

import type { CursorPage } from './pagination';

/**
 * What happened.
 *
 * The vocabulary is the documentation's: docs/03-domain-model.md names these
 * and plans/phase7.md §2 adds the remaining domain actions. It is kept in step
 * with internal/careevents.Types and with the database CHECK constraint.
 */
export const CARE_EVENT_TYPES = [
  'MEMBER_INVITED',
  'MEMBER_JOINED',
  'MEMBER_REVOKED',
  'TASK_CREATED',
  'TASK_COMPLETED',
  'TASK_SKIPPED',
  'TASK_MISSED',
  'MEDICATION_CREATED',
  'MEDICATION_TAKEN',
  'MEDICATION_SKIPPED',
  'MEDICATION_MISSED',
  'APPOINTMENT_CREATED',
  'APPOINTMENT_COMPLETED',
  'APPOINTMENT_CANCELLED',
  'NOTE_ADDED',
] as const;
export type CareEventType = (typeof CARE_EVENT_TYPES)[number];

/** What an event is about, so a screen can route without parsing the type. */
export const CARE_EVENT_ENTITIES = [
  'task',
  'medication',
  'appointment',
  'relationship',
  'invitation',
  'note',
] as const;
export type CareEventEntity = (typeof CARE_EVENT_ENTITIES)[number];

/**
 * The structured detail carried with an event.
 *
 * Short labels only — a task's title, a medicine's name and dosage — and never
 * a copy of the record. The server strips empty values, so an absent key means
 * nobody supplied one (plans/phase7.md §9).
 */
export interface CareEventMetadata {
  taskTitle?: string;
  medicationName?: string;
  dosage?: string;
  appointmentTitle?: string;
  memberName?: string;
  role?: string;
}

/** One thing that happened in a senior's care. */
export interface CareEvent {
  id: string;
  seniorId: string;

  type: CareEventType;

  /** Null for an event no person performed. Never supplied by a client. */
  actorUserId: string | null;

  entityType: CareEventEntity;
  entityId: string;

  metadata: CareEventMetadata;

  /**
   * ISO-8601 instant. Render it in the senior's timezone: an event at 00:30 in
   * Karachi belongs to that day there, whatever the reader's device says
   * (plans/phase7.md §§13, 16).
   */
  occurredAt: string;
}

/** `GET /v1/seniors/{id}/activity` response. */
export type ActivityResponse = CursorPage<CareEvent>;
