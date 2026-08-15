# Implementation Status

Last updated: 2026-08-15

## Current Phase

**Phase 2 — User + Senior: complete.** Next up is Phase 3 (invitations and the
care circle).

## Repository State

```text
apps/
  api/                     Go modular monolith
    cmd/api/               HTTP server entrypoint
    cmd/migrate/           migration CLI (up | status)
    internal/
      auth/                Supabase JWT verification, Principal, RequireAuth
      authz/               relationship-based authorization middleware
      care/                roles, permissions, statuses, per-role defaults
      config/              environment configuration
      database/            pgx pool, embedded migrations + runner, error helpers
        migrations/        0001_init.sql, 0002_seniors_and_relationships.sql
      relationships/       care relationship model and repository
      seniors/             senior profiles, /v1/seniors
      server/              router wiring, health/readiness
      testsupport/         integration-test database helper
      users/               application user model, repository, service, /v1/me
    pkg/
      httpx/               error envelope, JSON helpers, middleware
      logging/             slog setup
      validation/          request validation helpers
  mobile/                  Expo SDK 57 / React Native 0.86 / Expo Router
    src/
      app/                 routes: index, sign-in, home, onboarding, seniors/[id]
      components/ui/       Button, Card, OptionCard, Screen, Text, TextField
      features/auth/       session restore, sign in/up/out
      features/profile/    /v1/me query + mutation
      features/seniors/    senior queries and mutations
      lib/                 env, supabase, secure storage, api client, query client
      stores/              small Zustand UI store
      theme/               semantic design tokens + ThemeProvider
packages/
  config/                  shared tsconfig / prettier config
  contracts/               TypeScript contracts mirroring the Go API
docker-compose.yml         local PostgreSQL for development and tests
.github/workflows/ci.yml   Go + mobile CI
```

## Completed — Phase 1

| # | Item | Where |
|---|------|-------|
| 1 | Inspect repository | — |
| 2 | Project structure | `apps/`, `packages/` per docs/16 |
| 3 | pnpm workspace | `pnpm-workspace.yaml`, `.npmrc` (`node-linker=hoisted`) |
| 4 | Expo application | `apps/mobile` (SDK 57, RN 0.86.2, Expo Router, typed routes) |
| 5 | Go API | `apps/api`, module `github.com/meracare/api` |
| 6 | TypeScript | `packages/config/tsconfig.base.json`, strict everywhere |
| 7 | Lint / format | `eslint-config-expo`, Prettier, `gofmt`/`go vet` |
| 8 | Environment variables | `apps/api/.env.example`, `apps/mobile/.env.example` |
| 9 | Supabase | Auth client + JWT verification wired; project provisioning pending (see Blockers) |
| 10 | PostgreSQL migrations | embedded SQL + advisory-locked runner, `cmd/migrate` |
| 11 | Database connection | `internal/database`, pgx pool, startup ping, `/readyz` |
| 12 | Supabase Auth integration | `SessionProvider` restores the stored session; `users.Service` maps identity → app user |
| 13 | Go JWT verification | `internal/auth`, HS256 with audience/issuer/expiry checks |
| 14 | API error handling | `pkg/httpx` error envelope per docs/05 |
| 15 | Logging | `pkg/logging` structured slog, request-scoped, no tokens or health data |
| 16 | Testing infrastructure | Go unit + integration tests, Jest/jest-expo, CI workflow |

### Verified end to end

- `go build ./...`, `go vet ./...`, `gofmt -l .` clean.
- `go test ./...` — 10 packages pass, including integration tests against real
  PostgreSQL (`TEST_DATABASE_URL` set).
- Migration applied to a live database; `users` table has the expected
  constraints, partial unique email index, and `updated_at` trigger.
- API booted locally: `/healthz` 200, `/readyz` 200, `/v1/me` 401 without a
  token, 200 with a valid token (creating the application user on first call),
  `PATCH /v1/me` updates and rejects invalid input with `VALIDATION_FAILED`.
- Mobile: `tsc --noEmit` clean, `expo lint` clean, 16 Jest tests pass.

## Completed — Phase 2

