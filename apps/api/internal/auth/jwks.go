package auth

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Asymmetric access-token verification.
//
// Supabase signs access tokens with a private key it never shares and publishes
// the matching public keys as a JWKS document. The API can therefore verify
// tokens but can never mint them: leaking everything this service knows does
// not let an attacker impersonate a user. That is the property a shared HS256
// secret cannot provide, and the reason this is the default.

// asymmetricAlgorithms are the signing algorithms accepted from the key set.
//
// Pinning this list is what prevents algorithm-confusion: without it an
// attacker could take the published public key and present an HS256 token
// signed with that key as the shared secret.
var asymmetricAlgorithms = []string{"ES256", "ES384", "ES512", "RS256", "RS384", "RS512"}

// JWKSVerifierOptions configures asymmetric verification.
type JWKSVerifierOptions struct {
	// JWKSURL is the key set document, e.g.
	// https://<ref>.supabase.co/auth/v1/.well-known/jwks.json
	JWKSURL string
	// Audience is the expected `aud` claim, normally "authenticated".
	Audience string
	// Issuer, when set, is required to match the token's `iss` claim.
	Issuer string
	// Leeway tolerates small clock differences between Supabase and the API.
	Leeway time.Duration
	// HTTPClient fetches the key set. Defaults to a client with a short timeout.
	HTTPClient *http.Client
	// MinRefreshInterval throttles refetching after an unknown key ID, so a
	// stream of forged tokens cannot turn this service into a request amplifier
	// against Supabase. Defaults to one minute.
	MinRefreshInterval time.Duration
}

// JWKSVerifier validates asymmetric access tokens against a cached key set.
//
// Keys are fetched lazily and refreshed when a token presents an unknown `kid`,
// which is exactly what a key rotation looks like. No background timer runs.
type JWKSVerifier struct {
	jwksURL    string
	parser     *jwt.Parser
	client     *http.Client
	minRefresh time.Duration

	// mu guards keys for concurrent verification.
	mu   sync.RWMutex
	keys map[string]crypto.PublicKey

	// fetchMu serialises refreshes so a burst of misses causes one fetch.
	fetchMu     sync.Mutex
	lastAttempt time.Time
}

// NewJWKSVerifier builds a verifier for asymmetric Supabase access tokens.
func NewJWKSVerifier(opts JWKSVerifierOptions) (*JWKSVerifier, error) {
	if strings.TrimSpace(opts.JWKSURL) == "" {
		return nil, errors.New("auth: JWKS URL is required")
	}
	if strings.TrimSpace(opts.Audience) == "" {
		return nil, errors.New("auth: expected audience is required")
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods(asymmetricAlgorithms),
		jwt.WithAudience(opts.Audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(opts.Leeway),
	}
	if issuer := strings.TrimRight(strings.TrimSpace(opts.Issuer), "/"); issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(issuer))
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	minRefresh := opts.MinRefreshInterval
	if minRefresh <= 0 {
		minRefresh = time.Minute
	}

	return &JWKSVerifier{
		jwksURL:    opts.JWKSURL,
		parser:     jwt.NewParser(parserOptions...),
		client:     client,
		minRefresh: minRefresh,
		keys:       map[string]crypto.PublicKey{},
	}, nil
}

// Warmup loads the key set ahead of the first request. A failure is not fatal:
// the caller may log it and continue, since Verify fetches on demand.
func (v *JWKSVerifier) Warmup(ctx context.Context) error {
	keys, err := v.fetch(ctx)
	if err != nil {
		return err
	}

	v.mu.Lock()
	v.keys = keys
	v.mu.Unlock()

	v.fetchMu.Lock()
	v.lastAttempt = time.Now()
	v.fetchMu.Unlock()
	return nil
}

// Verify parses and validates rawToken against the published signing keys.
func (v *JWKSVerifier) Verify(ctx context.Context, rawToken string) (*Claims, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, fmt.Errorf("%w: empty token", ErrInvalidToken)
	}

	claims := &supabaseClaims{}
	token, err := v.parser.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("token has no kid header")
		}
		return v.resolveKey(ctx, kid)
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

// resolveKey returns the public key for kid, refreshing the key set once if the
// key is unknown — the normal path after a key rotation.
func (v *JWKSVerifier) resolveKey(ctx context.Context, kid string) (crypto.PublicKey, error) {
	if key, ok := v.cachedKey(kid); ok {
		return key, nil
	}

	// Serialise refreshes: concurrent misses share one fetch.
	v.fetchMu.Lock()
	defer v.fetchMu.Unlock()

	// Another goroutine may have refreshed while we waited.
	if key, ok := v.cachedKey(kid); ok {
		return key, nil
	}
	if !v.lastAttempt.IsZero() && time.Since(v.lastAttempt) < v.minRefresh {
		return nil, fmt.Errorf("unknown signing key %q", kid)
	}
	v.lastAttempt = time.Now()

	keys, err := v.fetch(ctx)
	if err != nil {
		return nil, err
	}

	v.mu.Lock()
	v.keys = keys
	v.mu.Unlock()

	if key, ok := keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("unknown signing key %q", kid)
}

func (v *JWKSVerifier) cachedKey(kid string) (crypto.PublicKey, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	key, ok := v.keys[kid]
	return key, ok
}

// maxJWKSBytes caps the key set document. A real one is well under 4KB.
const maxJWKSBytes = 1 << 20 // 1 MiB

func (v *JWKSVerifier) fetch(ctx context.Context) (map[string]crypto.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build JWKS request: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := v.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch JWKS: unexpected status %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxJWKSBytes))
	if err != nil {
		return nil, fmt.Errorf("read JWKS: %w", err)
	}
	return parseJWKS(body)
}

// SupabaseJWKSURL builds the key set URL for a Supabase project URL.
func SupabaseJWKSURL(supabaseURL string) string {
	return strings.TrimRight(strings.TrimSpace(supabaseURL), "/") + "/auth/v1/.well-known/jwks.json"
}

// SupabaseIssuer builds the expected `iss` claim for a Supabase project URL.
func SupabaseIssuer(supabaseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(supabaseURL), "/")
	if trimmed == "" {
		return ""
	}
	return trimmed + "/auth/v1"
}
