# exist.io — Reverse-Engineering Completeness & Gap Analysis

> Produced with the `reverse-engineer-anything` (REA) skill against the shipped
> Origin artifacts, plus external research on the Cursor Origin product.
> This is the source of truth for "what we cloned" vs "what we must build."

## Evidence sources (what we actually have)

Snapshot check — 2026-10-14: the local OpenAPI file has 119 unique `operationId` entries (not 120), SHA-256 `8e23b604a66d529f595be5949404092264b8513bd2e4713ffb346c499d88b74d`. The derived checked-in operation inventory is `docs/reverse-engineering/origin-operation-inventory.csv`; provenance and regeneration instructions are in `operation-inventory.md`. The CLI artifact SHA-256 is `d653ba55831cb698cdce49167a85ffc5a7d7d839d633c8a62fa17b073be04011`. Both source inputs currently live under `/tmp/opencode`; the operation count is evidence about this snapshot, not a guarantee that the live service has not changed.

| Source | Path | What it authoritatively gives us |
|---|---|---|
| Origin OpenAPI 3.1 snapshot | `/tmp/opencode/origin-spec/openapi.yaml` (1.0 MB; SHA-256 above) | 119 `operationId` entries in this snapshot, schemas, scope extensions, webhook event names, error envelope, enum values |
| Origin CLI (Bun binary) | `/tmp/opencode/origin-bin/co` == `origin` (98 MB, ELF x86-64, SHA-256 above) | Embedded TypeScript module paths, strings, descriptors, CLI command tree, env var/header names, error/refusal text, mergeability and auth-policy clues |
| Terraform provider (source) | `/tmp/opencode/tfcursor/` (Go) | Real client usage of the API: repo + grant CRUD shapes, auth flow, pagination |
| Cursor product docs | `cursor.com/docs/origin`, `flaviocopes.com/cursor-origin` | Product architecture, what's public vs proprietary, agent-native positioning |

---

## PART A — What the static artifacts describe

The OpenAPI snapshot gives exact static wire descriptions for its captured version. This does not establish live service behavior, complete client compatibility, or current product coverage. Reconcile every claim through the checked-in operation inventory planned in `docs/superpowers/plans/2026-10-14-exist-io.md` Task 0.

1. **REST surface (119 operations in this snapshot).** The checked-in operation inventory maps each operation to a plan task and marks the 28 baseline registrations as unverified, not complete. Path shapes include `:action` verbs (`:syncMirror`,
   `:batchGet`, `:entries:replace`, `:batchRedeliver`, `:grep`), all under `/v1/origin/`.
2. **Auth model.** 3 token types (`app` JWT, `oit_` install ≤15 min, user cred),
   `key_`/`crsr_` exchange via `POST /auth/exchange_user_api_key`, `Authorization: Bearer`,
   ambient `CURSOR_API_KEY`/`CURSOR_AUTH_TOKEN`, **token attenuation** (`isAttenuatedOriginSession`).
3. **Scope catalog (34).** 33 in `x-origin-scopes.scopes` + `namespace:user_tokens:write`
   (installation-only). Enforcement semantics: `tokenTypes` gate, `ambient` scopes come
   with the credential, `installationScopes` extras for install tokens. ✅ *implemented in `pkg/scopes` + `internal/auth`.*
4. **Every request/response schema & enum.** `PERMISSION_READ`, `MergeConflictState`
   (CLEAN/PENDING/CONFLICT), check-run status/conclusion, ruleset enforcement/kind/rules,
   PR states, mergeability verdicts (`mergeable`/`stale`/`blocked`), blocker kinds.
5. **`Status` error envelope.** `{code (google.rpc.Code int32), message, details}`;
   `PERMISSION_DENIED=7`, `UNAUTHENTICATED=16`. ✅ *implemented in `internal/auth`.*
6. **ID prefixes.** `app_ inst_ ns_ repo_ pr_ cr_ crg_ cra_ nsca_ nsip_ whd_ wha_ evt_ grp_ user_ stk_ oit_`. ✅ *implemented in `pkg/idgen`.*
7. **Webhooks (35 events).** Ed25519 `v1ed` signature over `SHA256({webhook-id}.{webhook-timestamp}.{body})`, verified against `/v1/origin/keys`.
8. **MCP surface (33 tools).** Stateless streamable HTTP, `POST` JSON-RPC, `/mcp` + `/mcp/readonly`, `x-mcp-readonly: true` header, `_meta["cursor/requiresConfirmation"]` for destructive tools, structured error `{category, message, requestId, retryAfterSeconds?, rateLimit?}` with categories `not_found|forbidden|quota_exceeded|conflict|validation_error|upstream_failure`. **Note: MCP host is `api.origin.cursor.com` (different from the REST host).**
9. **CLI command tree.** Far larger than the plan's Task 20: adds `stack`, `label`,
   `identity`, `namespace access`, `repo access`, `gh`, `agent`, `config`,
   `credential-helper`, `completion`, `ruleset`, `ssh-key`, `watch`, `push`, `update`.
10. **Internal client modules** (names + responsibilities, from embedded manifest):
    `origin-auth-policy-catalog`, `origin-blocker-kinds`, `origin-gate-checks`,
    `origin-rate-limit-backoff`, `origin-refusal-messages`, `merge-expected-heads`,
    `repo-sot`, `origin-sync-client/engine/*` (diff/fold/tuning/v2 multiplexing),
    `pulls-search/*` (ast/lexer/parse/values — a real query language).

---

## PART B — What CANNOT be reverse-engineered (the gaps)

