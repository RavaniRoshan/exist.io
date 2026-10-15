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
        'relative flex items-center h-12 w-full max-w-sm rounded-2xl bg-slate-900 border border-purple-500/30 overflow-hidden select-none p-1 shadow-inner',
        disabled && 'opacity-50 pointer-events-none',
        className
      )}
    >
      {/* Background fill based on drag */}
      <motion.div
        className="absolute inset-0 bg-gradient-to-r from-purple-600/30 via-indigo-600/30 to-emerald-600/30 pointer-events-none"
        style={{ opacity: isConfirmed ? 0.8 : bgOpacity }}
      />

      {/* Center Label */}
      <motion.span
        style={{ opacity }}
        className="absolute inset-0 flex items-center justify-center text-xs font-semibold tracking-wide text-slate-300 pointer-events-none pl-6 gap-2"
      >
        <span>{isConfirmed ? confirmedLabel : label}</span>
        {!isConfirmed && <ArrowRight className="w-3.5 h-3.5 text-purple-400 animate-pulse" />}
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
          className="relative z-10 flex items-center justify-center h-10 w-12 rounded-xl bg-gradient-to-r from-purple-500 to-indigo-600 text-white cursor-grab active:cursor-grabbing shadow-lg shadow-purple-500/30 border border-purple-300/40"
        >
          <GitMerge className="w-5 h-5" />
        </motion.div>
      ) : (
        <div className="relative z-10 flex items-center justify-center h-10 w-full rounded-xl bg-purple-600 text-white font-medium text-xs gap-2 shadow-md">
          <Check className="w-4 h-4 text-emerald-300" />
          <span>{confirmedLabel}</span>
        </div>
      )}
    </div>
  )
}
