import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { CommandPalette } from '../CommandPalette'

function renderPalette(overrides: Partial<React.ComponentProps<typeof CommandPalette>> = {}) {
  const props = {
    open: true,
    onOpenChange: vi.fn(),
    onTabChange: vi.fn(),
    onOpenTour: vi.fn(),
    onOpenManual: vi.fn(),
    onToggleDemoMode: vi.fn(),
    isDemo: false,
    ...overrides,
  }
  render(<CommandPalette {...props} />)
  return props
}

describe('CommandPalette', () => {
  it('lists navigation commands when open', () => {
    renderPalette()
    for (const label of ['Code', 'Issues', 'Pull requests', 'Agent Console', 'Repository settings', 'Admin console']) {
      expect(screen.getByText(label)).toBeInTheDocument()
    }
    expect(screen.getByText('Navigate')).toBeInTheDocument()
    expect(screen.getByText('Actions')).toBeInTheDocument()
  })

  it('jumps to a tab and closes on selection', () => {
    const props = renderPalette()
    fireEvent.click(screen.getByText('Pull requests'))
    expect(props.onTabChange).toHaveBeenCalledWith('prs')
    expect(props.onOpenChange).toHaveBeenCalledWith(false)
  })

  it('filters commands by query', () => {
    renderPalette()
    const input = screen.getByPlaceholderText('Type a command or search…')
    fireEvent.change(input, { target: { value: 'webhook' } })
    expect(screen.getByText('Webhooks')).toBeInTheDocument()
    expect(screen.queryByText('Pull requests')).not.toBeInTheDocument()
  })

  it('shows the empty state for unmatched queries', () => {
    renderPalette()
    const input = screen.getByPlaceholderText('Type a command or search…')
    fireEvent.change(input, { target: { value: 'zzzqqq' } })
    expect(screen.getByText('No matching command.')).toBeInTheDocument()
  })

  it('runs the tour action', () => {
    const props = renderPalette()
    fireEvent.click(screen.getByText('Take the guided tour'))
    expect(props.onOpenTour).toHaveBeenCalled()
    expect(props.onOpenChange).toHaveBeenCalledWith(false)
  })

  it('reflects demo mode in the toggle action label', () => {
    renderPalette({ isDemo: true })
    expect(screen.getByText('Switch to Live Production')).toBeInTheDocument()
  })
})
