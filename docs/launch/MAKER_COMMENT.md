# Product Hunt First Maker Comment (`MAKER_COMMENT.md`)

*Post as the first comment under the Product Hunt listing immediately upon going live.*

---

Hey Product Hunt community! 👋 I’m Roshan, the creator of **exist.io**.

### Why I built exist.io

Git platforms like GitHub and GitLab were built 15+ years ago for human developers clicking through web browsers. Over the past year, my day-to-day workflow shifted dramatically: AI coding agents (Claude Code, Cursor, Cline, Codex) now write significant portions of my codebase.

When Cursor announced Origin, I loved the concept of an agent-first forge—but I wanted two things they couldn't offer:
1. **Neutrality:** I didn't want my Git infrastructure tied to a single proprietary editor or cloud account.
2. **Self-hosted sovereignty:** I wanted to run the forge on my own hardware, air-gapped or on a cheap VPS, with zero third-party telemetry or vendor lock-in.

So I engineered **exist.io** from first principles.

---

### What makes exist.io different?

Traditional Git permissions are all-or-nothing: either an agent has full SSH/PAT write access to every branch, or it cannot push at all. If an agent hallucinates a `git push origin --force main`, disaster strikes.

exist.io solves this with **first-class agent boundaries**:
- **Ephemeral Capability Leases (`pkg/agentlease`):** Agents are issued short-lived (15–60 min) HMAC-SHA256 leases.
- **Ref Isolation:** Leases are strictly restricted to `refs/heads/agent/*`. If an agent attempts to mutate `main` or delete existing refs, our pre-receive hook intercepts and blocks the push with `GH007`.
- **Pre-Receive Secret Scanner:** Blocks hardcoded AWS keys, GitHub tokens, and unencrypted PEM private keys before commits are written.
- **Native Model Context Protocol (MCP):** Connects out of the box to Claude Code, Cursor, Cline, or Codex with 8 core primitives plus lease management tools.
- **Hermetic Pure-Go Engine:** No CGO, no OpenSSH host dependency, embedded SQLite WAL mode, and embedded React 19 UI. It downloads as a single binary and starts in < 2 seconds.

---

### Truth over marketing: What is real vs. what is frozen

We believe in radical engineering honesty:

- **What is rock-solid and verified today:**
  - Pure-Go Smart-HTTP and embedded SSH daemon (`:2222`).
  - Native MCP server (`/mcp`) and sandboxed exploration (`/mcp/readonly`).
  - Ephemeral capability lease issuance and Git transport enforcement.
  - Pull requests, deterministic 3-way/squash merges, and policy gates.
  - Concurrency benchmarks verified on real hardware ([`docs/benchmarks.md`](https://github.com/RavaniRoshan/exist.io/blob/main/docs/benchmarks.md)).
  - Full claims ledger with line-by-line code proofs in [`CLAIMS.md`](https://github.com/RavaniRoshan/exist.io/blob/main/CLAIMS.md).

- **What is frozen as experimental for launch:**
  - Enterprise federation (SAML 2.0, SCIM 2.0, OIDC): Implemented and passing unit tests, but frozen pending external security audit.
  - Git LFS locking (`/info/lfs/locks`): CAS is verified; binary locking is labeled experimental.
  - Terraform provider & Helm chart: Kept compiling, frozen for post-launch roadmap.

---

### How to try it in 60 seconds

```bash
# 1. Download or build from source
go build -o bin/existd ./cmd/existd

# 2. Bootstrap your admin account and first repo
./bin/existd bootstrap --email admin@exist.io --namespace acme

# 3. Start the daemon and open http://localhost:8080
./bin/existd
```

Copy the generated Claude Code or Cursor config snippet from the terminal output, point your favorite agent at it, and watch it open its first isolated PR!

I'd love your brutal feedback, feature requests, and bug reports. I'll be in the comments all day! 🚀
