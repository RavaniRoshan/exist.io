// exist.io Isolated Mock Fixtures for Sandbox Demo Mode
// NOTE: These fixtures are strictly quarantined for evaluation & demo mode.
// In Production Mode, all data is fetched live from existd REST & MCP endpoints.

import type {
  RepoInfo,
  FileItem,
  CommitItem,
  PullRequest,
  AgentActivityItem,
  AuditEvent,
  ClusterMetrics,
  SsoProvider,
  EnterpriseUser,
} from './types'

export const mockDefaultRepo: RepoInfo = {
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

export const mockSampleFiles: FileItem[] = [
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

export const mockSampleReadme = `# exist.io

<p align="center">
  <strong>The Autonomous-Agent-Ready Git Forge</strong>
</p>

exist.io is a modern, pure-Go Git forge designed from the ground up for **human developers** and **autonomous AI coding agents**.

- ⚡ **100% Smart-HTTP Git Engine:** Pure-Go authenticated transport (clone, push, fetch).
- 🤖 **Model Context Protocol (MCP) First-Class:** 34 JSON-RPC tools mounted at \`/mcp\`.
- 🛡️ **Deterministic 3-Way Merge & Ruleset Engine:** Enterprise policy gates, bypass hierarchies, and CIDR allowlists.
- 💎 **Space UI Living Interface:** Tactile squircle components, glowing status pulses, and living agent telemetry.
`

export const mockSampleCommits: CommitItem[] = [
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

export const mockSamplePullRequests: PullRequest[] = [
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

export const mockSampleAgentActivities: AgentActivityItem[] = [
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
]

export const mockAuditEvents: AuditEvent[] = [
  {
    id: 'aud-101',
    timestamp: '2026-10-15 03:42:18 UTC',
    actor: 'roshan@exist.io',
    actorType: 'user',
    action: 'repo.create_ruleset',
    resource: 'exist/exist.io:rs-main-prod',
    clientIp: '192.168.1.104',
    severity: 'info',
    details: 'Enforced required approvals count = 1 and passing checks check_run.matrix',
  },
  {
    id: 'aud-102',
    timestamp: '2026-10-15 03:31:05 UTC',
    actor: 'claude-agent',
    actorType: 'agent',
    action: 'mcp.tool_call:create_pull_request_inline_comment',
    resource: 'exist/exist.io:pr-42',
    clientIp: '127.0.0.1',
    severity: 'info',
    details: 'Line review comment created on internal/mcp/server.go:128',
  },
  {
    id: 'aud-103',
    timestamp: '2026-10-15 02:14:50 UTC',
    actor: 'unknown-client',
    actorType: 'system',
    action: 'git.push_blocked_by_cidr',
    resource: 'exist/exist.io',
    clientIp: '203.0.113.88',
    severity: 'warning',
    details: 'Connection rejected: Client IP 203.0.113.88 not permitted by CIDR allowlist 192.168.1.0/24',
  },
  {
    id: 'aud-104',
    timestamp: '2026-10-15 01:50:22 UTC',
    actor: 'admin@exist.io',
    actorType: 'user',
    action: 'security.ssh_ca_added',
    resource: 'namespace:exist',
    clientIp: '10.0.4.12',
    severity: 'info',
    details: 'Added trusted HashiCorp Vault SSH Root CA fingerprint SHA256:d8a0c23...',
  },
  {
    id: 'aud-105',
    timestamp: '2026-10-14 22:18:11 UTC',
    actor: 'git-transport',
    actorType: 'system',
    action: 'security.secret_push_blocked',
    resource: 'exist/exist.io:refs/heads/feature/test-keys',
    clientIp: '192.168.1.55',
    severity: 'critical',
    details: 'Pre-receive hook rejected commit: High-entropy AWS Access Key (AKIA...) detected in src/config.go',
  },
]

export const mockClusterMetrics: ClusterMetrics = {
  serverStatus: 'healthy',
  uptime: '4 days, 16 hours, 22 minutes',
  version: 'existd v0.2.0-enterprise (linux/amd64)',
  activeGitSessions: 2,
  sqliteWalSizeBytes: 14680064, // ~14MB
  databasePath: '/var/lib/exist/meta.db (WAL mode, SQLite 3.44+)',
  goroutines: 38,
  allocatedMemoryMb: 64.2,
  cpuUsagePercent: 1.8,
  litestreamSyncStatus: 'active',
  lastBackupTimestamp: '3 minutes ago (s3://exist-backups-prod/wal/)',
  mcpMountedToolsCount: 34,
}

export const mockSsoProviders: SsoProvider[] = [
  {
    id: 'sso-1',
    name: 'Okta Enterprise SSO',
    type: 'saml',
    enabled: true,
    issuerUrl: 'https://identity.company.com/app/exist/sso/saml',
    ssoUrl: 'https://identity.company.com/app/exist/sso/saml/login',
    certificateFingerprint: 'SHA256:4A:8B:12:F4:99:... (Valid until 2028)',
    scimEnabled: true,
    enforceForDomain: 'exist.io',
  },
  {
    id: 'sso-2',
    name: 'Keycloak OIDC',
    type: 'oidc',
    enabled: false,
    issuerUrl: 'https://auth.internal.corp/realms/engineering',
    ssoUrl: 'https://auth.internal.corp/realms/engineering/protocol/openid-connect/auth',
    scimEnabled: false,
  },
]

export const mockEnterpriseUsers: EnterpriseUser[] = [
  {
    id: 'u-1',
    username: 'roshan',
    email: 'roshan@exist.io',
    name: 'Roshan Ravani',
    role: 'owner',
    isAgent: false,
    status: 'active',
    twoFactorEnabled: true,
    lastActive: 'Just now',
    sshKeyCount: 2,
    patCount: 3,
  },
  {
    id: 'u-2',
    username: 'claude-agent',
    email: 'claude@anthropic.ai',
    name: 'Claude Code Agent',
    role: 'maintainer',
    isAgent: true,
    status: 'active',
    twoFactorEnabled: false,
    lastActive: '2 minutes ago',
    sshKeyCount: 1,
    patCount: 1,
  },
  {
    id: 'u-3',
    username: 'sec-auditor',
    email: 'compliance@corp.exist.io',
    name: 'Security Compliance Bot',
    role: 'auditor',
    isAgent: true,
    status: 'active',
    twoFactorEnabled: true,
    lastActive: '1 hour ago',
    sshKeyCount: 0,
    patCount: 1,
  },
  {
    id: 'u-4',
    username: 'dev-contractor',
    email: 'contractor@external.org',
    name: 'Dev Contractor',
    role: 'contributor',
    isAgent: false,
    status: 'invited',
    twoFactorEnabled: false,
    lastActive: 'Never',
    sshKeyCount: 0,
    patCount: 0,
  },
]

export const mockRulesets: import('./types').RulesetRecord[] = [
  {
    id: 'rs-main-prod',
    name: 'Production Mainline Protection',
    enforcement: 'active',
    kind: 'branch',
    includedRefNames: ['refs/heads/main'],
    rules: [
      { type: 'pull_request', parameters: { required_approving_review_count: 1 } },
      { type: 'required_status_checks', parameters: { checks: ['Go Build & Test Matrix'] } },
      { type: 'required_signatures', parameters: {} },
      { type: 'non_fast_forward', parameters: {} },
    ],
  },
]

export const mockIPAllowlistEntries: import('./types').IPAllowlistEntry[] = [
  { id: 'ip-1', cidr: '10.0.0.0/8', description: 'Internal VPC / Container Mesh', created_at: '2026-03-01T10:00:00Z' },
  { id: 'ip-2', cidr: '192.168.1.0/24', description: 'On-premise Engineering Subnet', created_at: '2026-03-01T10:00:00Z' },
  { id: 'ip-3', cidr: '172.16.0.0/12', description: 'Agent Sandbox Isolated Bridge', created_at: '2026-03-01T10:00:00Z' },
]

export const mockSSHCAs: import('./types').SSHCARecord[] = [
  {
    id: 'ca-1',
    name: 'Vault Production CA',
    fingerprint: 'SHA256:4Xk9p0vV+eM3p0eQ4rT2Y9z8lM1nO7kP8wJ2lQ5rS0=',
    key_type: 'ssh-ed25519',
    required: false,
    created_at: '2026-02-15T08:30:00Z',
  },
]

export const mockWebhookDeliveries = [
  {
    id: 'wh-1',
    event: 'pull_request.opened',
    targetUrl: 'https://ci.internal.exist.io/hooks/forge',
    statusCode: 200,
    timestamp: '3 hours ago',
    latency: '34ms',
    payload: { action: 'opened', number: 42, repository: 'exist/exist.io', sender: 'claude-agent' },
  },
  {
    id: 'wh-2',
    event: 'check_run.completed',
    targetUrl: 'https://ci.internal.exist.io/hooks/forge',
    statusCode: 200,
    timestamp: '2 hours ago',
    latency: '22ms',
    payload: { check_name: 'Go Build & Test Matrix', conclusion: 'success', duration_s: 42 },
  },
  {
    id: 'wh-3',
    event: 'push',
    targetUrl: 'https://audit.corp.exist.io/log',
    statusCode: 204,
    timestamp: '1 hour ago',
    latency: '18ms',
    payload: { ref: 'refs/heads/main', after: '7aa844697bf92d4f58c707d8ceca47e1d882b524' },
  },
]

export const mockSampleIssues: import('./types').Issue[] = [
  {
    id: 'issue-1',
    number: 1,
    title: 'Support FIDO2 / WebAuthn hardware security keys for admin authentication',
    body: 'For strict air-gapped sovereign clusters, we need hardware-bound WebAuthn credentials alongside SAML/OIDC and SSH CAs.',
    state: 'open',
    author: 'Roshan Ravani',
    assignees: ['shiva', 'Exist Agent Reviewer'],
    labels: [
      { name: 'security', color: '#b60205', description: 'Security hardening and cryptographic policy' },
      { name: 'enhancement', color: '#a2eeef', description: 'New feature or improvement' },
    ],
    createdAt: '2 days ago',
    comments: [
      {
        id: 'ic-1',
        author: 'Exist Agent Reviewer',
        isAgent: true,
        body: 'I inspected the auth handlers in `internal/auth/`. Adding WebAuthn requires implementing the FIDO2 credential creation and challenge-response ceremonies. I can prepare a draft implementation.',
        createdAt: '1 day ago',
      },
      {
        id: 'ic-2',
        author: 'Roshan Ravani',
        isAgent: false,
        body: 'Approved. Proceed with implementing the WebAuthn challenge storage in SQLite WAL.',
        createdAt: '5 hours ago',
      },
    ],
  },
  {
    id: 'issue-2',
    number: 2,
    title: 'Optimize git-upload-pack packfile negotiation for large monorepos',
    body: 'When fetching repositories with over 100,000 refs, packfile negotiation roundtrips can be reduced using shallow-since and want-ref capabilities.',
    state: 'open',
    author: 'dev-lead',
    assignees: ['shiva'],
    labels: [
      { name: 'performance', color: '#d4c5f9', description: 'Performance and throughput optimization' },
      { name: 'git-engine', color: '#1d76db', description: 'Smart-HTTP Git wire protocol' },
    ],
    createdAt: '3 days ago',
    comments: [],
  },
  {
    id: 'issue-3',
    number: 3,
    title: 'Verify SQLite WAL checkpoint concurrency under heavy parallel git push',
    body: 'Load test SQLite WAL mode with 50 concurrent SSH and Smart-HTTP push workers to verify zero database locks.',
    state: 'closed',
    author: 'qa-engineer',
    assignees: ['Exist Agent Reviewer'],
    labels: [
      { name: 'verified', color: '#0e8a16', description: 'Tested and verified in benchmark suite' },
    ],
    createdAt: '1 week ago',
    closedAt: '2 days ago',
    comments: [
      {
        id: 'ic-3',
        author: 'Exist Agent Reviewer',
        isAgent: true,
        body: 'Completed 50-worker concurrent push benchmark. 0 lock timeouts recorded, WAL checkpoint completed in 1.4ms.',
        createdAt: '2 days ago',
      },
    ],
  },
]

export const mockSampleReleases: import('./types').ReleaseItem[] = [
  {
    id: 'rel-1',
    tagName: 'v0.4.0-enterprise',
    name: 'v0.4.0 Sovereign Enterprise & AI Agent Bridge',
    body: `### What's Changed\n- **Sovereign Git Engine**: Pure-Go Smart-HTTP and Git SSH daemon on port 2222.\n- **Model Context Protocol**: 35 native tools mounted at \`/mcp\` for autonomous agents.\n- **SQLite WAL Ledger**: Transactional ref storage, cryptographic commit signing, and audit logging.\n- **GitHub Primer UI**: Pixel-accurate Primer design system with 6px squircle radius and OKLCH color theme.\n- **Enterprise Identity**: SAML 2.0, OIDC SSO, SCIM 2.0 user directory, and SSH CA enrollment.`,
    author: 'shiva',
    isDraft: false,
    isPrerelease: false,
    createdAt: '2026-10-15T06:00:00Z',
    publishedAt: '2026-10-15T06:30:00Z',
    commitSha: '5c92f55',
    assets: [
      { name: 'existd-linux-amd64.tar.gz', size: 34500000, downloadCount: 412, sha256: '9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b' },
      { name: 'existd-linux-arm64.tar.gz', size: 32800000, downloadCount: 289, sha256: '1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b' },
      { name: 'existd-darwin-arm64.tar.gz', size: 33100000, downloadCount: 351, sha256: '4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9a8b7c6d5e' },
      { name: 'exist-iac-terraform.zip', size: 1420000, downloadCount: 198, sha256: '3e2f1a0b9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d' },
      { name: 'v0.4.0-enterprise.sig', size: 512, downloadCount: 140, sha256: '0d9e8f7a6b5c4d3e2f1a0b9a8b7c6d5e4f3a2b1c' },
    ],
  },
  {
    id: 'rel-2',
    tagName: 'v0.3.0',
    name: 'v0.3.0 Pure-Go Git Engine & MCP Server',
    body: `Initial release featuring pure-Go Git transport, deterministic 3-way merge algorithm, and basic agent tool execution.`,
    author: 'shiva',
    isDraft: false,
    isPrerelease: false,
    createdAt: '2026-10-01T12:00:00Z',
    publishedAt: '2026-10-01T14:00:00Z',
    commitSha: '1a2b3c4',
    assets: [
      { name: 'existd-linux-amd64.tar.gz', size: 28400000, downloadCount: 820 },
      { name: 'existd-darwin-arm64.tar.gz', size: 27900000, downloadCount: 654 },
    ],
  },
]


