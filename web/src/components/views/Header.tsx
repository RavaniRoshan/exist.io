import React, { useState, useEffect } from 'react'
import { motion } from 'motion/react'
import { GitBranch, Terminal, ExternalLink, Copy, Check, ShieldCheck, Sparkles, BookOpen, Layers, LogIn, LogOut } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { SearchPill } from '@/components/ui/Input'
import { Dialog } from '@/components/ui/Dialog'
import { isDemoModeActive, setDemoModeActive, checkActiveSession, logoutUser, type RepoInfo, type UserProfile } from '@/lib/api'
import { AuthModal } from '@/components/views/AuthModal'

export interface HeaderProps {
  repo: RepoInfo
  activeTab: string
  onTabChange: (tab: string) => void
  onOpenAgentConsole: () => void
}

export const Header: React.FC<HeaderProps> = ({
  repo,
  activeTab: _activeTab,
  onTabChange: _onTabChange,
  onOpenAgentConsole,
}) => {
  const [cloneModalOpen, setCloneModalOpen] = useState(false)
  const [cloneType, setCloneType] = useState<'http' | 'ssh' | 'mcp'>('http')
  const [copied, setCopied] = useState(false)
  const [isDemo, setIsDemo] = useState(isDemoModeActive())
  const [currentUser, setCurrentUser] = useState<UserProfile | null>(null)
  const [authModalOpen, setAuthModalOpen] = useState(false)

  useEffect(() => {
    checkActiveSession().then((res) => {
      if (res.authenticated && res.user) {
        setCurrentUser(res.user)
      }
    })
  }, [])

  const handleLogout = async () => {
    await logoutUser()
    setCurrentUser(null)
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

  return (
    <>
      {/* Sandbox Demo Mode Top Announcement Banner (Visible ONLY in Demo Mode) */}
      {isDemo && (
        <div className="bg-gradient-to-r from-amber-600/20 via-orange-600/20 to-amber-600/20 border-b border-amber-500/30 px-4 py-1.5 text-center text-xs text-amber-300 font-medium flex items-center justify-center gap-2">
          <Layers className="w-3.5 h-3.5 text-amber-400" />
          <span>Running in Offline Sandbox Demo Mode — Data is simulated for evaluation.</span>
          <button
            onClick={toggleDemoMode}
            className="underline hover:text-white font-semibold ml-2 cursor-pointer"
          >
            Switch to Live Production API
          </button>
        </div>
      )}

      <header className="sticky top-0 z-40 w-full border-b border-border bg-card/85 backdrop-blur-xl">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16 gap-4">
            {/* Logo & Breadcrumb */}
            <div className="flex items-center gap-3">
              <motion.div
                whileHover={{ rotate: 10, scale: 1.05 }}
                className="flex items-center justify-center w-9 h-9 rounded-xl bg-primary text-primary-foreground shadow-sm border border-primary/40"
              >
                <GitBranch className="w-5 h-5 text-current" />
              </motion.div>
              <div className="flex items-center gap-2 text-sm font-semibold tracking-tight">
                <span className="text-muted-foreground hover:text-foreground cursor-pointer">{repo.ownerSlug}</span>
                <span className="text-border">/</span>
                <span className="text-foreground hover:text-primary cursor-pointer">{repo.name}</span>
                <Badge variant="outline" className="ml-1 text-[10px] py-0 border-border text-muted-foreground">
                  Public
                </Badge>
              </div>
            </div>

            {/* Central Search Pill */}
            <div className="hidden md:flex flex-1 max-w-md mx-4">
              <SearchPill />
            </div>

            {/* Right Action Bar */}
            <div className="flex items-center gap-2.5">
              {/* Mode Toggle: Production API vs Sandbox Demo */}
              <button
                onClick={toggleDemoMode}
                className={`hidden sm:inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium border transition-colors ${
                  isDemo
                    ? 'bg-amber-950/40 text-amber-300 border-amber-500/40 hover:bg-amber-950/60'
                    : 'bg-emerald-950/40 text-emerald-300 border-emerald-500/40 hover:bg-emerald-950/60'
                }`}
                title="Click to toggle between Live Production API and Offline Sandbox Demo mode"
              >
                <span className={`w-2 h-2 rounded-full ${isDemo ? 'bg-amber-400' : 'bg-emerald-400 animate-pulse'}`} />
                <span>{isDemo ? 'Sandbox Demo' : 'Live Production'}</span>
              </button>

              {/* Live MCP Agent Status Pill */}
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={onOpenAgentConsole}
                className="hidden sm:inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium bg-primary/10 text-primary border border-primary/30 hover:border-primary hover:bg-primary/20 transition-all shadow-xs"
              >
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-primary" />
                </span>
                <Sparkles className="w-3.5 h-3.5 text-primary" />
                <span className="font-semibold">MCP Active</span>
                <span className="text-[10px] text-primary bg-primary/20 px-1.5 py-0.2 rounded border border-primary/40">
                  35 tools
                </span>
              </motion.button>

              {/* Quick Clone Button */}
              <Button
                variant="glow"
                size="sm"
                onClick={() => setCloneModalOpen(true)}
                className="gap-1.5"
              >
                <Terminal className="w-3.5 h-3.5" />
                <span>Clone</span>
              </Button>

              {/* Documentation Site Link */}
              <a
                href="https://ravaniroshan.github.io/exist.io/"
                target="_blank"
                rel="noreferrer"
                className="p-2 text-slate-400 hover:text-white rounded-xl hover:bg-slate-900 border border-transparent hover:border-slate-800 transition-colors"
                title="Documentation Site"
              >
                <BookOpen className="w-4 h-4" />
              </a>

              {/* GitHub Remote Link */}
              <a
                href="https://github.com/RavaniRoshan/exist.io"
                target="_blank"
                rel="noreferrer"
                className="p-2 text-slate-400 hover:text-white rounded-xl hover:bg-slate-900 border border-transparent hover:border-slate-800 transition-colors"
                title="GitHub Repository"
              >
                <ExternalLink className="w-4 h-4" />
              </a>

              {/* Authentication Status / Sign In Button */}
              {currentUser ? (
                <div className="flex items-center gap-2 pl-2 border-l border-border">
                  <div className="flex items-center gap-2">
                    <div className="w-7 h-7 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-bold shadow-xs">
                      {currentUser.login.charAt(0).toUpperCase()}
                    </div>
                    <div className="hidden lg:flex flex-col text-left">
                      <span className="text-xs font-medium text-foreground leading-tight">{currentUser.name || currentUser.login}</span>
                      <span className="text-[10px] text-muted-foreground capitalize">{currentUser.role}</span>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleLogout}
                    className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 cursor-pointer"
                    title="Sign Out"
                  >
                    <LogOut className="w-3.5 h-3.5" />
                  </Button>
                </div>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setAuthModalOpen(true)}
                  className="gap-1.5 border-border bg-secondary hover:bg-secondary/80 text-foreground cursor-pointer"
                >
                  <LogIn className="w-3.5 h-3.5 text-primary" />
                  <span>Sign In</span>
                </Button>
              )}
            </div>
          </div>
        </div>

        {/* Clone Instructions Modal */}
        <Dialog
          isOpen={cloneModalOpen}
          onClose={() => setCloneModalOpen(false)}
          title="Clone or Connect to exist.io"
          description="Select your preferred transport protocol or connect an autonomous coding agent."
        >
          <div className="space-y-4 pt-2">
            {/* Transport Selector Tabs */}
            <div className="flex rounded-xl bg-background p-1 border border-border text-xs font-medium">
              <button
                onClick={() => setCloneType('http')}
                className={`flex-1 py-1.5 rounded-lg transition-colors ${
                  cloneType === 'http' ? 'bg-primary/20 text-primary font-semibold' : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                Smart-HTTP
              </button>
              <button
                onClick={() => setCloneType('ssh')}
                className={`flex-1 py-1.5 rounded-lg transition-colors ${
                  cloneType === 'ssh' ? 'bg-primary/20 text-primary font-semibold' : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                SSH
              </button>
              <button
                onClick={() => setCloneType('mcp')}
                className={`flex-1 py-1.5 rounded-lg transition-colors ${
                  cloneType === 'mcp' ? 'bg-primary/20 text-primary font-semibold' : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                MCP Agent
              </button>
            </div>

            {/* Code snippet display */}
            <div className="relative rounded-xl bg-background border border-border p-3 font-mono text-xs text-primary overflow-x-auto flex items-center justify-between">
              <code className="text-foreground select-all">{getCloneCommand()}</code>
              <Button
                variant="ghost"
                size="xs"
                onClick={() => copyToClipboard(getCloneCommand())}
                className="shrink-0 ml-2"
              >
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              </Button>
            </div>

            {/* Protocol notes */}
            <div className="text-xs text-slate-400 bg-slate-900/60 p-3 rounded-xl border border-slate-800/80 space-y-1.5">
              {cloneType === 'http' && (
                <p>Uses pure-Go Smart-HTTP transfer. Authenticate using your username and personal access token (PAT).</p>
              )}
              {cloneType === 'ssh' && (
                <p>Enterprise SSH key transport. Enforces CIDR IP allowlists and ruleset bypass hierarchies.</p>
              )}
              {cloneType === 'mcp' && (
                <div className="flex items-start gap-2 text-sky-300">
                  <ShieldCheck className="w-4 h-4 shrink-0 text-sky-400 mt-0.5" />
                  <p>Enables 34 forge tools (pull requests, reviews, diffs, checks, rulesets) for Claude, Cursor, and Codex.</p>
                </div>
              )}
            </div>
          </div>
        </Dialog>

        {/* Authentication Modal */}
        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(u) => setCurrentUser(u)}
        />
      </header>
    </>
  )
}
