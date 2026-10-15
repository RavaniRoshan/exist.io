import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { AuditLogs } from '../admin/AuditLogs'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api')>('@/lib/api')
  return {
    ...actual,
    getAuditEvents: vi.fn(),
  }
})

describe('AuditLogs Admin View', () => {
  const mockAuditEvents: api.AuditEvent[] = [
    {
      id: 'aud-1',
      timestamp: 'Today at 10:30 AM',
      actor: 'shiva',
      actorType: 'user',
      action: 'ruleset.updated',
      resource: 'shiva/exist.io',
      clientIp: '192.168.1.100',
      severity: 'warning',
      details: 'Enforced required signed commits and 2 approving reviews on main',
    },
    {
      id: 'aud-2',
      timestamp: 'Today at 09:15 AM',
      actor: 'exist-agent',
      actorType: 'agent',
      action: 'pr.merged',
      resource: 'shiva/exist.io#1',
      clientIp: '127.0.0.1',
      severity: 'info',
      details: 'Autonomously merged telemetry pull request after check suite passed',
    },
    {
      id: 'aud-3',
      timestamp: 'Yesterday at 11:00 PM',
      actor: '10.0.0.55',
      actorType: 'system',
      action: 'auth.failed',
      resource: 'ssh:2222',
      clientIp: '10.0.0.55',
      severity: 'critical',
      details: 'Unauthorized public key authentication attempt rejected',
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getAuditEvents).mockResolvedValue(mockAuditEvents)
  })

  it('renders audit events with severity levels and metadata', async () => {
    render(<AuditLogs />)

    await waitFor(() => {
      expect(screen.getByText('ruleset.updated')).toBeInTheDocument()
      expect(screen.getByText('pr.merged')).toBeInTheDocument()
      expect(screen.getByText('auth.failed')).toBeInTheDocument()
    })

    expect(screen.getByText(/Enforced required signed commits/i)).toBeInTheDocument()
  })

  it('filters audit log entries by severity button', async () => {
    render(<AuditLogs />)

    await waitFor(() => {
      expect(screen.getByText('ruleset.updated')).toBeInTheDocument()
    })

    const criticalBtn = screen.getByRole('button', { name: /^Critical$/i })
    fireEvent.click(criticalBtn)

    expect(screen.getByText('auth.failed')).toBeInTheDocument()
    expect(screen.queryByText('pr.merged')).not.toBeInTheDocument()
    expect(screen.queryByText('ruleset.updated')).not.toBeInTheDocument()
  })

  it('filters audit entries by keyword search', async () => {
    render(<AuditLogs />)

    await waitFor(() => {
      expect(screen.getByText('ruleset.updated')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText(/Search action, actor, IP, details/i)
    fireEvent.change(searchInput, { target: { value: 'telemetry' } })

    expect(screen.getByText('pr.merged')).toBeInTheDocument()
    expect(screen.queryByText('auth.failed')).not.toBeInTheDocument()
  })
})
