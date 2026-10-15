import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import App from '../App'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    checkActiveSession: vi.fn().mockResolvedValue({
      authenticated: true,
      user: {
        id: 'u-admin-1',
        name: 'Roshan Ravani',
        login: 'shiva',
        email: 'roshan@exist.io',
        role: 'admin',
        status: 'active',
        createdAt: '2026-01-01T00:00:00Z',
      },
    }),
    getRepoInfo: vi.fn().mockResolvedValue({
      ownerSlug: 'shiva',
      name: 'exist.io',
      description: 'Sovereign GitHub Enterprise alternative with MCP',
      defaultBranch: 'main',
      isPrivate: false,
      starsCount: 284,
      forksCount: 36,
      updatedAt: 'Just now',
      cloneUrlHttp: 'http://localhost:8080/shiva/exist.io.git',
      cloneUrlSsh: 'git@localhost:shiva/exist.io.git',
    }),
    getFiles: vi.fn().mockResolvedValue([
      { name: 'cmd', path: 'cmd', type: 'tree', size: 0, lastCommitMessage: 'refactor cmd', lastCommitDate: '2 hours ago' },
      { name: 'README.md', path: 'README.md', type: 'blob', size: 1024, lastCommitMessage: 'update docs', lastCommitDate: '1 hour ago' },
    ]),
    getFileContent: vi.fn().mockResolvedValue('# exist.io\nSovereign GitHub alternative.'),
    getCommits: vi.fn().mockResolvedValue([
      {
        sha: '5c92f551234567890abcdef',
        shortSha: '5c92f55',
        message: 'feat: add enterprise auth',
        author: { name: 'Roshan Ravani', email: 'roshan@exist.io', isAgent: false },
        timestamp: '2026-10-15T08:00:00Z',
        verified: true,
      },
    ]),
    getPullRequests: vi.fn().mockResolvedValue([
      {
        id: 'pr-4',
        number: 4,
        title: 'feat: add SAML/OIDC SSO enterprise authentication',
        state: 'open',
        author: { name: 'Roshan Ravani', isAgent: false },
        sourceBranch: 'feat/enterprise-auth',
        targetBranch: 'main',
        createdAt: '2026-10-15T07:00:00Z',
        description: 'Implements SAML 2.0 and OIDC identity federation',
        commitsCount: 3,
        changedFilesCount: 8,
      },
    ]),
    fetchUserSSHKeys: vi.fn().mockResolvedValue([]),
    fetchUserTokens: vi.fn().mockResolvedValue([]),
  }
})

describe('App End-to-End Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders repository overview with Code tab active by default', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getAllByText('exist.io').length).toBeGreaterThan(0)
    })
    expect(screen.getByRole('tab', { name: /code/i })).toHaveAttribute('aria-selected', 'true')
  })

  it('navigates to Issues view when clicking Issues tab', async () => {
    render(<App />)

    const issuesTab = screen.getByRole('tab', { name: /issues/i })
    fireEvent.click(issuesTab)

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/is:issue/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /new issue/i })).toBeInTheDocument()
    })
  })

  it('navigates to Pull Requests view when clicking PRs tab', async () => {
    render(<App />)

    const prTab = screen.getByRole('tab', { name: /pull requests/i })
    fireEvent.click(prTab)

    await waitFor(() => {
      expect(screen.getByText('feat: add SAML/OIDC SSO enterprise authentication')).toBeInTheDocument()
    })
  })

  it('navigates to Commits view when clicking Commits tab', async () => {
    render(<App />)

    const commitsTab = screen.getByRole('tab', { name: /commits/i })
    fireEvent.click(commitsTab)

    await waitFor(() => {
      expect(screen.getByText('feat: add enterprise auth')).toBeInTheDocument()
    })
  })

  it('navigates to Releases view when clicking Releases tab', async () => {
    render(<App />)

    const releasesTab = screen.getByRole('tab', { name: /releases/i })
    fireEvent.click(releasesTab)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /draft a new release/i })).toBeInTheDocument()
    })
  })

  it('navigates to Repository Settings view when clicking Settings tab', async () => {
    render(<App />)

    const settingsTab = screen.getByRole('tab', { name: /settings/i })
    fireEvent.click(settingsTab)

    await waitFor(() => {
      expect(screen.getByText('Repository Settings')).toBeInTheDocument()
      expect(screen.getByText('General Repository Configuration')).toBeInTheDocument()
      expect(screen.getByText('Pull Request Merge Buttons')).toBeInTheDocument()
    })
  })

  it('navigates to Enterprise Settings view via the user avatar menu', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByTitle(/signed in as @shiva/i)).toBeInTheDocument()
    })

    const avatarBtn = screen.getByTitle(/signed in as @shiva/i)
    fireEvent.click(avatarBtn)

    const adminItem = screen.getByText('Enterprise Settings')
    fireEvent.click(adminItem)

    await waitFor(() => {
      expect(screen.getByText('Cluster Health & Metrics')).toBeInTheDocument()
      expect(screen.getByText('Single Sign-On (SAML/OIDC)')).toBeInTheDocument()
    })
  })

  it('navigates to User Profile view via the user avatar menu', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByTitle(/signed in as @shiva/i)).toBeInTheDocument()
    })

    // Click user avatar to open dropdown
    const avatarBtn = screen.getByTitle(/signed in as @shiva/i)
    fireEvent.click(avatarBtn)

    // Click "Your profile"
    const profileItem = screen.getByText('Your profile')
    fireEvent.click(profileItem)

    await waitFor(() => {
      expect(screen.getByText('Roshan Ravani')).toBeInTheDocument()
      expect(screen.getByText('@shiva')).toBeInTheDocument()
      expect(screen.getByText(/contributions in the last year/i)).toBeInTheDocument()
    })
  })

  it('navigates to Global Enterprise Feed when clicking the exist.io brand logo', async () => {
    render(<App />)

    const brandBtn = screen.getByTitle('Go to Enterprise Home Feed')
    fireEvent.click(brandBtn)

    await waitFor(() => {
      expect(screen.getByText('Enterprise Activity Feed')).toBeInTheDocument()
      expect(screen.getByText('Cluster Telemetry')).toBeInTheDocument()
    })
  })
})
