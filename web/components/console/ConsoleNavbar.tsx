'use client'

import React, { useState, useEffect } from 'react'
import Link from 'next/link'
import { createClient } from '@/utils/supabase/client'

interface ConsoleNavbarProps {
  onOpenCommandPalette: () => void
  onToggleMobileMenu: () => void
}

export function ConsoleNavbar({
  onOpenCommandPalette,
  onToggleMobileMenu,
}: ConsoleNavbarProps) {
  const [userEmail, setUserEmail] = useState('admin@anarva.io')

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('anarva_user_email')
      if (stored) setUserEmail(stored)

      try {
        const supabase = createClient()
        supabase.auth.getUser().then(({ data }) => {
          if (data?.user?.email) {
            setUserEmail(data.user.email)
            localStorage.setItem('anarva_user_email', data.user.email)
          }
        })
      } catch {}
    }
  }, [])

  return (
    <header className="h-14 bg-gray-950/90 backdrop-blur border-b border-gray-800/80 px-4 flex items-center justify-between z-30 sticky top-0">
      {/* Left: Brand Identity & Selectors */}
      <div className="flex items-center gap-4">
        {/* Mobile Toggle Button */}
        <button
          onClick={onToggleMobileMenu}
          className="lg:hidden p-1.5 text-gray-400 hover:text-white rounded hover:bg-gray-800/60"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>

        {/* Brand Logo & Name */}
        <Link href="/console" className="flex items-center gap-2.5 font-mono text-sm tracking-tight text-white font-extrabold group">
          <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center text-white shadow-lg shadow-cyan-500/20 group-hover:scale-105 transition-transform">
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
            </svg>
          </div>
          <span>ANARVA</span>
        </Link>

        {/* Divider */}
        <span className="hidden sm:block text-gray-800 font-light">/</span>

        {/* Organization & Project Context Selectors */}
        <div className="hidden sm:flex items-center gap-2 text-xs">
          <div className="flex items-center gap-1.5 px-2.5 py-1 bg-gray-900 border border-gray-800 rounded-md text-gray-300">
            <span className="w-2 h-2 rounded-full bg-cyan-400"></span>
            <span className="font-semibold">Org:</span>
            <span className="font-mono text-white">Anarva Production Org</span>
          </div>

          <div className="flex items-center gap-1.5 px-2.5 py-1 bg-gray-900 border border-gray-800 rounded-md text-gray-300">
            <span className="font-semibold">Project:</span>
            <span className="font-mono text-white">prod-main</span>
          </div>

          <div className="px-2 py-0.5 bg-emerald-500/10 border border-emerald-500/20 rounded text-[10px] font-mono font-bold text-emerald-400 tracking-wider">
            PRODUCTION
          </div>
        </div>
      </div>

      {/* Right: Command Palette, Notifications & User Profile */}
      <div className="flex items-center gap-3">
        {/* Command Palette Trigger */}
        <button
          onClick={onOpenCommandPalette}
          className="flex items-center gap-2 px-3 py-1.5 bg-gray-900 hover:bg-gray-800 border border-gray-800 text-gray-400 hover:text-gray-200 rounded-lg text-xs transition-colors"
        >
          <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <span className="hidden md:inline font-mono">Search resources...</span>
          <kbd className="hidden sm:inline-block px-1.5 py-0.5 text-[10px] font-mono bg-gray-800 border border-gray-700 rounded text-gray-400">
            ⌘K
          </kbd>
        </button>

        {/* Notifications Button */}
        <button className="p-1.5 text-gray-400 hover:text-white rounded-lg hover:bg-gray-800/60 relative">
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
          <span className="w-1.5 h-1.5 bg-cyan-400 rounded-full absolute top-1.5 right-1.5"></span>
        </button>

        {/* User Profile Avatar */}
        <div className="flex items-center gap-2 pl-2 border-l border-gray-800/80">
          <div className="w-7 h-7 rounded-full bg-cyan-500/20 border border-cyan-500/30 flex items-center justify-center font-mono text-xs font-bold text-cyan-400">
            {userEmail.charAt(0).toUpperCase()}
          </div>
          <span className="hidden lg:inline text-xs font-mono text-gray-300 truncate max-w-[140px]">{userEmail}</span>
        </div>
      </div>
    </header>
  )
}
