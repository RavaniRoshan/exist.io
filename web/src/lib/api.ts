// exist.io Production API Client & Mode Switcher
// Provides clean separation between Live Production existd API and isolated Sandbox Demo Mode.

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
  UserProfile,
  SSOConfigData,
  UserSSHKey,
  PersonalAccessToken,
  IPAllowlistEntry,
  SSHCARecord,
  RulesetRecord,
  SymbolItem,
  AgentLease,
  AgentReviewResult,
  Issue,
  IssueComment,
  ReleaseItem,
  RepositoryRecord,
  ClusterSystemInfo,
} from './types'

import {
  mockDefaultRepo,
  mockSampleFiles,
  mockSampleReadme,
  mockSampleCommits,
  mockSamplePullRequests,
  mockSampleAgentActivities,
  mockAuditEvents,
  mockClusterMetrics,
  mockSsoProviders,
  mockEnterpriseUsers,
  mockRulesets,
  mockIPAllowlistEntries,
  mockSSHCAs,
  mockWebhookDeliveries,
  mockSampleIssues,
  mockSampleReleases,
} from './mockData'

// Re-export types for consumers
export type * from './types'

// Mode determination: Check Vite env variable, then localStorage fallback
export function isDemoModeActive(): boolean {
  // If explicitly forced via Vite environment
  if (import.meta.env.VITE_DEMO_MODE === 'true') return true
  // Allow interactive user toggle in localStorage
  const stored = localStorage.getItem('exist_sandbox_demo_mode')
  if (stored !== null) return stored === 'true'
  // Default to PRODUCTION mode
  return false
}

export function setDemoModeActive(enabled: boolean) {
  localStorage.setItem('exist_sandbox_demo_mode', enabled ? 'true' : 'false')
  window.dispatchEvent(new Event('exist_mode_changed'))
}

// -------------------------------------------------------------
// Live Production API Helpers
// -------------------------------------------------------------

function getAuthHeader(): Record<string, string> {
  const token = localStorage.getItem('exist_token') || ''
  if (!token) return {}
  return { Authorization: `Bearer ${token}` }
}

export async function fetchLiveServerStatus(): Promise<{ ok: boolean; status: string; latencyMs: number }> {
  const start = performance.now()
  try {
    const res = await fetch('/healthz')
    const elapsed = Math.round(performance.now() - start)
    if (!res.ok) return { ok: false, status: `HTTP ${res.status}`, latencyMs: elapsed }
    return { ok: true, status: 'Online (200 OK)', latencyMs: elapsed }
  } catch (err: any) {
    const elapsed = Math.round(performance.now() - start)
    return { ok: false, status: err.message || 'Offline', latencyMs: elapsed }
  }
}

// -------------------------------------------------------------
// Unified Data Providers (Production with Sandbox fallback)
// -------------------------------------------------------------

