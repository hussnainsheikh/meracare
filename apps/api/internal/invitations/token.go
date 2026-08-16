// Package invitations issues, validates and redeems the invitations that add a
// person to a senior's care circle.
//
// An invitation is a proposal. It grants nothing until it is accepted, and
// acceptance is what creates the care relationship (docs/04).
package invitations

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// tokenBytes is the entropy in a raw invitation token.
//
// 256 bits makes guessing infeasible, which is the property that matters: the
// token is the sole bearer of authority to join a care circle.
const tokenBytes = 32

// Token is a raw invitation token. It exists only in memory and in the single
// response that delivers it; the database stores its hash.
type Token string

// NewToken mints a cryptographically random token.
func NewToken() (Token, error) {
	buffer := make([]byte, tokenBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate invitation token: %w", err)
	}
	// URL-safe and unpadded, so the token survives a path segment or a link
	// without escaping.
	return Token(base64.RawURLEncoding.EncodeToString(buffer)), nil
}

// Hash returns the value stored in the database.
//
// A plain SHA-256 is the correct primitive for a 256-bit random token. Password
// hashes such as bcrypt or argon2 exist to slow brute force against low-entropy
// human secrets; applying one here would add cost to every lookup while
// defending against an attack that cannot succeed anyway.
func (t Token) Hash() []byte {
	sum := sha256.Sum256([]byte(t))
	return sum[:]
}

// Valid reports whether the token is well-formed.
//
// Rejecting malformed tokens before touching the database keeps a stream of
// junk from turning into a stream of queries.
func (t Token) Valid() bool {
	decoded, err := base64.RawURLEncoding.DecodeString(string(t))
	return err == nil && len(decoded) == tokenBytes
}

// EqualHash compares two token hashes in constant time.
//
// Lookups are by indexed hash rather than by scan, so this is not on the hot
// path; it exists for the places that compare a candidate against a known
// value, where timing should not leak how much of the hash matched.
func EqualHash(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
