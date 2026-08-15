# Agent Development Guide

## Visual System

Use the locked visual source of truth in
`18-visual-theme-and-illustrations.md`.

-   Primary direction: green with a slight blue/teal bias.
-   Primary brand color: `#0F766E` Deep Teal.
-   Illustration sources: unDraw (primary) and Storyset (secondary).
-   Do not introduce a new primary color or illustration library for MVP
    without explicit approval.

## Mission

Build the Senior Care MVP according to the product and engineering
specifications in `/docs`.

## Architecture Status

The following decisions are locked for MVP:

-   React Native + Expo
-   TypeScript
-   Expo Router
-   TanStack Query
-   small Zustand
-   native React Native styling
-   `expo-sqlite`
-   Go backend
-   REST API
-   PostgreSQL hosted on Supabase
-   Supabase Auth
-   Supabase Storage
-   pnpm workspace
-   modular monolith backend

Do not replace these technologies during MVP unless a concrete blocker
is demonstrated and explicitly approved.

## Source of Truth

Read:

1.  `00-product-overview.md`
2.  `01-roles-and-care-model.md`
3.  `02-permissions-and-authorization.md`
4.  `03-domain-model.md`
5.  `04-care-events-and-workflows.md`
6.  `05-api-and-backend-spec.md`
7.  `06-mobile-architecture.md`
8.  `07-database-and-sync.md`
9.  `08-notifications-and-background.md`
10. `09-security-privacy.md`
11. `10-testing-and-quality.md`
12. `11-performance-requirements.md`
13. `12-tech-stack.md`
14. `13-mvp-screen-map.md`
15. `14-mvp-roadmap-and-tasks.md`
16. `15-v2-v3-v4-roadmap.md`

## Product Modes

The MVP must support all four modes using the same application:

1.  Solo self-care.
2.  Family care.
3.  Professional caregiver.
4.  Mixed family + professional care.

A user may have different relationships with different seniors.

## Implementation Rules

-   Do not create separate family and professional applications.
-   Do not require a caregiver for a senior.
-   Do not introduce microservices.
-   Do not introduce Redux.
-   Do not introduce a styling framework.
-   Do not duplicate server state into Zustand.
-   Do not bypass Go API authorization.
-   Do not put business-critical database writes directly in the mobile
    client.
-   Do not implement continuous polling for timers/reminders.
-   Do not introduce V2/V3 features before MVP is complete unless
    explicitly requested.
-   Do not make medical diagnoses or unsupported medical claims.

## Repository Rules

Use a simple pnpm workspace.

Recommended:

``` text
apps/
  mobile/
  api/

packages/
  contracts/
  config/
```

The Go backend remains independently typed and compiled.

Do not create artificial shared packages between Go and TypeScript
merely to avoid duplication.

Use OpenAPI as the API contract when shared generated types become
useful.

## Feature Workflow

For every feature:

1.  Read the relevant specification.
2.  Inspect existing code.
3.  Identify dependencies.
4.  Implement the smallest correct change.
5.  Add/update tests.
6.  Run lint/type checks/tests.
7.  Verify mobile/backend behavior.
8.  Update documentation if architecture changes.
9.  Do not silently change locked decisions.

## Feature Completion Checklist

-   [ ] Data model defined.
-   [ ] API defined.
-   [ ] Authorization defined.
-   [ ] UI implemented.
-   [ ] Loading state.
-   [ ] Empty state.
-   [ ] Error state.
-   [ ] Offline behavior considered.
-   [ ] Notifications considered.
-   [ ] Care event generated.
-   [ ] Tests added.
-   [ ] Performance considered.
-   [ ] Accessibility considered.

## First Implementation Order

1.  Repository setup.
2.  Supabase project/Auth.
3.  PostgreSQL schema/migrations.
4.  Go API foundation.
5.  React Native/Expo foundation.
6.  User/session flow.
7.  Senior + CareRelationship.
8.  Invitations.
9.  Tasks.
10. Medication.
11. Appointments.
12. Care events.
13. Notifications.
14. Family dashboard.
15. Professional caregiver dashboard.
16. Solo mode verification.
17. Care-circle messaging.
18. Offline/sync hardening.
19. Automated tests.
20. MVP analytics/validation instrumentation.
