# Claims Reconciliation — EXIST-OPS-001 WP-A2

Baseline: `53bce30`. Governing ledger: `CLAIMS.md`. Code governs numeric facts (DR-08).
Every row has an executed action; none are left open.

| # | Location | Statement as written | REF-01 status | Code status at baseline | Action taken |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | `site/docs/index.mdx:11` | "wire-compatible with Origin forge interfaces" | Unverified (entry 1) | No public Origin specification; no conformance proof | Deleted the sentence |
| 2 | `site/docs/index.mdx` core list | "Git LFS & CAS storage" as a core capability | Verified but frozen (entry 7) | Implemented and tested | Moved to "Experimental modules (frozen for launch)" |
| 3 | `site/docs/index.mdx` core list | "Enterprise Federation: OIDC & SAML 2.0 … SCIM 2.0" as core | Experimental (entry 12) | Implemented; callback trusts client input | Moved to experimental and stated the callback fails closed |
| 4 | `site/docs/index.mdx` core list | "exposing 34 specialized tools" | Unverified count (DR-03) | Registry has a different count | Removed the count; links to the MCP reference |
| 5 | `site/docs/index.mdx` core list | "Official Terraform / OpenTofu provider and Kubernetes Helm chart" as core | Experimental | Source present under `deploy/` | Moved to experimental, unsupported |
| 6 | `site/docs/index.mdx` cards | Core cards for SSO/SCIM and Terraform/Helm | Experimental | Frozen | Retitled "Experimental" with constraint text |
| 7 | `site/docs/overview.mdx` included table | SSO/SCIM row | Experimental | Frozen | Removed from "Included"; added to experimental list |
| 8 | `site/docs/overview.mdx` included table | Terraform/Helm row | Experimental | Frozen | Removed from "Included"; added to experimental list |
| 9 | `site/docs/overview.mdx` included table | "Git LFS v1 Batch CAS engine & file locking" | Verified but frozen | Implemented | Moved to experimental list |
| 10 | `site/docs/overview.mdx` included table | "Embedded MCP server (34 tools)" | Unverified count | Count differs | Removed count |
| 11 | `site/docs/overview.mdx` included table | "Web UI Living Dashboard … Space UI" | False naming | Custom React components, not Space UI | Renamed to "Web UI (React 19 / Tailwind CSS)" |
| 12 | `site/docs/roadmap.mdx:30` | "Enterprise Identity Federation & SCIM 2.0 (Complete)" | Experimental | Frozen | Relabeled "Experimental — frozen for launch"; added 501 note |
| 13 | `site/docs/roadmap.mdx:35` | "Official Infrastructure as Code (IaC) (Complete)" | Experimental | Frozen | Relabeled "Experimental — frozen for launch" |
| 14 | `site/docs/roadmap.mdx:45` | "exposing 35 specialized tools" | Unverified count | Count differs | Removed count |
| 15 | `site/docs/roadmap.mdx:48` | "Space UI Living Dashboard" | False naming | Custom components | Renamed "Web UI & Environment Separation" |
| 16 | `site/docs/development/status.mdx:23` | "REST API surface (119 operations)" | Unverified count | Count not reproduced | Removed count; conformance command cited |
| 17 | `site/docs/development/status.mdx:24` | "100% coverage of the 119 Origin REST API operations" | Unverified | No external conformance proof | Replaced with "checked with `exist conformance`" |
| 18 | `site/docs/development/status.mdx:33` | "34 specialized agent tools" | Unverified count | Count differs | Removed count |
| 19 | `site/docs/development/status.mdx:49` | "Production-ready Terraform/OpenTofu provider" | Experimental | Frozen | Relabeled experimental, unsupported |
| 20 | `site/docs/development/status.mdx:59` | "SCIM 2.0 … Enterprise SSO" as shipped | Experimental | Frozen | Relabeled experimental; added 501 note |
| 21 | `site/docs/architecture/index.mdx:17` | "Space UI Living Dashboard" | False naming | Custom components | Renamed |
| 22 | `site/docs/architecture/index.mdx:18` | IdP node presented as integrated | Experimental | Frozen | Labeled experimental |
| 23 | `site/docs/architecture/index.mdx:19` | IaC node presented as integrated | Experimental | Frozen | Labeled experimental |
| 24 | `site/docs/architecture/index.mdx:28` | "34 AI Agent Tools" | Unverified count | Count differs | Removed count |
| 25 | `site/docs/architecture/index.mdx:29` | "119 Forge Operations" | Unverified count | Count not reproduced | Removed count |
| 26 | `site/docs/architecture/index.mdx:30` | Federation engine presented as shipped | Experimental | Frozen | Labeled experimental |
| 27 | `site/docs/architecture/index.mdx:83` | "Enterprise Controls" including SSO/SCIM | Experimental | Frozen | Labeled experimental |
| 28 | `site/docs/architecture/index.mdx:123` | Federation section presented as shipped | Experimental | Frozen | Labeled experimental; added 501 note |
| 29 | `site/docs/architecture/index.mdx:133` | "Exposes 34 specialized tools" | Unverified count | Count differs | Removed count |
| 30 | `site/docs/getting-started/index.mdx:84` | "Autonomous AI agent 35+ Git tools" | Unverified count | Count differs | Removed count |
| 31 | `site/docs/getting-started/index.mdx:88` | SCIM endpoint row in core table | Experimental | Frozen | Removed from core table; added experimental note |
| 32 | `site/docs/reference/mcp-tools.mdx:3` | "all 35 MCP tools" | Unverified count | Count differs | Removed count; points to `tools/list` as authoritative |
| 33 | `site/docs/reference/mcp-tools.mdx:7` | "35 specialized tools" | Unverified count | Count differs | Removed count |
| 34 | `site/docs/how-to/enterprise-sso-scim.mdx:6` | SSO/SCIM presented as available | Experimental | Callback trusts client input | Added "Not available in this build" caution; B2 makes the callback fail closed |
| 35 | `site/docs/how-to/terraform-helm.mdx:6` | IaC presented as official and production-ready | Experimental | Frozen | Added experimental/frozen caution |
| 36 | `README.md` comparison matrix | Rows asserting Cursor Origin, GitHub, and Forgejo capabilities | Not covered by REF-01 | No evidence in repo | Replaced with a code-backed "Verified Capabilities" table |
| 37 | `README.md:216` | "37 suites, 137 tests" | Unverified count | Test count changes | Removed the count |
| 38 | `README.md:199` | "frozen for the v0.5.0-beta release" | Version drift | Latest tag is v0.5.0-beta | Changed to "frozen for launch" |
| 39 | `CLAIMS.md` row 1 remediation | "replaced with verified comparison matrix" | Stale action | Comparison matrix removed | Corrected the remediation text |
| 40 | `CLAIMS.md` whole table | No verification commit recorded | Requirement gap | n/a | Added `Verified at commit` column |
| 41 | `CLAIMS.md` (new block) | No machine-enforced forbidden phrases | Requirement gap | n/a | Added the `forbidden-claims` block consumed by `internal/claims/coherence_test.go` |

## Verification

- `go test ./internal/claims -run TestPublishedClaimsAreCoherent -v` — failed on 12 baseline phrases, now passes.
- `grep -rn "34 specialized tools\|35 specialized tools\|35+ Git tools\|wire-compatible with Origin\|Space UI" README.md site/docs` — zero matches.
- `./scripts/check-links.sh` — passes.
- `cd site && npm run build` — passes.
