import { CARE_PERMISSIONS, type CarePermission, type CareRole } from './care';

/**
 * Human-readable descriptions of permissions.
 *
 * Raw identifiers like `medications.record` are for the API, not for a person
 * deciding what their mother's caregiver should be able to do. Every permission
 * shown in the app is described in plain language (docs/18 asks for simple
 * language and clear permission descriptions).
 */
export interface PermissionLabel {
  permission: CarePermission;
  /** Short phrase for a checkbox or list row. */
  label: string;
  /** One line explaining what it actually allows. */
  description: string;
  /** Grouping for the invite screen. */
  group: PermissionGroup;
}

export const PERMISSION_GROUPS = [
  'Care information',
  'Daily care',
  'Medications',
  'Appointments',
  'Notes and messages',
  'People',
] as const;
export type PermissionGroup = (typeof PERMISSION_GROUPS)[number];

const LABELS: Record<CarePermission, Omit<PermissionLabel, 'permission'>> = {
  'senior.view': {
    label: 'View care information',
    description: 'See the profile and overall care picture.',
    group: 'Care information',
  },
  'senior.edit': {
    label: 'Edit the profile',
    description: 'Change name, date of birth, contact and address details.',
    group: 'Care information',
  },
  'tasks.view': {
    label: 'View tasks',
    description: 'See what needs doing and what has been done.',
    group: 'Daily care',
  },
  'tasks.manage': {
    label: 'Create and change tasks',
    description: 'Add tasks, set when they repeat, and assign them to people.',
    group: 'Daily care',
  },
  'tasks.complete': {
    label: 'Complete tasks',
    description: 'Mark tasks as done or skipped.',
    group: 'Daily care',
  },
  'medications.view': {
    label: 'View medications',
    description: 'See the medication list and whether doses were taken.',
    group: 'Medications',
  },
  'medications.manage': {
    label: 'Create and change medications',
    description: 'Add medications and set their schedules.',
    group: 'Medications',
  },
  'medications.record': {
    label: 'Record medications',
    description: 'Mark a dose as taken or missed.',
    group: 'Medications',
  },
  'appointments.view': {
    label: 'View appointments',
    description: 'See upcoming and past appointments.',
    group: 'Appointments',
  },
  'appointments.manage': {
    label: 'Create and change appointments',
    description: 'Add appointments and assign who is going.',
    group: 'Appointments',
  },
  'notes.view': {
    label: 'View care notes',
    description: 'Read notes written by the care circle.',
    group: 'Notes and messages',
  },
  'notes.create': {
    label: 'Write care notes',
    description: 'Add notes about how things are going.',
    group: 'Notes and messages',
  },
  'activity.view': {
    label: 'View activity',
    description: 'See a history of what the care circle has done.',
    group: 'Notes and messages',
  },
  'messages.participate': {
    label: 'Join the conversation',
    description: 'Send and read messages with the care circle.',
    group: 'Notes and messages',
  },
  'members.view': {
    label: 'See who is involved',
    description: 'View the list of family and caregivers.',
    group: 'People',
  },
  'members.invite': {
    label: 'Invite people',
    description: 'Invite family members and caregivers into the circle.',
    group: 'People',
  },
  'members.manage': {
    label: 'Manage people',
    description: 'Change what others can do, and remove them from the circle.',
    group: 'People',
  },
};

/** Describes one permission in plain language. */
export function permissionLabel(permission: CarePermission): PermissionLabel {
  return { permission, ...LABELS[permission] };
}

/** Every permission, described, in a stable order suitable for a form. */
export function permissionLabels(): PermissionLabel[] {
  return CARE_PERMISSIONS.map(permissionLabel);
}

/** Permissions grouped for display, preserving group order. */
export function permissionLabelsByGroup(
  permissions: CarePermission[] = [...CARE_PERMISSIONS],
): { group: PermissionGroup; permissions: PermissionLabel[] }[] {
  const described = permissions.map(permissionLabel);

  return PERMISSION_GROUPS.map((group) => ({
    group,
    permissions: described.filter((entry) => entry.group === group),
  })).filter((section) => section.permissions.length > 0);
}

/** Plain-language name for a role. */
export function roleLabel(role: CareRole): string {
  switch (role) {
    case 'senior':
      return 'The person being cared for';
    case 'family_member':
      return 'Family member';
    case 'professional_caregiver':
      return 'Professional caregiver';
    default:
      return 'Care circle member';
  }
}
