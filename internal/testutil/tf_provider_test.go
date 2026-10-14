package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exist.io/internal/app"
	"exist.io/internal/auth"
	"exist.io/internal/config"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
)

type testToken struct {
	scopes []string
}

func (t testToken) Type() string     { return scopes.TypeUser }
func (t testToken) Scopes() []string { return t.scopes }
func (t testToken) Subject() string  { return "user_test" }

func TestTerraformProvider_OriginResourcesCompatibility(t *testing.T) {
	ctx := context.Background()

	// 1. Initialize exist.io application in test mode
	cfg := config.Config{
		DataDir:   ":memory:",
		PublicURL: "http://localhost:8080/v1/origin",
		GitURL:    "localhost:8080",
	}
	application, err := app.Open(cfg)
	if err != nil {
		t.Fatalf("app.Open: %v", err)
	}
	t.Cleanup(func() { _ = application.Close() })

	// Bypass auth checks for test runner with omnipotent test token
	application.Origin.SetResolver(func(*http.Request) (auth.Token, error) {
		return testToken{scopes: scopes.All()}, nil
	})

	// 2. Seed namespace, repository, and user
	ns := store.Namespace{
		ID:   idgen.NewNS(),
		Slug: "acme",
		Type: "team",
		Name: "Acme Corp",
	}
	if err := application.Store.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	repo := store.Repo{
		ID:                  idgen.NewRepo(),
		NamespaceID:         ns.ID,
		Name:                "rocket",
		FullName:            "acme/rocket",
		DefaultBranch:       "main",
		Visibility:          "private",
		AllowMergeCommit:    true,
		AllowSquashMerge:    true,
		DeleteBranchOnMerge: false,
	}
	if err := application.Store.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	userID, err := application.Store.UpsertUserByEmail(ctx, "alice@acme.dev", idgen.NewUser)
	if err != nil {
		t.Fatalf("UpsertUserByEmail: %v", err)
	}

	handler := application.Handler()

	doReq := func(method, path string, body any) *httptest.ResponseRecorder {
		var r *bytes.Reader
		if body != nil {
			raw, _ := json.Marshal(body)
			r = bytes.NewReader(raw)
		} else {
			r = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(method, path, r)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("User-Agent", "terraform-provider-cursor/0.8.0")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	// -------------------------------------------------------------
	// Resource 1: cursor_origin_repo (Data Source & View)
	// -------------------------------------------------------------
	t.Run("cursor_origin_repo by slug", func(t *testing.T) {
		rec := doReq(http.MethodGet, "/v1/origin/repos/acme/rocket", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("get repo expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var rDTO map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &rDTO)
		if rDTO["name"] != "rocket" || rDTO["id"] != repo.ID {
			t.Fatalf("unexpected repo DTO: %+v", rDTO)
		}
	})

	t.Run("cursor_origin_repo by id (_ sentinel)", func(t *testing.T) {
		rec := doReq(http.MethodGet, "/v1/origin/repos/_/"+repo.ID, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("get repo by id expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// -------------------------------------------------------------
	// Resource 2: cursor_origin_repo_grant
	// -------------------------------------------------------------
	t.Run("cursor_origin_repo_grant CRUD", func(t *testing.T) {
		// PUT (upsert grant)
		grantBody := map[string]any{
			"user": map[string]any{
				"id": userID,
			},
			"permission": "admin",
		}
		rec := doReq(http.MethodPost, "/v1/origin/repos/acme/rocket/grants", grantBody)
		if rec.Code != http.StatusOK {
			t.Fatalf("upsert grant expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// GET (list grants)
		rec = doReq(http.MethodGet, "/v1/origin/repos/acme/rocket/grants", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("list grants expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var listResp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
		grants, ok := listResp["grants"].([]any)
		if !ok || len(grants) != 1 {
			t.Fatalf("expected 1 grant, got: %+v", listResp)
		}

		// DELETE (remove grant)
		delBody := map[string]any{
			"user": map[string]any{
				"id": userID,
			},
		}
		rec = doReq(http.MethodDelete, "/v1/origin/repos/acme/rocket/grants", delBody)
		if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
			t.Fatalf("delete grant expected 204 or 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// -------------------------------------------------------------
	// Resource 3: cursor_origin_repo_ruleset
	// -------------------------------------------------------------
	var rulesetID string
	t.Run("cursor_origin_repo_ruleset CRUD", func(t *testing.T) {
		// POST (create ruleset)
		rsBody := map[string]any{
			"name":        "Protect Main Branch",
			"description": "Require status checks and approvals",
			"enforcement": "active",
			"kind":        "merge_branch",
			"includedRefNames": []string{
				"refs/heads/main",
			},
			"rules": []map[string]any{
				{
					"ruleType": "pull_request_review_gate",
					"parameters": map[string]any{
						"requiredApprovals": 1,
					},
				},
			},
		}
		rec := doReq(http.MethodPost, "/v1/origin/repos/acme/rocket/rulesets", rsBody)
		if rec.Code != http.StatusOK {
			t.Fatalf("create ruleset expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var rsResp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &rsResp)
		rulesetID = rsResp["id"].(string)
		if rulesetID == "" {
			t.Fatalf("expected non-empty ruleset ID: %+v", rsResp)
		}

		// GET (list rulesets)
		rec = doReq(http.MethodGet, "/v1/origin/repos/acme/rocket/rulesets", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("list rulesets expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// GET (view ruleset by id)
		rec = doReq(http.MethodGet, "/v1/origin/repos/acme/rocket/rulesets/"+rulesetID, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("get ruleset expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// DELETE (delete ruleset)
		rec = doReq(http.MethodDelete, "/v1/origin/repos/acme/rocket/rulesets/"+rulesetID, nil)
		if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
			t.Fatalf("delete ruleset expected 204 or 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// -------------------------------------------------------------
	// Resource 4: cursor_origin_ssh_certificate_authority
	// -------------------------------------------------------------
	t.Run("cursor_origin_ssh_certificate_authority CRUD", func(t *testing.T) {
		caBody := map[string]any{
			"name":      "Corp SSH CA",
			"publicKey": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIInJ8YxG668jK9+k5kU8vGfK9w2Z3mR8aX7n0p2y9z8Q corp-ca",
		}
		rec := doReq(http.MethodPost, "/v1/origin/namespaces/acme/ssh-certificate-authorities", caBody)
		if rec.Code != http.StatusOK {
			t.Fatalf("create SSH CA expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var caResp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &caResp)
		caID := caResp["id"].(string)

		rec = doReq(http.MethodGet, "/v1/origin/namespaces/acme/ssh-certificate-authorities", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("list SSH CAs expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		rec = doReq(http.MethodDelete, "/v1/origin/namespaces/acme/ssh-certificate-authorities/"+caID, nil)
		if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
			t.Fatalf("delete SSH CA expected 204 or 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// -------------------------------------------------------------
	// Resource 5: cursor_origin_inbound_ip_allowlist
	// -------------------------------------------------------------
	t.Run("cursor_origin_inbound_ip_allowlist CRUD", func(t *testing.T) {
		entryBody := map[string]any{
			"cidr":        "192.0.2.0/24",
			"description": "Office VPN Gateway",
			"enabled":     true,
		}
		rec := doReq(http.MethodPost, "/v1/origin/namespaces/acme/inbound-ip-allowlist/entries", entryBody)
		if rec.Code != http.StatusOK {
			t.Fatalf("create IP allowlist entry expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var entryResp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &entryResp)
		entryID := entryResp["id"].(string)

		rec = doReq(http.MethodGet, "/v1/origin/namespaces/acme/inbound-ip-allowlist", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("get IP allowlist expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		rec = doReq(http.MethodDelete, "/v1/origin/namespaces/acme/inbound-ip-allowlist/entries/"+entryID, nil)
		if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
			t.Fatalf("delete IP allowlist entry expected 204 or 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}
