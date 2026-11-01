#!/usr/bin/env bash
# Asserts that the installer, README, and docs print the same first-run sequence.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BOOTSTRAP='existd bootstrap --email admin@corp.exist.io --namespace acme --data-dir ./data'

FILES=(
    "scripts/install.sh"
    "README.md"
    "site/docs/index.mdx"
    "site/docs/getting-started/index.mdx"
)

status=0
for file in "${FILES[@]}"; do
    if ! grep -qF "$BOOTSTRAP" "$ROOT_DIR/$file"; then
        echo "missing canonical bootstrap command: $file" >&2
        status=1
    fi
done

if grep -rqF 'exist init' "$ROOT_DIR/scripts/install.sh" "$ROOT_DIR/README.md" "$ROOT_DIR/site/docs"; then
    echo "found the non-existent 'exist init' command in a first-run surface" >&2
    status=1
fi

if [ "$status" -eq 0 ]; then
    echo "quickstart surfaces agree"
fi
exit "$status"
