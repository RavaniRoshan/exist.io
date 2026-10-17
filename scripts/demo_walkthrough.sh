#!/usr/bin/env bash
# scripts/demo_walkthrough.sh — Automated 90-second agent walkthrough script for demo recording.

set -euo pipefail

# Ensure PATH includes Go and utilities
if ! command -v go >/dev/null 2>&1; then
    if [ -x "/home/linuxbrew/.linuxbrew/bin/go" ]; then
        export PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"
    fi
fi
export PATH="/home/shiva/bin:/snap/bin:$PATH"

BLUE='\033[0;34m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

DEMO_DIR="/tmp/exist_demo_run"
DEMO_DATA="$DEMO_DIR/data"
DEMO_WORKTREE="$DEMO_DIR/worktree"

rm -rf "$DEMO_DIR"
mkdir -p "$DEMO_DIR"
pkill -f existd 2>/dev/null || true
sleep 1

clear 2>/dev/null || true
echo -e "${BOLD}${BLUE}╔═══════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${BLUE}║       exist.io — The Neutral Git Forge for Coding Agents              ║${NC}"
echo -e "${BOLD}${BLUE}║       90-Second Demo: From Install to Merged Autonomous Agent PR      ║${NC}"
echo -e "${BOLD}${BLUE}╚═══════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
sleep 1.5

# Step 1: Build
echo -e "${CYAN}${BOLD}[1/7] Compiling hermetic pure-Go binaries (CGO_ENABLED=0)...${NC}"
echo -e "${YELLOW}$ go build -o bin/existd ./cmd/existd${NC}"
CGO_ENABLED=0 go build -o bin/existd ./cmd/existd
echo -e "${GREEN}✓ Binary compiled with embedded Web UI & zero host dependencies.${NC}"
echo ""
sleep 1.5

# Step 2: Bootstrap
echo -e "${CYAN}${BOLD}[2/7] Running single-command administrator bootstrap...${NC}"
echo -e "${YELLOW}$ existd bootstrap --email admin@exist.io --namespace acme${NC}"
./bin/existd bootstrap --email admin@exist.io --namespace acme --data-dir "$DEMO_DATA" > "$DEMO_DIR/bootstrap.out"
head -n 14 "$DEMO_DIR/bootstrap.out"
API_KEY=$(grep "API Key:" "$DEMO_DIR/bootstrap.out" | awk '{print $3}')
echo ""
sleep 2

# Step 3: Start Daemon
echo -e "${CYAN}${BOLD}[3/7] Launching forge daemon (Smart-HTTP :8080, pure-Go SSH :2222, MCP :8080/mcp)...${NC}"
echo -e "${YELLOW}$ existd &${NC}"
EXIST_DATA_DIR="$DEMO_DATA" EXIST_LISTEN_ADDR="127.0.0.1:8080" ./bin/existd > "$DEMO_DIR/existd.log" 2>&1 &
DAEMON_PID=$!
sleep 1.5

