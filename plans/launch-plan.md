# exist.io: Realignment and Product Hunt Launch Plan

Drafted October 16, 2026. Target launch: Tuesday, November 17, 2026. Fallback: Tuesday, December 1, 2026 (skip Thanksgiving week). Four weeks is tight for one person, so the launch gate below decides the date, not the calendar.

## The thesis in three lines
• exist.io is the neutral, self-hosted Git forge for coding agents. It is not an Origin clone and not a GitHub clone.  
• The launch has to pass one test: a stranger installs it in 5 minutes and watches an agent go from lease to merged pull request, and the whole thing fits in a 90-second demo.  
• Anything that does not serve that test is frozen until after launch.  

---

## Part 1: Fixes (do these first)
These make the project honest and coherent. Nothing new gets built until they are done.

### F1. Reposition
• Drop 'alternative to Cursor Origin' as the headline. It borrows a competitor's news cycle and rots if their API changes. Keep it as a one-line comparison lower in the README.  
• The wedge: Origin is tied to Cursor accounts. exist.io is vendor-neutral and runs on your own server, for any agent (Claude Code, Codex, Cline, Cursor).  
• Tagline candidates (Product Hunt limit is 60 characters; verify current limits in the submit form):  
    ◦ Self-hosted Git forge built for coding agents (45)  
    ◦ Any agent, your server: a self-hosted Git forge (47)  
    ◦ The neutral Git forge for coding agents (39)  

### F2. Claims audit
Build a claims ledger (`CLAIMS.md`): every claim in the README, its evidence (a test or a file), and its status. No evidence means the claim is deleted or softened. Known problem claims:  
• Full wire compatibility with Origin forge interfaces: delete it, or replace it with a conformance table of what is actually matched and tested. No public Origin API spec was found, so the 119-operation REST surface needs a documented source.  
• Enterprise-ready: delete until someone external runs it and the auth code is reviewed.  
• 100% open source with zero external proprietary dependencies: verify (S3 replication in the enterprise compose, the SQLite driver, the zero-cgo claim).  
• 35 MCP tools and 119 REST operations: report only counts that have tests, and trim the MCP surface to a core set.  
• Tests passed and version badges: they are hardcoded strings. Replace them with CI-driven badges.  

### F3. README and docs contradictions
• Two site URLs (exist.os and exist.io): pick one, fix every link, add a link checker to CI.  
• The quickstart says to open :8080, the endpoint table puts the Web UI on :3000, and the build steps never build the web UI. Embed the built UI in the daemon so one binary serves everything.  
• The quickstart must be copy-paste correct from a clean machine, and that is tested in CI (Tier 2).  

### F4. Scope freeze

| Area | Decision |
| --- | --- |
| Smart-HTTP and SSH transport, repos, PRs, 3-way merge, merge queue | Keep and deepen |
| MCP server (core tools), read-only MCP sandbox, capability leases, agent PR review, audit log | Keep and deepen |
| Basic issues, embedded web UI | Keep, minimal |
| SAML, SCIM, OIDC, LFS locking, webhook polish | Freeze, label experimental |
| Terraform provider, Helm chart, enterprise compose | Freeze, label experimental |
| Contribution heatmaps, Primer pixel polish, product tour | Freeze |
| SaaS, multi-tenancy, Postgres backend, CI runner | Not now |

Frozen modules must still build and must not break CI. They just stop getting attention.

### F5. Security baseline
• A threat model (`THREAT_MODEL.md`): assets (source, tokens, leases), attackers (malicious agent, stolen token, network attacker, malicious repo content), trust boundaries (HTTP, SSH, MCP, UI).  
• Lease semantics specified and tested: scope, TTL, revocation, no privilege escalation, one audit entry per use.  
• SSH hardening: host key handling, rate limiting, auth timeouts, key type policy.  
• CI tooling: go vet, staticcheck, govulncheck, gosec, the race detector.  
• `SECURITY.md` with a disclosure address.  
• No 'enterprise' language until an external review happens.  

