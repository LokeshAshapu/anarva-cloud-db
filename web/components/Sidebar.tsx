'use client'

import React from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'

export function Sidebar() {
  const pathname = usePathname()

  const navItems = [
    { name: 'Overview', href: '/dashboard', icon: '📊' },
    { name: 'Projects & Orgs', href: '/dashboard/projects', icon: '📁' },
    { name: 'Managed Databases', href: '/dashboard/databases', icon: '⚡' },
    { name: 'Unstructured Storage', href: '/dashboard/storage', icon: '📦' },
    { name: 'SQL Query Console', href: '/dashboard/query', icon: '💻' },
    { name: 'Backups & PITR', href: '/dashboard/backups', icon: '💾' },
    { name: 'API Keys & Security', href: '/dashboard/apikeys', icon: '🗝️' },
  ]

  return (
    <aside className="hidden md:flex w-64 border-r border-slate-800 bg-slate-900/50 p-4 flex-col justify-between shrink-0">
      <div className="space-y-1">
        <div className="px-3 py-2 text-xs font-semibold text-slate-500 uppercase tracking-wider">
          Platform Controls
        </div>
        {navItems.map((item) => {
          const isActive = pathname === item.href
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-2.5 px-3 py-2.5 text-sm font-medium rounded-lg transition ${
                isActive
                  ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20 shadow-sm'
                  : 'text-slate-400 hover:bg-slate-800/60 hover:text-slate-200'
              }`}
            >
              <span>{item.icon}</span>
              <span>{item.name}</span>
            </Link>
          )
        })}
      </div>

      <div className="p-3 bg-slate-900 border border-slate-800 rounded-xl text-xs text-slate-400 space-y-1 mt-4">
        <div className="font-semibold text-slate-200">System Status</div>
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
          <span>Microservices Operational</span>
        </div>
      </div>
    </aside>
  )
}
