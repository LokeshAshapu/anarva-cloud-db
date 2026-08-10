'use client'

import React from 'react'
import { CloudButton } from './CloudButton'

export interface CloudEmptyStateProps {
  title: string
  description: string
  actionLabel?: string
  onAction?: () => void
  icon?: React.ReactNode
}

export function CloudEmptyState({ title, description, actionLabel, onAction, icon }: CloudEmptyStateProps) {
  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-8 text-center space-y-4 max-w-md mx-auto">
      {icon && <div className="flex justify-center text-slate-400">{icon}</div>}
      <div className="space-y-1">
        <h3 className="text-base font-bold text-white tracking-tight">{title}</h3>
        <p className="text-xs text-slate-400">{description}</p>
      </div>
      {actionLabel && onAction && (
        <div className="pt-2 flex justify-center">
          <CloudButton variant="primary" size="sm" onClick={onAction}>
            {actionLabel}
          </CloudButton>
        </div>
      )}
    </div>
  )
}
