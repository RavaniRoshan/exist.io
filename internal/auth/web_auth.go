package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"exist.io/internal/audit"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"golang.org/x/crypto/ssh"
)

// LoginRequest defines credentials for web login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest defines fields for new user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// UserProfileDTO renders sanitized user data.
type UserProfileDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

func toUserProfileDTO(u store.User) UserProfileDTO {
	return UserProfileDTO{
		ID:        u.ID,
		Email:     u.Email,
		Login:     u.Login,
		Name:      u.Name,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

// LoginHandler processes user authentication, issuing session cookies and CSRF tokens.
func LoginHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
			return
		}

		req.Email = strings.TrimSpace(req.Email)
		if req.Email == "" || req.Password == "" {
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		user, err := db.GetUserByEmail(ctx, req.Email)
		if err != nil {
			// Check if login was provided instead of email
			user, err = db.GetUserByLogin(ctx, req.Email)
			if err != nil {
				_ = audit.Log(audit.Event{
					Actor:     req.Email,
					ActorType: "user",
					Action:    "auth.login_failed",
					Resource:  "auth",
					ClientIP:  r.RemoteAddr,
					Severity:  audit.SeverityWarning,
					Details:   "Login attempt with unknown email or login",
				})
				http.Error(w, "Invalid email or password", http.StatusUnauthorized)
				return
			}
		}

		if user.Status == "suspended" {
			_ = audit.Log(audit.Event{
				Actor:     user.Email,
				ActorType: "user",
				Action:    "auth.login_suspended",
				Resource:  user.ID,
				ClientIP:  r.RemoteAddr,
				Severity:  audit.SeverityWarning,
				Details:   "Suspended user attempted login",
			})
			http.Error(w, "Account has been suspended. Contact your forge administrator.", http.StatusForbidden)
			return
		}

		if !CheckPassword(req.Password, user.PasswordHash) {
			_ = audit.Log(audit.Event{
				Actor:     user.Email,
				ActorType: "user",
				Action:    "auth.login_failed",
				Resource:  user.ID,
				ClientIP:  r.RemoteAddr,
				Severity:  audit.SeverityWarning,
				Details:   "Invalid password provided",
			})
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		rawToken, tokenHash, err := GenerateSessionToken()
		if err != nil {
			http.Error(w, "Failed to generate session", http.StatusInternalServerError)
			return
		}

		csrfToken, err := GenerateCSRFToken()
		if err != nil {
			http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().UTC().Add(DefaultSessionDuration).Format(time.RFC3339Nano)
		sess := store.Session{
			ID:        idgen.New("sess"),
			UserID:    user.ID,
			TokenHash: tokenHash,
			CSRFToken: csrfToken,
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
			ExpiresAt: expiresAt,
		}

		if err := db.CreateSession(ctx, sess); err != nil {
			http.Error(w, "Failed to create session record", http.StatusInternalServerError)
			return
		}

		SetSessionCookie(w, rawToken, DefaultSessionDuration)

		_ = audit.Log(audit.Event{
			Actor:     user.Email,
			ActorType: "user",
			Action:    "auth.login_success",
			Resource:  user.ID,
			ClientIP:  r.RemoteAddr,
			Severity:  audit.SeverityInfo,
			Details:   "User successfully logged in",
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user":         toUserProfileDTO(user),
			"sessionToken": rawToken,
			"csrfToken":    csrfToken,
		})
	}
}

// RegisterHandler registers a new enterprise user. The first registered user automatically becomes admin.
func RegisterHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
			return
		}

		req.Email = strings.TrimSpace(req.Email)
		req.Login = strings.TrimSpace(req.Login)
		req.Name = strings.TrimSpace(req.Name)

		if req.Email == "" || req.Login == "" || req.Password == "" {
			http.Error(w, "Email, login username, and password are required", http.StatusBadRequest)
			return
		}

		if len(req.Password) < 8 {
			http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		totalUsers, err := db.CountUsers(ctx)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		role := "member"
		if totalUsers == 0 {
			role = "admin" // First user in forge is bootstrap admin
		}

		passHash, err := HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Failed to hash password: "+err.Error(), http.StatusBadRequest)
			return
		}

		userID := idgen.NewUser()
		user := store.User{
			ID:           userID,
			Email:        req.Email,
			Login:        req.Login,
			Name:         req.Name,
			PasswordHash: passHash,
			Role:         role,
			Status:       "active",
		}

		if err := db.CreateUserWithPassword(ctx, user); err != nil {
			if errors.Is(err, store.ErrConflict) {
				http.Error(w, "A user with this email or username already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create personal namespace for user
		nsID := idgen.NewNS()
		ns := store.Namespace{
			ID:          nsID,
			Slug:        req.Login,
			Type:        "user",
			Name:        req.Name,
			OwnerUserID: userID,
		}
		_ = db.CreateNamespace(ctx, ns)

		// Create admin grant on personal namespace
		grant := store.Grant{
			ID:            idgen.New("grant"),
			Scope:         "namespace",
			NamespaceID:   nsID,
			PrincipalKind: "user",
			UserID:        userID,
			Permission:    "PERMISSION_ADMIN",
		}
		_, _ = db.UpsertGrant(ctx, grant)

		// Auto-login newly registered user
		rawToken, tokenHash, _ := GenerateSessionToken()
		csrfToken, _ := GenerateCSRFToken()
		expiresAt := time.Now().UTC().Add(DefaultSessionDuration).Format(time.RFC3339Nano)
		sess := store.Session{
			ID:        idgen.New("sess"),
			UserID:    user.ID,
			TokenHash: tokenHash,
			CSRFToken: csrfToken,
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
			ExpiresAt: expiresAt,
		}
		_ = db.CreateSession(ctx, sess)
		SetSessionCookie(w, rawToken, DefaultSessionDuration)

		_ = audit.Log(audit.Event{
			Actor:     user.Email,
			ActorType: "user",
			Action:    "auth.user_register",
			Resource:  user.ID,
			ClientIP:  r.RemoteAddr,
			Severity:  audit.SeverityInfo,
			Details:   fmt.Sprintf("User registered with role %s", role),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user":         toUserProfileDTO(user),
			"sessionToken": rawToken,
			"csrfToken":    csrfToken,
		})
	}
}

// LogoutHandler revokes the active session and clears cookies.
func LogoutHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()
		if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
			sessHash := HashCredential(cookie.Value)
			if sess, user, err := db.GetSessionByHash(ctx, sessHash); err == nil {
				_ = db.DeleteSession(ctx, sess.ID)
				_ = audit.Log(audit.Event{
					Actor:     user.Email,
					ActorType: "user",
					Action:    "auth.user_logout",
					Resource:  user.ID,
					ClientIP:  r.RemoteAddr,
					Severity:  audit.SeverityInfo,
					Details:   "User logged out",
				})
			}
		}

		ClearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

// SessionHandler returns the currently authenticated user's profile and CSRF token.
func SessionHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		// 1. Try session cookie
		if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
			sessHash := HashCredential(cookie.Value)
			sess, user, err := db.GetSessionByHash(ctx, sessHash)
			if err == nil {
				exp, _ := time.Parse(time.RFC3339Nano, sess.ExpiresAt)
				if time.Now().Before(exp) && user.Status != "suspended" {
					_ = json.NewEncoder(w).Encode(map[string]any{
						"authenticated": true,
						"user":          toUserProfileDTO(user),
						"csrfToken":     sess.CSRFToken,
					})
					return
				}
			}
		}

		// 2. Try Bearer token
		if raw, ok := bearerCredential(r); ok && raw != "" {
			tokHash := HashCredential(raw)
			tok, err := db.GetAccessTokenByHash(ctx, tokHash)
			if err == nil {
				user, err := db.GetUser(ctx, tok.UserID)
				if err == nil && user.Status != "suspended" {
					_ = json.NewEncoder(w).Encode(map[string]any{
						"authenticated": true,
						"user":          toUserProfileDTO(user),
						"csrfToken":     "bearer-authenticated",
					})
					return
				}
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"authenticated": false,
			"user":          nil,
		})
	}
}

