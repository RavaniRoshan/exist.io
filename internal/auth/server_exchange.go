package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
)

const (
	codeInvalidArgument = 3
	codeInternal        = 13
	codeUnauthenticated = 16
)

// ExchangeHandler handles POST /auth/exchange_user_api_key.
// It verifies the caller's raw API key against the store, touches the key's last_used_at,
// generates a scoped access token, stores it in access_tokens, and returns it.
func ExchangeHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeStatus(w, http.StatusMethodNotAllowed, codeInvalidArgument, "method not allowed")
			return
		}

		rawKey, ok := bearerCredential(r)
		if !ok || rawKey == "" {
			writeStatus(w, http.StatusUnauthorized, codeUnauthenticated, "missing API key in Authorization header")
			return
		}

		if !strings.HasPrefix(rawKey, "key_") && !strings.HasPrefix(rawKey, "crsr_") {
			writeStatus(w, http.StatusBadRequest, codeInvalidArgument, "invalid API key prefix")
			return
		}

		hash := HashCredential(rawKey)
		keyRecord, err := db.GetAPIKeyByHash(r.Context(), hash)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeStatus(w, http.StatusUnauthorized, codeUnauthenticated, "invalid API key")
				return
			}
			writeStatus(w, http.StatusInternalServerError, codeInternal, "failed to query key")
			return
		}

		if keyRecord.RevokedAt.Valid && keyRecord.RevokedAt.String != "" {
			writeStatus(w, http.StatusUnauthorized, codeUnauthenticated, "API key has been revoked")
			return
		}

		if keyRecord.ExpiresAt.Valid && keyRecord.ExpiresAt.String != "" {
			exp, err := time.Parse(time.RFC3339, keyRecord.ExpiresAt.String)
			if err == nil && time.Now().After(exp) {
				writeStatus(w, http.StatusUnauthorized, codeUnauthenticated, "API key has expired")
				return
			}
		}

		_ = db.TouchAPIKey(r.Context(), keyRecord.ID)

		tokenScopes := keyRecord.Scopes
		if tokenScopes == "*" || tokenScopes == "" {
			tokenScopes = strings.Join(scopes.All(), " ")
		}

		rawToken, tokHash, err := GenerateAccessToken()
		if err != nil {
			writeStatus(w, http.StatusInternalServerError, codeInternal, "failed to generate access token")
			return
		}

		expiresAt := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
		tok := store.AccessToken{
			ID:        idgen.New("tok"),
			TokenHash: tokHash,
			UserID:    keyRecord.UserID,
			APIKeyID:  sql.NullString{String: keyRecord.ID, Valid: true},
			TokenType: TypeUser,
			Scopes:    tokenScopes,
			CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
			ExpiresAt: expiresAt,
		}

		if err := db.CreateAccessToken(r.Context(), tok); err != nil {
			writeStatus(w, http.StatusInternalServerError, codeInternal, "failed to persist access token")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(exchangeResponse{
			AccessToken:  rawToken,
			RefreshToken: "",
		})
	}
}
