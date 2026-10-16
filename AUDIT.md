# exist.io Phase 0 Comprehensive Audit (`AUDIT.md`)

Conducted on October 16, 2026. Target Launch: November 17, 2026.

---

## 1. Ground Truth Build & Test Results

Every toolchain and test suite was executed against the clean repository:

| Component / Step | Command | Status | Details / Failure Mode |
| :--- | :--- | :--- | :--- |
| **Go Compilation** | `go build ./cmd/exist ./cmd/existd` | **PASS** | Binaries compile cleanly with Go 1.27.2. |
| **Zero-CGO Pure-Go Build** | `CGO_ENABLED=0 go build ./cmd/exist ./cmd/existd` | **PASS** | Successful hermetic compilation with zero CGO dependencies across both binaries. |
| **Go Vet Analysis** | `go vet ./...` | **PASS** | Zero syntax errors, zero unreachable code warnings. |
| **Go Race Detector Tests** | `go test -race ./...` | **PASS** | All 25 Go packages passed with zero data races detected across the entire test suite. |
| **Frontend Web Compilation** | `npm run build` (in `/web`) | **PASS** | Vite v8.3.4 + TypeScript compiles clean client bundle (`dist/index.html` + assets). |
| **Frontend Automated Tests** | `npm test` (Vitest in `/web`) | **FAIL** | 36 suites passed, 1 failed (`src/__tests__/App.test.tsx:135` timed out in `waitFor` when running all 37 suites in parallel). Passed when run in isolation. |
| **Documentation Site Build** | `npm run build` (in `/site`) | **PASS** | Blume builds all 18 pages to static output cleanly. |
| **Continuous Integration** | `.github/workflows/` | **FAIL** | No CI workflow exists for Go tests or frontend builds; only `pages.yml` exists for docs. |

---

## 2. Top 10 Findings Ranked by Launch Risk

### Finding 1: Broken Agent Capability Lease Enforcement (CRITICAL)
- **Risk Level**: High / Architectural Showstopper
- **Evidence**: `internal/app/app.go:162`, `pkg/agentlease/lease.go:35`, `internal/security/hook.go:27`
- **Details**: `pkg/agentlease` defines capability leases with `AllowedRefPrefixes` and HMAC-SHA256 signatures, and `internal/app/app.go` provides a lease creation endpoint at `POST /api/v1/agent/leases`. However, `agentlease.Verify` is **never called anywhere** in the Git Smart-HTTP handler (`internal/gittransport`), the SSH server (`internal/sshserver`), or the pre-receive security hook (`internal/security`). Furthermore, no agent lease token can be passed or authenticated via Git CLI. The core promise—that agents operate under ephemeral capability leases strictly restricted to `refs/heads/agent/*`—is completely non-functional in the push path.
- **Remediation**: In Phase 2, integrate `agentlease.Verify` into Smart-HTTP Basic Auth / Bearer auth, the SSH server environment, and the pre-receive hook to reject non-`refs/heads/agent/*` ref pushes under lease authentication.

### Finding 2: Web UI Never Served by the existd Daemon (HIGH)
- **Risk Level**: High / User Experience Blocker
- **Evidence**: `internal/app/app.go:2980`, `README.md:36`, `README.md:230`
- **Details**: The README quickstart instructs first-time users to run `existd` and open `http://localhost:8080`. However, `internal/app/app.go` mounts the Git transport handler on root (`mux.Handle("/", gitTransport)`). Visiting `http://localhost:8080/` in a browser returns a 404 or Git protocol error, not the web application. Furthermore, the endpoint table in `README.md` claims the Web UI is on `:3000`, which is neither started by `existd` nor mentioned in the build steps.
- **Remediation**: In Phase 1, embed the built `/web/dist` assets directly into `existd` using Go 1.16+ `go:embed` and serve index/assets on `/` whenever an incoming request is not a Git transport request (`/info/refs`, `git-upload-pack`, `git-receive-pack`).

### Finding 3: Insecure Client-Forged SSO Callback (CRITICAL)
- **Risk Level**: High / Security Vulnerability
- **Evidence**: `internal/auth/sso.go:161-166`
- **Details**: The SSO callback handler (`POST /api/v1/auth/sso/callback`) does not contact an OpenID Connect token endpoint or validate ID token signatures. Instead, it extracts the `email` field directly from the client JSON request body:
  ```go
  email := strings.TrimSpace(req.Email)
  if email == "" {
      email = fmt.Sprintf("sso_user_%s@enterprise.internal", req.Code)
  }
  ```
  Any unauthenticated client can send a POST request with an arbitrary email and immediately log in as that user or trigger auto-provisioning. Furthermore, there is zero SAML 2.0 XML assertion verification in the codebase.
