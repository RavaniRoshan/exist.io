# Launch Day Social Media Kit (`SOCIAL_POSTS.md`)

Drafted announcements tailored for developer communities and social networks on launch day (Tuesday, November 17, 2026).

---

## 1. Twitter / X (Launch Thread)

### Tweet 1 (Hook & Announcement)
Traditional Git platforms were built 15+ years ago for humans clicking in browsers.

Today, we're launching **exist.io**: the neutral, self-hosted Git forge built from first principles for autonomous AI coding agents.

Single binary. Pure Go. Zero cgo. Zero cloud lock-in. 🧵👇

[Link to Product Hunt / GitHub]

### Tweet 2 (The Problem with Current Tools)
Cursor Origin proved the world wants an agent-first forge—but tying your Git infrastructure to a single proprietary editor or cloud account is a non-starter for self-hosters and teams.

exist.io is 100% vendor-neutral: Claude Code, Cursor, Cline, or Codex all connect over native MCP.

### Tweet 3 (Agent Capability Leases)
The nightmare of agent pair programming is a hallucinated `git push --force origin main`.

exist.io solves this with **Capability Leases**:
• Ephemeral HMAC-SHA256 tokens (15–60 min TTL)
• Strictly isolated to `refs/heads/agent/*`
• Pre-receive hook blocks pushes to main (GH007)
• In-memory secret scanner halts pushed AWS/API keys

### Tweet 4 (Architecture & Performance)
Under the hood:
• Single Go binary (`existd`) with embedded React 19 UI
• Embedded pure-Go SSH server on :2222 (no host OpenSSH daemon)
• SQLite WAL mode with zero lock contention
• Sub-10ms MCP tool latency, sub-3s full agent loop (clone, push, PR)

### Tweet 5 (Try it in 60s)
Install it in 60 seconds on any Linux, macOS, or Windows machine:
`go build ./cmd/existd && ./existd bootstrap`

We published an empirical claims ledger verifying every single technical claim in our README:
https://github.com/RavaniRoshan/exist.io/blob/main/CLAIMS.md

Feedback and stars welcome! ⭐🚀

---

## 2. Hacker News (`Show HN`)

**Title:** Show HN: exist.io – Neutral, self-hosted Git forge built for AI coding agents

**URL:** https://github.com/RavaniRoshan/exist.io

**Text:**
Hey HN,

Over the past year, my daily workflow moved from writing code manually to directing autonomous AI agents (Claude Code, Cline, Cursor). But traditional Git platforms weren't designed for agents pushing dozens of automated commits an hour, and proprietary agent forges tie your repo to closed cloud ecosystems.

I built exist.io to be an open, self-hosted forge where coding agents are first-class citizens alongside human engineers.

Key technical decisions:
1. **Hermetic pure-Go architecture:** Zero CGO, zero host OpenSSH dependency. It compiles with `CGO_ENABLED=0` and ships as a single binary with the React 19 web UI embedded via `go:embed`.
2. **Embedded SSH server (:2222):** Custom pure-Go SSH transport built on `golang.org/x/crypto/ssh`, persisting host keys and enforcing repository boundaries without touching host files.
3. **Capability Leases (`pkg/agentlease`):** Agents don't get blanket write keys. They receive short-lived HMAC-SHA256 capability tokens restricted to `refs/heads/agent/*`. The pre-receive hook inspects commits and blocks unauthorized ref updates (`GH007`) or hardcoded credentials.
4. **Native MCP Server:** Implements standard Model Context Protocol at `/mcp` and an isolated read-only sandbox at `/mcp/readonly`.
5. **Truth over marketing:** We audited every claim in the repository into a verified claims ledger (CLAIMS.md) and published reproducible benchmarks (bench/main.go) showing sub-10ms MCP latency and sub-3s end-to-end agent loops under 5x concurrency.

Code is Apache-2.0 licensed. Would love your thoughts on agent permissions, self-hosted architectures, and edge cases.

---

## 3. Reddit — r/selfhosted

**Title:** I built exist.io: A lightweight, self-hosted Git forge built for AI coding agents (Pure Go, single binary, embedded UI & SSH)

**Post:**
Hey r/selfhosted,

