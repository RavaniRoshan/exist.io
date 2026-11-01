# Security Policy

## Supported Versions

Security updates and patches are actively applied to the current release:

| Version | Supported |
| :--- | :--- |
| `0.6.x-beta` | Yes |
| `< 0.6.0` | No |

## Lease Signing Key

Capability leases are signed with HMAC-SHA256 using a 32-byte key resolved at startup:

1. If `EXIST_JWT_SECRET` is set, it is used and must be at least 32 bytes.
2. Otherwise the key is read from `${EXIST_DATA_DIR}/jwt_secret`.
3. If no key exists, one is generated with `crypto/rand`, written with mode `0600`, and reused on restart.

**Rotation requirement.** Any instance started before this change with the previously published literal default key must treat all leases signed with that key as compromised. To rotate: stop the daemon, delete `${EXIST_DATA_DIR}/jwt_secret` (or set a new `EXIST_JWT_SECRET` of at least 32 bytes), and restart. Leases signed with the old key are invalidated.

## Reporting a Vulnerability

The exist.io maintainers take security seriously. If you discover a security vulnerability or potential exploit, please do **NOT** open a public issue.

Please report all vulnerabilities privately via email to:

📧 **`security@exist.io`**

### What to Include
- Detailed description of the vulnerability and attack vector
- Steps to reproduce or proof-of-concept script
- Impact assessment (e.g. unauthorized ref mutation, authentication bypass, data exposure)
- Any proposed remediation or patches

### Response Timeline
- **Acknowledgement**: Within 48 hours of initial report.
- **Triage & Assessment**: Within 5 business days.
- **Fix & Public Disclosure**: Coordinated disclosure after release of patch.
