import type { CarePermission, CareRole } from './care';

/**
 * A senior profile as returned by `/v1/seniors`.
 *
 * `role` and `permissions` describe the *caller's* relationship to this senior,
 * not the senior themselves — the same profile looks different to a daughter
 * and to a visiting caregiver. The client renders actions from `permissions`;
 * the API enforces them regardless (docs/02-permissions-and-authorization.md).
 */
export interface Senior {
  id: string;
  displayName: string;
  /** `YYYY-MM-DD`, or null when unknown. */
  dateOfBirth: string | null;
  photoUrl: string | null;
  phone: string | null;
  address: string | null;
  emergencyContact: string | null;
  /** True when this profile represents the caller themselves (Solo Mode). */
  isSelf: boolean;
  role: CareRole;
  permissions: CarePermission[];
  createdAt: string;
  updatedAt: string;
}

/** How the creator relates to the senior they are adding. */
export const SENIOR_CREATE_MODES = ['self', 'family_member', 'professional_caregiver'] as const;
export type SeniorCreateMode = (typeof SENIOR_CREATE_MODES)[number];

/** `POST /v1/seniors` request body. */
export interface CreateSeniorRequest {
  mode: SeniorCreateMode;
  displayName: string;
  dateOfBirth?: string | null;
  phone?: string | null;
  address?: string | null;
  emergencyContact?: string | null;
}

/** `PATCH /v1/seniors/{id}` request body. Absent fields are left unchanged. */
export interface UpdateSeniorRequest {
  displayName?: string;
  dateOfBirth?: string | null;
  phone?: string | null;
  address?: string | null;
  emergencyContact?: string | null;
}

/** `GET /v1/seniors` response. */
export interface SeniorListResponse {
  items: Senior[];
}

/** Reports whether the caller may perform an action for this senior. */
export function can(senior: Senior, permission: CarePermission): boolean {
  return senior.permissions.includes(permission);
}
