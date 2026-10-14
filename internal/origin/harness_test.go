package origin

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"exist.io/internal/auth"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
)

// testServer wires a Server over in-memory store + temp git store and returns
// the router plus the store for seeding. The auth resolver is injected so tests
// control the acting credential's scopes directly.
func testServer(t *testing.T) (*Server, *store.DB, http.Handler) {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	git, err := gitstore.New(filepath.Join(t.TempDir(), "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	srv := New(db, git, auth.NewRegistry(), "https://api.cursor.com", "origin.cursor.com")
	// Default resolver: a user token with every scope, so handler tests focus
	// on behavior, not auth. Auth-denial paths are tested separately.
	srv.resolve = userResolver(allScopes()...)
	return srv, db, srv.Handler()
}

// seedNamespace creates a namespace directly in the store.
func seedNS(t *testing.T, db *store.DB, slug string) store.Namespace {
	t.Helper()
	ns := store.Namespace{ID: idgen.NewNS(), Slug: slug, Type: "team", Name: slug}
	if err := db.CreateNamespace(t.Context(), ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	return ns
}

// userResolver returns an auth.Resolver that yields a user token carrying the
// given scopes, bypassing scope resolution (which needs the store).
func userResolver(scopesList ...string) auth.Resolver {
	return func(*http.Request) (auth.Token, error) {
		return scopedUserToken{scopes: scopesList}, nil
	}
}

func allScopes() []string { return scopes.All() }

// scopedUserToken is a test Token with an explicit scope set.
type scopedUserToken struct {
	scopes []string
}

func (t scopedUserToken) Type() string     { return scopes.TypeUser }
func (t scopedUserToken) Scopes() []string { return t.scopes }
func (t scopedUserToken) Subject() string  { return "user_test" }

// do issues a request against the handler and returns the recorder.
func do(h http.Handler, method, target string, body string) *httptest.ResponseRecorder {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}
