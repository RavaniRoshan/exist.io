import { useState, useEffect, lazy, Suspense } from 'react'
import {
  Activity,
  Server,
  ShieldAlert,
  KeyRound,
  Users,
  Loader2,
} from 'lucide-react'
import { Header } from '@/components/views/Header'
import { RepoOverview } from '@/components/views/RepoOverview'
import { PullRequestsView } from '@/components/views/PullRequestsView'
import { CommitsView } from '@/components/views/CommitsView'
import { AgentConsoleView } from '@/components/views/AgentConsoleView'
import { RulesetsView } from '@/components/views/RulesetsView'
import { WebhooksView } from '@/components/views/WebhooksView'
import { GlobalHomeView } from '@/components/views/GlobalHomeView'
import { IssuesView } from '@/components/views/IssuesView'
import { ReleasesView } from '@/components/views/ReleasesView'
import { ProductTour } from '@/components/views/ProductTour'
import { UserManualModal } from '@/components/views/UserManualModal'
import { defaultRepo, getRepoInfo, getRepositories, isDemoModeActive } from '@/lib/api'
import { cn } from '@/lib/utils'

// Lazy-loaded views for bundle optimization & code splitting
const RepoSettingsView = lazy(() =>
  import('@/components/views/RepoSettingsView').then((m) => ({ default: m.RepoSettingsView }))
)
const UserProfileView = lazy(() =>
  import('@/components/views/UserProfileView').then((m) => ({ default: m.UserProfileView }))
)
const ClusterHealth = lazy(() =>
  import('@/components/views/admin/ClusterHealth').then((m) => ({ default: m.ClusterHealth }))
)
const AuditLogs = lazy(() =>
  import('@/components/views/admin/AuditLogs').then((m) => ({ default: m.AuditLogs }))
)
const SsoConfig = lazy(() =>
  import('@/components/views/admin/SsoConfig').then((m) => ({ default: m.SsoConfig }))
)
const EnterpriseUsers = lazy(() =>
  import('@/components/views/admin/EnterpriseUsers').then((m) => ({ default: m.EnterpriseUsers }))
)

const ViewFallback = () => (
  <div className="py-16 text-center text-xs text-muted-foreground">
    <Loader2 className="w-5 h-5 animate-spin mx-auto mb-2 text-primary" />
    <span>Loading view...</span>
  </div>
)

