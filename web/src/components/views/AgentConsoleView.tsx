import React, { useState, useEffect } from 'react'
import { motion } from 'motion/react'
import {
  Bot,
  Sparkles,
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
import { Card } from '@/components/ui/Card'
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
    if (toolName === 'list_pull_requests') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "state": "open"\n}')
    } else if (toolName === 'search_symbols') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "query": "Server"\n}')
    } else if (toolName === 'checks_read') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "headSha": "main"\n}')
    } else if (toolName === 'list_rulesets') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket"\n}')
    } else if (toolName === 'list_namespaces') {
      setToolArgsInput('{}')
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

      // Refresh activity list
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
    <div className="space-y-6">
      {/* Top Banner Hero */}
      <Card className="border-sky-500/30 bg-gradient-to-r from-sky-950/40 via-slate-900/60 to-purple-950/40 p-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <motion.div
              animate={{
                boxShadow: [
                  '0 0 15px rgba(56,189,248,0.2)',
                  '0 0 30px rgba(56,189,248,0.4)',
                  '0 0 15px rgba(56,189,248,0.2)',
                ],
              }}
              transition={{ duration: 3, repeat: Infinity }}
              className="flex items-center justify-center w-12 h-12 rounded-2xl bg-sky-500/20 border border-sky-400/40 text-sky-300"
            >
              <Bot className="w-6 h-6 text-sky-400" />
            </motion.div>
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <h2 className="text-xl font-bold text-white tracking-tight">AI Agent Telemetry & Capability Leases</h2>
                <Badge variant="agent" pulse>
                  <Sparkles className="w-3 h-3 mr-1" /> Live Protocol
                </Badge>
              </div>
              <p className="text-xs text-slate-400">
                Direct Model Context Protocol server mounted on <code className="text-sky-300 font-mono">http://localhost:8080/mcp</code> with 35 active tools.
              </p>
            </div>
          </div>

          {/* Quick Metrics */}
          <div className="flex items-center gap-4 text-xs font-mono">
            <div className="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 space-y-0.5">
              <span className="text-[10px] text-slate-500 block uppercase">MCP Tools</span>
              <span className="text-sm font-bold text-sky-400">{availableTools.length || 35} tools</span>
            </div>
            <div className="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 space-y-0.5">
              <span className="text-[10px] text-slate-500 block uppercase">Avg Latency</span>
              <span className="text-sm font-bold text-emerald-400">12ms</span>
            </div>
            <div className="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 space-y-0.5">
              <span className="text-[10px] text-slate-500 block uppercase">Engine State</span>
              <span className="text-sm font-bold text-purple-400">SQLite WAL</span>
            </div>
          </div>
        </div>
      </Card>

      {/* Sub-tabs */}
      <div className="flex items-center gap-2 border-b border-slate-800 pb-2">
        <button
          onClick={() => setActiveTab('activity')}
          className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
            activeTab === 'activity' ? 'bg-sky-500/20 text-sky-300 font-semibold' : 'text-slate-400 hover:text-white'
          }`}
        >
          <Activity className="w-3.5 h-3.5" />
          <span>Real-time Agent Activity</span>
          <span className="text-[10px] bg-slate-800 px-1.5 py-0.2 rounded-full">{activities.length}</span>
        </button>

        <button
          onClick={() => setActiveTab('tester')}
          className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
            activeTab === 'tester' ? 'bg-sky-500/20 text-sky-300 font-semibold' : 'text-slate-400 hover:text-white'
          }`}
        >
          <Terminal className="w-3.5 h-3.5" />
          <span>Live MCP Tool Runner</span>
        </button>

        <button
          onClick={() => setActiveTab('leases')}
          className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
            activeTab === 'leases' ? 'bg-purple-500/20 text-purple-300 font-semibold' : 'text-slate-400 hover:text-white'
          }`}
        >
          <KeyRound className="w-3.5 h-3.5 text-purple-400" />
          <span>Capability Leases</span>
        </button>

        <button
          onClick={() => setActiveTab('catalog')}
          className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
            activeTab === 'catalog' ? 'bg-sky-500/20 text-sky-300 font-semibold' : 'text-slate-400 hover:text-white'
          }`}
        >
          <ShieldCheck className="w-3.5 h-3.5" />
          <span>Tool Catalog</span>
        </button>
      </div>

      {/* Tab 1: Real-time Activity Feed */}
      {activeTab === 'activity' && (
        <div className="space-y-4">
          {activities.length === 0 ? (
            <Card className="border-slate-800 bg-slate-950/60 p-8 text-center space-y-4">
              <div className="w-12 h-12 rounded-2xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center mx-auto text-sky-400">
                <Activity className="w-6 h-6" />
              </div>
              <div className="max-w-md mx-auto space-y-1">
                <h3 className="text-base font-semibold text-slate-100">No agent operations recorded yet</h3>
                <p className="text-xs text-slate-400 leading-relaxed">
                  Autonomous agents invoking tools via the JSON-RPC Model Context Protocol or executing pull request reviews
                  will stream real-time operational telemetry into this dashboard.
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setActiveTab('tester')}
                className="gap-1.5 border-slate-700 text-slate-300"
              >
                <Terminal className="w-3.5 h-3.5 text-sky-400" />
                <span>Test an MCP Tool Now</span>
              </Button>
            </Card>
          ) : (
            <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/50">
              <div className="divide-y divide-slate-850/70">
                {activities.map((act) => (
                  <div key={act.id} className="p-4 hover:bg-slate-900/50 transition-colors space-y-2">
                    <div className="flex items-center justify-between gap-4">
                      <div className="flex items-center gap-3">
                        <LivingAvatar name={act.agentName} isAgent={true} size="sm" />
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-bold text-white">{act.agentName}</span>
                            <span className="text-slate-600 font-mono text-xs">called</span>
                            <span className="font-mono text-xs text-sky-300 bg-sky-950/60 px-2 py-0.5 rounded border border-sky-800/80">
                              {act.toolName}
                            </span>
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center gap-3 text-xs font-mono">
                        <span className="text-emerald-400 flex items-center gap-1">
                          <Clock className="w-3 h-3" /> {act.latencyMs}ms
                        </span>
                        <span className="text-slate-500 text-[11px]">{act.timestamp}</span>
                      </div>
                    </div>

                    <div className="pl-10 text-xs font-mono text-slate-400 bg-slate-950/80 p-2.5 rounded-xl border border-slate-900 space-y-1">
                      <div className="text-slate-500 text-[11px]">
                        Payload: <span className="text-slate-300">{JSON.stringify(act.requestPayload)}</span>
                      </div>
                      <div className="text-sky-400/90 text-[11px]">
                        Result: <span className="text-slate-200">{act.responseSummary}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          )}
        </div>
      )}

      {/* Tab 2: Live Tool Runner */}
      {activeTab === 'tester' && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-3">
            <h3 className="text-sm font-semibold text-slate-200">Select Tool to Test</h3>
            <div className="space-y-1 max-h-96 overflow-y-auto pr-1">
              {availableTools.map((t) => (
                <button
                  key={t.name}
                  onClick={() => handleToolSelect(t.name)}
                  className={`w-full text-left px-3 py-2 rounded-xl text-xs font-mono transition-colors flex items-center justify-between ${
                    selectedToolName === t.name
                      ? 'bg-sky-500/20 text-sky-300 border border-sky-500/30'
                      : 'hover:bg-slate-900 text-slate-400'
                  }`}
                >
                  <span className="truncate">{t.name}</span>
                  <ChevronRight className="w-3 h-3 opacity-50 shrink-0" />
                </button>
              ))}
            </div>
          </Card>

          <Card className="lg:col-span-2 border-slate-800 bg-slate-950/60 p-5 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-semibold text-white font-mono flex items-center gap-2">
                  <Terminal className="w-4 h-4 text-sky-400" />
                  <span>{selectedToolName}</span>
                </h3>
                <span className="text-xs text-slate-400">JSON-RPC 2.0 Payload Tester</span>
              </div>

              <Button
                variant="glow"
                size="sm"
                onClick={handleExecuteTool}
                disabled={executing}
                className="gap-1.5"
              >
                {executing ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Play className="w-3.5 h-3.5" />}
                <span>{executing ? 'Executing...' : 'Run Tool'}</span>
              </Button>
            </div>

            <div className="space-y-2">
              <label className="text-xs font-medium text-slate-400">Tool Arguments (JSON)</label>
              <textarea
                value={toolArgsInput}
                onChange={(e) => setToolArgsInput(e.target.value)}
                rows={6}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 font-mono text-xs text-sky-300 outline-none focus:border-sky-500"
              />
            </div>

            {executionResult && (
              <div className="space-y-2 pt-2 border-t border-slate-800">
                <div className="flex items-center justify-between text-xs">
                  <span className="font-semibold text-slate-200">Execution Response</span>
                  {executionTime && (
                    <span className="text-emerald-400 font-mono">{executionTime}ms</span>
                  )}
                </div>
                <pre className="p-3 bg-slate-950 border border-slate-800 rounded-xl font-mono text-xs text-slate-300 overflow-x-auto max-h-60 leading-relaxed">
                  {JSON.stringify(executionResult, null, 2)}
                </pre>
              </div>
            )}
          </Card>
        </div>
      )}

      {/* Tab 3: Capability Leases */}
      {activeTab === 'leases' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card className="border-slate-800 bg-slate-950/60 p-6 space-y-4">
            <div className="space-y-1">
              <h3 className="text-base font-semibold text-white flex items-center gap-2">
                <KeyRound className="w-4 h-4 text-purple-400" />
                <span>Issue Agent Capability Lease</span>
              </h3>
              <p className="text-xs text-slate-400 leading-relaxed">
                Generates a cryptographically signed HMAC-SHA256 capability token that permits an AI agent to push
                exclusively to authorized ref prefixes (e.g. <code className="text-purple-300 font-mono">refs/heads/agent/*</code>)
                with a guaranteed expiration TTL.
              </p>
            </div>

            <div className="space-y-3 pt-2">
              <div>
                <label className="text-xs font-medium text-slate-300">Agent Identifier</label>
                <input
                  type="text"
                  value={leaseAgentId}
                  onChange={(e) => setLeaseAgentId(e.target.value)}
                  className="mt-1 w-full bg-slate-900 border border-slate-800 rounded-xl p-2.5 text-xs text-slate-200 font-mono outline-none focus:border-purple-500"
                />
              </div>

              <div>
                <label className="text-xs font-medium text-slate-300">Target Repository</label>
                <input
                  type="text"
                  value={leaseRepoId}
                  onChange={(e) => setLeaseRepoId(e.target.value)}
                  className="mt-1 w-full bg-slate-900 border border-slate-800 rounded-xl p-2.5 text-xs text-slate-200 font-mono outline-none focus:border-purple-500"
                />
              </div>

              <div>
                <label className="text-xs font-medium text-slate-300">Allowed Ref Prefix</label>
                <input
                  type="text"
                  value={leaseRefPrefix}
                  onChange={(e) => setLeaseRefPrefix(e.target.value)}
                  className="mt-1 w-full bg-slate-900 border border-slate-800 rounded-xl p-2.5 text-xs text-slate-200 font-mono outline-none focus:border-purple-500"
                />
              </div>

              <div>
                <label className="text-xs font-medium text-slate-300">TTL Duration (Minutes)</label>
                <input
                  type="number"
                  value={leaseTtlMinutes}
                  onChange={(e) => setLeaseTtlMinutes(parseInt(e.target.value, 10) || 15)}
                  className="mt-1 w-full bg-slate-900 border border-slate-800 rounded-xl p-2.5 text-xs text-slate-200 font-mono outline-none focus:border-purple-500"
                />
              </div>

              <div className="pt-2">
                <Button
                  variant="glow"
                  size="sm"
                  onClick={handleGenerateLease}
                  disabled={generatingLease}
                  className="w-full gap-1.5 bg-gradient-to-r from-purple-600 to-indigo-600"
                >
                  <Lock className="w-3.5 h-3.5" />
                  <span>{generatingLease ? 'Signing Lease...' : 'Sign & Issue Capability Lease'}</span>
                </Button>
              </div>
            </div>
          </Card>

          <Card className="border-slate-800 bg-slate-950/60 p-6 space-y-4">
            <h3 className="text-base font-semibold text-white">Active Lease Credential</h3>
            {issuedLease ? (
              <div className="space-y-4">
                <div className="p-3 bg-purple-950/30 border border-purple-800/60 rounded-xl space-y-2">
                  <div className="flex items-center justify-between text-xs font-mono text-purple-300">
                    <span>Lease ID: {issuedLease.id}</span>
                    <Badge variant="outline" className="text-[10px] text-purple-300 border-purple-500/40">Active</Badge>
                  </div>
                  <div className="text-[11px] text-slate-400 font-mono">
                    Expires: {new Date(issuedLease.expiresAt).toLocaleTimeString()} ({issuedLease.expiresAt})
                  </div>
                </div>

                <div className="space-y-1.5">
                  <div className="flex items-center justify-between text-xs text-slate-400">
                    <span>Signed Capability Token</span>
                    <button
                      onClick={handleCopyToken}
                      className="hover:text-white transition-colors flex items-center gap-1 text-[11px] text-sky-400"
                    >
                      {copiedToken ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                      <span>{copiedToken ? 'Copied' : 'Copy Token'}</span>
                    </button>
                  </div>
                  <pre className="p-3 bg-slate-950 border border-slate-800 rounded-xl text-xs font-mono text-slate-300 overflow-x-auto break-all whitespace-pre-wrap">
                    {issuedLease.token}
                  </pre>
                </div>

                <div className="p-3 bg-slate-900 border border-slate-800 rounded-xl space-y-1 text-xs text-slate-400">
                  <span className="font-semibold text-slate-200">How to use:</span>
                  <p className="text-[11px] leading-relaxed">
                    Pass in agent HTTP requests via header:
                    <br />
                    <code className="text-sky-300 font-mono">Authorization: Bearer {issuedLease.token.slice(0, 16)}...</code>
                  </p>
                </div>
              </div>
            ) : (
              <div className="p-8 text-center text-slate-500 text-xs font-mono">
                No active capability lease issued in this session. Configure the parameters and click "Sign & Issue" to generate one.
              </div>
            )}
          </Card>
        </div>
      )}

      {/* Tab 4: 35-Tool Catalog */}
      {activeTab === 'catalog' && (
        <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/50">
          <div className="grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-slate-850">
            <div className="p-4 space-y-3">
              <h4 className="text-xs font-semibold text-sky-400 uppercase tracking-wider">Repository & Tree Tools</h4>
              <div className="space-y-2 text-xs">
                {availableTools.slice(0, Math.ceil(availableTools.length / 2)).map((t) => (
                  <div key={t.name} className="p-2.5 rounded-xl bg-slate-900/60 border border-slate-800/60 space-y-1">
                    <span className="font-mono font-semibold text-white">{t.name}</span>
                    <p className="text-slate-400 text-[11px]">{t.description}</p>
                  </div>
                ))}
              </div>
            </div>

            <div className="p-4 space-y-3">
              <h4 className="text-xs font-semibold text-purple-400 uppercase tracking-wider">Policy & Review Tools</h4>
              <div className="space-y-2 text-xs">
                {availableTools.slice(Math.ceil(availableTools.length / 2)).map((t) => (
                  <div key={t.name} className="p-2.5 rounded-xl bg-slate-900/60 border border-slate-800/60 space-y-1">
                    <span className="font-mono font-semibold text-white">{t.name}</span>
                    <p className="text-slate-400 text-[11px]">{t.description}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </Card>
      )}
    </div>
  )
}
