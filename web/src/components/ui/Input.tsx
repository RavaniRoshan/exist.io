import React, { forwardRef } from 'react'
import { Search } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  icon?: React.ReactNode
  error?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(({
  className,
  type = 'text',
  icon,
  error,
  ...props
}, ref) => {
  return (
    <div className="relative w-full">
      {icon && (
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
          {icon}
        </div>
      )}
      <input
        ref={ref}
        type={type}
        className={cn(
          'w-full bg-slate-900/90 text-slate-100 placeholder:text-slate-500 text-sm rounded-xl border border-slate-800/90 py-2 transition-all duration-200 outline-none',
          'focus:border-sky-500/70 focus:ring-1 focus:ring-sky-500/50 focus:shadow-[0_0_15px_rgba(56,189,248,0.15)]',
          icon ? 'pl-9 pr-3' : 'px-3',
          error && 'border-rose-500/80 focus:border-rose-500',
          className
        )}
        {...props}
      />
      {error && <span className="text-xs text-rose-400 mt-1 block">{error}</span>}
    </div>
  )
})

Input.displayName = 'Input'

export interface SearchPillProps extends React.InputHTMLAttributes<HTMLInputElement> {
  onSearch?: (value: string) => void
}

export const SearchPill: React.FC<SearchPillProps> = ({
  className,
  placeholder = 'Search branches, commits, PRs... (⌘K)',
  ...props
}) => {
  return (
    <div className={cn('relative flex items-center w-full max-w-md', className)}>
      <Search className="absolute left-3 w-4 h-4 text-slate-400 pointer-events-none" />
      <input
        type="text"
        placeholder={placeholder}
        className="w-full bg-slate-900/60 hover:bg-slate-900 text-slate-200 placeholder:text-slate-500 text-xs rounded-full border border-slate-800 py-1.5 pl-9 pr-12 transition-all duration-200 focus:outline-none focus:border-sky-500/60 focus:ring-1 focus:ring-sky-500/40 focus:bg-slate-900"
        {...props}
      />
      <div className="absolute right-2.5 flex items-center gap-0.5 pointer-events-none">
        <kbd className="text-[10px] bg-slate-800 border border-slate-700/80 text-slate-400 px-1.5 py-0.5 rounded shadow-inner">
          ⌘K
        </kbd>
      </div>
    </div>
  )
}
