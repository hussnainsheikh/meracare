# Performance Requirements

## Product Goal

The application should feel like a native, responsive care tool.

## Initial Targets

These are engineering targets, not contractual SLAs.

### Mobile

-   Usable UI after cold launch: target \< 2 seconds on representative
    modern devices.
-   Local navigation: target \< 200ms perceived response.
-   Task completion: immediate optimistic UI response.
-   Normal lists: smooth scrolling around 60fps.
-   No unnecessary continuous background JS execution.

### API

Target:

-   p50 \< 200ms for simple reads.
-   p95 \< 500ms for normal API operations.

Measure in production-like environments rather than relying on local
development timing.

## Mobile Load Expectations

Typical senior dashboard:

-   5--20 current tasks
-   1--10 medications
-   0--10 upcoming appointments
-   20--50 recent activity events

This is a low rendering workload.

The app should not load an entire multi-year care history into React
state.

## Scaling Example

At 100,000 seniors, assume:

-   5 medications/senior
-   10 recurring task templates/senior
-   5--20 care events/day/senior

The backend may process millions of records, but each mobile client
should only fetch the relevant slice.

## Performance Rules

-   Cursor pagination.
-   Database indexes.
-   Query only required fields.
-   Avoid N+1 queries.
-   Cache stable data.
-   Batch where appropriate.
-   Use optimistic mutations.
-   Use push instead of aggressive polling.
-   Profile before optimizing.
