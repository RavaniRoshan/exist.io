// exist.io Enterprise Domain & Data Types

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

export interface AuditEvent {
  id: string
  timestamp: string
  actor: string
  actorType: 'user' | 'agent' | 'system'
  action: string
  resource: string
  clientIp: string
  severity: 'info' | 'warning' | 'critical'
  details: string
}

export interface ClusterMetrics {
  serverStatus: 'healthy' | 'degraded' | 'syncing'
  uptime: string
  version: string
  activeGitSessions: number
  sqliteWalSizeBytes: number
  databasePath: string
  goroutines: number
  allocatedMemoryMb: number
  cpuUsagePercent: number
  litestreamSyncStatus: 'active' | 'standby' | 'disabled'
  lastBackupTimestamp: string
  mcpMountedToolsCount: number
}

export interface SsoProvider {
  id: string
  name: string
  type: 'saml' | 'oidc'
  enabled: boolean
  issuerUrl: string
  ssoUrl: string
  certificateFingerprint?: string
  scimEnabled: boolean
  enforceForDomain?: string
}

export interface EnterpriseUser {
  id: string
  username: string
  email: string
  name: string
  role: 'owner' | 'admin' | 'maintainer' | 'contributor' | 'auditor'
  isAgent?: boolean
  status: 'active' | 'suspended' | 'invited'
  twoFactorEnabled: boolean
  lastActive: string
  sshKeyCount: number
  patCount: number
}

export interface UserProfile {
  id: string
  email: string
  login: string
  name: string
  role: 'admin' | 'member' | 'service'
  status: 'active' | 'suspended'
  createdAt: string
}
