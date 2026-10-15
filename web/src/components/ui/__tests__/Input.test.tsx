import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { Input, SearchPill } from '../Input'

describe('Input and SearchPill Components', () => {
  it('renders Input with placeholder and handles text entry', () => {
    const onChange = vi.fn()
    render(<Input placeholder="Enter username..." onChange={onChange} />)

    const input = screen.getByPlaceholderText('Enter username...')
    expect(input).toBeInTheDocument()

    fireEvent.change(input, { target: { value: 'shiva' } })
    expect(onChange).toHaveBeenCalled()
  })

  it('renders error message when error prop is provided', () => {
    render(<Input placeholder="Repository name" error="Repository name is already taken" />)

    expect(screen.getByText('Repository name is already taken')).toBeInTheDocument()
    const input = screen.getByPlaceholderText('Repository name')
    expect(input.className).toContain('border-destructive')
  })

  it('renders SearchPill with keyboard shortcut indicator', () => {
    render(<SearchPill placeholder="Type / to search repositories..." />)

    expect(screen.getByPlaceholderText('Type / to search repositories...')).toBeInTheDocument()
    expect(screen.getByText('/')).toBeInTheDocument()
  })
})
