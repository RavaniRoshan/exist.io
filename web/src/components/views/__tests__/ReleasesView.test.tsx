import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { ReleasesView } from '../ReleasesView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    getReleases: vi.fn().mockResolvedValue([
      {
        id: 'rel-1',
        tagName: 'v0.4.0-enterprise',
        name: 'v0.4.0 Sovereign Enterprise',
        body: 'Enterprise Git and MCP agent bridge release notes.',
        author: 'shiva',
        isDraft: false,
        isPrerelease: false,
        createdAt: '2026-10-15T06:00:00Z',
        publishedAt: '2026-10-15T06:30:00Z',
        commitSha: '5c92f55',
        assets: [
          {
            name: 'existd-linux-amd64.tar.gz',
            size: 34500000,
            downloadCount: 412,
            sha256: '9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b',
          },
        ],
      },
    ]),
  }
})

describe('ReleasesView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders releases with latest badge, commit sha, and assets', async () => {
    render(<ReleasesView />)

    await waitFor(() => {
      expect(screen.getByText('v0.4.0-enterprise')).toBeInTheDocument()
      expect(screen.getByText('v0.4.0 Sovereign Enterprise')).toBeInTheDocument()
    })
    expect(screen.getByText('Latest')).toBeInTheDocument()
    expect(screen.getByText('5c92f55')).toBeInTheDocument()
    expect(screen.getByText('existd-linux-amd64.tar.gz')).toBeInTheDocument()
    expect(screen.getByText('34.5 MB')).toBeInTheDocument()
    expect(screen.getByText('412 downloads')).toBeInTheDocument()
  })

  it('copies SHA256 checksum to clipboard when clicking checksum pill', async () => {
    render(<ReleasesView />)

    await waitFor(() => {
      expect(screen.getByText('existd-linux-amd64.tar.gz')).toBeInTheDocument()
    })

    const checksumBtn = screen.getByTitle(/SHA256: 9a8b7c6d/i)
    fireEvent.click(checksumBtn)

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b')
  })

  it('switches to Tags tab and displays verified Git tags', async () => {
    render(<ReleasesView />)

    await waitFor(() => {
      expect(screen.getByText('v0.4.0-enterprise')).toBeInTheDocument()
    })

    const tagsTab = screen.getByRole('button', { name: /tags/i })
    fireEvent.click(tagsTab)

    expect(screen.getByText('All Git Tags')).toBeInTheDocument()
    expect(screen.getByText('Verified')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /zip/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /tar\.gz/i })).toBeInTheDocument()
  })

  it('opens Draft a new release modal and publishes a new release', async () => {
    render(<ReleasesView />)

    await waitFor(() => {
      expect(screen.getByText('v0.4.0-enterprise')).toBeInTheDocument()
    })

    const draftBtn = screen.getByRole('button', { name: /draft a new release/i })
    fireEvent.click(draftBtn)

    expect(screen.getByRole('heading', { name: /draft a new release/i })).toBeInTheDocument()
    const tagInput = screen.getByPlaceholderText('v0.5.0')
    const titleInput = screen.getByPlaceholderText('Release title')

    fireEvent.change(tagInput, { target: { value: 'v0.5.0' } })
    fireEvent.change(titleInput, { target: { value: 'v0.5.0 Agentic Autonomy' } })

    const publishBtn = screen.getByRole('button', { name: /publish release/i })
    fireEvent.click(publishBtn)

    await waitFor(() => {
      expect(screen.getByText('v0.5.0 Agentic Autonomy')).toBeInTheDocument()
      expect(screen.getByText('v0.5.0')).toBeInTheDocument()
    })
  })
})
