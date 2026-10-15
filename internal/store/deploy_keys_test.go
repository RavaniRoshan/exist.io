package store_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"golang.org/x/crypto/ssh"
)

func generateTestSSHKey(t *testing.T) string {
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

func TestDeployKeyLifecycle(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open store: %v", err)
	}
	defer db.Close()

	ctx := t.Context()
	ns := store.Namespace{ID: idgen.NewNS(), Slug: "corp", Type: "team", Name: "Corporate"}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "prod-service",
		FullName:      "corp/prod-service",
		DefaultBranch: "main",
		Visibility:    "private",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	key1Str := generateTestSSHKey(t)
	key2Str := generateTestSSHKey(t)

	// 1. Create read-only deploy key
	k1, err := db.CreateDeployKey(ctx, repo.ID, "CI Runner Read-Only", key1Str, true)
	if err != nil {
		t.Fatalf("CreateDeployKey read-only: %v", err)
	}
	if !k1.ReadOnly {
		t.Errorf("expected k1 to be read-only")
	}
	if k1.Fingerprint == "" {
		t.Errorf("expected non-empty fingerprint")
	}

	// 2. Create write deploy key
	k2, err := db.CreateDeployKey(ctx, repo.ID, "CD Release Bot", key2Str, false)
	if err != nil {
		t.Fatalf("CreateDeployKey write: %v", err)
	}
	if k2.ReadOnly {
		t.Errorf("expected k2 to have write access")
	}

	// 3. List deploy keys
	keys, err := db.ListDeployKeys(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListDeployKeys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 deploy keys, got %d", len(keys))
	}

	// 4. Get by fingerprint
	foundKey, foundRepo, err := db.GetDeployKeyByFingerprint(ctx, k1.Fingerprint)
	if err != nil {
		t.Fatalf("GetDeployKeyByFingerprint: %v", err)
	}
	if foundKey.ID != k1.ID {
		t.Errorf("foundKey.ID = %s, want %s", foundKey.ID, k1.ID)
	}
	if foundRepo.ID != repo.ID {
		t.Errorf("foundRepo.ID = %s, want %s", foundRepo.ID, repo.ID)
	}

	// 5. Touch last used
	if err := db.TouchDeployKey(ctx, k1.ID); err != nil {
		t.Fatalf("TouchDeployKey: %v", err)
	}
	updatedK1, err := db.GetDeployKey(ctx, k1.ID)
	if err != nil {
		t.Fatalf("GetDeployKey: %v", err)
	}
	if updatedK1.LastUsedAt == "" {
		t.Errorf("expected updated LastUsedAt timestamp")
	}

	// 6. Delete deploy key
	if err := db.DeleteDeployKey(ctx, k2.ID); err != nil {
		t.Fatalf("DeleteDeployKey: %v", err)
	}
	remaining, err := db.ListDeployKeys(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListDeployKeys after delete: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining key, got %d", len(remaining))
	}
}
