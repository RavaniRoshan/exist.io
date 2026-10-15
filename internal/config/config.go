// Package config manages runtime configuration for exist.io services,
// reading environment variables and providing self-hosted defaults.
package config

import "os"

const (
	defaultListen    = "127.0.0.1:8080"
	defaultPublicURL = "http://127.0.0.1:8080"
	defaultGitURL    = "127.0.0.1:8080"
	defaultDataDir   = "data"
)

// Config holds runtime configuration for the exist.io server and CLI.
// Defaults are strictly self-hosted and never reference Cursor infrastructure.
type Config struct {
	ListenAddr string
	DataDir    string
	PublicURL  string
	GitURL     string

	// Endpoint and GitHost are compatibility aliases for PublicURL and GitURL.
	Endpoint string
	GitHost  string

	SSHAddr        string
	SSHHostKeyPath string
}

// Load reads configuration from environment variables, falling back to safe self-hosted defaults.
func Load() Config {
	listenAddr := getEnv("EXIST_LISTEN_ADDR", defaultListen)
	dataDir := getEnv("EXIST_DATA_DIR", defaultDataDir)

	publicURL := getEnvFirst([]string{"EXIST_PUBLIC_URL", "EXIST_ENDPOINT"}, defaultPublicURL)
	gitURL := getEnvFirst([]string{"EXIST_GIT_URL", "EXIST_GIT_HOST"}, defaultGitURL)
	sshAddr := getEnv("EXIST_SSH_ADDR", "127.0.0.1:2222")
	sshHostKeyPath := getEnv("EXIST_SSH_HOST_KEY", dataDir+"/ssh_host_key")

	return Config{
		ListenAddr:     listenAddr,
		DataDir:        dataDir,
		PublicURL:      publicURL,
		GitURL:         gitURL,
		Endpoint:       publicURL,
		GitHost:        gitURL,
		SSHAddr:        sshAddr,
		SSHHostKeyPath: sshHostKeyPath,
	}
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
