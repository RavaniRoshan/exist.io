# Testing Strategy & Mutation Checks (`docs/testing.md`)

This document outlines the testing architecture, validation invariants, and mutation verification protocols for **exist.io**.

---

## 1. Automated Test Hierarchy

| Layer | Packages | Runner | Purpose |
| :--- | :--- | :--- | :--- |
| **Unit Tests** | `internal/store`, `internal/security`, `pkg/agentlease`, `internal/auth` | `go test -race` | Verifies low-level algorithms, SQL queries, crypto signatures, secret scanning, and lease token parsing. |
| **Protocol & API Tests** | `internal/mcp`, `internal/gittransport`, `internal/origin` | `go test -race` | Validates JSON-RPC 2.0 MCP tools, Smart-HTTP clone/push transport, and REST API handlers. |
| **End-to-End Integration** | `internal/app/agent_loop_e2e_test.go` | `go test -v -run TestAgentLoop_EndToEnd_RealServer` | Validates full real-world agent workflow on a live HTTP daemon using native git CLI with zero mocks. |
| **Frontend UI Tests** | `web/src/**/__tests__/*.test.tsx` | `vitest run` | 37 test suites verifying Primer components, repository views, settings modals, and demo-mode gating. |

---

## 2. End-to-End Agent Loop Verification (`TestAgentLoop_EndToEnd_RealServer`)

The canonical autonomous agent workflow is executed without any mocks against a running `existd` instance:

```mermaid
sequenceDiagram
    autonumber
    actor Agent as Autonomous Agent
    participant MCP as MCP Server (:8080/mcp)
    participant Git as Git Smart-HTTP (:8080)
    participant Hook as Pre-Receive Hook (existd)
    participant Engine as Ruleset & Policy Engine

    Agent->>MCP: lease_request (ttl: 600s, ref: refs/heads/agent/*)
    MCP-->>Agent: Ephemeral HMAC-SHA256 Capability Token
    Agent->>MCP: lease_inspect (verify token validity)
    MCP-->>Agent: Lease Details (active, unrevoked)
    Agent->>Git: git clone http://agent:<token>@host/owner/repo.git
    Git-->>Agent: Cloned Worktree
    Note over Agent: Create branch refs/heads/agent/patch-v1<br/>Commit changes to disk
    Agent->>Git: git push origin agent/patch-v1:refs/heads/agent/patch-v1
    Git->>Hook: Execute hook with EXIST_LEASE_* env
    Hook-->>Git: Allowed (prefix matches refs/heads/agent/*)
    Git-->>Agent: Push Accepted (200 OK)
    Note over Agent: Negative check: attempt push to refs/heads/main
    Agent->>Git: git push origin agent/patch-v1:refs/heads/main
    Git->>Hook: Execute hook with EXIST_LEASE_* env
    Hook-->>Git: Rejection: GH007 (Prohibited ref modification)
    Git-->>Agent: Push Rejected (exit status 1)
    Agent->>MCP: pull_request_create (head: agent/patch-v1, base: main)
    MCP-->>Agent: PR #1 Created
    Agent->>MCP: thread_comment (Post review explanation)
    MCP-->>Agent: Comment Recorded
    Note over Engine: Enforce ruleset: require passing status check 'ci/build'
    Agent->>MCP: pull_request_merge (PR #1)
    MCP->>Engine: EvaluateMerge
    Engine-->>MCP: Blocked (required check 'ci/build' pending)
    MCP-->>Agent: Error (merge blocked by ruleset)
    Agent->>MCP: status_check_report (name: 'ci/build', status: 'completed', conclusion: 'success')
    MCP-->>Agent: Check Run Recorded
    Agent->>MCP: pull_request_merge (PR #1)
    MCP->>Engine: EvaluateMerge
    Engine-->>MCP: Allowed
    MCP-->>Agent: PR #1 Merged (merge commit recorded)
    Agent->>MCP: lease_revoke (leaseId)
    MCP-->>Agent: Lease Revoked
    Note over Agent: Negative check: subsequent push with revoked token
    Agent->>Git: git push with revoked token
    Git-->>Agent: 401 Unauthorized (agent capability lease has been revoked)
```

---

## 3. Mutation Checks (Anti-False-Positive Verification)

Per Project Rule 2 (*No false positives*), every security gate and policy check must be validated with deliberate mutation testing.

### Mutation 1: Capability Lease Ref Prefix Gate
- **Target**: `pkg/agentlease/lease.go` (`CanPushRef`)
- **Mutation**: Force `CanPushRef` to return `true` unconditionally.
- **Expected Failure**: `go test ./pkg/agentlease` fails with:
  ```text
  lease_test.go:44: expected rejection for refs/heads/main, got true
  lease_test.go:47: expected rejection for refs/heads/master, got true
  lease_test.go:50: expected rejection for random branch outside allowed prefix, got true
  ```
- **Observed Result**: Test failed as expected. Code restored, test returned to 100% pass.

### Mutation 2: MCP Merge Policy Evaluation Gate
- **Target**: `internal/mcp/tools.go` (`pull_request_merge` handler)
- **Mutation**: Bypass ruleset blocker evaluation `if false && !evalRes.Allowed`.
- **Expected Failure**: `go test -run TestMCP_CanonicalAgentLoopTools ./internal/mcp` fails with:
  ```text
  server_test.go:535: expected pull_request_merge to be blocked by missing check run, but it succeeded: isError:false
  ```
- **Observed Result**: Test failed as expected. Code restored, test returned to 100% pass.

### Mutation 3: Pre-Receive Hook Default Branch Protection
- **Target**: `internal/security/hook.go` (Agent capability lease ref prefix check)
- **Mutation**: Invert `permitted = false` default to `permitted = true`.
- **Expected Failure**: `go test -run TestPreReceiveHook_AgentLeaseEnforcement ./internal/security` fails on unauthorized branch push.
- **Observed Result**: Test failed as expected. Code restored, test returned to 100% pass.

---

## 4. Running the Complete Verification Suite

```bash
# 1. Run all Go tests with race detection
export PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"
go test -race ./...

# 2. Run end-to-end agent loop integration test
go test -v -run TestAgentLoop_EndToEnd_RealServer ./internal/app

# 3. Run frontend Vitest test suite
cd web
npm test
npm run build
```
