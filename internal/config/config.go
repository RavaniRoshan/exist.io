// Package config manages runtime configuration for exist.io services,
// reading environment variables and providing self-hosted defaults.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultListen    = "127.0.0.1:8080"
	defaultPublicURL = "http://127.0.0.1:8080"
	defaultGitURL    = "127.0.0.1:8080"
	defaultDataDir   = "data"

	minJWTSecretBytes = 32
)

// Config holds runtime configuration for the exist.io server and CLI.
// Defaults are strictly self-hosted and never reference Cursor infrastructure.
type Config struct {
	Version    string
	ListenAddr string
	DataDir    string
	PublicURL  string
	GitURL     string

	// Endpoint and GitHost are compatibility aliases for PublicURL and GitURL.
	Endpoint string
	GitHost  string

	SSHAddr        string
	SSHHostKeyPath string
	JWTSecret      string

	ClusterRole string
	ReplicaDir  string
	NodeID      string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPTLS      bool

	TelemetryEnabled  bool
	TelemetryEndpoint string
}

// Load reads configuration from environment variables, falling back to safe self-hosted defaults.
// The lease signing key is resolved from EXIST_JWT_SECRET or generated once under the data directory.
func Load() (Config, error) {
	listenAddr := getEnv("EXIST_LISTEN_ADDR", defaultListen)
	dataDir := getEnv("EXIST_DATA_DIR", defaultDataDir)

	publicURL := getEnvFirst([]string{"EXIST_PUBLIC_URL", "EXIST_ENDPOINT"}, defaultPublicURL)
	gitURL := getEnvFirst([]string{"EXIST_GIT_URL", "EXIST_GIT_HOST"}, defaultGitURL)
	sshAddr := getEnv("EXIST_SSH_ADDR", "127.0.0.1:2222")
	sshHostKeyPath := getEnv("EXIST_SSH_HOST_KEY", dataDir+"/ssh_host_key")

	jwtSecret, err := resolveJWTSecret(dataDir)
	if err != nil {
		return Config{}, err
	}

	clusterRole := getEnv("EXIST_CLUSTER_ROLE", "primary")
	replicaDir := getEnv("EXIST_REPLICA_DIR", "")
	nodeID := getEnv("EXIST_NODE_ID", "")

	smtpHost := getEnv("EXIST_SMTP_HOST", "")
	smtpPort := 587
	if pStr := getEnv("EXIST_SMTP_PORT", ""); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			smtpPort = p
		}
	}
	smtpUsername := getEnv("EXIST_SMTP_USERNAME", "")
	smtpPassword := getEnv("EXIST_SMTP_PASSWORD", "")
	smtpFrom := getEnv("EXIST_SMTP_FROM", "noreply@exist.io")
	smtpTLS := getEnv("EXIST_SMTP_TLS", "true") != "false"

	version := getEnv("EXIST_VERSION", "dev")

	telemetryEnabled := getEnv("EXIST_TELEMETRY_OPT_IN", "") == "1"
	if !telemetryEnabled {
		if body, err := os.ReadFile(filepath.Join(dataDir, "telemetry_opt_in")); err == nil { // #nosec G304 -- operator-configured data directory
			telemetryEnabled = strings.TrimSpace(string(body)) == "1"
		}
	}
	telemetryEndpoint := getEnv("EXIST_TELEMETRY_ENDPOINT", "")

	return Config{
		Version:        version,
		ListenAddr:     listenAddr,
		DataDir:        dataDir,
		PublicURL:      publicURL,
		GitURL:         gitURL,
		Endpoint:       publicURL,
		GitHost:        gitURL,
		SSHAddr:        sshAddr,
		SSHHostKeyPath: sshHostKeyPath,
		JWTSecret:      jwtSecret,
		ClusterRole:    clusterRole,
		ReplicaDir:     replicaDir,
		NodeID:         nodeID,
		SMTPHost:       smtpHost,
		SMTPPort:       smtpPort,
		SMTPUsername:   smtpUsername,
		SMTPPassword:   smtpPassword,
		SMTPFrom:       smtpFrom,
		SMTPTLS:        smtpTLS,

		TelemetryEnabled:  telemetryEnabled,
		TelemetryEndpoint: telemetryEndpoint,
	}, nil
}

// resolveJWTSecret returns the lease signing key, generating and persisting one
// under dataDir on first use. It never falls back to a literal in the source.
func resolveJWTSecret(dataDir string) (string, error) {
	if v := os.Getenv("EXIST_JWT_SECRET"); v != "" {
		if len(v) < minJWTSecretBytes {
			return "", fmt.Errorf("EXIST_JWT_SECRET must be at least %d bytes (got %d)", minJWTSecretBytes, len(v))
		}
		return v, nil
	}

	path := filepath.Join(dataDir, "jwt_secret")
	if body, err := os.ReadFile(path); err == nil { // #nosec G304 -- path comes from operator configuration or a validated internal identifier
		secret := strings.TrimSpace(string(body))
		if len(secret) >= minJWTSecretBytes {
			return secret, nil
		}
	}

	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", fmt.Errorf("create data directory %s: %w", dataDir, err)
	}

	buf := make([]byte, minJWTSecretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate lease signing key: %w", err)
	}
	secret := hex.EncodeToString(buf)

	if err := os.WriteFile(path, []byte(secret), 0o600); err != nil {
		return "", fmt.Errorf("write lease signing key %s: %w", path, err)
	}
	log.Printf("generated lease signing key at %s", path)
	return secret, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvFirst(keys []string, fallback string) string {
	for _, key := range keys {
		if v, ok := os.LookupEnv(key); ok && v != "" {
			return v
		}
	}
	return fallback
}
