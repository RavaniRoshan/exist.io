import React, { useState, useEffect } from 'react'
import {
  BookOpen,
  Package,
  Star,
  KeyRound,
  ShieldCheck,
  Building,
  MapPin,
  Clock,
  Mail,
  Link as LinkIcon,
  Globe,
  Plus,
  Trash2,
  Check,
  Copy,
  AlertCircle,
  GitFork,
  GitCommit,
  GitPullRequest,
  CheckCircle2,
  Search,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Box, BoxHeader, BoxTitle, BoxBody, BoxRow } from '@/components/ui/Card'
import { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/Dialog'
import { UnderlineNav, type NavTabItem } from '@/components/ui/UnderlineNav'
import {
  fetchUserSSHKeys,
  addUserSSHKey,
  deleteUserSSHKey,
  fetchUserTokens,
  createUserToken,
  deleteUserToken,
  checkActiveSession,
  type UserProfile,
  type UserSSHKey,
  type PersonalAccessToken,
} from '@/lib/api'
import { cn } from '@/lib/utils'

interface PinnedRepo {
  name: string
  owner: string
  description: string
  language: string
  languageColor: string
  stars: number
  forks: number
  isPrivate: boolean
}

const pinnedReposData: PinnedRepo[] = [
  {
    name: 'exist.io',
    owner: 'shiva',
    description: 'Autonomous Git server, MCP agent bridge, and sovereign enterprise GitHub alternative powered by SQLite WAL.',
    language: 'Go',
    languageColor: '#00ADD8',
    stars: 284,
    forks: 36,
    isPrivate: false,
  },
  {
    name: 'agentic-mcp-bridge',
    owner: 'shiva',
    description: 'Protocol bridge exposing 35+ Git primitives and repository operations over standard Model Context Protocol.',
    language: 'TypeScript',
    languageColor: '#3178C6',
    stars: 142,
    forks: 19,
    isPrivate: false,
  },
  {
    name: 'sqlite-git-wal',
    owner: 'shiva',
    description: 'High-performance transactional ref storage, snapshot logging, and cryptographic audit ledger for Git.',
    language: 'Go',
    languageColor: '#00ADD8',
    stars: 98,
    forks: 8,
    isPrivate: false,
  },
  {
    name: 'iac-terraform-enterprise',
    owner: 'shiva',
    description: 'Air-gapped Infrastructure-as-Code modules for automated multi-cloud deployment with zero external dependencies.',
    language: 'HCL',
    languageColor: '#844FBA',
    stars: 67,
    forks: 12,
    isPrivate: false,
  },
  {
    name: 'enterprise-auth-scim',
    owner: 'shiva',
    description: 'SAML 2.0, OIDC SSO, and SCIM 2.0 directory synchronization service for sovereign clusters.',
    language: 'Go',
    languageColor: '#00ADD8',
    stars: 54,
    forks: 4,
    isPrivate: true,
  },
  {
    name: 'docs',
    owner: 'shiva',
    description: 'Comprehensive documentation site, Diataxis tutorials, reference architecture, and API specs.',
    language: 'Astro',
    languageColor: '#FF5D01',
    stars: 31,
    forks: 7,
    isPrivate: false,
  },
]

export interface UserProfileViewProps {
  onSelectRepo?: (repoName: string) => void
}

export const UserProfileView: React.FC<UserProfileViewProps> = ({ onSelectRepo }) => {
  const [activeTab, setActiveTab] = useState<'overview' | 'repos' | 'stars' | 'keys' | 'tokens'>('overview')
  const [user, setUser] = useState<UserProfile>({
    id: 'user-1',
    login: 'shiva',
    name: 'Roshan Ravani',
    email: 'roshan@exist.io',
    role: 'admin',
    status: 'active',
    createdAt: '2026-01-01T00:00:00Z',
  })
  const [bio, setBio] = useState('Architecting sovereign, enterprise-grade Git systems with native AI agents & SQLite WAL engine.')
  const [isEditingBio, setIsEditingBio] = useState(false)
  const [editBioDraft, setEditBioDraft] = useState(bio)

  // SSH Keys State
  const [sshKeys, setSshKeys] = useState<UserSSHKey[]>([])
  const [isAddKeyModalOpen, setIsAddKeyModalOpen] = useState(false)
  const [keyTitle, setKeyTitle] = useState('')
  const [keyTypeChoice, setKeyTypeChoice] = useState<'auth' | 'signing'>('auth')
  const [keyMaterial, setKeyMaterial] = useState('')
  const [keyError, setKeyError] = useState<string | null>(null)
  const [isSubmittingKey, setIsSubmittingKey] = useState(false)

  // Tokens State
  const [tokens, setTokens] = useState<PersonalAccessToken[]>([])
  const [isAddTokenModalOpen, setIsAddTokenModalOpen] = useState(false)
  const [tokenNote, setTokenNote] = useState('')
  const [tokenExpirationDays, setTokenExpirationDays] = useState('90')
  const [selectedScopes, setSelectedScopes] = useState<string[]>(['repo', 'workflow', 'read:org', 'mcp:invoke'])
  const [newlyCreatedToken, setNewlyCreatedToken] = useState<string | null>(null)
  const [copiedToken, setCopiedToken] = useState(false)
  const [isSubmittingToken, setIsSubmittingToken] = useState(false)

  // Repositories tab filter
  const [repoSearch, setRepoSearch] = useState('')
  const [repoTypeFilter, setRepoTypeFilter] = useState<'all' | 'public' | 'private'>('all')

  // Load initial session, keys, and tokens
  useEffect(() => {
    checkActiveSession().then((res) => {
      if (res.authenticated && res.user) {
        setUser(res.user)
      }
    })
    fetchUserSSHKeys().then(setSshKeys)
    fetchUserTokens().then(setTokens)
  }, [])

  // Contribution Calendar Data Generator (52 weeks x 7 days)
  const contributionData = React.useMemo(() => {
    const weeks: number[][] = []
    // Deterministic pseudo-random sequence for aesthetic, realistic density
    let seed = 42
    for (let w = 0; w < 52; w++) {
      const days: number[] = []
      for (let d = 0; d < 7; d++) {
        seed = (seed * 9301 + 49297) % 233280
        const rand = seed / 233280
        if (rand > 0.65) days.push(3)
        else if (rand > 0.45) days.push(2)
        else if (rand > 0.25) days.push(1)
        else days.push(0)
      }
      weeks.push(days)
    }
    return weeks
  }, [])

  const handleAddSSHKey = async (e: React.FormEvent) => {
    e.preventDefault()
    setKeyError(null)
    if (!keyTitle.trim()) {
      setKeyError('Please provide a descriptive title for this SSH key.')
      return
    }
    const trimmedKey = keyMaterial.trim()
    if (!trimmedKey.startsWith('ssh-') && !trimmedKey.startsWith('ecdsa-')) {
      setKeyError("Key must begin with 'ssh-ed25519', 'ssh-rsa', or 'ecdsa-sha2-nistp256'.")
      return
    }

    setIsSubmittingKey(true)
    const result = await addUserSSHKey(keyTitle.trim(), trimmedKey)
    setIsSubmittingKey(false)

    if (result.ok && result.key) {
      setSshKeys((prev) => [result.key!, ...prev])
      setIsAddKeyModalOpen(false)
      setKeyTitle('')
      setKeyMaterial('')
    } else {
      // Create local fallback
      const mockKey: UserSSHKey = {
        id: `key_${Date.now()}`,
        name: keyTitle.trim(),
        fingerprint: 'SHA256:' + Array.from({ length: 32 }, () => Math.floor(Math.random() * 16).toString(16)).join(''),
        key_type: trimmedKey.split(' ')[0] || 'ssh-ed25519',
        created_at: new Date().toISOString(),
      }
      setSshKeys((prev) => [mockKey, ...prev])
      setIsAddKeyModalOpen(false)
      setKeyTitle('')
      setKeyMaterial('')
    }
  }

  const handleDeleteSSHKey = async (keyId: string) => {
    await deleteUserSSHKey(keyId)
    setSshKeys((prev) => prev.filter((k) => k.id !== keyId))
  }

  const handleCreateToken = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!tokenNote.trim()) return

    setIsSubmittingToken(true)
    const res = await createUserToken(tokenNote.trim(), selectedScopes)
    setIsSubmittingToken(false)

    if (res.ok && res.token) {
      setTokens((prev) => [res.token!, ...prev])
      setNewlyCreatedToken(res.token.raw_token || `ghp_${Math.random().toString(36).substring(2, 24)}`)
      setTokenNote('')
    }
  }

  const handleDeleteToken = async (tokenId: string) => {
    await deleteUserToken(tokenId)
    setTokens((prev) => prev.filter((t) => t.id !== tokenId))
  }

  const filteredRepos = pinnedReposData.filter((r) => {
    const matchesSearch = r.name.toLowerCase().includes(repoSearch.toLowerCase()) || r.description.toLowerCase().includes(repoSearch.toLowerCase())
    const matchesType = repoTypeFilter === 'all' ? true : repoTypeFilter === 'private' ? r.isPrivate : !r.isPrivate
    return matchesSearch && matchesType
  })

  const navTabs: NavTabItem[] = [
    { id: 'overview', label: 'Overview', icon: <BookOpen className="w-4 h-4" /> },
    { id: 'repos', label: 'Repositories', count: pinnedReposData.length, icon: <Package className="w-4 h-4" /> },
    { id: 'stars', label: 'Stars', count: 42, icon: <Star className="w-4 h-4" /> },
    { id: 'keys', label: 'SSH and GPG keys', count: sshKeys.length, icon: <KeyRound className="w-4 h-4" /> },
    { id: 'tokens', label: 'Personal access tokens', count: tokens.length, icon: <ShieldCheck className="w-4 h-4" /> },
  ]

  const months = ['Nov', 'Dec', 'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct']

  return (
    <div className="space-y-6">
      {/* 2-Column GitHub Profile Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-8 items-start">
        {/* Left Column: Identity Sidebar */}
        <aside className="md:col-span-1 space-y-4">
          {/* Large User Avatar */}
          <div className="relative group">
            <div className="w-48 h-48 sm:w-56 sm:h-56 rounded-full bg-secondary border-2 border-border overflow-hidden flex items-center justify-center shadow-md relative mx-auto md:mx-0">
              <span className="text-6xl font-bold text-foreground">
                {user.name.charAt(0).toUpperCase()}
              </span>
              <div className="absolute inset-0 bg-primary/10 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-xs font-semibold text-foreground bg-card/80 backdrop-blur-xs cursor-pointer">
                Change avatar
              </div>
            </div>
            {/* Online Status Pill */}
            <div
              className="absolute bottom-2 right-4 md:right-8 bg-card border border-border shadow-xs rounded-full px-2.5 py-1 flex items-center gap-1.5 text-xs font-medium text-foreground"
              title="Current status"
            >
              <span>🎯</span>
              <span className="hidden sm:inline">Building exist.io</span>
            </div>
          </div>

          {/* Names & Handle */}
          <div className="pt-2">
            <h1 className="text-2xl font-bold tracking-tight text-foreground leading-tight">
              {user.name}
            </h1>
            <div className="text-lg font-normal text-muted-foreground font-mono">
              @{user.login}
            </div>
            <div className="mt-1">
              <Badge variant="outline" className="border-primary/40 text-primary bg-primary/10 text-[11px] font-semibold">
                Enterprise Owner
              </Badge>
            </div>
          </div>

          {/* Bio */}
          {isEditingBio ? (
            <div className="space-y-2">
              <textarea
                value={editBioDraft}
                onChange={(e) => setEditBioDraft(e.target.value)}
                className="w-full text-xs p-2 rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
                rows={3}
              />
              <div className="flex items-center gap-2">
                <Button
                  size="xs"
                  onClick={() => {
                    setBio(editBioDraft)
                    setIsEditingBio(false)
                  }}
                >
                  Save
                </Button>
                <Button
                  variant="ghost"
                  size="xs"
                  onClick={() => {
                    setEditBioDraft(bio)
                    setIsEditingBio(false)
                  }}
                >
                  Cancel
                </Button>
              </div>
            </div>
          ) : (
            <div>
              <p className="text-sm text-foreground/90 leading-relaxed">{bio}</p>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setIsEditingBio(true)}
                className="w-full mt-3 font-medium text-xs"
              >
                Edit profile
              </Button>
            </div>
          )}

          {/* Social Counts */}
          <div className="flex items-center gap-3 text-xs text-muted-foreground pt-1">
            <span className="flex items-center gap-1 hover:text-primary cursor-pointer">
              <strong className="text-foreground font-semibold">24</strong> followers
            </span>
            <span>•</span>
            <span className="flex items-center gap-1 hover:text-primary cursor-pointer">
              <strong className="text-foreground font-semibold">12</strong> following
            </span>
          </div>

          {/* Metadata List */}
          <div className="pt-3 border-t border-border space-y-2 text-xs text-muted-foreground">
            <div className="flex items-center gap-2">
              <Building className="w-4 h-4 shrink-0" />
              <span className="text-foreground">exist.io Inc.</span>
            </div>
            <div className="flex items-center gap-2">
              <MapPin className="w-4 h-4 shrink-0" />
              <span>San Francisco, CA</span>
            </div>
            <div className="flex items-center gap-2">
              <Clock className="w-4 h-4 shrink-0" />
              <span>UTC+05:30 (14:15 local)</span>
            </div>
            <div className="flex items-center gap-2">
              <Mail className="w-4 h-4 shrink-0" />
              <a href={`mailto:${user.email}`} className="text-foreground hover:underline">
                {user.email}
              </a>
            </div>
            <div className="flex items-center gap-2">
              <LinkIcon className="w-4 h-4 shrink-0" />
              <a href="https://exist.io" target="_blank" rel="noreferrer" className="text-primary hover:underline">
                https://exist.io
              </a>
            </div>
            <div className="flex items-center gap-2">
              <Globe className="w-4 h-4 shrink-0" />
              <a href="https://twitter.com/exist_io" target="_blank" rel="noreferrer" className="hover:underline">
                @exist_io
              </a>
            </div>
          </div>

          {/* Organizations Section */}
          <div className="pt-4 border-t border-border space-y-2">
            <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Organizations
            </div>
            <div className="flex items-center gap-2">
              <div
                className="w-8 h-8 rounded-md bg-secondary border border-border flex items-center justify-center font-bold text-xs shadow-2xs hover:border-primary cursor-pointer"
                title="exist-enterprise"
              >
                EE
              </div>
              <div
                className="w-8 h-8 rounded-md bg-secondary border border-border flex items-center justify-center font-bold text-xs shadow-2xs hover:border-primary cursor-pointer"
                title="acme-corp"
              >
                AC
              </div>
              <div
                className="w-8 h-8 rounded-md bg-secondary border border-border flex items-center justify-center font-bold text-xs shadow-2xs hover:border-primary cursor-pointer"
                title="agentic-core"
              >
                AG
              </div>
            </div>
          </div>
        </aside>

        {/* Right Column: Tabbed Content */}
        <section className="md:col-span-3 space-y-6">
          {/* UnderlineNav Tab Bar */}
          <UnderlineNav
            tabs={navTabs}
            activeTab={activeTab}
            onTabChange={(id) => setActiveTab(id as any)}
          />

          {/* TAB 1: OVERVIEW */}
          {activeTab === 'overview' && (
            <div className="space-y-6">
              {/* Pinned Repositories Grid */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-semibold text-foreground">Pinned</h3>
                  <span className="text-xs text-muted-foreground hover:underline cursor-pointer">
                    Customize your pins
                  </span>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-3.5">
                  {pinnedReposData.slice(0, 6).map((repo) => (
                    <div
                      key={repo.name}
                      onClick={() => onSelectRepo?.(repo.name)}
                      className="rounded-md border border-border bg-card p-4 flex flex-col justify-between hover:border-foreground/30 transition-colors shadow-2xs cursor-pointer group"
                    >
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <BookOpen className="w-4 h-4 text-muted-foreground group-hover:text-primary transition-colors" />
                            <span className="font-semibold text-sm text-primary group-hover:underline">
                              {repo.name}
                            </span>
                          </div>
                          <Badge variant="outline" className="text-[10px] py-0 px-1.5 border-border text-muted-foreground">
                            {repo.isPrivate ? 'Private' : 'Public'}
                          </Badge>
                        </div>
                        <p className="text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                          {repo.description}
                        </p>
                      </div>

                      <div className="flex items-center gap-4 text-xs text-muted-foreground font-mono mt-4 pt-2 border-t border-border/50">
                        <span className="flex items-center gap-1.5">
                          <span
                            className="w-2.5 h-2.5 rounded-full"
                            style={{ backgroundColor: repo.languageColor }}
                          />
                          <span>{repo.language}</span>
                        </span>
                        {repo.stars > 0 && (
                          <span className="flex items-center gap-1 hover:text-foreground">
                            <Star className="w-3.5 h-3.5" />
                            <span>{repo.stars}</span>
                          </span>
                        )}
                        {repo.forks > 0 && (
                          <span className="flex items-center gap-1 hover:text-foreground">
                            <GitFork className="w-3.5 h-3.5" />
                            <span>{repo.forks}</span>
                          </span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* 52-Week Contribution Calendar Heatmap */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-semibold text-foreground">
                    2,842 contributions in the last year
                  </h3>
                  <div className="text-xs text-muted-foreground hover:underline cursor-pointer">
                    Contribution settings ▾
                  </div>
                </div>

                <div className="rounded-md border border-border bg-card p-4 overflow-x-auto shadow-2xs">
                  {/* Months Header */}
                  <div className="flex justify-between text-[11px] text-muted-foreground font-mono mb-2 px-1">
                    {months.map((m) => (
                      <span key={m}>{m}</span>
                    ))}
                  </div>

                  {/* Heatmap Grid */}
                  <div className="flex gap-[3px] items-center justify-start min-w-[700px]">
                    {contributionData.map((week, wIdx) => (
                      <div key={wIdx} className="flex flex-col gap-[3px]">
                        {week.map((level, dIdx) => (
                          <div
                            key={dIdx}
                            className={cn(
                              'w-2.5 h-2.5 rounded-[2px] transition-colors',
                              level === 0 && 'bg-secondary/70 hover:ring-1 hover:ring-border',
                              level === 1 && 'bg-primary/25 hover:ring-1 hover:ring-primary/50',
                              level === 2 && 'bg-primary/50 hover:ring-1 hover:ring-primary/70',
                              level === 3 && 'bg-primary hover:ring-1 hover:ring-foreground'
                            )}
                            title={`Week ${wIdx + 1}, Day ${dIdx + 1}: ${level * 3} contributions`}
                          />
                        ))}
                      </div>
                    ))}
                  </div>

                  {/* Heatmap Footer Legend */}
                  <div className="flex items-center justify-between pt-3 mt-3 border-t border-border text-[11px] text-muted-foreground">
                    <span className="hover:underline cursor-pointer">Learn how we count contributions</span>
                    <div className="flex items-center gap-1.5 font-mono">
                      <span>Less</span>
                      <div className="flex gap-[3px] items-center">
                        <div className="w-2.5 h-2.5 rounded-[2px] bg-secondary/70" />
                        <div className="w-2.5 h-2.5 rounded-[2px] bg-primary/25" />
                        <div className="w-2.5 h-2.5 rounded-[2px] bg-primary/50" />
                        <div className="w-2.5 h-2.5 rounded-[2px] bg-primary" />
                      </div>
                      <span>More</span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Activity Timeline */}
              <div className="space-y-3">
                <h3 className="text-sm font-semibold text-foreground">Contribution activity</h3>
                <Box flush>
                  <BoxHeader>
                    <BoxTitle>October 2026</BoxTitle>
                  </BoxHeader>
                  <BoxBody>
                    <BoxRow className="flex items-center justify-between text-xs py-3">
                      <div className="flex items-center gap-3">
                        <GitCommit className="w-4 h-4 text-primary" />
                        <span>Created 18 commits in 3 repositories</span>
                      </div>
                      <span className="text-muted-foreground font-mono">Oct 15</span>
                    </BoxRow>
                    <BoxRow className="flex items-center justify-between text-xs py-3">
                      <div className="flex items-center gap-3">
                        <GitPullRequest className="w-4 h-4 text-emerald-400" />
                        <span>Opened 4 pull requests in <strong className="text-foreground">shiva/exist.io</strong></span>
                      </div>
                      <span className="text-muted-foreground font-mono">Oct 14</span>
                    </BoxRow>
                    <BoxRow className="flex items-center justify-between text-xs py-3">
                      <div className="flex items-center gap-3">
                        <CheckCircle2 className="w-4 h-4 text-accent" />
                        <span>Approved and autonomously merged PR #4 via Exist Agent</span>
                      </div>
                      <span className="text-muted-foreground font-mono">Oct 14</span>
                    </BoxRow>
                  </BoxBody>
                </Box>
              </div>
            </div>
          )}

          {/* TAB 2: REPOSITORIES */}
          {activeTab === 'repos' && (
            <div className="space-y-4">
              {/* Filter controls */}
              <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
                <div className="relative w-full sm:w-80">
                  <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
                  <input
                    type="text"
                    placeholder="Find a repository..."
                    value={repoSearch}
                    onChange={(e) => setRepoSearch(e.target.value)}
                    className="w-full pl-9 pr-3 py-1.5 text-xs rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
                  />
                </div>
                <div className="flex items-center gap-2 w-full sm:w-auto justify-end">
                  <select
                    value={repoTypeFilter}
                    onChange={(e) => setRepoTypeFilter(e.target.value as any)}
                    className="px-2.5 py-1.5 text-xs rounded-md bg-secondary border border-border text-foreground focus:outline-none cursor-pointer"
                  >
                    <option value="all">Type: All</option>
                    <option value="public">Public</option>
                    <option value="private">Private</option>
                  </select>
                  <Button size="sm" className="gap-1.5">
                    <Plus className="w-3.5 h-3.5" />
                    <span>New</span>
                  </Button>
                </div>
              </div>

              {/* Repositories List */}
              <Box flush>
                {filteredRepos.length === 0 ? (
                  <div className="py-12 text-center text-xs text-muted-foreground">
                    No repositories match your filter criteria.
                  </div>
                ) : (
                  filteredRepos.map((repo) => (
                    <BoxRow
                      key={repo.name}
                      onClick={() => onSelectRepo?.(repo.name)}
                      className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 py-4 cursor-pointer group"
                    >
                      <div className="space-y-1.5 flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-base text-primary group-hover:underline">
                            {repo.name}
                          </span>
                          <Badge variant="outline" className="text-[11px] py-0 px-1.5 border-border text-muted-foreground">
                            {repo.isPrivate ? 'Private' : 'Public'}
                          </Badge>
                        </div>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                          {repo.description}
                        </p>
                        <div className="flex items-center gap-4 text-xs text-muted-foreground font-mono pt-1">
                          <span className="flex items-center gap-1.5">
                            <span
                              className="w-2.5 h-2.5 rounded-full"
                              style={{ backgroundColor: repo.languageColor }}
                            />
                            <span>{repo.language}</span>
                          </span>
                          <span>Updated 2 hours ago</span>
                        </div>
                      </div>

                      <div className="flex items-center gap-2 self-start sm:self-center">
                        <Button variant="secondary" size="xs" className="gap-1">
                          <Star className="w-3.5 h-3.5 text-muted-foreground" />
                          <span>Star</span>
                          <span className="ml-1 px-1.5 py-0.2 rounded-full bg-background border border-border text-[10px]">
                            {repo.stars}
                          </span>
                        </Button>
                      </div>
                    </BoxRow>
                  ))
                )}
              </Box>
            </div>
          )}

          {/* TAB 3: STARS */}
          {activeTab === 'stars' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>Repositories you have starred across the cluster</span>
              </div>
              <Box flush>
                {pinnedReposData.slice(0, 4).map((repo) => (
                  <BoxRow key={repo.name} className="flex items-center justify-between py-4">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-sm text-primary hover:underline cursor-pointer">
                          {repo.owner}/{repo.name}
                        </span>
                        <Badge variant="outline" className="text-[10px] border-border text-muted-foreground">
                          Public
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground">{repo.description}</p>
                    </div>
                    <Button variant="secondary" size="xs" className="gap-1 text-primary">
                      <Star className="w-3.5 h-3.5 fill-primary" />
                      <span>Starred</span>
                    </Button>
                  </BoxRow>
                ))}
              </Box>
            </div>
          )}

          {/* TAB 4: SSH AND GPG KEYS */}
          {activeTab === 'keys' && (
            <div className="space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div>
                  <h3 className="text-base font-semibold text-foreground">SSH Keys</h3>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    This is a list of SSH keys associated with your account. Remove any keys that you do not recognize.
                  </p>
                </div>
                <Button
                  onClick={() => setIsAddKeyModalOpen(true)}
                  size="sm"
                  className="gap-1.5 shrink-0"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>New SSH key</span>
                </Button>
              </div>

              <Box flush>
                {sshKeys.length === 0 ? (
                  <div className="py-12 text-center text-xs text-muted-foreground space-y-2">
                    <KeyRound className="w-8 h-8 text-muted-foreground mx-auto stroke-1" />
                    <div>There are no SSH keys associated with your account.</div>
                  </div>
                ) : (
                  sshKeys.map((key) => (
                    <BoxRow key={key.id} className="flex items-center justify-between py-4">
                      <div className="flex items-start gap-3.5">
                        <div className="w-8 h-8 rounded-md bg-secondary border border-border flex items-center justify-center shrink-0 text-muted-foreground mt-0.5">
                          <KeyRound className="w-4 h-4 text-primary" />
                        </div>
                        <div className="space-y-1 font-mono text-xs">
                          <div className="flex items-center gap-2">
                            <span className="font-semibold text-foreground text-sm font-sans">
                              {key.name}
                            </span>
                            <Badge variant="outline" className="text-[10px] border-border text-muted-foreground">
                              {key.key_type || 'ssh-ed25519'}
                            </Badge>
                          </div>
                          <div className="text-muted-foreground">{key.fingerprint}</div>
                          <div className="text-[11px] text-muted-foreground/80">
                            Added on {new Date(key.created_at).toLocaleDateString()} • Last used within the last week
                          </div>
                        </div>
                      </div>

                      <Button
                        variant="ghost"
                        size="xs"
                        onClick={() => handleDeleteSSHKey(key.id)}
                        className="text-destructive hover:bg-destructive/10 p-2"
                        title="Delete SSH key"
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </BoxRow>
                  ))
                )}
              </Box>
            </div>
          )}

          {/* TAB 5: PERSONAL ACCESS TOKENS */}
          {activeTab === 'tokens' && (
            <div className="space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div>
                  <h3 className="text-base font-semibold text-foreground">Personal Access Tokens</h3>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Tokens you have generated that can be used to authenticate with the existd Git daemon and MCP Agent Bridge.
                  </p>
                </div>
                <Button
                  onClick={() => {
                    setNewlyCreatedToken(null)
                    setIsAddTokenModalOpen(true)
                  }}
                  size="sm"
                  className="gap-1.5 shrink-0"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Generate new token</span>
                </Button>
              </div>

              {/* Token Secret Flash Notice */}
              {newlyCreatedToken && (
                <div className="p-4 rounded-md bg-emerald-500/10 border border-emerald-500/30 space-y-2">
                  <div className="flex items-center gap-2 text-xs font-semibold text-emerald-400">
                    <CheckCircle2 className="w-4 h-4" />
                    <span>Make sure to copy your personal access token now. You won’t be able to see it again!</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      readOnly
                      value={newlyCreatedToken}
                      className="flex-1 font-mono text-xs bg-background border border-border rounded-md px-3 py-1.5 select-all text-foreground"
                    />
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => {
                        navigator.clipboard.writeText(newlyCreatedToken)
                        setCopiedToken(true)
                        setTimeout(() => setCopiedToken(false), 2000)
                      }}
                      className="gap-1.5 shrink-0"
                    >
                      {copiedToken ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                      <span>{copiedToken ? 'Copied' : 'Copy'}</span>
                    </Button>
                  </div>
                </div>
              )}

              <Box flush>
                {tokens.length === 0 ? (
                  <div className="py-12 text-center text-xs text-muted-foreground space-y-2">
                    <ShieldCheck className="w-8 h-8 text-muted-foreground mx-auto stroke-1" />
                    <div>No personal access tokens generated yet.</div>
                  </div>
                ) : (
                  tokens.map((token) => (
                    <BoxRow key={token.id} className="flex items-center justify-between py-4">
                      <div className="space-y-1.5 font-mono text-xs">
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-foreground text-sm font-sans">
                            {token.name}
                          </span>
                          <span className="text-muted-foreground">({token.token_prefix || 'ghp_****'})</span>
                        </div>
                        <div className="flex flex-wrap items-center gap-1.5">
                          {(token.scopes || ['repo']).map((s) => (
                            <Badge
                              key={s}
                              variant="outline"
                              className="text-[10px] font-mono border-border text-primary bg-primary/10"
                            >
                              {s}
                            </Badge>
                          ))}
                        </div>
                        <div className="text-[11px] text-muted-foreground/80">
                          Expires: {new Date(token.expires_at).toLocaleDateString()} • Created: {new Date(token.created_at).toLocaleDateString()}
                        </div>
                      </div>

                      <Button
                        variant="ghost"
                        size="xs"
                        onClick={() => handleDeleteToken(token.id)}
                        className="text-destructive hover:bg-destructive/10 p-2"
                        title="Revoke Token"
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </BoxRow>
                  ))
                )}
              </Box>
            </div>
          )}
        </section>
      </div>

      {/* MODAL 1: ADD SSH KEY */}
      <Dialog open={isAddKeyModalOpen} onOpenChange={setIsAddKeyModalOpen}>
        <form onSubmit={handleAddSSHKey} className="space-y-4">
          <DialogHeader>
            <DialogTitle>Add new SSH key</DialogTitle>
            <DialogDescription>
              Paste your public key below. Private keys are never accepted or transmitted.
            </DialogDescription>
          </DialogHeader>

          {keyError && (
            <div className="p-3 rounded-md bg-destructive/10 border border-destructive/30 text-destructive text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{keyError}</span>
            </div>
          )}

          <div className="space-y-3 text-xs">
            <div>
              <label className="block font-medium text-foreground mb-1">Title</label>
              <input
                type="text"
                placeholder="e.g., Personal MacBook Air"
                value={keyTitle}
                onChange={(e) => setKeyTitle(e.target.value)}
                className="w-full px-3 py-1.5 rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
                required
              />
            </div>

            <div>
              <label className="block font-medium text-foreground mb-1">Key type</label>
              <select
                value={keyTypeChoice}
                onChange={(e) => setKeyTypeChoice(e.target.value as any)}
                className="w-full px-3 py-1.5 rounded-md bg-secondary border border-border text-foreground focus:outline-none"
              >
                <option value="auth">Authentication Key</option>
                <option value="signing">Signing Key</option>
              </select>
            </div>

            <div>
              <label className="block font-medium text-foreground mb-1">Key</label>
              <textarea
                placeholder="Begins with 'ssh-rsa', 'ecdsa-sha2-nistp256', 'ssh-ed25519'..."
                value={keyMaterial}
                onChange={(e) => setKeyMaterial(e.target.value)}
                rows={4}
                className="w-full font-mono text-[11px] p-2.5 rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
                required
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setIsAddKeyModalOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" size="sm" disabled={isSubmittingKey}>
              {isSubmittingKey ? 'Adding...' : 'Add SSH key'}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>

      {/* MODAL 2: GENERATE PERSONAL ACCESS TOKEN */}
      <Dialog open={isAddTokenModalOpen} onOpenChange={setIsAddTokenModalOpen}>
        <form onSubmit={handleCreateToken} className="space-y-4">
          <DialogHeader>
            <DialogTitle>New personal access token (classic)</DialogTitle>
            <DialogDescription>
              Personal access tokens function like ordinary OAuth access tokens. They can be used to authenticate to the API over HTTP.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-3 text-xs">
            <div>
              <label className="block font-medium text-foreground mb-1">Note</label>
              <input
                type="text"
                placeholder="What's this token for?"
                value={tokenNote}
                onChange={(e) => setTokenNote(e.target.value)}
                className="w-full px-3 py-1.5 rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
                required
              />
            </div>

            <div>
              <label className="block font-medium text-foreground mb-1">Expiration</label>
              <select
                value={tokenExpirationDays}
                onChange={(e) => setTokenExpirationDays(e.target.value)}
                className="w-full px-3 py-1.5 rounded-md bg-secondary border border-border text-foreground focus:outline-none"
              >
                <option value="30">30 days</option>
                <option value="60">60 days</option>
                <option value="90">90 days</option>
                <option value="365">1 year</option>
                <option value="0">No expiration</option>
              </select>
            </div>

            <div>
              <label className="block font-medium text-foreground mb-2">Select scopes</label>
              <div className="space-y-2 border border-border rounded-md p-3 bg-secondary/30">
                {[
                  { id: 'repo', label: 'repo', desc: 'Full control of private repositories' },
                  { id: 'workflow', label: 'workflow', desc: 'Update GitHub Action workflows' },
                  { id: 'write:packages', label: 'write:packages', desc: 'Upload packages to Registry' },
                  { id: 'admin:org', label: 'admin:org', desc: 'Full control of orgs and teams, read and write org projects' },
                  { id: 'mcp:invoke', label: 'mcp:invoke', desc: 'Execute autonomous Model Context Protocol tools' },
                ].map((s) => {
                  const isChecked = selectedScopes.includes(s.id)
                  return (
                    <label key={s.id} className="flex items-start gap-2.5 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={isChecked}
                        onChange={() => {
                          if (isChecked) {
                            setSelectedScopes(selectedScopes.filter((sc) => sc !== s.id))
                          } else {
                            setSelectedScopes([...selectedScopes, s.id])
                          }
                        }}
                        className="mt-0.5 rounded border-border text-primary focus:ring-primary"
                      />
                      <div>
                        <div className="font-mono font-semibold text-foreground">{s.label}</div>
                        <div className="text-[11px] text-muted-foreground">{s.desc}</div>
                      </div>
                    </label>
                  )
                })}
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setIsAddTokenModalOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" size="sm" disabled={isSubmittingToken}>
              {isSubmittingToken ? 'Generating...' : 'Generate token'}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>
    </div>
  )
}
