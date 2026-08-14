'use client'

import React from 'react'
import { ResourceStatus } from '@/types/resource'

export interface CloudStatusProps {
  status: ResourceStatus | string
  className?: string
}

export function CloudStatus({ status, className = '' }: CloudStatusProps) {
  let badgeVariant: 'emerald' | 'amber' | 'red' | 'blue' | 'slate' = 'slate'

  switch (status.toUpperCase()) {
    case 'AVAILABLE':
    case 'RUNNING':
    case 'COMPLETED':
    case 'ACTIVE':
    case 'HEALTHY':
    case 'IN_SYNC':
      badgeVariant = 'emerald'
      break
    case 'PROVISIONING':
    case 'TESTING':
    case 'MAINTENANCE':
      badgeVariant = 'blue'
      break
    case 'DEGRADED':
    case 'STOPPED':
    case 'PENDING':
    case 'DRIFTED':
      badgeVariant = 'amber'
      break
    case 'FAILED':
    case 'TERMINATED':
      badgeVariant = 'red'
      break
    default:
      badgeVariant = 'slate'
  }

  const dotColors = {
    emerald: 'bg-emerald-400',
    blue: 'bg-blue-400',
    amber: 'bg-amber-400',
    red: 'bg-red-400',
    slate: 'bg-slate-400',
  }

  return (
    <span
      className={`inline-flex items-center gap-2 px-2.5 py-1 rounded-full text-[10px] font-mono font-bold uppercase border shadow-sm ${
        badgeVariant === 'emerald'
          ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20 shadow-emerald-500/5'
          : badgeVariant === 'blue'
          ? 'bg-blue-500/10 text-blue-400 border-blue-500/20 shadow-blue-500/5'
          : badgeVariant === 'amber'
          ? 'bg-amber-500/10 text-amber-400 border-amber-500/20 shadow-amber-500/5'
          : badgeVariant === 'red'
          ? 'bg-red-500/10 text-red-400 border-red-500/20 shadow-red-500/5'
          : 'bg-slate-800/80 text-slate-400 border-slate-700/80'
      } ${className}`}
    >
      <span className="relative flex h-2 w-2">
        <span
          className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${
            dotColors[badgeVariant]
          }`}
        />
        <span className={`relative inline-flex rounded-full h-2 w-2 ${dotColors[badgeVariant]}`} />
      </span>
      <span>{status}</span>
    </span>
  )
}
