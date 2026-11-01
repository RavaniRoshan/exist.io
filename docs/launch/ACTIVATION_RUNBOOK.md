# Activation Runbook (`ACTIVATION_RUNBOOK.md`)

## What "activated" means

A first merge of an `agent/*` pull request by someone other than the bootstrap administrator. One record per instance, written once to `${EXIST_DATA_DIR}/activation.json`.

## Read the funnel locally

```bash
curl -s http://localhost:8080/api/v1/admin/activation | jq
```

```json
{
  "activated": true,
  "activated_at": "2026-11-01T17:20:00Z",
  "pull_request_id": "pr_...",
  "actor_user_id": "usr_..."
}
```

Local files:

- `${EXIST_DATA_DIR}/bootstrap_admin` — the administrator excluded from activation.
- `${EXIST_DATA_DIR}/activation.json` — the once-only activation record.
- `${EXIST_DATA_DIR}/telemetry_opt_in` — `1` or `0`, written by the bootstrap prompt.

## Consent

Telemetry is off unless `EXIST_TELEMETRY_OPT_IN=1` is set or the bootstrap prompt is answered `y`. Set `EXIST_TELEMETRY_ENDPOINT` to a collector you control; without it the event is only logged locally.

## When a stage is zero

| Stage at zero | Check |
| :--- | :--- |
| Daemon ready | `GET /healthz` returns 200 with a version |
| Lease issued | the MCP token can call `lease_request` |
| Agent push | the remote is a lease URL and the ref starts with `agent/` |
| PR opened | `pull_request_create` returned a number |
| Activated | the merge actor is not the bootstrap administrator and the head ref is `agent/*` |

## Privacy boundary

The payload contains only `instance_id`, `event`, `version`, `os`, and `date`. It never includes repository names, paths, commit content, user identifiers, or hostnames.
