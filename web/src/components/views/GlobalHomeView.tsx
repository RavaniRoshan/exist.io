import React, { useState, useEffect } from 'react'
import {
  BookOpen,
  GitCommit,
  GitPullRequest,
  Bot,
  ShieldCheck,
  Server,
  Activity,
  Plus,
  Search,
  Tag,
  Copy,
  Check,
  Compass,
  Terminal,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Box, BoxHeader, BoxTitle, BoxBody } from '@/components/ui/Card'
import { CreateRepoModal } from '@/components/modals/CreateRepoModal'
import {
  getRepositories,
  getAgentActivities,
  isDemoModeActive,
  type RepositoryRecord,
  type AgentActivityItem,
} from '@/lib/api'
import { cn } from '@/lib/utils'

export interface GlobalHomeViewProps {
  onSelectRepo: (repoName: string) => void
  onSelectTab: (tab: string) => void
  onOpenTour?: () => void
  onOpenManual?: () => void
}

interface ActivityEvent {
  id: string
  type: 'pr' | 'commit' | 'agent' | 'release' | 'ruleset'
  actor: string
  action: string
  target: string
  timestamp: string
  details?: string
  sha?: string
}

const enterpriseFeed: ActivityEvent[] = [
  {
    id: 'evt-1',
    type: 'agent',
    actor: 'Exist Agent (Autonomous)',
    action: 'approved and autonomously merged pull request',
    target: '#4 feat: add SAML/OIDC SSO enterprise authentication',
    timestamp: '12 minutes ago',
    details: 'Verified branch protection rules, ran full Vitest suite, and executed SQLite WAL fast-forward commit.',
    sha: '5c92f55',
  },
  {
    id: 'evt-2',
    type: 'release',
    actor: 'shiva',
    action: 'published enterprise release',
    target: 'v0.4.0-enterprise',
    timestamp: '2 hours ago',
    details: 'Includes Sovereign Git Server, SQLite WAL Engine, MCP Protocol Bridge, and Primer Design System.',
  },
  {
    id: 'evt-3',
    type: 'commit',
    actor: 'Roshan Ravani',
    action: 'pushed 3 commits to branch',
    target: 'shiva/exist.io (main)',
    timestamp: '3 hours ago',
    sha: 'e92bc41',
  },
  {
    id: 'evt-4',
    type: 'ruleset',
    actor: 'Enterprise Policy Enforcer',
    action: 'enforced branch ruleset',
    target: 'Production-Mainline-Protection',
    timestamp: '5 hours ago',
    details: 'Cryptographic Ed25519 commit signature verified; force-pushes rejected across protected branches.',
  },
  {
    id: 'evt-5',
    type: 'pr',
    actor: 'shiva',
    action: 'opened pull request',
    target: '#5 feat: add UserProfileView and enterprise SSH credentials',
    timestamp: 'Yesterday',
    details: 'Enables developer profiles, 52-week activity contribution calendars, and self-hosted PAT management.',
  },
]

