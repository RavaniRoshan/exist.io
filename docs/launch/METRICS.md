# Activation and Launch Metrics (`METRICS.md`)

## Activation event

**Activation** is the first successful merge of a pull request originating from a capability lease, by an agent that is not the bootstrap administrator.

This is the single value-bearing event for exist.io. Every funnel definition below derives from it.

## Funnel

| Stage | Event | Where it is observed |
| :--- | :--- | :--- |
| 1. Install | The release binary is installed | Installer / CI matrix |
| 2. Bootstrap | `existd bootstrap` creates the first administrator, namespace, and starter repository | Bootstrap output |
| 3. Daemon ready | `GET /healthz` returns `{"status":"ok","version":"…"}` | Health probe |
| 4. Lease issued | `lease_request` succeeds over MCP | MCP / audit log |
| 5. Agent push | A push to `refs/heads/agent/*` is accepted | Git transport / audit log |
| 6. PR opened | `pull_request_create` succeeds | MCP / audit log |
| 7. **Activated** | The PR is merged by an actor other than the bootstrap administrator | `activation.json` + telemetry |

## What is measured

- **Install→activation rate:** activated instances divided by installs that reached stage 3.
- **Time to activation:** `activated_at - bootstrapped_at`.
- **Drop-off stage:** the last completed stage for instances that did not activate.

## Privacy boundary

Telemetry is **off by default** and consent-based. It is enabled only by:

1. `EXIST_TELEMETRY_OPT_IN=1`, or
2. answering `y` to the bootstrap prompt (bare Enter means **no**).

The transmitted payload contains exactly:

```json
{
  "instance_id": "<random 32-hex identifier>",
  "event": "activation.first_merge",
  "version": "v0.6.0-beta",
  "os": "linux",
  "date": "2026-11-01"
}
```

It never contains repository names, commit content, file paths, user identifiers, or hostnames. There is no default remote collector: the operator must set `EXIST_TELEMETRY_ENDPOINT`. Without it, the event is recorded locally only.

## Referrer convention

Outbound links use `?ref=` values: `readme`, `docs-home`, `docs-getting-started`, `installer`, `product-hunt`. Canonical URLs are unchanged when `ref` is absent.

## Reproduce

```bash
go test ./internal/telemetry -v
./scripts/test_telemetry.sh
```
