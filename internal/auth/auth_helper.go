package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"exist.io/internal/store"
)

// ErrUnauthenticated is returned when a request carries no valid session or token.
var ErrUnauthenticated = errors.New("auth: unauthenticated request")

// ResolveUser resolves the authenticated user from either an active session cookie
// or an Authorization: Bearer token (API Key or Access Token).
// Returns ErrUnauthenticated if no valid credential is provided.
func ResolveUser(ctx context.Context, db *store.DB, r *http.Request) (*store.User, error) {
	if db == nil || r == nil {
		return nil, ErrUnauthenticated
	}

	// 1. Try session cookie first
	if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
		sessHash := HashCredential(cookie.Value)
		sess, user, err := db.GetSessionByHash(ctx, sessHash)
		if err == nil {
			exp, err := time.Parse(time.RFC3339Nano, sess.ExpiresAt)
			if err != nil {
				exp, err = time.Parse(time.RFC3339, sess.ExpiresAt)
			}
			if err == nil && time.Now().Before(exp) && user.Status != "suspended" {
				return &user, nil
			}
		}
	}

	// 2. Try Authorization: Bearer <token>
	if raw, ok := bearerCredential(r); ok && raw != "" {
		tokHash := HashCredential(raw)

		// 2a. Check personal access token / API key
		if strings.HasPrefix(raw, "key_") || strings.HasPrefix(raw, "pat_") || strings.HasPrefix(raw, "crsr_") {
			key, err := db.GetAPIKeyByHash(ctx, tokHash)
			if err == nil && key.ID != "" {
				if key.RevokedAt.Valid && key.RevokedAt.String != "" {
					return nil, errors.New("auth: API key has been revoked")
				}
				if key.ExpiresAt.Valid && key.ExpiresAt.String != "" {
					exp, err := time.Parse(time.RFC3339, key.ExpiresAt.String)
					if err == nil && time.Now().After(exp) {
						return nil, errors.New("auth: API key has expired")
					}
				}
				user, err := db.GetUser(ctx, key.UserID)
				if err == nil && user.Status != "suspended" {
					_ = db.TouchAPIKey(ctx, key.ID)
					return &user, nil
				}
			}
		}

		// 2b. Check active access token (session token or OAuth token)
		tokRecord, err := db.GetAccessTokenByHash(ctx, tokHash)
		if err == nil && tokRecord.ID != "" {
			if tokRecord.RevokedAt.Valid && tokRecord.RevokedAt.String != "" {
				return nil, errors.New("auth: access token has been revoked")
			}
			exp, err := time.Parse(time.RFC3339, tokRecord.ExpiresAt)
			if err == nil && time.Now().After(exp) {
				return nil, errors.New("auth: access token has expired")
			}
			user, err := db.GetUser(ctx, tokRecord.UserID)
			if err == nil && user.Status != "suspended" {
				return &user, nil
			}
		}

		// 2c. Check if Bearer token is a raw session token (e.g. from automated tests/scripts)
		sess, user, err := db.GetSessionByHash(ctx, tokHash)
		if err == nil {
			exp, err := time.Parse(time.RFC3339Nano, sess.ExpiresAt)
			if err != nil {
				exp, err = time.Parse(time.RFC3339, sess.ExpiresAt)
			}
			if err == nil && time.Now().Before(exp) && user.Status != "suspended" {
				return &user, nil
			}
		}
	}

	return nil, ErrUnauthenticated
}
