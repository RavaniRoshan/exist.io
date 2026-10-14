package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDetectType(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"install", "oit_01k9x2y3w4v5u6t7s8r9q0p", TypeInstallation},
		{"key", "key_01k9x2y3w4v5u6t7s8r9q0p", TypeExchangeable},
		{"crsr", "crsr_01k9x2y3w4v5u6t7s8r9q0p", TypeExchangeable},
		{"user", "eyJhbGciOiJFZERTQSJ9.eyJzdWIiOiJ1XzAifQ.sig", TypeUser},
		{"empty", "", TypeUser},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DetectType(tc.raw); got != tc.want {
				t.Errorf("DetectType(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestTokenImplementations(t *testing.T) {
	jwtPayload, _ := json.Marshal(map[string]string{
		"sub":   "app_01k9x2y3w4v5u6t7s8r9q0p",
		"scope": "read:user write:repo",
	})
	raw := base64.RawURLEncoding.EncodeToString([]byte("header")) + "." +
		base64.RawURLEncoding.EncodeToString(jwtPayload) + ".sig"

	app, err := ParseAppToken(raw)
	if err != nil {
		t.Fatalf("ParseAppToken: %v", err)
	}
	if app.Type() != TypeApp {
		t.Errorf("AppToken.Type() = %q, want %q", app.Type(), TypeApp)
	}
	if app.Subject() != "app_01k9x2y3w4v5u6t7s8r9q0p" {
		t.Errorf("AppToken.Subject() = %q, want app_01k9x2y3w4v5u6t7s8r9q0p", app.Subject())
	}
	wantScopes := []string{"read:user", "write:repo"}
	if len(app.Scopes()) != len(wantScopes) {
		t.Fatalf("AppToken.Scopes() = %v, want %v", app.Scopes(), wantScopes)
	}
	for i, s := range wantScopes {
		if app.Scopes()[i] != s {
			t.Errorf("AppToken.Scopes()[%d] = %q, want %q", i, app.Scopes()[i], s)
		}
	}

	install := NewInstallToken("oit_01k9x2y3w4v5u6t7s8r9q0p")
	if install.Type() != TypeInstallation {
		t.Errorf("InstallToken.Type() = %q, want %q", install.Type(), TypeInstallation)
	}
	if install.Subject() != "oit_01k9x2y3w4v5u6t7s8r9q0p" {
		t.Errorf("InstallToken.Subject() = %q", install.Subject())
	}
	if install.Expired() {
		t.Error("InstallToken.Expired() = true immediately after issuance")
	}
	if got := install.ExpiresAt(); got.Sub(install.issuedAt) != InstallTokenMaxLifetime {
		t.Errorf("install lifetime = %v, want %v", got.Sub(install.issuedAt), InstallTokenMaxLifetime)
	}

	user := NewUserToken("plain-user-credential")
	if user.Type() != TypeUser {
		t.Errorf("UserToken.Type() = %q, want %q", user.Type(), TypeUser)
	}
	if user.Subject() != "plain-user-credential" {
		t.Errorf("UserToken.Subject() = %q", user.Subject())
	}
	if len(user.Scopes()) != 0 {
		t.Errorf("UserToken.Scopes() = %v, want empty", user.Scopes())
	}
}

func TestParseAppTokenErrors(t *testing.T) {
	for _, raw := range []string{"not-a-jwt", "a.b", "a.!!!.c"} {
		if _, err := ParseAppToken(raw); err == nil {
			t.Errorf("ParseAppToken(%q) succeeded, want error", raw)
		}
	}
}

func TestInstallTokenExpired(t *testing.T) {
	tok := InstallToken{raw: "oit_x", issuedAt: time.Now().Add(-InstallTokenMaxLifetime - time.Second)}
	if !tok.Expired() {
		t.Error("InstallToken.Expired() = false after max lifetime, want true")
	}
}

func TestExchangeAPIKey(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotBody, gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"accessToken":"oit_exchanged","refreshToken":"rft_exchanged"}`)
	}))
	defer srv.Close()

	access, refresh, err := ExchangeAPIKey(context.Background(), srv.URL, "key_test123")
	if err != nil {
		t.Fatalf("ExchangeAPIKey: %v", err)
	}
	if access != "oit_exchanged" {
		t.Errorf("accessToken = %q, want oit_exchanged", access)
	}
	if refresh != "rft_exchanged" {
		t.Errorf("refreshToken = %q, want rft_exchanged", refresh)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/auth/exchange_user_api_key" {
		t.Errorf("path = %q, want /auth/exchange_user_api_key", gotPath)
	}
	if gotAuth != "Bearer key_test123" {
		t.Errorf("Authorization = %q, want Bearer key_test123", gotAuth)
	}
	if gotBody != "{}" {
		t.Errorf("body = %q, want {}", gotBody)
	}
	if !strings.HasPrefix(gotContentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
}

func TestExchangeAPIKeyTrailingSlash(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		fmt.Fprint(w, `{"accessToken":"a","refreshToken":"b"}`)
	}))
	defer srv.Close()

	if _, _, err := ExchangeAPIKey(context.Background(), srv.URL+"/", "crsr_test"); err != nil {
		t.Fatalf("ExchangeAPIKey: %v", err)
	}
	if gotPath != "/auth/exchange_user_api_key" {
		t.Errorf("path = %q, want /auth/exchange_user_api_key", gotPath)
	}
}

func TestExchangeAPIKeyErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error":"invalid key"}`)
	}))
	defer srv.Close()

	_, _, err := ExchangeAPIKey(context.Background(), srv.URL, "key_bad")
	if err == nil {
		t.Fatal("ExchangeAPIKey succeeded, want error")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error = %q, want it to mention status 403", err.Error())
	}
}

func TestExchangeAPIKeyBadResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not json`)
	}))
	defer srv.Close()

	if _, _, err := ExchangeAPIKey(context.Background(), srv.URL, "key_test"); err == nil {
		t.Fatal("ExchangeAPIKey succeeded with non-JSON body, want error")
	}
}

func TestExchangeAPIKeyMissingAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"refreshToken":"rft_only"}`)
	}))
	defer srv.Close()

	if _, _, err := ExchangeAPIKey(context.Background(), srv.URL, "key_test"); err == nil {
		t.Fatal("ExchangeAPIKey succeeded without accessToken, want error")
	}
}
