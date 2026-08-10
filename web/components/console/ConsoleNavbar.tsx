'use client'

import React, { useState, useEffect } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { AnarvaLogo } from '../AnarvaLogo'
import { createClient } from '@/utils/supabase/client'

interface ConsoleNavbarProps {
  onOpenCommandPalette: () => void
  onToggleMobileMenu?: () => void
}

export function ConsoleNavbar({ onOpenCommandPalette, onToggleMobileMenu }: ConsoleNavbarProps) {
  const router = useRouter()
  const [selectedRegion, setSelectedRegion] = useState('us-east-1')
  const [showNotifications, setShowNotifications] = useState(false)
  const [showProfileMenu, setShowProfileMenu] = useState(false)
  const [userEmail, setUserEmail] = useState('lokeshashapu@gmail.com')
  const [userName, setUserName] = useState('Lokesh Ashapu')

  useEffect(() => {
    async function loadUser() {
      if (typeof window !== 'undefined') {
        const storedEmail = localStorage.getItem('anarva_user_email')
        const storedName = localStorage.getItem('anarva_user_name')
        if (storedEmail) setUserEmail(storedEmail)
        if (storedName) setUserName(storedName)

        try {
          const supabase = createClient()
          const { data } = await supabase.auth.getUser()
          if (data?.user?.email) {
            setUserEmail(data.user.email)
            localStorage.setItem('anarva_user_email', data.user.email)
            const metaName = data.user.user_metadata?.full_name
            if (metaName) {
              setUserName(metaName)
              localStorage.setItem('anarva_user_name', metaName)
            }
          }
        } catch (e) {
          console.log('Supabase user check notice:', e)
        }
      }
    }
    loadUser()
  }, [])

  const regions = [
    { id: 'us-east-1', name: 'US East (N. Virginia)' },
    { id: 'us-west-2', name: 'US West (Oregon)' },
    { id: 'eu-west-1', name: 'Europe (Frankfurt)' },
    { id: 'ap-south-1', name: 'Asia Pacific (Mumbai)' },
    { id: 'ap-southeast-1', name: 'Asia Pacific (Singapore)' },
  ]

  const handleSignOut = async () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token')
      localStorage.removeItem('anarva_user_email')
      localStorage.removeItem('anarva_user_name')
    }
    try {
      const supabase = createClient()
      await supabase.auth.signOut()
    } catch {}
    router.push('/login')
  }

  const initials = userName
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase() || 'LA'

  return (
    <header className="border-b border-slate-800 bg-slate-950/90 backdrop-blur sticky top-0 z-40 h-14 flex items-center justify-between px-3 sm:px-4">
      {/* Left Brand & Mobile Drawer Trigger */}
      <div className="flex items-center gap-2 sm:gap-4">
        {/* Mobile Hamburger Toggle Button */}
        <button
          onClick={() => onToggleMobileMenu && onToggleMobileMenu()}
          className="p-1.5 text-slate-400 hover:text-white rounded-lg lg:hidden hover:bg-slate-900 transition"
          aria-label="Open mobile menu"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>

        <Link href="/console" className="flex items-center gap-2">
          <AnarvaLogo className="h-7 w-7 sm:h-8 sm:w-8" />
          <span className="text-sm sm:text-base font-bold text-white tracking-tight flex items-center gap-1.5">
            ANARVA <span className="text-blue-500 font-extrabold uppercase text-[10px] sm:text-xs tracking-widest bg-blue-500/10 px-1.5 py-0.5 rounded border border-blue-500/20">CLOUD</span>
          </span>
        </Link>

        {/* Global Search Bar Trigger */}
        <button
          onClick={onOpenCommandPalette}
          className="hidden sm:flex items-center gap-3 px-3 py-1.5 bg-slate-900 hover:bg-slate-800 border border-slate-800 hover:border-slate-700 text-slate-400 text-xs rounded-lg transition w-48 md:w-64 justify-between"
        >
          <span className="flex items-center gap-2">
            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <span className="truncate">Search cloud...</span>
          </span>
          <kbd className="px-1.5 py-0.5 bg-slate-950 border border-slate-800 rounded text-[10px] text-slate-500 font-mono">⌘K</kbd>
        </button>
      </div>

      {/* Right Tools & Profile */}
      <div className="flex items-center gap-2 sm:gap-3 text-xs">
        {/* Region Selector */}
        <div className="relative hidden md:block">
          <select
            value={selectedRegion}
            onChange={(e) => setSelectedRegion(e.target.value)}
            className="bg-slate-900 border border-slate-800 text-slate-300 rounded-lg px-2.5 py-1.5 text-xs font-medium focus:outline-none focus:border-blue-500 cursor-pointer"
          >
            {regions.map((r) => (
              <option key={r.id} value={r.id}>
                {r.name}
              </option>
            ))}
          </select>
        </div>

        {/* Notifications Tray Toggle */}
        <div className="relative">
          <button
            onClick={() => setShowNotifications(!showNotifications)}
            className="p-1.5 text-slate-400 hover:text-slate-200 hover:bg-slate-900 rounded-lg transition relative"
            title="Notifications"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            <span className="absolute top-1 right-1 h-1.5 w-1.5 rounded-full bg-blue-500"></span>
          </button>

          {showNotifications && (
            <div className="absolute right-0 mt-2 w-72 sm:w-80 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl p-4 space-y-3 z-50 animate-in fade-in">
              <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                <span className="font-bold text-white text-xs">System Notifications</span>
                <span className="text-[10px] text-blue-400 font-mono">0 Alerts</span>
              </div>
              <div className="text-slate-400 text-xs py-4 text-center">
                All cloud systems operating normally.
              </div>
            </div>
          )}
        </div>

        {/* User Profile Menu */}
        <div className="relative">
          <button
            onClick={() => setShowProfileMenu(!showProfileMenu)}
            className="flex items-center gap-2 px-2 py-1 bg-slate-900 hover:bg-slate-800 border border-slate-800 rounded-lg text-slate-200 transition"
          >
            <span className="h-5 w-5 rounded-full bg-blue-600 flex items-center justify-center font-bold text-[10px] text-white">
              {initials}
            </span>
            <span className="hidden md:inline font-semibold">{userName}</span>
          </button>

          {showProfileMenu && (
            <div className="absolute right-0 mt-2 w-52 sm:w-56 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl p-2 space-y-1 z-50 animate-in fade-in">
              <div className="px-3 py-2 border-b border-slate-800 text-[11px]">
                <div className="font-bold text-white truncate">{userName}</div>
                <div className="text-slate-400 font-mono text-[10px] truncate">{userEmail}</div>
              </div>
              <Link href="/console/iam" className="block px-3 py-1.5 text-xs text-slate-300 hover:bg-slate-800 rounded-lg">
                IAM & Account Settings
              </Link>
              <Link href="/console/billing" className="block px-3 py-1.5 text-xs text-slate-300 hover:bg-slate-800 rounded-lg">
                Billing & Usage
              </Link>
              <button
                onClick={handleSignOut}
                className="w-full text-left px-3 py-1.5 text-xs text-red-400 hover:bg-red-500/10 rounded-lg font-semibold"
              >
                Sign Out
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
