import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import React from 'react'
import { AgentThinking } from '../agent-thinking'

describe('AgentThinking (BoardUI Free)', () => {
  it('announces its status label', () => {
    render(<AgentThinking label="Searching the docs" />)
    expect(screen.getByRole('status')).toBeInTheDocument()
    expect(screen.getByLabelText('Searching the docs')).toBeInTheDocument()
  })

  it('renders an elapsed timer by default and hides it on request', () => {
    const { rerender } = render(<AgentThinking label="Thinking" />)
    expect(screen.getByText(/^\d+\.\d+s$/)).toBeInTheDocument()

    rerender(<AgentThinking label="Thinking" showTimer={false} />)
    expect(screen.queryByText(/^\d+\.\d+s$/)).not.toBeInTheDocument()
  })

  it('applies the shimmer only when enabled', () => {
    const { rerender } = render(<AgentThinking label="Thinking" />)
    expect(screen.getByLabelText('Thinking')).toHaveClass('bui-agent-thinking-label')

    rerender(<AgentThinking label="Thinking" shimmer={false} />)
    expect(screen.getByLabelText('Thinking')).not.toHaveClass('bui-agent-thinking-label')
  })
})
