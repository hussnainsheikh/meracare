import { ApiError } from '@/lib/api-error';

import type { ReplayOutcome } from './sync-queue';

/**
 * Deciding whether a failed replay is worth trying again.
 *
 * This distinction is the whole point of the queue, and it is the same
 * judgement whatever kind of care the operation records: retrying a lost
 * connection is right; retrying a dose somebody else has already skipped would
 * never succeed, and would keep a stale action alive indefinitely.
 */
export function classify(error: unknown): ReplayOutcome {
  if (!(error instanceof ApiError)) {
    return { kind: 'transient', message: 'Something went wrong. We will try again.' };
  }

  if (error.isOffline) {
    return { kind: 'transient', message: error.message };
  }

  // 409: the record already has a different outcome, and the server is
  // authoritative. 404: it is gone, or access was revoked while the action sat
  // in the queue. Neither improves with time.
  if (error.status === 409 || error.status === 404 || error.isForbidden) {
    return { kind: 'permanent', message: error.message };
  }

  // A 5xx is the server's problem, not the request's.
  if (error.isRetryable) {
    return { kind: 'transient', message: error.message };
  }

  // Anything else — a rejected body, an expired session — will not fix itself
  // by being sent again unchanged.
  return { kind: 'permanent', message: error.message };
}

/** Reads the optional note a queued operation carries. */
export function readNotes(payload: string | null): string | null {
  if (payload === null) return null;

  try {
    const parsed = JSON.parse(payload) as { notes?: string };
    return typeof parsed.notes === 'string' ? parsed.notes : null;
  } catch {
    return null;
  }
}
