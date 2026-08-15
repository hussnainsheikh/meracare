// Package testsupport provides shared helpers for integration tests that need
// a real PostgreSQL database.
//
// Integration tests are skipped unless TEST_DATABASE_URL is set, so
// `go test ./...` stays fast and hermetic while CI and local developers can run
// the full suite against `docker compose up -d postgres`.
package testsupport

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/meracare/api/internal/database"
	"github.com/meracare/api/pkg/logging"
)

// DatabaseURLEnv names the environment variable that enables integration tests.
const DatabaseURLEnv = "TEST_DATABASE_URL"

// RequireDatabase returns a migrated connection pool, skipping the test when no
// test database is configured.
//
// Each call truncates application tables so tests start from a known state.
func RequireDatabase(t *testing.T) *database.Pool {
	t.Helper()

	url := os.Getenv(DatabaseURLEnv)
	if url == "" {
		t.Skipf("set %s to run integration tests (see docker-compose.yml)", DatabaseURLEnv)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, database.Options{URL: url, MaxConns: 4})
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := database.Migrate(ctx, pool, logging.New(io.Discard, logging.Options{Level: "error"})); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	truncate(ctx, t, pool)

	return pool
}

// truncate empties every application table. schema_migrations is preserved so
// the schema is migrated once per database rather than once per test.
func truncate(ctx context.Context, t *testing.T, pool *database.Pool) {
	t.Helper()

	if _, err := pool.Exec(ctx, `TRUNCATE TABLE users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate test database: %v", err)
	}
}
