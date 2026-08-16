'use client'

import React from 'react'

export interface CloudAlertProps {
  type?: 'info' | 'success' | 'warning' | 'error'
  title?: string
  requestId?: string
  onRetry?: () => void
  children: React.ReactNode
  className?: string
}

export function CloudAlert({
  type = 'info',
  title,
  requestId,
  onRetry,
  children,
  className = '',
}: CloudAlertProps) {
  const styles = {
    info: 'bg-sky-500/10 text-sky-300 border-sky-500/20',
    success: 'bg-emerald-500/10 text-emerald-300 border-emerald-500/20',
    warning: 'bg-amber-500/10 text-amber-300 border-amber-500/20',
    error: 'bg-red-500/10 text-red-300 border-red-500/20',
  }

  return (
    <div className={`p-4 rounded-xl border text-xs font-sans space-y-2 ${styles[type]} ${className}`}>
      {title && <div className="font-bold text-white text-xs tracking-tight">{title}</div>}
      <div className="leading-relaxed">{children}</div>

      {(requestId || onRetry) && (
        <div className="pt-1 flex items-center justify-between gap-2 border-t border-current/10 text-[11px] font-mono">
          {requestId && (
            <span className="opacity-80">
              Request ID: <code className="bg-slate-900/60 px-1.5 py-0.5 rounded text-white">{requestId}</code>
            </span>
          )}
          {onRetry && (
            <button
              onClick={onRetry}
              className="px-2 py-1 bg-slate-900/80 hover:bg-slate-900 text-white rounded font-bold transition-colors"
            >
              Retry
            </button>
          )}
        </div>
      )}
    </div>
  )
}
