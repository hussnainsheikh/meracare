# Mobile Architecture

## Locked Decision

The MVP mobile application will use:

-   React Native
-   Expo
-   TypeScript
-   Expo Router
-   TanStack Query
-   Small Zustand store
-   React Native StyleSheet/native styling
-   `expo-sqlite` for durable local/offline data
-   `expo-secure-store` where secure local credential/session storage is
    appropriate

## Visual System

Use the locked design system defined in:

`18-visual-theme-and-illustrations.md`

The mobile app must consume semantic theme tokens rather than hardcoded
colors.

Primary brand color:

`#0F766E`

Illustration sources:

-   unDraw
-   Storyset

Illustration assets must be bundled/served according to their applicable
licenses and should not be hotlinked from third-party illustration sites
in production.

## Architecture

``` text
Screens
  ↓
Feature Hooks / View Models
  ↓
TanStack Query + Small Zustand
  ↓
Repository Layer
  ├── Go API
  └── SQLite
  ↓
Native OS APIs
```

## State Rules

### TanStack Query

Use for all server state:

-   queries
-   mutations
-   caching
-   invalidation
-   synchronization status

Do not copy server state into Zustand.

### Zustand

Use only for small client/UI state:

-   selected senior
-   UI preferences
-   temporary filters
-   modal state
-   onboarding state

Do not turn Zustand into an application-wide database.

## Styling

Use React Native `StyleSheet` and platform primitives.

No Tailwind/native-styling framework is required.

Create a small internal design system:

-   spacing tokens
-   typography
-   colors
-   radii
-   elevation/shadows
-   buttons
-   inputs
-   cards
-   list rows
-   status indicators
-   illustration containers

The theme must work on iOS, Android, and future web through semantic
tokens.

## Navigation

Use Expo Router.

The route structure should reflect product features rather than backend
tables.

## Performance Rules

-   Virtualize long lists.
-   Avoid unnecessary global state.
-   Keep server data in TanStack Query.
-   Paginate activity and messages.
-   Optimize images.
-   Avoid polling when push/realtime is sufficient.
-   Avoid long-running JS loops.
-   Use OS scheduling for local notifications.
-   Keep heavy processing off the JS thread.

## Offline Strategy

Core workflows must remain usable with weak/intermittent connectivity.

At minimum:

-   read cached senior/task data
-   complete a task locally
-   mark medication as taken locally
-   queue mutation for synchronization
-   display sync state

Example:

``` text
Tap Complete
  ↓
SQLite transaction
  ↓
UI updates immediately
  ↓
Sync queue
  ↓
Go API mutation
  ↓
Server confirms
```

## Native Escape Hatch

React Native remains the application platform.

Native modules can be introduced for:

-   Apple Health / HealthKit
-   Android Health Connect
-   Bluetooth medical devices
-   advanced sensor processing
-   platform-specific background execution
-   other OS-specific capabilities

A native subsystem is an extension of the React Native application, not
a reason to rewrite the application.
