package main

import (
	"fmt"
	"net/http"

	"exist.io/internal/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	fmt.Printf("existd listening on %s (endpoint=%s gitHost=%s dataDir=%s)\n",
		cfg.ListenAddr, cfg.Endpoint, cfg.GitHost, cfg.DataDir)

	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		panic(err)
	}
}
