package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier is the subset of pgx shared by the connection pool and a transaction.
//
// Repositories hold one of these rather than the pool itself, so the same
// method body serves an ordinary request and a transactional one. That is what
// lets a domain change and the care event describing it be written through the
// same connection, and so commit or roll back together
// (plans/phase7.md §§6, 26).
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// InTx runs fn inside a single transaction, committing when it returns nil and
// rolling back when it returns an error.
//
// The rollback is deferred rather than called on each error path: a panic in fn
// must not leave a transaction open on a pooled connection, where it would be
// handed to the next request. Rolling back an already-committed transaction is
// a no-op in pgx, so the deferred call is safe on the success path too.
func InTx(ctx context.Context, pool *Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
