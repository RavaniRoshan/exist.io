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
    default: 'bg-slate-800 text-slate-300 border-slate-700/60',
    open: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 shadow-[0_0_10px_rgba(16,185,129,0.15)]',
    merged: 'bg-purple-500/15 text-purple-300 border-purple-500/30 shadow-[0_0_10px_rgba(168,85,247,0.15)]',
    closed: 'bg-rose-500/10 text-rose-400 border-rose-500/30',
    success: 'bg-teal-500/10 text-teal-300 border-teal-500/30',
    running: 'bg-sky-500/10 text-sky-400 border-sky-500/30',
    warning: 'bg-amber-500/10 text-amber-300 border-amber-500/30',
    agent: 'bg-gradient-to-r from-sky-500/15 to-violet-500/15 text-sky-300 border-sky-400/40 shadow-[0_0_12px_rgba(56,189,248,0.2)]',
    outline: 'bg-transparent text-slate-400 border-slate-700',
  }

  const dotColors = {
    default: 'bg-slate-400',
    open: 'bg-emerald-400',
    merged: 'bg-purple-400',
    closed: 'bg-rose-400',
    success: 'bg-teal-400',
    running: 'bg-sky-400',
    warning: 'bg-amber-400',
    agent: 'bg-sky-400 animate-pulse',
    outline: 'bg-slate-500',
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
