package seniors

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// DisplayName returns just the senior's name.
//
// The invitation preview needs to say who the circle is for, and nothing else
// about them — the preview is readable by anyone holding the token.
func (r *Repository) DisplayName(ctx context.Context, seniorID uuid.UUID) (string, error) {
	var name string
	err := r.pool.QueryRow(ctx,
		`SELECT display_name FROM senior_profiles WHERE id = $1`, seniorID).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get senior display name: %w", err)
	}
	return name, nil
}