I wanted to share a project I've been working on: **exist.io** (https://github.com/RavaniRoshan/exist.io).

Like many of you, I self-host my development stack and dislike being forced into proprietary cloud accounts to use new developer tools. As I started using agents like Claude Code and Cline more often, giving them unrestricted PATs to my repos felt dangerous.

### What is it?
A self-hosted Git forge (think lightweight, agent-native Gitea/Forgejo) designed for both humans and AI agents.

### Why you might find it interesting:
- **True single-binary deployment:** No Docker required (though compose files are provided). It embeds SQLite in WAL mode (`modernc.org/sqlite`, pure Go, zero CGO), the Web UI, and a pure-Go SSH server on port 2222.
- **Ref isolation for agents:** Agents authenticate using capability leases that restrict pushes strictly to `refs/heads/agent/*`. If an agent hallucinates a push to `main` or tries to delete a branch, the pre-receive hook rejects it.
- **Pre-receive secret scanner:** Rejects pushes containing AWS keys, GitHub tokens, or private keys before they hit the disk.
- **Runs anywhere:** Linux (amd64/arm64), macOS, and Windows. Uses ~35MB RAM on idle.
- **100% Open Source:** Apache-2.0 license.

You can try it with a single command:
```bash
existd bootstrap --email admin@exist.io --namespace myteam
./existd
```

Check out our claims audit (`CLAIMS.md`) and let me know what you think!

---

## 4. Reddit — r/LocalLLaMA

**Title:** exist.io — An open-source, self-hosted Git forge with native MCP & capability lease isolation for local agents

**Post:**
Hey r/LocalLLaMA,

If you run local coding models or pair with tools like Claude Code, Cline, or Cursor, you've probably noticed that giving an autonomous agent direct write access to your Git branches is risky.

We just released **exist.io** (https://github.com/RavaniRoshan/exist.io), an open-source (Apache-2.0) Git forge built specifically for agentic workflows:

- **Native Model Context Protocol (MCP) server at `/mcp`:** 8 core tools (`repo_tree`, `file_read`, `diff_compare`, `symbol_search`, `pull_request_create`, `thread_comment`, `pull_request_merge`) plus lease management.
- **Capability Leases:** Ephemeral HMAC-signed leases that lock agents into `refs/heads/agent/*`. Pushing to main or deleting branches is intercepted and blocked at the Git protocol level.
- **Read-Only Sandbox (`/mcp/readonly`):** Low-privilege sandbox for untrusted exploration or review agents.
- **Pure-Go, Zero CGO:** Built with Go + SQLite WAL. Single binary, no external database or dependencies required.

We ran a 20-agent concurrency benchmark on local hardware: lease issuance was 5ms p50, and full agent loops (lease -> clone -> branch -> push -> PR) completed in ~2.7s. Detailed numbers in `docs/benchmarks.md`.

Let me know what local agent integrations or tool primitives you'd like to see next!

---

## 5. LinkedIn

**Post:**
Traditional version control platforms were designed for human beings interacting via browser tabs. But software engineering is undergoing a generational shift: autonomous AI coding agents now read code, generate diffs, and push commits continuously.

Today, we're introducing **exist.io** — the neutral, self-hosted Git forge built from first principles for autonomous coding agents (Claude Code, Cursor, Cline, Codex) alongside engineering teams.

Key Highlights:
🔹 **Vendor-Neutral Sovereignty:** Runs on your own server or private cloud. Zero external proprietary dependencies.
🔹 **Cryptographic Capability Leases:** Short-lived leases with ref-prefix isolation (`refs/heads/agent/*`) prevent rogue pushes to production branches.
🔹 **Hermetic Pure-Go Engine:** Single static binary with embedded React 19 UI, embedded SSH (:2222), and SQLite WAL storage.
🔹 **Native Model Context Protocol (MCP):** Connect your favorite agents out of the box with zero custom glue code.
🔹 **Empirical Transparency:** Every claim backed by our verified claims ledger and reproducible benchmarks.

Explore the open-source repository (Apache-2.0):
https://github.com/RavaniRoshan/exist.io

#DevOps #OpenSource #SoftwareEngineering #ArtificialIntelligence #Git #AIagents #MCP
