package main

import (
	"fmt"
	"net/http"

	"exist.io/internal/auth"
	"exist.io/internal/config"
	"exist.io/pkg/scopes"
)

func main() {
	cfg := config.Load()

	// The operation registry maps Origin operationIds to their x-origin-scopes
	// requirements. Later handler groups populate it incrementally; the auth
	// middleware enforces each operation against the acting credential.
	registry := auth.NewRegistry()
	registry.Register("OriginService_ListRepos", scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser},
		Scopes:     []string{scopes.NamespaceRepositoriesRead},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Example guarded route: enforces the registered operation's scopes before
	// dispatching to the (not-yet-implemented) handler.
	mux.Handle("/v1/origin/repos", auth.Require(registry, "OriginService_ListRepos",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `[]`)
		})))

	fmt.Printf("existd listening on %s (endpoint=%s gitHost=%s dataDir=%s)\n",
		cfg.ListenAddr, cfg.Endpoint, cfg.GitHost, cfg.DataDir)

	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		panic(err)
	}
}
