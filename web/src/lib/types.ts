// exist.io Enterprise Domain & Data Types

export interface RepoInfo {
  ownerSlug: string
  name: string
  description: string
  defaultBranch: string
  isPrivate: boolean
  isArchived?: boolean
  starsCount: number
  forksCount: number
  commitsCount?: number
  branchesCount?: number
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
  verification?: {
    status: 'verified' | 'unverified' | 'unsigned' | 'invalid'
    key_type?: 'ssh' | 'gpg'
    key_id?: string
    signer_email?: string
    reason?: string
  }
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
  side?: string
  threadId?: string
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
  autoMerge?: boolean
  diffHunks: {
    oldPath: string
    newPath: string
    status: 'modified' | 'added' | 'deleted'
    additions: number
    deletions: number
    hunk: string
  }[]
}

export interface NotificationItem {
  id: string
  userId: string
  eventType: string
  title: string
  body: string
  url: string
  read: boolean
  createdAt: string
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

export interface SSOConfigData {
  configured: boolean
  provider_type?: 'oidc' | 'saml'
  issuer_url?: string
  client_id?: string
  client_secret?: string
  metadata_xml?: string
  enabled?: boolean
  enforce_sso?: boolean
}

export interface UserSSHKey {
  id: string
  name: string
  fingerprint: string
  key_type: string
  created_at: string
}

export interface PersonalAccessToken {
  id: string
  name: string
  token_prefix?: string
  raw_token?: string
  scopes?: string[]
  expires_at: string
  created_at: string
}

export interface IPAllowlistEntry {
  id: string
  cidr: string
  description: string
  created_at: string
}

export interface SSHCARecord {
  id: string
  name: string
  fingerprint: string
  key_type: string
  required?: boolean
  created_at: string
}

export interface RulesetRecord {
  id: string
  name: string
  enforcement: 'active' | 'evaluate' | 'disabled'
  kind: string
  includedRefNames?: string[]
  rules?: Array<{ type: string; parameters?: Record<string, any> }>
}

export interface SymbolItem {
  name: string
  kind: string
  file: string
  line: number
  signature?: string
  container?: string
  doc?: string
}

export interface AgentLease {
  token: string
  id: string
  agentId: string
  repoId: string
  issuedAt: string
  expiresAt: string
}

export interface AgentReviewResult {
  verdict: 'approve' | 'request_changes' | 'comment'
  conclusion: 'success' | 'failure' | 'neutral'
  score: number
  summary: string
  overviewMd: string
  violations: Array<{
    file: string
    line: number
    severity: 'error' | 'warning' | 'info'
    rule: string
    message: string
    snippet?: string
  }>
  reviewId: string
  checkRunId: string
  createdAt: string
}

export interface IssueLabel {
  name: string
  color: string
  description?: string
}

export interface IssueComment {
  id: string
  author: string
  isAgent?: boolean
  body: string
  createdAt: string
}

export interface Issue {
  id: string
  number: number
  title: string
  body: string
  state: 'open' | 'closed'
  author: string
  assignees?: string[]
  labels?: IssueLabel[]
  createdAt: string
  closedAt?: string
  comments?: IssueComment[]
}

export interface ReleaseAsset {
  name: string
  size: number
  downloadCount: number
  sha256?: string
}

export interface ReleaseItem {
  id: string
  tagName: string
  name: string
  body: string
  author: string
  isDraft: boolean
  isPrerelease: boolean
  createdAt: string
  publishedAt: string
  commitSha: string
  assets: ReleaseAsset[]
}

export interface RepositoryRecord {
  id: string
  name: string
  full_name: string
  default_branch: string
  visibility: string
  is_private: boolean
  clone_url: string
  web_url: string
  updated_at: string
  description?: string
  archived?: boolean
  stars?: number
  forks?: number
  language?: string
  languageColor?: string
}

export interface ClusterSystemInfo {
  version: string
  publicUrl: string
  gitHttpUrl: string
  sshAddr: string
  mcpUrl: string
  walStatus: string
  engine: string
}

export interface DeployKeyItem {
  id: string
  repo_id: string
  title: string
  key: string
  fingerprint: string
  read_only: boolean
  last_used_at?: string
  created_at: string
}

export interface RepoWebhookItem {
  id: string
  repo_id: string
  url: string
  content_type: string
  secret?: string
  events: string
  active: boolean
  insecure_ssl: boolean
  created_at: string
  updated_at: string
}

export interface RepoWebhookDeliveryItem {
  id: string
  webhook_id: string
  event: string
  guid: string
  request_headers: string
  request_payload: string
  response_status_code: number
  response_headers: string
  response_body: string
  duration_ms: number
  delivered: boolean
  created_at: string
}


