import {
  MAX_RETRIES,
  newOperation,
  processQueue,
  type ReplayOutcome,
  type SyncOperation,
  type SyncStore,
} from '../sync-queue';

/**
 * The offline queue holds care that has already happened — somebody completed a
 * task, and the phone had no signal. Losing an entry loses the record of that
 * care, so these tests pin what the queue does with each kind of failure.
 */

/** An in-memory store, so the logic is testable without the native module. */
function newStore(initial: SyncOperation[] = []): SyncStore & { all: SyncOperation[] } {
  const all = [...initial];

  return {
    all,
    async enqueue(operation) {
      const existing = all.findIndex((entry) => entry.operationId === operation.operationId);
      if (existing >= 0) all[existing] = operation;
      else all.push(operation);
    },
    async pending() {
      return all.filter((entry) => entry.status === 'pending');
    },
    async remove(operationId) {
      const index = all.findIndex((entry) => entry.operationId === operationId);
      if (index >= 0) all.splice(index, 1);
    },
    async markFailed(operationId, error, permanent) {
      const entry = all.find((item) => item.operationId === operationId);
      if (!entry) return;
      entry.retryCount += 1;
      entry.lastError = error;
      entry.status = permanent ? 'failed' : 'pending';
    },
    async forEntity(entityType, entityId) {
      return all.filter((entry) => entry.entityType === entityType && entry.entityId === entityId);
    },
  };
}

function operation(id: string, taskId = 'task-1'): SyncOperation {
  return newOperation(id, taskId, 'complete', null, `2026-08-17T09:0${id.slice(-1)}:00Z`);
}

const applied = async (): Promise<ReplayOutcome> => ({ kind: 'applied' });

describe('processQueue', () => {
  it('sends every queued mutation and clears it', async () => {
    const store = newStore([operation('op-1'), operation('op-2', 'task-2')]);

    const report = await processQueue(store, applied);

    expect(report.applied).toBe(2);
    expect(store.all).toHaveLength(0);
  });

  it('does nothing when the queue is empty', async () => {
    const store = newStore();

    const report = await processQueue(store, applied);

    expect(report).toEqual({ applied: 0, retrying: 0, failed: [] });
  });

  // Order is the user's order. Two actions on one task must arrive as they were
  // performed, or the second is judged against a state the first has not made.
  it('replays in the order the actions were taken', async () => {
    const store = newStore([operation('op-1'), operation('op-2'), operation('op-3')]);
    const seen: string[] = [];

    await processQueue(store, async (op) => {
      seen.push(op.operationId);
      return { kind: 'applied' };
    });

    expect(seen).toEqual(['op-1', 'op-2', 'op-3']);
  });

  // A lost connection is temporary. The entry stays, ready for the next pass.
  it('keeps an operation whose connection failed', async () => {
    const store = newStore([operation('op-1')]);

    const report = await processQueue(store, async () => ({
      kind: 'transient',
      message: 'offline',
    }));

    expect(report.applied).toBe(0);
    expect(report.retrying).toBe(1);
    expect(store.all).toHaveLength(1);
    expect(store.all[0]?.status).toBe('pending');
    expect(store.all[0]?.retryCount).toBe(1);
  });

  // If the connection has dropped, the rest will fail identically. Trying them
  // all would burn the retries that protect a genuinely stuck operation.
  it('stops the pass at the first connection failure', async () => {
    const store = newStore([operation('op-1'), operation('op-2'), operation('op-3')]);
    let attempts = 0;

    await processQueue(store, async () => {
      attempts += 1;
      return { kind: 'transient', message: 'offline' };
    });

    expect(attempts).toBe(1);
    expect(store.all).toHaveLength(3);
  });

  // A refusal will not improve with time; retrying it forever would keep a
  // stale action alive and hide the problem from the user.
  it('sets aside an operation the server refused outright', async () => {
    const store = newStore([operation('op-1'), operation('op-2')]);

    const report = await processQueue(store, async (op) =>
      op.operationId === 'op-1'
        ? { kind: 'permanent', message: 'Already skipped.' }
        : { kind: 'applied' },
    );

    expect(report.failed).toHaveLength(1);
    expect(report.failed[0]?.operationId).toBe('op-1');
    expect(report.applied).toBe(1);

    // It is kept, marked failed, rather than deleted: somebody has to be told.
    expect(store.all).toHaveLength(1);
    expect(store.all[0]?.status).toBe('failed');
    expect(store.all[0]?.lastError).toBe('Already skipped.');
  });

  it('gives up on an operation that keeps failing', async () => {
    const stubborn = { ...operation('op-1'), retryCount: MAX_RETRIES - 1 };
    const store = newStore([stubborn]);

    const report = await processQueue(store, async () => ({
      kind: 'transient',
      message: 'still offline',
    }));

    expect(report.retrying).toBe(0);
    expect(report.failed).toHaveLength(1);
    expect(store.all[0]?.status).toBe('failed');
  });

  it('ignores operations already set aside', async () => {
    const store = newStore([{ ...operation('op-1'), status: 'failed' }]);
    let attempts = 0;

    await processQueue(store, async () => {
      attempts += 1;
      return { kind: 'applied' };
    });

    expect(attempts).toBe(0);
  });
});

describe('newOperation', () => {
  it('starts pending, unretried, and carries its payload as JSON', () => {
    const created = newOperation(
      'op-1',
      'task-1',
      'skip',
      { notes: 'She was asleep' },
      '2026-08-17T09:00:00Z',
    );

    expect(created).toMatchObject({
      operationId: 'op-1',
      entityType: 'task',
      entityId: 'task-1',
      operationType: 'skip',
      retryCount: 0,
      lastError: null,
      status: 'pending',
    });
    expect(JSON.parse(created.payload as string)).toEqual({ notes: 'She was asleep' });
  });

  it('carries no payload when there is nothing to say', () => {
    expect(newOperation('op-1', 'task-1', 'complete', null, '2026-08-17T09:00:00Z').payload).toBe(
      null,
    );
  });
});

describe('the queue survives a duplicated enqueue', () => {
  // The same user action must never be queued twice, whatever retries the UI
  // performs, because two completions of one task is a nonsense record.
  it('replaces an operation with the same id', async () => {
    const store = newStore();

    await store.enqueue(operation('op-1'));
    await store.enqueue(operation('op-1'));

    expect(await store.pending()).toHaveLength(1);
  });
});
