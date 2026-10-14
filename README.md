# exist.io

**An open-source, self-hosted alternative to Cursor Origin—a Git forge for people and coding agents.**

Product documentation: **[exist.io docs](https://ravaniroshan.github.io/exist.io/)**.

exist.io is built for teams and autonomous coding agents who need repositories, collaboration workflows, policy enforcement, and agent integrations on infrastructure they control. It provides full wire compatibility with Origin interfaces while remaining 100% self-hosted, independent, and free of proprietary cloud services.

## Core Capabilities

- **Git Smart-HTTP Transport:** Pure Go Git transport (`git clone`, `git fetch`, `git push`) over HTTP with credential-helper and basic authentication support.
- **Full API Surface:** 100% coverage of the 119 Origin REST operations spanning repositories, namespaces, branches, commits, trees, diffs, pull requests, review threads, reactions, labels, status checks, and rulesets.
- **Agent Integration via MCP:** Built-in Model Context Protocol (MCP) server mounted at `/mcp` (read-write) and `/mcp/readonly` (read-only) exposing 34 specialized tools for AI coding agents.
- **Pull Request Collaboration & Merge Engine:** Native 3-way merge engine, squash merge, conflict detection, review approvals gate, required checks enforcement, and auto-merge queue.
- **Enterprise Controls & Security:** Branch and tag rulesets, bypass actors, inbound IP allowlists, SSH Certificate Authorities, and fine-grained repository/namespace grants.
- **Mirroring & Webhooks:** GitHub mirror transition and synchronization, app registrations, Ed25519 webhook signing, and redelivery queue.
- **Origin CLI:** Compatible command-line tool (`origin`) in `cmd/origin` for managing repos, pull requests, mirrors, apps, and credentials.
- **OpenAPI 3.1 Publication:** Dynamic and embedded OpenAPI 3.1 specification served at `/v1/origin/openapi.yaml`.
- **Pure Go Architecture:** Built on pure Go (`go-git/v6` and `modernc.org/sqlite` with WAL mode and foreign keys enabled), compiling to a single static binary without cgo.

## Getting Started

### Prerequisites

- Go 1.22+ (tested through Go 1.26)
- Git 2.30+

### Installation & Build

Build the server (`existd`) and CLI (`origin`):

```bash
# Build the forge server daemon
go build -o bin/existd ./cmd/existd

# Build the compatibility CLI
go build -o bin/origin ./cmd/origin
```

### Bootstrap First Admin

Initialize the database, create the primary administrator account, and issue your root API key:

```bash
./bin/existd bootstrap --email admin@example.com --namespace acme --data-dir ./data
```

### Start Server

Launch the daemon:

```bash
# Run exist.io daemon
./bin/existd
```

The server binds to `http://localhost:8080` by default.

- REST API: `http://localhost:8080/v1/origin`
- MCP Server: `http://localhost:8080/mcp` (and `http://localhost:8080/mcp/readonly`)
- OpenAPI Spec: `http://localhost:8080/v1/origin/openapi.yaml`
- Git Remotes: `http://localhost:8080/{owner}/{repo}.git`
- Health Probe: `http://localhost:8080/healthz`

### Using the CLI

```bash
export CURSOR_ORIGIN_ENDPOINT="http://localhost:8080/v1/origin"
export CURSOR_AUTH_TOKEN="<your-api-key>"

# Create repository
./bin/origin repo create acme/my-project

# Open a pull request
./bin/origin pr create -R acme/my-project --title "Add new feature" --head feat --base main

# Inspect MCP tools
curl -X POST http://localhost:8080/mcp -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

## Running Tests

All packages are covered with automated contract and unit tests:

```bash
go test -count=1 ./...
go vet ./...
```

## Technology Stack

- **Language:** Go
- **Storage:** SQLite (`modernc.org/sqlite`, zero cgo)
- **Git Engine:** `github.com/go-git/go-git/v6`
- **ID Generation:** Monotonic ULID (`github.com/oklog/ulid`)

## License

Apache 2.0 / MIT
