'use client'

import React from 'react'

export interface CloudAlertProps {
  type?: 'info' | 'success' | 'warning' | 'error'
  title?: string
  children: React.ReactNode
  className?: string
}

export function CloudAlert({ type = 'info', title, children, className = '' }: CloudAlertProps) {
  const styles = {
    info: 'bg-blue-500/10 text-blue-300 border-blue-500/20',
    success: 'bg-emerald-500/10 text-emerald-300 border-emerald-500/20',
    warning: 'bg-amber-500/10 text-amber-300 border-amber-500/20',
    error: 'bg-red-500/10 text-red-300 border-red-500/20',
  }

  return (
    <div className={`p-4 rounded-2xl border text-xs font-sans space-y-1 ${styles[type]} ${className}`}>
      {title && <div className="font-bold text-white text-xs">{title}</div>}
      <div>{children}</div>
    </div>
  )
}
