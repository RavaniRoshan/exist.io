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

| # | Claim | Evidence (Test / File Path) | Status | Required Action / Remediation | Verified at commit |
| :--- | :--- | :--- | :--- | :--- | :---: |
| 1 | **Full wire compatibility with Origin forge interfaces** | `internal/openapi/origin.openapi.yaml`, `internal/origin/dto.go`, `internal/openapi/golden_test.go:TestOpenAPI_Coverage` | **Unverified** | Removed the wire-compatibility claim in Phase 1 and removed the competitor comparison matrix in EXIST-OPS-001; only code-backed capabilities are published. Cursor Origin has no open specification. | 53bce30 |
| 2 | **Enterprise-ready** | `internal/config/config.go:40` (static JWT secret), `internal/app/app.go:137` (hardcoded litestream status), `internal/auth/sso.go:161` (unverified SSO callback) | **False** | Removed all "enterprise-ready" and "enterprise" suffixes across README and health endpoint; standardized on `v0.5.0-beta`. | 53bce30 |
| 3 | **100% open source with zero external proprietary dependencies** | `go.mod` (pure Go), `LICENSE` (Apache-2.0) | **Verified** | Added official Apache-2.0 `LICENSE` in Phase 1. Core Go backend has zero proprietary dependencies. | 53bce30 |
| 4 | **Pure-Go zero cgo backend** | Verified via `CGO_ENABLED=0 go build ./cmd/exist ./cmd/existd` using `modernc.org/sqlite` and `go-git/v6` | **Verified** | Hermetic pure-Go compilation verified in CI and local test builds. | 53bce30 |
| 5 | **Core Model Context Protocol (MCP) tools** | `internal/mcp/tools.go`, `internal/mcp/server_test.go` | **Verified** | Standardized to core ~8 tools + capability lease tools (`lease_request`, `lease_inspect`, `lease_revoke`). | 53bce30 |
| 6 | **REST API Conformance** | `internal/cli/conformance.go`, `internal/cli/conformance_test.go` | **Verified** | Automated conformance test suite (`exist conformance`) verifies active endpoints with 100% pass rate. | 53bce30 |
| 7 | **Git LFS v1 CAS Engine with file locking** | `internal/gitlfs/server.go`, `internal/gitlfs/server_test.go:TestServer_LFS` | **Verified** | CAS and `/info/lfs/locks` pass tests. Labeled experimental and frozen for launch. | 53bce30 |
| 8 | **OpenSSH SIGSPEC and GPG RFC 4880 signature verification** | `internal/gitstore/signature.go`, `internal/gitstore/signature_test.go` | **Verified** | Both OpenSSH binary armor and PGP commit signature verification pass tests with fingerprint matching. | 53bce30 |
| 9 | **Automated merge queues** | `internal/gitstore/merge.go`, `internal/app/app.go:2170` | **Remediated** | Removed "automated merge queue" claim from README. PRs are merged synchronously via deterministic 3-way/squash merge. | 53bce30 |
| 10 | **AST symbol search** | `internal/symbolsearch/search.go`, `internal/symbolsearch/search_test.go` | **Verified** | Real Go AST (`go/parser`) and regex heuristics for TS/Python verified in tests. | 53bce30 |
| 11 | **Cryptographic capability leases with ref-prefix isolation** | `pkg/agentlease/lease.go`, `internal/gittransport/server.go`, `internal/security/hook.go`, `internal/app/agent_loop_e2e_test.go` | **Verified** | Verified end-to-end: lease issuing, Smart-HTTP transport check, pre-receive hook isolation to `refs/heads/agent/*`, mutation checks. | 53bce30 |
| 12 | **Enterprise SSO (OIDC / SAML 2.0)** | `internal/auth/sso.go:161-175`, `internal/auth/sso_test.go` | **Experimental** | SSO callback trusts client-provided email string. Frozen and labeled experimental in docs. | 53bce30 |
| 13 | **SCIM 2.0 user sync (/scim/v2/Users)** | `internal/auth/scim.go`, `internal/auth/scim_test.go` | **Verified** | RFC 7643/7644 user provisioning and deprovisioning endpoints pass unit tests. Frozen and labeled experimental in docs. | 53bce30 |
| 14 | **Vitest test badge: 137 Passed (37 Suites)** | `web/src/__tests__` | **Verified** | Fixed concurrency timeout in `App.test.tsx:135`. All 37 test suites (137 tests) pass deterministically. Replaced hardcoded badge with CI badge. | 53bce30 |
| 15 | **Version badge: v0.4.7-enterprise** | `README.md:12`, `internal/app/app.go:130` (`v0.5.0-beta`) | **Verified** | Standardized all version badges, API responses, and docs on `v0.5.0-beta`. | 53bce30 |
| 16 | **Quickstart opens http://localhost:8080 for Web UI** | `internal/webui/webui.go`, `internal/app/app.go:2980`, `internal/app/app_test.go` | **Verified** | Web UI embedded directly into `existd` via `go:embed`. Port `:8080` serves Web UI and SPA routes with automatic fallback. | 53bce30 |
| 17 | **Web UI runs on http://localhost:3000** | `README.md:230` | **Remediated** | Removed `:3000` from endpoint table. Unified single-port `:8080` for Web UI, REST API, Git Smart-HTTP, and MCP. | 53bce30 |
| 18 | **Default SSH bind address is 0.0.0.0:2222** | `internal/config/config.go:38` (`127.0.0.1:2222`) | **Remediated** | Documentation corrected to state safe localhost default (`127.0.0.1:2222`). | 53bce30 |
| 19 | **Litestream sync status active in admin health API** | `internal/app/app.go:137` | **False** | Retained as experimental stub; documented in audit report. | 53bce30 |
| 20 | **Dual documentation URLs: legacy host vs exist.io** | `README.md`, `site/blume.config.ts` | **Remediated** | All legacy-domain references removed. Unified strictly on `https://ravaniroshan.github.io/exist.io/`. | 53bce30 |
| 21 | **Marketing + docs site rewrite (2026-11) introduces no new capability claims** | `site/` copy vs this ledger; `site/scripts/audit-claims.mjs`; `internal/claims/coherence_test.go` (now scans `site/{content,app,components}`) | **Verified** | Every capability statement on the landing traces to prior entries (#3–8, #10, #11, #14, #16–18). Claims gate re-pointed at the migrated docs path and extended to the site sources; `npm run claims-audit` enforced in CI. | 81720f9 |

---

## Forbidden claims (machine-enforced)

`internal/claims/coherence_test.go` fails the build if any phrase below appears in
`README.md` or under `site/content`, `site/app`, `site/components` (`.md`, `.mdx`,
`.ts`, `.tsx`). Add a phrase here when a claim is removed or
demoted to experimental, so it cannot silently return.

<!-- forbidden-claims:start -->
wire-compatible with Origin forge interfaces
34 specialized tools
35 specialized tools
35+ Git tools
119 operations
100% coverage of the 119
enterprise-ready
Space UI
production-ready Terraform/OpenTofu provider
37 suites, 137 tests
<!-- forbidden-claims:end -->
