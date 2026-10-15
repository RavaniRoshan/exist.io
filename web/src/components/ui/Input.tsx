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
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-muted-foreground">
          {icon}
        </div>
      )}
      <input
        ref={ref}
        type={type}
        className={cn(
          'w-full bg-input text-foreground placeholder:text-muted-foreground text-sm rounded-xl border border-border py-2 transition-all duration-200 outline-none',
          'focus:border-primary focus:ring-1 focus:ring-ring focus:shadow-xs',
          icon ? 'pl-9 pr-3' : 'px-3',
          error && 'border-destructive focus:border-destructive',
          className
        )}
        {...props}
      />
      {error && <span className="text-xs text-destructive mt-1 block">{error}</span>}
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
      <Search className="absolute left-3 w-4 h-4 text-muted-foreground pointer-events-none" />
      <input
        type="text"
        placeholder={placeholder}
        className="w-full bg-card/80 hover:bg-card text-foreground placeholder:text-muted-foreground text-xs rounded-full border border-border py-1.5 pl-9 pr-12 transition-all duration-200 focus:outline-none focus:border-primary focus:ring-1 focus:ring-ring"
        {...props}
      />
      <div className="absolute right-2.5 flex items-center gap-0.5 pointer-events-none">
        <kbd className="text-[10px] bg-secondary border border-border text-muted-foreground px-1.5 py-0.5 rounded shadow-2xs">
          ⌘K
        </kbd>
      </div>
    </div>
  )
}
