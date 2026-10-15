package symbolsearch

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"exist.io/internal/gitstore"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func author() object.Signature {
	return object.Signature{
		Name:  "Test Dev",
		Email: "dev@exist.io",
		When:  time.Now(),
	}
}

func TestExtractGoSymbols(t *testing.T) {
	src := []byte(`package sample

import "fmt"

// Config represents application settings
type Config struct {
	Port int
	Host string
}

type Runner interface {
	Run() error
}

const DefaultPort = 8080

func NewConfig(port int) *Config {
	return &Config{Port: port}
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
`)

	syms, err := ExtractGoSymbols("sample.go", src)
	if err != nil {
		t.Fatalf("ExtractGoSymbols: %v", err)
	}

	foundStruct := false
	foundInterface := false
	foundFunc := false
	foundMethod := false
	foundConst := false

	for _, s := range syms {
		switch s.Name {
		case "Config":
			if s.Kind == "struct" {
				foundStruct = true
			}
		case "Runner":
			if s.Kind == "interface" {
				foundInterface = true
			}
		case "NewConfig":
			if s.Kind == "function" {
				foundFunc = true
			}
		case "Address":
			if s.Kind == "method" && s.Container == "*Config" {
				foundMethod = true
			}
		case "DefaultPort":
			if s.Kind == "constant" {
				foundConst = true
			}
		}
	}

	if !foundStruct {
		t.Error("expected to find struct Config")
	}
	if !foundInterface {
		t.Error("expected to find interface Runner")
	}
	if !foundFunc {
		t.Error("expected to find function NewConfig")
	}
	if !foundMethod {
		t.Error("expected to find method Address on *Config")
	}
	if !foundConst {
		t.Error("expected to find constant DefaultPort")
	}
}

func TestExtractTSSymbols(t *testing.T) {
	src := []byte(`export interface UserProfile {
  id: string;
  name: string;
}

export type Role = 'admin' | 'member';

export class AuthManager {
  login() {}
}

export function verifyToken(token: string): boolean {
  return true;
}

export const useAuth = () => {
  return {};
};
`)

	syms := ExtractTSSymbols("auth.ts", src)
	names := make(map[string]string)
	for _, s := range syms {
		names[s.Name] = s.Kind
	}

	if names["UserProfile"] != "interface" {
		t.Errorf("expected UserProfile interface, got %q", names["UserProfile"])
	}
	if names["Role"] != "type" {
		t.Errorf("expected Role type, got %q", names["Role"])
	}
	if names["AuthManager"] != "class" {
		t.Errorf("expected AuthManager class, got %q", names["AuthManager"])
	}
	if names["verifyToken"] != "function" {
		t.Errorf("expected verifyToken function, got %q", names["verifyToken"])
	}
	if names["useAuth"] != "function" {
		t.Errorf("expected useAuth function, got %q", names["useAuth"])
	}
}

func TestSymbolSearch_Tree(t *testing.T) {
	ctx := context.Background()
	gs, err := gitstore.New(filepath.Join(t.TempDir(), "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	owner, repo := "origin", "forge"
	if _, err := gs.Init(owner, repo, "main"); err != nil {
		t.Fatalf("gs.Init: %v", err)
	}

	c1, err := gs.WriteFiles(owner, repo, "main", []gitstore.FileChange{
		{
			Path: "server/server.go",
			Content: `package server

type OriginServer struct {
	Addr string
}

func NewServer(addr string) *OriginServer {
	return &OriginServer{Addr: addr}
}
`,
		},
		{
			Path: "web/src/client.ts",
			Content: `export class ForgeClient {
  constructor(public baseUrl: string) {}
}

export function connect(): ForgeClient {
  return new ForgeClient('http://localhost');
}
`,
		},
	}, author(), "initial commit")
	if err != nil {
		t.Fatalf("commit 1: %v", err)
	}

	// Search for "Server"
	results, err := Search(ctx, gs, owner, repo, c1.String(), "Server", "", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results for 'Server', got %d", len(results))
	}

	// Filter by kind "class"
	classResults, err := Search(ctx, gs, owner, repo, c1.String(), "ForgeClient", "class", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(classResults) != 1 || classResults[0].Name != "ForgeClient" {
		t.Fatalf("expected 1 class result 'ForgeClient', got %+v", classResults)
	}
}
