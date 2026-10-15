import React from 'react'
import { cn } from '@/lib/utils'

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'open' | 'merged' | 'closed' | 'success' | 'running' | 'warning' | 'agent' | 'outline'
  pulse?: boolean
}

export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = 'default',
  pulse = false,
  className,
  ...props
}) => {
  const variantStyles = {
    default: 'bg-secondary text-secondary-foreground border-border',
    open: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 shadow-[0_0_10px_rgba(16,185,129,0.15)]',
    merged: 'bg-accent/15 text-foreground border-accent/40 shadow-xs',
    closed: 'bg-destructive/15 text-destructive border-destructive/30',
    success: 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30',
    running: 'bg-primary/10 text-primary border-primary/30',
    warning: 'bg-amber-500/10 text-amber-300 border-amber-500/30',
    agent: 'bg-gradient-to-r from-primary/15 to-accent/20 text-primary border-primary/40 shadow-xs',
    outline: 'bg-transparent text-muted-foreground border-border',
  }

  const dotColors = {
    default: 'bg-muted-foreground',
    open: 'bg-emerald-400',
    merged: 'bg-accent',
    closed: 'bg-destructive',
    success: 'bg-emerald-400',
    running: 'bg-primary',
    warning: 'bg-amber-400',
    agent: 'bg-primary animate-pulse',
    outline: 'bg-muted-foreground',
  }

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium border select-none transition-colors duration-200',
        variantStyles[variant],
        className
      )}
      {...props}
    >
      {pulse && (
        <span className="relative flex h-2 w-2">
          <span className={cn('animate-ping absolute inline-flex h-full w-full rounded-full opacity-75', dotColors[variant])} />
          <span className={cn('relative inline-flex rounded-full h-2 w-2', dotColors[variant])} />
        </span>
      )}
      {children}
    </span>
  )
}
