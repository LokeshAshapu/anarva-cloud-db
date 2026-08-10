'use client'

import React, { useEffect, useState } from 'react'
import { useRouter, usePathname } from 'next/navigation'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const pathname = usePathname()
  const [authorized, setAuthorized] = useState(false)

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('access_token')
      if (!token) {
        setAuthorized(false)
        router.push('/login')
      } else {
        setAuthorized(true)
        if (pathname.includes('/databases')) {
          router.replace('/console/databases')
        } else if (pathname.includes('/storage')) {
          router.replace('/console/storage')
        } else if (pathname.includes('/apikeys')) {
          router.replace('/console/devtools')
        } else if (pathname.includes('/backups')) {
          router.replace('/console/backups')
        } else if (pathname.includes('/query')) {
          router.replace('/console/databases')
        } else {
          router.replace('/console')
        }
      }
    }
  }, [router, pathname])

  if (!authorized) {
    return (
      <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-4 text-center space-y-4">
        <div className="p-3 rounded-2xl bg-blue-600/10 border border-blue-500/20 text-blue-400 font-mono text-xs flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-blue-500 animate-ping"></span>
          Redirecting to Anarva Enterprise Cloud Console...
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col antialiased">
      {children}
    </div>
  )
}
