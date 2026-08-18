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

/**
 * The care event vocabulary lives in `care-event.ts`.
 *
 * A speculative list stood here from Phase 1, written before any of these
 * domains existed. It had drifted from the documentation in three places —
 * `MEMBER_REMOVED` for what the domain calls revoking, plus `APPOINTMENT_UPDATED`
 * and `PERMISSIONS_CHANGED`, which docs/03 does not name and which no action
 * emits — and it was missing every creation event. Two vocabularies is the
 * parallel naming system plans/phase7.md §2 forbids, so Phase 7 kept the
 * documented one and deleted this. Nothing referenced it.
 */
