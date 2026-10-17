import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { NotificationsDropdown } from '../NotificationsDropdown'
import * as api from '@/lib/api'

vi.mock('@/lib/api', () => ({
  getNotifications: vi.fn(),
  markNotificationRead: vi.fn(),
  markAllNotificationsRead: vi.fn(),
}))

describe('NotificationsDropdown', () => {
  it('does not render when closed', () => {
    const { container } = render(<NotificationsDropdown isOpen={false} onClose={() => {}} />)
    expect(container.firstChild).toBeNull()
  })

  it('renders notifications list when open', async () => {
    vi.mocked(api.getNotifications).mockResolvedValueOnce([
      {
        id: 'notif-1',
        userId: 'usr-1',
        eventType: 'pr.opened',
        title: 'New PR #42: Feature X',
        body: 'alice opened a new pull request',
        url: '/owner/repo/pull/42',
        read: false,
        createdAt: new Date().toISOString(),
      },
    ])

    render(<NotificationsDropdown isOpen={true} onClose={() => {}} />)

    expect(screen.getByText('Notifications')).toBeDefined()
    await waitFor(() => {
      expect(screen.getByText('New PR #42: Feature X')).toBeDefined()
      expect(screen.getByText('alice opened a new pull request')).toBeDefined()
    })
  })

  it('renders empty state when there are no notifications', async () => {
    vi.mocked(api.getNotifications).mockResolvedValueOnce([])

    render(<NotificationsDropdown isOpen={true} onClose={() => {}} />)

    await waitFor(() => {
      expect(screen.getByText('All caught up!')).toBeDefined()
    })
  })

  it('calls markAllNotificationsRead when clicking mark all read', async () => {
    vi.mocked(api.getNotifications).mockResolvedValueOnce([
      {
        id: 'notif-1',
        userId: 'usr-1',
        eventType: 'pr.opened',
        title: 'PR #1',
        body: 'opened',
        url: '/owner/repo/pull/1',
        read: false,
        createdAt: new Date().toISOString(),
      },
    ])
    vi.mocked(api.markAllNotificationsRead).mockResolvedValueOnce({ ok: true })

    render(<NotificationsDropdown isOpen={true} onClose={() => {}} />)

    await waitFor(() => {
      expect(screen.getByText('PR #1')).toBeDefined()
    })

    const markAllBtn = screen.getByTitle('Mark all as read')
    fireEvent.click(markAllBtn)

    await waitFor(() => {
      expect(api.markAllNotificationsRead).toHaveBeenCalled()
    })
  })
})
