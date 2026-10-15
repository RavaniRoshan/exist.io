import React, { useState, useEffect } from 'react'
import {
  GitBranch,
  Terminal,
  ExternalLink,
  Copy,
  Check,
  Sparkles,
  BookOpen,
  Layers,
  LogIn,
  LogOut,
  Star,
  Eye,
  GitFork,
  Code2,
  GitPullRequest,
  GitCommit,
  Bot,
  ShieldCheck,
  Webhook,
  Settings,
  ChevronDown,
  AlertCircle,
  Tag,
  Compass,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { SearchPill } from '@/components/ui/Input'
import { Dialog } from '@/components/ui/Dialog'
import { UnderlineNav, type NavTabItem } from '@/components/ui/UnderlineNav'
import { isDemoModeActive, setDemoModeActive, checkActiveSession, logoutUser, type RepoInfo, type UserProfile } from '@/lib/api'
import { AuthModal } from '@/components/views/AuthModal'
import { cn } from '@/lib/utils'

export interface HeaderProps {
  repo: RepoInfo
  activeTab: string
  onTabChange: (tab: string) => void
  openPrCount?: number
  onOpenTour?: () => void
  onOpenManual?: () => void
}

/**
 * GitHub Primer Repository Header.
 * Top bar (56px) + Repository Title Breadcrumb & Actions + UnderlineNav bar.
 */
export const Header: React.FC<HeaderProps> = ({
  repo,
  activeTab,
  onTabChange,
  openPrCount = 1,
  onOpenTour,
  onOpenManual,
}) => {
  const [cloneModalOpen, setCloneModalOpen] = useState(false)
  const [cloneType, setCloneType] = useState<'http' | 'ssh' | 'mcp'>('http')
  const [copied, setCopied] = useState(false)
  const [isDemo, setIsDemo] = useState(isDemoModeActive())
  const [currentUser, setCurrentUser] = useState<UserProfile | null>(null)
  const [authModalOpen, setAuthModalOpen] = useState(false)
  const [starred, setStarred] = useState(false)
  const [starCount, setStarCount] = useState(14)

  const [userMenuOpen, setUserMenuOpen] = useState(false)

  useEffect(() => {
    checkActiveSession().then((res) => {
      if (res.authenticated && res.user) {
        setCurrentUser(res.user)
      } else {
        // Fallback default enterprise profile
        setCurrentUser({
          id: 'u-admin-1',
          login: 'shiva',
          name: 'Roshan Ravani',
          email: 'roshan@exist.io',
          role: 'admin',
          status: 'active',
          createdAt: '2026-01-01T00:00:00Z',
        })
      }
    })
  }, [])

  const handleLogout = async () => {
    await logoutUser()
    setCurrentUser(null)
    setUserMenuOpen(false)
  }

  const toggleDemoMode = () => {
    const next = !isDemo
    setIsDemo(next)
    setDemoModeActive(next)
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const getCloneCommand = () => {
    switch (cloneType) {
      case 'http':
        return `git clone ${repo.cloneUrlHttp}`
      case 'ssh':
        return `git clone ${repo.cloneUrlSsh}`
      case 'mcp':
        return `claude mcp add exist -- http://localhost:8080/mcp`
    }
  }

  // Canonical Repository Tabs
  const repoTabs: NavTabItem[] = [
    { id: 'code', label: 'Code', icon: <Code2 className="w-4 h-4" /> },
    { id: 'issues', label: 'Issues', count: 2, icon: <AlertCircle className="w-4 h-4" /> },
    { id: 'prs', label: 'Pull requests', count: openPrCount, icon: <GitPullRequest className="w-4 h-4" /> },
    { id: 'commits', label: 'Commits', icon: <GitCommit className="w-4 h-4" /> },
    { id: 'releases', label: 'Releases', count: 2, icon: <Tag className="w-4 h-4" /> },
    { id: 'agents', label: 'Agent Console', count: '35', icon: <Bot className="w-4 h-4" /> },
    { id: 'rulesets', label: 'Rulesets & Security', icon: <ShieldCheck className="w-4 h-4" /> },
    { id: 'webhooks', label: 'Webhooks', icon: <Webhook className="w-4 h-4" /> },
    { id: 'admin', label: 'Settings', icon: <Settings className="w-4 h-4" /> },
  ]

  return (
    <>
      {/* Sandbox Demo Mode Top Announcement Banner (Visible ONLY in Demo Mode) */}
      {isDemo && (
        <div className="bg-amber-950/30 border-b border-amber-500/30 px-4 py-1.5 text-center text-xs text-amber-300 font-medium flex items-center justify-center gap-2">
          <Layers className="w-3.5 h-3.5 text-amber-400" />
          <span>Running in Offline Sandbox Demo Mode — Data is simulated for evaluation.</span>
          <button
            onClick={toggleDemoMode}
            className="underline hover:text-foreground font-semibold ml-2 cursor-pointer"
          >
            Switch to Live Production API
          </button>
        </div>
      )}

      {/* 1. Global App Bar (56px) */}
      <header className="sticky top-0 z-40 w-full border-b border-border bg-card/95 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-14 gap-4">
            {/* Left: Brand Glyph & Title */}
            <div className="flex items-center gap-3">
              <button
                onClick={() => onTabChange('home')}
                className="flex items-center gap-2.5 text-foreground hover:opacity-90 transition-opacity cursor-pointer text-left"
                title="Go to Enterprise Home Feed"
              >
                <div className="w-7 h-7 rounded-md bg-primary text-primary-foreground flex items-center justify-center font-bold text-sm shadow-2xs">
                  <GitBranch className="w-4 h-4" />
                </div>
                <span className="font-bold text-sm tracking-tight text-foreground hidden sm:inline">exist.io</span>
              </button>

              {/* Mode Toggle Capsule */}
              <button
                onClick={toggleDemoMode}
                className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[11px] font-medium border transition-colors cursor-pointer ${
                  isDemo
                    ? 'bg-amber-500/10 text-amber-300 border-amber-500/30 hover:bg-amber-500/20'
                    : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 hover:bg-emerald-500/20'
                }`}
                title="Click to toggle between Live Production API and Offline Sandbox Demo mode"
              >
                <span className={`w-1.5 h-1.5 rounded-full ${isDemo ? 'bg-amber-400' : 'bg-emerald-400'}`} />
                <span>{isDemo ? 'Sandbox Demo' : 'Live Production'}</span>
              </button>
            </div>

            {/* Central Search Bar */}
            <div className="hidden md:flex flex-1 max-w-sm mx-4">
              <SearchPill />
            </div>

            {/* Right Action Group */}
            <div className="flex items-center gap-2">
              {/* MCP Tool Status Badge */}
              <button
                onClick={() => onTabChange('agents')}
                className="hidden sm:inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-secondary text-foreground border border-border hover:bg-secondary/80 transition-colors cursor-pointer"
                title="35 Model Context Protocol tools active"
              >
                <Sparkles className="w-3.5 h-3.5 text-primary" />
                <span>MCP: 35 tools</span>
              </button>

              {/* Tour and User Manual Buttons */}
              {onOpenTour && (
                <button
                  onClick={onOpenTour}
                  className="hidden sm:inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium bg-secondary text-foreground border border-border hover:bg-secondary/80 transition-colors cursor-pointer"
                  title="Take Interactive Onboarding Tour"
                >
                  <Compass className="w-3.5 h-3.5 text-primary" />
                  <span>Tour</span>
                </button>
              )}

              {onOpenManual && (
                <button
                  onClick={onOpenManual}
                  className="p-1.5 text-muted-foreground hover:text-foreground rounded-md hover:bg-secondary/60 transition-colors cursor-pointer"
                  title="In-App System Manual & Architecture Map"
                >
                  <BookOpen className="w-4 h-4 text-primary" />
                </button>
              )}

              {/* GitHub Link */}
              <a
                href="https://github.com/RavaniRoshan/exist.io"
                target="_blank"
                rel="noreferrer"
                className="p-1.5 text-muted-foreground hover:text-foreground rounded-md hover:bg-secondary/60 transition-colors"
                title="GitHub Repository"
              >
                <ExternalLink className="w-4 h-4" />
              </a>

              {/* User Authentication Menu & Profile Dropdown */}
              {currentUser ? (
                <div className="relative pl-2 border-l border-border">
                  <button
                    onClick={() => setUserMenuOpen(!userMenuOpen)}
                    className="flex items-center gap-1.5 p-0.5 rounded-full hover:ring-2 hover:ring-primary/40 transition-all cursor-pointer"
                    title={`Signed in as @${currentUser.login}`}
                  >
                    <div className="w-7 h-7 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-bold shadow-2xs">
                      {currentUser.login.charAt(0).toUpperCase()}
                    </div>
                  </button>

                  {/* Dropdown Menu */}
                  {userMenuOpen && (
                    <div className="absolute right-0 mt-2 w-52 rounded-md bg-card border border-border shadow-xl py-1 z-50 text-xs font-sans animate-in fade-in zoom-in-95 duration-100">
                      <div className="px-3 py-2 border-b border-border">
                        <div className="text-muted-foreground text-[11px]">Signed in as</div>
                        <div className="font-semibold text-foreground truncate">@{currentUser.login}</div>
                      </div>
                      <button
                        onClick={() => {
                          setUserMenuOpen(false)
                          onTabChange('profile')
                        }}
                        className="w-full text-left px-3 py-1.5 hover:bg-secondary/70 text-foreground transition-colors cursor-pointer flex items-center gap-2"
                      >
                        <span>Your profile</span>
                      </button>
                      <button
                        onClick={() => {
                          setUserMenuOpen(false)
                          onTabChange('admin')
                        }}
                        className="w-full text-left px-3 py-1.5 hover:bg-secondary/70 text-foreground transition-colors cursor-pointer flex items-center gap-2"
                      >
                        <span>Enterprise Settings</span>
                      </button>
                      <button
                        onClick={() => {
                          setUserMenuOpen(false)
                          onTabChange('home')
                        }}
                        className="w-full text-left px-3 py-1.5 hover:bg-secondary/70 text-foreground transition-colors cursor-pointer flex items-center gap-2"
                      >
                        <span>Enterprise Feed</span>
                      </button>
                      {onOpenTour && (
                        <button
                          onClick={() => {
                            setUserMenuOpen(false)
                            onOpenTour()
                          }}
                          className="w-full text-left px-3 py-1.5 hover:bg-secondary/70 text-foreground transition-colors cursor-pointer flex items-center gap-2"
                        >
                          <Compass className="w-3.5 h-3.5 text-primary" />
                          <span>Product Tour</span>
                        </button>
                      )}
                      {onOpenManual && (
                        <button
                          onClick={() => {
                            setUserMenuOpen(false)
                            onOpenManual()
                          }}
                          className="w-full text-left px-3 py-1.5 hover:bg-secondary/70 text-foreground transition-colors cursor-pointer flex items-center gap-2"
                        >
                          <BookOpen className="w-3.5 h-3.5 text-primary" />
                          <span>User Manual</span>
                        </button>
                      )}
                      <div className="border-t border-border my-1" />
                      <button
                        onClick={handleLogout}
                        className="w-full text-left px-3 py-1.5 hover:bg-destructive/10 text-destructive transition-colors cursor-pointer flex items-center gap-2"
                      >
                        <LogOut className="w-3.5 h-3.5" />
                        <span>Sign out</span>
                      </button>
                    </div>
                  )}
                </div>
              ) : (
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setAuthModalOpen(true)}
                  className="gap-1.5"
                >
                  <LogIn className="w-3.5 h-3.5 text-primary" />
                  <span>Sign In</span>
                </Button>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* 2. Repository Context Title Bar & UnderlineNav */}
      <div className="w-full bg-card/50 border-b border-border pt-4">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-3">
          {/* Top Row: Breadcrumb & Repo Actions */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            {/* Breadcrumb: owner / repo */}
            <div className="flex items-center gap-2 text-sm font-semibold tracking-tight">
              <span className="text-muted-foreground hover:underline cursor-pointer">{repo.ownerSlug}</span>
              <span className="text-muted-foreground">/</span>
              <span className="text-foreground hover:underline font-bold cursor-pointer text-base">{repo.name}</span>
              <Badge variant="outline" className="ml-1 text-[11px] py-0 px-1.5 border-border text-muted-foreground">
                Public
              </Badge>
            </div>

            {/* Action Group: Watch, Fork, Star, Code Dropdown */}
            <div className="flex items-center gap-1.5">
              <Button variant="secondary" size="sm" className="gap-1.5">
                <Eye className="w-3.5 h-3.5 text-muted-foreground" />
                <span>Watch</span>
                <span className="ml-1 px-1.5 py-0.2 rounded-full bg-background border border-border text-[10px] font-mono">1</span>
              </Button>

              <Button variant="secondary" size="sm" className="gap-1.5">
                <GitFork className="w-3.5 h-3.5 text-muted-foreground" />
                <span>Fork</span>
                <span className="ml-1 px-1.5 py-0.2 rounded-full bg-background border border-border text-[10px] font-mono">0</span>
              </Button>

              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  setStarred(!starred)
                  setStarCount(starred ? starCount - 1 : starCount + 1)
                }}
                className="gap-1.5"
              >
                <Star className={cn('w-3.5 h-3.5', starred ? 'text-amber-400 fill-amber-400' : 'text-muted-foreground')} />
                <span>{starred ? 'Starred' : 'Star'}</span>
                <span className="ml-1 px-1.5 py-0.2 rounded-full bg-background border border-border text-[10px] font-mono">{starCount}</span>
              </Button>

              {/* Code Clone Button */}
              <Button
                variant="primary"
                size="sm"
                onClick={() => setCloneModalOpen(true)}
                className="gap-1.5 ml-1"
              >
                <Terminal className="w-3.5 h-3.5" />
                <span>Code</span>
                <ChevronDown className="w-3 h-3 ml-0.5" />
              </Button>
            </div>
          </div>

          {/* UnderlineNav Repository Tabs */}
          <UnderlineNav
            tabs={repoTabs}
            activeTab={activeTab === 'cluster' || activeTab === 'users' || activeTab === 'sso' || activeTab === 'audit' ? 'admin' : activeTab}
            onChange={onTabChange}
          />
        </div>
      </div>

      {/* Clone Instructions Modal (Primer Dialog) */}
      <Dialog
        isOpen={cloneModalOpen}
        onClose={() => setCloneModalOpen(false)}
        title="Clone repository"
        description="Select your preferred transport protocol to clone or interact with autonomous agents."
      >
        <div className="space-y-4 pt-1">
          {/* Transport Selector Tabs */}
          <div className="flex rounded-md bg-secondary/50 p-1 border border-border text-xs font-semibold">
            <button
              onClick={() => setCloneType('http')}
              className={`flex-1 py-1.5 rounded-md transition-colors cursor-pointer ${
                cloneType === 'http' ? 'bg-card text-foreground shadow-2xs' : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Smart-HTTP
            </button>
            <button
              onClick={() => setCloneType('ssh')}
              className={`flex-1 py-1.5 rounded-md transition-colors cursor-pointer ${
                cloneType === 'ssh' ? 'bg-card text-foreground shadow-2xs' : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              SSH (Port 2222)
            </button>
            <button
              onClick={() => setCloneType('mcp')}
              className={`flex-1 py-1.5 rounded-md transition-colors cursor-pointer ${
                cloneType === 'mcp' ? 'bg-card text-foreground shadow-2xs' : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              MCP Agent
            </button>
          </div>

          {/* Code snippet display */}
          <div className="relative rounded-md bg-input border border-border p-2.5 font-mono text-xs text-foreground overflow-x-auto flex items-center justify-between gap-2">
            <code className="select-all text-[11px] truncate flex-1">{getCloneCommand()}</code>
            <Button
              variant="secondary"
              size="xs"
              onClick={() => copyToClipboard(getCloneCommand())}
              className="shrink-0"
            >
              {copied ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
              <span>{copied ? 'Copied' : 'Copy'}</span>
            </Button>
          </div>

          <p className="text-[11px] text-muted-foreground">
            {cloneType === 'http' && 'Clone using Git Smart-HTTP transfer over standard HTTPS/HTTP.'}
            {cloneType === 'ssh' && 'Clone using embedded zero-cgo pure-Go SSH server with enrolled keys.'}
            {cloneType === 'mcp' && 'Add exist.io Model Context Protocol endpoint to Claude Desktop or cursor-agent.'}
          </p>
        </div>
      </Dialog>

      {/* Auth Modal */}
      <AuthModal
        isOpen={authModalOpen}
        onClose={() => setAuthModalOpen(false)}
        onSuccess={(user) => setCurrentUser(user)}
      />
    </>
  )
}
