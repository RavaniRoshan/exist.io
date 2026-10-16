# Product Hunt Screenshot & Media Kit (`SCREENSHOTS.md`)

Product Hunt gallery requires 16:9 or 1200x800 high-resolution images or GIFs showcasing authentic product workflows.

---

### Asset 1: Hero Demo Walkthrough (GIF / Video)
- **File:** `assets/demo.gif` / `assets/demo.mp4` (generated via `assets/demo.tape`)
- **Dimensions:** 1200 x 720 px (Catppuccin Macchiato dark theme)
- **Caption:** *From single-command bootstrap to merged AI agent pull request in under 90 seconds.*
- **What it shows:**
  1. Instant `existd bootstrap` output generating admin PAT and MCP snippets.
  2. Background daemon startup and health check.
  3. Claude Code requesting an ephemeral capability lease.
  4. Git shallow clone and isolated branch push (`refs/heads/agent/*`).
  5. Pull request opened and merged via MCP.

---

### Asset 2: Pull Request & Merge Gate UI
- **File:** `assets/launch/02-web-pr-view.png`
- **Dimensions:** 1280 x 800 px
- **Caption:** *Pixel-accurate Primer UI: Clean pull request view with real-time status checks, agent telemetry badges, and 3-way merge engine.*
- **What it shows:**
  - Active Pull Request `#1` opened by `claude-code`.
  - Passing status checks and code reviewer approval.
  - Flush Primer `BoxRow` diff viewer and squircle `MergeBox`.

---

### Asset 3: Multi-Agent MCP Integration
- **File:** `assets/launch/03-mcp-agent-integration.png`
- **Dimensions:** 1280 x 800 px
- **Caption:** *Vendor-neutral Model Context Protocol: Pair program with Claude Code, Cursor, Cline, or Codex with identical tools.*
- **What it shows:**
  - VS Code with Cline / Cursor side-by-side with terminal running Claude Code.
  - Both agents communicating with the local `http://localhost:8080/mcp` endpoint simultaneously.

---

### Asset 4: Security Defense — Ref Isolation & Secret Scanner
- **File:** `assets/launch/04-agent-ref-isolation.png`
- **Dimensions:** 1200 x 720 px
- **Caption:** *Defense-in-depth: Pre-receive hook rejecting unauthorized pushes to main (GH007) and blocking hardcoded secrets.*
- **What it shows:**
  - An agent attempting `git push origin main` rejected with `GH007: Agent capability lease ref violation`.
  - A commit containing an AWS key rejected with detailed line numbers and redaction snippet.

---

### Asset 5: Empirical Benchmarks & Conformance
- **File:** `assets/launch/05-benchmarks-conformance.png`
- **Dimensions:** 1200 x 720 px
- **Caption:** *Proven performance: Sub-10ms MCP tool latency and 100% API conformance verified on real hardware.*
- **What it shows:**
  - Output of `exist conformance` verifying all active endpoints.
  - Concurrency benchmark table displaying p50 and p95 latencies across 1, 5, and 20 concurrent agents.
