'use client'

import React, { useEffect, useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { AnarvaLogo } from '@/components/AnarvaLogo'

export default function HomePage() {
  const router = useRouter()
  const [isLoggedIn, setIsLoggedIn] = useState(false)

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('access_token')
      setIsLoggedIn(!!token)
    }
  }, [])

  const handleLaunchConsole = (e: React.MouseEvent) => {
    e.preventDefault()
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('access_token')
      if (token) {
        router.push('/console')
      } else {
        router.push('/login')
      }
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col justify-between p-6 md:p-12">
      {/* Header Navigation */}
      <header className="max-w-7xl mx-auto w-full flex items-center justify-between">
        <div className="flex items-center gap-3">
          <AnarvaLogo className="h-14 w-14" />
          <span className="text-2xl font-bold tracking-tight text-white">
            Anarva <span className="text-blue-500">Cloud DB</span>
          </span>
        </div>

        <div className="flex items-center gap-4">
          <Link
            href="/login"
            className="px-4 py-2 text-sm font-medium text-slate-300 hover:text-white transition"
          >
            {isLoggedIn ? 'Dashboard' : 'Sign In'}
          </Link>
          <button
            onClick={handleLaunchConsole}
            className="px-5 py-2.5 text-sm font-semibold bg-blue-600 hover:bg-blue-500 text-white rounded-xl transition shadow-lg shadow-blue-600/25"
          >
            Open Dashboard
          </button>
        </div>
      </header>

      {/* Hero Content */}
      <main className="max-w-5xl mx-auto w-full text-center space-y-8 my-auto py-12">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-xs font-semibold text-blue-400">
          <span className="h-2 w-2 rounded-full bg-blue-400 animate-pulse"></span>
          Anarva Enterprise Managed Cloud Database Engine
        </div>

        <h1 className="text-5xl md:text-7xl font-extrabold text-white tracking-tight leading-tight">
          Next-Generation Managed <br />
          <span className="bg-gradient-to-r from-blue-400 via-cyan-400 to-indigo-400 bg-clip-text text-transparent">
            Anarva Database Infrastructure
          </span>
        </h1>

        <p className="text-lg md:text-xl text-slate-400 max-w-3xl mx-auto leading-relaxed">
          Provision Anarva Serverless PostgreSQL & MySQL instances in seconds with automated WAL backups, 
          real-time query parsing telemetry, zero-trust security, and global multi-tenancy.
        </p>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
          <button
            onClick={handleLaunchConsole}
            className="w-full sm:w-auto px-8 py-4 bg-blue-600 hover:bg-blue-500 text-white font-bold rounded-xl transition shadow-xl shadow-blue-600/25 text-lg cursor-pointer"
          >
            Launch Console →
          </button>
          <button
            onClick={handleLaunchConsole}
            className="w-full sm:w-auto px-8 py-4 bg-slate-900 hover:bg-slate-800 text-slate-200 font-bold rounded-xl transition border border-slate-800 text-lg cursor-pointer"
          >
            Try SQL Console
          </button>
        </div>

        {/* Platform Highlights */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-16 text-left">
          <div className="bg-slate-900/60 border border-slate-800 p-6 rounded-2xl space-y-2">
            <div className="text-blue-400 font-bold text-lg">⚡ Anarva Auto-Provisioning</div>
            <p className="text-xs text-slate-400">Deploy isolated database instances with automatic port assignment & credentials encryption.</p>
          </div>

          <div className="bg-slate-900/60 border border-slate-800 p-6 rounded-2xl space-y-2">
            <div className="text-cyan-400 font-bold text-lg">🛡️ Zero-Trust Security</div>
            <p className="text-xs text-slate-400">JWT token pair rotation, SHA-256 API Keys, AES-256 GCM encryption, and immutable audit logs.</p>
          </div>

          <div className="bg-slate-900/60 border border-slate-800 p-6 rounded-2xl space-y-2">
            <div className="text-indigo-400 font-bold text-lg">💾 Snapshot Backups</div>
            <p className="text-xs text-slate-400">Stream snapshots and WAL archives directly to Object Storage providers with one-click restore.</p>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="max-w-7xl mx-auto w-full text-center text-xs text-slate-500 py-6 border-t border-slate-900">
        © 2026 Anarva Cloud DB. Enterprise Production Edition.
      </footer>
    </div>
  )
}
