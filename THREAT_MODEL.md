# exist.io — Threat Model & Agent Security Architecture (`THREAT_MODEL.md`)

<p align="center">
  <strong>Comprehensive Security Analysis, Trust Boundaries, and Attack Surface Mitigations for Autonomous Coding Agents.</strong>
</p>

---

## 1. System Overview & Core Philosophy

**exist.io** is designed to treat autonomous AI coding agents (Claude Code, Codex, Cline, Cursor) as first-class, untrusted or semi-trusted collaborators operating on shared codebases.

Unlike traditional human-oriented Git platforms where access is binary (all-or-nothing SSH keys or long-lived personal access tokens), exist.io enforces **defense-in-depth at the protocol and transport layers**:
1. **Model Context Protocol (MCP)** boundary with strict read-only sandboxing (`/mcp/readonly` vs `/mcp`).
2. **Cryptographic Capability Leases** (`pkg/agentlease`) with short TTLs and ref-prefix isolation (`refs/heads/agent/*`).
3. **Pure-Go Git Pre-Receive Gate** (`internal/security/hook.go`) with secret pattern scanning and branch protection.
4. **Append-Only Immutable Compliance Audit Log** (`internal/audit`).

---

## 2. Threat Analysis: The Agent Boundary

This section analyzes the primary attack vectors available to a malicious, misconfigured, or compromised AI agent operating against an exist.io forge instance.

---

### Threat 1: Escalate Capability Lease to Forge Administrator

- **Attack Vector:** An agent holding a capability lease token attempts to invoke administrative endpoints (e.g. `/api/v1/admin/users`, `/v1/origin/meta`, `/api/v1/sso/*`, deleting organizations or namespaces) or mint long-lived Personal Access Tokens (PATs).
- **Impact:** Complete forge takeover, administrative user impersonation, arbitrary repository manipulation.
- **Mitigation:**
  - Capability leases are strictly typed (`TypeAgentLease`) and separate from user sessions and PATs (`TypeUser`).
  - Auth resolvers (`internal/auth/db_resolver.go`, `internal/auth/auth_helper.go`) verify credential type and reject any agent lease presenting against administrative or user management routes with `401 Unauthorized` or `403 Forbidden`.
  - The MCP read-only sandbox (`/mcp/readonly`) unconditionally restricts tool invocation to non-mutating primitives (`repo_tree`, `file_read`, `diff_compare`, `symbol_search`), preventing administrative or mutating tool execution.