- **Remediation**: Freeze SSO/SAML, label as experimental, remove claims of enterprise-grade OIDC/SAML, and restrict SSO routes to non-production sandbox until external review.

### Finding 4: Hardcoded Static JWT Secret for Agent Leases (HIGH)
- **Risk Level**: High / Security Vulnerability
- **Evidence**: `internal/config/config.go:40`
- **Details**: If `EXIST_JWT_SECRET` is omitted from the environment, `config.Load()` falls back to a public static string:
  ```go
  jwtSecret := getEnv("EXIST_JWT_SECRET", "exist-ephemeral-agent-lease-signing-key")
  ```
  Any attacker or agent with knowledge of this hardcoded default can forge arbitrary capability leases.
- **Remediation**: In Phase 4, generate an ephemeral, cryptographically secure random secret on startup if none is configured, or require explicit secret configuration.

### Finding 5: Total Absence of Automated CI Pipeline (HIGH)
- **Risk Level**: High / Quality Risk
- **Evidence**: `.github/workflows/pages.yml:1-65`, `README.md:22`
- **Details**: The repository contains only one GitHub Actions workflow (`pages.yml`), which exclusively builds and deploys documentation to GitHub Pages. There is no automated workflow running `go build`, `go test -race`, `go vet`, `npm test`, or `npm run build` on push or pull request. The README test badge (`Tests-137 Passed`) is a static shields.io image with no CI backing.
- **Remediation**: In Phase 1, implement a comprehensive CI matrix (`.github/workflows/ci.yml`) covering Go builds, vet, staticcheck, govulncheck, go test with race detector, web build, Vitest suite, and broken link checks.

### Finding 6: Missing License Files Despite Dual-License Claims (MEDIUM)
- **Risk Level**: Medium / Legal & Open-Source Hygiene
- **Evidence**: `README.md:13`, `README.md:408-411`
- **Details**: The README advertises dual licensing: `Licensed under the Apache License, Version 2.0 (LICENSE) or MIT License (LICENSE-MIT)`. Neither `LICENSE` nor `LICENSE-MIT` exists in the repository root.
- **Remediation**: Add standard `LICENSE` (Apache-2.0) and add `TRADEMARK.md` protecting the project name and logo.

### Finding 7: Fictitious "Automated Merge Queue" Feature (HIGH)
- **Risk Level**: High / False Claim
- **Evidence**: `README.md:80`, `README.md:144`, `AGENTS.md:144`, `internal/gitstore/merge.go:1-250`
- **Details**: The README and architectural diagrams prominently list "automated merge queues" and a "3-Way Merge & Ruleset Policy Engine". Code inspection of `internal/gitstore` and `internal/app` reveals deterministic 3-way merging, squash merging, and policy checks, but **no merge queue worker, data model, or asynchronous ordering exists**. Merges are either executed synchronously or rejected immediately.
- **Remediation**: Delete the "automated merge queue" claim from README, docs, and diagrams.

### Finding 8: MCP Tool Naming Inconsistencies & Missing Tools (MEDIUM)
- **Risk Level**: Medium / Agent Integration Friction
- **Evidence**: `README.md:309-316`, `site/docs/tutorials/first-agent.mdx`, `internal/mcp/tools.go:90-1126`
- **Details**: The README documents MCP tools named `repo_tree`, `file_read`, `diff_compare`, `pull_request_create`, `thread_comment`, and `status_check_report`. In `internal/mcp/tools.go`, the tools are actually named `get_git_tree`, `get_file_contents`, `compare_commits`, `create_pull_request`, and `create_pull_request_comment`. Moreover, `status_check_report` does not exist in `tools.go` at all (only read operations exist), and capability lease management tools are missing from the MCP surface.
- **Remediation**: In Phase 2, standardize the core MCP tool surface on clean, canonical names matching the specification, add `status_check_report` and lease tools, and update all tutorials and docs.

