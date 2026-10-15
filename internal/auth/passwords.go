package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	// SessionCookieName is the canonical cookie key for web session authentication.
	SessionCookieName = "exist_session"

	// DefaultSessionDuration is the standard lifetime of a user login session (30 days).
	DefaultSessionDuration = 30 * 24 * time.Hour
)

// HashPassword hashes a raw password using bcrypt with standard enterprise work factor.
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters long")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword verifies a plaintext password against a stored bcrypt hash.
func CheckPassword(password, hash string) bool {
	if hash == "" || password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateSessionToken creates a cryptographically secure random session string and its SHA-256 hash.
func GenerateSessionToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = "sess_" + hex.EncodeToString(b)
	hash = HashCredential(raw)
	return raw, hash, nil
}

// GenerateCSRFToken creates a high-entropy synchronizer token for CSRF mitigation.
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SetSessionCookie configures a secure, HttpOnly session cookie on the response.
func SetSessionCookie(w http.ResponseWriter, rawToken string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true behind TLS termination
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge.Seconds()),
	})
}

// ClearSessionCookie invalidates the browser session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
