package testsupport

import (
	"strings"
	"testing"
)

// The integration suite truncates every application table. These tests are the
// guarantee that it can only ever do so to a development container.

func TestLocalDatabasesAreAllowed(t *testing.T) {
	cases := []string{
		// What docker-compose.yml and docs/IMPLEMENTATION_STATUS.md tell people
		// to use.
		"postgres://meracare:meracare@localhost:55432/meracare?sslmode=disable",
		// What CI uses, so restoring the workflow needs no exception.
		"postgres://meracare:meracare@localhost:5432/meracare?sslmode=disable",
		"postgresql://meracare@127.0.0.1:5432/meracare",
		"postgres://meracare@[::1]:5432/meracare",
		"postgres://meracare@LOCALHOST:5432/meracare",
		// No port.
		"postgres://meracare@localhost/meracare",
		// Keyword form, which libpq also accepts.
		"host=localhost port=5432 dbname=meracare",
		"dbname=meracare host=127.0.0.1",
		// No host at all: a Unix socket, which is local by definition.
		"dbname=meracare user=meracare",
	}

	for _, url := range cases {
		t.Run(url, func(t *testing.T) {
			if err := RequireLocalHost(url); err != nil {
				t.Errorf("RequireLocalHost(%q) = %v, want it allowed", url, err)
			}
		})
	}
}

// The reason this check exists. Pointing the suite at the hosted project would
// erase real care records, with no confirmation step and nothing to notice
// until afterwards.
func TestHostedDatabasesAreRefused(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{
			"a Supabase session pooler",
			"postgres://postgres.abcdefgh:secret@aws-0-eu-west-2.pooler.supabase.com:5432/postgres",
		},
		{
			"a Supabase direct connection",
			"postgres://postgres:secret@db.axrfytnnnabjdnmwnese.supabase.co:5432/postgres",
		},
		{
			"any other remote host",
			"postgres://meracare@10.0.0.7:5432/meracare",
		},
		{
			"a hostname that merely looks local",
			"postgres://meracare@localhost.example.com:5432/meracare",
		},
		{
			"a remote host in keyword form",
			"host=db.axrfytnnnabjdnmwnese.supabase.co port=5432 dbname=postgres",
		},
		{
			// A remote host cannot be smuggled in as the userinfo half of the
			// URL: net/url reads the part after the last @ as the host.
			"a remote host disguised as a username",
			"postgres://localhost:secret@db.supabase.co:5432/postgres",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := RequireLocalHost(tc.url)
			if err == nil {
				t.Fatalf("RequireLocalHost(%q) allowed a hosted database", tc.url)
			}
			// The message has to name the variable and point at the fix: this
			// is read by somebody who has just had their test run stopped.
			if !strings.Contains(err.Error(), DatabaseURLEnv) {
				t.Errorf("message does not name %s: %v", DatabaseURLEnv, err)
			}
			if !strings.Contains(err.Error(), "docker-compose.yml") {
				t.Errorf("message does not say what to use instead: %v", err)
			}
		})
	}
}

// Anything unreadable is refused rather than assumed local: failing a test run
// is recoverable, and truncating a hosted database is not.
func TestUnreadableConnectionStringsAreRefused(t *testing.T) {
	for _, url := range []string{"nonsense", "mysql://localhost/meracare", ""} {
		t.Run(url, func(t *testing.T) {
			if err := RequireLocalHost(url); err == nil {
				t.Errorf("RequireLocalHost(%q) was trusted", url)
			}
		})
	}
}
