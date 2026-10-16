# exist.io 90-Second Demo Walkthrough
# Output configuration
Output assets/demo.gif
Output assets/demo.mp4

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 720
Set Padding 20
Set Theme "Catppuccin Macchiato"

# Title & intro
Type "# 1. Install & Bootstrap exist.io in under 5 seconds"
Enter
Sleep 1s

Type "go build -o bin/existd ./cmd/existd"
Enter
Sleep 2s

Type "./bin/existd bootstrap --email admin@exist.io --namespace acme"
Enter
Sleep 2s

# Show generated agent snippets
Type "# Notice: Administrator PAT created, demo repo initialized, and agent MCP configs ready!"
Enter
Sleep 2s

# Start daemon in background
Type "./bin/existd > /tmp/existd.log 2>&1 &"
Enter
Sleep 1.5s

Type "curl -s http://localhost:8080/healthz"
Enter
Sleep 1s

# 2. Agent Workflow: Request Capability Lease
Type "# 2. Autonomous Agent requests an ephemeral capability lease"
Enter
Sleep 1s

Type "curl -s -X POST http://localhost:8080/mcp -H 'Authorization: Bearer <API_KEY>' -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"lease_request\",\"arguments\":{\"agentId\":\"claude-code\",\"repoId\":\"acme/demo\",\"allowedRefPrefixes\":[\"refs/heads/agent/\"]}}}' | jq .result.content[0].text"
Enter
Sleep 2.5s

# 3. Agent clones under lease credentials
Type "# 3. Agent clones repository using capability lease token"
Enter
Sleep 1s

Type "git clone http://lease:<TOKEN>@localhost:8080/acme/demo.git worktree"
Enter
Sleep 2s

Type "cd worktree && git checkout -b agent/healthcheck"
Enter
Sleep 1s

Type "echo '{\"status\":\"healthy\",\"uptime\":\"active\"}' > health.json"
Enter
Sleep 1s

Type "git add health.json && git commit -m 'feat: add healthcheck descriptor'"
Enter
Sleep 1.5s

# 4. Agent pushes to refs/heads/agent/* (verified by pre-receive hook)
Type "# 4. Agent pushes branch; pre-receive hook verifies ref isolation & scans secrets"
Enter
Sleep 1s

Type "git push origin agent/healthcheck"
Enter
Sleep 2s

# 5. Open Pull Request via MCP
Type "# 5. Agent opens a Pull Request programmatically via MCP"
Enter
Sleep 1s

Type "curl -s -X POST http://localhost:8080/mcp -H 'Authorization: Bearer <API_KEY>' -d '{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"pull_request_create\",\"arguments\":{\"ownerSlug\":\"acme\",\"repoName\":\"demo\",\"title\":\"Autonomous PR: Add healthcheck\",\"headRef\":\"agent/healthcheck\",\"baseRef\":\"main\"}}}' | jq ."
Enter
Sleep 3s

# 6. Merge Pull Request & verify
Type "# 6. Pull Request is evaluated against policy gates and merged"
Enter
Sleep 1s

Type "curl -s -X POST http://localhost:8080/mcp -H 'Authorization: Bearer <API_KEY>' -d '{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"pull_request_merge\",\"arguments\":{\"ownerSlug\":\"acme\",\"repoName\":\"demo\",\"pullNumber\":1}}}' | jq ."
Enter
Sleep 2s

Type "git fetch origin main && git log -n 1 --stat origin/main"
Enter
Sleep 3s

Type "# Complete: From install to merged agent PR in under 90 seconds!"
Enter
Sleep 2s
