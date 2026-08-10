'use client'

import React, { useState } from 'react'
import { ConsoleNavbar } from '@/components/console/ConsoleNavbar'
import { ConsoleSidebar } from '@/components/console/ConsoleSidebar'
import { GlobalCommandPalette } from '@/components/console/GlobalCommandPalette'

export default function CloudConsoleLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false)

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col font-sans antialiased text-slate-100 selection:bg-blue-500 selection:text-white">
      {/* Enterprise Top Navigation Bar */}
      <ConsoleNavbar onOpenCommandPalette={() => setCommandPaletteOpen(true)} />

      {/* Main Console Body */}
      <div className="flex flex-1 overflow-hidden">
        <ConsoleSidebar />
        <main className="flex-1 p-4 sm:p-6 lg:p-8 overflow-y-auto max-w-full">
          {children}
        </main>
      </div>

      {/* Global Command Palette (⌘K) */}
      <GlobalCommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
      />
    </div>
  )
}
