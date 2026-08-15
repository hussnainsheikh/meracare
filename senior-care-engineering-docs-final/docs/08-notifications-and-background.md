# Notifications and Background Processing

## Principle

Do not keep JavaScript running continuously just to check time.

Use OS-native scheduling and push notifications.

## Local Notifications

Use for:

-   medication reminders
-   scheduled care tasks
-   appointments

The schedule should be synced to the device.

Example:

``` text
Server schedule
  ↓
Mobile sync
  ↓
OS local notification
  ↓
User action
```

## Remote Push

Use push notifications for:

-   missed tasks
-   caregiver activity
-   invitations
-   messages
-   relevant care-circle events

Architecture:

``` text
Domain event
  ↓
Notification decision
  ↓
Push provider
  ↓
APNs / FCM
  ↓
User device
```

## Background Work

Use background execution only when justified.

Potential MVP jobs:

-   sync pending mutations
-   refresh relevant cached data
-   reconcile notification schedules

Do not run high-frequency timers.

Avoid:

``` text
setInterval(..., 1000)
```

for care logic.

## Escalation

MVP should support configurable escalation.

Example:

``` text
Task due
  ↓
Reminder
  ↓
Grace period
  ↓
Overdue
  ↓
Notify assigned caregiver
  ↓
Optional family notification
```

The escalation system should not imply medical emergency unless a
separately implemented emergency feature exists.

## Notification Preferences

Users should control:

-   task reminders
-   medication reminders
-   activity notifications
-   messages
-   invitations
-   escalation alerts

Respect OS notification permissions.
