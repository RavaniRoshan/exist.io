#!/usr/bin/env bash
# Seeds a running instance with an open agent pull request for gallery capture.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORK="${1:-/tmp/exist-capture}"
PORT="${2:-18080}"

rm -rf "$WORK"
mkdir -p "$WORK/data" "$WORK/bin"

VERSION="$(tr -d '[:space:]' < "$ROOT_DIR/VERSION")"
CGO_ENABLED=0 go build -ldflags="-X main.version=$VERSION" -o "$WORK/bin/existd" "$ROOT_DIR/cmd/existd"
"$WORK/bin/existd" bootstrap --email admin@corp.exist.io --namespace acme --data-dir "$WORK/data" > "$WORK/bootstrap.out" 2>&1
API_KEY="$(grep 'API Key:' "$WORK/bootstrap.out" | awk '{print $3}')"
if [ -z "$API_KEY" ]; then
    echo "failed to parse the bootstrap API key" >&2
    cat "$WORK/bootstrap.out" >&2
    exit 1
fi
printf '%s' "$API_KEY" > "$WORK/key"
printf '%s' "$PORT" > "$WORK/port"

EXIST_DATA_DIR="$WORK/data" \
EXIST_LISTEN_ADDR="127.0.0.1:$PORT" \
EXIST_PUBLIC_URL="http://127.0.0.1:$PORT" \
EXIST_GIT_URL="http://127.0.0.1:$PORT" \
EXIST_SSH_ADDR="none" \
"$WORK/bin/existd" > "$WORK/daemon.log" 2>&1 &
echo $! > "$WORK/daemon.pid"

for _ in $(seq 1 60); do
    if curl -sf "http://127.0.0.1:$PORT/healthz" >/dev/null; then break; fi
    sleep 0.25
done

LEASE_RESPONSE="$(curl -sS -X POST "http://127.0.0.1:$PORT/mcp" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lease_request","arguments":{"agentId":"forge-capture","repoId":"*","allowedRefPrefixes":["refs/heads/agent/"],"ttlSeconds":3600}}}')"
LEASE_TOKEN="$(printf '%s' "$LEASE_RESPONSE" | python3 -c 'import json,sys; d=json.load(sys.stdin); print(json.loads(d["result"]["content"][0]["text"])["token"])')"

git clone "http://lease:${LEASE_TOKEN}@127.0.0.1:${PORT}/acme/demo.git" "$WORK/worktree" >/dev/null 2>&1
cd "$WORK/worktree"
git config user.name "Forge Capture"
git config user.email "capture@exist.io"
git checkout -b agent/forge-demo >/dev/null 2>&1
printf '\n## Captured with the Forge UI\n\nThis change was pushed by an agent under a capability lease.\n' >> README.md
git commit -am "feat: add Forge capture note" >/dev/null 2>&1
git push origin agent/forge-demo >/dev/null 2>&1

curl -sS -X POST "http://127.0.0.1:$PORT/mcp" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"pull_request_create","arguments":{"ownerSlug":"acme","repoName":"demo","title":"Add Forge capture note","headRef":"agent/forge-demo","baseRef":"main","body":"Captured from a live instance."}}}' > "$WORK/pr.json"

echo "seeded instance on http://127.0.0.1:${PORT} (data: $WORK/data)"
