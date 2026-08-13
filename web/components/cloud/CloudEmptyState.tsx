'use client'

import React from 'react'
import { CloudButton } from './CloudButton'

export interface CloudEmptyStateProps {
  title: string
  description: string
  actionLabel?: string
  onAction?: () => void
  icon?: React.ReactNode
  docsLink?: string
}

export function CloudEmptyState({ title, description, actionLabel, onAction, icon, docsLink }: CloudEmptyStateProps) {
  return (
    <div className="py-12 px-4 text-center border border-slate-800/80 rounded-2xl bg-slate-950/40 space-y-4">
      {/* Sleek Icon Badge */}
      <div className="flex justify-center">
        <div className="w-12 h-12 rounded-2xl bg-slate-900 border border-slate-800 flex items-center justify-center text-slate-400 text-xl shadow-inner">
          {icon || (
            <svg className="w-6 h-6 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          )}
        </div>
      </div>

      {/* Title & Description */}
      <div className="max-w-md mx-auto space-y-1">
        <h3 className="text-sm font-bold text-white tracking-tight">{title}</h3>
        <p className="text-xs text-slate-400 leading-relaxed font-sans">{description}</p>
      </div>

      {/* Action Button & Docs */}
      <div className="pt-2 flex flex-col sm:flex-row items-center justify-center gap-3 text-xs">
        {actionLabel && onAction && (
          <CloudButton variant="primary" size="sm" onClick={onAction}>
            {actionLabel}
          </CloudButton>
        )}
        {docsLink && (
          <a
            href={docsLink}
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs font-mono text-slate-400 hover:text-blue-400 transition flex items-center gap-1 py-1 px-2"
          >
            Documentation ↗
          </a>
        )}
      </div>
    </div>
  )
}