| # | Item | Where |
|---|------|-------|
| 17 | User model | Phase 1; extended by the relationships that reference it |
| 18 | SeniorProfile | `senior_profiles` table, `internal/seniors` |
| 19 | CareRelationship | `care_relationships` table, `internal/relationships` |
| 20 | Solo mode | `mode=self` links the profile to the caller and grants the `senior` role, with no invitation |
| 21 | Create senior | `POST /v1/seniors`, mobile onboarding screen |
| 22 | Edit senior | `PATCH /v1/seniors/{id}`, mobile edit screen |
| 23 | Senior dashboard | `GET /v1/seniors/{id}`, mobile dashboard shell |
| — | Role/permission enforcement | `internal/authz` guard on every senior-scoped route |

### Verified end to end

- `go test -race ./...` green across 13 packages, including integration tests
  against real PostgreSQL.
- Over HTTP, with two separate accounts: a stranger listing seniors gets an
  empty list; reading or editing another circle's senior by ID returns 404, not
  403, so senior IDs cannot be probed.
- A professional caregiver — a legitimate member — is refused `PATCH` on their
  own client, proving the check is per-permission and not merely per-membership.
- Unknown request fields are rejected, so a client cannot submit its own
  `permissions`; a second Solo Mode profile returns `CONFLICT`.
- Constraint-level: an unrecognised permission cannot be stored, a circle
  cannot hold two seniors, a user cannot hold two memberships of one circle, and
  a failed create leaves no orphan profile.
- Mobile: `tsc --noEmit` clean, `expo lint` clean, 20 Jest tests pass.

## Architectural Decisions Taken in Phase 1

These are implementation choices within the locked architecture — nothing in
docs/12 or docs/17 was changed.

1. **HTTP router: `go-chi/chi/v5`.** Standard library `http.ServeMux` cannot
   express the nested middleware groups the `/v1` authorization layer needs.
   chi is `http.Handler`-compatible, so it is not a framework lock-in.
2. **Database driver: `pgx/v5` with `pgxpool`.** The standard PostgreSQL driver
   for Go; no ORM, matching "avoid premature database abstraction" (docs/07).
3. **Migrations: embedded SQL with an in-repo runner** rather than a third-party
   CLI. Each migration runs in its own transaction with the bookkeeping insert,
   under a PostgreSQL advisory lock so concurrent deploys serialise. No extra
   tool to install; `go run ./cmd/migrate up`.
4. **JWT verification is asymmetric (JWKS) by default.** The API fetches the
   project's public signing keys from
   `<SUPABASE_URL>/auth/v1/.well-known/jwks.json` and verifies ES256/RS256
   tokens against them. It therefore holds no key capable of minting a token:
   a full compromise of the API's environment does not let an attacker
   impersonate a user. Keys are cached and refreshed only when a token presents
   an unknown `kid` — which is what a rotation looks like — with a one-minute
   throttle so forged tokens cannot amplify requests against Supabase. No
   background timer runs.

   `SUPABASE_JWT_MODE=legacy_hs256` selects the shared-secret verifier for
   projects still on the legacy JWT secret. Configuring a secret while in
   asymmetric mode is a startup error, so an unused forgeable key cannot sit
   unnoticed in a deployment.
5. **Application users are created lazily on first authenticated request.** The
   upsert is idempotent on `auth_user_id`, so no separate registration endpoint
   is needed and concurrent first requests collapse onto one row.
6. **Session storage is chunked SecureStore.** SecureStore warns above ~2KB and
   a Supabase session exceeds that, so values are split across numbered keys.
   docs/09 forbids plain local storage for tokens.
7. **Email/password sign-in first.** Apple and Google are the documented launch
   providers (docs/12) but require provider configuration in the Supabase
   project. Email keeps the foundation verifiable end to end meanwhile.
8. **`/healthz` and `/readyz` are separate.** Liveness performs no dependency
   checks so a database blip never restarts the API; readiness pings PostgreSQL.
9. **Authentication runs before routing inside `/v1`.** An unauthenticated
   caller gets 401 for any `/v1` path, so the API is not a probe for which
   endpoints exist.
10. **Local PostgreSQL uses host port 55432**, leaving the conventional 5432
    free for other projects on a developer machine.

## Architectural Decisions Taken in Phase 2

