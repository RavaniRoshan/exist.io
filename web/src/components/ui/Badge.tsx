import React from 'react'
import { cn } from '@/lib/utils'

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'open' | 'merged' | 'closed' | 'success' | 'running' | 'warning' | 'agent' | 'outline'
  pulse?: boolean
}

const variantStyles = {
  default: 'bg-raised text-text-1 border-hairline',
  open: 'bg-signal/15 text-signal border-signal/30',
  merged: 'bg-chart-4/15 text-chart-4 border-chart-4/30',
  closed: 'bg-destructive/15 text-destructive border-destructive/30',
  success: 'bg-signal/15 text-signal border-signal/30',
  running: 'bg-signal/15 text-signal border-signal/30',
  warning: 'bg-ember/15 text-ember border-ember/30',
  agent: 'bg-signal/10 text-signal border-signal/40 font-mono',
  outline: 'bg-transparent text-text-3 border-hairline',
}

const dotColors = {
  default: 'bg-text-3',
  open: 'bg-signal',
  merged: 'bg-chart-4',
  closed: 'bg-destructive',
  success: 'bg-signal',
  running: 'bg-signal',
  warning: 'bg-ember',
  agent: 'bg-signal',
  outline: 'bg-text-3',
}

/** Forge state badge. Solid semantic tones, no raw palette colors. */
export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = 'default',
  pulse = false,
  className,
  ...props
}) => (
  <span
    className={cn(
      'inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-micro font-medium border select-none transition-colors duration-150',
      variantStyles[variant],
      className
    )}
    {...props}
  >
    {pulse && <span className={cn('size-1.5 rounded-full motion-pulse', dotColors[variant])} />}
    {children}
  </span>
)

/** Large uppercase state pill used on PR and issue headers. */
export const StateLabel: React.FC<{
  state: 'open' | 'merged' | 'closed' | 'draft'
  icon?: React.ReactNode
  children?: React.ReactNode
  className?: string
}> = ({ state, icon, children, className }) => {
  const styles = {
    open: 'bg-signal text-signal-ink border-signal',
    merged: 'bg-chart-4 text-white border-chart-4',
    closed: 'bg-destructive text-white border-destructive',
    draft: 'bg-raised text-text-3 border-hairline',
  }[state]

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-caption font-semibold border select-none',
        styles,
        className
      )}
    >
      {icon}
      <span>{children || state.charAt(0).toUpperCase() + state.slice(1)}</span>
    </span>
  )
}
