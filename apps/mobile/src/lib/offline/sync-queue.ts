/**
 * The offline mutation queue.
 *
 * Care recorded on a phone with no signal must not be lost, so a completion or
 * a skip is written locally and replayed when the connection returns
 * (docs/07-database-and-sync.md, "Sync Queue").
 *
 * This is deliberately not a general sync engine. It carries the two task
 * mutations Phase 4 needs, in the shape docs/07 specifies, so a later phase can
 * add an entity type without rebuilding the mechanism (plans/phase4.md §26).
 */

/** What kind of thing the operation acts on. */
export type SyncEntityType = 'task';

/** What the operation does. Both are idempotent on the server. */
export type SyncOperationType = 'complete' | 'skip';

/** Where an operation has got to. */
export type SyncStatus = 'pending' | 'failed';

/** One queued mutation, matching the record docs/07 describes. */
export interface SyncOperation {
  operationId: string;
  entityType: SyncEntityType;
  entityId: string;
  operationType: SyncOperationType;
  payload: string | null;
  createdAt: string;
  retryCount: number;
  lastError: string | null;
  status: SyncStatus;
}

/**
 * Storage for the queue.
 *
 * An interface rather than a direct SQLite dependency: the replay logic below
 * is the part that can be wrong in ways that lose somebody's care record, and
 * it needs to be testable without a native module.
 */
export interface SyncStore {
  enqueue(operation: SyncOperation): Promise<void>;
  pending(): Promise<SyncOperation[]>;
  remove(operationId: string): Promise<void>;
  markFailed(operationId: string, error: string, permanent: boolean): Promise<void>;
  /** Operations for one entity, used to reason about what is in flight. */
  forEntity(entityType: SyncEntityType, entityId: string): Promise<SyncOperation[]>;
}

/** How the server answered a replayed operation. */
export type ReplayOutcome =
  /** Recorded. The operation can be dropped. */
  | { kind: 'applied' }
  /** The connection failed. Keep it and try again later. */
  | { kind: 'transient'; message: string }
  /**
   * The server refused it and always will — the task now has a different
   * outcome, or the caller lost access. Retrying cannot help.
   */
  | { kind: 'permanent'; message: string };

/** Sends one operation to the server. */
export type Replay = (operation: SyncOperation) => Promise<ReplayOutcome>;

/** What one pass of the queue did. */
export interface SyncReport {
  applied: number;
  retrying: number;
  failed: SyncOperation[];
}

/**
 * How many times an operation is retried before it is set aside.
 *
 * Not unlimited: an operation that keeps failing needs to be surfaced to
 * somebody, not retried silently until the end of time.
 */
export const MAX_RETRIES = 5;

/**
 * Replays every queued mutation, oldest first.
 *
 * Order matters and is preserved: two actions on the same task must reach the
 * server in the order the user performed them, or the second could be judged
 * against a state the first has not produced yet.
 *
 * A transient failure stops the pass. If the connection has dropped, the
 * remaining operations will fail the same way, and hammering them only burns
 * retries that are meant to protect against a genuinely stuck operation.
 */
export async function processQueue(store: SyncStore, replay: Replay): Promise<SyncReport> {
  const operations = await store.pending();
  const report: SyncReport = { applied: 0, retrying: 0, failed: [] };

  for (const operation of operations) {
    const outcome = await replay(operation);

    if (outcome.kind === 'applied') {
      await store.remove(operation.operationId);
      report.applied += 1;
      continue;
    }

    if (outcome.kind === 'permanent') {
      await store.markFailed(operation.operationId, outcome.message, true);
      report.failed.push({ ...operation, lastError: outcome.message, status: 'failed' });
      continue;
    }

    // Transient. Give up on this pass and keep the queue for the next one.
    const exhausted = operation.retryCount + 1 >= MAX_RETRIES;
    await store.markFailed(operation.operationId, outcome.message, exhausted);

    if (exhausted) {
      report.failed.push({ ...operation, lastError: outcome.message, status: 'failed' });
      continue;
    }

    report.retrying += 1;
    break;
  }

  return report;
}

/** Builds a queued operation. */
export function newOperation(
  operationId: string,
  entityId: string,
  operationType: SyncOperationType,
  payload: Record<string, unknown> | null,
  createdAt: string,
): SyncOperation {
  return {
    operationId,
    entityType: 'task',
    entityId,
    operationType,
    payload: payload === null ? null : JSON.stringify(payload),
    createdAt,
    retryCount: 0,
    lastError: null,
    status: 'pending',
  };
}
