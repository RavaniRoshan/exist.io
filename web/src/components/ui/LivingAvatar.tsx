import React from 'react'
import { motion } from 'motion/react'
import { Bot, User } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface LivingAvatarProps {
  name: string
  isAgent?: boolean
  status?: 'online' | 'busy' | 'offline'
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

export const LivingAvatar: React.FC<LivingAvatarProps> = ({
  name,
  isAgent = false,
  status = 'online',
  size = 'md',
  className,
}) => {
  const sizeMap = {
    sm: 'w-7 h-7 text-xs',
    md: 'w-9 h-9 text-sm',
    lg: 'w-12 h-12 text-base',
  }

  const indicatorSize = {
    sm: 'w-2 h-2',
    md: 'w-2.5 h-2.5',
    lg: 'w-3 h-3',
  }

  const statusColor = {
    online: 'bg-emerald-400',
    busy: 'bg-amber-400',
    offline: 'bg-slate-500',
  }

  const initials = name
    .split(/[-_ ]/)
    .map(n => n[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()

  return (
    <div className={cn('relative inline-flex items-center justify-center shrink-0', className)}>
      <motion.div
        whileHover={{ scale: 1.08 }}
        whileTap={{ scale: 0.95 }}
        transition={{ type: 'spring', stiffness: 400, damping: 20 }}
        className={cn(
          'relative flex items-center justify-center rounded-xl font-bold select-none shadow-md overflow-hidden border',
          sizeMap[size],
          isAgent
            ? 'bg-gradient-to-br from-sky-600 via-indigo-600 to-purple-700 text-white border-sky-400/50 shadow-sky-500/20'
            : 'bg-gradient-to-br from-slate-700 to-slate-800 text-slate-200 border-slate-600/60'
        )}
      >
        {isAgent ? (
          <Bot className={size === 'sm' ? 'w-3.5 h-3.5' : size === 'lg' ? 'w-6 h-6' : 'w-4 h-4'} />
        ) : initials.length > 0 ? (
          initials
        ) : (
          <User className="w-4 h-4" />
        )}

        {/* Agent Cyber Shimmer */}
        {isAgent && (
          <div className="absolute inset-0 bg-gradient-to-t from-sky-400/20 to-transparent pointer-events-none" />
        )}
      </motion.div>

      {/* Status Ring */}
      <span
        className={cn(
          'absolute -bottom-0.5 -right-0.5 rounded-full ring-2 ring-slate-950',
          indicatorSize[size],
          statusColor[status]
        )}
      />
    </div>
  )
}
