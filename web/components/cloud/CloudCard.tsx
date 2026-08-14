'use client'

import React from 'react'

export interface CloudCardProps {
  title?: React.ReactNode
  subtitle?: string
  action?: React.ReactNode
  children: React.ReactNode
  className?: string
}

export function CloudCard({ title, subtitle, action, children, className = '' }: CloudCardProps) {
  return (
    <div className={`bg-slate-900 border border-slate-800 rounded-xl p-5 sm:p-6 space-y-4 shadow-sm ${className}`}>
      {(title || action) && (
        <div className="flex items-center justify-between border-b border-slate-800/80 pb-3">
          <div>
            {typeof title === 'string' ? (
              <h3 className="text-base font-bold text-white tracking-tight">{title}</h3>
            ) : (
              title
            )}
            {subtitle && <p className="text-xs text-slate-400 mt-0.5">{subtitle}</p>}
          </div>
          {action && <div>{action}</div>}
        </div>
      )}
      <div>{children}</div>
    </div>
  )
}
