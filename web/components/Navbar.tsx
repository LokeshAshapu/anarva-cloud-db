'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { useRouter, usePathname } from 'next/navigation'
import { AnarvaLogo } from './AnarvaLogo'

export function Navbar() {
  const router = useRouter()
  const pathname = usePathname()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const handleSignOut = () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token')
    }
    router.push('/login')
  }

  const navLinks = [
    { name: 'Overview', href: '/dashboard' },
    { name: 'Databases', href: '/dashboard/databases' },
    { name: 'Unstructured Storage', href: '/dashboard/storage' },
    { name: 'Backups', href: '/dashboard/backups' },
    { name: 'Projects', href: '/dashboard/projects' },
    { name: 'SQL Console', href: '/dashboard/query' },
    { name: 'API Keys', href: '/dashboard/apikeys' },
  ]

  return (
    <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand & Logo */}
        <div className="flex items-center gap-3">
          <Link href="/dashboard" className="flex items-center gap-2.5">
            <AnarvaLogo className="h-9 w-9 sm:h-11 sm:w-11" />
            <span className="text-lg sm:text-xl font-bold text-white tracking-tight">
              Anarva <span className="text-blue-500">Cloud DB</span>
            </span>
          </Link>
          <span className="hidden md:inline-block px-2.5 py-0.5 text-xs font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-full">
            v1.0 Enterprise
          </span>
        </div>

        {/* Desktop Navigation & Actions */}
        <div className="hidden md:flex items-center gap-4">
          <Link
            href="/dashboard/query"
            className="px-3.5 py-1.5 text-sm font-medium bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg transition border border-slate-700"
          >
            SQL Console ⚡
          </Link>
          <button
            onClick={handleSignOut}
            className="px-3.5 py-1.5 text-sm font-medium bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-500/20 rounded-lg transition cursor-pointer"
          >
            Sign Out
          </button>
        </div>

        {/* Mobile Hamburger Toggle Button */}
        <div className="flex md:hidden items-center gap-2">
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="p-2 bg-slate-800 text-slate-300 hover:text-white rounded-lg border border-slate-700 transition focus:outline-none"
            aria-label="Toggle navigation menu"
          >
            {mobileMenuOpen ? (
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            ) : (
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            )}
          </button>
        </div>
      </div>

      {/* Mobile Drawer Menu */}
      {mobileMenuOpen && (
        <div className="md:hidden bg-slate-900 border-b border-slate-800 px-4 py-4 space-y-3 animate-in slide-in-from-top">
          <div className="grid grid-cols-2 gap-2 text-xs font-semibold">
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                onClick={() => setMobileMenuOpen(false)}
                className={`p-2.5 rounded-lg border transition ${
                  pathname === link.href
                    ? 'bg-blue-600/10 text-blue-400 border-blue-500/30'
                    : 'bg-slate-950 text-slate-300 border-slate-800 hover:border-slate-700'
                }`}
              >
                {link.name}
              </Link>
            ))}
          </div>

          <div className="pt-2 border-t border-slate-800 flex gap-2">
            <Link
              href="/dashboard/query"
              onClick={() => setMobileMenuOpen(false)}
              className="flex-1 py-2 text-center text-xs font-semibold bg-slate-800 text-slate-200 rounded-lg border border-slate-700"
            >
              SQL Console ⚡
            </Link>
            <button
              onClick={handleSignOut}
              className="flex-1 py-2 text-center text-xs font-semibold bg-red-600/20 text-red-400 rounded-lg border border-red-500/20"
            >
              Sign Out
            </button>
          </div>
        </div>
      )}
    </header>
  )
}
