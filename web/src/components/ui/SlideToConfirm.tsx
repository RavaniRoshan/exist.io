import React, { useState, useRef } from 'react'
import { motion, useMotionValue, useTransform } from 'motion/react'
import { GitMerge, Check, ArrowRight } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface SlideToConfirmProps {
  onConfirm: () => void
  label?: string
  confirmedLabel?: string
  disabled?: boolean
  className?: string
}

export const SlideToConfirm: React.FC<SlideToConfirmProps> = ({
  onConfirm,
  label = 'Slide to Merge Pull Request',
  confirmedLabel = 'Merged Successfully',
  disabled = false,
  className,
}) => {
  const [isConfirmed, setIsConfirmed] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const x = useMotionValue(0)

  // Maximum drag distance calculation
  const maxDrag = 210

  const opacity = useTransform(x, [0, maxDrag * 0.7], [1, 0])
  const bgOpacity = useTransform(x, [0, maxDrag], [0.1, 0.4])

  const handleDragEnd = () => {
    if (disabled || isConfirmed) return
    if (x.get() >= maxDrag - 15) {
      setIsConfirmed(true)
      onConfirm()
    } else {
      x.set(0)
    }
  }

  return (
    <div
      ref={containerRef}
      className={cn(
        'relative flex items-center h-12 w-full max-w-sm rounded-2xl bg-card border border-border overflow-hidden select-none p-1 shadow-inner',
        disabled && 'opacity-50 pointer-events-none',
        className
      )}
    >
      {/* Background fill based on drag */}
      <motion.div
        className="absolute inset-0 bg-gradient-to-r from-primary/20 via-accent/20 to-emerald-500/20 pointer-events-none"
        style={{ opacity: isConfirmed ? 0.8 : bgOpacity }}
      />

      {/* Center Label */}
      <motion.span
        style={{ opacity }}
        className="absolute inset-0 flex items-center justify-center text-xs font-semibold tracking-wide text-foreground/80 pointer-events-none pl-6 gap-2"
      >
        <span>{isConfirmed ? confirmedLabel : label}</span>
        {!isConfirmed && <ArrowRight className="w-3.5 h-3.5 text-primary animate-pulse" />}
      </motion.span>

      {/* Draggable Handle */}
      {!isConfirmed ? (
        <motion.div
          drag="x"
          dragConstraints={{ left: 0, right: maxDrag }}
          dragElastic={0.1}
          dragSnapToOrigin={false}
          onDragEnd={handleDragEnd}
          style={{ x }}
          className="relative z-10 flex items-center justify-center h-10 w-12 rounded-xl bg-primary text-primary-foreground cursor-grab active:cursor-grabbing shadow-sm border border-primary/40"
        >
          <GitMerge className="w-5 h-5" />
        </motion.div>
      ) : (
        <div className="relative z-10 flex items-center justify-center h-10 w-full rounded-xl bg-primary text-primary-foreground font-medium text-xs gap-2 shadow-sm">
          <Check className="w-4 h-4 text-primary-foreground" />
          <span>{confirmedLabel}</span>
        </div>
      )}
    </div>
  )
}
