package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// FindIDByEmail resolves an account by email address.
//
// Used by the invitation flow to tell "already a member" from "not signed up
// yet". Matching is case-insensitive and uses the partial unique index on
// lower(email).
func (r *Repository) FindIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	normalised := strings.ToLower(strings.TrimSpace(email))
	if normalised == "" {
		return uuid.Nil, ErrNotFound
	}

	var id uuid.UUID
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM users WHERE lower(email) = $1`, normalised).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("find user by email: %w", err)
	}
	return id, nil
}
