# Testing and Quality Strategy

## Testing Pyramid

### Unit Tests

Test:

-   recurrence calculations
-   authorization rules
-   task state transitions
-   medication schedule logic
-   escalation rules
-   data validation
-   sync conflict handling

### Integration Tests

Test Go API against a test PostgreSQL database.

Important flows:

-   create senior
-   invite member
-   accept invitation
-   create task
-   complete task
-   medication completion
-   appointment creation
-   unauthorized access

### Mobile Tests

Test:

-   authentication
-   onboarding
-   senior creation
-   care circle
-   task completion
-   medication completion
-   offline completion
-   notification navigation

### Web E2E

Use Playwright for any web/admin interface.

## Critical E2E Scenarios

1.  Solo user creates profile.
2.  Family creates senior.
3.  Family invites caregiver.
4.  Caregiver accepts invitation.
5.  Caregiver sees assigned seniors.
6.  Caregiver completes task.
7.  Family sees completion.
8.  Medication is completed.
9.  Missed task escalates.
10. Unauthorized user cannot access senior.

## Performance Testing

Track:

-   cold start
-   warm start
-   screen transition
-   list scrolling
-   task completion latency
-   API p50/p95 latency
-   sync duration
-   memory usage
-   battery impact of background features

Test on:

-   modern iPhone
-   older supported iPhone
-   mid-range Android
-   lower-end Android

## Quality Gates

A feature is not complete until:

-   happy path works
-   unauthorized path is tested
-   loading state exists
-   empty state exists
-   error state exists
-   offline behavior is defined
-   analytics/event behavior is defined if applicable
-   accessibility is considered