1. **Permissions are stored per relationship, not derived from the role.** A
   circle can narrow or widen one person's access without inventing new roles,
   which is what Phase 3's invitations will set. The role remains a label and a
   source of defaults; the stored set is what the API enforces.
2. **A denied senior returns 404, never 403.** Distinguishing "you may not see
   this" from "this does not exist" would let anyone enumerate the platform's
   seniors by probing IDs. Malformed IDs take the same path, without a query.
3. **Authorization is middleware, not a handler convention.** A handler obtains
   the senior it operates on from `authz.MustRelationship`, which only exists
   after a guard has run — so a route cannot accidentally skip the check and
   still function.
4. **A profile and its creator's membership are committed in one transaction.**
   A senior with no relationship would be visible to nobody, and unreachable
   even by the person who created it.
5. **The database mirrors the domain vocabulary.** CHECK constraints on role,
   status, and the permissions array mean the API cannot persist a value the
   domain does not define, whatever reaches the handler.
6. **The emergency contact is visible to every member, including professional
   caregivers.** It was initially withheld from caregivers as "private family
   information", which was the wrong reading of docs/02: a caregiver present
   with a senior is exactly who needs to know who to call. The restriction in
   docs/02 concerns information unrelated to the senior's care.
7. **Care events are deferred to Phase 7.** Creating a circle is a
   `MEMBER_JOINED` event under docs/04, but the events table belongs to Phase 7
   and the plan is explicit about building incrementally. No backfill will be
   needed, since there is no production data.
8. **`emergency_contact` is a single text field**, matching docs/03. It may
   split into a name and a phone number when the UI calls for it.

## Blockers

1. **Supabase anon key still needed.** The project at
   `https://axrfytnnnabjdnmwnese.supabase.co` exists and signs with an ES256
   key, which the API loads successfully at startup. The mobile app still needs
   `EXPO_PUBLIC_SUPABASE_ANON_KEY` before a real user can sign in, so the
   round trip from a genuine Supabase-issued token through `/v1/me` has not yet
   been exercised.
2. **Brand assets are still the Expo template placeholders.** `assets/images`
   holds the generated icon/splash. Real MeraCare icon, splash, and the first
   unDraw/Storyset illustrations are needed, along with `ASSET_LICENSES.md`
   (docs/18).

## Pending

Near-term, before or alongside Phase 2:

- **Apple and Google sign-in** once the Supabase project has the providers
  configured (docs/12).
- **OpenAPI document** for `/v1` (docs/05 lists OpenAPI as the contract format).
  Deferred until there are enough endpoints for it to be worth maintaining.
- **Inter typography** — the type scale is in place, but the Inter font files are
  not yet bundled, so the platform default renders at those sizes.

Later phases, unchanged from docs/14 and the Phase 1 plan:

- Phase 2 — User + SeniorProfile + CareRelationship, solo mode, senior dashboard
- Phase 3 — invitations and care circle
- Phase 4 — tasks; Phase 5 — medication; Phase 6 — appointments
- Phase 7 — care events and activity; Phase 8 — notifications
- Phase 9 — dashboards; Phase 10 — messaging
- Phase 11 — offline (`expo-sqlite`, sync queue); Phase 12 — quality

## Next Tasks (Phase 3 — invitations and the care circle)

1. Migration `0003`: `invitations` with token, proposed role and permissions,
   expiry, and status.
2. `GET /v1/seniors/{id}/members` — the care circle, behind `members.view`.
3. `POST /v1/seniors/{id}/invitations` — behind `members.invite`, validating the
   inviter's authority and refusing to grant permissions the inviter lacks.
4. Accept an invitation: create the `active` relationship, and reconcile with
   the pending status this phase already models.
5. `DELETE`/`PATCH` on a member — revoke rather than delete, so care history
   keeps its author.
6. Mobile: care-circle screen, invite form, accept-invitation flow.

## Running It

```bash
pnpm install

# Database
pnpm db:up                       # PostgreSQL on localhost:55432
cp apps/api/.env.example apps/api/.env
pnpm api:migrate

# API
pnpm api:run                     # http://localhost:8080

# Mobile
cp apps/mobile/.env.example apps/mobile/.env
pnpm mobile

# Checks
pnpm api:test                    # add TEST_DATABASE_URL for integration tests
pnpm typecheck && pnpm test
```
