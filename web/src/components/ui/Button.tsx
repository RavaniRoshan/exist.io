import React, { forwardRef } from 'react'
import { motion, type HTMLMotionProps } from 'motion/react'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface ButtonProps extends Omit<HTMLMotionProps<'button'>, 'children'> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive' | 'glow' | 'cyber'
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'icon'
  loading?: boolean
  children?: React.ReactNode
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(({
  className,
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled,
  children,
  ...props
}, ref) => {
  const baseStyles = 'relative inline-flex items-center justify-center font-medium select-none transition-all duration-200 outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:opacity-50 disabled:pointer-events-none rounded-xl'

  const sizeStyles = {
    xs: 'h-7 px-2.5 text-xs gap-1.5',
    sm: 'h-8 px-3 text-xs gap-1.5',
    md: 'h-9 px-4 text-sm gap-2',
    lg: 'h-11 px-5 text-base gap-2.5',
    icon: 'h-9 w-9 p-0',
  }

  const variantStyles = {
    primary: 'bg-primary hover:opacity-90 active:opacity-100 text-primary-foreground font-semibold shadow-xs border border-primary/40 hover:shadow-sm',
    secondary: 'bg-secondary hover:bg-secondary/80 text-secondary-foreground border border-border hover:border-muted-foreground/40',
    outline: 'bg-transparent hover:bg-secondary/60 text-foreground border border-border hover:border-muted-foreground/60',
    ghost: 'bg-transparent hover:bg-secondary/50 text-muted-foreground hover:text-foreground border-none',
    destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90 shadow-xs border border-destructive/40',
    glow: 'bg-gradient-to-r from-primary to-accent text-primary-foreground font-semibold shadow-sm hover:shadow-md border border-primary/40 hover:brightness-105',
    cyber: 'bg-card/90 hover:bg-card text-primary border border-primary/40 hover:border-primary hover:shadow-[0_0_15px_color-mix(in_oklch,var(--primary)_25%,transparent)]',
  }

  return (
    <motion.button
      ref={ref}
      whileTap={{ scale: disabled || loading ? 1 : 0.98 }}
      whileHover={{ scale: disabled || loading ? 1 : 1.01 }}
      transition={{ type: 'spring', stiffness: 500, damping: 30 }}
      disabled={disabled || loading}
      className={cn(baseStyles, sizeStyles[size], variantStyles[variant], className)}
      {...props}
    >
      {loading && <Loader2 className="w-4 h-4 mr-2 animate-spin text-current" />}
      {children}
    </motion.button>
  )
})

Button.displayName = 'Button'