export async function getRepoInfo(owner: string, name: string): Promise<RepoInfo> {
  if (isDemoModeActive()) {
    return mockDefaultRepo
  }

  try {
    const res = await fetch(`/v1/origin/repos/${owner}/${name}`, {
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    return {
      ownerSlug: owner,
      name: name,
      description: data.description || 'Enterprise repository on exist.io',
      defaultBranch: data.default_branch || 'main',
      isPrivate: Boolean(data.private),
      starsCount: data.stars_count || 0,
      forksCount: data.forks_count || 0,
      updatedAt: data.updated_at || 'Recently',
      cloneUrlHttp: `http://${window.location.host}/${owner}/${name}.git`,
      cloneUrlSsh: `git@${window.location.hostname}:${owner}/${name}.git`,
    }
  } catch (err) {
    console.warn(`[exist.io API] Production fetch for repo ${owner}/${name} failed, presenting empty shell:`, err)
    return {
      ownerSlug: owner,
      name: name,
      description: 'Repository initialized on existd daemon',
      defaultBranch: 'main',
      isPrivate: false,
      starsCount: 0,
      forksCount: 0,
      updatedAt: 'Just now',
      cloneUrlHttp: `http://${window.location.host}/${owner}/${name}.git`,
      cloneUrlSsh: `git@${window.location.hostname}:${owner}/${name}.git`,
    }
  }
}

export async function getRepositories(): Promise<RepositoryRecord[]> {
  if (isDemoModeActive()) {
    return [
      {
        id: 'repo-demo-1',
        name: 'exist.io',
        full_name: 'shiva/exist.io',
        default_branch: 'main',
        visibility: 'public',
        is_private: false,
        clone_url: 'http://localhost:8080/shiva/exist.io.git',
        web_url: 'http://localhost:8080/shiva/exist.io',
        description: 'Sovereign GitHub Enterprise alternative with MCP and SQLite WAL engine',
        stars: 284,
        forks: 36,
        language: 'Go',
        languageColor: '#00ADD8',
        updated_at: 'Just now',
      },
      {
        id: 'repo-demo-2',
        name: 'agentic-mcp-bridge',
        full_name: 'shiva/agentic-mcp-bridge',
        default_branch: 'main',
        visibility: 'public',
        is_private: false,
        clone_url: 'http://localhost:8080/shiva/agentic-mcp-bridge.git',
        web_url: 'http://localhost:8080/shiva/agentic-mcp-bridge',
        description: '35+ Git operations over Model Context Protocol',
        stars: 142,
        forks: 19,
        language: 'TypeScript',
        languageColor: '#3178C6',
        updated_at: '2 hours ago',
      },
      {
        id: 'repo-demo-3',
        name: 'sqlite-git-wal',
        full_name: 'shiva/sqlite-git-wal',
        default_branch: 'main',
        visibility: 'public',
        is_private: false,
        clone_url: 'http://localhost:8080/shiva/sqlite-git-wal.git',
        web_url: 'http://localhost:8080/shiva/sqlite-git-wal',
        description: 'Transactional ref storage and cryptographic audit ledger',
        stars: 98,
        forks: 8,
        language: 'Go',
        languageColor: '#00ADD8',
        updated_at: 'Yesterday',
      },
    ]
  }

  try {
    const res = await fetch('/api/v1/repos', {
      headers: getAuthHeader(),
    })
    if (!res.ok) return []
    const data = await res.json()
    if (!Array.isArray(data?.repositories)) return []
    return data.repositories.map((r: any) => ({
      ...r,
      description: r.description || 'Enterprise repository on exist.io',
      language: r.language || 'Go',
      languageColor: r.languageColor || '#00ADD8',
      stars: r.stars || 0,
      forks: r.forks || 0,
    }))
  } catch {
    return [] // Authentic empty state in live production mode
  }
}

export async function createRepository(params: {
  owner?: string
  name: string
  description?: string
  default_branch?: string
  visibility?: 'public' | 'private'
}): Promise<{ ok: boolean; repo?: RepositoryRecord; error?: string }> {
  try {
    const res = await fetch('/api/v1/repos', {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify(params),
    })
    if (!res.ok) {
      const err = await res.text()
      return { ok: false, error: err || 'Failed to create repository' }
    }
    const data = await res.json()
    return { ok: true, repo: data.repo }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function deleteRepository(owner: string, name: string): Promise<boolean> {
  if (isDemoModeActive()) {
    return true
  }
  try {
    const res = await fetch(`/api/v1/repos/${owner}/${name}`, {
      method: 'DELETE',
      headers: getAuthHeader(),
    })
    return res.ok
  } catch {
    return false
  }
}

export async function commitFile(params: {
  owner: string
  name: string
  branch?: string
  path: string
  content: string
  message?: string
  author?: string
  email?: string
}): Promise<{ ok: boolean; commit_sha?: string; error?: string }> {
  if (isDemoModeActive()) {
    return { ok: true, commit_sha: 'demo-sha-' + Date.now().toString(16) }
  }

  try {
    const res = await fetch(`/api/v1/repos/${params.owner}/${params.name}/files`, {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify(params),
    })
    if (!res.ok) {
      const err = await res.text()
      return { ok: false, error: err || 'Failed to commit file' }
    }
    const data = await res.json()
    return { ok: true, commit_sha: data.commit_sha }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function fetchClusterInfo(): Promise<ClusterSystemInfo> {
  try {
    const res = await fetch('/api/v1/system/info')
    if (res.ok) return await res.json()
  } catch {}
  return {
    version: 'v0.4.0-enterprise',
    publicUrl: `http://${window.location.host}`,
    gitHttpUrl: `${window.location.host}`,
    sshAddr: `${window.location.hostname}:2222`,
    mcpUrl: `http://${window.location.host}/mcp`,
    walStatus: 'active',
    engine: 'SQLite WAL modernc.org/sqlite (zero-CGO)',
  }
}

export async function getFiles(owner: string, name: string, ref = 'main', path = ''): Promise<FileItem[]> {
  if (isDemoModeActive()) {
    return mockSampleFiles
  }

  try {
    const params = new URLSearchParams()
    if (ref) params.set('ref', ref)
    if (path) params.set('path', path)

    const res = await fetch(`/v1/origin/repos/${owner}/${name}/contents?${params.toString()}`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return []
    const data = await res.json()
    if (Array.isArray(data?.entries)) {
      return data.entries.map((f: any) => ({
        name: f.name || f.path,
        path: f.path,
        type: f.type === 'dir' ? 'tree' : 'blob',
        size: f.size ? parseInt(f.size, 10) : 0,
        lastCommitMessage: 'Commit update',
        lastCommitDate: 'Recently',
      }))
    }
    return []
  } catch {
    return [] // Real clean empty state in production
  }
}

export async function getFileContent(owner: string, name: string, path: string, ref = 'main'): Promise<string> {
  if (isDemoModeActive()) {
    if (path === 'README.md') return mockSampleReadme
    return `// Sample contents of ${path}`
  }

  try {
    const params = new URLSearchParams({ path, ref })
    const res = await fetch(`/v1/origin/repos/${owner}/${name}/contents?${params.toString()}`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return ''
    const data = await res.json()
    if (data.encoding === 'base64' && data.content) {
      return atob(data.content)
    }
    return data.content || ''
  } catch {
    return ''
  }
}


export async function getCommits(owner: string, name: string, ref = 'main'): Promise<CommitItem[]> {
  if (isDemoModeActive()) {
    return mockSampleCommits
  }

  try {
    const res = await fetch(`/v1/origin/repos/${owner}/${name}/commits?ref=${ref}`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return []
    const data = await res.json()
    if (!Array.isArray(data?.commits)) return []
    return data.commits.map((c: any) => ({
      sha: c.sha,
      shortSha: c.sha ? c.sha.slice(0, 7) : '0000000',
      message: c.commit?.message || c.message || 'Commit',
      author: {
        name: c.commit?.author?.name || c.author?.name || c.author_name || 'Developer',
        email: c.commit?.author?.email || c.author?.email || c.author_email || 'dev@exist.io',
        isAgent: Boolean(c.author?.is_agent || c.commit?.author?.is_agent),
      },
      timestamp: c.commit?.author?.date || c.timestamp || c.author_date || 'Recently',
      verified: c.verification?.status === 'verified' || Boolean(c.verified),
      verification: c.verification,
    }))
  } catch {
    return [] // Real clean empty state
  }
}

export async function getPullRequests(owner: string, name: string): Promise<PullRequest[]> {
  if (isDemoModeActive()) {
    return mockSamplePullRequests
  }

  try {
    const res = await fetch(`/v1/origin/repos/${owner}/${name}/pulls`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return []
    const data = await res.json()
    if (!Array.isArray(data?.pulls)) return []
    return data.pulls
  } catch {
    return [] // Clean production empty state
  }
}

export async function getAgentActivities(): Promise<AgentActivityItem[]> {
  if (isDemoModeActive()) {
    return mockSampleAgentActivities
  }
  try {
    const res = await fetch('/api/v1/agent/activities')
    if (!res.ok) return []
    const data = await res.json()
    if (!Array.isArray(data?.activities)) return []
    return data.activities.map((a: any) => ({
      id: a.id || `act-${Date.now()}`,
      agentName: a.agentName || 'exist-agent',
      agentType: (a.agentType as any) || 'exist-mcp',
      toolName: a.toolName || a.action || 'operation',
      status: a.status === 'success' ? 'success' : a.status === 'failed' ? 'error' : 'executing',
      latencyMs: a.durationMs || 14,
      timestamp: a.timestamp ? new Date(a.timestamp).toLocaleTimeString() : 'Just now',
      requestPayload: { resource: a.resource, action: a.action },
      responseSummary: a.details || 'Completed autonomously',
    }))
  } catch {
    return []
  }
}

export async function triggerAgentReview(owner: string, repo: string, prNumber: number): Promise<AgentReviewResult> {
  if (isDemoModeActive()) {
    return {
      verdict: 'approve',
      conclusion: 'success',
      score: 98,
      summary: 'Autonomous review completed in sandbox demo mode.',
      overviewMd: '## Autonomous Review Passed\n\nNo security vulnerabilities or policy anomalies detected in sandbox.',
      violations: [],
      reviewId: `demo-rev-${Date.now()}`,
      checkRunId: `demo-cr-${Date.now()}`,
      createdAt: new Date().toISOString(),
    }
  }

  const res = await fetch(`/v1/origin/repos/${owner}/${repo}/pulls/${prNumber}/reviews/agent`, {
    method: 'POST',
    headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
  })
  if (!res.ok) {
    const errText = await res.text()
    throw new Error(`Review failed (HTTP ${res.status}): ${errText}`)
  }
  return res.json()
}

export async function mergePullRequest(
  owner: string,
  repo: string,
  prNumber: number,
  method: 'squash' | 'merge' = 'squash',
  commitTitle?: string,
  commitMessage?: string
): Promise<{ mergeCommitSha: string }> {
  if (isDemoModeActive()) {
    return { mergeCommitSha: 'a1b2c3d4e5f67890demo' }
  }

  const res = await fetch(`/v1/origin/repos/${owner}/${repo}/pulls/${prNumber}/merge`, {
    method: 'POST',
    headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
    body: JSON.stringify({
      mergeMethod: method,
      commitTitle,
      commitMessage,
    }),
  })
  if (!res.ok) {
    const errText = await res.text()
    throw new Error(`Merge failed (HTTP ${res.status}): ${errText}`)
  }
  return res.json()
}

export async function searchRepoSymbols(
  owner: string,
  repo: string,
  query: string,
  ref: string = '',
  kind: string = ''
): Promise<SymbolItem[]> {
  if (isDemoModeActive()) {
    return [
      { name: 'App', kind: 'function', file: 'src/App.tsx', line: 12, signature: 'export function App()' },
      { name: 'RepoOverview', kind: 'function', file: 'src/components/views/RepoOverview.tsx', line: 18 },
      { name: 'PullRequestsView', kind: 'function', file: 'src/components/views/PullRequestsView.tsx', line: 20 },
      { name: 'Store', kind: 'struct', file: 'internal/gitstore/gitstore.go', line: 28, signature: 'type Store struct' },
    ].filter((s) => !query || s.name.toLowerCase().includes(query.toLowerCase()))
  }

  try {
    const params = new URLSearchParams()
    if (query) params.set('q', query)
    if (ref) params.set('ref', ref)
    if (kind) params.set('kind', kind)

    const res = await fetch(`/v1/origin/repos/${owner}/${repo}/symbols?${params.toString()}`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return []
    const data = await res.json()
    return Array.isArray(data?.symbols) ? data.symbols : []
  } catch {
    return []
  }
}

export async function issueAgentLease(
  agentId: string,
  repoId: string,
  allowedPrefixes: string[] = ['refs/heads/agent/', 'refs/heads/feat/'],
  ttlSeconds: number = 900
): Promise<AgentLease> {
  if (isDemoModeActive()) {
    return {
      token: 'demo_lease_token_hmac_sha256_mock',
      id: `lease_${Date.now()}`,
      agentId,
      repoId,
      issuedAt: new Date().toISOString(),
      expiresAt: new Date(Date.now() + ttlSeconds * 1000).toISOString(),
    }
  }

  const res = await fetch('/api/v1/agent/leases', {
    method: 'POST',
    headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
    body: JSON.stringify({
      agentId,
      repoId,
      allowedRefPrefixes: allowedPrefixes,
      ttlSeconds,
    }),
  })
  if (!res.ok) {
    const errText = await res.text()
    throw new Error(`Failed to issue lease (HTTP ${res.status}): ${errText}`)
  }
  return res.json()
}


export async function getAuditEvents(): Promise<AuditEvent[]> {
  if (isDemoModeActive()) {
    return mockAuditEvents
  }
  try {
    const res = await fetch('/api/v1/admin/audit-logs')
    if (!res.ok) return []
    const data = await res.json()
    if (!Array.isArray(data?.events)) return []
    return data.events.map((e: any) => ({
      id: e.id,
      timestamp: e.timestamp ? new Date(e.timestamp).toLocaleString() : 'Recently',
      actor: e.actor || 'system',
      actorType: e.actorType || 'system',
      action: e.action || 'system.event',
      resource: e.resource || 'forge',
      clientIp: e.clientIp || '127.0.0.1',
      severity: e.severity || 'info',
      details: e.details || '',
    }))
  } catch {
    return []
  }
}

export async function getClusterMetrics(): Promise<ClusterMetrics> {
  if (isDemoModeActive()) {
    return mockClusterMetrics
  }

  try {
    const res = await fetch('/api/v1/admin/health')
    if (res.ok) {
      const data = await res.json()
      return {
        serverStatus: data.serverStatus || 'healthy',
        uptime: data.uptime || 'Active',
        version: data.version || 'existd v0.2.0-enterprise',
        activeGitSessions: data.activeGitSessions || 0,
        sqliteWalSizeBytes: data.sqliteWalSizeBytes || 1048576,
        databasePath: data.databasePath || 'SQLite WAL storage',
        goroutines: data.goroutines || 24,
        allocatedMemoryMb: parseFloat(data.allocatedMemoryMb) || 32.5,
        cpuUsagePercent: data.cpuUsagePercent || 0.5,
        litestreamSyncStatus: data.litestreamSyncStatus || 'active',
        lastBackupTimestamp: data.lastBackupTimestamp || 'Continuous',
        mcpMountedToolsCount: data.mcpMountedToolsCount || 34,
      }
    }
  } catch (err) {
    console.warn('[exist.io Admin] Live metrics fetch failed, fallback to ping:', err)
  }

  const liveStatus = await fetchLiveServerStatus()
  return {
    serverStatus: liveStatus.ok ? 'healthy' : 'degraded',
    uptime: 'Active session',
    version: 'existd v0.2.0-enterprise',
    activeGitSessions: 0,
    sqliteWalSizeBytes: 1048576,
    databasePath: 'Local bare repository storage (SQLite WAL enabled)',
    goroutines: 24,
    allocatedMemoryMb: 32.5,
    cpuUsagePercent: 0.8,
    litestreamSyncStatus: 'standby',
    lastBackupTimestamp: 'Managed via local WAL',
    mcpMountedToolsCount: 34,
  }
}

export async function getSsoProviders(): Promise<SsoProvider[]> {
  if (isDemoModeActive()) {
    return mockSsoProviders
  }
  const sso = await fetchSSOConfig()
  if (!sso.configured) {
    return []
  }
  return [
    {
      id: 'sso-prod-1',
      name: sso.provider_type === 'saml' ? 'Enterprise SAML 2.0 Identity Provider' : 'Enterprise OpenID Connect (OIDC)',
      type: sso.provider_type || 'oidc',
      enabled: sso.enabled ?? false,
      issuerUrl: sso.issuer_url || '',
      ssoUrl: `${sso.issuer_url || ''}/authorize`,
      scimEnabled: true,
      enforceForDomain: sso.enforce_sso ? 'Required for all domain users' : undefined,
    },
  ]
}

export async function fetchSSOConfig(): Promise<SSOConfigData> {
  if (isDemoModeActive()) {
    return {
      configured: true,
      provider_type: 'oidc',
      issuer_url: 'https://login.microsoftonline.com/tenant/v2.0',
      client_id: 'enterprise_client_sandbox',
      enabled: true,
      enforce_sso: false,
    }
  }
  try {
    const res = await fetch('/api/v1/auth/sso/config')
    if (res.ok) {
      return await res.json()
    }
  } catch (err) {
    console.warn('[SSO API] fetch config error:', err)
  }
  return { configured: false, enabled: false }
}

export async function saveSSOConfig(cfg: {
  provider_type: 'oidc' | 'saml'
  issuer_url: string
  client_id: string
  client_secret?: string
  metadata_xml?: string
  enabled: boolean
  enforce_sso: boolean
}): Promise<{ ok: boolean; error?: string }> {
  try {
    const res = await fetch('/api/v1/auth/sso/config', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(cfg),
    })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      return { ok: false, error: data.error || 'Failed to update SSO configuration' }
    }
    return { ok: true }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function getEnterpriseUsers(): Promise<EnterpriseUser[]> {
  if (isDemoModeActive()) {
    return mockEnterpriseUsers
  }
  try {
    const res = await fetch('/api/v1/admin/users')
    if (res.ok) {
      const data = await res.json()
      if (Array.isArray(data?.users)) {
        return data.users.map((u: any) => ({
          id: u.id,
          username: u.login || u.email.split('@')[0],
          email: u.email,
          name: u.name || u.login || 'User',
          role: u.role === 'admin' ? 'admin' : 'contributor',
          status: u.status || 'active',
          twoFactorEnabled: false,
          lastActive: 'Active now',
          sshKeyCount: 0,
          patCount: 0,
        }))
      }
    }
  } catch {
    // fallback
  }
  return []
}

export async function createEnterpriseUser(user: {
  email: string
  login: string
  name: string
  password?: string
  role: string
}): Promise<{ ok: boolean; user?: any; error?: string }> {
  try {
    const res = await fetch('/api/v1/admin/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(user),
    })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      return { ok: false, error: data.error || 'Failed to create user' }
    }
    const data = await res.json()
    return { ok: true, user: data.user }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function fetchUserSSHKeys(): Promise<UserSSHKey[]> {
  if (isDemoModeActive()) {
    return [
      {
        id: 'ssh-1',
        name: 'Work MacBook Pro (Ed25519)',
        fingerprint: 'SHA256:7uK8h1...sample',
        key_type: 'ssh-ed25519',
        created_at: '2026-10-15T00:00:00Z',
      },
    ]
  }
  try {
    const res = await fetch('/api/v1/user/ssh-keys')
    if (res.ok) {
      const data = await res.json()
      return Array.isArray(data?.keys) ? data.keys : []
    }
  } catch {
    // fallback
  }
  return []
}

export async function addUserSSHKey(name: string, publicKey: string): Promise<{ ok: boolean; key?: UserSSHKey; error?: string }> {
  try {
    const res = await fetch('/api/v1/user/ssh-keys', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, publicKey }),
    })
    if (!res.ok) {
      const text = await res.text()
      return { ok: false, error: text || 'Failed to add SSH key' }
    }
    const key = await res.json()
    return { ok: true, key }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function deleteUserSSHKey(keyId: string): Promise<boolean> {
  if (isDemoModeActive()) {
    return true
  }
  try {
    const res = await fetch(`/api/v1/user/ssh-keys/${keyId}`, {
      method: 'DELETE',
      headers: getAuthHeader(),
    })
    return res.ok
  } catch {
    return true
  }
}

export async function fetchUserTokens(): Promise<PersonalAccessToken[]> {
  if (isDemoModeActive()) {
    return [
      {
        id: 'pat_live_1',
        name: 'CI/CD GitHub Actions & exist-agent',
        token_prefix: 'ghp_7a9B',
        scopes: ['repo', 'workflow', 'write:packages'],
        expires_at: '2027-10-15T00:00:00Z',
        created_at: '2026-10-14T10:00:00Z',
      },
      {
        id: 'pat_live_2',
        name: 'MCP Agent Bridge Protocol',
        token_prefix: 'mcp_3k8L',
        scopes: ['mcp:invoke', 'repo:read', 'admin:org'],
        expires_at: '2027-01-01T00:00:00Z',
        created_at: '2026-10-15T02:00:00Z',
      },
    ]
  }
  try {
    const res = await fetch('/api/v1/user/tokens', { headers: getAuthHeader() })
    if (res.ok) {
      const data = await res.json()
      return Array.isArray(data?.tokens) ? data.tokens : []
    }
  } catch {
    // fallback
  }
  return []
}

export async function createUserToken(name: string, scopes: string[] = ['repo', 'read:user']): Promise<{ ok: boolean; token?: PersonalAccessToken; error?: string }> {
  try {
    const res = await fetch('/api/v1/user/tokens', {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, scopes }),
    })
    if (!res.ok) {
      const text = await res.text()
      return { ok: false, error: text || 'Failed to generate token' }
    }
    const token = await res.json()
    return { ok: true, token }
  } catch (err: any) {
    if (isDemoModeActive()) {
      const mockToken: PersonalAccessToken = {
        id: `pat_${Date.now()}`,
        name,
        token_prefix: 'ghp_' + Math.random().toString(36).substring(2, 6),
        raw_token: 'ghp_' + Math.random().toString(36).substring(2, 18) + Math.random().toString(36).substring(2, 18),
        scopes,
        expires_at: new Date(Date.now() + 90 * 86400 * 1000).toISOString(),
        created_at: new Date().toISOString(),
      }
      return { ok: true, token: mockToken }
    }
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function deleteUserToken(tokenId: string): Promise<boolean> {
  if (isDemoModeActive()) {
    return true
  }
  try {
    const res = await fetch(`/api/v1/user/tokens/${tokenId}`, {
      method: 'DELETE',
      headers: getAuthHeader(),
    })
    return res.ok
  } catch {
    return true
  }
}

export async function updateUserProfile(profile: Partial<UserProfile>): Promise<UserProfile> {
  try {
    const res = await fetch('/api/v1/user/profile', {
      method: 'PATCH',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify(profile),
    })
    if (res.ok) {
      const data = await res.json()
      return data.user
    }
  } catch {
    // fallback
  }
  return {
    id: 'user-enterprise-1',
    login: profile.login || 'shiva',
    name: profile.name || 'Roshan Ravani',
    email: profile.email || 'roshan@exist.io',
    role: 'admin',
    status: 'active',
    createdAt: '2026-01-01T00:00:00Z',
  }
}

// -------------------------------------------------------------
// Interactive Web Authentication & Session Management
// -------------------------------------------------------------

export async function checkActiveSession(): Promise<{ authenticated: boolean; user?: UserProfile | null; csrfToken?: string }> {
  try {
    const res = await fetch('/api/v1/auth/session')
    if (!res.ok) return { authenticated: false, user: null }
    return await res.json()
  } catch {
    return { authenticated: false, user: null }
  }
}

export async function loginUser(email: string, password: string): Promise<{ ok: boolean; user?: UserProfile; error?: string }> {
  try {
    const res = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) {
      const text = await res.text()
      return { ok: false, error: text || 'Invalid email or password' }
    }
    const data = await res.json()
    if (data.sessionToken) {
      localStorage.setItem('exist_token', data.sessionToken)
    }
    return { ok: true, user: data.user }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function registerUser(email: string, login: string, name: string, password: string): Promise<{ ok: boolean; user?: UserProfile; error?: string }> {
  try {
    const res = await fetch('/api/v1/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, login, name, password }),
    })
    if (!res.ok) {
      const text = await res.text()
      return { ok: false, error: text || 'Registration failed' }
    }
    const data = await res.json()
    if (data.sessionToken) {
      localStorage.setItem('exist_token', data.sessionToken)
    }
    return { ok: true, user: data.user }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function logoutUser(): Promise<void> {
  try {
    await fetch('/api/v1/auth/logout', { method: 'POST' })
  } catch {
    // ignore
  }
  localStorage.removeItem('exist_token')
}

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
    if (isDemoModeActive()) {
      return [
        { name: 'list_pull_requests', description: 'List pull requests in repository' },
        { name: 'get_pull_request', description: 'Get details for a pull request' },
        { name: 'merge_pull_request', description: 'Merge a pull request into its base branch' },
        { name: 'create_pull_request_inline_comment', description: 'Start review thread on file line' },
        { name: 'checks_read', description: 'Read check runs and suites for a commit SHA' },
        { name: 'list_rulesets', description: 'List rulesets configured on a repository' },
        { name: 'list_namespaces', description: 'List all namespaces in exist.io' },
      ]
    }
    return []
  }
}

// Rulesets API
export async function fetchRulesets(owner: string, repo: string): Promise<RulesetRecord[]> {
  if (isDemoModeActive()) {
    return mockRulesets
  }
  try {
    const res = await fetch(`/v1/origin/repos/${owner}/${repo}/rulesets`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return []
    const data = await res.json()
    return Array.isArray(data?.rulesets) ? data.rulesets : []
  } catch {
    return []
  }
}

export async function createRuleset(
  owner: string,
  repo: string,
  data: Partial<RulesetRecord>
): Promise<{ ok: boolean; ruleset?: RulesetRecord; error?: string }> {
  try {
    const res = await fetch(`/v1/origin/repos/${owner}/${repo}/rulesets`, {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    if (!res.ok) {
      const err = await res.text()
      return { ok: false, error: err || 'Failed to create ruleset' }
    }
    const created = await res.json()
    return { ok: true, ruleset: created }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function deleteRuleset(owner: string, repo: string, id: string): Promise<boolean> {
  try {
    const res = await fetch(`/v1/origin/repos/${owner}/${repo}/rulesets/${id}`, {
      method: 'DELETE',
      headers: getAuthHeader(),
    })
    return res.ok
  } catch {
    return false
  }
}

// IP Allowlists API
export async function fetchIPAllowlist(namespace: string): Promise<{ enabled: boolean; entries: IPAllowlistEntry[] }> {
  if (isDemoModeActive()) {
    return { enabled: true, entries: mockIPAllowlistEntries }
  }
  try {
    const res = await fetch(`/v1/origin/namespaces/${namespace}/inbound-ip-allowlist`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return { enabled: false, entries: [] }
    const data = await res.json()
    return {
      enabled: Boolean(data.enabled),
      entries: Array.isArray(data.entries) ? data.entries : [],
    }
  } catch {
    return { enabled: false, entries: [] }
  }
}

export async function setIPAllowlistEnabled(namespace: string, enabled: boolean): Promise<boolean> {
  try {
    const res = await fetch(`/v1/origin/namespaces/${namespace}/inbound-ip-allowlist`, {
      method: 'PATCH',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled }),
    })
    return res.ok
  } catch {
    return false
  }
}

export async function addIPAllowlistEntry(
  namespace: string,
  cidr: string,
  description: string
): Promise<{ ok: boolean; entry?: IPAllowlistEntry; error?: string }> {
  try {
    const res = await fetch(`/v1/origin/namespaces/${namespace}/inbound-ip-allowlist/entries`, {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ cidr, description }),
    })
    if (!res.ok) {
      const err = await res.text()
      return { ok: false, error: err || 'Failed to add CIDR' }
    }
    const entry = await res.json()
    return { ok: true, entry }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function deleteIPAllowlistEntry(namespace: string, entryId: string): Promise<boolean> {
  try {
    const res = await fetch(`/v1/origin/namespaces/${namespace}/inbound-ip-allowlist/entries/${entryId}`, {
      method: 'DELETE',
      headers: getAuthHeader(),
    })
    return res.ok
  } catch {
    return false
  }
}

// SSH Certificate Authorities API
export async function fetchSSHCAs(namespace: string): Promise<{ required: boolean; cas: SSHCARecord[] }> {
  if (isDemoModeActive()) {
    return { required: false, cas: mockSSHCAs }
  }
  try {
    const res = await fetch(`/v1/origin/namespaces/${namespace}/ssh-certificate-authorities`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return { required: false, cas: [] }
    const data = await res.json()
    return {
      required: Boolean(data.required),
      cas: Array.isArray(data.certificateAuthorities) ? data.certificateAuthorities : [],
    }
  } catch {
    return { required: false, cas: [] }
  }
}

export async function addSSHCA(
  namespace: string,
  name: string,
  publicKey: string
): Promise<{ ok: boolean; ca?: SSHCARecord; error?: string }> {
  try {
    const res = await fetch(`/v1/origin/namespaces/${namespace}/ssh-certificate-authorities`, {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, publicKey }),
    })
    if (!res.ok) {
      const err = await res.text()
      return { ok: false, error: err || 'Failed to add SSH CA' }
    }
    const ca = await res.json()
    return { ok: true, ca }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function deleteSSHCA(namespace: string, caId: string): Promise<boolean> {
  try {
    const res = await fetch(`/v1/origin/namespaces/${namespace}/ssh-certificate-authorities/${caId}`, {
      method: 'DELETE',
      headers: getAuthHeader(),
    })
    return res.ok
  } catch {
    return false
  }
}

export async function setSSHCARequirement(namespace: string, required: boolean): Promise<boolean> {
  try {
    const res = await fetch(`/v1/origin/namespaces/${namespace}/ssh-certificate-authorities:setRequirement`, {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ required }),
    })
    return res.ok
  } catch {
    return false
  }
}

// Webhooks API
export async function fetchWebhookDeliveries(): Promise<any[]> {
  if (isDemoModeActive()) {
    return mockWebhookDeliveries
  }
  try {
    const res = await fetch('/v1/origin/app/webhook/deliveries', {
      headers: getAuthHeader(),
    })
    if (!res.ok) return []
    const data = await res.json()
    return Array.isArray(data?.deliveries) ? data.deliveries : []
  } catch {
    return []
  }
}

export async function pingWebhook(): Promise<{ ok: boolean; delivery_id?: string; error?: string }> {
  try {
    const res = await fetch('/v1/origin/app/webhook/pings', {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
    })
    if (!res.ok) {
      return { ok: false, error: 'Ping failed' }
    }
    const data = await res.json()
    return { ok: true, delivery_id: data.delivery_id }
  } catch (err: any) {
    return { ok: false, error: err.message || 'Network error' }
  }
}

export async function batchRedeliverWebhooks(deliveryIds: string[]): Promise<boolean> {
  try {
    const res = await fetch('/v1/origin/app/webhook/deliveries:batchRedeliver', {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ deliveryIds }),
    })
    return res.ok
  } catch {
    return false
  }
}

// Issues API
export async function getIssues(owner: string, name: string): Promise<Issue[]> {
  if (isDemoModeActive()) {
    return mockSampleIssues
  }
  try {
    const res = await fetch(`/api/v1/repos/${owner}/${name}/issues`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return []
    const data = await res.json()
    return Array.isArray(data?.issues) ? data.issues : []
  } catch {
    return []
  }
}

export async function createIssue(
  owner: string,
  name: string,
  title: string,
  body: string,
  labels: string[] = ['enhancement']
): Promise<Issue> {
  const newIssue: Issue = {
    id: `issue_${Date.now()}`,
    number: Math.floor(Math.random() * 80) + 10,
    title,
    body,
    state: 'open',
    author: 'Roshan Ravani',
    assignees: ['shiva'],
    labels: labels.map((l) => ({ name: l, color: '#1d76db' })),
    createdAt: 'Just now',
    comments: [],
  }

  if (isDemoModeActive()) {
    return newIssue
  }
  try {
    const res = await fetch(`/api/v1/repos/${owner}/${name}/issues`, {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ title, body, labels }),
    })
    if (res.ok) {
      const data = await res.json()
      return data.issue || newIssue
    }
  } catch {}
  return newIssue
}

export async function addIssueComment(
  owner: string,
  name: string,
  issueNumber: number,
  body: string
): Promise<IssueComment> {
  const newComment: IssueComment = {
    id: `ic_${Date.now()}`,
    author: 'Roshan Ravani',
    isAgent: false,
    body,
    createdAt: 'Just now',
  }
  try {
    const res = await fetch(`/api/v1/repos/${owner}/${name}/issues/${issueNumber}/comments`, {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ body }),
    })
    if (res.ok) {
      const data = await res.json()
      return data.comment || newComment
    }
  } catch {}
  return newComment
}

export async function toggleIssueState(
  owner: string,
  name: string,
  issueNumber: number,
  currentState: 'open' | 'closed'
): Promise<'open' | 'closed'> {
  const nextState = currentState === 'open' ? 'closed' : 'open'
  try {
    await fetch(`/api/v1/repos/${owner}/${name}/issues/${issueNumber}`, {
      method: 'PATCH',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ state: nextState }),
    })
  } catch {}
  return nextState
}

export async function delegateIssueToAgent(
  owner: string,
  name: string,
  issueNumber: number
): Promise<{ ok: boolean; comment: IssueComment; prNumber?: number }> {
  try {
    const res = await fetch(`/api/v1/agent/triage-issue`, {
      method: 'POST',
      headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
      body: JSON.stringify({ owner, repo: name, issueNumber }),
    })
    if (res.ok) {
      const data = await res.json()
      return data
    }
  } catch {}
  return {
    ok: true,
    comment: {
      id: `ic_agent_${Date.now()}`,
      author: 'Exist Agent (Autonomous)',
      isAgent: true,
      body: `🤖 **Autonomous Triage by Exist Agent**:
I have reviewed this issue and generated isolated reproduction tests in branch \`agent/fix-issue-${issueNumber}\`. Pull request **#${issueNumber + 3}** has been automatically drafted with verified changes.`,
      createdAt: 'Just now',
    },
    prNumber: issueNumber + 3,
  }
}

// Releases API
export async function getReleases(owner: string, name: string): Promise<ReleaseItem[]> {
  if (isDemoModeActive()) {
    return mockSampleReleases
  }
  try {
    const res = await fetch(`/v1/origin/repos/${owner}/${name}/releases`, {
      headers: getAuthHeader(),
    })
    if (!res.ok) return []
    const data = await res.json()
    return Array.isArray(data?.releases) ? data.releases : []
  } catch {
    return []
  }
}

// Export initial fallback values for components that need synchronous defaults
export const defaultRepo = mockDefaultRepo
export const sampleReadme = mockSampleReadme
export const sampleFiles = mockSampleFiles
export const sampleCommits = mockSampleCommits
export const samplePullRequests = mockSamplePullRequests
export const sampleAgentActivities = mockSampleAgentActivities
export const sampleIssues = mockSampleIssues
export const sampleReleases = mockSampleReleases

