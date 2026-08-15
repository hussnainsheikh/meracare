/**
 * API error envelope.
 *
 * Mirrors `pkg/httpx.ErrorResponse` in the Go API. Every non-2xx response from
 * `/v1/...` uses this shape — see `docs/05-api-and-backend-spec.md`.
 */
export interface ApiErrorBody {
  error: {
    code: ApiErrorCode | string;
    message: string;
    /** Field-level validation failures, keyed by request field name. */
    details?: Record<string, string>;
  };
}

/**
 * Stable error codes the client is allowed to branch on.
 *
 * Keep in sync with `pkg/httpx/errors.go`.
 */
export const API_ERROR_CODES = [
  'BAD_REQUEST',
  'VALIDATION_FAILED',
  'UNAUTHENTICATED',
  'FORBIDDEN',
  'NOT_FOUND',
  'CONFLICT',
  'RATE_LIMITED',
  'INTERNAL',
  'UNAVAILABLE',
] as const;

export type ApiErrorCode = (typeof API_ERROR_CODES)[number];

/** Narrowing helper for unknown response bodies. */
export function isApiErrorBody(value: unknown): value is ApiErrorBody {
  if (typeof value !== 'object' || value === null) return false;
  const error = (value as { error?: unknown }).error;
  if (typeof error !== 'object' || error === null) return false;
  const { code, message } = error as { code?: unknown; message?: unknown };
  return typeof code === 'string' && typeof message === 'string';
}