### Finding 9: Unsubstantiated "Full Wire Compatibility with Cursor Origin" (HIGH)
- **Risk Level**: High / False Claim
- **Evidence**: `README.md:33-35`, `internal/openapi/origin.openapi.yaml`
- **Details**: The README claims "full wire compatibility with Origin forge interfaces". No public Cursor Origin specification exists. The 119 REST operations implemented in `internal/origin` were reverse-engineered from internal models. There is no automated conformance suite proving compatibility with real Cursor client software.
- **Remediation**: Drop "alternative to Cursor Origin" as the headline. Remove wire compatibility claims in favor of an honest conformance table showing exactly what is implemented and tested.

### Finding 10: Vitest Concurrency Test Flake (MEDIUM)
- **Risk Level**: Medium / Testing Reliability
- **Evidence**: `web/src/__tests__/App.test.tsx:135`
- **Details**: Running `npm test` across all 37 Vitest suites spawns 37 JSDOM instances simultaneously, causing CPU starvation on local runners. This causes `App.test.tsx:135` (`waitFor` on repository settings tabs) to fail intermittently.
- **Remediation**: Update `web/vite.config.ts` test configuration to optimize worker pool concurrency (`threads: false` or `pool: 'vmThreads'`) and increase `waitFor` timeout defaults to ensure 100% deterministic green passes.

---

## 3. Discrepancies, Contradictions & Dead Links Audit

1. **Brand & URL Split (`exist.os` vs `exist.io`)**:
   - `README.md:10`, `README.md:25`, and `README.md:39` link to `https://ravaniroshan.github.io/exist.os/`.
   - `README.md:11` links to `https://ravaniroshan.github.io/exist.io/`.
   - `site/blume.config.ts` defines base as `/exist.io`.
   - Action: Unify all references to `https://ravaniroshan.github.io/exist.io/`.

2. **Web UI Port Mismatch (`:8080` vs `:3000`)**:
   - `README.md:36` states: *"then open http://localhost:8080 (HTTP :8080 · SSH :2222)"*.
   - `README.md:230` table states: *"`http://localhost:3000` \| Web UI"*.
   - `cmd/existd` never opens port 3000.
   - Action: Embed Web UI into `existd` and document `:8080` as the sole HTTP endpoint.

3. **Hardcoded Litestream Status**:
   - `internal/app/app.go:137` returns `"litestreamSyncStatus": "active"` hardcoded in the health JSON response, even if Litestream is not present.
   - Action: Return actual status or remove fake replication claims.

4. **Version Badge Discrepancy**:
   - `README.md:12` displays `v0.4.7--enterprise`.
   - `internal/app/app.go:130` returns `existd v0.4.6-enterprise`.
   - Action: Update both to `v0.5.0-beta`.

---

## 4. Security Surface & Attack Vector Map

### 4.1 Network Listeners
1. **HTTP Gateway** (`127.0.0.1:8080`):
   - Exposes REST API, Smart-HTTP transport, Git LFS, SCIM 2.0, SSO endpoints, MCP JSON-RPC, metrics, and health probes.
2. **Pure-Go SSH Server** (`127.0.0.1:2222`):
   - Exposes Git clone and push over SSH. Handled via in-process `go-git` transport. Binds to `127.0.0.1:2222` by default.

### 4.2 Authentication & Token Lifecycles
- **Web UI Sessions**: `exist_session_token` cookie. Stored in SQLite `sessions` table.
- **Personal Access Tokens (PAT)**: Prefixed `token_...`, hashed with SHA-256 in `user_tokens`. Verified on REST requests.
- **Agent Capability Leases**: Prefixed `lease_...`, signed with HMAC-SHA256. Verified in unit tests, but not wired to transport.
- **SCIM Bearer Tokens**: Stored in `scim_tokens`, verified on `/scim/v2/*`.
- **SSH Public Keys & Deploy Keys**: Stored in `user_public_keys` and `repo_deploy_keys`. Verified via SSH daemon public key callback.

---

## 5. Repository Hygiene & Files to Exclude from Shipping

1. **`research/github-origin-reverse-engineering-and-enterprise-roadmap.md`**:
   - Internal competitor analysis and notes comparing exist.io to GitHub Enterprise.
   - Action: Keep in internal dev branches or exclude from public release artifacts.
2. **`agent.md`**:
   - 4-line redundant pointer file to `AGENTS.md`.
   - Action: Consolidate into `AGENTS.md` and remove `agent.md`.
3. **Missing Hygiene Files**:
   - `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `CHANGELOG.md`, `TRADEMARK.md` must be created.
