import { replayMedicationOperation } from '@/features/medications/medication-sync';
import { replayTaskOperation } from '@/features/tasks/task-sync';
import { sqliteSyncStore } from '@/lib/offline/database';
import {
  processQueue,
  type ReplayOutcome,
  type SyncOperation,
  type SyncReport,
} from '@/lib/offline/sync-queue';

/**
 * Sending everything that was recorded offline.
 *
 * One queue, drained in the order the user acted, whatever kind of care each
 * entry records. Two queues would replay a task completion and a dose taken a
 * minute earlier in whichever order the passes happened to run
 * (plans/phase5.md §20).
 */

/** Routes one queued operation to the endpoint that records it. */
export function replayOperation(operation: SyncOperation): Promise<ReplayOutcome> {
  switch (operation.entityType) {
    case 'task':
      return replayTaskOperation(operation);
    case 'medication':
      return replayMedicationOperation(operation);
  }
}

/** Sends everything queued while offline. */
export function syncQueuedOperations(): Promise<SyncReport> {
  return processQueue(sqliteSyncStore, replayOperation);
}
