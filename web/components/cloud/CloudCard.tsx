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
    <div className={`bg-gradient-to-b from-slate-900/90 to-slate-950/90 backdrop-blur-xl border border-slate-800/80 hover:border-slate-700/80 rounded-2xl p-5 sm:p-6 space-y-4 shadow-2xl transition-all duration-300 relative overflow-hidden group ${className}`}>
      {/* Subtle top accent gradient line */}
      <div className="absolute top-0 left-0 right-0 h-[1px] bg-gradient-to-r from-transparent via-blue-500/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
      
      {(title || action) && (
        <div className="flex items-center justify-between border-b border-slate-800/80 pb-3.5">
          <div>
            {typeof title === 'string' ? (
              <h3 className="text-base font-bold text-white tracking-tight flex items-center gap-2">
                {title}
              </h3>
            ) : (
              title
            )}
            {subtitle && <p className="text-xs text-slate-400 mt-0.5 font-sans">{subtitle}</p>}
          </div>
          {action && <div>{action}</div>}
        </div>
      )}
      <div>{children}</div>
    </div>
  )
}
