import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import React from 'react'
import { ProductTour } from '../ProductTour'

describe('ProductTour Component', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('renders step 1 (Sovereign Cluster Architecture) when open', () => {
    const onClose = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} />)

    expect(screen.getByText('Sovereign Cluster Architecture')).toBeInTheDocument()
    expect(screen.getByText(/core engine/i)).toBeInTheDocument()
    expect(screen.getByText(':8080')).toBeInTheDocument()
    expect(screen.getByText(':2222')).toBeInTheDocument()
    expect(screen.getByText('SQLite WAL')).toBeInTheDocument()
  })

  it('navigates through steps when clicking Next and Back', () => {
    const onClose = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} />)

    // Step 1 -> Step 2
    const nextBtn = screen.getByRole('button', { name: /next/i })
    fireEvent.click(nextBtn)

    expect(screen.getByText('Connecting Your Git Clients')).toBeInTheDocument()
    expect(screen.getAllByText(/git wire protocols/i).length).toBeGreaterThan(0)

    // Step 2 -> Step 3
    fireEvent.click(screen.getByRole('button', { name: /next/i }))
    expect(screen.getByText('Autonomous AI Agent Bridge (MCP)')).toBeInTheDocument()

    // Step 3 -> Back to Step 2
    const backBtn = screen.getByRole('button', { name: /back/i })
    fireEvent.click(backBtn)
    expect(screen.getByText('Connecting Your Git Clients')).toBeInTheDocument()
  })

  it('marks tour as completed in localStorage on finish', () => {
    const onClose = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} />)

    // Advance to last step (Step 5)
    for (let i = 0; i < 4; i++) {
      fireEvent.click(screen.getByRole('button', { name: /next/i }))
    }

    expect(screen.getByText('Enterprise Identity & Governance')).toBeInTheDocument()
    const finishBtn = screen.getByRole('button', { name: /get started/i })
    fireEvent.click(finishBtn)

    expect(localStorage.getItem('exist_onboarding_tour_completed')).toBe('true')
    expect(onClose).toHaveBeenCalled()
  })

  it('triggers onOpenManual when clicking Open Manual', () => {
    const onClose = vi.fn()
    const onOpenManual = vi.fn()
    render(<ProductTour isOpen={true} onClose={onClose} onOpenManual={onOpenManual} />)

    const manualBtn = screen.getByRole('button', { name: /open manual/i })
    fireEvent.click(manualBtn)

    expect(onClose).toHaveBeenCalled()
    expect(onOpenManual).toHaveBeenCalled()
  })
})
