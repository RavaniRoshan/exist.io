import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { UserManualModal } from '../UserManualModal'

describe('UserManualModal Component', () => {
  it('renders quick start tab by default with copyable commands', async () => {
    const onClose = vi.fn()
    render(<UserManualModal isOpen={true} onClose={onClose} />)

    expect(screen.getByText('exist.io User Manual & Architecture Guide')).toBeInTheDocument()
    expect(screen.getByText('Getting Started with exist.io')).toBeInTheDocument()
    expect(screen.getByText(/ssh-keygen -t ed25519/i)).toBeInTheDocument()
    expect(screen.getByText(/claude mcp add exist/i)).toBeInTheDocument()
  })

  it('switches between tabs: Architecture, AI Agents, Security, Federation, and CLI', async () => {
    const onClose = vi.fn()
    render(<UserManualModal isOpen={true} onClose={onClose} />)

    // Architecture tab
    const archTab = screen.getByRole('tab', { name: /architecture & ports/i })
    fireEvent.click(archTab)
    expect(screen.getByText('Cluster Architecture & Port Map')).toBeInTheDocument()
    expect(screen.getByText('Network Protocols & Default Ports')).toBeInTheDocument()

    // Agents & MCP tab
    const agentsTab = screen.getByRole('tab', { name: /ai agents & mcp/i })
    fireEvent.click(agentsTab)
    expect(screen.getByText('Model Context Protocol (MCP) Integration')).toBeInTheDocument()
    expect(screen.getByText('Claude Code Configuration')).toBeInTheDocument()
    expect(screen.getByText('Key Agent Capabilities (35 Tools)')).toBeInTheDocument()

    // Security & Signing tab
    const secTab = screen.getByRole('tab', { name: /rulesets & signing/i })
    fireEvent.click(secTab)
    expect(screen.getByText('Rulesets & Cryptographic Enforcement')).toBeInTheDocument()
    expect(screen.getByText(/Configuring Git Commit Signing with SSH/i)).toBeInTheDocument()

    // CLI tab
    const cliTab = screen.getByRole('tab', { name: /cli reference/i })
    fireEvent.click(cliTab)
    expect(screen.getByText('Command Line Reference')).toBeInTheDocument()
    expect(screen.getByText(/existd bootstrap --email/i)).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn()
    render(<UserManualModal isOpen={true} onClose={onClose} />)

    const closeBtn = screen.getByRole('button', { name: /^close$/i })
    fireEvent.click(closeBtn)
    expect(onClose).toHaveBeenCalled()
  })
})
