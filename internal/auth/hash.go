package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashCredential computes the SHA-256 hex digest of a raw secret (API key or access token).
func HashCredential(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// GenerateAPIKey generates a high-entropy secret API key.
// Returns the raw key, the display prefix (first 12 chars), and the stored SHA-256 hash.
func GenerateAPIKey() (rawKey, prefix, hash string, err error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("auth: generate random key: %w", err)
	}
	rawKey = "key_" + hex.EncodeToString(b)
	prefix = rawKey[:12]
	hash = HashCredential(rawKey)
	return rawKey, prefix, hash, nil
}

// GenerateAccessToken generates a high-entropy access token for an active session.
// Returns the raw token and its stored SHA-256 hash.
func GenerateAccessToken() (rawToken, hash string, err error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("auth: generate random token: %w", err)
	}
	rawToken = "eit_" + hex.EncodeToString(b)
	hash = HashCredential(rawToken)
	return rawToken, hash, nil
}
