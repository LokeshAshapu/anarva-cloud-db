'use client'

import React from 'react'

export interface CloudButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'danger' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  isLoading?: boolean
  children: React.ReactNode
}

export function CloudButton({
  variant = 'primary',
  size = 'md',
  isLoading = false,
  children,
  className = '',
  disabled,
  ...props
}: CloudButtonProps) {
  const baseStyles =
    'font-semibold rounded-xl transition duration-150 flex items-center justify-center gap-2 focus:outline-none focus:ring-2 focus:ring-blue-500/50 disabled:opacity-50 disabled:cursor-not-allowed'

  const sizeStyles = {
    sm: 'px-3 py-1.5 text-xs',
    md: 'px-4 py-2 text-xs sm:text-sm',
    lg: 'px-6 py-3 text-sm sm:text-base',
  }

  const variantStyles = {
    primary: 'bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-600/20 border border-blue-500/30',
    secondary: 'bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700',
    outline: 'bg-transparent hover:bg-slate-900 text-slate-300 border border-slate-800 hover:border-slate-700',
    danger: 'bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-500/30',
    ghost: 'bg-transparent hover:bg-slate-900 text-slate-400 hover:text-slate-200',
  }

  return (
    <button
      className={`${baseStyles} ${sizeStyles[size]} ${variantStyles[variant]} ${className}`}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading && <span className="h-3 w-3 rounded-full border-2 border-current border-t-transparent animate-spin"></span>}
      {children}
    </button>
  )
}
