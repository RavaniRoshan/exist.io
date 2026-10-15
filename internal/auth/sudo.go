package auth

import (
	"context"
	"strings"

	"exist.io/internal/store"
)

// VerifySudoChallenge verifies whether a security secret provided by a user
// satisfies the sudo mode re-authentication challenge.
// It accepts:
// 1. The caller's plaintext account password.
// 2. An active, unrevoked personal access token / API key belonging strictly to the caller.
// It returns false if the secret is empty, belongs to a different user, or is invalid.
func VerifySudoChallenge(ctx context.Context, db *store.DB, caller *store.User, secret string) bool {
	if db == nil || caller == nil {
		return false
	}
	cleanSecret := strings.TrimSpace(secret)
	if cleanSecret == "" {
		return false
	}

	// 1. Check against caller's password
	if caller.PasswordHash != "" && CheckPassword(cleanSecret, caller.PasswordHash) {
		return true
	}

	// 2. Check against caller's API keys (must belong strictly to caller.ID)
	keyHash := HashCredential(cleanSecret)
	if key, err := db.GetAPIKeyByHash(ctx, keyHash); err == nil && key.ID != "" {
		// Strictly enforce ownership: the key must belong to the caller!
		if key.UserID == caller.ID {
			if !key.RevokedAt.Valid || key.RevokedAt.String == "" {
				return true
			}
		}
	}

	return false
}
