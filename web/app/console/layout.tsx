'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ConsoleNavbar } from '@/components/console/ConsoleNavbar'
import { ConsoleSidebar } from '@/components/console/ConsoleSidebar'
import { GlobalCommandPalette } from '@/components/console/GlobalCommandPalette'

export default function CloudConsoleLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const pathname = usePathname()
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const mobileNavLinks = [
    { name: 'Home Overview', href: '/console' },
    { name: 'Anarva Compute Engine (ACE)', href: '/console/compute' },
    { name: 'Managed Databases', href: '/console/databases' },
    { name: 'Object Storage (AOS)', href: '/console/storage' },
    { name: 'Networking (VPC)', href: '/console/networking' },
    { name: 'IAM & Access Control', href: '/console/iam' },
    { name: 'Observability & Metrics', href: '/console/monitoring' },
    { name: 'Backups & Recovery', href: '/console/backups' },
    { name: 'Audit Logs & Activity History', href: '/console/audit' },
    { name: 'Billing & Usage Costs', href: '/console/billing' },
    { name: 'Developer Tools & CLI', href: '/console/devtools' },
    { name: 'Platform Settings', href: '/console/settings' },
  ]

  const bottomTabItems = [
    {
      name: 'Home',
      href: '/console',
      icon: (
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 00-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
        </svg>
      ),
    },
    {
      name: 'Compute',
      href: '/console/compute',
      icon: (
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
        </svg>
      ),
    },
    {
      name: 'Databases',
      href: '/console/databases',
      icon: (
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
        </svg>
      ),
    },
    {
      name: 'Storage',
      href: '/console/storage',
      icon: (
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
        </svg>
      ),
    },
    {
      name: 'Billing',
      href: '/console/billing',
      icon: (
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 0 0 3 3z" />
        </svg>
      ),
    },
  ]

  return (
    <div className="h-screen overflow-hidden bg-slate-950 flex flex-col font-sans antialiased text-slate-100 selection:bg-blue-500 selection:text-white">
      {/* Enterprise Top Navigation Bar */}
      <ConsoleNavbar
        onOpenCommandPalette={() => setCommandPaletteOpen(true)}
        onToggleMobileMenu={() => setMobileMenuOpen(!mobileMenuOpen)}
      />

      {/* Main Console Body */}
      <div className="flex flex-1 overflow-hidden">
        <ConsoleSidebar />
        <main className="flex-1 p-3 sm:p-6 lg:p-8 overflow-y-auto max-w-full pb-20 lg:pb-8">
          {children}
        </main>
      </div>

      {/* Mobile Slide-Over Navigation Drawer */}
      {mobileMenuOpen && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur z-50 lg:hidden flex flex-col animate-in fade-in">
          <div className="p-4 bg-slate-900 border-b border-slate-800 flex items-center justify-between">
            <span className="font-bold text-white text-sm">Cloud Platform Navigation</span>
            <button
              onClick={() => setMobileMenuOpen(false)}
              className="p-1.5 text-slate-400 hover:text-white bg-slate-800 rounded-lg text-xs"
            >
              ✕ Close
            </button>
          </div>

          <div className="p-4 overflow-y-auto space-y-2 flex-1">
            {mobileNavLinks.map((link) => {
              const isActive = pathname === link.href || (link.href !== '/console' && pathname.startsWith(link.href))
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  onClick={() => setMobileMenuOpen(false)}
                  className={`block p-3 rounded-xl border text-xs font-semibold transition ${
                    isActive
                      ? 'bg-blue-600/10 text-blue-400 border-blue-500/30 font-bold'
                      : 'bg-slate-900/60 text-slate-300 border-slate-800/80 hover:border-slate-700'
                  }`}
                >
                  {link.name}
                </Link>
              )
            })}
          </div>
        </div>
      )}

      {/* Mobile Fixed Bottom Navigation Bar */}
      <nav className="lg:hidden fixed bottom-0 left-0 right-0 bg-slate-950/95 backdrop-blur border-t border-slate-800 z-40 px-2 py-1.5 flex items-center justify-around">
        {bottomTabItems.map((item) => {
          const isActive = pathname === item.href
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex flex-col items-center gap-1 px-3 py-1 rounded-lg text-[10px] font-semibold transition ${
                isActive ? 'text-blue-400 font-extrabold' : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <span>{item.icon}</span>
              <span>{item.name}</span>
            </Link>
          )
        })}
      </nav>

      {/* Global Command Palette (⌘K) */}
      <GlobalCommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
      />
    </div>
  )
}
