import { newOperation } from '@/lib/offline/sync-queue';

import { replayOperation } from '../replay';

/**
 * One queue, drained in the order the user acted, whatever kind of care each
 * entry records (plans/phase5.md §20). This pins the routing: a task must not
 * be sent to a medication route, and a dose must not be sent to a task route.
 */

const mockReplayTask = jest.fn(async () => ({ kind: 'applied' }) as const);
const mockReplayMedication = jest.fn(async () => ({ kind: 'applied' }) as const);

jest.mock('@/features/tasks/task-sync', () => ({
  replayTaskOperation: (...args: unknown[]) => mockReplayTask(...(args as [])),
}));

jest.mock('@/features/medications/medication-sync', () => ({
  replayMedicationOperation: (...args: unknown[]) => mockReplayMedication(...(args as [])),
}));

jest.mock('@/lib/offline/database', () => ({ sqliteSyncStore: {} }));

beforeEach(() => {
  mockReplayTask.mockClear();
  mockReplayMedication.mockClear();
});

describe('replayOperation', () => {
  it('sends a queued task completion to the task route', async () => {
    const operation = newOperation(
      'op-1',
      'task',
      'task-1',
      'complete',
      null,
      '2026-08-17T09:00:00Z',
    );

    await replayOperation(operation);

    expect(mockReplayTask).toHaveBeenCalledWith(operation);
    expect(mockReplayMedication).not.toHaveBeenCalled();
  });

  it('sends a queued dose to the medication route', async () => {
    const operation = newOperation(
      'op-2',
      'medication',
      'med-1/dose-1',
      'take',
      null,
      '2026-08-17T09:01:00Z',
    );

    await replayOperation(operation);

    expect(mockReplayMedication).toHaveBeenCalledWith(operation);
    expect(mockReplayTask).not.toHaveBeenCalled();
  });
});
