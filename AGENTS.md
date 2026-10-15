# exist.io — Autonomous Agent Operational Handbook (`AGENTS.md`)

<p align="center">
  <strong>Master Architecture, Operational Invariants, and Pair-Programming Runbook for AI Coding Agents.</strong>
</p>

---

## 1. Project Mission & Identity

**exist.io** is the next-generation, self-hosted Git forge engineered from first principles for both **human engineering teams** and **autonomous AI coding agents**.

Traditional Git platforms were built 15+ years ago for human browser interaction. exist.io bridges this generational shift by treating autonomous coding agents as first-class citizens alongside human developers:
- **Dual Transports:** Pure-Go Smart-HTTP (`:8080`) and standalone embedded pure-Go SSH server (`:2222`, zero cgo, zero host OpenSSH dependency).
- **Git LFS v1 CAS Engine:** Content-Addressable Storage with SHA-256 validation, streaming chunked transfer, and binary file locking protocol (`/info/lfs/locks`).
- **AI Agent Native Ecosystem:** Native Model Context Protocol (MCP) server at `/mcp` (34 tools) and sandboxed `/mcp/readonly`, ephemeral HMAC-SHA256 capability leases (`pkg/agentlease`) with ref-prefix isolation (`refs/heads/agent/*`), autonomous PR review agent (`internal/agentreviewer`), and pure-Go semantic AST symbol search (`internal/symbolsearch`).
- **Cryptographic Verification:** Real-time OpenSSH SIGSPEC and GPG (RFC 4880) commit signature parsing and fingerprint matching.
- **Enterprise Federation & IaC:** SCIM 2.0 user lifecycle sync (`/scim/v2/Users`), OIDC & SAML 2.0 Single Sign-On, official Terraform/OpenTofu provider, and production Kubernetes Helm chart.
- **GitHub Primer UI Parity:** Pixel-accurate GitHub Primer design system with 6px squircle radius, 1px borders, custom OKLCH tokens, authentic empty states, and zero mock data in production.

---

## 2. Core Operational Invariants (Non-Negotiable Rules)

All agents working on this codebase MUST uphold these invariants at all times:

1. **Zero Mock Data in Production (`isDemoModeActive() === false`):**
   - The UI must NEVER flash or default to mock data when connecting to a live cluster.
   - All mock objects in `web/src/lib/mockData.ts` are strictly gated by `isDemoModeActive() === true`.
   - If an endpoint returns empty results, render authentic GitHub Primer empty states with actionable buttons, never placeholder demo items.

2. **Strict GitHub Primer Design System:**
   - Exact `rounded-md` (6px squircle radius).
   - `border border-border` (1px crisp borders).
   - Flush `BoxRow` list items with hover transitions.
   - Custom OKLCH color token architecture (`text-foreground`, `text-muted-foreground`, `bg-secondary`, `border-border`, `bg-card`).
   - Accessible `UnderlineNav` with `role="tablist"` and `aria-selected` attributes.

3. **Hermetic Pure-Go Backend (Zero CGO):**
   - The Go backend must compile with `CGO_ENABLED=0` across all architectures (`linux/amd64`, `linux/arm64`, `darwin/arm64`, `windows/amd64`).
   - SQLite is driven by `modernc.org/sqlite` in WAL mode with `PRAGMA foreign_keys = ON`.
   - Git operations use `github.com/go-git/go-git/v6` directly against bare repository directories on disk.

4. **Security Friction & Danger Zone Safeguards:**
   - High-impact operations (visibility change from public to private, repository deletion, repository transfer) MUST enforce:
     1. Consequence acknowledgement checkbox.
     2. Exact repository path confirmation string (`owner/repo-name`).
     3. Sudo mode authentication challenge (user password or PAT key hash verification).
   - Deploy keys must be strictly isolated to their assigned repository UUID; cross-repo access is rejected.
   - Read-only deploy keys must intercept and abort `git-receive-pack` before spawning subprocesses.
   - Archived repositories are completely frozen: write operations via Git Smart-HTTP, SSH daemon, issues, PRs, comments, and releases must be rejected with `403 Forbidden` / exit status 1.

5. **Test Integrity & Zero-Error Compilation:**
   - Every backend feature must have corresponding unit and integration tests in `internal/<pkg>/<pkg>_test.go`.
   - Every frontend view/modal must have automated Vitest unit tests in `src/**/__tests__/*.test.tsx`.
   - `npm run build` (`tsc -b && vite build`) and `go test ./...` must pass with 100% success and zero warnings before pushing.

---

## 3. Directory Layout & Architecture Map

