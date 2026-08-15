/**
 * Cursor pagination envelope used for activity, messages and other large
 * histories (`docs/05-api-and-backend-spec.md`).
 */
export interface CursorPage<T> {
  items: T[];
  /** Opaque cursor for the next page; `null` when the list is exhausted. */
  nextCursor: string | null;
}

export interface CursorPageParams {
  cursor?: string;
  limit?: number;
}
