import React, { useState, useEffect } from 'react'
import { ShieldAlert, Download, Search, AlertCircle, Info, Bot, User, Terminal } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { getAuditEvents, type AuditEvent } from '@/lib/api'

export const AuditLogs: React.FC = () => {
  const [events, setEvents] = useState<AuditEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [severityFilter, setSeverityFilter] = useState<'all' | 'info' | 'warning' | 'critical'>('all')

  useEffect(() => {
    async function load() {
      setLoading(true)
      try {
        const data = await getAuditEvents()
        setEvents(data)
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  const filteredEvents = events.filter((ev) => {
    if (severityFilter !== 'all' && ev.severity !== severityFilter) return false
    if (!searchQuery) return true
    const q = searchQuery.toLowerCase()
    return (
      ev.action.toLowerCase().includes(q) ||
      ev.actor.toLowerCase().includes(q) ||
      ev.resource.toLowerCase().includes(q) ||
      ev.clientIp.toLowerCase().includes(q) ||
      ev.details.toLowerCase().includes(q)
    )
  })

  const exportToJson = () => {
    const blob = new Blob([JSON.stringify(filteredEvents, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `exist-audit-log-${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const exportToCsv = () => {
    const headers = ['ID', 'Timestamp', 'Actor', 'ActorType', 'Action', 'Resource', 'ClientIP', 'Severity', 'Details']
    const rows = filteredEvents.map((e) => [
      e.id,
      e.timestamp,
      e.actor,
      e.actorType,
      e.action,
      e.resource,
      e.clientIp,
      e.severity,
      `"${e.details.replace(/"/g, '""')}"`,
    ])
    const csvContent = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n')
    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `exist-audit-log-${Date.now()}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
      {/* Header & Export Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-5 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md">
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center w-11 h-11 rounded-xl bg-purple-500/10 border border-purple-500/30 text-purple-400">
            <ShieldAlert className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-base font-semibold text-white">Security & Compliance Audit Log</h3>
            <p className="text-xs text-slate-400">
              Immutable audit trail of Git pushes, access grants, ruleset bypasses, and MCP agent calls.
            </p>
          </div>
        </div>

        {/* Export Buttons */}
        <div className="flex items-center gap-2 self-start sm:self-auto">
          <Button
            variant="outline"
            size="sm"
            onClick={exportToJson}
            disabled={filteredEvents.length === 0}
            className="text-xs gap-1.5"
          >
            <Download className="w-3.5 h-3.5" />
            <span>JSON</span>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={exportToCsv}
            disabled={filteredEvents.length === 0}
            className="text-xs gap-1.5"
          >
            <Download className="w-3.5 h-3.5" />
            <span>CSV</span>
          </Button>
        </div>
      </div>

      {/* Filter & Search Bar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
        {/* Severity Tabs */}
        <div className="flex items-center gap-1 p-1 bg-slate-900 border border-slate-800 rounded-xl text-xs w-full sm:w-auto">
          <button
            onClick={() => setSeverityFilter('all')}
            className={`px-3 py-1.5 rounded-lg transition-colors ${
              severityFilter === 'all' ? 'bg-slate-800 text-white font-semibold' : 'text-slate-400 hover:text-white'
            }`}
          >
            All Severities
          </button>
          <button
            onClick={() => setSeverityFilter('info')}
            className={`px-3 py-1.5 rounded-lg transition-colors ${
              severityFilter === 'info' ? 'bg-sky-500/20 text-sky-300 font-semibold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Info
          </button>
          <button
            onClick={() => setSeverityFilter('warning')}
            className={`px-3 py-1.5 rounded-lg transition-colors ${
              severityFilter === 'warning' ? 'bg-amber-500/20 text-amber-300 font-semibold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Warnings
          </button>
          <button
            onClick={() => setSeverityFilter('critical')}
            className={`px-3 py-1.5 rounded-lg transition-colors ${
              severityFilter === 'critical' ? 'bg-rose-500/20 text-rose-300 font-semibold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Critical
          </button>
        </div>

        {/* Search Input */}
        <div className="relative w-full sm:max-w-xs">
          <Search className="w-3.5 h-3.5 absolute left-3 top-3 text-slate-400" />
          <input
            type="text"
            placeholder="Search action, actor, IP, details..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-slate-900/80 text-xs rounded-xl border border-slate-800 pl-8 pr-3 py-2 text-slate-200 placeholder:text-slate-500 outline-none focus:border-sky-500/60"
          />
        </div>
      </div>

      {/* Events Table */}
      <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/60">
        {loading ? (
          <div className="p-12 text-center text-xs text-slate-500 font-mono">
            Loading audit stream...
          </div>
        ) : filteredEvents.length === 0 ? (
          <div className="p-12 text-center space-y-2">
            <ShieldAlert className="w-8 h-8 text-slate-600 mx-auto" />
            <h4 className="text-sm font-semibold text-slate-300">No Security Audit Events</h4>
            <p className="text-xs text-slate-500 max-w-sm mx-auto">
              No audit logs recorded matching your search. Live events stream automatically when users or autonomous agents interact with the forge.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-slate-850/70">
            {filteredEvents.map((ev) => (
              <div key={ev.id} className="p-4 hover:bg-slate-900/40 transition-colors space-y-2">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <div className="flex items-center gap-2.5">
                    {ev.severity === 'critical' && (
                      <Badge variant="closed" className="text-[10px] py-0 gap-1">
                        <AlertCircle className="w-3 h-3" /> Critical
                      </Badge>
                    )}
                    {ev.severity === 'warning' && (
                      <Badge variant="warning" className="text-[10px] py-0 gap-1">
                        <AlertCircle className="w-3 h-3" /> Warning
                      </Badge>
                    )}
                    {ev.severity === 'info' && (
                      <Badge variant="default" className="text-[10px] py-0 gap-1">
                        <Info className="w-3 h-3 text-sky-400" /> Info
                      </Badge>
                    )}

                    <span className="font-mono text-xs font-semibold text-slate-100">
                      {ev.action}
                    </span>
                  </div>

                  <span className="font-mono text-[11px] text-slate-500">{ev.timestamp}</span>
                </div>

                {/* Details line */}
                <div className="text-xs text-slate-300 leading-relaxed pl-1">
                  {ev.details}
                </div>

                {/* Meta details */}
                <div className="flex items-center gap-4 text-[11px] font-mono text-slate-500 pt-1">
                  <span className="flex items-center gap-1 text-slate-400">
                    {ev.actorType === 'agent' ? (
                      <Bot className="w-3 h-3 text-sky-400" />
                    ) : ev.actorType === 'user' ? (
                      <User className="w-3 h-3 text-slate-400" />
                    ) : (
                      <Terminal className="w-3 h-3 text-purple-400" />
                    )}
                    <span>{ev.actor}</span>
                  </span>
                  <span>•</span>
                  <span>IP: <strong className="text-slate-400">{ev.clientIp}</strong></span>
                  <span>•</span>
                  <span>Resource: <strong className="text-slate-400">{ev.resource}</strong></span>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
