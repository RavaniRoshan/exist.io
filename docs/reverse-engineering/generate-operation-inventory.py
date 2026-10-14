#!/usr/bin/env python3
"""Generate the checked-in Origin operation matrix from a caller-supplied OpenAPI YAML."""

import argparse
import csv
import hashlib
import re
from pathlib import Path

import yaml


def task_for(operation_id, path):
    if path == "/v1/origin/rate_limit" or path == "/v1/origin/search/pulls" or "/collaborators/" in path:
        return "Task 23 — search, rate limits, and residual API"
    if "/grants" in path:
        return "Task 7 — grants"
    if "/git/" in path or "/contents" in path or "/commits" in path or "/branches" in path or "/compare/" in path or "/tarball" in path or ":grep" in path:
        return "Task 9/10 — contents and refs"
    if "/namespaces" in path and ("/apps" in path or "/installations" in path):
        return "Task 17 — apps and installations"
    if "/namespaces" in path and ("ssh-certificate" in path or "inbound-ip-allowlist" in path):
        return "Task 15 — SSH CAs and IP allowlist"
    if "/repos" in path and ("/pulls" in path or "/search/pulls" in path):
        if "/comments" in path or "/reviews" in path or "/labels" in path or "/reactions" in path or "/threads" in path or "reviewers" in path:
            return "Task 12 — PR discussion, reviews, labels"
        return "Task 11 — pull requests"
    if "/check-" in path or "/check-runs" in path or "/check-suites" in path:
        return "Task 13 — checks"
    if "/rulesets" in path:
        return "Task 14 — rulesets"
    if "/mirror" in path or ":syncMirror" in path:
        return "Task 16 — mirrors"
    if "/app/webhook/" in path:
        return "Task 18 — webhooks"
    if "/app" in path or "/apps" in path or "/installation" in path:
        return "Task 17 — apps and installations"
    if "/repos" in path:
        return "Task 6 — namespace/repository CRUD"
    if "/namespaces" in path:
        return "Task 6 — namespace/repository CRUD"
    return "Task 23 — search and residual API"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("openapi_yaml", type=Path, help="Caller-supplied Origin OpenAPI snapshot")
    parser.add_argument("--output", type=Path, default=Path(__file__).with_name("origin-operation-inventory.csv"))
    args = parser.parse_args()

    raw = args.openapi_yaml.read_bytes()
    document = yaml.safe_load(raw)
    repo_root = Path(__file__).resolve().parents[2]
    registrations = set()
    for source in (repo_root / "internal/origin").glob("*.go"):
        registrations.update(re.findall(r's\.register\("([^"]+)"', source.read_text(encoding="utf-8")))
    rows = []
    for path, path_item in document.get("paths", {}).items():
        for method, operation in path_item.items():
            if not isinstance(operation, dict) or "operationId" not in operation:
                continue
            extension = operation.get("x-origin-scopes") or {}
            operation_id = operation["operationId"]
            rows.append({
                "operationId": operation_id,
                "method": method.upper(),
                "path": path,
                "tokenTypes": "|".join(extension.get("tokenTypes", [])),
                "scopes": "|".join(extension.get("scopes", [])),
                "ambient": str(bool(extension.get("ambient", False))).lower(),
                "installationScopes": "|".join(extension.get("installationScopes", [])),
                "plannedTask": task_for(operation_id, path),
                "baselineStatus": "registered-unverified" if operation_id in registrations else "not-implemented",
            })

    rows.sort(key=lambda row: (row["operationId"], row["method"], row["path"]))
    if len({row["operationId"] for row in rows}) != len(rows):
        raise SystemExit("OpenAPI contains duplicate operationId values; resolve before inventory generation")
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("w", newline="", encoding="utf-8") as output:
        writer = csv.DictWriter(output, fieldnames=list(rows[0]))
        writer.writeheader()
        writer.writerows(rows)
    print(f"wrote {len(rows)} operations; source_sha256={hashlib.sha256(raw).hexdigest()}; output={args.output}")


if __name__ == "__main__":
    main()
