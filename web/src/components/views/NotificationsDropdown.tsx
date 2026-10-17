import React, { useEffect, useState } from 'react'
import {
  Bell,
  Check,
  CheckCheck,
  GitPullRequest,
  GitMerge,
  AlertCircle,
} from 'lucide-react'
import {
  getNotifications,
  markNotificationRead,
  markAllNotificationsRead,
} from '@/lib/api'
import type { NotificationItem } from '@/lib/types'

interface NotificationsDropdownProps {
  isOpen: boolean
  onClose: () => void
  onNavigate?: (url: string) => void
  onCountChange?: (count: number) => void
}

export const NotificationsDropdown: React.FC<NotificationsDropdownProps> = ({
  isOpen,
  onClose,
  onNavigate,
  onCountChange,
}) => {
  const [notifications, setNotifications] = useState<NotificationItem[]>([])
  const [loading, setLoading] = useState(false)
  const [filterUnread, setFilterUnread] = useState(false)

  const loadNotifications = async () => {
    setLoading(true)
    try {
      const items = await getNotifications(filterUnread)
      setNotifications(items)
      const unread = items.filter((n) => !n.read).length
      onCountChange?.(unread)
    } catch {
      // silently handle
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (isOpen) {
      loadNotifications()
    }
  }, [isOpen, filterUnread])

  const handleMarkRead = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await markNotificationRead(id)
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, read: true } : n))
      )
      const newUnread = notifications.filter((n) => n.id !== id && !n.read).length
      onCountChange?.(newUnread)
    } catch {
      // ignore
    }
  }

  const handleMarkAllRead = async () => {
    try {
      await markAllNotificationsRead()
      setNotifications((prev) => prev.map((n) => ({ ...n, read: true })))
      onCountChange?.(0)
    } catch {
      // ignore
    }
  }

  const handleItemClick = async (n: NotificationItem) => {
    if (!n.read) {
      try {
        await markNotificationRead(n.id)
      } catch {
        // ignore
      }
    }
    onClose()
    if (onNavigate && n.url) {
      onNavigate(n.url)
    }
  }

  if (!isOpen) return null

  const getEventIcon = (eventType: string) => {
    switch (eventType) {
      case 'pr.merged':
        return <GitMerge className="w-3.5 h-3.5 text-purple-400" />
      case 'pr.opened':
      case 'pr.reviewed':
        return <GitPullRequest className="w-3.5 h-3.5 text-emerald-400" />
      case 'issue.assigned':
        return <AlertCircle className="w-3.5 h-3.5 text-amber-400" />
      default:
        return <Bell className="w-3.5 h-3.5 text-primary" />
    }
  }

  return (
    <div
      role="menu"
      aria-label="Notifications panel"
      className="absolute right-0 mt-2 w-80 sm:w-96 rounded-md bg-card border border-border shadow-2xl py-0 z-50 text-xs font-sans animate-in fade-in zoom-in-95 duration-100 overflow-hidden"
    >
      {/* Header */}
      <div className="px-3 py-2.5 bg-secondary/40 border-b border-border flex items-center justify-between">
        <div className="flex items-center gap-2 font-semibold text-foreground">
          <Bell className="w-4 h-4 text-primary" />
          <span>Notifications</span>
          {notifications.some((n) => !n.read) && (
            <span className="px-1.5 py-0.2 bg-primary/20 text-primary text-[10px] font-bold rounded-full">
              {notifications.filter((n) => !n.read).length}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setFilterUnread(!filterUnread)}
            className={`text-[11px] px-1.5 py-0.5 rounded transition-colors cursor-pointer ${
              filterUnread
                ? 'bg-primary/20 text-primary font-medium'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {filterUnread ? 'Unread only' : 'All'}
          </button>
          <button
            onClick={handleMarkAllRead}
            title="Mark all as read"
            className="text-[11px] text-muted-foreground hover:text-foreground flex items-center gap-1 cursor-pointer transition-colors"
          >
            <CheckCheck className="w-3.5 h-3.5" />
            <span className="hidden sm:inline">Mark all read</span>
          </button>
        </div>
      </div>

      {/* List */}
      <div className="max-h-80 overflow-y-auto divide-y divide-border">
        {loading ? (
          <div className="py-8 text-center text-muted-foreground text-xs">
            Loading notifications...
          </div>
        ) : notifications.length === 0 ? (
          <div className="py-8 text-center text-muted-foreground text-xs flex flex-col items-center gap-1">
            <span className="text-base">🎉</span>
            <span className="font-medium text-foreground">All caught up!</span>
            <span>No notifications at this time.</span>
          </div>
        ) : (
          notifications.map((n) => (
            <div
              key={n.id}
              onClick={() => handleItemClick(n)}
              className={`px-3 py-2.5 hover:bg-secondary/50 transition-colors cursor-pointer flex items-start gap-2.5 ${
                !n.read ? 'bg-primary/5' : ''
              }`}
            >
              <div className="mt-0.5 shrink-0">{getEventIcon(n.eventType)}</div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between gap-1">
                  <div
                    className={`font-medium truncate ${
                      !n.read ? 'text-foreground font-semibold' : 'text-muted-foreground'
                    }`}
                  >
                    {n.title}
                  </div>
                  {!n.read && (
                    <span className="w-2 h-2 rounded-full bg-primary shrink-0" />
                  )}
                </div>
                {n.body && (
                  <div className="text-muted-foreground text-[11px] line-clamp-2 mt-0.5">
                    {n.body}
                  </div>
                )}
                <div className="text-[10px] text-muted-foreground/70 mt-1 flex items-center justify-between">
                  <span>{new Date(n.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                  {!n.read && (
                    <button
                      onClick={(e) => handleMarkRead(n.id, e)}
                      title="Mark as read"
                      className="hover:text-primary transition-colors cursor-pointer flex items-center gap-0.5"
                    >
                      <Check className="w-3 h-3" />
                      <span>Read</span>
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
