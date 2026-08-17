import { RECURRENCE_FREQUENCIES, type Recurrence, type RecurrenceFrequency } from './recurrence';

/**
 * Care tasks: the daily routine a circle carries out.
 *
 * Mirrors internal/tasks in the Go API. The distinction the domain draws
 * (docs/03-domain-model.md) is between a *template* — the rule, "every weekday
 * at 09:00" — and an *instance*, one concrete occurrence somebody completes or
 * skips. A one-time task is an instance with no template.
 */

/**
 * The state of one occurrence.
 *
 * `overdue` is derived by the server from the clock, never stored: a task is
 * overdue exactly when its due time has passed and nobody has acted on it. The
 * client can therefore render this value directly.
 */
export const TASK_STATUSES = ['pending', 'overdue', 'completed', 'skipped', 'cancelled'] as const;
export type TaskStatus = (typeof TASK_STATUSES)[number];

/** Statuses that mean somebody has already decided the outcome. */
export const SETTLED_TASK_STATUSES = ['completed', 'skipped', 'cancelled'] as const;

/** Reports whether a task still needs doing. */
export function isOpen(task: Pick<CareTask, 'status'>): boolean {
  return task.status === 'pending' || task.status === 'overdue';
}

/** Reports whether the task is past due and still waiting. */
export function isOverdue(task: Pick<CareTask, 'status'>): boolean {
  return task.status === 'overdue';
}

/** One occurrence of a care task. */
export interface CareTask {
  id: string;
  /** Null for a one-time task. */
  templateId: string | null;
  seniorId: string;

  title: string;
  description: string | null;

  /** Null when the task belongs to whoever is available. */
  assignedUserId: string | null;

  /** ISO-8601 instant. Render it in the senior's timezone, not the device's. */
  scheduledFor: string;
  status: TaskStatus;
  /** True when this came from a repeating rule. */
  recurring: boolean;

  completedAt: string | null;
  completedBy: string | null;
  skippedAt: string | null;
  skippedBy: string | null;

  notes: string | null;

  createdAt: string;
  updatedAt: string;
}

/**
 * How often a task repeats.
 *
 * The rule grammar is shared with medication schedules — see `recurrence.ts`.
 * These names stay because a task screen talks about tasks, but they are the
 * same type, not a parallel one.
 */
export type TaskFrequency = RecurrenceFrequency;
export const TASK_FREQUENCIES = RECURRENCE_FREQUENCIES;
export type TaskRecurrence = Recurrence;

/** The definition of a recurring care task. */
export interface CareTaskTemplate {
  id: string;
  seniorId: string;
  title: string;
  description: string | null;
  assignedUserId: string | null;
  recurrence: TaskRecurrence;
  /** Wall-clock `HH:MM` in the senior's timezone, not an instant. */
  dueTime: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

/** Which set of tasks to fetch. */
export const TASK_SCOPES = ['today', 'upcoming', 'overdue', 'window'] as const;
export type TaskScope = (typeof TASK_SCOPES)[number];

/** `POST /v1/seniors/{id}/tasks` request body. */
export interface CreateTaskRequest {
  title: string;
  description?: string | null;
  assignedUserId?: string | null;
  /** A one-time task: an ISO-8601 instant. */
  scheduledFor?: string;
  /** A recurring task: the rule and the time of day it falls due. */
  recurrence?: TaskRecurrence;
  dueTime?: string;
}

/** `POST /v1/seniors/{id}/tasks` response. */
export interface CreateTaskResponse {
  tasks: CareTask[];
  /** Present only for a recurring task. */
  template?: CareTaskTemplate;
}

/** `PATCH /v1/tasks/{id}` request body. Absent fields are left unchanged. */
export interface UpdateTaskRequest {
  title?: string;
  description?: string | null;
  scheduledFor?: string;
  assignedUserId?: string | null;
  /** Sent with a null assignee to unassign, which absence alone cannot express. */
  clearAssignee?: boolean;
}

/** `PATCH /v1/seniors/{id}/tasks/templates/{id}` request body. */
export interface UpdateTaskTemplateRequest {
  title?: string;
  description?: string | null;
  recurrence?: TaskRecurrence;
  dueTime?: string;
  /** Set false to retire the routine; its history is kept. */
  active?: boolean;
  assignedUserId?: string | null;
  clearAssignee?: boolean;
}

/** An optional note recorded with a completion or a skip. */
export interface TaskActionRequest {
  notes?: string;
}

/** `GET /v1/seniors/{id}/tasks` and `GET /v1/tasks` response. */
export interface TaskListResponse {
  items: CareTask[];
}

/** `GET /v1/seniors/{id}/tasks/templates` response. */
export interface TaskTemplateListResponse {
  items: CareTaskTemplate[];
}
