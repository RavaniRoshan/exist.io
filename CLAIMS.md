# Claims Ledger (`CLAIMS.md`)

This ledger documents every technical and product claim made in the README, documentation (`/site`), and internal handbooks (`AGENTS.md`), mapped to empirical code evidence, verification status, and required remediation actions.

## Legend
- **Verified**: Confirmed by an automated passing test or code implementation.
- **Partial**: Implemented or tested, but with significant discrepancies, missing tools, or limitations not stated.
- **Unverified**: Implemented in code against internal schemas, but lacks external conformance testing or proof.
- **False**: Claimed in documentation or badges, but absent, failing, or contradicted by the codebase.

---

## Claims Audit Table

| # | Claim | Evidence (Test / File Path) | Status | Required Action |
| :--- | :--- | :--- | :--- | :--- |
| 1 | **Full wire compatibility with Origin forge interfaces** | `internal/openapi/origin.openapi.yaml`, `internal/origin/dto.go`, `internal/openapi/golden_test.go:TestOpenAPI_Coverage` | **Unverified** | Cursor Origin has no published open specification. The 119 REST operations match an internal schema. Delete the wire compatibility claim or replace with a factual conformance table of tested endpoints. |
| 2 | **Enterprise-ready** | `internal/config/config.go:40` (static JWT secret), `internal/app/app.go:137` (hardcoded litestream status), `internal/auth/sso.go:161` (unverified SSO callback) | **False** | Delete all "enterprise-ready" and "enterprise" suffixes across the codebase until independent security reviews and live deployments occur. |
| 3 | **100% open source with zero external proprietary dependencies** | `go.mod` (pure Go), missing `LICENSE` file, `docker-compose.enterprise.yml` (references MinIO and Litestream) | **Partial** | Add the missing Apache-2.0 `LICENSE` file immediately. The core Go backend has zero proprietary dependencies, but compose bundles third-party services. |
| 4 | **Pure-Go zero cgo backend** | Verified via `CGO_ENABLED=0 go build ./cmd/exist ./cmd/existd` using `modernc.org/sqlite` and `go-git/v6` | **Verified** | Keep and highlight this architectural strength. Note that SQLite WAL is single-writer. |
| 5 | **35 Model Context Protocol (MCP) tools** | `internal/mcp/tools.go` (defines 35 tools) | **Partial** | Tool names in code (`get_git_tree`, `get_file_contents`, `compare_commits`) do NOT match README names (`repo_tree`, `file_read`, `diff_compare`). `status_check_report` is completely missing. Trim and standardize MCP to core 8 tools + lease tools. |
| 6 | **119 REST operations** | `internal/openapi/origin.openapi.yaml` (119 operations), tested in `internal/openapi/golden_test.go` | **Verified** | Retain count, but clarify that it covers the internal OpenAPI specification. |
| 7 | **Git LFS v1 CAS Engine with file locking** | `internal/gitlfs/server.go`, `internal/gitlfs/server_test.go:TestServer_LFS` | **Verified** | CAS and `/info/lfs/locks` work and pass tests. Per launch plan, freeze LFS locking and label experimental to focus on core agent loop. |
| 8 | **OpenSSH SIGSPEC and GPG RFC 4880 signature verification** | `internal/gitstore/signature.go`, `internal/gitstore/signature_test.go` | **Verified** | Both OpenSSH binary armor and PGP commit signature verification pass tests with fingerprint matching. |
| 9 | **Automated merge queues** | `internal/gitstore/merge.go`, `internal/app/app.go:2170` | **False** | Only immediate 3-way merge and squash merge exist. There is no queue worker, data structure, or background task for merge queues. Delete claim immediately. |
| 10 | **AST symbol search** | `internal/symbolsearch/search.go`, `internal/symbolsearch/search_test.go` | **Verified** | Implements real Go AST (`go/parser`) and regex heuristics for TS/Python. Clarify in docs that Go uses standard AST while TS/Python use regex extractors. |
| 11 | **Cryptographic capability leases with ref-prefix isolation** | `pkg/agentlease/lease.go`, `internal/app/app.go:163` | **False** | Leases can be issued at `POST /api/v1/agent/leases`, but `agentlease.Verify` is NEVER called on Smart-HTTP, SSH, pre-receive hooks, or PR merges. `refs/heads/agent/*` isolation is completely unenforced in the push path. Must be implemented in Phase 2. |
| 12 | **Enterprise SSO (OIDC / SAML 2.0)** | `internal/auth/sso.go:161-175`, `internal/auth/sso_test.go` | **False** | The SSO callback trusts an unverified client-provided email string and does not contact an IdP token endpoint or verify SAML assertions. Freeze and label experimental. |
| 13 | **SCIM 2.0 user sync (/scim/v2/Users)** | `internal/auth/scim.go`, `internal/auth/scim_test.go` | **Verified** | RFC 7643/7644 user provisioning and deprovisioning endpoints pass unit tests. Freeze and label experimental. |
| 14 | **Vitest test badge: 137 Passed (37 Suites)** | `web/src/__tests__` | **False** | When the full test suite runs concurrently, `src/__tests__/App.test.tsx` fails. The badge is a hardcoded static shield, not dynamic, and ignores backend Go tests. Replace with CI-driven badge. |
| 15 | **Version badge: v0.4.7-enterprise** | `README.md:12`, `internal/app/app.go:130` (`v0.4.6-enterprise`) | **False** | Version numbers are contradictory and hardcoded. Remove enterprise suffix and standardize on `v0.5.0-beta`. |
| 16 | **Quickstart opens http://localhost:8080 for Web UI** | `internal/app/app.go:2980` | **False** | `existd` mounts `gitTransport` on `/` and never serves the built React web UI. Visiting `:8080` in a browser returns Git transport or 404. Must embed Web UI via `go:embed`. |
| 17 | **Web UI runs on http://localhost:3000** | `README.md:230` | **False** | `existd` does not listen on 3000; the quickstart never instructs running `npm run dev` in `/web`. Resolve by serving embedded UI on `:8080`. |
| 18 | **Default SSH bind address is 0.0.0.0:2222** | `internal/config/config.go:38` (`127.0.0.1:2222`) | **False** | README config table lists `0.0.0.0:2222`, but code defaults to `127.0.0.1:2222`. Update docs to reflect localhost default. |
| 19 | **Litestream sync status active in admin health API** | `internal/app/app.go:137` | **False** | `"litestreamSyncStatus": "active"` is a hardcoded JSON string returned by the admin endpoint regardless of whether Litestream is installed or running. |
| 20 | **Dual site URLs: exist.os vs exist.io** | `README.md:10`, `site/blume.config.ts` | **False** | Brand is split between `exist.os` and `exist.io`. Standardize exclusively on `exist.io`. |
