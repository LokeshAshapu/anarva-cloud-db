'use client'

import React from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { AnarvaLogo } from './AnarvaLogo'

export function Navbar() {
  const router = useRouter()

  const handleSignOut = () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token')
    }
    router.push('/login')
  }

  return (
    <header className="border-b border-slate-800 bg-slate-900/80 backdrop-blur sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <AnarvaLogo className="h-11 w-11" />
          <Link href="/dashboard" className="text-xl font-bold text-white tracking-tight flex items-center gap-1">
            Anarva <span className="text-blue-500">Cloud DB</span>
          </Link>
          <span className="px-2.5 py-0.5 text-xs font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-full">
            v1.0 Enterprise
          </span>
        </div>

        <div className="flex items-center gap-4">
          <Link
            href="/dashboard/query"
            className="px-3.5 py-1.5 text-sm font-medium bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg transition border border-slate-700"
          >
            SQL Console
          </Link>
          <button
            onClick={handleSignOut}
            className="px-3.5 py-1.5 text-sm font-medium bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-500/20 rounded-lg transition"
          >
            Sign Out
          </button>
        </div>
      </div>
    </header>
  )
}