- **Verification & Test:**
  - Automated integration test in [`internal/app/agent_loop_e2e_test.go`](file:///home/shiva/projects/exist.io/internal/app/agent_loop_e2e_test.go) explicitly asserts that mutating operations on `/mcp/readonly` and unauthorized administrative calls are blocked.

---

### Threat 2: Escape `allowed_ref_prefixes` (Pushing to Production or Deleting Refs)

- **Attack Vector:** An agent granted a capability lease scoped to `refs/heads/agent/*` attempts to push directly to production branches (`refs/heads/main`, `refs/heads/master`, `refs/tags/*`) or delete existing branches.
- **Impact:** Poisoning production branch histories, bypassing code reviews, accidental or malicious data loss.
- **Mitigation:**
  - Multi-tier transport and hook enforcement:
    1. **Transport Layer (`internal/gittransport/server.go`):** On `git-receive-pack`, the lease token is validated and decoded. Allowed ref prefixes, disallowed ref prefixes, and expiration timestamps are passed as sealed environment variables (`EXIST_LEASE_*`) to the execution environment.
    2. **Pre-Receive Hook (`internal/security/hook.go`):** Every ref update tuple (`<old-rev> <new-rev> <ref-name>`) is evaluated against `EXIST_LEASE_ALLOWED_REFS` and `EXIST_LEASE_DISALLOWED_REFS`.
    3. **Branch Deletion Prohibition:** An agent push attempting ref deletion (new-rev `0000000000000000000000000000000000000000`) is rejected with `GH007: Agent capability lease does not permit deleting refs`.
    4. **Protected Branch Guards:** Default branches (`main`, `master`) enforce `GH006` protection against non-fast-forward force-pushes and deletions.
- **Verification & Test:**
  - Unit test: [`internal/security/hook_test.go`](file:///home/shiva/projects/exist.io/internal/security/hook_test.go) (`TestPreReceiveHook_AgentLeaseEnforcement`).
  - Transport test: [`internal/gittransport/server_test.go`](file:///home/shiva/projects/exist.io/internal/gittransport/server_test.go) (`TestGitSmartHTTP_AgentLeaseEnforcement`).
  - First-run install script mutation check: [`scripts/test_first_run.sh --mutation`](file:///home/shiva/projects/exist.io/scripts/test_first_run.sh).

---

### Threat 3: Forge or Tamper with a Capability Lease Token

- **Attack Vector:** An attacker or compromised agent attempts to forge a capability lease token, modify `allowed_ref_prefixes`, or extend `expires_at` without holding the daemon's signing key.
- **Impact:** Arbitrary branch push capability across any repository on the instance.
- **Mitigation:**
  - Capability leases are cryptographically signed using HMAC-SHA256 (`pkg/agentlease/lease.go`).
  - The signing key is derived from the server's private secret (`EXIST_JWT_SECRET` / `cfg.JWTSecret`).
  - Token verification uses constant-time comparison (`crypto/subtle` / `hmac.Equal`) to prevent timing side-channel attacks.
  - Any altered payload or invalid signature returns `ErrInvalidSignature` and terminates authentication immediately.
- **Verification & Test:**
  - Unit tests: [`pkg/agentlease/lease_test.go`](file:///home/shiva/projects/exist.io/pkg/agentlease/lease_test.go) (`TestTamperedLeaseFails`).
  - Automated fuzz testing: [`pkg/agentlease/lease_fuzz_test.go`](file:///home/shiva/projects/exist.io/pkg/agentlease/lease_fuzz_test.go) (`FuzzTokenParser`).

---

### Threat 4: Denial of Service via Giant Push or Recursive Cycles

- **Attack Vector:** An agent floods the Git Smart-HTTP endpoint with massive binary packfiles, recursive submodules (`.gitmodules`), or endless concurrent connections to exhaust system memory, disk, and CPU.
- **Impact:** Server crash, memory exhaustion (OOM), denial of service for other human and agent developers.
- **Mitigation:**
  - **Bounded Concurrency Semaphores:** Both Smart-HTTP (`internal/gittransport/server.go`) and SSH (`internal/sshserver/server.go`) maintain bounded connection pools (`sem = make(chan struct{}, 64)`).
  - **HTTP Timeouts & Header Limits:** `http.Server` configured with `ReadHeaderTimeout: 10s`, `ReadTimeout: 60s`, `WriteTimeout: 120s`, `IdleTimeout: 120s`, and `MaxHeaderBytes: 1MB`.
  - **SSH Handshake Timeout:** 30-second handshake deadline on incoming raw TCP connections prevents Slowloris attacks.
  - **Git LFS CAS Engine:** Large files must be transferred via Content-Addressable Storage (CAS) with chunked SHA-256 verification, preventing Git packfile bloat.
- **Verification & Test:**
  - Integration tests in [`internal/gittransport/server_test.go`](file:///home/shiva/projects/exist.io/internal/gittransport/server_test.go) and [`internal/sshserver/server_test.go`](file:///home/shiva/projects/exist.io/internal/sshserver/server_test.go).

---

### Threat 5: Poisoning or Deleting the Compliance Audit Log

- **Attack Vector:** An agent attempts to delete audit logs, rewrite event histories, or inject false compliance records to conceal unauthorized branch modifications or secret leaks.
- **Impact:** Loss of enterprise audit trail, undetected compliance violations (SOC 2, ISO 27001).
- **Mitigation:**
  - **Append-Only Architecture:** The `audit_logs` table in SQLite has no `DELETE` or `UPDATE` operations exposed through REST or MCP interfaces.
  - **Process-Space Isolation:** Audit entries are written directly by internal server components (`internal/audit/logger.go`) during transport events, lease creation, and PR merges. Agents have no direct access to the database layer.
  - **Immutable JSON Records:** Every audit entry contains immutable metadata: unique audit ID (`aud-*`), nanosecond timestamp, authenticated actor ID, actor type (`agent`, `user`), action, target resource, client IP, severity, and details.
- **Verification & Test:**
  - Audit tests in [`internal/audit/audit_test.go`](file:///home/shiva/projects/exist.io/internal/audit/audit_test.go) and [`internal/store/audit_logs_test.go`](file:///home/shiva/projects/exist.io/internal/store/audit_logs_test.go).

---

## 3. Network Listeners & Attack Surface Matrix

| Service | Default Bind | Default Port | Auth Mechanism | Intended Clients | Hardening Controls |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **HTTP Forge API & UI** | `127.0.0.1` | `:8080` | Bearer Token / Session Cookie | Human Devs & Agent Clients | ReadHeader/Read/Write timeouts, max header limits, localhost bind |
| **Model Context Protocol** | `127.0.0.1` | `:8080/mcp` | Bearer Token (PAT or Lease) | Claude Code, Cursor, Cline | Sandboxed `/mcp/readonly` endpoint, strict JSON-RPC input parsing |
| **Git Smart-HTTP** | `127.0.0.1` | `:8080/{owner}/{repo}.git` | HTTP Basic Auth (PAT or Lease) | Git CLI & Agent Git subprocesses | Capability lease ref isolation, pre-receive secret scanning, hook sandbox |
| **Standalone SSH Server** | `127.0.0.1` | `:2222` | Public Key (Ed25519/RSA/CAs) | Git CLI (clone & push) | 30s handshake deadline, modern cipher suites, persisted Ed25519 host key |

---

## 4. Security Invariants Summary

1. **Localhost Default:** The daemon binds to `127.0.0.1` by default; public internet exposure requires explicit operator configuration.
2. **Ephemeral Capability Leases:** Autonomous agents must operate under short-lived capability leases restricted to `refs/heads/agent/*`.
3. **Pre-Receive Defense:** Secrets, private keys, and unauthorized ref pushes are blocked before Git commits are written to bare repos.
4. **Append-Only Auditing:** Every agent transport action is irreversibly recorded in the local SQLite WAL audit trail.
