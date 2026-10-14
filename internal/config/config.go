package config

import "os"

const (
	defaultEndpoint = "https://api.cursor.com"
	defaultGitHost  = "origin.cursor.com"
	defaultListen   = "127.0.0.1:8080"
)

type Config struct {
	ListenAddr string
	DataDir    string
	Endpoint   string
	GitHost    string
}

func Load() Config {
	return Config{
		ListenAddr: getEnv("EXIST_LISTEN_ADDR", defaultListen),
		DataDir:    getEnv("EXIST_DATA_DIR", "data"),
		Endpoint:   getEnv("EXIST_ENDPOINT", defaultEndpoint),
		GitHost:    getEnv("EXIST_GIT_HOST", defaultGitHost),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
