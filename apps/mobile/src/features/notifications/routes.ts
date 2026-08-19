import type { ReminderPayload } from '@meracare/contracts';
import type { Href } from 'expo-router';

/**
 * Where tapping a notification should take the user.
 *
 * The server does not decide this, and should not: routes are a property of the
 * app, and a server that named them would have to be redeployed to move a
 * screen. It sends what the reminder is about; this turns that into a
 * destination in the app's existing navigation (plans/phase8.md §15).
 */
export function reminderDestination(payload: ReminderPayload): Href {
  switch (payload.entityType) {
    case 'task_instance':
      return { pathname: '/tasks/[taskId]', params: { taskId: payload.entityId } };

    case 'appointment':
      return {
        pathname: '/appointments/[appointmentId]',
        params: { appointmentId: payload.entityId },
      };

    case 'medication_dose':
      // A single dose has no screen of its own, and giving it one for the sake
      // of a notification would be a screen nobody navigates to any other way.
      // Today's medication is where the dose is, in the context of the rest of
      // the day's medicine.
      return {
        pathname: '/seniors/[seniorId]/medications',
        params: { seniorId: payload.seniorId },
      };
  }
}
