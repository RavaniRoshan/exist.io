import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { GlobalHomeView } from '../GlobalHomeView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    isDemoModeActive: vi.fn().mockReturnValue(true),
  }
})

describe('GlobalHomeView', () => {
  it('renders top repositories and filters repositories based on search input', () => {
    const onSelectRepo = vi.fn()
    const onSelectTab = vi.fn()

    render(<GlobalHomeView onSelectRepo={onSelectRepo} onSelectTab={onSelectTab} />)

    expect(screen.getByText(/top repositories/i)).toBeInTheDocument()
    expect(screen.getByText('shiva/exist.io')).toBeInTheDocument()
    expect(screen.getByText('shiva/agentic-mcp-bridge')).toBeInTheDocument()

    // Filter by text
    const searchInput = screen.getByPlaceholderText('Find a repository...')
    fireEvent.change(searchInput, { target: { value: 'terraform' } })

    expect(screen.getByText('shiva/iac-terraform-enterprise')).toBeInTheDocument()
    expect(screen.queryByText('shiva/sqlite-git-wal')).not.toBeInTheDocument()
  })

  it('calls onSelectRepo when clicking a repository', () => {
    const onSelectRepo = vi.fn()
    const onSelectTab = vi.fn()

    render(<GlobalHomeView onSelectRepo={onSelectRepo} onSelectTab={onSelectTab} />)

    fireEvent.click(screen.getByText('shiva/exist.io'))
    expect(onSelectRepo).toHaveBeenCalledWith('exist.io')
  })

  it('filters activity feed by event type (Agents, Pull Requests, Commits)', () => {
    const onSelectRepo = vi.fn()
    const onSelectTab = vi.fn()

    render(<GlobalHomeView onSelectRepo={onSelectRepo} onSelectTab={onSelectTab} />)

    expect(screen.getByText('Enterprise Activity Feed')).toBeInTheDocument()
    expect(screen.getByText('Exist Agent (Autonomous)')).toBeInTheDocument()

    // Filter by PRs only
    const prFilterBtn = screen.getByRole('button', { name: /pull requests/i })
    fireEvent.click(prFilterBtn)

    expect(screen.queryByText('Exist Agent (Autonomous)')).not.toBeInTheDocument()
    expect(screen.getByText(/#5 feat: add UserProfileView/i)).toBeInTheDocument()

    // Filter by Commits
    const commitFilterBtn = screen.getByRole('button', { name: /commits/i })
    fireEvent.click(commitFilterBtn)

    expect(screen.getByText('Roshan Ravani')).toBeInTheDocument()
    expect(screen.queryByText(/#5 feat: add UserProfileView/i)).not.toBeInTheDocument()
  })

  it('renders cluster telemetry metrics and triggers agent console tab callback', () => {
    const onSelectRepo = vi.fn()
    const onSelectTab = vi.fn()

    render(<GlobalHomeView onSelectRepo={onSelectRepo} onSelectTab={onSelectTab} />)

    expect(screen.getByText('Cluster Telemetry')).toBeInTheDocument()
    expect(screen.getByText('Online (8080)')).toBeInTheDocument()
    expect(screen.getByText('Port 2222')).toBeInTheDocument()
    expect(screen.getByText('SQLite WAL')).toBeInTheDocument()
    expect(screen.getByText('35')).toBeInTheDocument()

    const openAgentBtn = screen.getByRole('button', { name: /open agent console/i })
    fireEvent.click(openAgentBtn)
    expect(onSelectTab).toHaveBeenCalledWith('agents')
  })
})
