import { useState, useEffect } from 'react'
import {
  Activity,
  Server,
  ShieldAlert,
  KeyRound,
  Users,
} from 'lucide-react'
import { Header } from '@/components/views/Header'
import { RepoOverview } from '@/components/views/RepoOverview'
import { PullRequestsView } from '@/components/views/PullRequestsView'
import { CommitsView } from '@/components/views/CommitsView'
import { AgentConsoleView } from '@/components/views/AgentConsoleView'
import { RulesetsView } from '@/components/views/RulesetsView'
import { WebhooksView } from '@/components/views/WebhooksView'
import { UserProfileView } from '@/components/views/UserProfileView'
import { GlobalHomeView } from '@/components/views/GlobalHomeView'
import { IssuesView } from '@/components/views/IssuesView'
import { ReleasesView } from '@/components/views/ReleasesView'
import { ProductTour } from '@/components/views/ProductTour'
import { UserManualModal } from '@/components/views/UserManualModal'
import { ClusterHealth } from '@/components/views/admin/ClusterHealth'
import { AuditLogs } from '@/components/views/admin/AuditLogs'
import { SsoConfig } from '@/components/views/admin/SsoConfig'
import { EnterpriseUsers } from '@/components/views/admin/EnterpriseUsers'
import { RepoSettingsView } from '@/components/views/RepoSettingsView'
import { defaultRepo, getRepoInfo } from '@/lib/api'
import { cn } from '@/lib/utils'

export default function App() {
  const [activeTab, setActiveTab] = useState<'home' | 'profile' | 'code' | 'issues' | 'prs' | 'commits' | 'releases' | 'agents' | 'rulesets' | 'webhooks' | 'admin' | 'settings'>('code')
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
    setCurrentRepo({
      ownerSlug: owner,
      name: name,
      description: `Enterprise repository ${owner}/${name}`,
      defaultBranch: 'main',
      isPrivate: false,
      starsCount: 0,
      forksCount: 0,
      updatedAt: 'Just now',
      cloneUrlHttp: `http://${window.location.host}/${owner}/${name}.git`,
      cloneUrlSsh: `git@${window.location.hostname}:2222/${owner}/${name}.git`,
    })
    setActiveTab('code')
  }

  useEffect(() => {
    // If first-time user hasn't completed tour yet, auto-open tour
    const completed = localStorage.getItem('exist_onboarding_tour_completed')
    if (completed !== 'true') {
      setTourOpen(true)
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
          <UserProfileView
            onSelectRepo={(repoSlug) => handleSelectRepo(repoSlug)}
          />
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
              {adminSection === 'cluster' && <ClusterHealth />}
              {adminSection === 'users' && <EnterpriseUsers />}
              {adminSection === 'sso' && <SsoConfig />}
              {adminSection === 'audit' && <AuditLogs />}
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
              exist.io v0.4.4-enterprise
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
