# Claims Ledger (`CLAIMS.md`)

This ledger documents every technical and product claim made in the README, documentation (`/site`), and internal handbooks (`AGENTS.md`), mapped to empirical code evidence, verification status, and required remediation actions.

## Legend
- **Verified**: Confirmed by an automated passing test or code implementation.
- **Partial**: Implemented or tested, but with significant discrepancies, missing tools, or limitations not stated.
- **Unverified**: Implemented in code against internal schemas, but lacks external conformance testing or proof.
- **Remediated**: Previously false or contradictory claim resolved during Phase 1.
- **False**: Claimed in documentation or badges, but absent, failing, or contradicted by the codebase.

---

## Claims Audit Table

| # | Claim | Evidence (Test / File Path) | Status | Required Action / Remediation |
| :--- | :--- | :--- | :--- | :--- |
| 1 | **Full wire compatibility with Origin forge interfaces** | `internal/openapi/origin.openapi.yaml`, `internal/origin/dto.go`, `internal/openapi/golden_test.go:TestOpenAPI_Coverage` | **Unverified** | Removed wire compatibility claim in Phase 1; replaced with verified comparison matrix. Cursor Origin has no open specification. |
| 2 | **Enterprise-ready** | `internal/config/config.go:40` (static JWT secret), `internal/app/app.go:137` (hardcoded litestream status), `internal/auth/sso.go:161` (unverified SSO callback) | **False** | Removed all "enterprise-ready" and "enterprise" suffixes across README and health endpoint; standardized on `v0.5.0-beta`. |
| 3 | **100% open source with zero external proprietary dependencies** | `go.mod` (pure Go), `LICENSE` (Apache-2.0) | **Verified** | Added official Apache-2.0 `LICENSE` in Phase 1. Core Go backend has zero proprietary dependencies. |
| 4 | **Pure-Go zero cgo backend** | Verified via `CGO_ENABLED=0 go build ./cmd/exist ./cmd/existd` using `modernc.org/sqlite` and `go-git/v6` | **Verified** | Hermetic pure-Go compilation verified in CI and local test builds. |
| 5 | **35 Model Context Protocol (MCP) tools** | `internal/mcp/tools.go` (defines 35 tools) | **Partial** | Tool names in code do not match README names. In Phase 2, standardize MCP surface to core ~8 tools + lease tools. |
| 6 | **119 REST operations** | `internal/openapi/origin.openapi.yaml` (119 operations), tested in `internal/openapi/golden_test.go` | **Verified** | Retain count, documented as the internal OpenAPI specification. |
| 7 | **Git LFS v1 CAS Engine with file locking** | `internal/gitlfs/server.go`, `internal/gitlfs/server_test.go:TestServer_LFS` | **Verified** | CAS and `/info/lfs/locks` pass tests. Labeled experimental and frozen for launch. |
| 8 | **OpenSSH SIGSPEC and GPG RFC 4880 signature verification** | `internal/gitstore/signature.go`, `internal/gitstore/signature_test.go` | **Verified** | Both OpenSSH binary armor and PGP commit signature verification pass tests with fingerprint matching. |
| 9 | **Automated merge queues** | `internal/gitstore/merge.go`, `internal/app/app.go:2170` | **Remediated** | Removed "automated merge queue" claim from README. PRs are merged synchronously via deterministic 3-way/squash merge. |
| 10 | **AST symbol search** | `internal/symbolsearch/search.go`, `internal/symbolsearch/search_test.go` | **Verified** | Real Go AST (`go/parser`) and regex heuristics for TS/Python verified in tests. |
| 11 | **Cryptographic capability leases with ref-prefix isolation** | `pkg/agentlease/lease.go`, `internal/app/app.go:163` | **Pending Phase 2** | Leases can be issued at `POST /api/v1/agent/leases`. Integration into Smart-HTTP, SSH server, and pre-receive hook to be completed in Phase 2. |
| 12 | **Enterprise SSO (OIDC / SAML 2.0)** | `internal/auth/sso.go:161-175`, `internal/auth/sso_test.go` | **Experimental** | SSO callback trusts client-provided email string. Frozen and labeled experimental in docs. |
| 13 | **SCIM 2.0 user sync (/scim/v2/Users)** | `internal/auth/scim.go`, `internal/auth/scim_test.go` | **Verified** | RFC 7643/7644 user provisioning and deprovisioning endpoints pass unit tests. Frozen and labeled experimental in docs. |
| 14 | **Vitest test badge: 137 Passed (37 Suites)** | `web/src/__tests__` | **Verified** | Fixed concurrency timeout in `App.test.tsx:135`. All 37 test suites (137 tests) pass deterministically. Replaced hardcoded badge with CI badge. |
| 15 | **Version badge: v0.4.7-enterprise** | `README.md:12`, `internal/app/app.go:130` (`v0.5.0-beta`) | **Verified** | Standardized all version badges, API responses, and docs on `v0.5.0-beta`. |
| 16 | **Quickstart opens http://localhost:8080 for Web UI** | `internal/webui/webui.go`, `internal/app/app.go:2980`, `internal/app/app_test.go` | **Verified** | Web UI embedded directly into `existd` via `go:embed`. Port `:8080` serves Web UI and SPA routes with automatic fallback. |
| 17 | **Web UI runs on http://localhost:3000** | `README.md:230` | **Remediated** | Removed `:3000` from endpoint table. Unified single-port `:8080` for Web UI, REST API, Git Smart-HTTP, and MCP. |
| 18 | **Default SSH bind address is 0.0.0.0:2222** | `internal/config/config.go:38` (`127.0.0.1:2222`) | **Remediated** | Documentation corrected to state safe localhost default (`127.0.0.1:2222`). |
| 19 | **Litestream sync status active in admin health API** | `internal/app/app.go:137` | **False** | Retained as experimental stub; documented in audit report. |
| 20 | **Dual site URLs: exist.os vs exist.io** | `README.md`, `site/blume.config.ts` | **Remediated** | All `exist.os` references removed. Unified strictly on `https://ravaniroshan.github.io/exist.io/`. |
