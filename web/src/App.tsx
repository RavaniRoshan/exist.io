import { useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import {
  Code2,
  GitPullRequest,
  GitCommit,
  Bot,
  Shield,
  Webhook,
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
import { ClusterHealth } from '@/components/views/admin/ClusterHealth'
import { AuditLogs } from '@/components/views/admin/AuditLogs'
import { SsoConfig } from '@/components/views/admin/SsoConfig'
import { EnterpriseUsers } from '@/components/views/admin/EnterpriseUsers'
import { Tabs } from '@/components/ui/Tabs'
import { defaultRepo } from '@/lib/api'

export default function App() {
  const [navSection, setNavSection] = useState<'forge' | 'admin'>('forge')
  const [activeTab, setActiveTab] = useState('code')
  const repo = defaultRepo

  const forgeTabs = [
    { id: 'code', label: 'Code', icon: <Code2 className="w-3.5 h-3.5" /> },
    { id: 'pulls', label: 'Pull Requests', count: 1, icon: <GitPullRequest className="w-3.5 h-3.5" /> },
    { id: 'commits', label: 'Commits', count: 148, icon: <GitCommit className="w-3.5 h-3.5" /> },
    { id: 'agents', label: 'AI Agents & MCP', count: '34 tools', icon: <Bot className="w-3.5 h-3.5 text-sky-400" /> },
    { id: 'rulesets', label: 'Rulesets & Policy', icon: <Shield className="w-3.5 h-3.5" /> },
    { id: 'webhooks', label: 'Webhooks', icon: <Webhook className="w-3.5 h-3.5" /> },
  ]

  const adminTabs = [
    { id: 'cluster', label: 'Cluster Health', icon: <Server className="w-3.5 h-3.5 text-emerald-400" /> },
    { id: 'audit', label: 'Compliance Audit Logs', icon: <ShieldAlert className="w-3.5 h-3.5 text-purple-400" /> },
    { id: 'sso', label: 'Enterprise SSO & SCIM', icon: <KeyRound className="w-3.5 h-3.5 text-sky-400" /> },
    { id: 'users', label: 'Users & RBAC', icon: <Users className="w-3.5 h-3.5 text-teal-400" /> },
  ]

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col font-sans selection:bg-primary/25 selection:text-primary">
      {/* Space UI Subtle Mesh Grid Background */}
      <div className="fixed inset-0 pointer-events-none bg-space-grid opacity-80 -z-10" />

      {/* Top Application Header */}
      <Header
        repo={repo}
        activeTab={activeTab}
        onTabChange={setActiveTab}
        onOpenAgentConsole={() => {
          setNavSection('forge')
          setActiveTab('agents')
        }}
      />

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
        {/* Navigation Section Switcher: Forge vs Enterprise Admin */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-border">
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1 p-1 bg-card border border-border rounded-xl text-xs font-semibold">
              <button
                onClick={() => {
                  setNavSection('forge')
                  setActiveTab('code')
                }}
                className={`px-3 py-1.5 rounded-lg transition-colors ${
                  navSection === 'forge'
                    ? 'bg-primary/20 text-primary border border-primary/30 shadow-xs'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                Forge Workspace
              </button>
              <button
                onClick={() => {
                  setNavSection('admin')
                  setActiveTab('cluster')
                }}
                className={`px-3 py-1.5 rounded-lg transition-colors ${
                  navSection === 'admin'
                    ? 'bg-accent/25 text-foreground border border-accent/40 shadow-xs'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                Enterprise Admin
              </button>
            </div>
          </div>

          {/* Sub-tabs for the active section */}
          <Tabs
            tabs={navSection === 'forge' ? forgeTabs : adminTabs}
            activeTab={activeTab}
            onChange={setActiveTab}
          />
        </div>

        {/* Tab Content Panes with Fluid Transitions */}
        <AnimatePresence mode="wait">
          <motion.div
            key={`${navSection}-${activeTab}`}
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.18, ease: 'easeOut' }}
          >
            {/* Forge Workspace Views */}
            {navSection === 'forge' && activeTab === 'code' && (
              <RepoOverview
                repo={repo}
                onSelectCommits={() => setActiveTab('commits')}
                onSelectAgentConsole={() => setActiveTab('agents')}
              />
            )}

            {navSection === 'forge' && activeTab === 'pulls' && (
              <PullRequestsView
                repo={repo}
                onSelectAgentConsole={() => setActiveTab('agents')}
              />
            )}

            {navSection === 'forge' && activeTab === 'commits' && (
              <CommitsView repo={repo} />
            )}

            {navSection === 'forge' && activeTab === 'agents' && (
              <AgentConsoleView />
            )}

            {navSection === 'forge' && activeTab === 'rulesets' && (
              <RulesetsView repo={repo} />
            )}

            {navSection === 'forge' && activeTab === 'webhooks' && (
              <WebhooksView />
            )}

            {/* Enterprise Admin Console Views */}
            {navSection === 'admin' && activeTab === 'cluster' && (
              <ClusterHealth />
            )}

            {navSection === 'admin' && activeTab === 'audit' && (
              <AuditLogs />
            )}

            {navSection === 'admin' && activeTab === 'sso' && (
              <SsoConfig />
            )}

            {navSection === 'admin' && activeTab === 'users' && (
              <EnterpriseUsers />
            )}
          </motion.div>
        </AnimatePresence>
      </main>

      {/* Space UI Live Telemetry Footer */}
      <footer className="border-t border-border bg-card/70 backdrop-blur-md py-4 mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-muted-foreground font-mono">
          <div className="flex items-center gap-3">
            <span className="flex items-center gap-1.5 text-foreground/80">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-400" />
              </span>
              <span>existd 127.0.0.1:8080</span>
            </span>
            <span>•</span>
            <span className="text-foreground/80 flex items-center gap-1">
              <Activity className="w-3 h-3 text-primary" />
              <span>SQLite WAL</span>
            </span>
            <span>•</span>
            <span className="text-foreground/80">35 MCP Tools Mounted</span>
          </div>

          <div className="flex items-center gap-4">
            <a
              href="https://ravaniroshan.github.io/exist.io/"
              target="_blank"
              rel="noreferrer"
              className="hover:text-primary transition-colors"
            >
              Docs & Guide
            </a>
            <a
              href="https://github.com/RavaniRoshan/exist.io"
              target="_blank"
              rel="noreferrer"
              className="hover:text-primary transition-colors"
            >
              GitHub
            </a>
            <span className="text-border">|</span>
            <span className="flex items-center gap-1 text-foreground/80">
              <span>Enterprise Ready v0.3.0</span>
            </span>
          </div>
        </div>
      </footer>
    </div>
  )
}
