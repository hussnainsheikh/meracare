/** The authenticated application user, as returned by `GET /v1/me`. */
export interface Me {
  id: string;
  displayName: string;
  avatarUrl: string | null;
  phone: string | null;
  createdAt: string;
  updatedAt: string;
}

/** `PATCH /v1/me` request body. */
export interface UpdateMeRequest {
  displayName?: string;
  avatarUrl?: string | null;
  phone?: string | null;
}
