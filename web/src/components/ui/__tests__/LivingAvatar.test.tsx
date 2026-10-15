import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import React from 'react'
import { LivingAvatar } from '../LivingAvatar'

describe('LivingAvatar Component', () => {
  it('renders user initials when given a human name', () => {
    render(<LivingAvatar name="Roshan Ravani" isAgent={false} />)

    expect(screen.getByText('RR')).toBeInTheDocument()
  })

  it('renders single initial for single word name', () => {
    render(<LivingAvatar name="shiva" isAgent={false} />)

    expect(screen.getByText('S')).toBeInTheDocument()
  })

  it('renders Bot icon when isAgent is true', () => {
    const { container } = render(<LivingAvatar name="exist-agent" isAgent={true} />)

    // Agent Cyber Shimmer is rendered for agent
    expect(container.querySelector('svg.lucide-bot')).toBeInTheDocument()
  })

  it('renders correct indicator color for status', () => {
    const { container: onlineContainer } = render(
      <LivingAvatar name="alex" status="online" />
    )
    expect(onlineContainer.querySelector('span.bg-emerald-400')).toBeInTheDocument()

    const { container: busyContainer } = render(
      <LivingAvatar name="alex" status="busy" />
    )
    expect(busyContainer.querySelector('span.bg-amber-400')).toBeInTheDocument()

    const { container: offlineContainer } = render(
      <LivingAvatar name="alex" status="offline" />
    )
    expect(offlineContainer.querySelector('span.bg-slate-500')).toBeInTheDocument()
  })
})
