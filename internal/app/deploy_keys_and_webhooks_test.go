package app

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exist.io/internal/config"
	"exist.io/internal/store"
	"golang.org/x/crypto/ssh"
)

func generateTestSSHAuthorizedKey(t *testing.T) string {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key: %v", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("new ssh public key: %v", err)
	}
	return string(ssh.MarshalAuthorizedKey(sshPub))
}

func TestDeployKeysAndRepoWebhooksAPI(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:0",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}

	application, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer application.Close()

	handler := application.Handler()
	db := application.Store
	ctx := t.Context()

	// 1. Create owner namespace and repository
	ns := store.Namespace{ID: "ns_deploy_test", Slug: "octocorp", Type: "team", Name: "OctoCorp"}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	repo := store.Repo{
		ID:            "repo_deploy_test",
		NamespaceID:   ns.ID,
		Name:          "payment-service",
		FullName:      "octocorp/payment-service",
		DefaultBranch: "main",
		Visibility:    "private",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// 2. Test GET Deploy Keys (should be empty array)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/repos/octocorp/payment-service/keys", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET keys status = %d, want 200", rr.Code)
	}
	var keys []*store.RepoDeployKey
	if err := json.Unmarshal(rr.Body.Bytes(), &keys); err != nil {
		t.Fatalf("unmarshal keys: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(keys))
	}

	// 3. Test POST Deploy Key
	testKey := generateTestSSHAuthorizedKey(t)
	keyPayload := map[string]any{
		"title":     "Production CI Deploy Key",
		"key":       testKey,
		"read_only": true,
	}
	b, _ := json.Marshal(keyPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/octocorp/payment-service/keys", bytes.NewReader(b))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("POST key status = %d, body: %s", rr.Code, rr.Body.String())
	}
	var createdKey store.RepoDeployKey
	if err := json.Unmarshal(rr.Body.Bytes(), &createdKey); err != nil {
		t.Fatalf("unmarshal created key: %v", err)
	}
	if createdKey.Title != "Production CI Deploy Key" {
		t.Errorf("createdKey.Title = %s, want Production CI Deploy Key", createdKey.Title)
	}
	if !createdKey.ReadOnly {
		t.Errorf("expected key to be read-only")
	}

	// 4. Test DELETE Deploy Key
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/repos/octocorp/payment-service/keys/"+createdKey.ID, nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE key status = %d", rr.Code)
	}

	// 5. Test POST Repo Webhook
	hookPayload := map[string]any{
		"url":          "https://ci.octocorp.internal/webhook",
		"content_type": "json",
		"secret":       "octosecret999",
		"events":       []string{"push", "pull_request"},
		"active":       true,
		"insecure_ssl": false,
	}
	b, _ = json.Marshal(hookPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/octocorp/payment-service/hooks", bytes.NewReader(b))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("POST hook status = %d, body: %s", rr.Code, rr.Body.String())
	}
	var createdHook store.RepoWebhook
	if err := json.Unmarshal(rr.Body.Bytes(), &createdHook); err != nil {
		t.Fatalf("unmarshal hook: %v", err)
	}
	if createdHook.URL != "https://ci.octocorp.internal/webhook" {
		t.Errorf("createdHook.URL = %s, want https://ci.octocorp.internal/webhook", createdHook.URL)
	}

	// 6. Test POST Ping Test Event on Webhook
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/octocorp/payment-service/hooks/"+createdHook.ID+"/tests", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST hook test status = %d", rr.Code)
	}

	// 7. Test GET Deliveries for Webhook
	req = httptest.NewRequest(http.MethodGet, "/api/v1/repos/octocorp/payment-service/hooks/"+createdHook.ID+"/deliveries", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET hook deliveries status = %d", rr.Code)
	}
	var deliveries []*store.RepoWebhookDelivery
	if err := json.Unmarshal(rr.Body.Bytes(), &deliveries); err != nil {
		t.Fatalf("unmarshal deliveries: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(deliveries))
	}
	if deliveries[0].Event != "ping" {
		t.Errorf("deliveries[0].Event = %s, want ping", deliveries[0].Event)
	}

	// 8. Test POST Redeliver
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/octocorp/payment-service/hooks/"+createdHook.ID+"/deliveries/"+deliveries[0].ID+"/redeliver", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST redeliver status = %d", rr.Code)
	}

	// 9. Test DELETE Webhook
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/repos/octocorp/payment-service/hooks/"+createdHook.ID, nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE hook status = %d", rr.Code)
	}
}
