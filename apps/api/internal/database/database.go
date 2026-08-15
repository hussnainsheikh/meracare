// Package database owns the PostgreSQL connection pool and schema migrations.
//
// PostgreSQL (hosted by Supabase) is the authoritative store; the Go API is the
// only writer for business-critical data (docs/07-database-and-sync.md).
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the shared connection pool type used by every repository.
type Pool = pgxpool.Pool

// Options configures the connection pool.
type Options struct {
	URL      string
	MaxConns int32
	// ConnectTimeout bounds the initial connectivity check.
	ConnectTimeout time.Duration
}

// Connect opens the pool and verifies connectivity before returning, so a bad
// DATABASE_URL fails at startup rather than on the first request.
func Connect(ctx context.Context, opts Options) (*Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	if opts.MaxConns > 0 {
		poolConfig.MaxConns = opts.MaxConns
	}
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	connectTimeout := opts.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 10 * time.Second
	}
	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Ping reports whether the database is reachable. Used by the readiness probe.
func Ping(ctx context.Context, pool *Pool) error {
	if pool == nil {
		return fmt.Errorf("database pool is not initialised")
	}
	return pool.Ping(ctx)
}
