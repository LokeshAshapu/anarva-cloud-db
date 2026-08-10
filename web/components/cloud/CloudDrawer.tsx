'use client'

import React, { useEffect } from 'react'

export interface CloudDrawerProps {
  isOpen: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
  side?: 'right' | 'left'
}

export function CloudDrawer({ isOpen, onClose, title, children, side = 'right' }: CloudDrawerProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) onClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 flex">
      <div className={`bg-slate-900 border-l border-slate-800 w-full max-w-md h-full p-6 flex flex-col justify-between shadow-2xl animate-in slide-in-from-${side} ${side === 'right' ? 'ml-auto' : 'mr-auto'}`}>
        <div className="space-y-4 flex-1 overflow-y-auto">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <h3 className="text-base font-bold text-white tracking-tight">{title}</h3>
            <button onClick={onClose} className="p-1 text-slate-500 hover:text-slate-200 text-xs font-mono">
              ✕
            </button>
          </div>
          <div>{children}</div>
        </div>
      </div>
    </div>
  )
}
