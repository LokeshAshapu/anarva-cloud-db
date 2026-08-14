'use client'

import React from 'react'
import { ResourceStatus } from '@/types/resource'

export interface CloudStatusProps {
  status: ResourceStatus | string
  className?: string
}

export function CloudStatus({ status, className = '' }: CloudStatusProps) {
  let badgeVariant: 'emerald' | 'amber' | 'red' | 'blue' | 'slate' = 'slate'
  let isPulse = false

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
      isPulse = true
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

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase border ${
        badgeVariant === 'emerald'
          ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
          : badgeVariant === 'blue'
          ? 'bg-blue-500/10 text-blue-400 border-blue-500/20'
          : badgeVariant === 'amber'
          ? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
          : badgeVariant === 'red'
          ? 'bg-red-500/10 text-red-400 border-red-500/20'
          : 'bg-slate-800 text-slate-400 border-slate-700'
      } ${className}`}
    >
      <span
        className={`h-1.5 w-1.5 rounded-full ${
          badgeVariant === 'emerald'
            ? 'bg-emerald-400'
            : badgeVariant === 'blue'
            ? 'bg-blue-400'
            : badgeVariant === 'amber'
            ? 'bg-amber-400'
            : badgeVariant === 'red'
            ? 'bg-red-400'
            : 'bg-slate-400'
        } ${isPulse ? 'animate-ping' : ''}`}
      />
      <span>{status}</span>
    </span>
  )
}
