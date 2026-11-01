import React, { forwardRef } from 'react'
import { Search } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Input as PrimitiveInput } from '@/components/primitives/input'

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  icon?: React.ReactNode
  error?: string
}

/** Forge text input: React Aria input behind the existing API. */
export const Input = forwardRef<HTMLInputElement, InputProps>(({
  className,
  type = 'text',
  icon,
  error,
  ...props
}, ref) => (
  <div className="relative w-full">
    {icon && (
      <div className="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-text-3">
        {icon}
      </div>
    )}
    <PrimitiveInput
      ref={ref}
      type={type}
      className={cn(
        'h-8 rounded-md text-xs',
        icon ? 'pl-8 pr-2.5' : 'px-2.5',
        error && 'border-destructive',
        className
      )}
      {...props}
    />
    {error && <span className="text-caption text-destructive mt-1 block">{error}</span>}
  </div>
))

Input.displayName = 'Input'

export interface SearchPillProps extends React.InputHTMLAttributes<HTMLInputElement> {
  onSearch?: (value: string) => void
}

/** Forge search field with the `/` affordance. */
export const SearchPill: React.FC<SearchPillProps> = ({
  className,
  placeholder = 'Type / to search...',
  ...props
}) => (
  <div className={cn('relative flex items-center w-full max-w-sm', className)}>
    <Search className="absolute left-2.5 size-3.5 text-text-3 pointer-events-none" />
    <PrimitiveInput
      type="text"
      placeholder={placeholder}
      className="h-8 w-full rounded-md text-xs pl-8 pr-8"
      {...props}
    />
    <div className="absolute right-2 flex items-center pointer-events-none">
      <kbd className="text-micro bg-raised border border-hairline text-text-3 px-1.5 py-0.5 rounded font-mono">
        /
      </kbd>
    </div>
  </div>
)
