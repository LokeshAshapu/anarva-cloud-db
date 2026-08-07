'use client'

import React from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'

export function Sidebar() {
  const pathname = usePathname()

  const navItems = [
    { name: 'Overview', href: '/dashboard' },
    { name: 'Projects & Orgs', href: '/dashboard/projects' },
    { name: 'Managed Databases', href: '/dashboard/databases' },
    { name: 'SQL Query Console', href: '/dashboard/query' },
    { name: 'Backups & PITR', href: '/dashboard/backups' },
    { name: 'API Keys & Security', href: '/dashboard/apikeys' },
  ]

  return (
    <aside className="w-64 border-r border-slate-800 bg-slate-900/50 p-4 min-h-[calc(100vh-4rem)] flex flex-col justify-between">
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
              className={`flex items-center px-3 py-2 text-sm font-medium rounded-lg transition ${
                isActive
                  ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20'
                  : 'text-slate-400 hover:bg-slate-800/60 hover:text-slate-200'
              }`}
            >
              {item.name}
            </Link>
          )
        })}
      </div>

      <div className="p-3 bg-slate-900 border border-slate-800 rounded-lg text-xs text-slate-400 space-y-1">
        <div className="font-semibold text-slate-200">System Status</div>
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
          <span>All Microservices Operational</span>
        </div>
      </div>
    </aside>
  )
}
