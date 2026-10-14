package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exist.io/pkg/scopes"
)

// okHandler is a terminal handler that records that it was reached.
func okHandler(reached *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reached != nil {
			*reached = true
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
}

// resolverFrom returns a Resolver that always yields tok (ignoring the request).
func resolverFrom(tok Token, err error) Resolver {
	return func(*http.Request) (Token, error) { return tok, err }
}

// statusBody decodes the Origin Status envelope.
type statusBody struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details"`
}

func TestMiddlewareAllowsScopedToken(t *testing.T) {
	// A user token holding namespace:repositories:read may call ListRepos.
	tok := NewUserToken("user-cred")
	// Wrap so Scopes() reports the granted scope (user tokens get scopes from a
	// store in production; here we simulate a resolved credential).
	scoped := scopedToken{tok, []string{scopes.NamespaceRepositoriesRead}}

	var reached bool
	h := authMiddleware(resolverFrom(scoped, nil), okHandler(&reached), OperationMeta{
		ID: "OriginService_ListRepos",
		Reqs: scopes.Requirements{
			TokenTypes: []string{scopes.TypeUser},
			Scopes:     []string{scopes.NamespaceRepositoriesRead},
		},
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/origin/repos", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !reached {
		t.Error("next handler was not reached")
	}
}

func TestMiddlewareDeniesMissingScope(t *testing.T) {
	tok := scopedToken{NewUserToken("user-cred"), []string{scopes.NamespaceRepositoriesRead}}
	var reached bool
	h := authMiddleware(resolverFrom(tok, nil), okHandler(&reached), OperationMeta{
		ID: "OriginService_CreateRepo",
		Reqs: scopes.Requirements{
			TokenTypes: []string{scopes.TypeUser},
			Scopes:     []string{scopes.NamespaceRepositoriesCreate},
		},
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/origin/repos/acme", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	if reached {
		t.Error("next handler was reached despite missing scope")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body statusBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode Status envelope: %v; body=%s", err, rec.Body.String())
	}
	if body.Code != scopes.CodePermissionDenied {
		t.Errorf("envelope code = %d, want %d", body.Code, scopes.CodePermissionDenied)
	}
	if body.Message == "" {
		t.Error("envelope message is empty")
	}
}

func TestMiddlewareDeniesWrongTokenType(t *testing.T) {
	// A user token calling an app-only operation is rejected (403).
	tok := scopedToken{NewUserToken("user-cred"), []string{scopes.AppMetadataRead}}
	var reached bool
	h := authMiddleware(resolverFrom(tok, nil), okHandler(&reached), OperationMeta{
		ID:   "OriginService_GetAuthenticatedApp",
		Reqs: scopes.Requirements{TokenTypes: []string{scopes.TypeApp}, Scopes: []string{scopes.AppMetadataRead}, Ambient: true},
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/origin/app", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	var body statusBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Code != scopes.CodePermissionDenied {
		t.Errorf("envelope code = %d, want %d", body.Code, scopes.CodePermissionDenied)
	}
}

func TestMiddlewareUnauthenticatedWhenNoToken(t *testing.T) {
	var reached bool
	h := authMiddleware(resolverFrom(nil, errMissingToken), okHandler(&reached), OperationMeta{
		ID:   "OriginService_ListRepos",
		Reqs: scopes.Requirements{TokenTypes: []string{scopes.TypeUser}, Scopes: []string{scopes.NamespaceRepositoriesRead}},
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/origin/repos", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if reached {
		t.Error("next handler was reached without a token")
	}
	var body statusBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Code != scopes.CodeUnauthenticated {
		t.Errorf("envelope code = %d, want %d", body.Code, scopes.CodeUnauthenticated)
	}
}

func TestMiddlewareAmbientAppToken(t *testing.T) {
	// A bare app token satisfies an ambient operation.
	app, err := ParseAppToken(makeAppJWT(t, "app_01", ""))
	if err != nil {
		t.Fatalf("ParseAppToken: %v", err)
	}
	var reached bool
	h := authMiddleware(resolverFrom(app, nil), okHandler(&reached), OperationMeta{
		ID:   "OriginService_GetAuthenticatedApp",
		Reqs: scopes.Requirements{TokenTypes: []string{scopes.TypeApp}, Scopes: []string{scopes.AppMetadataRead}, Ambient: true},
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/origin/app", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegistry(t *testing.T) {
	reg := NewRegistry()
	if _, ok := reg.Requirements("OriginService_ListRepos"); ok {
		t.Fatal("empty registry returned an operation")
	}
	reg.Register("OriginService_ListRepos", scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser},
		Scopes:     []string{scopes.NamespaceRepositoriesRead},
	})
	reqs, ok := reg.Requirements("OriginService_ListRepos")
	if !ok {
		t.Fatal("registered operation not found")
	}
	if len(reqs.Scopes) != 1 || reqs.Scopes[0] != scopes.NamespaceRepositoriesRead {
		t.Errorf("reqs.Scopes = %v, want [namespace:repositories:read]", reqs.Scopes)
	}
}

func TestRequireHelper(t *testing.T) {
	reg := NewRegistry()
	reg.Register("OriginService_ListRepos", scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser},
		Scopes:     []string{scopes.NamespaceRepositoriesRead},
	})

	var reached bool
	// Registered op + scoped user token → allowed.
	h := Require(reg, "OriginService_ListRepos", okHandler(&reached))
	r := httptest.NewRequest(http.MethodGet, "/v1/origin/repos", nil)
	// BearerResolver maps an unprefixed credential to a user token; simulate the
	// granted scope via a wrapper is not possible through BearerResolver, so use
	// a raw user credential and confirm the type gate passes but scope gate is
	// store-dependent. Here we only assert the unregistered-op fail-closed path
	// and the 401 path, which do not depend on scope resolution.
	r.Header.Set("Authorization", "Bearer some-user-cred")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	// A bare user token has no scopes resolved by BearerResolver, so ListRepos
	// (which requires namespace:repositories:read) is denied 403.
	if rec.Code != http.StatusForbidden {
		t.Fatalf("registered op with scopeless user: status = %d, want 403", rec.Code)
	}
	if reached {
		t.Error("handler reached despite missing scope")
	}

	// Unregistered op → fail closed with 403.
	reached = false
	h2 := Require(reg, "OriginService_DoesNotExist", okHandler(&reached))
	rec2 := httptest.NewRecorder()
	h2.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/v1/origin/nope", nil))
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("unregistered op: status = %d, want 403", rec2.Code)
	}
	var body statusBody
	_ = json.Unmarshal(rec2.Body.Bytes(), &body)
	if body.Code != scopes.CodePermissionDenied {
		t.Errorf("unregistered op envelope code = %d, want %d", body.Code, scopes.CodePermissionDenied)
	}
	if reached {
		t.Error("handler reached for unregistered op")
	}
}

// BearerResolver parses the Authorization header into a Token.
func TestBearerResolver(t *testing.T) {
	// App JWT.
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+makeAppJWT(t, "app_01", "app:metadata:read"))
	tok, err := BearerResolver(r)
	if err != nil {
		t.Fatalf("BearerResolver(app JWT): %v", err)
	}
	if tok.Type() != TypeApp {
		t.Errorf("type = %q, want app", tok.Type())
	}

	// Installation token.
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer oit_01kabc")
	tok, err = BearerResolver(r)
	if err != nil {
		t.Fatalf("BearerResolver(oit_): %v", err)
	}
	if tok.Type() != TypeInstallation {
		t.Errorf("type = %q, want installation", tok.Type())
	}

	// Missing header.
	if _, err := BearerResolver(httptest.NewRequest(http.MethodGet, "/", nil)); err == nil {
		t.Error("BearerResolver(no header) = nil error, want error")
	}

	// Raw API key must be exchanged first.
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer key_01kabc")
	if _, err := BearerResolver(r); err == nil {
		t.Error("BearerResolver(key_) = nil error, want must-exchange error")
	}
}

// scopedToken adapts a Token with an explicit scope list (simulating a
// credential whose scopes were resolved from a store).
type scopedToken struct {
	Token
	scopes []string
}

func (s scopedToken) Scopes() []string { return s.scopes }

func makeAppJWT(t *testing.T, sub, scope string) string {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{"sub": sub, "scope": scope})
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA"}`))
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}
