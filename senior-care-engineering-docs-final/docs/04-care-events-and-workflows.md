# Care Events and Core Workflows

## Care Event Architecture

Every meaningful care action should produce a `CareEvent`.

Flow:

``` text
User action
   ↓
Domain command
   ↓
Database transaction
   ↓
CareEvent
   ↓
Notifications / activity / analytics
```

Do not make the activity timeline depend on UI state.

## Task Completion

1.  User opens task.
2.  Client immediately updates UI optimistically.
3.  Client sends completion request.
4.  Backend authorizes user.
5.  Backend updates task instance.
6.  Backend creates `TASK_COMPLETED`.
7.  Backend queues notifications if required.
8.  Other clients receive/update data.

## Medication Completion

Same pattern:

``` text
Medication reminder
→ user confirms
→ optimistic local update
→ server mutation
→ MEDICATION_TAKEN event
→ care-circle visibility
```

## Missed Task Escalation

MVP rules should be simple and configurable.

Example:

-   Due at 08:00.
-   Reminder before/during due time.
-   If not completed after configured grace period, mark as overdue.
-   Notify assigned caregiver.
-   Optionally notify selected family member.

Avoid aggressive medical/emergency language.

Use:

> "Mom's morning medication has not been marked complete."

Not:

> "Mom is in danger."

## Appointment Workflow

1.  Create appointment.
2.  Assign caregiver.
3.  Notify assigned person.
4.  Display on senior dashboard.
5.  Allow completion/cancellation.
6.  Record event.

## Invitation Workflow

1.  Authorized member creates invitation.
2.  Backend creates invitation.
3.  Notification/email is sent.
4.  Invitee signs in or creates account.
5.  Invitee accepts.
6.  Care relationship becomes active.
7.  `MEMBER_JOINED` event is created.

## Solo Workflow

1.  User creates their own senior profile.
2.  Relationship to self is created with role `senior`.
3.  No invitation is required.
4.  User can create tasks, medication, and appointments.
5.  User may invite others later.
