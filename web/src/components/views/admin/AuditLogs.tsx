import React, { useState, useEffect } from 'react'
import { ShieldAlert, Download, Search, AlertCircle, Info, Bot, User, Terminal } from 'lucide-react'
import { Box, BoxHeader, BoxBody, BoxRow } from '@/components/ui/Card'
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
    <div className="space-y-4">
      {/* Header & Export Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
            <ShieldAlert className="w-4 h-4 text-primary" />
            <span>Security & Compliance Audit Log</span>
          </h2>
          <p className="text-xs text-muted-foreground">
            Immutable audit trail of Git pushes, access grants, ruleset bypasses, and MCP agent calls.
          </p>
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
            <span>Export JSON</span>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={exportToCsv}
            disabled={filteredEvents.length === 0}
            className="text-xs gap-1.5"
          >
            <Download className="w-3.5 h-3.5" />
            <span>Export CSV</span>
          </Button>
        </div>
      </div>

      {/* Filter & Search Bar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
        {/* Severity Tabs */}
        <div className="flex items-center gap-1 p-0.5 bg-secondary border border-border rounded-md text-xs w-full sm:w-auto">
          <button
            onClick={() => setSeverityFilter('all')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'all' ? 'bg-background text-foreground font-semibold shadow-2xs' : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            All
          </button>
          <button
            onClick={() => setSeverityFilter('info')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'info' ? 'bg-background text-foreground font-semibold shadow-2xs' : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            Info
          </button>
          <button
            onClick={() => setSeverityFilter('warning')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'warning' ? 'bg-background text-foreground font-semibold shadow-2xs' : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            Warnings
          </button>
          <button
            onClick={() => setSeverityFilter('critical')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'critical' ? 'bg-background text-foreground font-semibold shadow-2xs' : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            Critical
          </button>
        </div>

        {/* Search Input */}
        <div className="relative w-full sm:max-w-xs">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search action, actor, IP, details..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full h-8 pl-8 pr-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
          />
        </div>
      </div>

      {/* Events Table */}
      <Box>
        <BoxHeader className="bg-secondary/35 py-2.5">
          <span className="font-semibold text-foreground text-xs">Audit Events Stream</span>
          <span className="text-muted-foreground font-mono text-[11px]">{filteredEvents.length} events</span>
        </BoxHeader>

        <BoxBody className="p-0">
          {loading ? (
            <div className="p-8 text-center text-xs text-muted-foreground font-mono">
              Loading audit stream...
            </div>
          ) : filteredEvents.length === 0 ? (
            <div className="p-8 text-center space-y-2">
              <ShieldAlert className="w-6 h-6 text-muted-foreground mx-auto" />
              <h4 className="text-xs font-semibold text-foreground">No Security Audit Events</h4>
              <p className="text-xs text-muted-foreground max-w-sm mx-auto">
                No audit logs recorded matching your filter. Live events stream automatically when users or autonomous agents interact with the forge.
              </p>
            </div>
          ) : (
            filteredEvents.map((ev) => (
              <BoxRow key={ev.id} className="p-3.5 hover:bg-secondary/20 transition-colors space-y-1.5">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
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
                        <Info className="w-3 h-3 text-primary" /> Info
                      </Badge>
                    )}

                    <span className="font-mono text-xs font-semibold text-foreground">
                      {ev.action}
                    </span>
                  </div>

                  <span className="font-mono text-[11px] text-muted-foreground">{ev.timestamp}</span>
                </div>

                {/* Details line */}
                <div className="text-xs text-foreground/90 leading-relaxed pl-0.5">
                  {ev.details}
                </div>

                {/* Meta details */}
                <div className="flex items-center gap-3 text-[11px] font-mono text-muted-foreground pt-0.5">
                  <span className="flex items-center gap-1 text-foreground">
                    {ev.actorType === 'agent' ? (
                      <Bot className="w-3 h-3 text-primary" />
                    ) : ev.actorType === 'user' ? (
                      <User className="w-3 h-3 text-muted-foreground" />
                    ) : (
                      <Terminal className="w-3 h-3 text-muted-foreground" />
                    )}
                    <span>{ev.actor}</span>
                  </span>
                  <span>•</span>
                  <span>IP: <strong className="text-foreground">{ev.clientIp}</strong></span>
                  <span>•</span>
                  <span>Resource: <strong className="text-foreground">{ev.resource}</strong></span>
                </div>
              </BoxRow>
            ))
          )}
        </BoxBody>
      </Box>
    </div>
  )
}
