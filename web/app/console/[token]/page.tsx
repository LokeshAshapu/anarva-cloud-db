'use client'

import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
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
      <div className="flex flex-col items-center justify-center p-12 text-center space-y-4">
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
      case 'databases':
        return <DatabasesPage />
      case 'enc-7d4e11':
      case 'storage':
        return <UnstructuredStoragePage />
      case 'enc-2c6b4d':
      case 'projects':
        return <ProjectsPage />
      case 'enc-5f9e8a':
      case 'query':
        return <SQLConsolePage />
      case 'enc-1d3a7e':
      case 'backups':
        return <BackupsPage />
      case 'enc-9b2c4f':
      case 'apikeys':
        return <APIKeysPage />
      case 'enc-0a1b9c':
      default:
        return <DashboardOverview />
    }
  }

  return (
    <div className="w-full">
      {renderView()}
    </div>
  )
}
