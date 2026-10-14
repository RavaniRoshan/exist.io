package config

import "testing"

func TestConfigDefaults(t *testing.T) {
	cfg := Load()
	if cfg.Endpoint != "https://api.cursor.com" {
		t.Errorf("endpoint = %s", cfg.Endpoint)
	}
	if cfg.GitHost != "origin.cursor.com" {
		t.Errorf("gitHost = %s", cfg.GitHost)
	}
	if cfg.ListenAddr == "" {
		t.Error("listenAddr is empty")
	}
	if cfg.DataDir == "" {
		t.Error("dataDir is empty")
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	t.Setenv("EXIST_ENDPOINT", "https://api.example.com")
	t.Setenv("EXIST_DATA_DIR", "/tmp/exist-data")

	cfg := Load()
	if cfg.Endpoint != "https://api.example.com" {
		t.Errorf("endpoint = %s", cfg.Endpoint)
	}
	if cfg.DataDir != "/tmp/exist-data" {
		t.Errorf("dataDir = %s", cfg.DataDir)
	}
}