// AdminListUsersHandler returns all users registered on the forge.
func AdminListUsersHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		users, err := db.ListUsers(r.Context())
		if err != nil {
			http.Error(w, "Failed to list users", http.StatusInternalServerError)
			return
		}
		var dtos []UserProfileDTO
		for _, u := range users {
			dtos = append(dtos, toUserProfileDTO(u))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"users": dtos,
		})
	}
}

// AdminCreateUserHandler allows an administrator to provision new enterprise users.
func AdminCreateUserHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		var req struct {
			Email    string `json:"email"`
			Login    string `json:"login"`
			Name     string `json:"name"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Email == "" || req.Login == "" || req.Password == "" {
			http.Error(w, "Email, login, and password are required", http.StatusBadRequest)
			return
		}

		passHash, err := HashPassword(req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Role == "" {
			req.Role = "member"
		}

		userID := idgen.NewUser()
		user := store.User{
			ID:           userID,
			Email:        req.Email,
			Login:        req.Login,
			Name:         req.Name,
			PasswordHash: passHash,
			Role:         req.Role,
			Status:       "active",
		}

		if err := db.CreateUserWithPassword(r.Context(), user); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		_ = audit.Log(audit.Event{
			Actor:     "admin",
			ActorType: "user",
			Action:    "admin.user_created",
			Resource:  userID,
			ClientIP:  r.RemoteAddr,
			Severity:  audit.SeverityWarning,
			Details:   fmt.Sprintf("Admin created user %s (%s)", user.Email, user.Role),
		})

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(toUserProfileDTO(user))
	}
}

// UserCreateTokenHandler creates a self-service Personal Access Token (PAT).
func UserCreateTokenHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		var req struct {
			Name   string `json:"name"`
			Scopes string `json:"scopes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			req.Name = "Personal Access Token"
		}
		if req.Scopes == "" {
			req.Scopes = "*"
		}

		// Identify acting user from session or context
		actingUser := "user_root"
		if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
			if sess, _, err := db.GetSessionByHash(r.Context(), HashCredential(cookie.Value)); err == nil {
				actingUser = sess.UserID
			}
		}

		rawKey, prefix, hash, err := GenerateAPIKey()
		if err != nil {
			http.Error(w, "Failed to generate key", http.StatusInternalServerError)
			return
		}

		key := store.APIKey{
			ID:        idgen.New("key"),
			UserID:    actingUser,
			KeyHash:   hash,
			KeyPrefix: prefix,
			Name:      req.Name,
			Scopes:    req.Scopes,
			CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		}

		if err := db.CreateAPIKey(r.Context(), key); err != nil {
			http.Error(w, "Failed to store key", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        key.ID,
			"name":      key.Name,
			"keyPrefix": key.KeyPrefix,
			"token":     rawKey,
			"message":   "Save this token now; it will not be displayed again.",
		})
	}
}

