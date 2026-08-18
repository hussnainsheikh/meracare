import type { ReminderPayload } from '@meracare/contracts';

import { reminderDestination } from '../routes';

/**
 * Every notification has somewhere to go. A tap that lands on the wrong screen
 * — or nowhere — makes the reminder useless at exactly the moment it mattered.
 */

function payload(overrides: Partial<ReminderPayload> = {}): ReminderPayload {
  return {
    reminderId: 'reminder-1',
    type: 'MEDICATION_REMINDER',
    seniorId: 'senior-1',
    entityType: 'medication_dose',
    entityId: 'dose-1',
    ...overrides,
  };
}

it('opens today’s medication for a dose', () => {
  expect(reminderDestination(payload())).toEqual({
    pathname: '/seniors/[seniorId]/medications',
    params: { seniorId: 'senior-1' },
  });
});

it('opens the task for a task reminder', () => {
  const destination = reminderDestination(
    payload({ type: 'TASK_REMINDER', entityType: 'task_instance', entityId: 'task-9' }),
  );

  expect(destination).toEqual({
    pathname: '/tasks/[taskId]',
    params: { taskId: 'task-9' },
  });
});

it('opens the appointment for an appointment reminder', () => {
  const destination = reminderDestination(
    payload({
      type: 'APPOINTMENT_REMINDER',
      entityType: 'appointment',
      entityId: 'appointment-3',
    }),
  );

  expect(destination).toEqual({
    pathname: '/appointments/[appointmentId]',
    params: { appointmentId: 'appointment-3' },
  });
});
