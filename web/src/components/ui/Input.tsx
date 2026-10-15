import React, { forwardRef } from 'react'
import { Search } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  icon?: React.ReactNode
  error?: string
}

/**
 * GitHub Primer TextInput.
 * Features 6px (rounded-md) corners, 32px height, inset surface fill, and crisp focus ring.
 */
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
        <div className="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-muted-foreground">
          {icon}
        </div>
      )}
      <input
        ref={ref}
        type={type}
        className={cn(
          'w-full h-8 bg-input text-foreground placeholder:text-muted-foreground text-xs rounded-md border border-border py-1 transition-colors duration-150 outline-none',
          'focus:border-primary focus:ring-1 focus:ring-ring shadow-2xs',
          icon ? 'pl-8 pr-2.5' : 'px-2.5',
          error && 'border-destructive focus:border-destructive',
          className
        )}
        {...props}
      />
      {error && <span className="text-[11px] text-destructive mt-1 block">{error}</span>}
    </div>
  )
})

Input.displayName = 'Input'

export interface SearchPillProps extends React.InputHTMLAttributes<HTMLInputElement> {
  onSearch?: (value: string) => void
}

/**
 * GitHub Primer Search Bar.
 */
export const SearchPill: React.FC<SearchPillProps> = ({
  className,
  placeholder = 'Type / to search...',
  ...props
}) => {
  return (
    <div className={cn('relative flex items-center w-full max-w-sm', className)}>
      <Search className="absolute left-2.5 w-3.5 h-3.5 text-muted-foreground pointer-events-none" />
      <input
        type="text"
        placeholder={placeholder}
        className="w-full h-8 bg-input text-foreground placeholder:text-muted-foreground text-xs rounded-md border border-border py-1 pl-8 pr-8 transition-colors duration-150 focus:outline-none focus:border-primary focus:ring-1 focus:ring-ring shadow-2xs"
        {...props}
      />
      <div className="absolute right-2 flex items-center pointer-events-none">
        <kbd className="text-[10px] bg-secondary border border-border text-muted-foreground px-1.5 py-0.2 rounded font-mono shadow-2xs">
          /
        </kbd>
      </div>
    </div>
  )
}
