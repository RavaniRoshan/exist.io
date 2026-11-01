# Launch Readiness Scorecard (`LAUNCH_READINESS.md`)

**Gates passed: 5 · Blocked: 1 · Not verified: 1**
**Evaluated: 2026-11-01 · Commit: `9baeec3` · Release: `v0.6.0-beta`**

A gate is **PASS** only when the cited command exits zero and its output supports the claim (EXIST-OPS-001 DR-10). A gate is **BLOCKED** when it depends on an action the executing agent cannot perform. **NOT VERIFIED** means evidence could not be produced in this environment.

This revision supersedes the 2026-10-17 scorecard, which asserted readiness in its header while recording Gate 6 as blocked and Gate 7 as passed without an artefact.

---

## Gate summary

| Gate | Criterion | Status | Reproducible evidence |
| :--- | :--- | :---: | :--- |
| **Gate 1** | Clean install and first run in < 5 minutes | 🟢 **PASS** | `./scripts/test_first_run.sh` — exit 0, agent loop completed in 2 s (target < 180 s). |
| **Gate 2** | 90-second demo runs end-to-end without manual intervention | 🟢 **PASS** | `vhs /tmp/demo.tape` — exit 0, produced a 2.7 MB GIF and 1.5 MB MP4 from a running instance. |
| **Gate 3** | Claims ledger has 0 false entries and 0 claims without evidence | 🟢 **PASS** | `go test ./internal/claims -v` — exit 0; `docs/CLAIMS_RECONCILIATION.md` committed with no empty actions. |
| **Gate 4** | Zero unverified claims in the README and docs site | 🟢 **PASS** | `go test ./internal/claims -v` — exit 0; forbidden-phrase scan returns zero matches. |
| **Gate 5** | Security audit clean (0 reachable vulnerabilities) | 🟢 **PASS** | `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — "Your code is affected by 0 vulnerabilities." |
| **Gate 6** | External beta user completed the agent loop | 🔴 **BLOCKED** | Requires a named external person. The executing agent cannot satisfy this gate and did not attempt to. |
| **Gate 7** | Positioning test: 3 of 5 target users say "I want this" | ⚪ **NOT VERIFIED** | No recorded artefact exists in the repository. The prior PASS cited no evidence. |

---

## Gate detail

### Gate 1 — Clean install and first run

- **Command:** `./scripts/test_first_run.sh`
- **Observed:** exit 0; "PASS: exist.io first-run installation and agent loop verified in 2s".
- **What it exercises:** CGO-free build, offline bootstrap, daemon start, health probe, MCP lease request, git clone, agent branch push under the lease, MCP PR create, MCP PR merge, and a post-merge content check on `origin/main`.
- **Mutation check:** `./scripts/test_first_run.sh --mutation` — exit 0; "MUTATION CHECK PASSED: Clone rejected with invalid lease token as expected."
- **Target:** < 180 s. Observed runs: 2 s and 7 s on the evaluation host.

### Gate 2 — Unattended 90-second demo

- **Commands:**
  ```bash
  sed 's|Output "assets/demo.gif"|Output "/tmp/exist-demo.gif"|; s|Output "assets/demo.mp4"|Output "/tmp/exist-demo.mp4"|' assets/demo.tape > /tmp/demo.tape
  vhs /tmp/demo.tape
  ```
- **Observed:** exit 0; `Creating /tmp/exist-demo.gif`, `Creating /tmp/exist-demo.mp4`; 2.7 MB GIF, 1.5 MB MP4.
- **Note:** the tape no longer contains host-specific paths; it resolves the repository root at runtime.

### Gate 3 — Claims ledger integrity

- **Command:** `go test ./internal/claims -run TestPublishedClaimsAreCoherent -v`
- **Observed:** PASS.
- **Supporting artefact:** `docs/CLAIMS_RECONCILIATION.md` records one row per capability statement with a non-empty action.
- **Baseline evidence:** the same test failed on 12 forbidden phrases before reconciliation; the failure output is recorded in the EXIST-OPS-001 report.

### Gate 4 — Zero unverified claims in published surfaces

- **Command:** `go test ./internal/claims -v`
- **Observed:** PASS. The test reads the `forbidden-claims` block in `CLAIMS.md` and scans `README.md` and `site/docs/**/*.mdx`.
- **Additional check:** `git grep -n "34 specialized tools\|35 specialized tools\|35+ Git tools\|wire-compatible with Origin\|Space UI"` — zero matches.

### Gate 5 — Security baseline

- **Command:** `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
- **Toolchain:** Go 1.27.3 (pinned in `go.mod` under `toolchain`).
- **Observed:** "Your code is affected by 0 vulnerabilities." The scan also found 3 vulnerabilities in required modules that are not reachable from the code.
- **Scope note:** this gate proves reachable-vulnerability status only. It does not assert that every CI security tool is fail-closed; that is WP-B3.

### Gate 6 — External beta user

**BLOCKED.** The condition is that a named external person, not the maintainer, installs exist.io, connects an agent, and merges a pull request. The executing agent cannot perform this action or fabricate its result. This gate remains blocked until a human tester completes the loop and their result is recorded.

### Gate 7 — Positioning validation

**NOT VERIFIED.** No interview notes, survey results, or recorded test artefact exists in the repository. The 2026-10-17 revision marked this PASS without evidence. It is not a PASS until a recorded artefact is committed.

---

## Release integrity (VG-3)

- **Tag:** `v0.6.0-beta` at `9baeec3`, published by the release workflow.
- **Assets:** 5 platform archives (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`) plus `SHA256SUMS`.
- **Installer:** `INSTALL_DIR=/tmp/install-final bash scripts/install.sh` resolved `v0.6.0-beta`, verified the checksum, and installed it.
- **Version surfaces:** `existd v0.6.0-beta`, `exist version v0.6.0-beta`, `GET /healthz` → `v0.6.0-beta`, latest tag `v0.6.0-beta`, and the installer resolution agree. The README badge is dynamic (`shields.io/github/v/release`).
- **Install matrix:** `linux-amd64`, `linux-arm64`, `darwin-arm64`, and `windows-amd64` jobs succeeded in run `41332477902`.

## Reproduce this scorecard

```bash
git clone https://github.com/RavaniRoshan/exist.io
cd exist.io
git checkout 68e01e4

# Gate 1
./scripts/test_first_run.sh
./scripts/test_first_run.sh --mutation

# Gate 2 (requires vhs and ffmpeg)
sed 's|Output "assets/demo.gif"|Output "/tmp/exist-demo.gif"|; s|Output "assets/demo.mp4"|Output "/tmp/exist-demo.mp4"|' assets/demo.tape > /tmp/demo.tape
vhs /tmp/demo.tape

# Gates 3 and 4
go test ./internal/claims -v

# Gate 5
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Expected: Gates 1–5 exit 0 with the outputs quoted above. Gates 6 and 7 cannot be reproduced by a command.

---

## What this scorecard does not assert

- It does not assert that CI security gates fail closed. That is WP-B3 and is recorded separately.
- It does not assert external validation or positioning.
