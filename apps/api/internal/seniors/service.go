package seniors

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/meracare/api/internal/auth"
	"github.com/meracare/api/internal/care"
	"github.com/meracare/api/internal/relationships"
)

// Service coordinates senior profiles with the care relationships that grant
// access to them.
type Service struct {
	seniors       *Repository
	relationships *relationships.Repository
}

// NewService builds the service.
func NewService(seniors *Repository, relationships *relationships.Repository) *Service {
	return &Service{seniors: seniors, relationships: relationships}
}

// CreateMode is how the caller relates to the senior they are creating.
type CreateMode string

const (
	// CreateModeSelf is Solo Mode: the caller is the senior. The profile links
	// to their account and they take the `senior` role, with no invitation and
	// no other member required (docs/04, "Solo Workflow").
	CreateModeSelf CreateMode = "self"
	// CreateModeFamily is a relative creating a profile for someone else.
	CreateModeFamily CreateMode = "family_member"
	// CreateModeProfessional is a caregiver taking on a client.
	CreateModeProfessional CreateMode = "professional_caregiver"
)

// Valid reports whether the mode is recognised.
func (m CreateMode) Valid() bool {
	switch m {
	case CreateModeSelf, CreateModeFamily, CreateModeProfessional:
		return true
	default:
		return false
	}
}

// role maps the creation mode onto the creator's role in the new circle.
func (m CreateMode) role() care.Role {
	switch m {
	case CreateModeSelf:
		return care.RoleSenior
	case CreateModeProfessional:
		return care.RoleProfessionalCaregiver
	default:
		return care.RoleFamilyMember
	}
}

// CreateInput is a request to create a senior profile.
type CreateInput struct {
	Mode             CreateMode
	DisplayName      string
	DateOfBirth      *time.Time
	Phone            string
	Address          string
	EmergencyContact string
}

// Create makes a senior profile and the creator's membership atomically.
//
// The two are inseparable: a profile with no relationship would be visible to
// nobody, and a relationship without a profile is meaningless. They are
// committed together or not at all.
func (s *Service) Create(ctx context.Context, principal auth.Principal, input CreateInput) (Membership, error) {
	tx, err := s.seniors.BeginTx(ctx)
	if err != nil {
		return Membership{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(context.WithoutCancel(ctx))
	}()

	// Solo Mode links the profile to the caller's account; a profile created
	// for someone else has no account until that person is invited and joins.
	var seniorUserID uuid.NullUUID
	if input.Mode == CreateModeSelf {
		seniorUserID = uuid.NullUUID{UUID: principal.UserID, Valid: true}
	}

	senior, err := s.seniors.CreateTx(ctx, tx, CreateParams{
		UserID:           seniorUserID,
		CreatedByUserID:  principal.UserID,
		DisplayName:      input.DisplayName,
		DateOfBirth:      input.DateOfBirth,
		Phone:            input.Phone,
		Address:          input.Address,
		EmergencyContact: input.EmergencyContact,
	})
	if err != nil {
		return Membership{}, err
	}

	role := input.Mode.role()
	relationship, err := s.relationships.CreateTx(ctx, tx, relationships.CreateParams{
		SeniorID: senior.ID,
		UserID:   principal.UserID,
		Role:     role,
		// Permissions come from the role's defaults, never from the client.
		Permissions: care.Normalise(care.DefaultPermissions(role)),
		Status:      care.StatusActive,
	})
	if err != nil {
		return Membership{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Membership{}, fmt.Errorf("commit senior profile: %w", err)
	}

	return Membership{Senior: senior, Relationship: relationship}, nil
}

// List returns every senior the caller can currently reach.
func (s *Service) List(ctx context.Context, principal auth.Principal) ([]Membership, error) {
	return s.seniors.ListForUser(ctx, principal.UserID)
}

// Get loads one senior profile. The caller's access is already established by
// the authorization middleware, which supplies the relationship.
func (s *Service) Get(ctx context.Context, seniorID uuid.UUID) (Senior, error) {
	return s.seniors.GetByID(ctx, seniorID)
}

// Update applies profile changes.
func (s *Service) Update(ctx context.Context, seniorID uuid.UUID, params UpdateParams) (Senior, error) {
	return s.seniors.Update(ctx, seniorID, params)
}
