package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/meracare/api/internal/invitations"
	"github.com/meracare/api/internal/users"
)

// userLookup adapts the users repository to the narrow interface the invitation
// flow needs.
//
// The adapter lives here, at the composition root, so internal/invitations
// depends on a two-method interface rather than on the whole users package.
type userLookup struct {
	repo *users.Repository
}

func (l userLookup) GetByID(ctx context.Context, id uuid.UUID) (invitations.UserSummary, error) {
	user, err := l.repo.GetByID(ctx, id)
	if err != nil {
		return invitations.UserSummary{}, err
	}
	return invitations.UserSummary{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}, nil
}

func (l userLookup) FindIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	return l.repo.FindIDByEmail(ctx, email)
}
