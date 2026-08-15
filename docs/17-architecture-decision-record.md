# Architecture Decision Record --- MVP Baseline

## Status

**Accepted / Locked for MVP**

Date: 2026-08-15

## Decision

Build Senior Care as a cross-platform mobile application using React
Native + Expo.

Use:

-   TypeScript
-   Expo Router
-   TanStack Query
-   small Zustand store
-   native React Native styling
-   expo-sqlite
-   expo-secure-store

Use a Go modular-monolith backend exposing a versioned REST API.

Use PostgreSQL hosted by Supabase.

Use Supabase Auth for identity and social authentication.

Use Supabase Storage for private files where appropriate.

Use a pnpm workspace rather than microservices or a complex monorepo
orchestration layer.

## Why

The MVP workload is CRUD-heavy, list-heavy, notification-driven, and
moderately offline-capable. It is not a sensor/graphics-heavy product.

React Native therefore gives the desired native-quality experience while
keeping one mobile codebase.

Go gives a small, efficient backend with strong concurrency and a
straightforward deployment model.

PostgreSQL is the natural relational database for the care domain.

Supabase removes infrastructure work around database hosting and
authentication while preserving PostgreSQL and allowing the same
identity system to be used by a future web application.

## What Is Not Locked

These can evolve without changing the core architecture:

-   exact cloud compute provider for Go
-   exact push notification provider beyond APNs/FCM
-   exact object storage implementation if Supabase Storage becomes
    insufficient
-   exact background job implementation
-   observability vendor
-   analytics vendor
-   web framework details when web is introduced

## What Would Justify Reconsideration

Only reconsider the mobile framework if measured production requirements
demonstrate that React Native cannot satisfy a specific critical
workload after native-module optimization.

Only reconsider PostgreSQL if concrete workload/data-model requirements
make relational PostgreSQL unsuitable.

Only reconsider Go if there is a demonstrated engineering constraint
that cannot reasonably be solved within the modular monolith.

Do not switch technologies because another framework is fashionable or
marginally faster in a benchmark unrelated to this product.

## Stability Principle

The MVP should be built on these decisions. Future features should
extend the architecture rather than trigger unnecessary rewrites.

## Visual System Decision

The MVP visual system is also locked:

-   Primary: `#0F766E` Deep Teal.
-   Supporting palette: mint/teal neutrals plus semantic
    success/warning/danger colors.
-   Light mode is primary; dark mode is supported.
-   Inter is the typography direction.
-   Native React Native styling and semantic theme tokens are used.
-   unDraw is the primary lightweight illustration source.
-   Storyset is the secondary richer illustration source.

The detailed visual specification is
`18-visual-theme-and-illustrations.md`.

Illustration licensing must be recorded per asset. unDraw currently
permits commercial use without attribution under its license, while
Storyset's current free content requires attribution; the applicable
current license must be followed for every Storyset asset.
citeturn0search0turn0search2
