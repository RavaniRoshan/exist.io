# Security Audit Report (`docs/security-audit.md`)

Date: October 16, 2026  
Tool: `govulncheck` v1.9.0 (Go 1.27.2)  
Scope: Entire exist.io repository (`./...`, 28 root packages, 22 direct and transitive modules)

---

## Executive Summary

`govulncheck` was executed against all 28 packages across the repository.

```
=== Symbol Results ===
No vulnerabilities found.

=== Package Results ===
No other vulnerabilities found.

Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 1
vulnerability in modules you require, but your code doesn't appear to call these
vulnerabilities.
```

---

## Transitive Dependency Analysis & Exception Rationale

### Exception 1: `GO-2026-5932` in `golang.org/x/crypto`

- **Vulnerability ID:** [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932)
- **Module:** `golang.org/x/crypto@v0.57.0`
- **Component:** `golang.org/x/crypto/openpgp` (deprecated)
- **Description:** The legacy `openpgp` package in `x/crypto` is deprecated upstream and unmaintained.
- **Transitive Source:** Required by `github.com/go-git/go-git/v6` as part of legacy PGP support.
- **Usage Analysis & Exception Rationale:**
  - `exist.io` **does not import or invoke** `golang.org/x/crypto/openpgp`.
  - OpenPGP commit signature parsing and verification in exist.io uses `github.com/ProtonMail/go-crypto/openpgp` (`v1.5.2`), which is actively maintained and secure.
  - No symbol from `golang.org/x/crypto/openpgp` is referenced or reachable from exist.io production binaries (`existd`, `exist`).
  - **Verdict:** Accepted transitive risk with zero reachable attack surface.

---

## Resolved Vulnerabilities in this Audit Cycle

During this security baseline hardening pass, the following transitive dependencies were upgraded:

- **Upgraded `golang.org/x/net` from `v0.59.0` to `v0.60.0`**:
  - Eliminated `GO-2026-6617` (HPACK encoder race in net/http)
  - Eliminated `GO-2026-6612` (Double flow control refund on HTTP/2 server streams)
  - Eliminated `GO-2026-6611` (Excessive CPU consumption from repeated initial window changes)
  - Eliminated `GO-2026-6610` (HTTP/2 transport accepts malformed framing headers)
  - Eliminated `GO-2026-6603` (HTTP/2 server memory exhaustion via Trailer headers)

---

## Hardened Security Controls

1. **Pre-Receive Secret Scanner (`internal/security/scanner.go`):**
   - Active regex scanning for unencrypted PEM/OpenSSH private keys, GitHub PATs/OAuth/fine-grained tokens, AWS access/secret keys, Slack tokens, high-entropy API keys, and JWT secrets.
   - 100% test coverage verified with `internal/security/scanner_test.go`.
2. **SSH Server Hardening (`internal/sshserver/server.go`):**
   - Modern cipher suites only (`chacha20-poly1305@openssh.com`, `aes256-gcm@openssh.com`, `aes128-gcm@openssh.com`, `aes256-ctr`).
   - 30-second TCP handshake deadline to prevent Slowloris / connection exhaustion attacks.
   - Persisted host key across daemon restarts.
3. **HTTP Server Hardening (`cmd/existd/main.go`):**
   - Localhost default binding (`127.0.0.1:8080`).
   - ReadHeaderTimeout (10s), ReadTimeout (60s), WriteTimeout (120s), IdleTimeout (120s), MaxHeaderBytes (1MB).
4. **Fuzz Testing:**
   - Git pkt-line parser (`internal/gittransport/pktline_fuzz_test.go`): 645,188 executions in 30.4s without failures.
   - Capability lease token parser (`pkg/agentlease/lease_fuzz_test.go`): 3,062,243 executions in 31.1s without failures.
