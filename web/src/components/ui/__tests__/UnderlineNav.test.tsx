import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { UnderlineNav } from '../UnderlineNav'

describe('Primer UnderlineNav Component', () => {
  const tabs = [
    { id: 'code', label: 'Code' },
    { id: 'prs', label: 'Pull requests', count: 3 },
    { id: 'commits', label: 'Commits', count: 12 },
  ]

  it('renders all navigation tabs with labels and counts', () => {
    render(<UnderlineNav tabs={tabs} activeTab="code" onTabChange={() => {}} />)

    expect(screen.getByText('Code')).toBeInTheDocument()
    expect(screen.getByText('Pull requests')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('Commits')).toBeInTheDocument()
    expect(screen.getByText('12')).toBeInTheDocument()
  })

  it('indicates the active tab with bottom border style', () => {
    render(<UnderlineNav tabs={tabs} activeTab="prs" onTabChange={() => {}} />)

    const prTab = screen.getByRole('tab', { name: /pull requests/i })
    expect(prTab).toHaveClass('border-primary')
    expect(prTab).toHaveAttribute('aria-selected', 'true')
  })

  it('calls onTabChange when a tab is clicked', () => {
    const handleTabChange = vi.fn()
    render(<UnderlineNav tabs={tabs} activeTab="code" onTabChange={handleTabChange} />)

    fireEvent.click(screen.getByRole('tab', { name: /commits/i }))
    expect(handleTabChange).toHaveBeenCalledWith('commits')
  })
})