### F6. Architecture honesty
• State the design limits in the docs: SQLite WAL means one writer at a time, and go-git is slower than native git on large repos. Say what scale it is built for.  
• Publish a benchmark against Forgejo (Tier 3). Do not claim machine-speed throughput.  

### F7. License and trademark
• Decide now. Default recommendation: Apache-2.0 for the code (simplify from the current Apache/MIT dual license), plus a `TRADEMARK.md` protecting the name and logo, because managed hosting will be your moat, not the code.  
• Flip condition: if you expect a larger player to host it as a competing service, consider AGPL or a source-available license before traction, not after. The cost is lower adoption among companies that avoid those licenses.  

### F8. Repo hygiene
• Audit `research/` and `agent.md` for internal notes that should not ship.  
• Add `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, issue and PR templates, `SECURITY.md`, `CHANGELOG.md`.  
• CI on every push: Go tests, vet, lint, vulnerability scan, web build, docs build, link check.  
• Versioning: drop the 'enterprise' suffix. Ship v0.5.0-beta as the launch candidate.  

---

## Part 2: Build plan to launch (current development)
Each tier has a done-when test. Do not start the next tier until the current one passes, except where noted as overlapping.

### Tier 0: Truth and hygiene (week 1)
Everything in F1 to F8 that is docs, CI, and cleanup. Done when: `CLAIMS.md` exists with every claim backed or removed, CI is green, all links resolve, and the site URL is unified.

### Tier 1: The agent loop (weeks 1-2). This is the actual product.
Flow: an agent requests a lease, clones, branches, pushes, opens a PR, another agent reviews, the merge queue merges through policy, and the lease is revoked.  
• Trim MCP to a core of about 8 tools: `repo_tree`, `file_read`, `diff_compare`, `symbol_search`, `pull_request_create`, `thread_comment`, `status_check_report`, `pull_request_merge`, plus the lease tools. Make each one excellent, with clear schemas and errors an agent can act on.  
• Build an end-to-end test driven by a real MCP client against a real running existd, with no mocks.  
Done when: the e2e test passes in CI; an expired lease is rejected; a revoked lease is rejected; the read-only endpoint blocks every write tool; every action leaves an audit entry; and removing a policy gate makes a test fail (mutation check).

### Tier 2: Install and first run (week 2)
• One binary with the UI embedded, one command to bootstrap. Docker image on ghcr. Release binaries for Linux, macOS and Windows (amd64 and arm64) with checksums.  
• Done when: a clean Ubuntu container in CI goes from nothing to a first agent-created PR in under 5 minutes following only the README, and the Docker path does the same.

### Tier 3: Proof (week 3)
• Benchmark against Forgejo on the same machine with 1, 5 and 20 agents pushing, cloning and opening PRs in parallel. Report latency and failure rate, including where exist.io breaks.  
• A conformance table for anything you call compatible, with how it was tested.  
• Done when: numbers and method are in the repo and reproducible with one script. Bad numbers are fine. Hidden numbers are not.

### Tier 4: Private beta (weeks 3-4)
• Recruit about 10 outside users: AI agent builders on X, LinkedIn, Discord and Reddit communities. Ask them to run the quickstart and push their own agent through the loop.  
• Track install success rate, time to first PR, and the top 5 complaints. Fix what breaks first.  
• Done when: at least 10 people who are not you completed the install, and the top issues are fixed or documented.

### Tier 5: Launch assets (weeks 3-4, overlaps the beta)
• A 90-second demo, recorded from a script so it can be re-recorded: lease, push, PR, review, a blocked write on the read-only endpoint, merge.  
• README rewritten: tagline, demo GIF, 5-minute quickstart, one comparison table (versus Origin, GitHub, Forgejo) with verified facts only, and an honest limits section.  
• A one-screen landing page: demo, install command, GitHub link.  
• Product Hunt kit: name (max 40 characters), tagline (max 60), description (max 260), gallery images (about 1270x760), maker first comment, and answers to the 5 likely questions.

### Tier 6: Launch week (week of November 16)
• Launch Tuesday to Thursday. Product Hunt days start at 12:01 AM Pacific, which is the afternoon in IST, so you are live for the first hours.  
• Same week: an X thread and a LinkedIn post with the demo, and a Show HN (not in the same hour as the Product Hunt launch). Reply to every comment fast.  
• If the gate below fails, move the date. Moving costs a week. Launching broken costs the launch.

### Launch gate (all must be true)
1. Claims ledger clean.
2. Agent-loop e2e green in CI.
3. Clean-machine install under 5 minutes, tested in CI.
4. Benchmark published.
5. Security baseline done, no enterprise language.
6. 10 outside users completed the install.
7. 90-second demo recorded.

---

## Part 3: Future scope (explicitly after launch)

### 0-3 months after launch
• GitHub import and mirror sync. Origin ships GitHub sync, and it is the adoption path for anyone with existing repos.  
• Stacked-PR-friendly review flow, if users ask for it.  
• A Postgres backend option, only if the benchmark and real users show SQLite is the actual ceiling.  
• Governance features: lease policies and audit export. These are what companies pay for.  

### 3-6 months
• Managed hosting, single-tenant: one isolated instance per customer, provisioned on demand, with backups and updates. It reuses the SQLite design and avoids a multi-tenant rewrite. The code stays open.  
• Bring SAML and SCIM out of the freeze when a paying customer asks, with an external security review.  
• A CI runner or Actions-compatible workflows only if users demand them. That is a large surface.  

### SaaS decision gate
Start only when 10 or more outside self-hosters are active, 3 or more explicitly ask for hosted, and the security review is done. Not before.

---

## Part 4: Risks and kill criteria
• If the 90-second demo is not compelling, fix the loop, not the marketing.  
• If the benchmark shows it cannot handle 5 concurrent agents, document the limits and target small teams only.  
• If the beta cannot find 10 users, the problem is positioning or distribution. Test that before building more.  
• Solo bandwidth: the freeze list is how you protect it. Adding to scope requires removing something.  

---

## Part 5: The execution prompt
Paste this into Claude Code (or your coding agent) at the repo root. It is written to run end to end and stop only for things the agent cannot do.

```
ROLE
You are the lead engineer taking exist.io from an impressive README to a launch-ready product. Repo: github.com/RavaniRoshan/exist.io. Stack as I understand it: Go daemon (existd) and CLI (exist), go-git, SQLite WAL, React 19 and Vite web UI in /web, docs site in /site, deploy/ with a Terraform provider and Helm chart, an MCP server at /mcp. Verify all of this yourself. The README is aspirational, so trust only code and passing tests.

