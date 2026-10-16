# Changelog

All notable changes to exist.io will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0-beta] - 2026-10-16

### Added
- **Embedded Web UI**: Single binary distribution where `existd` serves the compiled React 19 / Vite UI on port 8080 via Go `embed.FS` with automatic SPA fallback.
- **Claims Ledger & Audit Documentation**: Comprehensive verification ledger (`CLAIMS.md`) and launch risk audit (`AUDIT.md`).
- **Comprehensive CI Workflow**: Automated GitHub Actions testing covering Go build, race detector, `go vet`, frontend Vite build, and Vitest suite.
- **Repository Governance**: Added `LICENSE` (Apache-2.0), `TRADEMARK.md`, `SECURITY.md`, `CONTRIBUTING.md`, and `CODE_OF_CONDUCT.md`.

### Changed
- **Repositioning**: Positioned as the vendor-neutral, self-hosted Git forge built for autonomous coding agents (Claude Code, Codex, Cline, Cursor).
- **Version Standardization**: Replaced legacy `v0.4.7-enterprise` and `v0.4.6-enterprise` labels with unified `v0.5.0-beta`.
- **Site URL Unification**: Standardized all documentation and homepage links to canonical URL `https://ravaniroshan.github.io/exist.io/`.
- **Vitest Concurrency Stabilization**: Configured explicit test timeouts and waitFor bounds to eliminate concurrency flakes under load.

### Deprecated / Frozen
- **Experimental Enterprise Modules**: LFS locking, SAML/SCIM SSO, Terraform provider, Helm chart, and multi-service enterprise compose marked as experimental and frozen for launch.
- **Claims Softening**: Removed unverified "full wire compatibility with Cursor Origin" and unbuilt "automated merge queue" claims.
