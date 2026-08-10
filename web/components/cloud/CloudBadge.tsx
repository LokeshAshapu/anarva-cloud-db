'use client'

import React from 'react'

export interface CloudBadgeProps {
  variant?: 'blue' | 'emerald' | 'amber' | 'red' | 'purple' | 'slate'
  children: React.ReactNode
  className?: string
}

export function CloudBadge({ variant = 'blue', children, className = '' }: CloudBadgeProps) {
  const styles = {
    blue: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
    emerald: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    amber: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
    red: 'bg-red-500/10 text-red-400 border-red-500/20',
    purple: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
    slate: 'bg-slate-800 text-slate-400 border-slate-700',
  }

  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-mono font-bold border ${styles[variant]} ${className}`}>
      {children}
    </span>
  )
}
