'use client'

import React, { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'

interface GlobalCommandPaletteProps {
  isOpen: boolean
  onClose: () => void
}

export function GlobalCommandPalette({ isOpen, onClose }: GlobalCommandPaletteProps) {
  const router = useRouter()
  const [query, setQuery] = useState('')

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        if (isOpen) onClose()
        else {
          // Trigger open via parent state
        }
      }
      if (e.key === 'Escape' && isOpen) {
        onClose()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  if (!isOpen) return null

  const items = [
    { name: 'Home Infrastructure Dashboard', category: 'Services', href: '/console' },
    { name: 'Anarva Compute Engine (ACE)', category: 'Services', href: '/console/compute' },
    { name: 'Managed Database Clusters', category: 'Services', href: '/console/databases' },
    { name: 'Anarva Object Storage (AOS)', category: 'Services', href: '/console/storage' },
    { name: 'Virtual Private Cloud (VPC) & Security', category: 'Services', href: '/console/networking' },
    { name: 'IAM Users, Roles & Policies', category: 'Security', href: '/console/iam' },
    { name: 'Observability & Time-Series Metrics', category: 'Operations', href: '/console/monitoring' },
    { name: 'Point-in-Time Backups & Recovery', category: 'Operations', href: '/console/backups' },
    { name: 'Billing, Usage & Cost Analytics', category: 'Management', href: '/console/billing' },
    { name: 'API Keys & CLI SDK Tools', category: 'Developer Tools', href: '/console/devtools' },
  ]

  const filtered = items.filter(
    (i) => i.name.toLowerCase().includes(query.toLowerCase()) || i.category.toLowerCase().includes(query.toLowerCase())
  )

  const handleSelect = (href: string) => {
    router.push(href)
    onClose()
  }

  return (
    <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 flex items-start justify-center pt-20 p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-xl shadow-2xl overflow-hidden animate-in zoom-in-95">
        <div className="p-4 border-b border-slate-800 flex items-center gap-3">
          <svg className="w-5 h-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Type a command or search cloud resources..."
            className="w-full bg-transparent text-white text-sm focus:outline-none placeholder-slate-500"
            autoFocus
          />
          <button onClick={onClose} className="text-xs text-slate-500 hover:text-slate-300 font-mono">
            ESC
          </button>
        </div>

        <div className="max-h-80 overflow-y-auto p-2 divide-y divide-slate-800/50">
          {filtered.length === 0 ? (
            <div className="p-8 text-center text-xs text-slate-500">No matching cloud resources found.</div>
          ) : (
            filtered.map((item, idx) => (
              <button
                key={idx}
                onClick={() => handleSelect(item.href)}
                className="w-full text-left p-3 hover:bg-slate-800/60 rounded-xl flex items-center justify-between transition text-xs group"
              >
                <span className="font-semibold text-slate-200 group-hover:text-blue-400">{item.name}</span>
                <span className="px-2 py-0.5 bg-slate-950 border border-slate-800 text-[10px] text-slate-400 font-mono rounded">
                  {item.category}
                </span>
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
