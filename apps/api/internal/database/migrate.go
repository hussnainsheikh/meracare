package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// migrationFiles holds the SQL migrations, embedded so the binary can migrate
// without shipping loose files.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

// migrationLockID is an arbitrary but stable key for the PostgreSQL advisory
// lock that serialises concurrent migration runs (e.g. rolling deploys).
const migrationLockID int64 = 7_265_631_001

// Migration is one forward-only schema change.
type Migration struct {
	Version int64
	Name    string
	SQL     string
}

// LoadMigrations reads and orders the embedded migrations.
//
// Files must be named `<version>_<name>.sql`, e.g. `0001_init.sql`.
func LoadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	migrations := make([]Migration, 0, len(entries))
	seen := make(map[int64]string, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		rawVersion, name, found := strings.Cut(strings.TrimSuffix(entry.Name(), ".sql"), "_")
		if !found {
			return nil, fmt.Errorf("migration %q must be named <version>_<name>.sql", entry.Name())
		}
		version, err := strconv.ParseInt(rawVersion, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("migration %q has a non-numeric version: %w", entry.Name(), err)
		}
		if previous, duplicate := seen[version]; duplicate {
			return nil, fmt.Errorf("migration version %d is used by both %q and %q", version, previous, entry.Name())
		}
		seen[version] = entry.Name()

		contents, err := migrationFiles.ReadFile(path.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{Version: version, Name: name, SQL: string(contents)})
	}

	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	return migrations, nil
}

// Migrate applies every migration that has not yet been recorded.
//
// Each migration runs inside its own transaction together with the bookkeeping
// insert, so a failure leaves the schema and the ledger consistent.
func Migrate(ctx context.Context, pool *Pool, logger *slog.Logger) error {
	migrations, err := LoadMigrations()
	if err != nil {
		return err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationLockID); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		// Best effort: the lock is also released when the connection closes.
		_, _ = conn.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, migrationLockID)
	}()

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     bigint      PRIMARY KEY,
			name        text        NOT NULL,
			applied_at  timestamptz NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := appliedVersions(ctx, conn.Conn())
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if _, done := applied[migration.Version]; done {
			continue
		}

		if err := applyMigration(ctx, conn.Conn(), migration); err != nil {
			return err
		}
		logger.Info("applied migration",
			slog.Int64("version", migration.Version),
			slog.String("name", migration.Name),
		)
	}

	return nil
}

func appliedVersions(ctx context.Context, conn *pgx.Conn) (map[int64]struct{}, error) {
	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]struct{})
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}
	return applied, nil
}

func applyMigration(ctx context.Context, conn *pgx.Conn, migration Migration) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", migration.Version, err)
	}
	defer func() {
		_ = tx.Rollback(context.WithoutCancel(ctx))
	}()

	if _, err := tx.Exec(ctx, migration.SQL); err != nil {
		return fmt.Errorf("apply migration %d (%s): %w", migration.Version, migration.Name, err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
		migration.Version, migration.Name,
	); err != nil {
		return fmt.Errorf("record migration %d: %w", migration.Version, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %d: %w", migration.Version, err)
	}
	return nil
}

// MigrationStatus describes one migration and whether it has been applied.
type MigrationStatus struct {
	Version int64
	Name    string
	Applied bool
}

// Status reports which migrations have been applied.
func Status(ctx context.Context, pool *Pool) ([]MigrationStatus, error) {
	migrations, err := LoadMigrations()
	if err != nil {
		return nil, err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	applied, err := appliedVersions(ctx, conn.Conn())
	if err != nil {
		// A missing ledger simply means nothing has been applied yet.
		var pgErr interface{ SQLState() string }
		if errors.As(err, &pgErr) && pgErr.SQLState() == "42P01" {
			applied = map[int64]struct{}{}
		} else {
			return nil, err
		}
	}

	statuses := make([]MigrationStatus, 0, len(migrations))
	for _, migration := range migrations {
		_, done := applied[migration.Version]
		statuses = append(statuses, MigrationStatus{
			Version: migration.Version,
			Name:    migration.Name,
			Applied: done,
		})
	}
	return statuses, nil
}
