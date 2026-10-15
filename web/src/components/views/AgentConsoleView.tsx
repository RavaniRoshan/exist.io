import React, { useState, useEffect } from 'react'
import {
  Bot,
  Terminal,
  Activity,
  Play,
  Clock,
  ShieldCheck,
  ChevronRight,
  KeyRound,
  Copy,
  Check,
  RefreshCw,
  Lock,
} from 'lucide-react'
import { Box, BoxHeader, BoxBody, BoxRow } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import {
  getAgentActivities,
  issueAgentLease,
  queryLocalMcpTools,
  isDemoModeActive,
  sampleAgentActivities,
  type AgentActivityItem,
  type AgentLease,
} from '@/lib/api'

export const AgentConsoleView: React.FC = () => {
  const [activities, setActivities] = useState<AgentActivityItem[]>([])
  const [, setLoadingActivities] = useState(true)
  const [availableTools, setAvailableTools] = useState<any[]>([])
  const [selectedToolName, setSelectedToolName] = useState('list_pull_requests')
  const [toolArgsInput, setToolArgsInput] = useState(
    '{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "state": "open"\n}'
  )
  const [executing, setExecuting] = useState(false)
  const [executionResult, setExecutionResult] = useState<any | null>(null)
  const [executionTime, setExecutionTime] = useState<number | null>(null)
  const [activeTab, setActiveTab] = useState<'activity' | 'tester' | 'catalog' | 'leases'>('activity')

  // Lease Generator State
  const [leaseAgentId, setLeaseAgentId] = useState('agent:cursor-worker')
  const [leaseRepoId, setLeaseRepoId] = useState('acme/rocket')
  const [leaseRefPrefix, setLeaseRefPrefix] = useState('refs/heads/agent/')
  const [leaseTtlMinutes, setLeaseTtlMinutes] = useState(30)
  const [generatingLease, setGeneratingLease] = useState(false)
  const [issuedLease, setIssuedLease] = useState<AgentLease | null>(null)
  const [copiedToken, setCopiedToken] = useState(false)

  const loadActivities = async () => {
    try {
      const items = await getAgentActivities()
      setActivities(items)
    } catch {
      if (isDemoModeActive()) {
        setActivities(sampleAgentActivities)
      } else {
        setActivities([])
      }
    } finally {
      setLoadingActivities(false)
    }
  }

  useEffect(() => {
    loadActivities()
    const interval = setInterval(loadActivities, 4000)

    const onModeChange = () => loadActivities()
    window.addEventListener('exist_mode_changed', onModeChange)

    return () => {
      clearInterval(interval)
      window.removeEventListener('exist_mode_changed', onModeChange)
    }
  }, [])

  useEffect(() => {
    async function loadMcpCatalog() {
      const tools = await queryLocalMcpTools()
      if (tools.length > 0) {
        setAvailableTools(tools)
      } else {
        setAvailableTools([
          { name: 'list_pull_requests', description: 'List pull requests in repository' },
          { name: 'get_pull_request', description: 'Get details for a pull request' },
          { name: 'merge_pull_request', description: 'Merge a pull request into its base branch' },
          { name: 'search_symbols', description: 'Semantic code symbol search across AST' },
          { name: 'create_pull_request_inline_comment', description: 'Start review thread on file line' },
          { name: 'checks_read', description: 'Read check runs and suites for a commit SHA' },
          { name: 'list_rulesets', description: 'List rulesets configured on a repository' },
          { name: 'list_namespaces', description: 'List all namespaces in exist.io' },
        ])
      }
    }
    loadMcpCatalog()
  }, [])

  const handleToolSelect = (toolName: string) => {
    setSelectedToolName(toolName)
    setExecutionResult(null)
    setExecutionTime(null)

    if (toolName === 'list_pull_requests') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "state": "open"\n}')
    } else if (toolName === 'get_pull_request') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "number": 104\n}')
    } else if (toolName === 'merge_pull_request') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "number": 104,\n  "method": "squash"\n}')
    } else if (toolName === 'search_symbols') {
      setToolArgsInput('{\n  "query": "Lease"\n}')
    } else {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket"\n}')
    }
  }

  const handleExecuteTool = async () => {
    setExecuting(true)
    setExecutionResult(null)
    const startTime = performance.now()

    try {
      let parsedArgs = {}
      try {
        parsedArgs = JSON.parse(toolArgsInput)
      } catch (e: any) {
        throw new Error('Invalid JSON arguments: ' + e.message)
      }

      const response = await fetch('/mcp', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: Date.now(),
          method: 'tools/call',
          params: {
            name: selectedToolName,
            arguments: parsedArgs,
          },
        }),
      })

      const elapsed = Math.round(performance.now() - startTime)
      setExecutionTime(elapsed)

      if (!response.ok) {
        throw new Error(`Server returned HTTP ${response.status}`)
      }

      const resData = await response.json()
      setExecutionResult(resData)

      setTimeout(loadActivities, 500)
    } catch (err: any) {
      setExecutionResult({ error: err.message || 'Execution failed' })
    } finally {
      setExecuting(false)
    }
  }

  const handleGenerateLease = async () => {
    setGeneratingLease(true)
    try {
      const lease = await issueAgentLease(
        leaseAgentId,
        leaseRepoId,
        [leaseRefPrefix, 'refs/heads/agent/*'],
        leaseTtlMinutes * 60
      )
      setIssuedLease(lease)
    } catch (err: any) {
      alert('Failed to issue capability lease: ' + err.message)
    } finally {
      setGeneratingLease(false)
    }
  }

  const handleCopyToken = () => {
    if (!issuedLease) return
    navigator.clipboard.writeText(issuedLease.token)
    setCopiedToken(true)
    setTimeout(() => setCopiedToken(false), 2000)
  }

  return (
    <div className="space-y-4">
      {/* Top Banner Hero */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <div className="flex items-center gap-2">
            <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
              <Bot className="w-4 h-4 text-primary" />
              <span>AI Agent Telemetry & Capability Leases</span>
            </h2>
            <Badge variant="agent" pulse>
              Live Protocol
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground">
            Direct Model Context Protocol server mounted on <code className="text-primary font-mono">http://localhost:8080/mcp</code> with 35 active tools.
          </p>
        </div>

        {/* Quick Metrics */}
        <div className="flex items-center gap-3 text-xs font-mono">
          <div className="px-2.5 py-1 rounded-md bg-secondary border border-border">
            <span className="text-muted-foreground mr-1">Tools:</span>
            <span className="font-semibold text-foreground">{availableTools.length || 35}</span>
          </div>
          <div className="px-2.5 py-1 rounded-md bg-secondary border border-border">
            <span className="text-muted-foreground mr-1">Avg Latency:</span>
            <span className="font-semibold text-emerald-500">12ms</span>
          </div>
        </div>
      </div>

      {/* Sub-Navigation UnderlineNav */}
      <div className="flex items-center gap-2 border-b border-border -mb-px text-xs font-medium">
        <button
          onClick={() => setActiveTab('activity')}
          className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors cursor-pointer ${
            activeTab === 'activity'
              ? 'border-primary text-foreground font-semibold'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <Activity className="w-3.5 h-3.5" />
          <span>Real-time Activity</span>
          <span className="text-[10px] bg-secondary text-secondary-foreground px-1.5 py-0.2 rounded-full font-mono">
            {activities.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab('tester')}
          className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors cursor-pointer ${
            activeTab === 'tester'
              ? 'border-primary text-foreground font-semibold'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <Terminal className="w-3.5 h-3.5" />
          <span>Live MCP Tool Runner</span>
        </button>

        <button
          onClick={() => setActiveTab('leases')}
          className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors cursor-pointer ${
            activeTab === 'leases'
              ? 'border-primary text-foreground font-semibold'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <KeyRound className="w-3.5 h-3.5" />
          <span>Capability Leases</span>
        </button>

        <button
          onClick={() => setActiveTab('catalog')}
          className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors cursor-pointer ${
            activeTab === 'catalog'
              ? 'border-primary text-foreground font-semibold'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <ShieldCheck className="w-3.5 h-3.5" />
          <span>Tool Catalog</span>
        </button>
      </div>

      {/* Tab 1: Real-time Activity Feed */}
      {activeTab === 'activity' && (
        <Box>
          <BoxHeader className="bg-secondary/35 py-2.5">
            <span className="font-semibold text-foreground text-xs">Autonomous Agent Invocations</span>
            <span className="text-muted-foreground font-mono text-[11px]">{activities.length} events</span>
          </BoxHeader>

          <BoxBody className="p-0">
            {activities.length === 0 ? (
              <div className="p-8 text-center space-y-3">
                <div className="w-10 h-10 rounded-md bg-secondary border border-border flex items-center justify-center mx-auto text-primary">
                  <Activity className="w-5 h-5" />
                </div>
                <div className="max-w-md mx-auto space-y-1">
                  <h3 className="text-sm font-semibold text-foreground">No agent operations recorded yet</h3>
                  <p className="text-xs text-muted-foreground leading-relaxed">
                    Autonomous agents invoking tools via the JSON-RPC Model Context Protocol or executing pull request reviews
                    will stream real-time operational telemetry into this dashboard.
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setActiveTab('tester')}
                  className="gap-1.5 text-xs"
                >
                  <Terminal className="w-3.5 h-3.5 text-primary" />
                  <span>Test an MCP Tool Now</span>
                </Button>
              </div>
            ) : (
              activities.map((act) => (
                <BoxRow key={act.id} className="p-3.5 hover:bg-secondary/20 transition-colors space-y-2">
                  <div className="flex items-center justify-between gap-4">
                    <div className="flex items-center gap-2.5">
                      <LivingAvatar name={act.agentName} isAgent={true} size="sm" />
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-semibold text-foreground">{act.agentName}</span>
                        <span className="text-muted-foreground font-mono text-xs">called</span>
                        <span className="font-mono text-xs text-primary bg-secondary px-2 py-0.5 rounded border border-border">
                          {act.toolName}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-3 text-xs font-mono">
                      <span className="text-emerald-500 flex items-center gap-1">
                        <Clock className="w-3 h-3" /> {act.latencyMs}ms
                      </span>
                      <span className="text-muted-foreground text-[11px]">{act.timestamp}</span>
                    </div>
                  </div>

                  <div className="text-xs font-mono text-muted-foreground bg-secondary/30 p-2.5 rounded-md border border-border space-y-1">
                    <div className="text-[11px]">
                      Payload: <span className="text-foreground">{JSON.stringify(act.requestPayload)}</span>
                    </div>
                    <div className="text-[11px]">
                      Result: <span className="text-foreground">{act.responseSummary}</span>
                    </div>
                  </div>
                </BoxRow>
              ))
            )}
          </BoxBody>
        </Box>
      )}

      {/* Tab 2: Live Tool Runner */}
      {activeTab === 'tester' && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          <Box className="p-3 space-y-2">
            <h3 className="text-xs font-semibold text-foreground">Select MCP Tool</h3>
            <div className="space-y-1 max-h-96 overflow-y-auto pr-1">
              {availableTools.map((t) => (
                <button
                  key={t.name}
                  onClick={() => handleToolSelect(t.name)}
                  className={`w-full text-left px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors flex items-center justify-between ${
                    selectedToolName === t.name
                      ? 'bg-secondary text-foreground font-semibold border border-primary/40'
                      : 'hover:bg-secondary/40 text-muted-foreground'
                  }`}
                >
                  <span className="truncate">{t.name}</span>
                  <ChevronRight className="w-3 h-3 opacity-50 shrink-0" />
                </button>
              ))}
            </div>
          </Box>

          <Box className="lg:col-span-2">
            <BoxHeader className="bg-secondary/35 py-2">
              <div>
                <h3 className="text-xs font-semibold text-foreground font-mono flex items-center gap-1.5">
                  <Terminal className="w-3.5 h-3.5 text-primary" />
                  <span>{selectedToolName}</span>
                </h3>
              </div>

              <Button
                variant="primary"
                size="sm"
                onClick={handleExecuteTool}
                disabled={executing}
                className="gap-1.5 text-xs"
              >
                {executing ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Play className="w-3.5 h-3.5" />}
                <span>{executing ? 'Executing...' : 'Run Tool'}</span>
              </Button>
            </BoxHeader>

            <BoxBody className="p-3.5 space-y-3">
              <div className="space-y-1">
                <label className="text-xs font-medium text-foreground">Tool Arguments (JSON)</label>
                <textarea
                  value={toolArgsInput}
                  onChange={(e) => setToolArgsInput(e.target.value)}
                  rows={6}
                  className="w-full bg-input border border-border rounded-md p-2.5 font-mono text-xs text-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                />
              </div>

              {executionResult && (
                <div className="space-y-1.5 pt-2 border-t border-border">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-semibold text-foreground">Execution Response</span>
                    {executionTime && (
                      <span className="text-emerald-500 font-mono">{executionTime}ms</span>
                    )}
                  </div>
                  <pre className="p-2.5 bg-background border border-border rounded-md font-mono text-xs text-foreground overflow-x-auto max-h-60 leading-relaxed">
                    {JSON.stringify(executionResult, null, 2)}
                  </pre>
                </div>
              )}
            </BoxBody>
          </Box>
        </div>
      )}

      {/* Tab 3: Capability Leases */}
      {activeTab === 'leases' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <Box>
            <BoxHeader className="bg-secondary/35 py-2.5">
              <div className="flex items-center gap-2">
                <KeyRound className="w-4 h-4 text-primary" />
                <span className="text-xs font-semibold text-foreground">Issue Agent Capability Lease</span>
              </div>
            </BoxHeader>

            <BoxBody className="p-4 space-y-3">
              <p className="text-xs text-muted-foreground leading-relaxed">
                Generates a cryptographically signed HMAC-SHA256 capability token that permits an AI agent to push
                exclusively to authorized ref prefixes (e.g. <code className="font-mono text-foreground">refs/heads/agent/*</code>)
                with a guaranteed expiration TTL.
              </p>

              <div className="space-y-2.5 pt-1">
                <div className="space-y-1">
                  <label className="text-xs font-medium text-foreground">Agent Identifier</label>
                  <input
                    type="text"
                    value={leaseAgentId}
                    onChange={(e) => setLeaseAgentId(e.target.value)}
                    className="w-full bg-input border border-border rounded-md p-2 text-xs text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                  />
                </div>

                <div className="space-y-1">
                  <label className="text-xs font-medium text-foreground">Target Repository</label>
                  <input
                    type="text"
                    value={leaseRepoId}
                    onChange={(e) => setLeaseRepoId(e.target.value)}
                    className="w-full bg-input border border-border rounded-md p-2 text-xs text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                  />
                </div>

                <div className="space-y-1">
                  <label className="text-xs font-medium text-foreground">Allowed Ref Prefix</label>
                  <input
                    type="text"
                    value={leaseRefPrefix}
                    onChange={(e) => setLeaseRefPrefix(e.target.value)}
                    className="w-full bg-input border border-border rounded-md p-2 text-xs text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                  />
                </div>

                <div className="space-y-1">
                  <label className="text-xs font-medium text-foreground">TTL Duration (Minutes)</label>
                  <input
                    type="number"
                    value={leaseTtlMinutes}
                    onChange={(e) => setLeaseTtlMinutes(parseInt(e.target.value, 10) || 15)}
                    className="w-full bg-input border border-border rounded-md p-2 text-xs text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                  />
                </div>

                <div className="pt-2">
                  <Button
                    variant="primary"
                    size="sm"
                    onClick={handleGenerateLease}
                    disabled={generatingLease}
                    className="w-full gap-1.5 text-xs"
                  >
                    <Lock className="w-3.5 h-3.5" />
                    <span>{generatingLease ? 'Signing Lease...' : 'Sign & Issue Capability Lease'}</span>
                  </Button>
                </div>
              </div>
            </BoxBody>
          </Box>

          <Box>
            <BoxHeader className="bg-secondary/35 py-2.5">
              <span className="text-xs font-semibold text-foreground">Active Lease Credential</span>
            </BoxHeader>

            <BoxBody className="p-4 space-y-3">
              {issuedLease ? (
                <div className="space-y-3">
                  <div className="p-2.5 bg-secondary/40 border border-border rounded-md space-y-1">
                    <div className="flex items-center justify-between text-xs font-mono text-foreground">
                      <span>Lease ID: {issuedLease.id}</span>
                      <Badge variant="success" className="text-[10px]">Active</Badge>
                    </div>
                    <div className="text-[11px] text-muted-foreground font-mono">
                      Expires: {new Date(issuedLease.expiresAt).toLocaleTimeString()} ({issuedLease.expiresAt})
                    </div>
                  </div>

                  <div className="space-y-1">
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <span>Signed Capability Token</span>
                      <button
                        onClick={handleCopyToken}
                        className="hover:text-foreground transition-colors flex items-center gap-1 text-[11px] text-primary cursor-pointer"
                      >
                        {copiedToken ? <Check className="w-3 h-3 text-emerald-500" /> : <Copy className="w-3 h-3" />}
                        <span>{copiedToken ? 'Copied' : 'Copy Token'}</span>
                      </button>
                    </div>
                    <pre className="p-2.5 bg-background border border-border rounded-md text-xs font-mono text-foreground overflow-x-auto break-all whitespace-pre-wrap">
                      {issuedLease.token}
                    </pre>
                  </div>

                  <div className="p-2.5 bg-secondary/30 border border-border rounded-md space-y-1 text-xs text-muted-foreground">
                    <span className="font-semibold text-foreground">How to use:</span>
                    <p className="text-[11px] leading-relaxed">
                      Pass in agent HTTP requests via header:
                      <br />
                      <code className="text-primary font-mono">Authorization: Bearer {issuedLease.token.slice(0, 16)}...</code>
                    </p>
                  </div>
                </div>
              ) : (
                <div className="p-8 text-center text-muted-foreground text-xs font-mono">
                  No active capability lease issued in this session. Configure the parameters and click "Sign & Issue" to generate one.
                </div>
              )}
            </BoxBody>
          </Box>
        </div>
      )}

      {/* Tab 4: 35-Tool Catalog */}
      {activeTab === 'catalog' && (
        <Box>
          <BoxHeader className="bg-secondary/35 py-2.5">
            <span className="text-xs font-semibold text-foreground">Model Context Protocol Tools (35 Tools)</span>
          </BoxHeader>

          <BoxBody className="p-0">
            <div className="grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-border">
              <div className="p-3.5 space-y-2">
                <h4 className="text-xs font-semibold text-foreground uppercase tracking-wider">Repository & Tree Tools</h4>
                <div className="space-y-1.5 text-xs">
                  {availableTools.slice(0, Math.ceil(availableTools.length / 2)).map((t) => (
                    <div key={t.name} className="p-2 rounded-md bg-secondary/30 border border-border space-y-0.5">
                      <span className="font-mono font-semibold text-foreground text-xs">{t.name}</span>
                      <p className="text-muted-foreground text-[11px]">{t.description}</p>
                    </div>
                  ))}
                </div>
              </div>

              <div className="p-3.5 space-y-2">
                <h4 className="text-xs font-semibold text-foreground uppercase tracking-wider">Policy & Review Tools</h4>
                <div className="space-y-1.5 text-xs">
                  {availableTools.slice(Math.ceil(availableTools.length / 2)).map((t) => (
                    <div key={t.name} className="p-2 rounded-md bg-secondary/30 border border-border space-y-0.5">
                      <span className="font-mono font-semibold text-foreground text-xs">{t.name}</span>
                      <p className="text-muted-foreground text-[11px]">{t.description}</p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </BoxBody>
        </Box>
      )}
    </div>
  )
}
