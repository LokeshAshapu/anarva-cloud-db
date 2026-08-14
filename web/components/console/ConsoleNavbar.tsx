'use client'

import React, { useState, useEffect } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { AnarvaLogo } from '../AnarvaLogo'
import { RegionSelector } from '../cloud/RegionSelector'
import { createClient } from '@/utils/supabase/client'

interface ConsoleNavbarProps {
  onOpenCommandPalette: () => void
  onToggleMobileMenu?: () => void
}

export function ConsoleNavbar({ onOpenCommandPalette, onToggleMobileMenu }: ConsoleNavbarProps) {
  const router = useRouter()
  const [showNotifications, setShowNotifications] = useState(false)
  const [showProfileMenu, setShowProfileMenu] = useState(false)
  const [userEmail, setUserEmail] = useState('operator@anarva.internal')
  const [userName, setUserName] = useState('Cloud Operator')
  const [notifications, setNotifications] = useState([
    {
      id: 'notif-1',
      title: 'S3 Object Storage Online',
      desc: 'Phase 23 storage engine ready',
      href: '/console/storage',
      time: 'Just now',
      unread: true,
    },
    {
      id: 'notif-2',
      title: 'Attack Protection Active',
      desc: 'Captcha & leaked password shield',
      href: '/console/security',
      time: '5m ago',
      unread: true,
    },
  ])

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

  const handleNotificationClick = (href: string) => {
    setShowNotifications(false)
    router.push(href)
  }

  const handleClearNotifications = () => {
    setNotifications([])
  }

  const unreadCount = notifications.filter((n) => n.unread).length

  const initials =
    userName
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase() || 'LA'

  return (
    <header className="border-b border-slate-800 bg-slate-950/95 backdrop-blur sticky top-0 z-40 h-14 flex items-center justify-between px-3 sm:px-4">
      {/* Click-away Backdrop overlay when any dropdown is open */}
      {(showNotifications || showProfileMenu) && (
        <div
          className="fixed inset-0 z-40 bg-transparent"
          onClick={() => {
            setShowNotifications(false)
            setShowProfileMenu(false)
          }}
        />
      )}

      {/* Left Brand & Mobile Drawer Trigger */}
      <div className="flex items-center gap-2 sm:gap-4 z-50">
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
            ANARVA{' '}
            <span className="text-blue-500 font-extrabold uppercase text-[10px] sm:text-xs tracking-widest bg-blue-500/10 px-1.5 py-0.5 rounded border border-blue-500/20">
              CLOUD
            </span>
          </span>
        </Link>

        {/* Global Search Bar Trigger */}
        <button
          onClick={onOpenCommandPalette}
          className="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-slate-900 border border-slate-800 hover:border-slate-700 text-slate-400 text-xs rounded-xl transition w-48 lg:w-64"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <span className="flex-1 text-left truncate">Search resources, databases...</span>
          <kbd className="px-1.5 py-0.5 text-[10px] font-mono bg-slate-950 text-slate-400 border border-slate-800 rounded">⌘K</kbd>
        </button>
      </div>

      {/* Right Region, Notifications & Profile */}
      <div className="flex items-center gap-2 sm:gap-3 z-50">
        {/* Region Selector */}
        <div className="hidden sm:block">
          <RegionSelector />
        </div>

        {/* Notifications Dropdown */}
        <div className="relative">
          <button
            onClick={() => {
              setShowProfileMenu(false)
              setShowNotifications(!showNotifications)
            }}
            className="p-2 text-slate-400 hover:text-white rounded-xl hover:bg-slate-900 transition relative focus:outline-none"
            aria-label="Notifications"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            {unreadCount > 0 && <span className="absolute top-1.5 right-1.5 h-2 w-2 bg-blue-500 rounded-full" />}
          </button>

          {showNotifications && (
            <div className="absolute right-0 mt-2 w-80 bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl p-4 z-50 space-y-3 font-sans">
              <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                <span className="font-bold text-white text-xs">Notifications</span>
                {notifications.length > 0 ? (
                  <button
                    onClick={handleClearNotifications}
                    className="text-[10px] text-blue-400 hover:underline font-mono font-bold"
                  >
                    Clear All
                  </button>
                ) : (
                  <span className="text-[10px] text-slate-500 font-mono font-bold">ALL CLEAR</span>
                )}
              </div>

              {notifications.length === 0 ? (
                <div className="text-xs text-slate-500 py-4 text-center">No new notifications</div>
              ) : (
                <div className="space-y-2 text-xs">
                  {notifications.map((n) => (
                    <button
                      key={n.id}
                      onClick={() => handleNotificationClick(n.href)}
                      className="w-full text-left p-3 bg-slate-950 hover:bg-slate-800/80 border border-slate-800/80 hover:border-blue-500/40 rounded-xl transition cursor-pointer active:scale-[0.98] block"
                    >
                      <div className="font-bold text-slate-200 flex items-center justify-between">
                        <span>{n.title}</span>
                        <span className="text-[9px] text-slate-500 font-mono font-normal">{n.time}</span>
                      </div>
                      <div className="text-[11px] text-slate-400 mt-1">{n.desc}</div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* User Profile */}
        <div className="relative">
          <button
            onClick={() => {
              setShowNotifications(false)
              setShowProfileMenu(!showProfileMenu)
            }}
            className="flex items-center gap-2 p-1 rounded-full hover:bg-slate-900 transition focus:outline-none"
          >
            <div className="h-8 w-8 rounded-full bg-blue-600 text-white font-bold text-xs flex items-center justify-center border border-blue-400/30">
              {initials}
            </div>
          </button>

          {showProfileMenu && (
            <div className="absolute right-0 mt-2 w-64 bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl p-4 z-50 space-y-3 font-sans">
              <div className="border-b border-slate-800 pb-2">
                <div className="font-bold text-white text-sm">{userName}</div>
                <div className="text-xs text-slate-400 font-mono truncate">{userEmail}</div>
              </div>
              <div className="space-y-1 text-xs">
                <Link
                  href="/console/security"
                  onClick={() => setShowProfileMenu(false)}
                  className="block px-3 py-2 text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl font-medium"
                >
                  Security Controls
                </Link>
                <Link
                  href="/console/billing"
                  onClick={() => setShowProfileMenu(false)}
                  className="block px-3 py-2 text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl font-medium"
                >
                  Billing & Quotas
                </Link>
              </div>
              <button
                onClick={handleSignOut}
                className="w-full text-left px-3 py-2 text-xs text-red-400 hover:bg-red-500/10 rounded-xl font-semibold transition"
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