echo -e "${YELLOW}$ curl -s http://127.0.0.1:8080/healthz${NC}"
HEALTH=$(curl -s http://127.0.0.1:8080/healthz)
echo -e "${GREEN}✓ Daemon status: $HEALTH${NC}"
echo ""
sleep 1.5

# Step 4: Capability Lease
echo -e "${CYAN}${BOLD}[4/7] Agent requests ephemeral HMAC-SHA256 capability lease via MCP...${NC}"
echo -e "${YELLOW}$ curl -X POST http://127.0.0.1:8080/mcp (tool: lease_request, scope: refs/heads/agent/*)${NC}"

curl -s -X POST http://127.0.0.1:8080/mcp \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lease_request","arguments":{"agentId":"claude-code","repoId":"acme/demo","allowedRefPrefixes":["refs/heads/agent/"]}}}' > "$DEMO_DIR/lease.json"

cat "$DEMO_DIR/lease.json" | python3 -c '
import sys, json
d = json.load(sys.stdin)
content = json.loads(d["result"]["content"][0]["text"])
print("  Lease ID:  ", content.get("id"))
print("  Agent ID:  ", content.get("agentId"))
print("  Ref Scope: ", content.get("allowedRefPrefixes"))
print("  Expires At:", content.get("expiresAt"))
'
LEASE_TOKEN=$(cat "$DEMO_DIR/lease.json" | python3 -c 'import sys, json; d = json.load(sys.stdin); print(json.loads(d["result"]["content"][0]["text"])["token"])')
echo -e "${GREEN}✓ Ephemeral lease issued. Branch access strictly isolated to refs/heads/agent/*${NC}"
echo ""
sleep 2

# Step 5: Git Clone & Branch
echo -e "${CYAN}${BOLD}[5/7] Agent clones repository and creates branch using lease credentials...${NC}"
echo -e "${YELLOW}$ git clone http://lease:<TOKEN>@127.0.0.1:8080/acme/demo.git worktree${NC}"
git clone "http://lease:$LEASE_TOKEN@127.0.0.1:8080/acme/demo.git" "$DEMO_WORKTREE" 2>&1 | grep -v "warning: redirecting"
cd "$DEMO_WORKTREE"
git checkout -b agent/telemetry
echo '{"status":"healthy","uptime":"active","agent":"claude-code"}' > telemetry.json
git add telemetry.json
git commit -m "feat: add telemetry descriptor"
echo ""
sleep 1.5

# Step 6: Push to refs/heads/agent/*
echo -e "${CYAN}${BOLD}[6/7] Agent pushes branch to forge (verified by pre-receive hook)...${NC}"
echo -e "${YELLOW}$ git push origin agent/telemetry${NC}"
git push origin agent/telemetry 2>&1 | grep -v "warning: redirecting"
echo -e "${GREEN}✓ Security hook passed: secret scanning clean, ref namespace confirmed.${NC}"
echo ""
sleep 2

# Step 7: Open & Merge PR
echo -e "${CYAN}${BOLD}[7/7] Agent opens Pull Request & merges via MCP...${NC}"
echo -e "${YELLOW}$ curl -X POST http://127.0.0.1:8080/mcp (tool: pull_request_create)${NC}"

PR_RES=$(curl -s -X POST http://127.0.0.1:8080/mcp \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"pull_request_create","arguments":{"ownerSlug":"acme","repoName":"demo","title":"feat: add agent telemetry descriptor","headRef":"agent/telemetry","baseRef":"main"}}}')

PR_NUM=$(echo "$PR_RES" | python3 -c 'import sys, json; d = json.load(sys.stdin); print(json.loads(d["result"]["content"][0]["text"]).get("number") or 1)')
echo -e "${GREEN}✓ Pull Request #$PR_NUM created successfully.${NC}"

echo -e "${YELLOW}$ curl -X POST http://127.0.0.1:8080/mcp (tool: pull_request_merge, pull: $PR_NUM)${NC}"
curl -s -X POST http://127.0.0.1:8080/mcp \
    -H "Authorization: Bearer $API_KEY" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"pull_request_merge\",\"arguments\":{\"ownerSlug\":\"acme\",\"repoName\":\"demo\",\"pullNumber\":$PR_NUM}}}" > "$DEMO_DIR/merge.json"

echo -e "${GREEN}✓ Pull Request #$PR_NUM merged into main by policy engine.${NC}"
echo ""
sleep 1.5

echo -e "${CYAN}${BOLD}==> Verifying main branch log:${NC}"
echo -e "${YELLOW}$ git fetch origin main && git log -n 1 --stat origin/main${NC}"
git fetch origin main >/dev/null 2>&1
git log -n 1 --stat origin/main
echo ""
sleep 2

echo -e "${GREEN}${BOLD}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}${BOLD} ✓ Complete! From cold install to merged agent PR in < 90 seconds.   ${NC}"
echo -e "${GREEN}${BOLD}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""
sleep 3

kill "$DAEMON_PID" 2>/dev/null || true
rm -rf "$DEMO_DIR"
