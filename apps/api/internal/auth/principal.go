package auth

import (
	"context"

	"github.com/google/uuid"
)

// Principal is the authenticated caller for one request: the Supabase identity
// plus the application user it maps to.
//
// A Principal says who the caller is. It says nothing about what they may do —
// that is decided per senior by relationship-based authorization
// (docs/02-permissions-and-authorization.md).
type Principal struct {
	// UserID is the application `users.id`.
	UserID uuid.UUID
	// AuthUserID is the Supabase `auth.users.id`.
	AuthUserID uuid.UUID
	Email      string
}

type principalContextKey struct{}

// WithPrincipal returns a context carrying the authenticated caller.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

// PrincipalFrom returns the authenticated caller, if the request passed through
// the authentication middleware.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

// MustPrincipal returns the authenticated caller. It panics when called outside
// an authenticated route, which is a programming error rather than a runtime
// condition — the recovery middleware turns it into a 500 instead of allowing
// an unauthenticated request to proceed as if it were authenticated.
func MustPrincipal(ctx context.Context) Principal {
	principal, ok := PrincipalFrom(ctx)
	if !ok {
		panic("auth: no principal in context; handler is not behind RequireAuth")
	}
	return principal
}
