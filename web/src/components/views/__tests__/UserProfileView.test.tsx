import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { UserProfileView } from '../UserProfileView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    isDemoModeActive: vi.fn().mockReturnValue(true),
    checkActiveSession: vi.fn().mockResolvedValue({
      authenticated: true,
      user: {
        id: 'u-1',
        name: 'Roshan Ravani',
        login: 'shiva',
        email: 'roshan@exist.io',
        role: 'admin',
        status: 'active',
        createdAt: '2026-01-01T00:00:00Z',
      },
    }),
    fetchUserSSHKeys: vi.fn().mockResolvedValue([
      {
        id: 'ssh-test-1',
        name: 'MacBook Pro Ed25519',
        fingerprint: 'SHA256:abcd1234efgh5678',
        key_type: 'ssh-ed25519',
        created_at: '2026-10-14T00:00:00Z',
      },
    ]),
    addUserSSHKey: vi.fn().mockImplementation((name, key) =>
      Promise.resolve({
        ok: true,
        key: {
          id: 'ssh-new-1',
          name,
          fingerprint: 'SHA256:newfingerprint',
          key_type: 'ssh-ed25519',
          created_at: '2026-10-15T00:00:00Z',
        },
      })
    ),
    deleteUserSSHKey: vi.fn().mockResolvedValue(true),
    fetchUserTokens: vi.fn().mockResolvedValue([
      {
        id: 'token-test-1',
        name: 'CI Deployment Token',
        token_prefix: 'ghp_test',
        scopes: ['repo', 'workflow'],
        expires_at: '2027-10-15T00:00:00Z',
        created_at: '2026-10-14T00:00:00Z',
      },
    ]),
    createUserToken: vi.fn().mockImplementation((name, scopes) =>
      Promise.resolve({
        ok: true,
        token: {
          id: 'token-new-1',
          name,
          token_prefix: 'ghp_new',
          raw_token: 'ghp_test_secret_token_12345',
          scopes,
          expires_at: '2027-01-01T00:00:00Z',
          created_at: '2026-10-15T00:00:00Z',
        },
      })
    ),
    deleteUserToken: vi.fn().mockResolvedValue(true),
  }
})

describe('UserProfileView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders identity profile card and metadata accurately', async () => {
    render(<UserProfileView />)

    await waitFor(() => {
      expect(screen.getByText('Roshan Ravani')).toBeInTheDocument()
      expect(screen.getByText('@shiva')).toBeInTheDocument()
    })
    expect(screen.getByText('Enterprise Owner')).toBeInTheDocument()
    expect(screen.getByText(/exist\.io Inc\./i)).toBeInTheDocument()
    expect(screen.getByText(/San Francisco, CA/i)).toBeInTheDocument()
  })

  it('renders 52-week contribution heatmap and pinned repositories in Overview tab', async () => {
    render(<UserProfileView />)

    await waitFor(() => {
      expect(screen.getByText(/2,842 contributions in the last year/i)).toBeInTheDocument()
    })
    expect(screen.getByText('Pinned')).toBeInTheDocument()
    expect(screen.getByText('agentic-mcp-bridge')).toBeInTheDocument()
    expect(screen.getByText('sqlite-git-wal')).toBeInTheDocument()
  })

  it('switches to Repositories tab and filters repositories', async () => {
    render(<UserProfileView />)

    await waitFor(() => {
      expect(screen.getByText('Roshan Ravani')).toBeInTheDocument()
    })

    const repoTab = screen.getByRole('tab', { name: /repositories/i })
    fireEvent.click(repoTab)

    const searchInput = screen.getByPlaceholderText('Find a repository...')
    expect(searchInput).toBeInTheDocument()

    // Filter by text
    fireEvent.change(searchInput, { target: { value: 'terraform' } })
    expect(screen.getByText('iac-terraform-enterprise')).toBeInTheDocument()
    expect(screen.queryByText('sqlite-git-wal')).not.toBeInTheDocument()
  })

  it('switches to SSH keys tab and renders existing keys and adds a new key', async () => {
    render(<UserProfileView />)

    const keysTab = screen.getByRole('tab', { name: /ssh and gpg keys/i })
    fireEvent.click(keysTab)

    await waitFor(() => {
      expect(screen.getByText('MacBook Pro Ed25519')).toBeInTheDocument()
      expect(screen.getByText('SHA256:abcd1234efgh5678')).toBeInTheDocument()
    })

    // Click "New SSH key" button
    const newKeyBtn = screen.getByRole('button', { name: /new ssh key/i })
    fireEvent.click(newKeyBtn)

    // Modal opens
    expect(screen.getByText('Add new SSH key')).toBeInTheDocument()
    const titleInput = screen.getByPlaceholderText('e.g., Personal MacBook Air')
    const keyTextarea = screen.getByPlaceholderText(/Begins with 'ssh-rsa'/i)

    fireEvent.change(titleInput, { target: { value: 'Server Key' } })
    fireEvent.change(keyTextarea, { target: { value: 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... server' } })

    const submitBtn = screen.getByRole('button', { name: /add ssh key/i })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(api.addUserSSHKey).toHaveBeenCalledWith('Server Key', 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... server')
    })
  })

  it('switches to Personal Access Tokens tab and generates a new token', async () => {
    render(<UserProfileView />)

    const tokensTab = screen.getByRole('tab', { name: /personal access tokens/i })
    fireEvent.click(tokensTab)

    await waitFor(() => {
      expect(screen.getByText('CI Deployment Token')).toBeInTheDocument()
    })

    // Open token generation modal
    const generateBtn = screen.getByRole('button', { name: /generate new token/i })
    fireEvent.click(generateBtn)

    expect(screen.getByText('New personal access token (classic)')).toBeInTheDocument()
    const noteInput = screen.getByPlaceholderText("What's this token for?")
    fireEvent.change(noteInput, { target: { value: 'Agentic Test Token' } })

    const modalSubmitBtn = screen.getByRole('button', { name: /generate token/i })
    fireEvent.click(modalSubmitBtn)

    await waitFor(() => {
      expect(api.createUserToken).toHaveBeenCalled()
      expect(screen.getByText(/Make sure to copy your personal access token now/i)).toBeInTheDocument()
      expect(screen.getByDisplayValue('ghp_test_secret_token_12345')).toBeInTheDocument()
    })
  })
})
