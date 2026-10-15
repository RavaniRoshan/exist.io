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

/**
 * GitHub Primer UnderlineNav.
 * Renders horizontal tab navigation with active bottom underline border.
 */
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
    <div className={cn('flex items-center justify-between border-b border-border w-full overflow-x-auto scrollbar-none', className)}>
      <nav className="flex items-center gap-1 -mb-[1px]" role="tablist">
        {tabs.map((tab) => {
          const isActive = activeTab === tab.id
          return (
            <button
              key={tab.id}
              role="tab"
              aria-selected={isActive}
              onClick={() => handleSelect(tab.id)}
              className={cn(
                'relative flex items-center gap-2 px-3.5 py-2.5 text-xs font-medium transition-colors cursor-pointer select-none whitespace-nowrap outline-none',
                isActive
                  ? 'text-foreground font-semibold border-b-2 border-primary'
                  : 'text-muted-foreground hover:text-foreground border-b-2 border-transparent hover:border-border'
              )}
            >
              {tab.icon && <span className={cn('w-4 h-4 shrink-0', isActive ? 'text-primary' : 'text-muted-foreground')}>{tab.icon}</span>}
              <span>{tab.label}</span>
              {tab.count !== undefined && (
                <span
                  className={cn(
                    'px-2 py-0.2 text-[10px] rounded-full font-mono font-semibold',
                    isActive
                      ? 'bg-primary/20 text-primary border border-primary/30'
                      : 'bg-secondary text-muted-foreground border border-border'
                  )}
                >
                  {tab.count}
                </span>
              )}
            </button>
          )
        })}
      </nav>

      {actions && (
        <div className="flex items-center gap-2 pb-1.5 shrink-0">
          {actions}
        </div>
      )}
    </div>
  )
}

// Backwards compatibility alias for Tabs
export const Tabs: React.FC<{
  tabs: NavTabItem[]
  activeTab: string
  onChange: (id: string) => void
  className?: string
}> = ({ tabs, activeTab, onChange, className }) => (
  <UnderlineNav tabs={tabs} activeTab={activeTab} onChange={onChange} className={className} />
)
