import React from 'react'
import { cn } from '@/lib/utils'

export interface NavTabItem {
  id: string
  label: string
  count?: number | string
  icon?: React.ReactNode
}

export interface UnderlineNavProps {
  tabs: NavTabItem[]
  activeTab: string
  onChange?: (id: string) => void
  onTabChange?: (id: string) => void
  className?: string
  actions?: React.ReactNode
}

/** Forge underline navigation. Active state is a 2px signal rule. */
export const UnderlineNav: React.FC<UnderlineNavProps> = ({
  tabs,
  activeTab,
  onChange,
  onTabChange,
  className,
  actions,
}) => {
  const handleSelect = (id: string) => {
    if (onTabChange) onTabChange(id)
    else if (onChange) onChange(id)
  }

  return (
    <div className={cn('flex items-center justify-between border-b border-hairline w-full overflow-x-auto scrollbar-none', className)}>
      <nav className="flex items-center gap-0.5 -mb-px" role="tablist">
        {tabs.map((tab) => {
          const isActive = activeTab === tab.id
          return (
            <button
              key={tab.id}
              role="tab"
              aria-selected={isActive}
              onClick={() => handleSelect(tab.id)}
              className={cn(
                'relative flex items-center gap-1.5 px-3 py-2 text-body-medium transition-colors cursor-pointer select-none whitespace-nowrap outline-none border-b-2',
                isActive
                  ? 'text-text-1 border-signal'
                  : 'text-text-3 border-transparent hover:text-text-1 hover:border-hairline'
              )}
            >
              {tab.icon && (
                <span className={cn('size-3.5 shrink-0', isActive ? 'text-signal' : 'text-text-3')}>
                  {tab.icon}
                </span>
              )}
              <span>{tab.label}</span>
              {tab.count !== undefined && (
                <span
                  className={cn(
                    'px-1.5 py-0.5 text-micro rounded-full font-mono',
                    isActive
                      ? 'bg-signal/15 text-signal border border-signal/30'
                      : 'bg-raised text-text-3 border border-hairline'
                  )}
                >
                  {tab.count}
                </span>
              )}
            </button>
          )
        })}
      </nav>

      {actions && <div className="flex items-center gap-2 pb-1.5 shrink-0">{actions}</div>}
    </div>
  )
}

/** Backwards-compatible alias. */
export const Tabs: React.FC<{
  tabs: NavTabItem[]
  activeTab: string
  onChange: (id: string) => void
  className?: string
}> = ({ tabs, activeTab, onChange, className }) => (
  <UnderlineNav tabs={tabs} activeTab={activeTab} onChange={onChange} className={className} />
)
