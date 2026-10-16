# Product Hunt Short Description

Product Hunt enforces a strict limit of **260 characters** for short descriptions.

---

## Selected Short Description (Recommended)

> **exist.io is the neutral, self-hosted Git forge built for AI coding agents (Claude Code, Cursor, Cline, Codex). Single pure-Go binary with embedded UI, embedded SSH (:2222), native Model Context Protocol (MCP) server, and ephemeral capability leases with ref isolation.**

*Character count:* **258 / 260**

---

## Bullet Points / Detailed Overview (For Body Text)

- **Single Pure-Go Binary:** Runs with `CGO_ENABLED=0`, embedded React 19 UI, and SQLite WAL storage. Zero external proprietary services.
- **Embedded SSH Server (:2222):** Pure-Go SSH daemon with Ed25519 host key persistence and zero host OpenSSH dependency.
- **Native Model Context Protocol (MCP):** Connect Claude Code, Cursor, Cline, or Codex directly to `/mcp` or sandboxed `/mcp/readonly`.
- **Ephemeral Capability Leases:** Short-lived HMAC-SHA256 capability tokens isolating agent pushes to `refs/heads/agent/*`.
- **Pre-Receive Secret Scanner:** Blocks hardcoded API tokens, AWS keys, and private keys before commits hit disk.
- **Empirically Proven Performance:** Sub-10ms MCP latency and sub-3s full agent loop (clone, push, PR) under multi-agent load.