These depend on **server-side state, Cursor's backend, or live runtime** that the
shipped spec + CLI + provider do **not** contain. We must **design and build our own**.

### B1. Server-side internals & data (no client artifact exposes these)
- **Storage schema.** The plan's 12 SQLite tables are *our design*, not Origin's.
  Origin's actual DB is proprietary. Column types, indexes, constraints = ours.
- **Merge evaluation engine.** The spec gives the *result shape* (`mergeable`, `blocker_kinds`,
  `mergeConflictState`) but NOT the algorithm: how conflicts are detected, how status-check
  gating aggregates, how "up to date" is computed, how bypass actors are applied. **We implement this.**
- **Git mirror sync engine.** `origin-sync-client/engine/*` is the *client-side* fold/diff
  logic. The **server-side** GitHub↔Origin bidirectional sync (conflict resolution, retry,
  cutover semantics, PR two-way sync) is not in any artifact. **We implement a compatible client
  contract and our own sync behavior.**
- **Rate limiting & backoff.** `origin-rate-limit-backoff` gives the *client* strategy; the
  server's actual limits/quotas are unknown. **We choose our own limits** (must be sane, not identical).
- **Webhook delivery/retry.** We know the signature and event list; the retry schedule,
  backoff, dead-lettering, and delivery guarantees are ours to define.
- **Ruleset evaluation.** Rule *shapes* are known; *how* Origin evaluates them against a
  push/merge (ordering, aggregation, bypass precedence) is ours.

### B2. Cursor-proprietary, agent-native features (structurally out of reach)
These are the "key parts of Origin" that are **not a git forge** and have **no public API**:
- **Cloud/background coding agents.** The autonomous agents that open PRs are Cursor's
  models + orchestration. Not cloneable; not in the API. Our forge is the *host*, not the agent.
- **Cursor Automations** (event→agent triggers) and **Bugbot** (AI review). Proprietary.
- **PR Routing & Approval** with **attenuated sessions** and write-approval tokens
  (`ORIGIN_WRITE_APPROVAL_*`, agent approval headers). The *enforcement points* are visible;
  the *policy brain* is Cursor's. We implement the mechanism; the "intelligence" is absent by design.
- **Cursor Portal / Codebase web UI**, **Grok Bot**, agent VMs (`CURSOR_HOSTED_AGENT_VM`).
- **Cursor account/team/entitlement system.** Namespaces = Cursor teams; privacy mode;
  plan gating (Pro/Teams/Enterprise). **We replace with our own identity/tenancy.**

### B3. Hosted-service concerns (we must operate, not clone)
- **api.origin.cursor.com / origin.cursor.com** infrastructure, DNS, TLS, CDN, the
  `/v1/origin/keys` JWT verification key set (origin = Cursor's signing keys; ours = our keys).
- **Git-over-SSH/HTTP transport auth.** Origin uses SSH CAs + inbound IP allowlists (in the
  API), but the actual `git` daemon, SSH cert signing, and smart-HTTP are ours to run.

---

## PART C — What we must BUILD OURSELVES to make a complete, open-source forge

The reverse-engineered part is the **static wire contract** for the captured client-facing artifact; unchanged-client operation is a goal that still requires contract testing.
The **open-source product** needs these self-built pillars, which no artifact provides:

1. **Metadata store + migrations** (Task 5) — our schema, our choice (SQLite via modernc, pure Go).
2. **Git object store & server transport** (Task 8) — go-git for storage + a `git` smart-HTTP/SSH
   server so `git clone/push/pull` works. The API describes git *data*; we run the *git server*.
3. **Mergeability + ruleset evaluation engines** (Tasks 11, 14) — our algorithms behind the known result shapes.
4. **Mirror sync** (Task 16) — our GitHub↔Origin sync implementing the known client contract.
5. **Webhook dispatch + delivery + retry** (Task 18) — our delivery subsystem, exact signature.
6. **MCP server** (Task 19) — our implementation of the 33-tool surface over our handlers.
7. **CLI** (Task 20) — our origin-compatible client.
8. **Identity, tenancy, namespaces, grants, principal resolution** (Tasks 6, 7) — our authz,
   replacing Cursor teams with our own user/group model.
9. **Admin bridges** — `GET /teams/members` (user_email→user_) and
   `GET /organizations/groups` (group_name→grp_) are Cursor admin APIs. We implement our own
   equivalents so the Terraform provider's principal resolution works.
10. **The "agent" seam** — Origin's whole point is agents acting on repos. We cannot ship Cursor's
    agents, but we MUST expose the surfaces they use (PR routing, attenuated sessions, approvals,
    MCP) so *third-party* agents (and Cline, etc.) can drive the forge. **This is the strategic gap:
    we are building an agent-native forge host, not an agent.**

---

## Bottom line

- **Static contract coverage: substantial, not yet proven complete.** The snapshot describes 119 operations and client artifacts reveal additional command/tool behavior. Compatibility with the current live service and unchanged-client operation remains unverified until contract tests pass against exist.io.
- **Server behavior & product brain: NOT reverse-engineerable and intentionally so.** Merge logic,
  sync, rate limits, agents, automations, review AI, Cursor tenancy = ours to design. We match the
  *contract*, not the *implementation*.
- **Net:** exist.io targets an open-source, self-hosted forge with a measured Origin-compatible surface. Server behavior absent from the artifacts remains our design, and proprietary Cursor agent functionality is outside this forge's scope.

See `docs/reverse-engineering/scopes-34.txt` and `cli-module-manifest.txt` for the raw evidence.
