'use client'

import React from 'react'
import { CloudButton } from './CloudButton'

export interface CloudEmptyStateProps {
  title: string
  description: string
  actionLabel?: string
  onAction?: () => void
  icon?: React.ReactNode
  features?: { title: string; desc: string; icon?: string }[]
}

export function CloudEmptyState({ title, description, actionLabel, onAction, icon, features }: CloudEmptyStateProps) {
  return (
    <div className="space-y-6 py-4">
      {/* Hero Banner */}
      <div className="bg-gradient-to-br from-slate-900 via-slate-900 to-blue-950/40 border border-slate-800 rounded-3xl p-8 space-y-6 text-center shadow-2xl relative overflow-hidden">
        <div className="absolute top-0 right-0 -mt-8 -mr-8 w-48 h-48 bg-blue-500/10 rounded-full blur-3xl pointer-events-none" />
        
        <div className="max-w-xl mx-auto space-y-3 relative z-10">
          {icon && <div className="flex justify-center text-4xl mb-2">{icon}</div>}
          
          <div className="inline-flex items-center gap-2 px-3 py-1 bg-blue-500/10 border border-blue-500/20 rounded-full text-blue-400 text-xs font-mono font-bold">
            ⚡ ANARVA CLOUD PLATFORM ENGINE
          </div>

          <h3 className="text-xl sm:text-2xl font-extrabold text-white tracking-tight">{title}</h3>
          <p className="text-slate-300 text-xs sm:text-sm leading-relaxed">{description}</p>
          
          {actionLabel && onAction && (
            <div className="pt-3 flex justify-center">
              <CloudButton variant="primary" size="sm" onClick={onAction}>
                {actionLabel}
              </CloudButton>
            </div>
          )}
        </div>
      </div>

      {/* Feature Highlights Grid */}
      {features && features.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 font-mono text-xs">
          {features.map((f, i) => (
            <div key={i} className="p-4 bg-slate-900/60 border border-slate-800 rounded-2xl space-y-1.5">
              <div className="text-base">{f.icon || '🚀'}</div>
              <h4 className="font-bold text-white text-xs">{f.title}</h4>
              <p className="text-slate-400 text-[11px] font-sans leading-relaxed">{f.desc}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
