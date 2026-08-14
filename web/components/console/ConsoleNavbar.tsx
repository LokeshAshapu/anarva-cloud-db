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
  const [selectedRegion, setSelectedRegion] = useState('ap-hyderabad-1')
  const [showServicesMenu, setShowServicesMenu] = useState(false)
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

  const servicesList = [
    { category: 'COMPUTE', name: 'ACE Containers / EC2', href: '/console/compute', icon: '⚡' },
    { category: 'DATABASES', name: 'PostgreSQL & MySQL / RDS', href: '/console/databases', icon: '🗄️' },
    { category: 'STORAGE', name: 'S3 Object Storage (AOS)', href: '/console/storage', icon: '🪣' },
    { category: 'NETWORKING', name: 'VPC, Subnets & Gateways', href: '/console/networking', icon: '🌐' },
    { category: 'LOAD BALANCERS', name: 'ALB & Edge Delivery', href: '/console/loadbalancers', icon: '🔀' },
    { category: 'SECURITY', name: 'IAM & Attack Defense', href: '/console/security', icon: '🛡️' },
    { category: 'OBSERVABILITY', name: 'CloudWatch & Metrics', href: '/console/monitoring', icon: '📊' },
    { category: 'BILLING', name: 'Billing & Cost Explorer', href: '/console/billing', icon: '💳' },
  ]

  const regions = [
    { id: 'ap-hyderabad-1', name: 'Asia Pacific (Hyderabad)', code: 'ap-hyderabad-1' },
    { id: 'us-east-1', name: 'US East (N. Virginia)', code: 'us-east-1' },
    { id: 'us-west-2', name: 'US West (Oregon)', code: 'us-west-2' },
    { id: 'eu-west-1', name: 'Europe (Frankfurt)', code: 'eu-west-1' },
    { id: 'ap-south-1', name: 'Asia Pacific (Mumbai)', code: 'ap-south-1' },
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

  const initials =
    userName
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase() || 'LA'

  return (
    <header className="border-b border-slate-800/90 bg-slate-950/95 backdrop-blur-xl sticky top-0 z-40 h-14 flex items-center justify-between px-3 sm:px-4 shadow-xl">
      {/* Left Brand, AWS Services Dropdown & Mobile Drawer Trigger */}
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
            ANARVA{' '}
            <span className="text-amber-400 font-extrabold uppercase text-[10px] sm:text-xs tracking-widest bg-amber-500/10 px-1.5 py-0.5 rounded border border-amber-500/20 shadow-sm">
              CONSOLE
            </span>
          </span>
        </Link>

        {/* AWS-Style Services Menu Dropdown */}
        <div className="relative hidden md:block">
          <button
            onClick={() => setShowServicesMenu(!showServicesMenu)}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-slate-900/90 hover:bg-slate-800 border border-slate-800 hover:border-slate-700 text-slate-200 text-xs font-semibold rounded-lg transition"
          >
            <span className="text-amber-400">❖</span>
            <span>Services</span>
            <span className="text-slate-400 text-[10px]">▾</span>
          </button>

          {showServicesMenu && (
            <div className="absolute top-full left-0 mt-2 w-72 bg-slate-950/95 backdrop-blur-xl border border-slate-800 rounded-xl shadow-2xl p-2 z-50 divide-y divide-slate-800/80">
              <div className="px-3 py-1.5 text-[10px] font-mono font-bold text-slate-400 uppercase tracking-widest">
                AWS & Cloud Services Matrix
              </div>
              <div className="py-1 space-y-0.5 max-h-80 overflow-y-auto">
                {servicesList.map((svc) => (
                  <Link
                    key={svc.name}
                    href={svc.href}
                    onClick={() => setShowServicesMenu(false)}
                    className="flex items-center gap-2.5 px-3 py-2 text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-900/80 rounded-lg transition"
                  >
                    <span>{svc.icon}</span>
                    <span>{svc.name}</span>
                  </Link>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Global Search Bar Trigger (⌘K) */}
        <button
          onClick={onOpenCommandPalette}
          className="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-slate-900/80 hover:bg-slate-900 border border-slate-800/80 hover:border-slate-700 text-slate-400 text-xs rounded-lg transition w-44 lg:w-64"
        >
          <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <span className="flex-1 text-left truncate">Search resources, services...</span>
          <kbd className="px-1.5 py-0.5 text-[10px] font-mono bg-slate-950 text-slate-400 border border-slate-800 rounded">⌘K</kbd>
        </button>
      </div>

      {/* Right Navbar Utility Controls */}
      <div className="flex items-center gap-2 sm:gap-3 font-mono text-xs">
        {/* AWS Account ID Badge */}
        <div className="hidden xl:flex items-center gap-1.5 px-2.5 py-1 bg-slate-900/80 border border-slate-800/80 rounded-lg text-slate-400 text-[11px]">
          <span className="text-slate-500 font-mono">Account ID:</span>
          <span className="text-slate-200 font-bold font-mono">4829-1029-3019</span>
        </div>

        {/* Region Selector */}
        <div className="relative">
          <select
            value={selectedRegion}
            onChange={(e) => setSelectedRegion(e.target.value)}
            className="bg-slate-900/90 border border-slate-800 hover:border-slate-700 text-slate-200 text-xs font-semibold rounded-lg px-2.5 py-1.5 focus:outline-none cursor-pointer"
          >
            {regions.map((r) => (
              <option key={r.id} value={r.id} className="bg-slate-950 text-slate-200">
                🌐 {r.name}
              </option>
            ))}
          </select>
        </div>

        {/* Notifications Dropdown */}
        <div className="relative">
          <button
            onClick={() => setShowNotifications(!showNotifications)}
            className="p-2 text-slate-400 hover:text-white rounded-lg hover:bg-slate-900 transition relative"
            aria-label="Notifications"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            <span className="absolute top-1 right-1 h-2 w-2 bg-blue-500 rounded-full animate-ping" />
            <span className="absolute top-1 right-1 h-2 w-2 bg-blue-500 rounded-full" />
          </button>

          {showNotifications && (
            <div className="absolute right-0 mt-2 w-80 bg-slate-950/95 backdrop-blur-xl border border-slate-800 rounded-xl shadow-2xl p-3 z-50 space-y-2">
              <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                <span className="font-bold text-white text-xs">Cloud Notifications</span>
                <span className="text-[10px] text-blue-400 font-mono">ALL SYSTEMS OPERATIONAL</span>
              </div>
              <div className="space-y-2 text-xs">
                <div className="p-2 bg-slate-900 rounded-lg">
                  <div className="font-semibold text-slate-200">Phase 23 Object Storage Active</div>
                  <div className="text-[10px] text-slate-400 mt-0.5">S3 compatibility layer enabled</div>
                </div>
                <div className="p-2 bg-slate-900 rounded-lg">
                  <div className="font-semibold text-slate-200">Attack & Bot Shield Enabled</div>
                  <div className="text-[10px] text-slate-400 mt-0.5">Turnstile / hCaptcha active</div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* User Profile Avatar */}
        <div className="relative">
          <button
            onClick={() => setShowProfileMenu(!showProfileMenu)}
            className="flex items-center gap-2 p-1 rounded-full hover:bg-slate-900 transition border border-slate-800"
          >
            <div className="h-7 w-7 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white font-bold text-xs flex items-center justify-center shadow-md">
              {initials}
            </div>
          </button>

          {showProfileMenu && (
            <div className="absolute right-0 mt-2 w-64 bg-slate-950/95 backdrop-blur-xl border border-slate-800 rounded-xl shadow-2xl p-3 z-50 space-y-3">
              <div className="border-b border-slate-800 pb-2">
                <div className="font-bold text-white text-xs">{userName}</div>
                <div className="text-[10px] text-slate-400 truncate">{userEmail}</div>
              </div>
              <div className="space-y-1">
                <Link
                  href="/console/security"
                  onClick={() => setShowProfileMenu(false)}
                  className="block px-2.5 py-1.5 text-xs text-slate-300 hover:text-white hover:bg-slate-900 rounded-lg"
                >
                  Security & Access Keys
                </Link>
                <Link
                  href="/console/billing"
                  onClick={() => setShowProfileMenu(false)}
                  className="block px-2.5 py-1.5 text-xs text-slate-300 hover:text-white hover:bg-slate-900 rounded-lg"
                >
                  Billing Dashboard
                </Link>
              </div>
              <button
                onClick={handleSignOut}
                className="w-full text-left px-2.5 py-1.5 text-xs text-red-400 hover:bg-red-500/10 rounded-lg font-semibold transition"
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
