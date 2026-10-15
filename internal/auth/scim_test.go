package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"exist.io/internal/store"
)

func TestSCIM2FullLifecycle(t *testing.T) {
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Create a SCIM bearer token
	scimTokenRaw := "scim_secret_token_123"
	err = db.CreateSCIMToken(ctx, store.SCIMToken{
		ID:          "token_scim",
		TokenHash:   HashCredential(scimTokenRaw),
		Description: "Okta SCIM Integration",
	})
	if err != nil {
		t.Fatalf("CreateSCIMToken: %v", err)
	}

	mux := http.NewServeMux()
	RegisterSCIMRoutes(mux, db)

	doRequest := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+scimTokenRaw)
		req.Header.Set("Content-Type", SCIMMediaType)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	// 2. Service Provider Config (public/unauthenticated)
	spReq := httptest.NewRequest(http.MethodGet, "/scim/v2/ServiceProviderConfig", nil)
	spRec := httptest.NewRecorder()
	mux.ServeHTTP(spRec, spReq)
	if spRec.Code != http.StatusOK {
		t.Fatalf("ServiceProviderConfig status = %d, want 200", spRec.Code)
	}

	// 3. Create User via SCIM POST
	createBody := `{
		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
		"userName": "sarah.connor@cyberdyne.com",
		"name": {
			"formatted": "Sarah Connor",
			"familyName": "Connor",
			"givenName": "Sarah"
		},
		"emails": [
			{"value": "sarah.connor@cyberdyne.com", "primary": true, "type": "work"}
		],
		"active": true
	}`
	createRec := doRequest(http.MethodPost, "/scim/v2/Users", createBody)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("createUser status = %d, want 201: %s", createRec.Code, createRec.Body.String())
	}

	var createdUser SCIMUser
	if err := json.Unmarshal(createRec.Body.Bytes(), &createdUser); err != nil {
		t.Fatalf("unmarshal created user: %v", err)
	}
	if createdUser.ID == "" || createdUser.UserName != "sarah.connor@cyberdyne.com" {
		t.Fatalf("unexpected created user: %+v", createdUser)
	}

	// 4. Query Users with filter
	listRec := doRequest(http.MethodGet, `/scim/v2/Users?filter=userName+eq+"sarah.connor@cyberdyne.com"`, "")
	if listRec.Code != http.StatusOK {
		t.Fatalf("listUsers status = %d, want 200", listRec.Code)
	}
	var listResp SCIMListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}
	if listResp.TotalResults != 1 || len(listResp.Resources) != 1 {
		t.Fatalf("expected 1 result in filter query, got %d", listResp.TotalResults)
	}

	// 5. Get User by ID
	getRec := doRequest(http.MethodGet, "/scim/v2/Users/"+createdUser.ID, "")
	if getRec.Code != http.StatusOK {
		t.Fatalf("getUser status = %d, want 200", getRec.Code)
	}

	// 6. Patch User to active: false (Deprovision / Suspend)
	patchBody := `{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations": [
			{
				"op": "replace",
				"path": "active",
				"value": false
			}
		]
	}`
	patchRec := doRequest(http.MethodPatch, "/scim/v2/Users/"+createdUser.ID, patchBody)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("patchUser status = %d, want 200", patchRec.Code)
	}
	var patchedUser SCIMUser
	_ = json.Unmarshal(patchRec.Body.Bytes(), &patchedUser)
	if patchedUser.Active {
		t.Fatalf("expected active=false after patch, got true")
	}

	// Verify in database that status is now "suspended"
	dbUser, err := db.GetUser(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("db.GetUser: %v", err)
	}
	if dbUser.Status != "suspended" {
		t.Fatalf("expected user status 'suspended', got %s", dbUser.Status)
	}

	// 7. Delete User (Deprovision)
	delRec := doRequest(http.MethodDelete, "/scim/v2/Users/"+createdUser.ID, "")
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("deleteUser status = %d, want 204", delRec.Code)
	}
}