```text
exist.io/
├── cmd/
│   ├── exist/               # Developer & operator companion CLI
│   └── existd/              # Production forge daemon (HTTP :8080 + SSH :2222)
├── deploy/
│   ├── helm/                # Production Kubernetes Helm chart
│   └── terraform-provider-exist/ # Official Terraform / OpenTofu provider
├── internal/
│   ├── agentactivity/       # Bounded memory ring buffer for agent telemetry stream
│   ├── agentreviewer/       # Autonomous PR review agent with static AST & secret scanner
│   ├── app/                 # Forge REST API v1 handlers, route mux, and HTTP server
│   ├── audit/               # Immutable compliance audit log engine
│   ├── auth/                # Password hashing (bcrypt), PAT tokens, and session manager
│   ├── config/              # Environment and YAML configuration parser
│   ├── gitlfs/              # Git LFS v1 Batch & CAS engine with SHA-256 chunk storage
│   ├── gitstore/            # Bare repository manager using go-git/v6 & hooks injector
│   ├── gittransport/        # Smart-HTTP Git protocol handler (upload-pack, receive-pack)
│   ├── mcp/                 # Model Context Protocol (MCP) JSON-RPC 2.0 server (34 tools)
│   ├── metrics/             # Prometheus exposition & cluster telemetry registry
│   ├── security/            # Pre-receive hook, secret pattern scanner, and branch protection
│   ├── sshserver/           # Standalone pure-Go SSH server (:2222) for git clone & push
│   ├── store/               # SQLite WAL data store, queries, and migrations
│   │   └── migrations/      # Sequential SQL migrations (0001_init.sql ... 0012_*.sql)
│   └── symbolsearch/        # Semantic AST symbol indexer (Go, TypeScript, Python)
├── pkg/
│   ├── agentlease/          # Cryptographic capability leases with ref-prefix isolation
│   ├── idgen/               # Canonical prefixed UUID generator (repo-*, pr-*, usr-*)
│   └── scopes/              # OAuth & PAT token permission matrix
├── web/                     # Vite + React 19 + TypeScript frontend application
│   ├── src/
│   │   ├── components/
│   │   │   ├── modals/      # Creation, transfer, visibility, and archival dialogs
│   │   │   ├── ui/          # Primer primitives (Button, Card, Badge, Dialog, Input)
│   │   │   └── views/       # Application views (Header, RepoOverview, Settings, etc.)
│   │   ├── lib/
│   │   │   ├── api.ts       # Typed fetch client with demo-mode gating
│   │   │   ├── types.ts     # TypeScript interface definitions
│   │   │   └── mockData.ts  # Isolated demo-mode fixtures
│   │   └── index.css        # Tailwind v4 theme & Primer color tokens
│   ├── package.json
│   └── vite.config.ts
├── ROADMAP.md               # Product vision and milestone tracking
├── README.md                # Enterprise forge overview and quickstart guide
└── AGENTS.md                # This operational handbook
```

---

## 4. Key Subsystems & How to Extend Them

### 4.1 Database Migrations (`internal/store/migrations/`)
- All schema changes MUST be declared as a new `.sql` file with the next sequential 4-digit prefix (e.g. `0013_your_feature.sql`).
- All migrations are applied automatically on startup by `store.Open()`.
- Use SQLite-compatible DDL; always provide indexes for foreign keys and queries.
- After creating a migration, add the corresponding CRUD queries in `internal/store/` and write a test in `internal/store/<domain>_test.go`.

### 4.2 Embedded SSH Server (`internal/sshserver/server.go`)
- Listens on `:2222` by default.
- Authenticates users via public keys registered in `user_public_keys` table.
- Authenticates deploy keys via `repo_deploy_keys` table.
- Enforces repo UUID boundaries and intercepts `git-receive-pack` if marked read-only or if the repository is archived.
- Runs without requiring any host OpenSSH binary or daemon.

### 4.3 Pre-Receive Hook & Security Pipeline (`internal/security/hook.go`)
- Injected automatically into bare Git repository directories (`hooks/pre-receive`).
- Validates commits pushed via Smart-HTTP or SSH:
  1. Detects and blocks hardcoded secrets, private keys, and API tokens.
  2. Protects default branches (`main`, `master`) against deletion (`GH006`).
  3. Rejects non-fast-forward force-pushes unless bypassed.

### 4.4 Model Context Protocol (MCP) (`internal/mcp/`)
- Standard JSON-RPC 2.0 endpoint at `/mcp` for agent pair-programming.
- Read-only sandbox at `/mcp/readonly` for low-privilege agent exploration.
- 34 tools available for file reading, semantic symbol search, commit inspection, PR reviews, and check executions.

### 4.5 Ephemeral Agent Leases (`pkg/agentlease/`)
- Short-lived capability tokens issued to autonomous agents.
- Cryptographically signed with HMAC-SHA256 and verified on every Git push and REST request.
- Strictly isolated to `refs/heads/agent/*` branches to prevent unauthorized tampering with production branches.

---

## 5. Development & Testing Runbook

### Backend (Go)
```bash
# Run all unit and integration tests across the repository
go test ./...

# Run tests without cache for specific packages
go test -v -count=1 ./internal/app ./internal/store ./internal/sshserver ./internal/security

# Build the exist companion CLI and existd daemon
go build -o bin/exist ./cmd/exist
go build -o bin/existd ./cmd/existd
```

### Frontend (React & TypeScript)
```bash
# Navigate to web directory
cd web

# Run full Vitest automated test suite (36 suites, 133+ tests)
npm test

# Run specific test suite
npx vitest run src/components/views/__tests__/RepoSettingsView.test.tsx

# Production bundle compilation and TypeScript verification
npm run build

# Code linting
npm run lint
```

---

## 6. Git Commit & Release Guidelines

- Follow Conventional Commits:
  - `feat(<domain>): description` (e.g. `feat(repo): add repository deploy keys with SSH isolation`)
  - `fix(<domain>): description` (e.g. `fix(ssh): enforce read-only check on receive-pack`)
  - `test(<domain>): description` (e.g. `test(ui): add TransferRepoModal test suite`)
  - `docs(<domain>): description` (e.g. `docs(roadmap): update Phase 5 milestones`)
- Keep working tree clean. Never leave dangling untracked test files.
- Always verify both `go test ./...` and `npm test` + `npm run build` pass before pushing.
