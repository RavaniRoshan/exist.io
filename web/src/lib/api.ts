// exist.io API Client & State Store

export interface RepoInfo {
  ownerSlug: string
  name: string
  description: string
  defaultBranch: string
  isPrivate: boolean
  starsCount: number
  forksCount: number
  updatedAt: string
  cloneUrlHttp: string
  cloneUrlSsh: string
}

export interface FileItem {
  name: string
  path: string
  type: 'tree' | 'blob'
  size?: number
  lastCommitMessage: string
  lastCommitDate: string
}

export interface CommitItem {
  sha: string
  shortSha: string
  message: string
  author: {
    name: string
    email: string
    isAgent?: boolean
  }
  timestamp: string
  verified: boolean
}

export interface CheckRun {
  id: string
  name: string
  status: 'completed' | 'in_progress' | 'queued'
  conclusion: 'success' | 'failure' | 'neutral'
  duration: string
}

export interface ReviewComment {
  id: string
  author: string
  isAgent?: boolean
  createdAt: string
  body: string
  path?: string
  line?: number
  resolved?: boolean
}

export interface PullRequest {
  number: number
  title: string
  body: string
  state: 'open' | 'merged' | 'closed'
  author: string
  isAgent?: boolean
  headRef: string
  baseRef: string
  createdAt: string
  labels: { name: string; color: string }[]
  diffStats: {
    additions: number
    deletions: number
    filesChanged: number
  }
  checks: CheckRun[]
  comments: ReviewComment[]
  mergeable: boolean
  diffHunks: {
    oldPath: string
    newPath: string
    status: 'modified' | 'added' | 'deleted'
    additions: number
    deletions: number
    hunk: string
  }[]
}

export interface AgentActivityItem {
  id: string
  agentName: string
  agentType: 'claude-code' | 'cursor' | 'codex' | 'exist-mcp'
  toolName: string
  status: 'success' | 'executing' | 'error'
  latencyMs: number
  timestamp: string
  requestPayload: any
  responseSummary: string
}

// Initial Mock Seed for exist.io Core Forge
export const defaultRepo: RepoInfo = {
  ownerSlug: 'exist',
  name: 'exist.io',
  description: 'Self-hosted, autonomous-agent-ready Git forge with 100% smart-HTTP transport and native MCP server.',
  defaultBranch: 'main',
  isPrivate: false,
  starsCount: 842,
  forksCount: 67,
  updatedAt: 'Just now',
  cloneUrlHttp: 'http://localhost:8080/exist/exist.io.git',
  cloneUrlSsh: 'git@localhost:exist/exist.io.git',
}

export const sampleFiles: FileItem[] = [
  { name: 'cmd', path: 'cmd', type: 'tree', lastCommitMessage: 'feat: add existd daemon and CLI entrypoints', lastCommitDate: '2 hours ago' },
  { name: 'internal', path: 'internal', type: 'tree', lastCommitMessage: 'feat(mcp): implement all 34 MCP forge tools and wire handlers', lastCommitDate: '1 hour ago' },
  { name: 'pkg', path: 'pkg', type: 'tree', lastCommitMessage: 'refactor: expose public git types and auth models', lastCommitDate: '3 hours ago' },
  { name: 'web', path: 'web', type: 'tree', lastCommitMessage: 'feat(ui): initialize Space UI living dashboard with motion', lastCommitDate: 'Just now' },
  { name: '.gitignore', path: '.gitignore', type: 'blob', size: 503, lastCommitMessage: 'chore: ignore private artifacts and SQLite WAL databases', lastCommitDate: '4 hours ago' },
  { name: 'go.mod', path: 'go.mod', type: 'blob', size: 1144, lastCommitMessage: 'build: bump go-git v6 and modernc.org/sqlite', lastCommitDate: '5 hours ago' },
  { name: 'go.sum', path: 'go.sum', type: 'blob', size: 8909, lastCommitMessage: 'build: checksum updates for pure-Go toolchain', lastCommitDate: '5 hours ago' },
  { name: 'README.md', path: 'README.md', type: 'blob', size: 10502, lastCommitMessage: 'docs: beautify README with animated vector logo and matrix', lastCommitDate: '2 hours ago' },
  { name: 'ROADMAP.md', path: 'ROADMAP.md', type: 'blob', size: 7377, lastCommitMessage: 'docs: document roadmap phases 1 through 5', lastCommitDate: '2 hours ago' },
]

