import React, { forwardRef } from 'react'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive' | 'danger' | 'glow' | 'cyber'
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'icon'
  loading?: boolean
}

/**
 * Forge button. 32px default height, 6px radius, 1px hairline border.
 * Press darkens instead of scaling.
 */
export const Button = forwardRef<HTMLButtonElement, ButtonProps>(({
  className,
  variant = 'secondary',
  size = 'md',
  loading = false,
  disabled,
  children,
  ...props
}, ref) => {
  const baseStyles = 'relative inline-flex items-center justify-center font-medium select-none transition-colors duration-150 outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 focus-visible:ring-offset-background disabled:opacity-50 disabled:pointer-events-none rounded-md cursor-pointer active:brightness-95'

  const sizeStyles = {
    xs: 'h-6 px-2 text-[11px] gap-1',
    sm: 'h-7 px-2.5 text-xs gap-1.5',
    md: 'h-8 px-3 text-xs gap-1.5',
    lg: 'h-9 px-4 text-sm gap-2',
    icon: 'h-8 w-8 p-0',
  }

  const variantStyles = {
    primary: 'bg-signal text-signal-ink border border-signal/60 hover:bg-signal/90',
    secondary: 'bg-raised text-text-1 border border-hairline hover:bg-raised/80',
    outline: 'bg-transparent text-text-1 border border-hairline hover:bg-raised/60',
    ghost: 'bg-transparent text-text-2 border border-transparent hover:bg-raised/50 hover:text-text-1',
    destructive: 'bg-destructive/15 text-destructive border border-destructive/40 hover:bg-destructive/25',
    danger: 'bg-destructive/15 text-destructive border border-destructive/40 hover:bg-destructive/25',
    glow: 'bg-signal text-signal-ink border border-signal/60 hover:bg-signal/90',
    cyber: 'bg-raised text-text-1 border border-hairline hover:bg-raised/80',
  }

  return (
    <button
      ref={ref}
      disabled={disabled || loading}
      className={cn(baseStyles, sizeStyles[size], variantStyles[variant], className)}
      {...props}
    >
      {loading && <Loader2 className="size-3.5 animate-spin" />}
      {children}
    </button>
  )
})

Button.displayName = 'Button'