export const GlobalHomeView: React.FC<GlobalHomeViewProps> = ({
  onSelectRepo,
  onSelectTab,
  onOpenTour,
  onOpenManual,
}) => {
  const [filterType, setFilterType] = useState<'all' | 'agent' | 'pr' | 'commit'>('all')
  const [repoSearch, setRepoSearch] = useState('')
  const [copiedClone, setCopiedClone] = useState(false)
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)

  // Real or sandbox repositories
  const [repos, setRepos] = useState<RepositoryRecord[]>([
    {
      id: 'repo-1',
      name: 'exist.io',
      full_name: 'shiva/exist.io',
      default_branch: 'main',
      visibility: 'public',
      is_private: false,
      clone_url: 'http://localhost:8080/shiva/exist.io.git',
      web_url: 'http://localhost:8080/shiva/exist.io',
      description: 'Sovereign GitHub Enterprise alternative with MCP',
      updated_at: 'Just now',
    },
    {
      id: 'repo-2',
      name: 'agentic-mcp-bridge',
      full_name: 'shiva/agentic-mcp-bridge',
      default_branch: 'main',
      visibility: 'public',
      is_private: false,
      clone_url: 'http://localhost:8080/shiva/agentic-mcp-bridge.git',
      web_url: 'http://localhost:8080/shiva/agentic-mcp-bridge',
      description: '35+ Git operations over Model Context Protocol',
      updated_at: '2 hours ago',
    },
    {
      id: 'repo-3',
      name: 'sqlite-git-wal',
      full_name: 'shiva/sqlite-git-wal',
      default_branch: 'main',
      visibility: 'public',
      is_private: false,
      clone_url: 'http://localhost:8080/shiva/sqlite-git-wal.git',
      web_url: 'http://localhost:8080/shiva/sqlite-git-wal',
      description: 'Transactional ref storage and cryptographic audit ledger',
      updated_at: 'Yesterday',
    },
    {
      id: 'repo-4',
      name: 'iac-terraform-enterprise',
      full_name: 'shiva/iac-terraform-enterprise',
      default_branch: 'main',
      visibility: 'public',
      is_private: false,
      clone_url: 'http://localhost:8080/shiva/iac-terraform-enterprise.git',
      web_url: 'http://localhost:8080/shiva/iac-terraform-enterprise',
      description: 'Air-gapped multi-cloud IaC provisioning',
      updated_at: '3 days ago',
    },
    {
      id: 'repo-5',
      name: 'enterprise-auth-scim',
      full_name: 'shiva/enterprise-auth-scim',
      default_branch: 'main',
      visibility: 'private',
      is_private: true,
      clone_url: 'http://localhost:8080/shiva/enterprise-auth-scim.git',
      web_url: 'http://localhost:8080/shiva/enterprise-auth-scim',
      description: 'SAML 2.0 and SCIM 2.0 enterprise identity sync',
      updated_at: '5 days ago',
    },
  ])

  // Real or sandbox events
  const [events, setEvents] = useState<ActivityEvent[]>(enterpriseFeed)

  useEffect(() => {
    // If running in live production mode, fetch live cluster repositories and agent activities
    if (!isDemoModeActive()) {
      getRepositories().then((liveRepos) => {
        if (liveRepos && liveRepos.length > 0) {
          setRepos(liveRepos)
        } else {
          setRepos([])
        }
      })

      getAgentActivities().then((acts: AgentActivityItem[]) => {
        if (acts && acts.length > 0) {
          const liveEvents: ActivityEvent[] = acts.map((a, idx) => ({
            id: a.id || `live-${idx}`,
            type: 'agent',
            actor: a.agentName || 'exist-agent',
            action: `executed ${a.toolName}`,
            target: typeof a.requestPayload?.resource === 'string' ? a.requestPayload.resource : 'repository',
            timestamp: a.timestamp || 'Recently',
            details: a.responseSummary,
          }))
          setEvents(liveEvents)
        } else {
          setEvents([])
        }
      })
    }
  }, [])

  const filteredRepos = repos.filter((r) =>
    r.full_name?.toLowerCase().includes(repoSearch.toLowerCase()) ||
    r.name.toLowerCase().includes(repoSearch.toLowerCase()) ||
    (r.description && r.description.toLowerCase().includes(repoSearch.toLowerCase()))
  )

  const filteredEvents = events.filter((evt) => {
    if (filterType === 'all') return true
    if (filterType === 'agent') return evt.type === 'agent' || evt.type === 'ruleset'
    if (filterType === 'pr') return evt.type === 'pr'
    if (filterType === 'commit') return evt.type === 'commit'
    return true
  })

  const primaryRepo = repos[0] || {
    full_name: 'shiva/exist.io',
    clone_url: 'http://localhost:8080/shiva/exist.io.git',
  }
  const quickCloneUrl = primaryRepo.clone_url || `http://localhost:8080/${primaryRepo.full_name}.git`

  const handleRepoCreated = (newRepo: RepositoryRecord) => {
    setRepos((prev) => [newRepo, ...prev])
    onSelectRepo(newRepo.name)
  }

  return (
    <>
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Left Column: Top Repositories (3 Cols) */}
        <aside className="lg:col-span-3 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xs font-semibold text-foreground tracking-tight uppercase">
              Top Repositories
            </h2>
            <Button
              size="xs"
              variant="primary"
              className="gap-1 text-xs"
              onClick={() => setIsCreateModalOpen(true)}
            >
              <Plus className="w-3.5 h-3.5" />
              <span>New</span>
            </Button>
          </div>

          {/* Search */}
          <div className="relative">
            <Search className="w-3.5 h-3.5 absolute left-2.5 top-2.5 text-muted-foreground" />
            <input
              type="text"
              placeholder="Find a repository..."
              value={repoSearch}
              onChange={(e) => setRepoSearch(e.target.value)}
              className="w-full pl-8 pr-2.5 py-1.5 text-xs rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
            />
          </div>

          {/* Repos list */}
          <div className="space-y-1">
            {filteredRepos.length > 0 ? (
              filteredRepos.map((repo) => (
                <div
                  key={repo.id || repo.full_name || repo.name}
                  onClick={() => onSelectRepo(repo.name)}
                  className="p-2 rounded-md hover:bg-secondary/60 transition-colors cursor-pointer group flex items-start gap-2.5"
                >
                  <div className="w-5 h-5 rounded-full bg-secondary border border-border flex items-center justify-center shrink-0 text-muted-foreground mt-0.5 group-hover:text-primary">
                    <BookOpen className="w-3 h-3" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-xs font-semibold text-foreground group-hover:text-primary group-hover:underline truncate">
                      {repo.full_name || repo.name}
                    </div>
                    {repo.description && (
                      <div className="text-[11px] text-muted-foreground truncate">
                        {repo.description}
                      </div>
                    )}
                  </div>
                  {repo.is_private && (
                    <Badge variant="outline" className="text-[9px] px-1 py-0 border-border text-muted-foreground">
                      Private
                    </Badge>
                  )}
                </div>
              ))
            ) : (
              <div className="p-4 rounded-md border border-dashed border-border text-center space-y-2">
                <p className="text-xs text-muted-foreground">No repositories found.</p>
                <Button
                  size="xs"
                  variant="primary"
                  onClick={() => setIsCreateModalOpen(true)}
                  className="text-xs"
                >
                  Create repository
                </Button>
              </div>
            )}
          </div>

          {/* Enterprise Organization Shortcut */}
          <div className="pt-4 border-t border-border space-y-2">
            <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Your Organizations
            </div>
            <div
              onClick={() => onSelectTab('profile')}
              className="flex items-center gap-2 p-1.5 rounded-md hover:bg-secondary/60 cursor-pointer"
            >
              <div className="w-6 h-6 rounded bg-primary text-primary-foreground font-bold text-xs flex items-center justify-center">
                E
              </div>
              <div className="text-xs font-medium text-foreground">exist-enterprise</div>
            </div>
          </div>
        </aside>

        {/* Center Feed: Activity Timeline / Onboarding Hero (6 Cols) */}
        <main className="lg:col-span-6 space-y-5">
          {/* Welcome / Onboarding Hero if in production with empty state */}
          {!isDemoModeActive() && repos.length === 0 && (
            <div className="p-5 rounded-lg border border-border bg-card space-y-4 shadow-sm">
              <div className="flex items-start gap-3">
                <div className="w-9 h-9 rounded-md bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0 mt-0.5">
                  <Compass className="w-5 h-5" />
                </div>
                <div>
                  <h2 className="text-sm font-semibold text-foreground">Welcome to your sovereign exist.io cluster</h2>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Your single-binary Git forge is running with SQLite WAL storage and MCP autonomous agent capabilities.
                  </p>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5 pt-1">
                <div
                  onClick={() => setIsCreateModalOpen(true)}
                  className="p-3 rounded-md border border-border bg-secondary/20 hover:bg-secondary/40 transition-colors cursor-pointer space-y-1"
                >
                  <div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
                    <Plus className="w-3.5 h-3.5 text-primary" />
                    <span>Create a new repository</span>
                  </div>
                  <p className="text-[11px] text-muted-foreground">
                    Host revision histories and collaborate securely.
                  </p>
                </div>

                {onOpenTour && (
                  <div
                    onClick={onOpenTour}
                    className="p-3 rounded-md border border-border bg-secondary/20 hover:bg-secondary/40 transition-colors cursor-pointer space-y-1"
                  >
                    <div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
                      <Compass className="w-3.5 h-3.5 text-primary" />
                      <span>Take interactive tour</span>
                    </div>
                    <p className="text-[11px] text-muted-foreground">
                      Learn how ports, protocols, and agents connect.
                    </p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Feed Header & Filters */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
            <h1 className="text-base font-bold text-foreground">Enterprise Activity Feed</h1>
            <div className="flex items-center gap-1.5 text-xs font-medium">
              <button
                onClick={() => setFilterType('all')}
                className={cn(
                  'px-2.5 py-1 rounded-md transition-colors cursor-pointer',
                  filterType === 'all'
                    ? 'bg-secondary text-foreground font-semibold'
                    : 'text-muted-foreground hover:bg-secondary/40'
                )}
              >
                All
              </button>
              <button
                onClick={() => setFilterType('agent')}
                className={cn(
                  'px-2.5 py-1 rounded-md transition-colors cursor-pointer flex items-center gap-1',
                  filterType === 'agent'
                    ? 'bg-secondary text-foreground font-semibold'
                    : 'text-muted-foreground hover:bg-secondary/40'
                )}
              >
                <Bot className="w-3 h-3 text-primary" />
                <span>Agents</span>
              </button>
              <button
                onClick={() => setFilterType('pr')}
                className={cn(
                  'px-2.5 py-1 rounded-md transition-colors cursor-pointer flex items-center gap-1',
                  filterType === 'pr'
                    ? 'bg-secondary text-foreground font-semibold'
                    : 'text-muted-foreground hover:bg-secondary/40'
                )}
              >
                <GitPullRequest className="w-3 h-3 text-emerald-400" />
                <span>Pull Requests</span>
              </button>
              <button
                onClick={() => setFilterType('commit')}
                className={cn(
                  'px-2.5 py-1 rounded-md transition-colors cursor-pointer flex items-center gap-1',
                  filterType === 'commit'
                    ? 'bg-secondary text-foreground font-semibold'
                    : 'text-muted-foreground hover:bg-secondary/40'
                )}
              >
                <GitCommit className="w-3 h-3 text-primary" />
                <span>Commits</span>
              </button>
            </div>
          </div>

          {/* Timeline Events Box */}
          <div className="space-y-3">
            {filteredEvents.length > 0 ? (
              filteredEvents.map((evt) => (
                <div
                  key={evt.id}
                  className="rounded-md border border-border bg-card p-4 space-y-2.5 hover:border-foreground/20 transition-colors shadow-2xs"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex items-center gap-2 text-xs">
                      {evt.type === 'agent' && (
                        <div className="w-6 h-6 rounded-md bg-primary/10 border border-primary/30 flex items-center justify-center text-primary">
                          <Bot className="w-3.5 h-3.5" />
                        </div>
                      )}
                      {evt.type === 'release' && (
                        <div className="w-6 h-6 rounded-md bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400">
                          <Tag className="w-3.5 h-3.5" />
                        </div>
                      )}
                      {evt.type === 'commit' && (
                        <div className="w-6 h-6 rounded-md bg-secondary border border-border flex items-center justify-center text-foreground">
                          <GitCommit className="w-3.5 h-3.5 text-primary" />
                        </div>
                      )}
                      {evt.type === 'ruleset' && (
                        <div className="w-6 h-6 rounded-md bg-primary/10 border border-primary/30 flex items-center justify-center text-primary">
                          <ShieldCheck className="w-3.5 h-3.5" />
                        </div>
                      )}
                      {evt.type === 'pr' && (
                        <div className="w-6 h-6 rounded-md bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
                          <GitPullRequest className="w-3.5 h-3.5" />
                        </div>
                      )}

                      <div className="leading-snug">
                        <strong className="text-foreground font-semibold hover:underline cursor-pointer">
                          {evt.actor}
                        </strong>{' '}
                        <span className="text-muted-foreground">{evt.action}</span>{' '}
                        <strong className="text-primary hover:underline cursor-pointer">
                          {evt.target}
                        </strong>
                      </div>
                    </div>

                    <span className="text-[11px] text-muted-foreground font-mono shrink-0">
                      {evt.timestamp}
                    </span>
                  </div>

                  {evt.details && (
                    <div className="text-xs text-muted-foreground bg-secondary/30 rounded-md p-2.5 border border-border/50 leading-relaxed">
                      {evt.details}
                    </div>
                  )}

                  {evt.sha && (
                    <div className="flex items-center gap-2 pt-1 font-mono text-[11px] text-muted-foreground">
                      <span className="px-1.5 py-0.5 rounded bg-secondary border border-border text-primary font-bold">
                        {evt.sha}
                      </span>
                      <span>Commit recorded in SQLite WAL ledger</span>
                    </div>
                  )}
                </div>
              ))
            ) : (
              <div className="p-6 rounded-lg border border-border bg-card space-y-4">
                <div className="flex items-center gap-2 text-xs font-semibold text-foreground">
                  <Terminal className="w-4 h-4 text-primary" />
                  <span>Getting Started Checklist</span>
                </div>
                <div className="space-y-2.5 text-xs">
                  <div className="flex items-center gap-2 text-emerald-500">
                    <Check className="w-4 h-4" />
                    <span className="text-foreground font-medium">Cluster daemon initialized (HTTP :8080 &amp; SSH :2222)</span>
                  </div>
                  <div className="flex items-center gap-2 text-muted-foreground">
                    <span className="w-4 h-4 rounded-full border border-border flex items-center justify-center text-[10px]">2</span>
                    <span>Create a new repository or push an existing Git repository</span>
                  </div>
                  <div className="flex items-center gap-2 text-muted-foreground">
                    <span className="w-4 h-4 rounded-full border border-border flex items-center justify-center text-[10px]">3</span>
                    <span>Enroll your SSH public key for sovereign push access</span>
                  </div>
                  <div className="flex items-center gap-2 text-muted-foreground">
                    <span className="w-4 h-4 rounded-full border border-border flex items-center justify-center text-[10px]">4</span>
                    <span>Connect your AI coding agent over MCP protocol at <code className="font-mono text-foreground">:8080/mcp</code></span>
                  </div>
                </div>
                {onOpenManual && (
                  <Button variant="secondary" size="xs" onClick={onOpenManual} className="gap-1 text-xs">
                    <BookOpen className="w-3.5 h-3.5" />
                    <span>View System Manual</span>
                  </Button>
                )}
              </div>
            )}
          </div>
        </main>

        {/* Right Sidebar: System Status & Quick Actions (3 Cols) */}
        <aside className="lg:col-span-3 space-y-5">
          {/* Daemon Status Card */}
          <Box>
            <BoxHeader>
              <BoxTitle className="flex items-center gap-2 text-xs">
                <Server className="w-4 h-4 text-primary" />
                <span>Cluster Telemetry</span>
              </BoxTitle>
            </BoxHeader>
            <BoxBody className="space-y-3 text-xs">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">existd Daemon:</span>
                <span className="flex items-center gap-1.5 text-emerald-400 font-medium">
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                  <span>Online (8080)</span>
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Git SSH Daemon:</span>
                <span className="text-foreground font-mono">Port 2222</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Storage Engine:</span>
                <span className="text-primary font-medium flex items-center gap-1">
                  <Activity className="w-3.5 h-3.5" />
                  <span>SQLite WAL</span>
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Mounted MCP Tools:</span>
                <span className="text-foreground font-mono font-bold">35</span>
              </div>
            </BoxBody>
          </Box>

          {/* Quick Clone Helper */}
          <div className="rounded-md border border-border bg-card p-4 space-y-2.5 shadow-2xs">
            <div className="text-xs font-semibold text-foreground">Quick Clone</div>
            <div className="flex items-center gap-1.5">
              <input
                type="text"
                readOnly
                value={`git clone ${quickCloneUrl}`}
                className="flex-1 font-mono text-[11px] bg-secondary/50 border border-border rounded px-2.5 py-1 text-foreground"
              />
              <Button
                variant="secondary"
                size="xs"
                onClick={() => {
                  navigator.clipboard.writeText(`git clone ${quickCloneUrl}`)
                  setCopiedClone(true)
                  setTimeout(() => setCopiedClone(false), 2000)
                }}
                className="shrink-0"
              >
                {copiedClone ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              </Button>
            </div>
          </div>

          {/* User Guide & Tour Shortcuts */}
          {(onOpenTour || onOpenManual) && (
            <div className="rounded-md border border-border bg-card p-4 space-y-2.5 shadow-2xs">
              <div className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                <Compass className="w-3.5 h-3.5 text-primary" />
                <span>Cluster Navigation</span>
              </div>
              <p className="text-[11px] text-muted-foreground">
                First time using exist.io? Explore how everything is connected.
              </p>
              <div className="flex items-center gap-2 pt-1">
                {onOpenTour && (
                  <Button variant="secondary" size="xs" onClick={onOpenTour} className="flex-1 text-xs">
                    Start Tour
                  </Button>
                )}
                {onOpenManual && (
                  <Button variant="secondary" size="xs" onClick={onOpenManual} className="flex-1 text-xs">
                    User Manual
                  </Button>
                )}
              </div>
            </div>
          )}

          {/* Autonomous Agents Spotlight */}
          <div className="rounded-md border border-border bg-card p-4 space-y-3 shadow-2xs">
            <div className="flex items-center justify-between">
              <div className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                <Bot className="w-3.5 h-3.5 text-primary" />
                <span>Autonomous Agents</span>
              </div>
              <Badge variant="outline" className="text-[10px] border-primary/40 text-primary bg-primary/10">
                3 Active
              </Badge>
            </div>
            <p className="text-xs text-muted-foreground leading-relaxed">
              Native AI agents autonomously triage pull requests, enforce cryptographic rulesets, and report telemetry over Model Context Protocol.
            </p>
            <Button
              variant="secondary"
              size="xs"
              onClick={() => onSelectTab('agents')}
              className="w-full text-xs font-medium"
            >
              Open Agent Console
            </Button>
          </div>
        </aside>
      </div>

      {/* Create Repo Modal */}
      <CreateRepoModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSuccess={handleRepoCreated}
      />
    </>
  )
}
