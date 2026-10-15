import React, { useState } from 'react'
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
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Box, BoxHeader, BoxTitle, BoxBody } from '@/components/ui/Card'
import { cn } from '@/lib/utils'

export interface GlobalHomeViewProps {
  onSelectRepo: (repoName: string) => void
  onSelectTab: (tab: string) => void
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
}) => {
  const [filterType, setFilterType] = useState<'all' | 'agent' | 'pr' | 'commit'>('all')
  const [repoSearch, setRepoSearch] = useState('')
  const [copiedClone, setCopiedClone] = useState(false)

  const recentRepos = [
    { name: 'shiva/exist.io', isPrivate: false, description: 'Sovereign GitHub Enterprise alternative with MCP' },
    { name: 'shiva/agentic-mcp-bridge', isPrivate: false, description: '35+ Git operations over Model Context Protocol' },
    { name: 'shiva/sqlite-git-wal', isPrivate: false, description: 'Transactional ref storage and cryptographic audit ledger' },
    { name: 'shiva/iac-terraform-enterprise', isPrivate: false, description: 'Air-gapped multi-cloud IaC provisioning' },
    { name: 'shiva/enterprise-auth-scim', isPrivate: true, description: 'SAML 2.0 and SCIM 2.0 enterprise identity sync' },
  ]

  const filteredRepos = recentRepos.filter((r) =>
    r.name.toLowerCase().includes(repoSearch.toLowerCase())
  )

  const filteredEvents = enterpriseFeed.filter((evt) => {
    if (filterType === 'all') return true
    if (filterType === 'agent') return evt.type === 'agent' || evt.type === 'ruleset'
    if (filterType === 'pr') return evt.type === 'pr'
    if (filterType === 'commit') return evt.type === 'commit'
    return true
  })

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
      {/* Left Column: Top Repositories (3 Cols) */}
      <aside className="lg:col-span-3 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xs font-semibold text-foreground tracking-tight uppercase">
            Top Repositories
          </h2>
          <Button size="xs" className="gap-1 text-xs">
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
          {filteredRepos.map((repo) => (
            <div
              key={repo.name}
              onClick={() => onSelectRepo(repo.name.split('/')[1] || repo.name)}
              className="p-2 rounded-md hover:bg-secondary/60 transition-colors cursor-pointer group flex items-start gap-2.5"
            >
              <div className="w-5 h-5 rounded-full bg-secondary border border-border flex items-center justify-center shrink-0 text-muted-foreground mt-0.5 group-hover:text-primary">
                <BookOpen className="w-3 h-3" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-xs font-semibold text-foreground group-hover:text-primary group-hover:underline truncate">
                  {repo.name}
                </div>
                <div className="text-[11px] text-muted-foreground truncate">
                  {repo.description}
                </div>
              </div>
            </div>
          ))}
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

      {/* Center Feed: Activity Timeline (6 Cols) */}
      <main className="lg:col-span-6 space-y-5">
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
          {filteredEvents.map((evt) => (
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
          ))}
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
              value="git clone http://localhost:8080/shiva/exist.io.git"
              className="flex-1 font-mono text-[11px] bg-secondary/50 border border-border rounded px-2.5 py-1 text-foreground"
            />
            <Button
              variant="secondary"
              size="xs"
              onClick={() => {
                navigator.clipboard.writeText('git clone http://localhost:8080/shiva/exist.io.git')
                setCopiedClone(true)
                setTimeout(() => setCopiedClone(false), 2000)
              }}
              className="shrink-0"
            >
              {copiedClone ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
            </Button>
          </div>
        </div>

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
  )
}
