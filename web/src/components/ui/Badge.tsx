import React from 'react'
import { cn } from '@/lib/utils'

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'open' | 'merged' | 'closed' | 'success' | 'running' | 'warning' | 'agent' | 'outline'
  pulse?: boolean
}

/**
 * Primer-grade StateLabel and CounterBadge.
 */
export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = 'default',
  pulse = false,
  className,
  ...props
}) => {
  const variantStyles = {
    default: 'bg-secondary text-secondary-foreground border-border',
    open: 'bg-emerald-600/15 text-emerald-400 border-emerald-500/30',
    merged: 'bg-purple-600/20 text-purple-300 border-purple-500/40',
    closed: 'bg-destructive/15 text-destructive border-destructive/30',
    success: 'bg-emerald-600/15 text-emerald-400 border-emerald-500/30',
    running: 'bg-primary/15 text-primary border-primary/30',
    warning: 'bg-amber-500/15 text-amber-300 border-amber-500/30',
    agent: 'bg-primary/15 text-primary border-primary/40 font-mono',
    outline: 'bg-transparent text-muted-foreground border-border',
  }

  const dotColors = {
    default: 'bg-muted-foreground',
    open: 'bg-emerald-400',
    merged: 'bg-purple-400',
    closed: 'bg-destructive',
    success: 'bg-emerald-400',
    running: 'bg-primary',
    warning: 'bg-amber-400',
    agent: 'bg-primary',
    outline: 'bg-muted-foreground',
  }

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[11px] font-medium border select-none transition-colors duration-150',
        variantStyles[variant],
        className
      )}
      {...props}
    >
      {pulse && (
        <span className={cn('w-1.5 h-1.5 rounded-full shrink-0', dotColors[variant])} />
      )}
      {children}
    </span>
  )
}

export const StateLabel: React.FC<{
  state: 'open' | 'merged' | 'closed' | 'draft'
  icon?: React.ReactNode
  children?: React.ReactNode
  className?: string
}> = ({ state, icon, children, className }) => {
  const styles = {
    open: 'bg-emerald-600 text-white border-emerald-500',
    merged: 'bg-purple-600 text-white border-purple-500',
    closed: 'bg-rose-600 text-white border-rose-500',
    draft: 'bg-secondary text-muted-foreground border-border',
  }[state]

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold border select-none shadow-2xs',
        styles,
        className
      )}
    >
      {icon}
      <span>{children || state.charAt(0).toUpperCase() + state.slice(1)}</span>
    </span>
  )
}
