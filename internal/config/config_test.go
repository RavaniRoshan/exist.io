package config

import (
	"strings"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Load()
	if strings.Contains(cfg.PublicURL, "cursor.com") {
		t.Errorf("self-hosted public URL must not default to Cursor: %s", cfg.PublicURL)
	}
	if strings.Contains(cfg.GitURL, "cursor.com") {
		t.Errorf("self-hosted git URL must not default to Cursor: %s", cfg.GitURL)
	}
	if cfg.PublicURL != "http://127.0.0.1:8080" {
		t.Errorf("publicURL = %s, want http://127.0.0.1:8080", cfg.PublicURL)
	}
	if cfg.GitURL != "127.0.0.1:8080" {
		t.Errorf("gitURL = %s, want 127.0.0.1:8080", cfg.GitURL)
	}
	if cfg.ListenAddr != "127.0.0.1:8080" {
		t.Errorf("listenAddr = %s, want 127.0.0.1:8080", cfg.ListenAddr)
	}
	if cfg.DataDir != "data" {
		t.Errorf("dataDir = %s, want data", cfg.DataDir)
	}
	// Aliases for compatibility
	if cfg.Endpoint != cfg.PublicURL {
		t.Errorf("endpoint alias mismatch: got %s, want %s", cfg.Endpoint, cfg.PublicURL)
	}
	if cfg.GitHost != cfg.GitURL {
		t.Errorf("gitHost alias mismatch: got %s, want %s", cfg.GitHost, cfg.GitURL)
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	t.Setenv("EXIST_PUBLIC_URL", "https://forge.example.com")
	t.Setenv("EXIST_GIT_URL", "git.example.com")
	t.Setenv("EXIST_LISTEN_ADDR", "0.0.0.0:9090")
	t.Setenv("EXIST_DATA_DIR", "/tmp/exist-data")

	cfg := Load()
	if cfg.PublicURL != "https://forge.example.com" {
		t.Errorf("publicURL = %s", cfg.PublicURL)
	}
	if cfg.GitURL != "git.example.com" {
		t.Errorf("gitURL = %s", cfg.GitURL)
	}
	if cfg.ListenAddr != "0.0.0.0:9090" {
		t.Errorf("listenAddr = %s", cfg.ListenAddr)
	}
	if cfg.DataDir != "/tmp/exist-data" {
		t.Errorf("dataDir = %s", cfg.DataDir)
	}
}

func TestConfigLegacyEnvOverrides(t *testing.T) {
	t.Setenv("EXIST_ENDPOINT", "https://legacy.example.com")
	t.Setenv("EXIST_GIT_HOST", "legacy-git.example.com")

	cfg := Load()
	if cfg.PublicURL != "https://legacy.example.com" {
		t.Errorf("publicURL from legacy EXIST_ENDPOINT = %s", cfg.PublicURL)
	}
	if cfg.GitURL != "legacy-git.example.com" {
		t.Errorf("gitURL from legacy EXIST_GIT_HOST = %s", cfg.GitURL)
	}
}
