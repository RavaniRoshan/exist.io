import React, { useState, useEffect } from 'react'
import { motion } from 'motion/react'
import {
  BookOpen,
  Bot,
  ShieldCheck,
  Users,
  Terminal,
  Copy,
  Check,
  X,
  Cpu,
  Lock,
  Network,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { UnderlineNav, type NavTabItem } from '@/components/ui/UnderlineNav'
import { fetchClusterInfo, type ClusterSystemInfo } from '@/lib/api'

export interface UserManualModalProps {
  isOpen: boolean
  onClose: () => void
  initialTab?: string
}

export const UserManualModal: React.FC<UserManualModalProps> = ({
  isOpen,
  onClose,
  initialTab = 'quickstart',
}) => {
  const [activeTab, setActiveTab] = useState(initialTab)
  const [clusterInfo, setClusterInfo] = useState<ClusterSystemInfo | null>(null)
  const [copiedSection, setCopiedSection] = useState<string | null>(null)

  useEffect(() => {
    if (isOpen) {
      fetchClusterInfo().then(setClusterInfo)
    }
  }, [isOpen])

  const handleCopy = (text: string, sectionId: string) => {
    navigator.clipboard.writeText(text)
    setCopiedSection(sectionId)
    setTimeout(() => setCopiedSection(null), 2000)
  }

  const manualTabs: NavTabItem[] = [
    { id: 'quickstart', label: 'Quick Start', icon: <Terminal className="w-4 h-4" /> },
    { id: 'architecture', label: 'Architecture & Ports', icon: <Network className="w-4 h-4" /> },
    { id: 'agents', label: 'AI Agents & MCP', icon: <Bot className="w-4 h-4" /> },
    { id: 'security', label: 'Rulesets & Signing', icon: <ShieldCheck className="w-4 h-4" /> },
    { id: 'federation', label: 'SSO & SCIM 2.0', icon: <Users className="w-4 h-4" /> },
    { id: 'cli', label: 'CLI Reference', icon: <Cpu className="w-4 h-4" /> },
  ]

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6">
      {/* Backdrop */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-background/80 backdrop-blur-xs"
        onClick={onClose}
      />

      {/* Modal Dialog */}
      <motion.div
        initial={{ opacity: 0, scale: 0.98, y: 10 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.98, y: 10 }}
        className="relative w-full max-w-4xl max-h-[85vh] rounded-lg border border-border bg-card text-card-foreground shadow-2xl z-10 flex flex-col overflow-hidden"
      >
        {/* Header */}
        <div className="px-5 py-3.5 border-b border-border bg-secondary/30 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-md bg-primary/10 border border-primary/20 flex items-center justify-center text-primary">
              <BookOpen className="w-4 h-4" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-sm font-semibold text-foreground">exist.io User Manual & Architecture Guide</h2>
                <Badge variant="outline" className="text-[10px] font-mono">
                  {clusterInfo?.version || 'v0.4.0-enterprise'}
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground">
                Sovereign Git forge, autonomous MCP agent bridge, and SQLite WAL engine
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 text-muted-foreground hover:text-foreground rounded-md hover:bg-secondary/60 transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Tab Navigation */}
        <div className="px-5 border-b border-border bg-secondary/10 shrink-0">
          <UnderlineNav
            tabs={manualTabs}
            activeTab={activeTab}
            onTabChange={setActiveTab}
            className="border-b-0"
          />
        </div>

        {/* Scrollable Content Body */}
        <div className="p-6 overflow-y-auto space-y-6 flex-1 text-xs">
          {/* TAB 1: QUICK START */}
          {activeTab === 'quickstart' && (
            <div className="space-y-5">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Getting Started with exist.io</h3>
                <p className="text-muted-foreground mt-0.5">
                  Follow these steps to connect your Git client, configure authentication, and push your first repository.
                </p>
              </div>

              {/* Step 1 */}
              <div className="p-4 rounded-md border border-border bg-secondary/20 space-y-2.5">
                <div className="flex items-center gap-2">
                  <span className="w-5 h-5 rounded-full bg-primary text-primary-foreground font-bold text-[10px] flex items-center justify-center shrink-0">
                    1
                  </span>
                  <h4 className="text-xs font-semibold text-foreground">Enroll Your SSH Public Key</h4>
                </div>
                <p className="text-muted-foreground leading-relaxed">
                  Generate an Ed25519 key on your local machine if you haven't already:
                </p>
                <div className="relative rounded bg-input border border-border p-2 font-mono text-[11px] text-foreground flex items-center justify-between">
                  <code>ssh-keygen -t ed25519 -C "developer@exist.io"</code>
                  <button
                    onClick={() => handleCopy('ssh-keygen -t ed25519 -C "developer@exist.io"', 'ssh-keygen')}
                    className="p-1 text-muted-foreground hover:text-foreground cursor-pointer"
                  >
                    {copiedSection === 'ssh-keygen' ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
                <p className="text-muted-foreground text-[11px]">
                  Copy the output of <code className="font-mono">cat ~/.ssh/id_ed25519.pub</code> and paste it into your <strong>User Profile &gt; SSH and GPG keys</strong>.
                </p>
              </div>

              {/* Step 2 */}
              <div className="p-4 rounded-md border border-border bg-secondary/20 space-y-2.5">
                <div className="flex items-center gap-2">
                  <span className="w-5 h-5 rounded-full bg-primary text-primary-foreground font-bold text-[10px] flex items-center justify-center shrink-0">
                    2
                  </span>
                  <h4 className="text-xs font-semibold text-foreground">Create or Push a Repository</h4>
                </div>
                <p className="text-muted-foreground">
                  Click the <strong>"New"</strong> button on the home dashboard or run the following Git commands in any existing local project:
                </p>
                <div className="relative rounded bg-input border border-border p-2.5 font-mono text-[11px] text-foreground space-y-1">
                  <div>git remote add origin http://localhost:8080/shiva/my-repo.git</div>
                  <div>git branch -M main</div>
                  <div>git push -u origin main</div>
                  <button
                    onClick={() => handleCopy('git remote add origin http://localhost:8080/shiva/my-repo.git\ngit branch -M main\ngit push -u origin main', 'git-push')}
                    className="absolute top-2 right-2 p-1 text-muted-foreground hover:text-foreground cursor-pointer"
                  >
                    {copiedSection === 'git-push' ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>

              {/* Step 3 */}
              <div className="p-4 rounded-md border border-border bg-secondary/20 space-y-2.5">
                <div className="flex items-center gap-2">
                  <span className="w-5 h-5 rounded-full bg-primary text-primary-foreground font-bold text-[10px] flex items-center justify-center shrink-0">
                    3
                  </span>
                  <h4 className="text-xs font-semibold text-foreground">Connect AI Coding Agents via MCP</h4>
                </div>
                <p className="text-muted-foreground">
                  Point any MCP-compatible agent to your local gateway:
                </p>
                <div className="relative rounded bg-input border border-border p-2 font-mono text-[11px] text-foreground flex items-center justify-between">
                  <code>claude mcp add exist http://localhost:8080/mcp</code>
                  <button
                    onClick={() => handleCopy('claude mcp add exist http://localhost:8080/mcp', 'mcp-cmd')}
                    className="p-1 text-muted-foreground hover:text-foreground cursor-pointer"
                  >
                    {copiedSection === 'mcp-cmd' ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* TAB 2: ARCHITECTURE & PORTS */}
          {activeTab === 'architecture' && (
            <div className="space-y-5">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Cluster Architecture & Port Map</h3>
                <p className="text-muted-foreground mt-0.5">
                  How the unified <code className="font-mono">existd</code> server daemon coordinates network protocols, storage, and agents.
                </p>
              </div>

              {/* Visual ASCII / Box Diagram */}
              <div className="p-4 rounded-lg bg-secondary/30 border border-border font-mono text-[11px] space-y-2 overflow-x-auto text-foreground">
                <div className="text-muted-foreground">// exist.io Unified System Topology</div>
                <div className="whitespace-pre text-xs leading-relaxed">
{`+-------------------------------------------------------------------------+
|                              exist.io Daemon (:8080)                    |
|                                                                         |
|  +--------------------+   +--------------------+   +------------------+ |
|  | Smart-HTTP Server  |   |   Pure-Go SSH      |   | MCP Agent Bridge | |
|  | (Port :8080)       |   |   (Port :2222)     |   | (/mcp Endpoint)  | |
|  +---------+----------+   +---------+----------+   +--------+---------+ |
|            |                        |                       |           |
|            +-------------------+    |    +------------------+           |
|                                |    |    |                              |
|                                v    v    v                              |
|  +--------------------------------------------------------------------+ |
|  |             Policy Engine (Rulesets, Signatures, Grants)           | |
|  +---------------------------------+----------------------------------+ |
|                                    |                                    |
|                                    v                                    |
|  +--------------------------------------------------------------------+ |
|  |    SQLite WAL Database Engine (modernc.org/sqlite, zero-CGO)       | |
|  |    data/meta.db  <--->  Litestream Hot Live Replicas               | |
|  +---------------------------------+----------------------------------+ |
|                                    |                                    |
|                                    v                                    |
|  +--------------------------------------------------------------------+ |
|  |    Bare Git Object Repositories (data/git/{owner}/{repo}.git)      | |
|  +--------------------------------------------------------------------+ |
+-------------------------------------------------------------------------+`}
                </div>
              </div>

              {/* Ports Table */}
              <div className="space-y-2">
                <h4 className="text-xs font-semibold text-foreground">Network Protocols & Default Ports</h4>
                <div className="rounded-md border border-border overflow-hidden">
                  <table className="w-full text-left text-[11px]">
                    <thead className="bg-secondary/40 border-b border-border text-foreground font-semibold">
                      <tr>
                        <th className="p-2.5">Port / Route</th>
                        <th className="p-2.5">Protocol</th>
                        <th className="p-2.5">Function</th>
                        <th className="p-2.5">Auth Method</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      <tr>
                        <td className="p-2.5 font-mono text-primary font-bold">:8080</td>
                        <td className="p-2.5">HTTP/1.1</td>
                        <td className="p-2.5 text-muted-foreground">Git Smart-HTTP transport & Web Dashboard</td>
                        <td className="p-2.5">Session Cookie / PAT</td>
                      </tr>
                      <tr>
                        <td className="p-2.5 font-mono text-primary font-bold">:2222</td>
                        <td className="p-2.5">SSHv2</td>
                        <td className="p-2.5 text-muted-foreground">Pure-Go Git SSH wire transport</td>
                        <td className="p-2.5">Public Key (Ed25519/RSA)</td>
                      </tr>
                      <tr>
                        <td className="p-2.5 font-mono text-primary font-bold">/mcp</td>
                        <td className="p-2.5">JSON-RPC 2.0</td>
                        <td className="p-2.5 text-muted-foreground">Model Context Protocol 35+ Git agent tools</td>
                        <td className="p-2.5">Bearer Token / Agent Lease</td>
                      </tr>
                      <tr>
                        <td className="p-2.5 font-mono text-primary font-bold">/scim/v2</td>
                        <td className="p-2.5">REST / JSON</td>
                        <td className="p-2.5 text-muted-foreground">SCIM 2.0 automated user sync (Okta, Entra)</td>
                        <td className="p-2.5">Bearer Token</td>
                      </tr>
                      <tr>
                        <td className="p-2.5 font-mono text-primary font-bold">/healthz</td>
                        <td className="p-2.5">HTTP GET</td>
                        <td className="p-2.5 text-muted-foreground">Cluster liveness check</td>
                        <td className="p-2.5">Unauthenticated</td>
                      </tr>
                      <tr>
                        <td className="p-2.5 font-mono text-primary font-bold">/metrics</td>
                        <td className="p-2.5">Prometheus</td>
                        <td className="p-2.5 text-muted-foreground">Cluster telemetry & Git transfer throughput</td>
                        <td className="p-2.5">Admin Bearer / Internal</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* TAB 3: AGENTS & MCP */}
          {activeTab === 'agents' && (
            <div className="space-y-5">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Model Context Protocol (MCP) Integration</h3>
                <p className="text-muted-foreground mt-0.5">
                  How autonomous coding agents interface safely with your repositories.
                </p>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="p-3.5 rounded-md border border-border bg-secondary/20 space-y-2">
                  <span className="font-semibold text-foreground">Claude Code Configuration</span>
                  <p className="text-[11px] text-muted-foreground">
                    Add exist.io as an MCP server in your Claude project settings:
                  </p>
                  <pre className="p-2 rounded bg-input border border-border font-mono text-[10px] overflow-x-auto text-foreground">
{`{
  "mcpServers": {
    "exist": {
      "url": "http://localhost:8080/mcp"
    }
  }
}`}
                  </pre>
                </div>

                <div className="p-3.5 rounded-md border border-border bg-secondary/20 space-y-2">
                  <span className="font-semibold text-foreground">Cursor Configuration (.cursor/mcp.json)</span>
                  <p className="text-[11px] text-muted-foreground">
                    Enable repository awareness for Cursor agents:
                  </p>
                  <pre className="p-2 rounded bg-input border border-border font-mono text-[10px] overflow-x-auto text-foreground">
{`{
  "mcpServers": {
    "exist": {
      "command": "curl",
      "args": ["http://localhost:8080/mcp"]
    }
  }
}`}
                  </pre>
                </div>
              </div>

              <div className="space-y-2">
                <h4 className="text-xs font-semibold text-foreground">Key Agent Capabilities (35 Tools)</h4>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
                  <div className="p-2.5 rounded bg-secondary/15 border border-border space-y-1">
                    <span className="font-mono text-primary text-[11px] font-bold">git_diff</span>
                    <p className="text-[10px] text-muted-foreground">Inspect file changes across commits or branches</p>
                  </div>
                  <div className="p-2.5 rounded bg-secondary/15 border border-border space-y-1">
                    <span className="font-mono text-primary text-[11px] font-bold">git_commit</span>
                    <p className="text-[10px] text-muted-foreground">Create signed atomic commits with agent metadata</p>
                  </div>
                  <div className="p-2.5 rounded bg-secondary/15 border border-border space-y-1">
                    <span className="font-mono text-primary text-[11px] font-bold">mcp_review_pr</span>
                    <p className="text-[10px] text-muted-foreground">Perform autonomous AST code review and approve PRs</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* TAB 4: RULES & SIGNING */}
          {activeTab === 'security' && (
            <div className="space-y-5">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Rulesets & Cryptographic Enforcement</h3>
                <p className="text-muted-foreground mt-0.5">
                  Protecting mainlines and verifying cryptographic commit signatures.
                </p>
              </div>

              <div className="p-4 rounded-md border border-border bg-secondary/20 space-y-3">
                <h4 className="text-xs font-semibold text-foreground flex items-center gap-2">
                  <Lock className="w-4 h-4 text-primary" />
                  <span>Configuring Git Commit Signing with SSH</span>
                </h4>
                <p className="text-muted-foreground">
                  You can sign Git commits with your existing SSH key instead of configuring complex GPG setups:
                </p>
                <div className="rounded bg-input border border-border p-2.5 font-mono text-[11px] text-foreground space-y-1">
                  <div>git config --global gpg.format ssh</div>
                  <div>git config --global user.signingkey ~/.ssh/id_ed25519.pub</div>
                  <div>git config --global commit.gpgsign true</div>
                </div>
                <p className="text-[11px] text-muted-foreground">
                  When enabled, all commits pushed to exist.io will display the green <strong>"Verified"</strong> badge in the Commits view and satisfy strict branch protection rulesets.
                </p>
              </div>
            </div>
          )}

          {/* TAB 5: FEDERATION */}
          {activeTab === 'federation' && (
            <div className="space-y-5">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Enterprise Identity Federation</h3>
                <p className="text-muted-foreground mt-0.5">
                  SAML 2.0 / OIDC SSO and SCIM 2.0 automated provisioning.
                </p>
              </div>

              <div className="p-4 rounded-md border border-border bg-secondary/20 space-y-3">
                <div className="flex items-center gap-2">
                  <Users className="w-4 h-4 text-primary" />
                  <h4 className="text-xs font-semibold text-foreground">SCIM 2.0 Directory Sync</h4>
                </div>
                <p className="text-muted-foreground">
                  Connect Okta, Microsoft Entra ID (Azure AD), or OneLogin to automatically sync employee lifecycle events:
                </p>
                <div className="rounded bg-input border border-border p-2.5 font-mono text-[11px] text-foreground space-y-1">
                  <div>Base URL: http://localhost:8080/scim/v2</div>
                  <div>Users Endpoint: http://localhost:8080/scim/v2/Users</div>
                  <div>Authentication: Bearer Token (generated in Enterprise Settings)</div>
                </div>
              </div>
            </div>
          )}

          {/* TAB 6: CLI REFERENCE */}
          {activeTab === 'cli' && (
            <div className="space-y-5">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Command Line Reference</h3>
                <p className="text-muted-foreground mt-0.5">
                  Administrative commands and runtime flags for <code className="font-mono">existd</code> and <code className="font-mono">exist</code> CLI.
                </p>
              </div>

              <div className="space-y-3">
                <div className="p-3 rounded-md border border-border bg-secondary/20 space-y-1.5">
                  <div className="font-mono font-semibold text-primary text-xs">existd bootstrap --email &lt;email&gt; --namespace &lt;slug&gt;</div>
                  <p className="text-muted-foreground">
                    Initializes the sovereign cluster's first root administrator account and creates an initial high-entropy API key.
                  </p>
                </div>

                <div className="p-3 rounded-md border border-border bg-secondary/20 space-y-1.5">
                  <div className="font-mono font-semibold text-primary text-xs">existd hook pre-receive</div>
                  <p className="text-muted-foreground">
                    Native pre-receive hook executed during git push to scan for secrets, verify cryptographic commit signatures, and enforce branch rulesets.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-5 py-3 border-t border-border bg-secondary/20 flex items-center justify-between shrink-0">
          <span className="text-[11px] text-muted-foreground">
            SQLite WAL Mode &bull; Pure-Go SSH &bull; Zero External Dependencies
          </span>
          <Button variant="secondary" size="xs" onClick={onClose}>
            Close
          </Button>
        </div>
      </motion.div>
    </div>
  )
}
