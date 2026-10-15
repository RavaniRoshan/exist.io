import React, { forwardRef } from 'react'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive' | 'danger' | 'glow' | 'cyber'
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'icon'
  loading?: boolean
}

/**
 * GitHub Primer Button.
 * Features 6px (rounded-md) corners, 32px standard height, crisp 1px borders, and tactile active press.
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
  const baseStyles = 'relative inline-flex items-center justify-center font-semibold select-none transition-all duration-150 outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 focus-visible:ring-offset-background disabled:opacity-50 disabled:pointer-events-none rounded-md cursor-pointer active:scale-[0.98]'

  const sizeStyles = {
    xs: 'h-6 px-2 text-[11px] gap-1',
    sm: 'h-7 px-2.5 text-xs gap-1.5',
    md: 'h-8 px-3 text-xs gap-1.5',
    lg: 'h-9 px-4 text-sm gap-2',
    icon: 'h-8 w-8 p-0',
  }

  const variantStyles = {
    primary: 'bg-primary hover:opacity-90 active:opacity-100 text-primary-foreground border border-primary/50 shadow-2xs',
    secondary: 'bg-secondary hover:bg-secondary/80 text-secondary-foreground border border-border shadow-2xs',
    outline: 'bg-transparent hover:bg-secondary/60 text-foreground border border-border shadow-2xs',
    ghost: 'bg-transparent hover:bg-secondary/50 text-muted-foreground hover:text-foreground border-transparent',
    destructive: 'bg-destructive/15 hover:bg-destructive/25 text-destructive border border-destructive/40 shadow-2xs',
    danger: 'bg-destructive/15 hover:bg-destructive/25 text-destructive border border-destructive/40 shadow-2xs',
    glow: 'bg-primary hover:opacity-90 text-primary-foreground border border-primary/50 shadow-2xs',
    cyber: 'bg-secondary hover:bg-secondary/80 text-foreground border border-border shadow-2xs',
  }

  return (
    <button
      ref={ref}
      disabled={disabled || loading}
      className={cn(baseStyles, sizeStyles[size], variantStyles[variant], className)}
      {...props}
    >
      {loading && <Loader2 className="w-3.5 h-3.5 mr-1.5 animate-spin text-current" />}
      {children}
    </button>
  )
})

Button.displayName = 'Button'
