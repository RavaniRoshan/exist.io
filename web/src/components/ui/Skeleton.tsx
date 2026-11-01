import React from 'react'
import { cn } from '@/lib/utils'

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  shimmer?: boolean
}

/** Forge loading placeholder. Sweep is disabled under reduced motion. */
export const Skeleton: React.FC<SkeletonProps> = ({
  className,
  shimmer = true,
  ...props
}) => (
  <div
    className={cn(
      'relative overflow-hidden rounded-lg bg-raised',
      shimmer && 'motion-sweep',
      className
    )}
    {...props}
  />
)
