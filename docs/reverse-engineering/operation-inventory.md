# Origin operation inventory

`origin-operation-inventory.csv` is the work queue for REST compatibility. It has one row per unique OpenAPI `operationId`, including exact method/path, token types, scopes, ambient flag, installation scopes, planned task, and baseline registration status.

## Provenance and regeneration

- Source: Cursor Origin API OpenAPI snapshot supplied as `/tmp/opencode/origin-spec/openapi.yaml` during the 2026-10-14 review.
- Source SHA-256: `8e23b604a66d529f595be5949404092264b8513bd2e4713ffb346c499d88b74d`.
- Source identifies itself as OpenAPI 3.1.0, `Cursor Origin API`, version `v1alpha1`; the file is not copied into this repository because redistribution permission has not been established.
- Rebuild with `python3 docs/reverse-engineering/generate-operation-inventory.py /path/to/openapi.yaml`. The generator requires PyYAML. It records the source hash in its stdout; verify it against the pinned hash above before accepting regenerated rows.
- `baselineStatus` is derived by scanning checked-in `internal/origin/*.go` for `s.register("<operationId>"` calls. It deliberately says only `registered-unverified`; it cannot assess route quality or handler behavior.
- The locally inspected CLI binary is `/tmp/opencode/origin-bin/co`, SHA-256 `d653ba55831cb698cdce49167a85ffc5a7d7d839d633c8a62fa17b073be04011`; provider evidence is in `/tmp/opencode/tfcursor/`. These are evidence inputs, not runtime dependencies.

## Reading the status column

`registered-unverified` means an `origin.Server` registration with that operation ID exists in the reviewed source. It does not claim that routing, behavior, scope enforcement, persistence, or compatibility tests are complete. `not-implemented` means no matching registration was found in the reviewed baseline; update status only after the handler and its contract tests land. Re-run the baseline scan before using this status as a release claim.

`plannedTask` is an execution assignment, not proof that the current plan already has sufficient acceptance criteria. Task 23 explicitly covers pull search, rate limits, collaborator permission lookup, and any residual rows that do not fit the existing groups. Tasks 19–22 (MCP, CLI, OpenAPI publication, Terraform) have independent inventories and must not be counted as Origin REST operations.