export const sampleReadme = `# exist.io

<p align="center">
  <strong>The Autonomous-Agent-Ready Git Forge</strong>
</p>

exist.io is a modern, pure-Go Git forge designed from the ground up for **human developers** and **autonomous AI coding agents**.

- ⚡ **100% Smart-HTTP Git Engine:** Pure-Go authenticated transport (clone, push, fetch).
- 🤖 **Model Context Protocol (MCP) First-Class:** 34 JSON-RPC tools mounted at \`/mcp\`.
- 🛡️ **Deterministic 3-Way Merge & Ruleset Engine:** Enterprise policy gates, bypass hierarchies, and CIDR allowlists.
- 💎 **Space UI Living Interface:** Tactile squircle components, glowing status pulses, and living agent telemetry.
`

export const sampleCommits: CommitItem[] = [
  {
    sha: '7aa844697bf92d4f58c707d8ceca47e1d882b524',
    shortSha: '7aa8446',
    message: 'docs: expand Diátaxis site to 15 pages and install global agent skills',
    author: { name: 'Roshan Ravani', email: 'roshan@exist.io', isAgent: false },
    timestamp: '2 hours ago',
    verified: true,
  },
  {
    sha: '3c85df1a2e9b114d7a8846bc04e8d32d09ff9201',
    shortSha: '3c85df1',
    message: 'feat(mcp): support batch pull request inline review threads',
    author: { name: 'Claude Agent', email: 'claude@anthropic.ai', isAgent: true },
    timestamp: '4 hours ago',
    verified: true,
  },
  {
    sha: 'e571c84b029283f1208933ba1107ec3200ff12aa',
    shortSha: 'e571c84',
    message: 'feat(ruleset): enforce CIDR IP allowlist filtering and bypass actors',
    author: { name: 'Roshan Ravani', email: 'roshan@exist.io', isAgent: false },
    timestamp: '6 hours ago',
    verified: true,
  },
  {
    sha: '839201af0029b9f77291a1348102bbcd91024312',
    shortSha: '839201a',
    message: 'feat(transport): wire go-git v6 bare repo storage engine with WAL',
    author: { name: 'Roshan Ravani', email: 'roshan@exist.io', isAgent: false },
    timestamp: '1 day ago',
    verified: true,
  },
]

export const samplePullRequests: PullRequest[] = [
  {
    number: 42,
    title: 'feat(mcp): expand autonomous agent review pipeline with streaming check runs',
    body: 'This pull request integrates continuous AI agent reviews directly with forge checks. When an agent runs via MCP, it registers a live check run on the commit SHA and reports inline review comments upon completion.',
    state: 'open',
    author: 'claude-agent',
    isAgent: true,
    headRef: 'feature/agent-checks',
    baseRef: 'main',
    createdAt: '3 hours ago',
    labels: [
      { name: 'mcp-agent', color: '#38bdf8' },
      { name: 'enhancement', color: '#a855f7' },
      { name: 'review-ready', color: '#10b981' },
    ],
    diffStats: { additions: 184, deletions: 23, filesChanged: 4 },
    checks: [
      { id: 'chk-1', name: 'Go Build & Test Matrix', status: 'completed', conclusion: 'success', duration: '42s' },
      { id: 'chk-2', name: 'MCP Schema Validator', status: 'completed', conclusion: 'success', duration: '12s' },
      { id: 'chk-3', name: 'Ruleset CIDR Policy Pass', status: 'completed', conclusion: 'success', duration: '4s' },
    ],
    comments: [
      {
        id: 'c-1',
        author: 'claude-agent',
        isAgent: true,
        createdAt: '2 hours ago',
        body: 'I have verified that all 34 MCP tools handle the new event schema deterministically. Test coverage across internal/mcp is 94.6%.',
      },
      {
        id: 'c-2',
        author: 'Roshan Ravani',
        isAgent: false,
        createdAt: '1 hour ago',
        body: 'Looks very clean. The check run dispatch aligns perfectly with our Phase 2 roadmap requirements. Ready to merge!',
      },
    ],
    mergeable: true,
    diffHunks: [
      {
        oldPath: 'internal/mcp/server.go',
        newPath: 'internal/mcp/server.go',
        status: 'modified',
        additions: 48,
        deletions: 5,
        hunk: `@@ -124,11 +124,19 @@ func (s *Server) handleToolCall(ctx context.Context, call Request) (Response, error) {
+	// Emit live check run telemetry for active agent sessions
+	if s.telemetrySink != nil {
+		s.telemetrySink.RecordAgentEvent(call.ToolName, call.Arguments)
+	}
+
 	switch call.ToolName {
 	case "list_pull_requests":
 		return s.handleListPullRequests(ctx, call.Arguments)
+	case "checks_read":
+		return s.handleChecksRead(ctx, call.Arguments)
 	case "merge_pull_request":
 		return s.handleMergePullRequest(ctx, call.Arguments)`,
      },
      {
        oldPath: 'internal/engine/ruleset.go',
        newPath: 'internal/engine/ruleset.go',
        status: 'modified',
        additions: 32,
        deletions: 2,
        hunk: `@@ -89,7 +89,12 @@ func (e *Engine) EvaluateMergeGate(pr *PullRequest) (EvaluationResult, error) {
+	if pr.RequirePassingChecks && !pr.AllChecksPassed() {
+		return EvaluationResult{Allowed: false, Reason: "Required status checks have not succeeded"}, nil
+	}
+	if pr.RequireReviewCount > 0 && pr.ApprovalCount < pr.RequireReviewCount {
+		return EvaluationResult{Allowed: false, Reason: "Minimum review approval threshold not met"}, nil
+	}`,
      },
    ],
  },
  {
    number: 41,
    title: 'docs: initialize Diátaxis framework site structure',
    body: 'Scaffolds tutorials, how-to guides, reference documents, and architectural explanations.',
    state: 'merged',
    author: 'Roshan Ravani',
    isAgent: false,
    headRef: 'docs/diataxis-overhaul',
    baseRef: 'main',
    createdAt: '1 day ago',
    labels: [
      { name: 'documentation', color: '#0ea5e9' },
      { name: 'completed', color: '#10b981' },
    ],
    diffStats: { additions: 620, deletions: 14, filesChanged: 15 },
    checks: [
      { id: 'chk-docs', name: 'Blume Doc Link Validator', status: 'completed', conclusion: 'success', duration: '6s' },
    ],
    comments: [
      { id: 'c-3', author: 'Roshan Ravani', isAgent: false, createdAt: '1 day ago', body: 'Merged cleanly into main.' },
    ],
    mergeable: true,
    diffHunks: [],
  },
]

