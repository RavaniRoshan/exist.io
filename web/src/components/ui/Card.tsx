import React from 'react'
import { cn } from '@/lib/utils'

export interface BoxProps extends React.HTMLAttributes<HTMLDivElement> {
  flush?: boolean
}

/**
 * Primer-grade Box container.
 * Features 6px (rounded-md) corners, 1px crisp border, and authentic surface elevation.
 */
export const Box = React.forwardRef<HTMLDivElement, BoxProps>(({
  children,
  className,
  flush = false,
  ...props
}, ref) => {
  return (
    <div
      ref={ref}
      className={cn(
        'rounded-md border border-border bg-card text-card-foreground',
        !flush && 'overflow-hidden shadow-2xs',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
})
Box.displayName = 'Box'

export const BoxHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(({
  className,
  children,
  ...props
}, ref) => {
  return (
    <div
      ref={ref}
      className={cn(
        'px-4 py-2.5 bg-secondary/35 border-b border-border text-xs font-semibold text-foreground flex items-center justify-between gap-3',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
})
BoxHeader.displayName = 'BoxHeader'

export const BoxTitle = React.forwardRef<HTMLHeadingElement, React.HTMLAttributes<HTMLHeadingElement>>(({
  className,
  children,
  ...props
}, ref) => {
  return (
    <h3
      ref={ref}
      className={cn('text-sm font-semibold tracking-tight text-foreground flex items-center gap-2', className)}
      {...props}
    >
      {children}
    </h3>
  )
})
BoxTitle.displayName = 'BoxTitle'

export const BoxDescription = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLParagraphElement>>(({
  className,
  children,
  ...props
}, ref) => {
  return (
    <p
      ref={ref}
      className={cn('text-xs text-muted-foreground leading-normal', className)}
      {...props}
    >
      {children}
    </p>
  )
})
BoxDescription.displayName = 'BoxDescription'

export const BoxBody = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(({
  className,
  children,
  ...props
}, ref) => {
  return (
    <div
      ref={ref}
      className={cn('p-4 text-xs', className)}
      {...props}
    >
      {children}
    </div>
  )
})
BoxBody.displayName = 'BoxBody'

export const BoxRow = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(({
  className,
  children,
  ...props
}, ref) => {
  return (
    <div
      ref={ref}
      className={cn(
        'px-4 py-2.5 border-b border-border last:border-b-0 hover:bg-secondary/30 transition-colors flex items-center justify-between gap-3 text-xs',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
})
BoxRow.displayName = 'BoxRow'

export const BoxFooter = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(({
  className,
  children,
  ...props
}, ref) => {
  return (
    <div
      ref={ref}
      className={cn(
        'px-4 py-2.5 bg-secondary/20 border-t border-border flex items-center justify-between gap-3 text-xs text-muted-foreground',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
})
BoxFooter.displayName = 'BoxFooter'

// Backwards-compatible aliases for existing views
export const Card = Box
export const CardHeader = BoxHeader
export const CardTitle = BoxTitle
export const CardDescription = BoxDescription
export const CardContent = BoxBody
