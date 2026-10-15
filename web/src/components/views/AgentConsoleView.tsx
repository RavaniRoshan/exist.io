import React, { useState, useEffect } from 'react'
import { motion } from 'motion/react'
import { Bot, Sparkles, Terminal, Activity, Play, Clock, ShieldCheck, ChevronRight } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { sampleAgentActivities, type AgentActivityItem, queryLocalMcpTools } from '@/lib/api'

export const AgentConsoleView: React.FC = () => {
  const [activities, setActivities] = useState<AgentActivityItem[]>(sampleAgentActivities)
  const [availableTools, setAvailableTools] = useState<any[]>([])
  const [selectedToolName, setSelectedToolName] = useState('list_pull_requests')
  const [toolArgsInput, setToolArgsInput] = useState('{\n  "ownerSlug": "exist",\n  "repoName": "exist.io",\n  "state": "open"\n}')
  const [executing, setExecuting] = useState(false)
  const [executionResult, setExecutionResult] = useState<any | null>(null)
  const [executionTime, setExecutionTime] = useState<number | null>(null)
  const [activeTab, setActiveTab] = useState<'activity' | 'tester' | 'catalog'>('activity')

  useEffect(() => {
    async function loadMcpCatalog() {
      const tools = await queryLocalMcpTools()
      if (tools.length > 0) {
        setAvailableTools(tools)
      } else {
        // Fallback default MCP tool list
        setAvailableTools([
          { name: 'list_pull_requests', description: 'List pull requests in repository' },
          { name: 'get_pull_request', description: 'Get details for a pull request' },
          { name: 'merge_pull_request', description: 'Merge a pull request into its base branch' },
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
      setToolArgsInput('{\n  "ownerSlug": "exist",\n  "repoName": "exist.io",\n  "state": "open"\n}')
    } else if (toolName === 'checks_read') {
      setToolArgsInput('{\n  "ownerSlug": "exist",\n  "repoName": "exist.io",\n  "headSha": "7aa8446"\n}')
    } else if (toolName === 'list_rulesets') {
      setToolArgsInput('{\n  "ownerSlug": "exist",\n  "repoName": "exist.io"\n}')
    } else if (toolName === 'list_namespaces') {
      setToolArgsInput('{}')
    } else {
      setToolArgsInput('{\n  "ownerSlug": "exist",\n  "repoName": "exist.io"\n}')
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

      // Live JSON-RPC call to existd daemon
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

      // Add to live activity feed
      const newAct: AgentActivityItem = {
        id: `act-${Date.now()}`,
        agentName: 'MCP Console Tester',
        agentType: 'exist-mcp',
        toolName: selectedToolName,
        status: 'success',
        latencyMs: elapsed,
        timestamp: 'Just now',
        requestPayload: parsedArgs,
        responseSummary: resData.error ? `Error: ${resData.error.message}` : `Success (${elapsed}ms)`,
      }
      setActivities((prev) => [newAct, ...prev])
    } catch (err: any) {
      const elapsed = Math.round(performance.now() - startTime)
      setExecutionTime(elapsed)
      setExecutionResult({
        error: { message: err.message || 'Execution failed' },
      })
    } finally {
      setExecuting(false)
    }
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
                <h2 className="text-xl font-bold text-white tracking-tight">AI Agent Telemetry & MCP Engine</h2>
                <Badge variant="agent" pulse>
                  <Sparkles className="w-3 h-3 mr-1" /> Live Protocol
                </Badge>
              </div>
              <p className="text-xs text-slate-400">
                Direct Model Context Protocol server mounted on <code className="text-sky-300 font-mono">http://localhost:8080/mcp</code> with 34 active tools.
              </p>
            </div>
          </div>

          {/* Quick Metrics */}
          <div className="flex items-center gap-4 text-xs font-mono">
            <div className="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 space-y-0.5">
              <span className="text-[10px] text-slate-500 block uppercase">MCP Tools</span>
              <span className="text-sm font-bold text-sky-400">{availableTools.length || 34} tools</span>
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
          onClick={() => setActiveTab('catalog')}
          className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
            activeTab === 'catalog' ? 'bg-sky-500/20 text-sky-300 font-semibold' : 'text-slate-400 hover:text-white'
          }`}
        >
          <ShieldCheck className="w-3.5 h-3.5" />
          <span>34-Tool Catalog</span>
        </button>
      </div>

      {/* Tab: Real-time Activity Feed */}
      {activeTab === 'activity' && (
        <div className="space-y-4">
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
        </div>
      )}

      {/* Tab: Interactive Tool Runner */}
      {activeTab === 'tester' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Tool Selector & Arguments */}
          <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
            <CardHeader className="p-0">
              <CardTitle className="text-base text-white">Execute MCP Tool</CardTitle>
              <CardDescription>
                Test any Model Context Protocol tool directly against the local existd daemon.
              </CardDescription>
            </CardHeader>

            <div className="space-y-3">
              <label className="text-xs font-medium text-slate-300 block">Select MCP Tool</label>
              <select
                value={selectedToolName}
                onChange={(e) => handleToolSelect(e.target.value)}
                className="w-full bg-slate-900 text-slate-200 text-xs rounded-xl border border-slate-800 p-2.5 outline-none focus:border-sky-500/70"
              >
                {availableTools.map((t) => (
                  <option key={t.name} value={t.name}>
                    {t.name}
                  </option>
                ))}
              </select>

              <label className="text-xs font-medium text-slate-300 block">JSON Arguments</label>
              <textarea
                value={toolArgsInput}
                onChange={(e) => setToolArgsInput(e.target.value)}
                rows={6}
                className="w-full bg-slate-950 font-mono text-xs text-sky-200 p-3 rounded-xl border border-slate-800 outline-none focus:border-sky-500/70 resize-y"
              />

              <Button
                variant="glow"
                size="md"
                onClick={handleExecuteTool}
                loading={executing}
                className="w-full gap-2"
              >
                <Play className="w-4 h-4 fill-current" />
                <span>Execute via MCP Engine</span>
              </Button>
            </div>
          </Card>

          {/* Execution Response Inspector */}
          <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
            <CardHeader className="p-0 flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-base text-white">Execution Response</CardTitle>
                <CardDescription>Real-time JSON-RPC 2.0 output</CardDescription>
              </div>
              {executionTime !== null && (
                <Badge variant="success" className="text-[10px] py-0 font-mono">
                  {executionTime}ms
                </Badge>
              )}
            </CardHeader>

            <div className="h-[280px] bg-slate-950 rounded-xl border border-slate-800 p-4 font-mono text-xs overflow-auto text-slate-200">
              {executing ? (
                <div className="h-full flex items-center justify-center text-slate-500 gap-2">
                  <Activity className="w-4 h-4 animate-spin text-sky-400" />
                  <span>Dispatching JSON-RPC call...</span>
                </div>
              ) : executionResult ? (
                <pre className="text-sky-300 whitespace-pre-wrap">
                  {JSON.stringify(executionResult, null, 2)}
                </pre>
              ) : (
                <div className="h-full flex items-center justify-center text-slate-600 text-xs">
                  Run a tool on the left to inspect live protocol responses.
                </div>
              )}
            </div>
          </Card>
        </div>
      )}

      {/* Tab: 34 Tools Catalog */}
      {activeTab === 'catalog' && (
        <Card className="border-slate-800 bg-slate-950/60 p-5">
          <CardHeader className="p-0 pb-4">
            <CardTitle className="text-base text-white">Complete MCP Forge Tool Catalog</CardTitle>
            <CardDescription>
              All 34 tools exposed to autonomous coding agents across repositories, pull requests, reviews, checks, and rulesets.
            </CardDescription>
          </CardHeader>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {availableTools.map((t) => (
              <div
                key={t.name}
                onClick={() => {
                  handleToolSelect(t.name)
                  setActiveTab('tester')
                }}
                className="p-3 rounded-xl bg-slate-900/60 border border-slate-800 hover:border-sky-500/40 hover:bg-slate-900 transition-all cursor-pointer group"
              >
                <div className="flex items-center justify-between">
                  <span className="font-mono text-xs font-semibold text-sky-300 group-hover:text-sky-400">
                    {t.name}
                  </span>
                  <ChevronRight className="w-3.5 h-3.5 text-slate-600 group-hover:text-sky-400 transition-colors" />
                </div>
                <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
                  {t.description || 'MCP protocol capability'}
                </p>
              </div>
            ))}
          </div>
        </Card>
      )}
    </div>
  )
}