MISSION
Make exist.io ready for a Product Hunt launch on Tuesday, November 17, 2026. Positioning: the neutral, self-hosted Git forge for coding agents (Claude Code, Codex, Cline, Cursor). The launch passes if a stranger can install it in under 5 minutes and watch an agent go from lease to merged pull request, as a 90-second demo.

NON-NEGOTIABLE RULES
1. Truth over marketing. Every claim in the README must map to a passing test or a file. No evidence, no claim.
2. No false positives. A feature counts as done only if a test fails when the feature is deliberately broken. Prove it with a mutation check: temporarily break the code, confirm the test fails, restore it.
3. No new features beyond this brief. If you think something is missing, write it in ROADMAP.md instead of building it.
4. No mocks in production code paths. Test doubles are fine inside tests only.
5. Work autonomously. Ask me only for things you cannot do yourself: recording the final demo on my screen, creating the Product Hunt account or listing, domain and DNS, GitHub repo secrets and settings, signing keys, posting on social accounts, recruiting beta users. Batch those requests at the end of each phase.
6. Commit per phase with conventional commits. Keep main green. Never leave the repo broken between phases.
7. No SaaS or multi-tenant work. Do not expand frozen modules.
8. If something fundamental fails (the agent loop cannot be made reliable, a core claim is false), stop and report instead of papering over it.

