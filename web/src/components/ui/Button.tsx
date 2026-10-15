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
  const baseStyles = 'relative inline-flex items-center justify-center font-medium select-none transition-all duration-200 outline-none focus-visible:ring-2 focus-visible:ring-sky-400 focus-visible:ring-offset-2 focus-visible:ring-offset-slate-950 disabled:opacity-50 disabled:pointer-events-none rounded-xl'

  const sizeStyles = {
    xs: 'h-7 px-2.5 text-xs gap-1.5',
    sm: 'h-8 px-3 text-xs gap-1.5',
    md: 'h-9 px-4 text-sm gap-2',
    lg: 'h-11 px-5 text-base gap-2.5',
    icon: 'h-9 w-9 p-0',
  }

  const variantStyles = {
    primary: 'bg-sky-500 hover:bg-sky-400 active:bg-sky-600 text-slate-950 font-semibold shadow-md shadow-sky-500/20 border border-sky-400/40 hover:shadow-sky-500/30',
    secondary: 'bg-slate-800/90 hover:bg-slate-800 text-slate-100 border border-slate-700/80 hover:border-slate-600',
    outline: 'bg-transparent hover:bg-slate-800/60 text-slate-300 hover:text-white border border-slate-700 hover:border-slate-600',
    ghost: 'bg-transparent hover:bg-slate-800/60 text-slate-400 hover:text-slate-100 border-none',
    destructive: 'bg-rose-500/15 hover:bg-rose-500/25 active:bg-rose-500/35 text-rose-400 border border-rose-500/30',
    glow: 'bg-gradient-to-r from-sky-500 to-indigo-500 text-white font-medium shadow-lg shadow-sky-500/25 hover:shadow-sky-500/40 border border-sky-300/30 hover:brightness-110',
    cyber: 'bg-slate-900/90 hover:bg-slate-900 text-sky-400 border border-sky-500/40 hover:border-sky-400 hover:shadow-[0_0_15px_rgba(56,189,248,0.25)]',
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
