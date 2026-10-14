package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"exist.io/internal/auth"
	"exist.io/internal/config"
	"exist.io/internal/gitstore"
	"exist.io/internal/gittransport"
	"exist.io/internal/mcp"
	"exist.io/internal/openapi"
	"exist.io/internal/origin"
	"exist.io/internal/store"
)

// App is the composition root for exist.io. It owns configuration, metadata
// storage, Git object storage, Origin compatibility routes, and server lifecycle.
type App struct {
	Config       config.Config
	Store        *store.DB
	Git          *gitstore.Store
	Origin       *origin.Server
	GitTransport *gittransport.Server

	handler http.Handler
}

// Handler returns the root HTTP handler for the server.
func (a *App) Handler() http.Handler {
	return a.handler
}

// Open initializes all subsystems according to cfg and returns the ready-to-serve App.
// If initialization fails at any stage, any resources already opened are closed cleanly.
func Open(cfg config.Config) (*App, error) {
	var dbPath string
	if cfg.DataDir == ":memory:" {
		dbPath = ":memory:"
	} else {
		if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
			return nil, fmt.Errorf("app: create data dir %s: %w", cfg.DataDir, err)
		}
		dbPath = filepath.Join(cfg.DataDir, "meta.db")
	}

	db, err := store.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("app: open store: %w", err)
	}

	var gitDir string
	if cfg.DataDir == ":memory:" {
		tmpDir, err := os.MkdirTemp("", "exist-git-*")
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("app: temp git dir: %w", err)
		}
		gitDir = tmpDir
	} else {
		gitDir = filepath.Join(cfg.DataDir, "git")
	}

	git, err := gitstore.New(gitDir)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("app: init gitstore: %w", err)
	}

	registry := auth.NewRegistry()
	originServer := origin.New(db, git, registry, cfg.PublicURL, cfg.GitURL)
	originServer.SetResolver(auth.NewDatabaseResolver(db))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("POST /auth/exchange_user_api_key", auth.ExchangeHandler(db))

	originServer.RegisterRoutes(mux)

	mux.Handle("GET /v1/origin/openapi.yaml", openapi.Handler())

	mcpServer := mcp.New(db, git)
	mux.Handle("/mcp", mcpServer.Handler(false))
	mux.Handle("/mcp/readonly", mcpServer.Handler(true))

	gitTransport := gittransport.NewServer(db, git)
	mux.Handle("/", gitTransport)

	return &App{
		Config:       cfg,
		Store:        db,
		Git:          git,
		Origin:       originServer,
		GitTransport: gitTransport,
		handler:      mux,
	}, nil
}

// Close shuts down all held resources.
func (a *App) Close() error {
	if a.Store != nil {
		return a.Store.Close()
	}
	return nil
}
