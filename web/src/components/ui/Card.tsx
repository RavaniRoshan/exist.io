import React from 'react'
import { cn } from '@/lib/utils'

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  glow?: boolean
  gradient?: boolean
}

export const Card: React.FC<CardProps> = ({
  children,
  className,
  glow = false,
  gradient = false,
  ...props
}) => {
  return (
    <div
      className={cn(
        'relative rounded-2xl border border-border bg-card/85 text-card-foreground backdrop-blur-xl p-5 shadow-sm transition-all duration-300',
        glow && 'hover:border-primary/50 hover:shadow-[0_0_25px_color-mix(in_oklch,var(--primary)_20%,transparent)]',
        gradient && 'before:absolute before:inset-0 before:-z-10 before:rounded-2xl before:bg-gradient-to-b before:from-primary/10 before:to-transparent before:opacity-50',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}

export const CardHeader: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}) => {
  return (
    <div className={cn('flex flex-col space-y-1.5 pb-4', className)} {...props}>
      {children}
    </div>
  )
}

export const CardTitle: React.FC<React.HTMLAttributes<HTMLHeadingElement>> = ({
  className,
  children,
  ...props
}) => {
  return (
    <h3 className={cn('text-lg font-semibold tracking-tight text-foreground flex items-center gap-2', className)} {...props}>
      {children}
    </h3>
  )
}

export const CardDescription: React.FC<React.HTMLAttributes<HTMLParagraphElement>> = ({
  className,
  children,
  ...props
}) => {
  return (
    <p className={cn('text-xs text-muted-foreground leading-relaxed', className)} {...props}>
      {children}
    </p>
  )
}

export const CardContent: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}) => {
  return (
    <div className={cn('pt-0', className)} {...props}>
      {children}
    </div>
  )
}
