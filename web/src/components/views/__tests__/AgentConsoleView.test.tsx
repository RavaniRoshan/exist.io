import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { AgentConsoleView } from '../AgentConsoleView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    getAgentActivities: vi.fn().mockResolvedValue([
      {
        id: 'act-1',
        agentName: 'exist-agent',
        agentType: 'exist-mcp',
        toolName: 'list_pull_requests',
        status: 'success',
        latencyMs: 24,
        timestamp: '11:20:00 AM',
        requestPayload: { resource: 'acme/rocket', action: 'query' },
        responseSummary: 'Fetched 3 open pull requests',
      },
    ]),
    queryLocalMcpTools: vi.fn().mockResolvedValue([
      { name: 'list_pull_requests', description: 'List pull requests in repository' },
      { name: 'search_symbols', description: 'Semantic code symbol search across AST' },
    ]),
    issueAgentLease: vi.fn().mockResolvedValue({
      id: 'lease-test-1',
      token: 'agl_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.dummy.sig',
      agentId: 'agent:cursor-worker',
      repoId: 'acme/rocket',
      refPrefixes: ['refs/heads/agent/'],
      expiresAt: '2026-10-15T12:00:00Z',
    }),
    isDemoModeActive: vi.fn().mockReturnValue(false),
  }
})

describe('AgentConsoleView Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders hero title and tabs', async () => {
    render(<AgentConsoleView />)

    expect(screen.getByText('AI Agent Telemetry & Capability Leases')).toBeInTheDocument()
    expect(screen.getByText('Live Protocol')).toBeInTheDocument()
    expect(screen.getByText('Real-time Activity')).toBeInTheDocument()
    expect(screen.getByText('Live MCP Tool Runner')).toBeInTheDocument()
    expect(screen.getByText('Capability Leases')).toBeInTheDocument()
    expect(screen.getByText('Tool Catalog')).toBeInTheDocument()
  })

  it('renders agent activity item when activities exist', async () => {
    render(<AgentConsoleView />)

    await waitFor(() => {
      expect(screen.getByText('exist-agent')).toBeInTheDocument()
    })

    expect(screen.getByText('list_pull_requests')).toBeInTheDocument()
    expect(screen.getByText('24ms')).toBeInTheDocument()
    expect(screen.getByText('Fetched 3 open pull requests')).toBeInTheDocument()
  })

  it('switches to Live MCP Tool Runner tab and selects a tool', async () => {
    render(<AgentConsoleView />)

    const testerTab = screen.getByText('Live MCP Tool Runner')
    fireEvent.click(testerTab)

    expect(screen.getByText('Select MCP Tool')).toBeInTheDocument()
    expect(screen.getByText('Run Tool')).toBeInTheDocument()
  })

  it('switches to Capability Leases tab and generates a lease token', async () => {
    render(<AgentConsoleView />)

    const leasesTab = screen.getByText('Capability Leases')
    fireEvent.click(leasesTab)

    expect(screen.getByText('Issue Agent Capability Lease')).toBeInTheDocument()

    const generateBtn = screen.getByRole('button', { name: /sign & issue capability lease/i })
    fireEvent.click(generateBtn)

    await waitFor(() => {
      expect(screen.getByText('Active Lease Credential')).toBeInTheDocument()
    })

    expect(screen.getByText('agl_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.dummy.sig')).toBeInTheDocument()
  })

  it('switches to Tool Catalog tab and lists MCP primitives', async () => {
    render(<AgentConsoleView />)

    const catalogTab = screen.getByText('Tool Catalog')
    fireEvent.click(catalogTab)

    expect(screen.getByText(/Model Context Protocol Tools/i)).toBeInTheDocument()
  })
})
