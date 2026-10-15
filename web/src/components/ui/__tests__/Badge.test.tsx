import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { Badge, StateLabel } from '../Badge'

describe('Primer Badge & StateLabel Components', () => {
  it('renders StateLabel for open, merged, and closed states', () => {
    const { rerender } = render(<StateLabel state="open">Open</StateLabel>)
    expect(screen.getByText('Open')).toBeInTheDocument()
    expect(screen.getByText('Open').parentElement).toHaveClass('bg-emerald-600')

    rerender(<StateLabel state="merged">Merged</StateLabel>)
    expect(screen.getByText('Merged')).toBeInTheDocument()
    expect(screen.getByText('Merged').parentElement).toHaveClass('bg-purple-600')

    rerender(<StateLabel state="closed">Closed</StateLabel>)
    expect(screen.getByText('Closed')).toBeInTheDocument()
    expect(screen.getByText('Closed').parentElement).toHaveClass('bg-rose-600')
  })

  it('renders Badge with variants and pulse indicator', () => {
    render(<Badge variant="agent" pulse>AI Agent</Badge>)
    expect(screen.getByText('AI Agent')).toBeInTheDocument()
  })
})
