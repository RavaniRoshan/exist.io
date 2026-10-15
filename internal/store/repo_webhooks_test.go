package store_test

import (
	"testing"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

func TestRepoWebhookLifecycle(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open store: %v", err)
	}
	defer db.Close()

	ctx := t.Context()
	ns := store.Namespace{ID: idgen.NewNS(), Slug: "webhookorg", Type: "team", Name: "Webhook Org"}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "web-app",
		FullName:      "webhookorg/web-app",
		DefaultBranch: "main",
		Visibility:    "public",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// 1. Create webhook
	hook, err := db.CreateRepoWebhook(ctx, repo.ID, "https://ci.example.com/webhook", "supersecret123", "json", `["push", "pull_request"]`, true, false)
	if err != nil {
		t.Fatalf("CreateRepoWebhook: %v", err)
	}
	if hook.ID == "" {
		t.Errorf("expected non-empty webhook ID")
	}
	if hook.URL != "https://ci.example.com/webhook" {
		t.Errorf("hook.URL = %s, want https://ci.example.com/webhook", hook.URL)
	}

	// 2. List webhooks
	hooks, err := db.ListRepoWebhooks(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListRepoWebhooks: %v", err)
	}
	if len(hooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(hooks))
	}

	// 3. Record delivery attempt
	del := &store.RepoWebhookDelivery{
		WebhookID:          hook.ID,
		Event:              "push",
		GUID:               "guid-12345",
		RequestHeaders:     `{"X-GitHub-Event": "push"}`,
		RequestPayload:     `{"ref": "refs/heads/main"}`,
		ResponseStatusCode: 200,
		ResponseHeaders:    `{"Content-Type": "application/json"}`,
		ResponseBody:       `{"status": "ok"}`,
		DurationMs:         45,
		Delivered:          true,
	}
	if err := db.RecordRepoWebhookDelivery(ctx, del); err != nil {
		t.Fatalf("RecordRepoWebhookDelivery: %v", err)
	}

	// 4. List deliveries
	deliveries, err := db.ListRepoWebhookDeliveries(ctx, hook.ID, 10)
	if err != nil {
		t.Fatalf("ListRepoWebhookDeliveries: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(deliveries))
	}
	if deliveries[0].ResponseStatusCode != 200 {
		t.Errorf("delivery status = %d, want 200", deliveries[0].ResponseStatusCode)
	}

	// 5. Get delivery
	fetchedDel, err := db.GetRepoWebhookDelivery(ctx, deliveries[0].ID)
	if err != nil {
		t.Fatalf("GetRepoWebhookDelivery: %v", err)
	}
	if fetchedDel.GUID != "guid-12345" {
		t.Errorf("fetchedDel.GUID = %s, want guid-12345", fetchedDel.GUID)
	}

	// 6. Delete webhook
	if err := db.DeleteRepoWebhook(ctx, hook.ID); err != nil {
		t.Fatalf("DeleteRepoWebhook: %v", err)
	}
	remainingHooks, err := db.ListRepoWebhooks(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListRepoWebhooks after delete: %v", err)
	}
	if len(remainingHooks) != 0 {
		t.Fatalf("expected 0 remaining hooks, got %d", len(remainingHooks))
	}
}
