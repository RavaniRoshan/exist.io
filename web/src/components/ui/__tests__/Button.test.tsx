import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { Button } from '../Button'

describe('Primer Button Component', () => {
  it('renders standard Primer button with correct styling', () => {
    render(<Button>Click Me</Button>)
    const button = screen.getByRole('button', { name: /click me/i })
    expect(button).toBeInTheDocument()
    expect(button).toHaveClass('rounded-md')
    expect(button).toHaveClass('h-8')
  })

  it('handles click events', () => {
    const handleClick = vi.fn()
    render(<Button onClick={handleClick}>Submit</Button>)
    fireEvent.click(screen.getByRole('button', { name: /submit/i }))
    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('renders loading state and disables interaction', () => {
    const handleClick = vi.fn()
    render(<Button loading onClick={handleClick}>Saving</Button>)
    const button = screen.getByRole('button')
    expect(button).toBeDisabled()
    fireEvent.click(button)
    expect(handleClick).not.toHaveBeenCalled()
  })

  it('supports variants (primary, outline, danger)', () => {
    const { rerender } = render(<Button variant="primary">Primary</Button>)
    expect(screen.getByRole('button')).toHaveClass('bg-primary')

    rerender(<Button variant="outline">Outline</Button>)
    expect(screen.getByRole('button')).toHaveClass('border-border')

    rerender(<Button variant="danger">Danger</Button>)
    expect(screen.getByRole('button')).toHaveClass('text-destructive')
  })
})
