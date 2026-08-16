import type { CarePermission, CareRelationshipStatus, CareRole } from './care';

/**
 * Care circle membership and invitations.
 *
 * Mirrors `internal/relationships` and `internal/invitations` in the Go API.
 */

/** A person in a senior's care circle. */
export interface CircleMember {
  /** The relationship ID — what the member endpoints address. */
  id: string;
  userId: string;
  displayName: string;
  role: CareRole;
  permissions: CarePermission[];
  status: CareRelationshipStatus;
  /** True for the person the circle exists for. Their membership is fixed. */
  isSenior: boolean;
  /** True for the reader's own membership. */
  isSelf: boolean;
  joinedAt: string;
  updatedAt: string;
}

export interface CircleMemberListResponse {
  items: CircleMember[];
}

/** Roles that can be invited. The senior's seat is not one of them. */
export const INVITABLE_ROLES = ['family_member', 'professional_caregiver'] as const;
export type InvitableRole = (typeof INVITABLE_ROLES)[number];

export const INVITATION_STATUSES = ['pending', 'accepted', 'revoked', 'expired'] as const;
export type InvitationStatus = (typeof INVITATION_STATUSES)[number];

/** An invitation, as seen by the circle managing it. */
export interface Invitation {
  id: string;
  seniorId: string;
  inviteeEmail: string;
  role: InvitableRole;
  permissions: CarePermission[];
  status: InvitationStatus;
  expiresAt: string;
  createdAt: string;
}

export interface InvitationListResponse {
  items: Invitation[];
}

/** `POST /v1/seniors/{id}/invitations` request body. */
export interface CreateInvitationRequest {
  email: string;
  role: InvitableRole;
  /**
   * Omit to use the role's defaults. Whatever is sent is validated against the
   * inviter's own permissions server-side — the client cannot widen a grant.
   */
  permissions?: CarePermission[];
}

/**
 * `POST /v1/seniors/{id}/invitations` response.
 *
 * The token is delivered here and nowhere else; it cannot be retrieved later.
 */
export interface CreateInvitationResponse {
  invitation: Invitation;
  token: string;
}

/**
 * What the holder of an invitation token sees before accepting.
 *
 * Deliberately thin: anyone holding the token can read it.
 */
export interface InvitationPreview {
  seniorName: string;
  inviterName: string;
  inviteeEmail: string;
  role: InvitableRole;
  permissions: CarePermission[];
  status: InvitationStatus;
  expiresAt: string;
}

/** `POST /v1/invitations/{token}/accept` response. */
export interface AcceptInvitationResponse {
  seniorId: string;
  role: CareRole;
  permissions: CarePermission[];
}

/** `PATCH /v1/seniors/{id}/members/{relationshipId}` request body. */
export interface UpdateMemberRequest {
  permissions: CarePermission[];
}
