#!/usr/bin/env bash
# Fail-closed link checker for published markdown.
set -euo pipefail
exec python3 "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check_links.py" "$@"
