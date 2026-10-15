import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { RulesetsView } from '../RulesetsView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    fetchRulesets: vi.fn().mockResolvedValue([
      {
        id: 'rs-1',
        name: 'Mainline Production Policy',
        enforcement: 'active',
        kind: 'branch',
        includedRefNames: ['refs/heads/main'],
        rules: [
          { type: 'pull_request', parameters: { required_approving_review_count: 1 } },
          { type: 'required_signatures', parameters: {} },
        ],
      },
    ]),
    createRuleset: vi.fn().mockResolvedValue({
      ok: true,
      ruleset: {
        id: 'rs-2',
        name: 'Release Tag Protection',
        enforcement: 'active',
        kind: 'tag',
        includedRefNames: ['refs/tags/*'],
      },
    }),
    deleteRuleset: vi.fn().mockResolvedValue(true),
    fetchIPAllowlist: vi.fn().mockResolvedValue({
      enabled: true,
      entries: [
        { id: 'ip-1', cidr: '10.0.0.0/8', description: 'Internal VPN', enabled: true },
      ],
    }),
    setIPAllowlistEnabled: vi.fn().mockResolvedValue(true),
    addIPAllowlistEntry: vi.fn().mockResolvedValue({
      ok: true,
      entry: { id: 'ip-2', cidr: '192.168.1.0/24', description: 'Office LAN', enabled: true },
    }),
    deleteIPAllowlistEntry: vi.fn().mockResolvedValue(true),
    fetchSSHCAs: vi.fn().mockResolvedValue({
      required: false,
      cas: [
        {
          id: 'ca-1',
          name: 'Corporate Root SSH CA',
          fingerprint: 'SHA256:abcd1234efgh5678',
          createdAt: '2026-10-15T00:00:00Z',
        },
      ],
    }),
    addSSHCA: vi.fn().mockResolvedValue({
      ok: true,
      ca: {
        id: 'ca-2',
        name: 'Backup SSH CA',
        fingerprint: 'SHA256:9988776655443322',
        createdAt: 'Just now',
      },
    }),
    deleteSSHCA: vi.fn().mockResolvedValue(true),
    setSSHCARequirement: vi.fn().mockResolvedValue(true),
  }
})

describe('RulesetsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders active branch rulesets and rules list', async () => {
    render(<RulesetsView />)

    await waitFor(() => {
      expect(screen.getByText('Repository Security & Merge Gates')).toBeInTheDocument()
      expect(screen.getByText('Mainline Production Policy')).toBeInTheDocument()
      expect(screen.getByText('Branch Protection & Merge Gates')).toBeInTheDocument()
    })
    expect(screen.getByText('Require Signed Commits')).toBeInTheDocument()
    expect(screen.getByText('Require 1 PR Approval')).toBeInTheDocument()
  })

  it('renders Inbound IP Allowlist and CIDR entries', async () => {
    render(<RulesetsView />)

    await waitFor(() => {
      expect(screen.getByText('Inbound IP Allowlists (CIDR)')).toBeInTheDocument()
      expect(screen.getByText('10.0.0.0/8')).toBeInTheDocument()
      expect(screen.getByText('Internal VPN')).toBeInTheDocument()
    })
  })

  it('renders SSH Certificate Authorities and opens enrollment form', async () => {
    render(<RulesetsView />)

    await waitFor(() => {
      expect(screen.getByText('Corporate Root SSH CA')).toBeInTheDocument()
      expect(screen.getByText('SHA256:abcd1234efgh5678')).toBeInTheDocument()
    })

    const enrollBtn = screen.getByRole('button', { name: /enroll trusted ssh ca/i })
    fireEvent.click(enrollBtn)

    expect(screen.getByPlaceholderText(/Vault Prod CA/i)).toBeInTheDocument()
  })
})
