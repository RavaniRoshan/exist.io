# Frequently Asked Questions (`FAQS.md`)

Common questions from developers, engineering leaders, and AI agent operators about exist.io.

---

### 1. Is exist.io a clone of Cursor Origin?

**No.** exist.io is an independent, vendor-neutral Git forge engineered from first principles in pure Go.

While Cursor Origin is a hosted service tied to Cursor accounts and proprietary infrastructure, exist.io is:
- **100% self-hosted:** Runs on your laptop, home server, air-gapped machine, or private cloud.
- **Vendor-neutral:** Designed equally for Claude Code, Cursor, Cline, OpenAI Codex, or any agent speaking the Model Context Protocol (MCP).
- **Hermetic & Open Source:** Zero external proprietary cloud dependencies, zero CGO, and licensed under the permissive Apache-2.0 license.

---

### 2. Does exist.io support my coding agent?

**Yes.** Any agent that supports the **Model Context Protocol (MCP)** or standard **Git CLI / Smart-HTTP** works out of the box:
- **Claude Code:** Full support via `claude mcp add exist http://localhost:8080/mcp`.
- **Cursor:** Native support via `.cursor/mcp.json`.
- **Cline:** Full support via `cline_mcp_settings.json`.
- **Codex / Custom Agents:** Standard JSON-RPC 2.0 endpoint at `/mcp` or sandboxed `/mcp/readonly`.

---

### 3. What is the software license?

exist.io is released under the **Apache License, Version 2.0** ([`LICENSE`](https://github.com/RavaniRoshan/exist.io/blob/main/LICENSE)).
- You are free to run it for personal, commercial, internal enterprise, or air-gapped use.
- You can modify the source code, build custom plugins, and run private clusters without paying licensing fees or royalties.

---

### 4. How is exist.io hosted and deployed?

exist.io is distributed as a single hermetic binary (`existd`) with the full Web UI embedded via `go:embed`.

**Deployment options:**
- **Local Machine / VPS:** Download the binary, run `./existd bootstrap`, and `./existd`. That's it.
- **Docker:** `docker run -p 8080:8080 -p 2222:2222 -v /data:/data ravaniroshan/existd:latest`.
- **Kubernetes:** Production Helm chart available in `deploy/helm/` (labeled experimental for launch).

No separate PostgreSQL, Redis, or OpenSSH server is required. SQLite WAL mode (`modernc.org/sqlite`) handles all database queries with zero lock contention.

---

### 5. What are Capability Leases and how do they protect my code?

Traditional Git tokens give agents full write access to all branches. An errant LLM prompt can trigger a catastrophic `git push --force origin main`.

exist.io introduces **Capability Leases (`pkg/agentlease`)**:
- Short-lived HMAC-SHA256 capability tokens (default: 15–60 minutes).
- Strictly isolated to specific ref prefixes (by default: `refs/heads/agent/*`).
- Pre-receive hook rejects any attempt to push to `main` or delete existing branches with error `GH007`.
- Secret scanners automatically inspect incoming commits for private keys and tokens before accepting the push.

---

### 6. Where can I inspect the code proofs and benchmarks?

We practice radical engineering transparency:
- Every README claim is audited and backed by file/test proofs in [`CLAIMS.md`](https://github.com/RavaniRoshan/exist.io/blob/main/CLAIMS.md).
- Automated concurrency benchmarks and p50/p95 latency metrics are published in [`docs/benchmarks.md`](https://github.com/RavaniRoshan/exist.io/blob/main/docs/benchmarks.md).
- Security threat analysis and attack surface mitigations are documented in [`THREAT_MODEL.md`](https://github.com/RavaniRoshan/exist.io/blob/main/THREAT_MODEL.md).