export const sampleAgentActivities: AgentActivityItem[] = [
  {
    id: 'act-1',
    agentName: 'Claude Code',
    agentType: 'claude-code',
    toolName: 'list_pull_requests',
    status: 'success',
    latencyMs: 14,
    timestamp: '2 mins ago',
    requestPayload: { ownerSlug: 'exist', repoName: 'exist.io', state: 'open' },
    responseSummary: 'Returned 1 open pull request (#42)',
  },
  {
    id: 'act-2',
    agentName: 'Cursor Agent',
    agentType: 'cursor',
    toolName: 'checks_read',
    status: 'success',
    latencyMs: 8,
    timestamp: '6 mins ago',
    requestPayload: { ownerSlug: 'exist', repoName: 'exist.io', headSha: '7aa8446' },
    responseSummary: '3 suites evaluated: all green',
  },
  {
    id: 'act-3',
    agentName: 'Codex CLI',
    agentType: 'codex',
    toolName: 'create_pull_request_inline_comment',
    status: 'success',
    latencyMs: 22,
    timestamp: '14 mins ago',
    requestPayload: { pullNumber: 42, path: 'internal/mcp/server.go', line: 128 },
    responseSummary: 'Review comment thread #87 opened',
  },
  {
    id: 'act-4',
    agentName: 'exist-mcp-daemon',
    agentType: 'exist-mcp',
    toolName: 'get_ruleset',
    status: 'success',
    latencyMs: 5,
    timestamp: '25 mins ago',
    requestPayload: { ownerSlug: 'exist', repoName: 'exist.io', rulesetId: 'rs-main-prod' },
    responseSummary: 'Policy confirmed: RequireReview=1, EnforceChecks=true',
  },
]

// Live MCP Caller to local existd daemon
export async function queryLocalMcpTools(): Promise<any[]> {
  try {
    const res = await fetch('/mcp', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        jsonrpc: '2.0',
        id: Date.now(),
        method: 'tools/list',
      }),
    })
    if (!res.ok) throw new Error('Daemon returned HTTP ' + res.status)
    const json = await res.json()
    return json?.result?.tools || []
  } catch (err) {
    console.warn('Using built-in MCP catalog (daemon proxy unavailable):', err)
    return []
  }
}
