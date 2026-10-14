package origin

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

func TestUpsertRepoGrantByEmail(t *testing.T) {
	_, db, h := testServer(t)
	seedNS(t, db, "acme")
	do(h, http.MethodPost, "/v1/origin/repos/acme", `{"repo":{"name":"widgets"}}`)

	// Upsert by user_email → resolves to a user_ id.
	rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/grants",
		`{"permission":"write","user":{"email":"dev@example.com"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("upsert = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var g grantDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &g)
	if g.Permission != "write" {
		t.Errorf("permission = %q, want write (lowercase repo value)", g.Permission)
	}
	if g.User == nil || !strings.HasPrefix(g.User.ID, "user_") {
		t.Errorf("user = %+v, want a user_ id resolved from email", g.User)
	}

	// List returns it.
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/grants", "")
	var list struct {
		Grants []grantDTO `json:"grants"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list.Grants) != 1 || list.Grants[0].Permission != "write" {
		t.Errorf("list = %+v, want 1 write grant", list.Grants)
	}
}

func TestUpsertNamespaceGrantByGroupName(t *testing.T) {
	srv, db, _ := testServer(t)
	seedNS(t, db, "acme")
	h := srv.Handler()

	// Seed a group by name so resolution finds it, then grant it.
	grpID, err := db.UpsertGroupByName(t.Context(), nsID(t, db, "acme"), "eng", idgen.NewGroup)
	if err != nil {
		t.Fatalf("seed group: %v", err)
	}
	rec := do(h, http.MethodPost, "/v1/origin/namespaces/acme/grants",
		`{"permission":"PERMISSION_CONTRIBUTOR","group":{"id":"`+grpID+`"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("upsert = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var g grantDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &g)
	if g.Permission != "PERMISSION_CONTRIBUTOR" {
		t.Errorf("permission = %q, want PERMISSION_CONTRIBUTOR (uppercase ns value)", g.Permission)
	}
	if g.Group == nil || !strings.HasPrefix(g.Group.ID, "grp_") {
		t.Errorf("group = %+v, want grp_ id", g.Group)
	}
}

func TestUpsertTeamGroupGrant(t *testing.T) {
	_, db, h := testServer(t)
	seedNS(t, db, "acme")

	rec := do(h, http.MethodPost, "/v1/origin/namespaces/acme/grants",
		`{"permission":"PERMISSION_READ","teamGroup":{"kind":"members"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("upsert = %d; body=%s", rec.Code, rec.Body.String())
	}
	var g grantDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &g)
	if g.TeamGroup == nil || g.TeamGroup.Kind != "members" {
		t.Errorf("teamGroup = %+v, want members", g.TeamGroup)
	}
}

func TestDeleteGrant(t *testing.T) {
	_, db, h := testServer(t)
	seedNS(t, db, "acme")
	do(h, http.MethodPost, "/v1/origin/repos/acme", `{"repo":{"name":"widgets"}}`)
	do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/grants", `{"permission":"read","user":{"email":"a@b.com"}}`)

	// Delete by principal (user id resolved from email).
	rec := do(h, http.MethodDelete, "/v1/origin/repos/acme/widgets/grants", `{"user":{"email":"a@b.com"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete = %d; body=%s", rec.Code, rec.Body.String())
	}
	// List is now empty.
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/grants", "")
	var list struct {
		Grants []grantDTO `json:"grants"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list.Grants) != 0 {
		t.Errorf("after delete, list = %+v, want empty", list.Grants)
	}
}

func TestGrantValidation(t *testing.T) {
	_, db, h := testServer(t)
	seedNS(t, db, "acme")

	// No principal → 400.
	if rec := do(h, http.MethodPost, "/v1/origin/namespaces/acme/grants", `{"permission":"PERMISSION_READ"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("no principal = %d, want 400", rec.Code)
	}
	// Wrong permission enum for namespace → 400.
	if rec := do(h, http.MethodPost, "/v1/origin/namespaces/acme/grants", `{"permission":"read","teamGroup":{"kind":"admins"}}`); rec.Code != http.StatusBadRequest {
		t.Errorf("lowercase perm on ns = %d, want 400", rec.Code)
	}
	// Unknown namespace → 404.
	if rec := do(h, http.MethodGet, "/v1/origin/namespaces/ghost/grants", ""); rec.Code != http.StatusNotFound {
		t.Errorf("unknown ns = %d, want 404", rec.Code)
	}
}

// nsID returns the namespace id for a slug.
func nsID(t *testing.T, db *store.DB, slug string) string {
	t.Helper()
	ns, err := db.GetNamespaceBySlug(t.Context(), slug)
	if err != nil {
		t.Fatalf("GetNamespaceBySlug: %v", err)
	}
	return ns.ID
}
