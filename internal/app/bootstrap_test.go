package app_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"exist.io/internal/app"
	"exist.io/internal/config"
	"exist.io/internal/gitstore"
)

func TestBootstrap(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.Config{
		DataDir:    tmpDir,
		ListenAddr: "127.0.0.1:0",
		PublicURL:  "http://localhost:8080",
		GitURL:     "http://localhost:8080",
		JWTSecret:  "test-secret-key-32-bytes-long!",
	}

	ctx := context.Background()

	// 1. Successful bootstrap
	res, err := app.Bootstrap(ctx, cfg, "admin@exist.io", "acme")
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if res.User.Email != "admin@exist.io" {
		t.Errorf("User email = %q, want %q", res.User.Email, "admin@exist.io")
	}
	if res.Namespace.Slug != "acme" {
		t.Errorf("Namespace slug = %q, want %q", res.Namespace.Slug, "acme")
	}
	if res.Repo.FullName != "acme/demo" {
		t.Errorf("Repo fullName = %q, want %q", res.Repo.FullName, "acme/demo")
	}
	if !strings.HasPrefix(res.RawKey, "key_") {
		t.Errorf("API key = %q, want prefix key_", res.RawKey)
	}

	// Verify repo on disk
	repoStore, err := gitstore.New(filepath.Join(tmpDir, "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}
	if !repoStore.Exists("acme", "demo") {
		t.Errorf("Repository acme/demo does not exist on disk")
	}

	// 2. Second bootstrap on the same data dir must fail
	_, err = app.Bootstrap(ctx, cfg, "other@exist.io", "other")
	if err == nil {
		t.Fatal("Second bootstrap succeeded, expected error")
	}
	if !strings.Contains(err.Error(), "already initialized") {
		t.Errorf("Second bootstrap error = %q, want 'already initialized'", err.Error())
	}
}

func TestBootstrapValidation(t *testing.T) {
	cfg := config.Config{
		DataDir: t.TempDir(),
	}
	ctx := context.Background()

	if _, err := app.Bootstrap(ctx, cfg, "", "acme"); err == nil {
		t.Error("Bootstrap with empty email succeeded, want error")
	}
	if _, err := app.Bootstrap(ctx, cfg, "admin@exist.io", ""); err == nil {
		t.Error("Bootstrap with empty namespace succeeded, want error")
	}
}
