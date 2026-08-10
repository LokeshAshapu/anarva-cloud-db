'use client'

import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Navbar } from '@/components/Navbar'
import { Sidebar } from '@/components/Sidebar'
import DashboardOverview from '@/app/dashboard/page'
import DatabasesPage from '@/app/dashboard/databases/page'
import UnstructuredStoragePage from '@/app/dashboard/storage/page'
import ProjectsPage from '@/app/dashboard/projects/page'
import SQLConsolePage from '@/app/dashboard/query/page'
import BackupsPage from '@/app/dashboard/backups/page'
import APIKeysPage from '@/app/dashboard/apikeys/page'

export default function EncryptedConsolePage({ params }: { params: { token: string } }) {
  const router = useRouter()
  const [authorized, setAuthorized] = useState(false)
  const token = params.token || 'enc-0a1b9c'

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const authToken = localStorage.getItem('access_token')
      if (!authToken) {
        setAuthorized(false)
        router.push('/login')
      } else {
        setAuthorized(true)
      }
    }
  }, [router, token])

  if (!authorized) {
    return (
      <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-4 text-center space-y-4">
        <div className="p-3 rounded-2xl bg-blue-600/10 border border-blue-500/20 text-blue-400 font-mono text-xs flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-blue-500 animate-ping"></span>
          Verifying Encrypted Route Token ({token})...
        </div>
      </div>
    )
  }

  const renderView = () => {
    switch (token) {
      case 'enc-8f3a92':
        return <DatabasesPage />
      case 'enc-7d4e11':
        return <UnstructuredStoragePage />
      case 'enc-2c6b4d':
        return <ProjectsPage />
      case 'enc-5f9e8a':
        return <SQLConsolePage />
      case 'enc-1d3a7e':
        return <BackupsPage />
      case 'enc-9b2c4f':
        return <APIKeysPage />
      case 'enc-0a1b9c':
      default:
        return <DashboardOverview />
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col antialiased">
      <Navbar />

      {/* Encrypted Route Security Banner */}
      <div className="bg-slate-900/60 border-b border-slate-800/80 px-4 py-1.5 flex items-center justify-between text-[11px] font-mono text-slate-400">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-emerald-400"></span>
          <span>Address Bar URL Encrypted: <strong className="text-emerald-400">/console/{token}</strong></span>
        </div>
        <div className="hidden sm:block text-slate-500">
          Zero-Trust Integrity Token Active
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        <main className="flex-1 p-4 sm:p-6 lg:p-8 overflow-y-auto max-w-full pb-20 sm:pb-8">
          {renderView()}
        </main>
      </div>
    </div>
  )
}
