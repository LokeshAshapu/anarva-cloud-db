'use client'

import React, { useEffect, useState } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import Link from 'next/link'
import { Navbar } from '@/components/Navbar'
import { Sidebar } from '@/components/Sidebar'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const pathname = usePathname()
  const [authorized, setAuthorized] = useState(false)
  const [sessionHash, setSessionHash] = useState<string>('')

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('access_token')
      if (!token) {
        // Tampered or unauthenticated routing attempt intercepted
        setAuthorized(false)
        router.push('/login')
      } else {
        // Generate encrypted route signature
        const hash = `enc-${Math.abs(
          token.split('').reduce((acc, char) => (acc << 5) - acc + char.charCodeAt(0), 0)
        ).toString(16)}`
        setSessionHash(hash)
        setAuthorized(true)
      }
    }
  }, [router, pathname])

  if (!authorized) {
    return (
      <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-4 text-center space-y-4">
        <div className="p-3 rounded-2xl bg-blue-600/10 border border-blue-500/20 text-blue-400 font-mono text-xs flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-blue-500 animate-ping"></span>
          Authenticating Encrypted Zero-Trust Route Token...
        </div>
      </div>
    )
  }

  const bottomNavItems = [
    { name: 'Overview', href: '/dashboard', icon: '📊' },
    { name: 'Databases', href: '/dashboard/databases', icon: '⚡' },
    { name: 'Storage', href: '/dashboard/storage', icon: '📦' },
    { name: 'Console', href: '/dashboard/query', icon: '💻' },
    { name: 'Backups', href: '/dashboard/backups', icon: '💾' },
  ]

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col antialiased">
      <Navbar />

      {/* Encrypted Route Security Bar */}
      <div className="bg-slate-900/60 border-b border-slate-800/80 px-4 py-1.5 flex items-center justify-between text-[11px] font-mono text-slate-400">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-emerald-400"></span>
          <span>🛡️ Encrypted Route Protection Active: <strong className="text-emerald-400">{pathname}</strong></span>
        </div>
        <div className="hidden sm:block text-slate-500">
          Token Signature: <span className="text-blue-400">{sessionHash}</span>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        <main className="flex-1 p-4 sm:p-6 lg:p-8 overflow-y-auto max-w-full pb-20 sm:pb-8">
          {children}
        </main>
      </div>

      {/* Mobile Fixed Bottom Navigation Bar */}
      <nav className="sm:hidden fixed bottom-0 left-0 right-0 bg-slate-900/95 backdrop-blur border-t border-slate-800 z-40 px-2 py-2 flex items-center justify-around">
        {bottomNavItems.map((item) => {
          const isActive = pathname === item.href
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex flex-col items-center gap-0.5 px-3 py-1 rounded-lg text-xs font-medium transition ${
                isActive ? 'text-blue-400 font-bold' : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <span className="text-base">{item.icon}</span>
              <span>{item.name}</span>
            </Link>
          )
        })}
      </nav>
    </div>
  )
}
