# Capture Required (`CAPTURE_REQUIRED.md`)

Assets 3–5 were not captured in this environment. Each entry below is the exact specification for a real screen capture. Do not synthesize these images.

## Asset 3 — Multi-Agent MCP Integration

- **File:** `assets/launch/03-mcp-agent-integration.png`
- **Viewport:** 1280 x 800
- **URL path:** an IDE window with Cline or Cursor open beside a terminal running Claude Code, both pointed at `http://localhost:8080/mcp`.
- **Seed data:** a bootstrapped instance (`scripts/capture/seed.sh`), two agent sessions connected with the printed API key, and one open agent pull request.
- **Caption:** *Vendor-neutral Model Context Protocol: pair program with Claude Code, Cursor, Cline, or Codex with identical tools.*
- **Acceptance:** both agent clients visibly connected to the same instance; the PR list visible in the browser behind them; no composited windows.

## Asset 4 — Security Defense: Ref Isolation & Secret Scanner

- **File:** `assets/launch/04-agent-ref-isolation.png`
- **Viewport:** 1200 x 720 terminal
- **URL path:** n/a
- **Seed data:** a lease token restricted to `refs/heads/agent/*`. Two commands:
  1. `git push origin HEAD:main` → rejected with `GH007`/ref violation.
  2. push a commit containing `AKIAIOSFODNN7EXAMPLE` to `agent/*` → rejected by the pre-receive secret scanner with line numbers.
- **Caption:** *Defense-in-depth: the pre-receive hook rejects unauthorized pushes to main and blocks hardcoded secrets.*
- **Acceptance:** both rejections visible in one terminal capture; exit status non-zero; no edited output.

## Asset 5 — Empirical Benchmarks & Conformance

- **File:** `assets/launch/05-benchmarks-conformance.png`
- **Viewport:** 1200 x 720 terminal
- **URL path:** n/a
- **Seed data:** a running instance; run `exist conformance --endpoint http://localhost:8080 --token <key>` and display the measured table from `docs/benchmarks.md`.
- **Caption:** *API conformance and the measured concurrency table from the current harness.*
- **Acceptance:** conformance output and the table are from the current build; the capture is a real terminal, not a rendered mockup.