FROZEN MODULES (label experimental in docs, keep them building, do not improve them)
SAML, SCIM, OIDC, LFS locking, webhook polish, Terraform provider, Helm chart, enterprise docker-compose, contribution heatmap and Primer pixel polish, interactive product tour.

PHASE 0: AUDIT (docs only, no code changes)
- Build and test everything. Record what passes, fails, or does not run: go build, go vet, go test with the race detector, web build, docs build.
- Create CLAIMS.md: a table of every claim in the README, site and docs, with columns claim, evidence (test name or file path), status (verified, partial, unverified, false), and action. Include at minimum: full wire compatibility with Origin, enterprise-ready, 100 percent open source with zero external proprietary dependencies, pure-Go zero cgo, 35 MCP tools, 119 REST operations, Git LFS v1 with locking, OpenSSH and GPG signature verification, merge queue, AST symbol search, capability leases, SAML, OIDC, SCIM, and the test-count and version badges.
- Find contradictions and dead links: the exist.os versus exist.io URLs, the :8080 versus :3000 web UI port, the missing web build in the quickstart, hardcoded badges.
- Map the security surface: every network listener, every auth path, every place a token or lease is created, checked, or revoked. List what is unreviewed.
- List what in research/ and agent.md is internal and should not ship. Check LICENSE files and dependency licenses.
- Write AUDIT.md with the top 10 findings ranked by launch risk, each with file and line evidence.
Gate: AUDIT.md and CLAIMS.md committed. Show me the top 10 findings, then continue without waiting unless something is fundamentally broken.

PHASE 1: FIXES
- Rewrite the README: new tagline (shortlist: Self-hosted Git forge built for coding agents; Any agent, your server: a self-hosted Git forge; The neutral Git forge for coding agents), a placeholder for the demo GIF, a 5-minute quickstart, a comparison table versus Cursor Origin, GitHub and Forgejo that states only verified facts, and an honest limits section (SQLite single writer, go-git performance, current scale).
- Remove or soften every claim marked false or unverified. Replace hardcoded badges with CI-driven ones. Remove enterprise-ready language. Replace wire compatibility with a conformance table of what is actually matched and tested, or delete it.
- Unify the site URL and fix every link. Add a link checker to CI.
- Embed the built web UI in the existd binary (go:embed) so one binary serves UI, API, MCP and Git on documented ports. Fix the quickstart and endpoint table to match reality.
- Mark frozen modules experimental in README and docs.
- Add CONTRIBUTING.md, CODE_OF_CONDUCT.md, SECURITY.md, issue and PR templates, CHANGELOG.md, and TRADEMARK.md. Set the license to Apache-2.0 only unless I say otherwise, and tell me exactly what you changed.
- Set the version to 0.5.0-beta and remove the enterprise suffix.
- CI on every push and PR: go build, vet, staticcheck, govulncheck, gosec, go test with the race detector, web build, docs build, link check.
Acceptance: CI green; every README claim appears in CLAIMS.md as verified; no dead links; no mention of the old URL.

PHASE 2: THE AGENT LOOP (the actual product)
- Define the core MCP tool set (about 8 tools plus lease tools): repo_tree, file_read, diff_compare, symbol_search, pull_request_create, thread_comment, status_check_report, pull_request_merge, and lease request, inspect and revoke. Give each a precise schema and clear error messages an agent can act on. Move the other tools out of the default surface or mark them experimental.
- Lease semantics, written in docs and enforced in code: scope (repo and ref), TTL, revocation, no privilege escalation, one audit-log entry per use.
- Build an end-to-end test using a real MCP client against a real running existd with a temp data dir (no mocks): lease, clone over HTTP and SSH, branch, push, open PR, post a review comment, report a status check, merge through the merge queue and policy gates, revoke the lease.
- Negative tests: expired lease rejected; revoked lease rejected; out-of-scope ref rejected; every write tool blocked on /mcp/readonly; merge blocked when required checks or approvals are missing; an audit entry exists for each action.
- Run mutation checks on the policy gates and lease checks to prove the tests fail when the logic is removed.
Acceptance: e2e and negative tests pass in CI; mutation checks documented in docs/testing.md.

