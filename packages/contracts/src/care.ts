/**
 * Care-domain constants shared between the mobile app and the Go API.
 *
 * These mirror `internal/care` in the Go API and the domain model in
 * `docs/03-domain-model.md`. Authorization is relationship-based: these values
 * describe a user's relationship to one senior, never a global capability.
 */

/** Role a user holds inside one senior's care circle. */
export const CARE_ROLES = ['senior', 'family_member', 'professional_caregiver'] as const;
export type CareRole = (typeof CARE_ROLES)[number];

/** Lifecycle of a care relationship. */
export const CARE_RELATIONSHIP_STATUSES = ['pending', 'active', 'revoked'] as const;
export type CareRelationshipStatus = (typeof CARE_RELATIONSHIP_STATUSES)[number];

/**
 * Discrete permissions granted per care relationship.
 *
 * Roles carry defaults, but the stored permission set is what the API
 * enforces — see `docs/02-permissions-and-authorization.md`.
 */
export const CARE_PERMISSIONS = [
  'senior.view',
  'senior.edit',
  'tasks.view',
  'tasks.manage',
  'tasks.complete',
  'medications.view',
  'medications.manage',
  'medications.record',
  'appointments.view',
  'appointments.manage',
  'notes.view',
  'notes.create',
  'activity.view',
  'members.view',
  'members.invite',
  'members.manage',
  'messages.participate',
] as const;
export type CarePermission = (typeof CARE_PERMISSIONS)[number];

/** Immutable events that drive the activity timeline (`docs/04-...`). */
export const CARE_EVENT_TYPES = [
  'TASK_COMPLETED',
  'TASK_MISSED',
  'TASK_SKIPPED',
  'MEDICATION_TAKEN',
  'MEDICATION_MISSED',
  'APPOINTMENT_CREATED',
  'APPOINTMENT_UPDATED',
  'NOTE_ADDED',
  'MEMBER_INVITED',
  'MEMBER_JOINED',
  'MEMBER_REMOVED',
  'PERMISSIONS_CHANGED',
] as const;
export type CareEventType = (typeof CARE_EVENT_TYPES)[number];
