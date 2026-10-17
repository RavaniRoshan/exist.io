# Changelog

All notable changes to exist.io will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0-beta] - 2026-10-17

### Added
- **Interactive Code Review & Suggestion Blocks**:
  - Gutter hover `+` controls to add inline comments to specific diff lines and sides (`LEFT`/`RIGHT`).
  - Markdown suggestion block parser (` ```suggestion ... ``` `) and 1-click "Apply suggestion" commit button calling `POST /api/v1/repos/{owner}/{name}/pulls/{number}/comments/{commentId}/apply`.
  - Batch review submission bar with modal supporting Comment, Approve, and Request Changes verdicts.
  - Inline conversation thread resolution (`PATCH .../threads/{threadId}`) with resolved indicators and reply composer.
- **Event-Driven Auto-Merge Engine**:
  - Dedicated background worker (`AutoMergeEngine`) with PR mutex locking to prevent race conditions and double-merges.
  - Automatically squashes/merges pull requests once all required check suites pass.
  - REST endpoints at `GET/PUT/DELETE /api/v1/repos/{owner}/{name}/pulls/{number}/auto-merge`.
  - Integrated auto-merge controls inside `MergeBox`.
- **In-App Notification Center & SMTP Email Engine**:
  - SQLite WAL-backed notification delivery log and unread counter (`0014_notifications.sql`, `0015_in_app_notifications.sql`).
  - In-app notification center dropdown in global header with unread count badge, real-time polling, and 1-click mark-all-read.
  - Pure-Go SMTP client (`internal/notifications`) supporting STARTTLS/TLS and authentic GitHub Primer HTML email templates for PR opened, merged, reviewed, and issue assignments.
- **Repository Migration Wizard**:
  - Multi-platform importer (`MigrateRepoModal`) supporting GitHub and generic Git providers.
  - Background async migration worker cloning bare repositories, importing issues, PRs, and comments with Link-header pagination (`RFC 5988`).
  - Real-time job polling at `GET /api/v1/repos/migrate/{jobId}/status`.
  - Mounted on both `UserProfileView` and `GlobalHomeView`.

## [0.5.0-beta] - 2026-10-17

### Added
- **Embedded Web UI**: Single binary distribution where `existd` serves the compiled React 19 / Vite UI on port 8080 via Go `embed.FS` with automatic SPA fallback.
- **Model Context Protocol (MCP) Server**: 16 core tools with `/mcp` full-access and `/mcp/readonly` sandbox for agent exploration.
- **Agent Capability Leases**: Ephemeral HMAC-SHA256 capability tokens (`pkg/agentlease`) with ref-prefix isolation (`refs/heads/agent/*`) enforced on Smart-HTTP and pre-receive hooks.
- **Pre-Receive Security Hook**: Automatic secret scanning, non-fast-forward push rejection, and default branch protection.
- **AST Symbol Search Indexer**: In-memory AST symbol search for Go, TypeScript, and Python (`internal/symbolsearch`).
- **One-Command Bootstrap**: `existd bootstrap` initializes admin user, namespace, repo, API key, and tailored agent MCP configuration snippets in < 5 seconds.
- **One-Line Installer**: `scripts/install.sh` for single-command installation across Linux, macOS, and Windows.
- **Automated Conformance Harness**: `exist conformance` validating REST endpoints and MCP JSON-RPC protocol compliance.
- **Concurrent Multi-Agent Benchmarks**: Rigorous benchmark suite measuring p50/p95 latency under 1, 5, and 20 concurrent autonomous agents.
- **Threat Model & Security Hardening**: Comprehensive threat model (`THREAT_MODEL.md`), pure-Go SSH modern ciphers (`curve25519-sha256`, `chacha20-poly1305`), 30s handshake deadline, HTTP timeouts, and pktline fuzz testing.
- **Beta Tester Runbook & Feedback Form**: Structured 5-minute intake runbook (`docs/launch/BETA_TESTING.md`) and GitHub Issue form.
- **Claims Ledger & Audit Documentation**: Comprehensive verification ledger (`CLAIMS.md`) and launch readiness audit (`LAUNCH_READINESS.md`).
- **Product Hunt Launch Kit**: Complete copy, taglines, first comment, and maker FAQs in `docs/launch/PRODUCT_HUNT.md`.
- **Repository Governance**: Added `LICENSE` (Apache-2.0), `TRADEMARK.md`, `SECURITY.md`, `CONTRIBUTING.md`, and `CODE_OF_CONDUCT.md`.

### Changed
- **Repositioning**: Positioned as the vendor-neutral, self-hosted Git forge built for autonomous coding agents (Claude Code, Codex, Cline, Cursor).
- **Version Standardization**: Replaced legacy `v0.4.7-enterprise` labels with unified `v0.5.0-beta`.
- **Site URL Unification**: Standardized all documentation and homepage links to canonical URL `https://ravaniroshan.github.io/exist.io/`.
- **Vitest Concurrency Stabilization**: Configured explicit test timeouts and waitFor bounds to eliminate concurrency flakes under load.

### Deprecated / Frozen
- **Experimental Enterprise Modules**: LFS locking, SAML/SCIM SSO, Terraform provider, Helm chart, and multi-service enterprise compose marked as experimental and frozen for launch.
- **Claims Softening**: Removed unverified "full wire compatibility with Cursor Origin" and unbuilt "automated merge queue" claims.
