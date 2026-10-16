// Command existd is the primary server daemon for exist.io.
// It hosts the Smart-HTTP Git transport, REST API, Model Context Protocol (MCP) server,
// and supports offline administrator bootstrap.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"exist.io/internal/app"
	"exist.io/internal/config"
	"exist.io/internal/security"
)

func main() {
	cfg := config.Load()

	// Check for CLI subcommands
	if len(os.Args) > 1 && os.Args[1] == "hook" {
		if len(os.Args) > 2 && os.Args[2] == "pre-receive" {
			if err := security.PreReceiveHook(os.Stdin, os.Stderr); err != nil {
				os.Exit(1)
			}
			return
		}
		fmt.Fprintln(os.Stderr, "Usage: existd hook pre-receive")
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "bootstrap" {
		fs := flag.NewFlagSet("bootstrap", flag.ExitOnError)
		email := fs.String("email", "", "Admin user email address (required)")
		namespace := fs.String("namespace", "", "Initial admin namespace slug (required)")
		dataDir := fs.String("data-dir", cfg.DataDir, "Path to exist.io data directory")
		if err := fs.Parse(os.Args[2:]); err != nil {
			log.Fatalf("bootstrap: %v", err)
		}

		if *email == "" || *namespace == "" {
			fmt.Fprintln(os.Stderr, "Usage: existd bootstrap --email <email> --namespace <slug> [--data-dir <path>]")
			os.Exit(1)
		}

		cfg.DataDir = *dataDir
		res, err := app.Bootstrap(context.Background(), cfg, *email, *namespace)
		if err != nil {
			log.Fatalf("bootstrap failed: %v", err)
		}

		fmt.Println("exist.io bootstrap complete!")
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Printf("User:       %s (id: %s)\n", res.User.Email, res.User.ID)
		fmt.Printf("Namespace:  %s (id: %s)\n", res.Namespace.Slug, res.Namespace.ID)
		fmt.Printf("Repository: %s\n", res.Repo.FullName)
		fmt.Printf("API Key:    %s\n", res.RawKey)
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println("IMPORTANT: Store this API key securely. It will not be shown again.")
		fmt.Println()
		fmt.Println("=== Agent Configuration Snippets ===")
		fmt.Println()
		fmt.Println("Claude Code (~/.claude/settings.json or command line):")
		fmt.Printf("  claude mcp add exist %s/mcp --header \"Authorization: Bearer %s\"\n\n", cfg.PublicURL, res.RawKey)
		fmt.Println("  Or add to ~/.claude/settings.json:")
		fmt.Printf(`  {
    "mcpServers": {
      "exist": {
        "url": "%s/mcp",
        "headers": {
          "Authorization": "Bearer %s"
        }
      }
    }
  }`+"\n", cfg.PublicURL, res.RawKey)
		fmt.Println()
		fmt.Println("Cursor (.cursor/mcp.json):")
		fmt.Printf(`  {
    "mcpServers": {
      "exist": {
        "url": "%s/mcp",
        "headers": {
          "Authorization": "Bearer %s"
        }
      }
    }
  }`+"\n", cfg.PublicURL, res.RawKey)
		fmt.Println()
		fmt.Println("Cline (cline_mcp_settings.json):")
		fmt.Printf(`  {
    "mcpServers": {
      "exist": {
        "url": "%s/mcp",
        "headers": {
          "Authorization": "Bearer %s"
        },
        "disabled": false,
        "autoApprove": []
      }
    }
  }`+"\n", cfg.PublicURL, res.RawKey)
		fmt.Println()
		fmt.Println("Codex (TOML):")
		fmt.Printf(`  [mcp.servers.exist]
  url = "%s/mcp"
  headers = { Authorization = "Bearer %s" }`+"\n", cfg.PublicURL, res.RawKey)
		return
	}

	application, err := app.Open(cfg)
	if err != nil {
		log.Fatalf("existd: failed to initialize application: %v", err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("existd: error closing application: %v", err)
		}
	}()

	if cfg.SSHAddr != "" && cfg.SSHAddr != "none" {
		if err := application.StartSSH(); err != nil {
			log.Printf("existd: failed to start SSH server on %s: %v", cfg.SSHAddr, err)
		} else {
			fmt.Printf("existd SSH listening on %s\n", cfg.SSHAddr)
		}
	}

	server := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: application.Handler(),
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("existd listening on %s (publicURL=%s gitURL=%s dataDir=%s sshAddr=%s)\n",
			cfg.ListenAddr, cfg.PublicURL, cfg.GitURL, cfg.DataDir, cfg.SSHAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		log.Fatalf("existd: server error: %v", err)
	case sig := <-shutdown:
		fmt.Printf("\nexistd: received signal %v, shutting down gracefully...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("existd: shutdown error: %v", err)
		}
	}
}
