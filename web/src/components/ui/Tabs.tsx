import React from 'react'
import { motion } from 'motion/react'
import { cn } from '@/lib/utils'

export interface TabItem {
  id: string
  label: string
  count?: number | string
  icon?: React.ReactNode
}

export interface TabsProps {
  tabs: TabItem[]
  activeTab: string
  onChange: (id: string) => void
  className?: string
}

export const Tabs: React.FC<TabsProps> = ({
  tabs,
  activeTab,
  onChange,
  className,
}) => {
  return (
    <div className={cn('flex items-center gap-1 p-1 bg-card/80 border border-border rounded-xl backdrop-blur-md overflow-x-auto scrollbar-none', className)}>
      {tabs.map((tab) => {
        const isActive = activeTab === tab.id
        return (
          <button
            key={tab.id}
            onClick={() => onChange(tab.id)}
            className={cn(
              'relative flex items-center gap-2 px-3.5 py-1.5 text-xs font-medium rounded-lg transition-colors duration-200 outline-none whitespace-nowrap',
              isActive ? 'text-foreground font-semibold' : 'text-muted-foreground hover:text-foreground hover:bg-secondary/50'
            )}
          >
            {isActive && (
              <motion.div
                layoutId="active-tab-indicator"
                className="absolute inset-0 bg-primary/15 border border-primary/30 rounded-lg shadow-xs"
                transition={{ type: 'spring', stiffness: 500, damping: 35 }}
              />
            )}
            <span className="relative z-10 flex items-center gap-1.5">
              {tab.icon}
              {tab.label}
              {tab.count !== undefined && (
                <span
                  className={cn(
                    'px-1.5 py-0.2 text-[10px] rounded-full font-semibold',
                    isActive ? 'bg-primary/25 text-primary' : 'bg-secondary text-muted-foreground'
                  )}
                >
                  {tab.count}
                </span>
              )}
            </span>
          </button>
        )
      })}
    </div>
  )
}
