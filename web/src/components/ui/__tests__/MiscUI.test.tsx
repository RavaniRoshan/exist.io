import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { Skeleton } from '../Skeleton'
import { SlideToConfirm } from '../SlideToConfirm'
import { Tabs } from '../Tabs'

describe('Miscellaneous UI Components', () => {
  it('renders Skeleton with shimmer effect by default', () => {
    const { container } = render(<Skeleton className="w-20 h-4" />)
    const el = container.firstChild as HTMLElement
    expect(el).toBeInTheDocument()
    expect(el.className).toContain('bg-secondary/60')
    expect(el.className).toContain('w-20')
  })

  it('renders SlideToConfirm wrapping MergeBox and triggers onConfirm', () => {
    const onConfirm = vi.fn()
    render(<SlideToConfirm onConfirm={onConfirm} />)

    const mergeBtn = screen.getByRole('button', { name: /create a merge commit/i })
    expect(mergeBtn).toBeInTheDocument()

    fireEvent.click(mergeBtn)
    const confirmBtn = screen.getByRole('button', { name: /confirm create a merge commit/i })
    expect(confirmBtn).toBeInTheDocument()

    fireEvent.click(confirmBtn)
    expect(onConfirm).toHaveBeenCalled()
  })

  it('renders Tabs forwarding to UnderlineNav and triggers onChange', () => {
    const onChange = vi.fn()
    const testTabs = [
      { id: 'tab-1', label: 'Overview' },
      { id: 'tab-2', label: 'Settings', count: 3 },
    ]

    render(<Tabs tabs={testTabs} activeTab="tab-1" onChange={onChange} />)

    expect(screen.getByText('Overview')).toBeInTheDocument()
    expect(screen.getByText('Settings')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Settings'))
    expect(onChange).toHaveBeenCalledWith('tab-2')
  })
})