PHASE 3: INSTALL AND FIRST RUN
- Release config (goreleaser or equivalent): Linux, macOS and Windows, amd64 and arm64, checksums, SBOM. Docker image published to ghcr. An install script, tested.
- A single-command bootstrap that prints the URL, the admin token, and the exact commands for connecting an MCP client.
- A CI job on a clean Ubuntu container that follows only the README quickstart from zero to a first agent-created PR, and fails if it takes more than 5 minutes or if any README command fails. Same check for the Docker path.
Acceptance: both install paths pass in CI from a clean environment.

PHASE 4: SECURITY BASELINE
- Write THREAT_MODEL.md: assets (source, tokens, leases), attackers (malicious agent, stolen token, network attacker, malicious repo content), trust boundaries (HTTP, SSH, MCP, UI).
- Harden SSH: host key persistence and rotation, auth timeouts, rate limiting, key type policy. Harden HTTP: secure cookies, CSRF where cookies are used, request size limits, timeouts, safe defaults (bind to localhost unless configured).
- Fix everything govulncheck and gosec report, or document an explicit justification for each exception.
- Fuzz or property-test the Git protocol parsing, signature parsing, and lease token parsing for panics and bypasses.
Acceptance: scanners clean or justified; threat model committed; fuzz targets run in CI on a short budget.

PHASE 5: PROOF
- Write a reproducible benchmark script comparing exist.io with Forgejo on the same machine: 1, 5 and 20 agents in parallel doing clone, push and PR creation. Report p50 and p95 latency, throughput, and failure rate. Include where exist.io breaks. Commit the script, the raw results and the method to /bench.
- Add a conformance table for anything the README calls compatible, with how each item is tested.
Acceptance: one command reproduces the numbers; the README links to the results and claims no more than they show.

PHASE 6: LAUNCH ASSETS
- Script the demo as a regenerable terminal recording (a scriptable recorder such as VHS or asciinema) covering lease, push, PR, review, a blocked write on the read-only endpoint, and merge. Target 90 seconds.
- Docs quickstart and a short MCP client setup guide for Claude Code, Codex and Cline.
- A one-screen landing page with the demo, the install command and the GitHub link.
- Draft the Product Hunt kit in docs/launch/: name (max 40 chars), tagline (max 60), description (max 260), maker first comment, answers to the 5 likely questions, and X, LinkedIn and Show HN drafts. Verify the current Product Hunt field limits before finalizing.
Acceptance: the demo regenerates from one command; all assets committed.

PHASE 7: READINESS REPORT
Produce LAUNCH_READINESS.md: a pass or fail matrix for the launch gate (claims ledger clean, agent-loop e2e green, clean-machine install under 5 minutes, benchmark published, security baseline done, demo recorded, 10 outside users, which only I can complete), plus the list of things only I can do in priority order, and any open risks. Be honest: anything not proven is marked not proven.

OUTPUT STYLE
At the end of each phase, give me five lines: what changed, what passed, what failed, what you need from me, what is next. No marketing language in any file you write.
```

---

## Part 6: Your own to-do (the agent cannot do these)
1. Decide the license (F7) before Phase 1 runs. Default: Apache-2.0 plus a trademark policy.  
2. Start recruiting beta users in week 2, not week 3. Post the repo and a short clip in agent-builder communities.  
3. Create the Product Hunt maker account and decide on a hunter or self-hunt.  
4. Record or approve the final 90-second demo.  
5. Line up an outside security review of the SSH and auth code if you want to use the word enterprise later.  
