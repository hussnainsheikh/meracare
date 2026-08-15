// Package auth validates Supabase-issued access tokens and establishes the
// authenticated application user for a request.
//
// The API never trusts user IDs, roles, or permissions sent in a request body —
// identity comes from the verified token alone (docs/09-security-privacy.md).
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken is returned for every token failure. The specific reason is
// deliberately not exposed to callers that render responses.
var ErrInvalidToken = errors.New("invalid access token")

// Claims is the subset of a Supabase access token the API relies on.
type Claims struct {
	// AuthUserID is the Supabase `auth.users.id` taken from the `sub` claim.
	AuthUserID uuid.UUID
	Email      string
	// Provider is the sign-in method Supabase recorded (`apple`, `google`, `email`).
	Provider  string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// Verifier validates a raw bearer token and returns its claims.
//
// It is an interface so the HS256 implementation can be replaced by a JWKS
// verifier when the project migrates to asymmetric signing keys, without
// touching the middleware or handlers.
type Verifier interface {
	Verify(ctx context.Context, rawToken string) (*Claims, error)
}

// HS256VerifierOptions configures symmetric token verification.
type HS256VerifierOptions struct {
	// Secret is the Supabase project JWT secret.
	Secret string
	// Audience is the expected `aud` claim, normally "authenticated".
	Audience string
	// Issuer, when set, is required to match the token's `iss` claim.
	// Supabase issues `<SUPABASE_URL>/auth/v1`.
	Issuer string
	// Leeway tolerates small clock differences between Supabase and the API.
	Leeway time.Duration
}

// HS256Verifier validates the symmetric (HS256) access tokens issued by
// Supabase projects using the legacy JWT secret.
type HS256Verifier struct {
	parser *jwt.Parser
	secret []byte
}

// NewHS256Verifier builds a verifier from the project JWT secret.
func NewHS256Verifier(opts HS256VerifierOptions) (*HS256Verifier, error) {
	if strings.TrimSpace(opts.Secret) == "" {
		return nil, errors.New("auth: JWT secret is required")
	}
	if strings.TrimSpace(opts.Audience) == "" {
		return nil, errors.New("auth: expected audience is required")
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithAudience(opts.Audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(opts.Leeway),
	}
	if issuer := strings.TrimRight(strings.TrimSpace(opts.Issuer), "/"); issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(issuer))
	}

	return &HS256Verifier{
		parser: jwt.NewParser(parserOptions...),
		secret: []byte(opts.Secret),
	}, nil
}

// supabaseClaims mirrors the Supabase access token payload.
type supabaseClaims struct {
	jwt.RegisteredClaims
	Email       string `json:"email"`
	AppMetadata struct {
		Provider string `json:"provider"`
	} `json:"app_metadata"`
}

// Verify parses and validates rawToken, returning its claims.
func (v *HS256Verifier) Verify(_ context.Context, rawToken string) (*Claims, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, fmt.Errorf("%w: empty token", ErrInvalidToken)
	}

	claims := &supabaseClaims{}
	token, err := v.parser.ParseWithClaims(rawToken, claims, func(*jwt.Token) (any, error) {
		return v.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	authUserID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("%w: subject is not a UUID", ErrInvalidToken)
	}

	result := &Claims{
		AuthUserID: authUserID,
		Email:      strings.TrimSpace(claims.Email),
		Provider:   claims.AppMetadata.Provider,
	}
	if claims.IssuedAt != nil {
		result.IssuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		result.ExpiresAt = claims.ExpiresAt.Time
	}
	return result, nil
}

// BearerToken extracts the token from an `Authorization: Bearer <token>` header.
func BearerToken(header string) (string, bool) {
	const prefix = "bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}
