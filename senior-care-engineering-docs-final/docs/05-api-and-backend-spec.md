# API and Backend Specification

## Backend

-   Go
-   REST API for MVP
-   PostgreSQL
-   Supabase Auth
-   Supabase Storage where appropriate
-   Background jobs for reminders/escalations
-   Structured logging
-   OpenAPI documentation

## API Principles

-   Version APIs: `/v1/...`
-   JSON request/response.
-   UUID primary identifiers.
-   ISO-8601 timestamps.
-   Explicit validation.
-   Relationship-based authorization.
-   Idempotency for mutation endpoints where retries are possible.

## Core Endpoint Groups

### Auth/Profile

The mobile client obtains the Supabase access token.

Backend validates the JWT and maps the Supabase user to the application
user.

### Seniors

``` text
GET    /v1/seniors
POST   /v1/seniors
GET    /v1/seniors/{id}
PATCH  /v1/seniors/{id}
```

### Care Circle

``` text
GET    /v1/seniors/{id}/members
POST   /v1/seniors/{id}/invitations
DELETE /v1/seniors/{id}/members/{memberId}
PATCH  /v1/seniors/{id}/members/{memberId}
```

### Tasks

``` text
GET    /v1/seniors/{id}/tasks
POST   /v1/seniors/{id}/tasks
GET    /v1/tasks/{id}
PATCH  /v1/tasks/{id}
POST   /v1/tasks/{id}/complete
POST   /v1/tasks/{id}/skip
```

### Medications

``` text
GET    /v1/seniors/{id}/medications
POST   /v1/seniors/{id}/medications
PATCH  /v1/medications/{id}
POST   /v1/medications/{id}/instances/{instanceId}/take
```

### Appointments

``` text
GET    /v1/seniors/{id}/appointments
POST   /v1/seniors/{id}/appointments
PATCH  /v1/appointments/{id}
```

### Notes

``` text
GET    /v1/seniors/{id}/notes
POST   /v1/seniors/{id}/notes
PATCH  /v1/notes/{id}
```

### Activity

``` text
GET /v1/seniors/{id}/activity?cursor=...
```

### Messages

``` text
GET  /v1/seniors/{id}/messages?cursor=...
POST /v1/seniors/{id}/messages
```

## Pagination

Use cursor-based pagination for activity, messages, and potentially task
histories.

Avoid offset pagination for large event timelines.

## Error Format

Use a consistent structure:

``` json
{
  "error": {
    "code": "TASK_NOT_ASSIGNED",
    "message": "You are not allowed to complete this task."
  }
}
```

Do not expose internal database or stack-trace details.

## Backend Package Structure

Recommended:

``` text
cmd/
  api/

internal/
  auth/
  users/
  seniors/
  relationships/
  invitations/
  tasks/
  medications/
  appointments/
  notes/
  events/
  notifications/
  messages/
  storage/

pkg/
  http/
  validation/
  logging/
```

Avoid premature microservices.