export default function App() {
  const [activeTab, setActiveTab] = useState<'home' | 'profile' | 'code' | 'issues' | 'prs' | 'commits' | 'releases' | 'agents' | 'rulesets' | 'webhooks' | 'admin' | 'settings'>(() => {
    if (typeof window !== 'undefined') {
      const p = new URLSearchParams(window.location.search)
      const t = p.get('tab')
      if (t && ['home', 'profile', 'code', 'issues', 'prs', 'commits', 'releases', 'agents', 'rulesets', 'webhooks', 'admin', 'settings'].includes(t)) {
        return t as any
      }
    }
    return 'code'
  })
  const [adminSection, setAdminSection] = useState<'cluster' | 'users' | 'sso' | 'audit'>('cluster')
  const [tourOpen, setTourOpen] = useState(false)
  const [manualOpen, setManualOpen] = useState(false)
  const [currentRepo, setCurrentRepo] = useState(defaultRepo)

  const handleSelectRepo = (slug?: string) => {
    if (!slug) {
      setActiveTab('code')
      return
    }
    let owner = 'shiva'
    let name = slug
    if (slug.includes('/')) {
      const parts = slug.split('/')
      owner = parts[0]
      name = parts[1]
    }
    getRepoInfo(owner, name).then((info) => {
      setCurrentRepo(info)
    })
    setActiveTab('code')
  }

  useEffect(() => {
    // If first-time user hasn't completed tour yet, auto-open tour (unless ?notour=1)
    const p = new URLSearchParams(window.location.search)
    if (p.get('notour') !== '1') {
      const completed = localStorage.getItem('exist_onboarding_tour_completed')
      if (completed !== 'true') {
        setTourOpen(true)
      }
    }

    // In live mode, dynamically resolve the active primary repository from cluster
    if (!isDemoModeActive()) {
      getRepositories().then((repos) => {
        if (repos && repos.length > 0) {
          const first = repos[0]
          let owner = 'shiva'
          let name = first.name
          if (first.full_name && first.full_name.includes('/')) {
            const parts = first.full_name.split('/')
            owner = parts[0]
            name = parts[1]
          }
          getRepoInfo(owner, name).then((info) => {
            setCurrentRepo(info)
          })
        }
      })
    }
  }, [])

  const adminNavItems = [
    { id: 'cluster', label: 'Cluster Health & Metrics', icon: <Server className="w-4 h-4" /> },
    { id: 'users', label: 'User Directory & SCIM', icon: <Users className="w-4 h-4" /> },
    { id: 'sso', label: 'Single Sign-On (SAML/OIDC)', icon: <KeyRound className="w-4 h-4" /> },
    { id: 'audit', label: 'Audit Log & Security', icon: <ShieldAlert className="w-4 h-4" /> },
  ] as const

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col font-sans selection:bg-primary/25 selection:text-primary">
      {/* Top Application Header & UnderlineNav */}
      <Header
        repo={currentRepo}
        activeTab={activeTab}
        onTabChange={(tab) => {
          if (tab === 'admin') {
            setActiveTab('admin')
          } else {
            setActiveTab(tab as any)
          }
        }}
        onOpenTour={() => setTourOpen(true)}
        onOpenManual={() => setManualOpen(true)}
      />

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* 0. Enterprise Home Activity Feed */}
        {activeTab === 'home' && (
          <GlobalHomeView
            onSelectRepo={(repoSlug) => handleSelectRepo(repoSlug)}
            onSelectTab={(tab) => setActiveTab(tab as any)}
            onOpenTour={() => setTourOpen(true)}
            onOpenManual={() => setManualOpen(true)}
          />
        )}

        {/* 0.1 User Profile & Keys Management */}
        {activeTab === 'profile' && (
          <Suspense fallback={<ViewFallback />}>
            <UserProfileView
              onSelectRepo={(repoSlug) => handleSelectRepo(repoSlug)}
            />
          </Suspense>
        )}

        {/* 1. Code & File Tree View */}
        {activeTab === 'code' && (
          <RepoOverview
            repo={currentRepo}
            onSelectCommits={() => setActiveTab('commits')}
            onSelectAgentConsole={() => setActiveTab('agents')}
          />
        )}

        {/* 1.5. Issues View */}
        {activeTab === 'issues' && (
          <IssuesView
            repo={currentRepo}
            onSelectPullRequest={() => setActiveTab('prs')}
          />
        )}

        {/* 2. Pull Requests View */}
        {activeTab === 'prs' && (
          <PullRequestsView
            repo={currentRepo}
            onSelectAgentConsole={() => setActiveTab('agents')}
          />
        )}

        {/* 3. Commits History View */}
        {activeTab === 'commits' && (
          <CommitsView repo={currentRepo} />
        )}

        {/* 3.5. Releases & Tags View */}
        {activeTab === 'releases' && (
          <ReleasesView repo={currentRepo} />
        )}

        {/* 4. Autonomous Agent Console & MCP Telemetry */}
        {activeTab === 'agents' && (
          <AgentConsoleView />
        )}

        {/* 5. Branch Rulesets & Policy View */}
        {activeTab === 'rulesets' && (
          <RulesetsView repo={currentRepo} />
        )}

        {/* 6. Webhooks Delivery View */}
        {activeTab === 'webhooks' && (
          <WebhooksView />
        )}

        {/* 6.5. Repository Settings & Collaborators */}
        {activeTab === 'settings' && (
          <Suspense fallback={<ViewFallback />}>
            <RepoSettingsView
              repo={currentRepo}
              onRepoUpdated={() => {
                getRepoInfo(currentRepo.ownerSlug, currentRepo.name).then(setCurrentRepo)
              }}
              onRepoDeleted={() => {
                handleSelectRepo('')
                setActiveTab('home')
              }}
            />
          </Suspense>
        )}

        {/* 7. Enterprise Admin & Settings View (Classic 2-column layout) */}
        {activeTab === 'admin' && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 items-start">
            {/* Left Settings Sidebar (25%) */}
            <aside className="md:col-span-1 space-y-1">
              <div className="px-3 pb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Enterprise Settings
              </div>
              <nav className="space-y-0.5">
                {adminNavItems.map((item) => (
                  <button
                    key={item.id}
                    onClick={() => setAdminSection(item.id)}
                    className={cn(
                      'w-full flex items-center gap-2.5 px-3 py-2 text-xs font-medium rounded-md transition-colors text-left cursor-pointer',
                      adminSection === item.id
                        ? 'bg-secondary text-foreground font-semibold border-l-2 border-primary'
                        : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
                    )}
                  >
                    <span className={cn('w-4 h-4 shrink-0', adminSection === item.id ? 'text-primary' : 'text-muted-foreground')}>
                      {item.icon}
                    </span>
                    <span>{item.label}</span>
                  </button>
                ))}
              </nav>
            </aside>

            {/* Right Settings Content (75%) */}
            <section className="md:col-span-3">
              <Suspense fallback={<ViewFallback />}>
                {adminSection === 'cluster' && <ClusterHealth />}
                {adminSection === 'users' && <EnterpriseUsers />}
                {adminSection === 'sso' && <SsoConfig />}
                {adminSection === 'audit' && <AuditLogs />}
              </Suspense>
            </section>
          </div>
        )}
      </main>

      {/* GitHub Primer Footer */}
      <footer className="border-t border-border bg-card/60 py-5 mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-muted-foreground font-mono">
          <div className="flex items-center gap-3">
            <span className="flex items-center gap-1.5 text-foreground">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-400" />
              </span>
              <span>existd (127.0.0.1:8080)</span>
            </span>
            <span>•</span>
            <span className="text-foreground flex items-center gap-1">
              <Activity className="w-3.5 h-3.5 text-primary" />
              <span>SQLite WAL</span>
            </span>
            <span>•</span>
            <span className="text-foreground">35 MCP Tools Mounted</span>
          </div>

          <div className="flex items-center gap-4">
            <a
              href="https://ravaniroshan.github.io/exist.io/"
              target="_blank"
              rel="noreferrer"
              className="hover:text-primary transition-colors hover:underline"
            >
              Docs & Guide
            </a>
            <a
              href="https://github.com/RavaniRoshan/exist.io"
              target="_blank"
              rel="noreferrer"
              className="hover:text-primary transition-colors hover:underline"
            >
              GitHub
            </a>
            <span className="text-border">|</span>
            <span className="text-foreground font-semibold">
              exist.io v0.5.0-beta
            </span>
          </div>
        </div>
      </footer>

      {/* Interactive First-Time Onboarding Tour */}
      <ProductTour
        isOpen={tourOpen}
        onClose={() => setTourOpen(false)}
        onOpenManual={() => setManualOpen(true)}
      />

      {/* In-App System Architecture User Manual */}
      <UserManualModal
        isOpen={manualOpen}
        onClose={() => setManualOpen(false)}
      />
    </div>
  )
}
