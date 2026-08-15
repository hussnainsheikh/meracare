package auth

import (
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
)

// JSON Web Key parsing.
//
// Only the key types Supabase issues access tokens with are supported: EC
// (ES256/384/512) and RSA (RS256/384/512). Anything else in the key set is
// skipped rather than rejected, so an unrelated future key cannot break
// authentication.

// jwkSet is the document served at `<issuer>/.well-known/jwks.json`.
type jwkSet struct {
	Keys []jwk `json:"keys"`
}

// jwk is one key. Field names follow RFC 7517.
type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`

	// EC
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`

	// RSA
	N string `json:"n"`
	E string `json:"e"`
}

// parseJWKS turns a key set document into public keys indexed by `kid`.
//
// Keys that are unusable, unsupported, or not intended for signatures are
// skipped. An error is returned only when the document yields no usable key at
// all, since that leaves the API unable to verify anything.
func parseJWKS(document []byte) (map[string]crypto.PublicKey, error) {
	var set jwkSet
	if err := json.Unmarshal(document, &set); err != nil {
		return nil, fmt.Errorf("parse JWKS: %w", err)
	}

	keys := make(map[string]crypto.PublicKey, len(set.Keys))
	for _, key := range set.Keys {
		// `use` is optional; when present it must mark a signature key.
		if key.Kid == "" || (key.Use != "" && key.Use != "sig") {
			continue
		}
		publicKey, err := key.publicKey()
		if err != nil {
			continue
		}
		keys[key.Kid] = publicKey
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("parse JWKS: no usable signing keys")
	}
	return keys, nil
}

// publicKey converts one JWK into a crypto.PublicKey.
func (k jwk) publicKey() (crypto.PublicKey, error) {
	switch k.Kty {
	case "EC":
		return k.ecPublicKey()
	case "RSA":
		return k.rsaPublicKey()
	default:
		return nil, fmt.Errorf("unsupported key type %q", k.Kty)
	}
}

func (k jwk) ecPublicKey() (crypto.PublicKey, error) {
	var curve elliptic.Curve
	var ecdhCurve ecdh.Curve
	switch k.Crv {
	case "P-256":
		curve, ecdhCurve = elliptic.P256(), ecdh.P256()
	case "P-384":
		curve, ecdhCurve = elliptic.P384(), ecdh.P384()
	case "P-521":
		curve, ecdhCurve = elliptic.P521(), ecdh.P521()
	default:
		return nil, fmt.Errorf("unsupported curve %q", k.Crv)
	}

	x, err := decodeBase64URLBigInt(k.X)
	if err != nil {
		return nil, fmt.Errorf("decode x: %w", err)
	}
	y, err := decodeBase64URLBigInt(k.Y)
	if err != nil {
		return nil, fmt.Errorf("decode y: %w", err)
	}

	// Reject a point that is not on the curve. crypto/ecdh performs that
	// validation when it parses an uncompressed point.
	byteLen := (curve.Params().BitSize + 7) / 8
	point := make([]byte, 1+2*byteLen)
	point[0] = 4 // uncompressed point marker
	if x.BitLen() > curve.Params().BitSize || y.BitLen() > curve.Params().BitSize {
		return nil, fmt.Errorf("coordinate is too large for curve %q", k.Crv)
	}
	x.FillBytes(point[1 : 1+byteLen])
	y.FillBytes(point[1+byteLen:])

	if _, err := ecdhCurve.NewPublicKey(point); err != nil {
		return nil, fmt.Errorf("invalid %s point: %w", k.Crv, err)
	}

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

func (k jwk) rsaPublicKey() (crypto.PublicKey, error) {
	modulus, err := decodeBase64URLBigInt(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode n: %w", err)
	}
	exponentBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode e: %w", err)
	}
	// A public exponent larger than 8 bytes is not a key we should accept.
	if len(exponentBytes) == 0 || len(exponentBytes) > 8 {
		return nil, fmt.Errorf("invalid exponent length %d", len(exponentBytes))
	}

	exponent := new(big.Int).SetBytes(exponentBytes)
	if !exponent.IsInt64() || exponent.Int64() < 3 || exponent.Int64() > int64(^uint32(0)) {
		return nil, fmt.Errorf("invalid public exponent")
	}
	// Reject undersized moduli outright rather than trusting a weak key.
	if modulus.BitLen() < 2048 {
		return nil, fmt.Errorf("RSA modulus is only %d bits", modulus.BitLen())
	}

	return &rsa.PublicKey{N: modulus, E: int(exponent.Int64())}, nil
}

// decodeBase64URLBigInt decodes an unpadded base64url integer, as JWK requires.
func decodeBase64URLBigInt(value string) (*big.Int, error) {
	if value == "" {
		return nil, fmt.Errorf("value is empty")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(decoded), nil
}
