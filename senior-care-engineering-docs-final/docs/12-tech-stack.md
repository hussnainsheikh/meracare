# Senior Care --- Final Technology Stack

## Visual System

Use the locked visual source of truth in
`18-visual-theme-and-illustrations.md`.

-   Primary direction: green with a slight blue/teal bias.
-   Primary brand color: `#0F766E` Deep Teal.
-   Illustration sources: unDraw (primary) and Storyset (secondary).
-   Do not introduce a new primary color or illustration library for MVP
    without explicit approval.

## Architecture Status

**LOCKED FOR MVP**

The following stack is the baseline implementation decision. It is
deliberately conservative and should not be replaced during MVP unless a
concrete technical blocker is demonstrated.

## Mobile

-   React Native
-   Expo
-   TypeScript
-   Expo Router
-   TanStack Query
-   Zustand
-   React Native `StyleSheet` / native styling
-   `expo-sqlite`
-   `expo-secure-store`
-   Expo Notifications initially

## Backend

-   Go
-   REST API
-   PostgreSQL
-   Supabase-hosted PostgreSQL
-   Supabase Auth
-   Supabase Storage
-   Background jobs in the Go backend
-   OpenAPI

## Web

The web application is **not required for the initial MVP**.

When introduced, use:

-   Next.js
-   TypeScript
-   same Go API
-   same Supabase Auth project

This allows the same identity to work across mobile and web.

## Repository / Tooling

Use:

-   pnpm workspaces
-   TypeScript
-   Go modules
-   GitHub
-   GitHub Actions
-   Docker for local/integration infrastructure

For the MVP, do not introduce Nx solely for orchestration. A simple pnpm
workspace is sufficient. If the repository later becomes large enough to
need task graph/caching tooling, it can be introduced without changing
the application architecture.

## State Management

### TanStack Query

The only default source for remote/server state in the mobile app.

Use it for:

-   queries
-   caching
-   mutations
-   invalidation
-   synchronization

### Zustand

Use only for local UI/application state.

Examples:

-   selected senior
-   filters
-   temporary form/UI state
-   onboarding state
-   preferences

Do not use Zustand as a server-state cache.

## Styling

Use native React Native styling.

No UI styling framework is required.

Build a small internal design system so that visual consistency does not
depend on an external styling abstraction.

## Database

PostgreSQL is the locked primary database.

It is the correct choice because the domain is relational:

-   users
-   seniors
-   care relationships
-   permissions
-   invitations
-   tasks
-   schedules
-   medications
-   appointments
-   notes
-   events
-   notifications
-   messages

## Supabase

Supabase is the managed infrastructure layer around PostgreSQL and Auth.

Use it for:

-   PostgreSQL hosting
-   Auth
-   Apple/Google social login
-   storage
-   database tooling

The Go API owns application business logic.

The mobile application should not directly perform business-critical
database operations through Supabase's generated data APIs.

## Authentication

Use one Supabase Auth project for:

-   iOS
-   Android
-   future web

Initial providers:

-   Apple
-   Google
-   email/password or email OTP as selected during onboarding

The application user record is linked to the Supabase Auth user ID.

## Why React Native

The MVP workload consists primarily of:

-   forms
-   lists
-   dashboards
-   navigation
-   notifications
-   API calls
-   local persistence
-   messaging
-   calendars
-   moderate health data

It does not require:

-   continuous high-frequency graphics
-   game rendering
-   advanced video processing
-   continuous DSP
-   large-scale on-device computer vision

Modern React Native is therefore an appropriate native-quality
cross-platform choice.

React Native 0.86 was released in June 2026, and Expo SDK 57 targets
React Native 0.86. We should use the stable Expo SDK/RN pairing
available when the repository is initialized rather than manually mixing
versions. citeturn0search1turn0search3

## Why Expo

Expo accelerates development and provides a mature path to native
capabilities.

We will use development builds/EAS/native configuration whenever the
application needs native functionality.

Expo is not a limitation on the architecture.

## Why Go

Go is the locked backend language because it provides:

-   low runtime overhead
-   strong concurrency
-   simple deployment
-   excellent HTTP support
-   strong typing
-   straightforward background processing
-   a good fit for this team's expertise

The backend is a modular monolith for MVP.

Do not create microservices.

## Why PostgreSQL

PostgreSQL gives:

-   transactions
-   constraints
-   indexes
-   relational integrity
-   mature tooling
-   good reporting/analytics foundations
-   straightforward future scaling

## Why Supabase

Supabase removes unnecessary infrastructure work for:

-   database hosting
-   authentication
-   social login
-   storage

Supabase officially supports React Native/Expo authentication and social
authentication flows, including Apple and Google.
citeturn0search0turn0search2

## Why TanStack Query + Zustand

This combination gives a clean separation:

``` text
Server state → TanStack Query
Local/UI state → Zustand
Durable/offline data → SQLite
```

There is no need for Redux or another global state framework.

## Why Native Styling

The app's UI is not dependent on a styling runtime or utility framework.

Native styling gives:

-   small dependency surface
-   predictable rendering
-   direct platform control
-   straightforward accessibility
-   easy future native customization

## Cross-Platform Decision

### React Native --- SELECTED

Best fit for this product and team.

### Flutter --- NOT SELECTED

Technically excellent, but introduces Dart and another ecosystem without
solving a requirement that React Native cannot solve.

### Kotlin Multiplatform --- NOT SELECTED

Excellent native integration, but unnecessary architectural complexity
for the MVP and less aligned with the existing TypeScript/React
ecosystem.

### Fully Native --- NOT SELECTED

Would increase development and maintenance cost without a corresponding
MVP benefit.

If a future feature requires heavy native processing, implement that
feature as a native module while retaining React Native for the
application.

## Performance Position

React Native is expected to provide a native-quality experience for this
workload.

The primary performance risks are application architecture, not the
framework choice.

Avoid:

-   unnecessary rerenders
-   unvirtualized large lists
-   aggressive polling
-   continuous JS timers
-   large images
-   huge client-side datasets
-   duplicated server state
-   unnecessary background work

Use:

-   SQLite
-   optimistic updates
-   TanStack Query caching
-   pagination
-   push notifications
-   OS scheduling
-   native APIs for sensor-heavy functionality

## Final Decision Summary

``` text
Mobile:
React Native + Expo + TypeScript

State:
TanStack Query + small Zustand

Styling:
Native React Native StyleSheet

Local:
expo-sqlite + expo-secure-store

Backend:
Go modular monolith + REST

Database:
PostgreSQL hosted by Supabase

Authentication:
Supabase Auth

Social Login:
Apple + Google initially

Storage:
Supabase Storage

Notifications:
Expo Notifications + APNs/FCM

Future Web:
Next.js + TypeScript

Repository:
pnpm workspace

CI:
GitHub Actions

Containers:
Docker
```