// UserAddSSHKeyHandler enrolls a user's SSH public key for Git SSH access.
func UserAddSSHKeyHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		var req struct {
			Name      string `json:"name"`
			PublicKey string `json:"publicKey"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		req.PublicKey = strings.TrimSpace(req.PublicKey)
		if req.PublicKey == "" {
			http.Error(w, "publicKey is required", http.StatusBadRequest)
			return
		}

		parsedKey, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey))
		if err != nil {
			http.Error(w, "Invalid SSH public key format: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			if comment != "" {
				req.Name = comment
			} else {
				req.Name = "SSH Key"
			}
		}

		fingerprint := ssh.FingerprintSHA256(parsedKey)
		actingUser := "user_root"
		if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
			if sess, _, err := db.GetSessionByHash(r.Context(), HashCredential(cookie.Value)); err == nil {
				actingUser = sess.UserID
			}
		}

		sshKey := store.SSHKey{
			ID:          idgen.New("sshkey"),
			UserID:      actingUser,
			Name:        req.Name,
			Fingerprint: fingerprint,
			PublicKey:   req.PublicKey,
			KeyType:     parsedKey.Type(),
			CreatedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		}

		if err := db.AddUserSSHKey(r.Context(), sshKey); err != nil {
			if errors.Is(err, store.ErrConflict) {
				http.Error(w, "This SSH key is already enrolled", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to add SSH key", http.StatusInternalServerError)
			return
		}

		_ = audit.Log(audit.Event{
			Actor:     actingUser,
			ActorType: "user",
			Action:    "auth.ssh_key_added",
			Resource:  fingerprint,
			ClientIP:  r.RemoteAddr,
			Severity:  audit.SeverityInfo,
			Details:   fmt.Sprintf("Enrolled SSH key %q (%s)", sshKey.Name, fingerprint),
		})

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(sshKey)
	}
}

// UserListSSHKeysHandler returns all SSH keys enrolled by the caller.
func UserListSSHKeysHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		actingUser := "user_root"
		if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
			if sess, _, err := db.GetSessionByHash(r.Context(), HashCredential(cookie.Value)); err == nil {
				actingUser = sess.UserID
			}
		}

		keys, err := db.ListUserSSHKeys(r.Context(), actingUser)
		if err != nil {
			http.Error(w, "Failed to list keys", http.StatusInternalServerError)
			return
		}
		if keys == nil {
			keys = []store.SSHKey{}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": keys})
	}
}

// UserListTokensHandler returns all active API keys (Personal Access Tokens) for the caller.
func UserListTokensHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		actingUser := "user_root"
		if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
			if sess, _, err := db.GetSessionByHash(r.Context(), HashCredential(cookie.Value)); err == nil {
				actingUser = sess.UserID
			}
		}

		keys, err := db.ListAPIKeysByUserID(r.Context(), actingUser)
		if err != nil {
			http.Error(w, "Failed to list tokens", http.StatusInternalServerError)
			return
		}
		type tokenItem struct {
			ID         string   `json:"id"`
			Name       string   `json:"name"`
			KeyPrefix  string   `json:"prefix"`
			Scopes     []string `json:"scopes"`
			CreatedAt  string   `json:"created_at"`
			LastUsedAt string   `json:"last_used_at,omitempty"`
		}
		out := make([]tokenItem, 0, len(keys))
		for _, k := range keys {
			var scopesList []string
			if k.Scopes == "*" {
				scopesList = []string{"all"}
			} else {
				scopesList = strings.Fields(k.Scopes)
			}
			lastUsed := ""
			if k.LastUsedAt.Valid {
				lastUsed = k.LastUsedAt.String
			}
			out = append(out, tokenItem{
				ID:         k.ID,
				Name:       k.Name,
				KeyPrefix:  k.KeyPrefix,
				Scopes:     scopesList,
				CreatedAt:  k.CreatedAt,
				LastUsedAt: lastUsed,
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"tokens": out})
	}
}

// UserDeleteTokenHandler revokes an API key.
func UserDeleteTokenHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		keyID := r.PathValue("id")
		if keyID == "" {
			keyID = strings.TrimPrefix(r.URL.Path, "/api/v1/user/tokens/")
		}
		if keyID == "" {
			http.Error(w, "token id is required", http.StatusBadRequest)
			return
		}

		if err := db.RevokeAPIKey(r.Context(), keyID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				http.Error(w, "Token not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to revoke token", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "revoked": keyID})
	}
}

// UserDeleteSSHKeyHandler removes an enrolled SSH key.
func UserDeleteSSHKeyHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		actingUser := "user_root"
		if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
			if sess, _, err := db.GetSessionByHash(r.Context(), HashCredential(cookie.Value)); err == nil {
				actingUser = sess.UserID
			}
		}

		keyID := r.PathValue("id")
		if keyID == "" {
			keyID = strings.TrimPrefix(r.URL.Path, "/api/v1/user/ssh-keys/")
		}
		if keyID == "" {
			http.Error(w, "ssh key id is required", http.StatusBadRequest)
			return
		}

		if err := db.DeleteUserSSHKey(r.Context(), actingUser, keyID); err != nil {
			http.Error(w, "Failed to delete SSH key", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "deleted": keyID})
	}
}
