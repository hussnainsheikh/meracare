import type { NotificationPreferences, ReminderPlan } from '@meracare/contracts';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { apiRequest } from '@/lib/api-client';

import { describeDevice } from './device';

/**
 * Notification settings and the reminder plan.
 *
 * All of it is server state, so all of it lives in TanStack Query. Zustand
 * holds nothing here: a preference the server has not accepted is not a
 * preference, and a plan cached in a store would go on scheduling reminders for
 * a senior the user no longer has access to (plans/phase8.md §35).
 */

export const notificationKeys = {
  preferences: ['notifications', 'preferences'] as const,
  reminders: ['notifications', 'reminders'] as const,
};

/** The caller's notification settings. */
export function useNotificationPreferences(enabled = true) {
  return useQuery({
    queryKey: notificationKeys.preferences,
    queryFn: () => apiRequest<NotificationPreferences>('/notifications/preferences'),
    enabled,
    // Settings change only when this user changes them, and the mutation below
    // writes the result straight into the cache.
    staleTime: 5 * 60_000,
  });
}

/**
 * Changes one or more categories.
 *
 * Deliberately not queued for offline replay. The offline queue exists for care
 * that was given — a dose taken, a task completed — where losing the record
 * would lose something that actually happened. A preference is a statement
 * about the future, and one applied optimistically offline would have the app
 * showing reminders as off while the server went on planning them
 * (plans/phase8.md §36).
 */
export function useUpdateNotificationPreferences() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (change: Partial<Omit<NotificationPreferences, 'updatedAt'>>) =>
      apiRequest<NotificationPreferences>('/notifications/preferences', {
        method: 'PATCH',
        body: change,
      }),
    onSuccess: (preferences) => {
      queryClient.setQueryData(notificationKeys.preferences, preferences);
      // A silenced category must disappear from the device, not just from the
      // screen, so the plan is refetched and reconciled (plans/phase8.md §22).
      void queryClient.invalidateQueries({ queryKey: notificationKeys.reminders });
    },
  });
}

/**
 * The reminders this device should have scheduled.
 *
 * Refetched rather than long-cached: the plan is how a revoked caregiver's
 * reminders stop, so a stale one is not merely out of date, it is a small
 * privacy problem (plans/phase8.md §23).
 */
export function useReminderPlan(enabled = true) {
  return useQuery({
    queryKey: notificationKeys.reminders,
    queryFn: () => apiRequest<ReminderPlan>('/notifications/reminders'),
    enabled,
    staleTime: 60_000,
  });
}

/**
 * Registers this installation with the server.
 *
 * Called on sign-in and after permission is granted, because both change what
 * the answer would be. Repeating it is free: the endpoint upserts on the device
 * identifier (plans/phase8.md §25).
 */
export function useRegisterDevice() {
  return useMutation({
    mutationFn: async () =>
      apiRequest<unknown>('/notifications/devices', {
        method: 'POST',
        body: await describeDevice(),
      }),
  });
}
