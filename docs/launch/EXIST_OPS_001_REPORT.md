# EXIST-OPS-001 Completion Report

**Directive:** EXIST-OPS-001 Rev 1.0  
**Baseline:** `53bce30`  
**Branch:** `launch/exist-ops-001`  
**Date:** 2026-11-01

## Status by phase

| Phase | Status | Evidence |
| :--- | :--- | :--- |
| A — Integrity | Complete | `go test ./internal/claims`, `scripts/check-links.sh`, `gencounts -check`, `LAUNCH_READINESS.md` |
| B — Security | Complete | `internal/config` JWT tests, `internal/auth` SSO tests, staticcheck/gosec/govulncheck/gitleaks, installer harness |
| C — Release mechanics | Code complete; cut pending | `VERSION`, `existd --version`, `/healthz`, installer resolution, release workflow |
| D — Measurement | Complete | `internal/telemetry`, `docs/launch/METRICS.md` |
| E — Activation | Complete | `internal/app/activation.go`, `/api/v1/admin/activation`, `ACTIVATION_RUNBOOK.md` |
| F — Media | Partial | Asset 02 captured; 03–05 specified in `CAPTURE_REQUIRED.md` |
| G — Closure | This report + scorecard | `LAUNCH_READINESS.md` |

## Defect index (Appendix A was not supplied; D-01–D-13 and D-17–D-20 are from the directive, D-14–D-16 inferred)

| ID | Defect | Resolution |
| :--- | :--- | :--- |
| D-01 | CI gates cannot fail | `|| true` removed; staticcheck/gosec/govulncheck/gitleaks/gofmt fail the build |
| D-02 | Hardcoded lease signing key | Generated with `crypto/rand`, persisted 0600 under the data dir |
| D-03 | SSO identity spoofing | Callback returns 501; no identity from client input; baseline test proof captured |
| D-04 | Installer installs unverified binaries | Fails closed in four cases; warned override only |
| D-05 | Installer prints `exist init` | Replaced with the bootstrap-first sequence; consistency check in CI |
| D-06 | Installer pins an old release | Resolves the newest GitHub release; no hardcoded pin |
| D-07/D-08 | Version discontinuity | `VERSION` is the single source; injected into both binaries and `/healthz` |
| D-09 | Origin wire-compatibility claim | Removed from the documentation site |
| D-10 | Federation/IaC presented as core | Moved to experimental sections with constraints |
| D-11 | Tool count drift | Generated counts; 46 actual tools; UI reads the API count |
| D-12 | Assertion divergence | Coherence test + `CLAIMS_RECONCILIATION.md` |
| D-13 | Two documentation domains | Zero `exist.os` strings; repo description fixed; H-02 recorded |
| D-14 (inferred) | CI Go pin vs `go.mod` | Workflows read `go-version-file: go.mod` |
| D-15 (inferred) | Docker Go pin + stale version ldflags | Docker uses Go 1.26 and `VERSION` |
| D-16 (inferred) | No `.dockerignore` | Added, covering data, credentials, and build outputs |
| D-17 | Dead local-scheme links | Zero; link checker covers the scheme |
| D-18 | Unevidenced readiness scorecard | Reissued with commands; Gate 6 BLOCKED, Gate 7 NOT VERIFIED |
| D-19 | No measurement | Consent-gated telemetry with a five-field payload |
| D-20 | No activation definition | Defined in `METRICS.md` and detected once per instance |

## Verification log

```text
go build ./...                                             PASS
go test -race ./...                                        PASS
gofmt -l .                                                 empty
staticcheck ./...                                          PASS
gosec ./...                                                0 issues (53 line-scoped justifications)
govulncheck ./...                                          0 affecting code (Go 1.27.3)
gitleaks detect --source . --redact                        0 leaks (81 commits)
./scripts/check-links.sh                                   PASS (22 external URLs, 40 files)
./scripts/check_quickstart_consistency.sh                  PASS
./scripts/test_install_verification.sh                     PASS (5 cases)
./scripts/test_first_run.sh                                PASS (2 s)
./scripts/test_first_run.sh --mutation                     PASS (rejected)
go run ./internal/tools/gencounts -check                   PASS
cd web && npm test                                         41 suites / 153 tests PASS
cd web && npm run build                                    PASS
cd site && npm run build                                   PASS (21 pages)
```

Security demonstrations:

- SSO test fails on `53bce30` (callback returned 200 and provisioned a user), passes after the fix (501).
- staticcheck fails on an injected unused symbol; gosec fails on an injected `AKIA…` credential.
- gitleaks history scan: zero findings.

## Deviations

1. **Installer release resolution (C2).** The literal rule (`releases/latest`) resolves to `v0.4.7-enterprise`; the implementation takes the newest release including prereleases, preserving the "installer installs the documented release" invariant.
2. **Release-cut timing.** The tag/publish step runs after the measurement and UI work so the release binary contains them. VG-3 mechanics are implemented and verified against local builds.
3. **Derived D–G.** The pasted directive stopped mid-Phase D; D–G were derived from Section 1.3, DR-07, REF-04, and REF-06.
4. **UI redesign scope.** The Forge foundation, shadcn/React Aria primitives, shell, and key surfaces are rebuilt; view-by-view migration continues after the launch cut. Asset 02 is captured from the new UI.
5. **Media.** Assets 03–05 are specified in `CAPTURE_REQUIRED.md`; they were not captured because no IDE session exists and the terminal captures were not produced in this session.

## Human actions

- **H-01** Branch protection was set successfully; no action.
- **H-02** `RavaniRoshan/exist.os` remains published outside this working tree and needs manual archival by the owner.
- **H-03** Gate 6 needs one named external tester to complete the agent loop.
- **H-04** If Actions cannot publish, run `./scripts/build-release.sh v0.6.0-beta` and `gh release create`.
- **H-05** No real secret was found; no rotation is required.

## Remaining before Product Hunt

1. Merge PR #2 and cut `v0.6.0-beta`.
2. Capture assets 03–05 per `CAPTURE_REQUIRED.md`.
3. Continue the view-by-view Forge migration (the foundation and shell are in place).
4. Complete Gate 6 with an external tester.
