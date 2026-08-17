import type { CareTask } from '@meracare/contracts';

import { apiRequest } from '@/lib/api-client';
import { ApiError } from '@/lib/api-error';
import type { ReplayOutcome, SyncOperation } from '@/lib/offline/sync-queue';

/**
 * Sending a queued task mutation back to the server.
 *
 * Kept apart from the queue mechanism so the decision that matters — which
 * failures are worth retrying and which are final — is testable on its own.
 */

/**
 * Replays one queued completion or skip.
 *
 * The operation id travels as the idempotency key, so a mutation that reached
 * the server before the connection dropped is recognised rather than applied
 * twice (plans/phase4.md §27).
 */
export async function replayTaskOperation(operation: SyncOperation): Promise<ReplayOutcome> {
  const notes = readNotes(operation.payload);

  try {
    await apiRequest<CareTask>(`/tasks/${operation.entityId}/${operation.operationType}`, {
      method: 'POST',
      body: notes === null ? undefined : { notes },
      idempotencyKey: operation.operationId,
    });
    return { kind: 'applied' };
  } catch (error) {
    return classify(error);
  }
}

/**
 * Decides whether a failure is worth trying again.
 *
 * The distinction is the whole point of the queue. Retrying a lost connection
 * is right; retrying a task somebody else has already skipped would never
 * succeed, and would keep a stale action alive indefinitely.
 */
export function classify(error: unknown): ReplayOutcome {
  if (!(error instanceof ApiError)) {
    return { kind: 'transient', message: 'Something went wrong. We will try again.' };
  }

  if (error.isOffline) {
    return { kind: 'transient', message: error.message };
  }

  // 409: the task already has a different outcome, and the server is
  // authoritative. 404: the task is gone, or access was revoked while the
  // action sat in the queue. Neither improves with time.
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

function readNotes(payload: string | null): string | null {
  if (payload === null) return null;

  try {
    const parsed = JSON.parse(payload) as { notes?: string };
    return typeof parsed.notes === 'string' ? parsed.notes : null;
  } catch {
    return null;
  }
}
