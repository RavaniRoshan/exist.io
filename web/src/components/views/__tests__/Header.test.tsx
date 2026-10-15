import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { Header } from '../Header'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    checkActiveSession: vi.fn().mockResolvedValue({
      authenticated: true,
      user: {
        id: 'u-admin-1',
        login: 'shiva',
        name: 'Roshan Ravani',
        email: 'roshan@exist.io',
        role: 'admin',
        status: 'active',
      },
    }),
    logoutUser: vi.fn().mockResolvedValue({ ok: true }),
    getIssues: vi.fn().mockResolvedValue([
      { id: 'issue-1', number: 1, title: 'Test issue', state: 'open' },
    ]),
    getPullRequests: vi.fn().mockResolvedValue([
      { number: 1, title: 'Test PR', state: 'open' },
    ]),
    getReleases: vi.fn().mockResolvedValue([
      { id: 'rel-1', tagName: 'v0.4.6-enterprise' },
    ]),
    isDemoModeActive: vi.fn().mockReturnValue(false),
  }
})

describe('Header Component', () => {
  const mockRepo = {
    ownerSlug: 'shiva',
    name: 'exist.io',
    description: 'Sovereign GitHub Enterprise alternative',
    defaultBranch: 'main',
    isPrivate: false,
    starsCount: 42,
    forksCount: 7,
    updatedAt: '2 hours ago',
    cloneUrlHttp: 'http://localhost:8080/shiva/exist.io.git',
    cloneUrlSsh: 'git@localhost:2222/shiva/exist.io.git',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders repository breadcrumbs, star count, and fork count', async () => {
    render(
      <Header
        repo={mockRepo}
        activeTab="code"
        onTabChange={vi.fn()}
      />
    )

    expect(screen.getByText('shiva')).toBeInTheDocument()
    expect(screen.getAllByText('exist.io').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('42')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
  })

  it('calls onTabChange when navigation tabs are clicked', async () => {
    const onTabChange = vi.fn()
    render(
      <Header
        repo={mockRepo}
        activeTab="code"
        onTabChange={onTabChange}
      />
    )

    const issuesTab = screen.getByRole('tab', { name: /issues/i })
    fireEvent.click(issuesTab)
    expect(onTabChange).toHaveBeenCalledWith('issues')

    const prsTab = screen.getByRole('tab', { name: /pull requests/i })
    fireEvent.click(prsTab)
    expect(onTabChange).toHaveBeenCalledWith('prs')
  })

  it('opens and closes the user profile dropdown menu', async () => {
    const onTabChange = vi.fn()
    render(
      <Header
        repo={mockRepo}
        activeTab="code"
        onTabChange={onTabChange}
      />
    )

    await waitFor(() => {
      expect(screen.getByTitle('Signed in as @shiva')).toBeInTheDocument()
    })

    const userButton = screen.getByTitle('Signed in as @shiva')
    fireEvent.click(userButton)

    expect(screen.getByText('Your profile')).toBeInTheDocument()
    expect(screen.getByText('Enterprise Settings')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Enterprise Settings'))
    expect(onTabChange).toHaveBeenCalledWith('admin')
  })

  it('opens code clone modal and displays clone commands', async () => {
    render(
      <Header
        repo={mockRepo}
        activeTab="code"
        onTabChange={vi.fn()}
      />
    )

    const codeButton = screen.getByTestId('repo-clone-btn')
    fireEvent.click(codeButton)

    await waitFor(() => {
      expect(screen.getByText('Clone repository')).toBeInTheDocument()
    })

    expect(screen.getByText('git clone http://localhost:8080/shiva/exist.io.git')).toBeInTheDocument()

    // Switch to SSH tab
    const sshTab = screen.getByRole('button', { name: /ssh \(port 2222\)/i })
    fireEvent.click(sshTab)
    expect(screen.getByText('git clone git@localhost:2222/shiva/exist.io.git')).toBeInTheDocument()
  })
})
