package origin

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateAndGetRepo(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")

	// GET before create → 404.
	if rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("GET before create = %d, want 404; body=%s", rec.Code, rec.Body.String())
	}

	// Create.
	rec := do(h, http.MethodPost, "/v1/origin/repos/acme", `{"repo":{"name":"widgets"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("create = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var created repoDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.ID == "" || created.Name != "widgets" || created.FullName != "acme/widgets" {
		t.Errorf("created = %+v", created)
	}
	if created.DefaultBranch != "main" {
		t.Errorf("defaultBranch = %q, want main (default)", created.DefaultBranch)
	}
	if created.Owner.Slug != "acme" {
		t.Errorf("owner.slug = %q, want acme", created.Owner.Slug)
	}
	if created.CloneURL == "" {
		t.Error("cloneUrl empty")
	}

	// The bare git repo was initialized on disk.
	if !srv.Git().Exists("acme", "widgets") {
		t.Error("git repo not initialized on disk")
	}

	// GET after create → 200.
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET after create = %d, want 200", rec.Code)
	}

	// Duplicate create → 409.
	if rec := do(h, http.MethodPost, "/v1/origin/repos/acme", `{"repo":{"name":"widgets"}}`); rec.Code != http.StatusConflict {
		t.Errorf("duplicate create = %d, want 409", rec.Code)
	}
}

func TestCreateRepoValidates(t *testing.T) {
	_, db, h := testServer(t)
	seedNS(t, db, "acme")

	// Missing name → 400.
	if rec := do(h, http.MethodPost, "/v1/origin/repos/acme", `{"repo":{}}`); rec.Code != http.StatusBadRequest {
		t.Errorf("missing name = %d, want 400", rec.Code)
	}
	// Unknown namespace → 404.
	if rec := do(h, http.MethodPost, "/v1/origin/repos/ghost", `{"repo":{"name":"x"}}`); rec.Code != http.StatusNotFound {
		t.Errorf("unknown ns = %d, want 404", rec.Code)
	}
}

func TestUpdateRepo(t *testing.T) {
	_, db, h := testServer(t)
	seedNS(t, db, "acme")
	do(h, http.MethodPost, "/v1/origin/repos/acme", `{"repo":{"name":"widgets"}}`)

	rec := do(h, http.MethodPatch, "/v1/origin/repos/acme/widgets", `{"defaultBranch":"develop"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var got repoDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got.DefaultBranch != "develop" {
		t.Errorf("defaultBranch = %q, want develop", got.DefaultBranch)
	}

	// Update missing repo → 404.
	if rec := do(h, http.MethodPatch, "/v1/origin/repos/acme/ghost", `{"defaultBranch":"x"}`); rec.Code != http.StatusNotFound {
		t.Errorf("update missing = %d, want 404", rec.Code)
	}
}

func TestListRepos(t *testing.T) {
	_, db, h := testServer(t)
	seedNS(t, db, "acme")
	for _, n := range []string{"a", "b", "c"} {
		do(h, http.MethodPost, "/v1/origin/repos/acme", `{"repo":{"name":"`+n+`"}}`)
	}

	rec := do(h, http.MethodGet, "/v1/origin/repos/acme", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list = %d", rec.Code)
	}
	var resp struct {
		Repositories  []repoDTO `json:"repositories"`
		NextPageToken string    `json:"nextPageToken"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Repositories) != 3 {
		t.Errorf("listed %d repos, want 3", len(resp.Repositories))
	}
}

func TestListNamespaces(t *testing.T) {
	_, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedNS(t, db, "beta")

	rec := do(h, http.MethodGet, "/v1/origin/namespaces", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list namespaces = %d", rec.Code)
	}
	var resp struct {
		Namespaces []namespaceDTO `json:"namespaces"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Namespaces) != 2 {
		t.Errorf("listed %d namespaces, want 2", len(resp.Namespaces))
	}
}

// TestAuthDenialInsufficientScope confirms the middleware blocks a user token
// lacking the required scope (403 Status envelope).
func TestAuthDenialInsufficientScope(t *testing.T) {
	srv, db, _ := testServer(t)
	seedNS(t, db, "acme")
	// Rebuild handler with a scopeless user token.
	srv.resolve = userResolver()
	h := srv.Handler()

	rec := do(h, http.MethodGet, "/v1/origin/repos/acme", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("scopeless list = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	if env.Code != 7 {
		t.Errorf("envelope code = %d, want 7 (PERMISSION_DENIED)", env.Code)
	}
}
