import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { CommitsView } from '../CommitsView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    getCommits: vi.fn(),
  }
})

describe('CommitsView Component', () => {
  const mockCommits = [
    {
      sha: 'abcdef1234567890abcdef1234567890abcdef12',
      shortSha: 'abcdef1',
      message: 'feat: implement capability leases in engine',
      timestamp: '2 hours ago',
      verified: true,
      author: {
        name: 'Roshan Ravani',
        email: 'roshan@exist.io',
        isAgent: false,
      },
      verification: {
        verified: true,
        key_type: 'ssh',
        key_id: 'SHA256:4Xk9p0vV+eM3p0eQ4rT2Y9z8lM1nO7kP8wJ2lQ5rS0=',
        signer_email: 'roshan@exist.io',
        reason: 'Signed with enrolled SSH key',
      },
    },
    {
      sha: '1234567890abcdef1234567890abcdef12345678',
      shortSha: '1234567',
      message: 'chore: update CI matrix',
      timestamp: 'Oct 14, 2026',
      verified: false,
      author: {
        name: 'Dev Bot',
        email: 'bot@exist.io',
        isAgent: true,
      },
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getCommits).mockResolvedValue(mockCommits)
  })

  it('renders date-grouped commits timeline in Primer Boxes', async () => {
    render(<CommitsView />)

    expect(await screen.findByText('feat: implement capability leases in engine')).toBeInTheDocument()
    expect(screen.getByText('chore: update CI matrix')).toBeInTheDocument()
    expect(screen.getByText('Branch:')).toBeInTheDocument()
    expect(screen.getByText('2 commits')).toBeInTheDocument()
    expect(screen.getByText('50% Verified')).toBeInTheDocument()
  })

  it('toggles cryptographic signature inspection drawer on click', async () => {
    render(<CommitsView />)

    const verifiedBtn = await screen.findByRole('button', { name: /verified \(ssh\)/i })
    expect(verifiedBtn).toBeInTheDocument()

    // Click verified button to open drawer
    fireEvent.click(verifiedBtn)

    expect(await screen.findByText(/Cryptographic Signature Audit Attestation/i)).toBeInTheDocument()
    expect(screen.getByText(/OpenSSH SIGSPEC/i)).toBeInTheDocument()
    expect(screen.getByText(/CRYPTOGRAPHICALLY VALID/i)).toBeInTheDocument()
  })
})
