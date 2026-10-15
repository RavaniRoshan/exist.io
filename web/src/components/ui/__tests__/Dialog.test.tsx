import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { Dialog } from '../Dialog'

describe('Dialog Component', () => {
  it('renders title, description, and children when open', () => {
    render(
      <Dialog isOpen={true} title="Test Title" description="Test Description">
        <p>Dialog inner content</p>
      </Dialog>
    )

    expect(screen.getByText('Test Title')).toBeInTheDocument()
    expect(screen.getByText('Test Description')).toBeInTheDocument()
    expect(screen.getByText('Dialog inner content')).toBeInTheDocument()
  })

  it('does not render content when closed', () => {
    render(
      <Dialog isOpen={false} title="Hidden Title">
        <p>Hidden content</p>
      </Dialog>
    )

    expect(screen.queryByText('Hidden Title')).not.toBeInTheDocument()
    expect(screen.queryByText('Hidden content')).not.toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn()
    render(
      <Dialog isOpen={true} title="Closable Dialog" onClose={onClose}>
        <p>Content</p>
      </Dialog>
    )

    const closeButton = screen.getByRole('button')
    fireEvent.click(closeButton)
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('calls onClose when Escape key is pressed', () => {
    const onClose = vi.fn()
    render(
      <Dialog isOpen={true} title="Escape Dialog" onClose={onClose}>
        <p>Press escape to close</p>
      </Dialog>
    )

    fireEvent.keyDown(window, { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(1)
  })
})
