import React from 'react'
import {
  LayoutDashboard,
  User,
  Code2,
  AlertCircle,
  GitPullRequest,
  GitCommit,
  Tag,
  Bot,
  ShieldCheck,
  Webhook,
  Settings,
  Server,
  Sparkles,
  BookOpen,
  FlaskConical,
} from 'lucide-react'
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandList,
  CommandItem,
} from '@/components/primitives/command'
import { Kbd, KbdGroup } from '@/components/primitives/kbd'
import type { LucideIcon } from 'lucide-react'

export interface CommandPaletteProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onTabChange: (tab: string) => void
  onOpenTour?: () => void
  onOpenManual?: () => void
  onToggleDemoMode?: () => void
  isDemo?: boolean
}

interface PaletteItem {
  id: string
  label: string
  hint?: string
  icon: LucideIcon
  keywords?: string
  run: () => void
}

/** GitHub-style command palette: navigate tabs, run actions. */
export const CommandPalette: React.FC<CommandPaletteProps> = ({
  open,
  onOpenChange,
  onTabChange,
  onOpenTour,
  onOpenManual,
  onToggleDemoMode,
  isDemo = false,
}) => {
  const close = () => onOpenChange(false)
  const go = (tab: string) => () => {
    onTabChange(tab)
    close()
  }
  const trigger = (fn?: () => void) => () => {
    fn?.()
    close()
  }

  const navigate: PaletteItem[] = [
    { id: 'home', label: 'Enterprise Feed', icon: LayoutDashboard, keywords: 'home dashboard activity', run: go('home') },
    { id: 'profile', label: 'Your profile', icon: User, keywords: 'user account settings', run: go('profile') },
    { id: 'code', label: 'Code', icon: Code2, keywords: 'repository files source', run: go('code') },
    { id: 'issues', label: 'Issues', icon: AlertCircle, keywords: 'bugs tickets', run: go('issues') },
    { id: 'prs', label: 'Pull requests', icon: GitPullRequest, keywords: 'pr review merge', run: go('prs') },
    { id: 'commits', label: 'Commits', icon: GitCommit, keywords: 'history log', run: go('commits') },
    { id: 'releases', label: 'Releases', icon: Tag, keywords: 'tags versions', run: go('releases') },
    { id: 'agents', label: 'Agent Console', icon: Bot, keywords: 'ai leases mcp agents', run: go('agents') },
    { id: 'rulesets', label: 'Rulesets & Security', icon: ShieldCheck, keywords: 'policy branch protection', run: go('rulesets') },
    { id: 'webhooks', label: 'Webhooks', icon: Webhook, keywords: 'integrations ci delivery', run: go('webhooks') },
    { id: 'settings', label: 'Repository settings', icon: Settings, keywords: 'config admin', run: go('settings') },
    { id: 'admin', label: 'Admin console', icon: Server, keywords: 'cluster enterprise scim', run: go('admin') },
  ]

  const actions: PaletteItem[] = [
    ...(onOpenTour
      ? [{ id: 'tour', label: 'Take the guided tour', icon: Sparkles, keywords: 'onboarding help', run: trigger(onOpenTour) }]
      : []),
    ...(onOpenManual
      ? [{ id: 'manual', label: 'Open user manual', icon: BookOpen, keywords: 'docs help guide', run: trigger(onOpenManual) }]
      : []),
    ...(onToggleDemoMode
      ? [{
          id: 'demo-toggle',
          label: isDemo ? 'Switch to Live Production' : 'Switch to Sandbox Demo',
          icon: FlaskConical,
          keywords: 'demo mode live sandbox',
          run: trigger(onToggleDemoMode),
        }]
      : []),
  ]

  const renderItem = (item: PaletteItem) => (
    <CommandItem key={item.id} textValue={`${item.label} ${item.keywords ?? ''}`} onAction={item.run}>
      <item.icon />
      <span>{item.label}</span>
      {item.hint && <Kbd className="ml-auto">{item.hint}</Kbd>}
    </CommandItem>
  )

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Command palette"
      description="Navigate the forge or run an action."
    >
      <Command aria-label="Command palette">
        <CommandInput placeholder="Type a command or search…" />
        <CommandList renderEmptyState={() => <CommandEmpty>No matching command.</CommandEmpty>}>
          <CommandGroup heading="Navigate">
            {navigate.map(renderItem)}
          </CommandGroup>
          {actions.length > 0 && (
            <CommandGroup heading="Actions">
              {actions.map(renderItem)}
            </CommandGroup>
          )}
        </CommandList>
        <div className="flex items-center justify-between gap-2 border-t border-border px-3 py-2 text-[11px] text-muted-foreground">
          <span className="flex items-center gap-2">
            <KbdGroup>
              <Kbd>↑</Kbd>
              <Kbd>↓</Kbd>
            </KbdGroup>
            <span>navigate</span>
            <Kbd>↵</Kbd>
            <span>run</span>
          </span>
          <span className="flex items-center gap-2">
            <Kbd>esc</Kbd>
            <span>close</span>
          </span>
        </div>
      </Command>
    </CommandDialog>
  )
}
