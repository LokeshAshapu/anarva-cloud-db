'use client'

import React from 'react'

export interface CloudSkeletonProps {
  className?: string
  variant?: 'text' | 'card' | 'table' | 'metric' | 'detail'
}

export function CloudSkeleton({ className = 'h-4 w-full', variant = 'text' }: CloudSkeletonProps) {
  if (variant === 'card') {
    return (
      <div className={`p-4 bg-slate-900/60 border border-slate-800/80 rounded-xl space-y-3 animate-pulse ${className}`}>
        <div className="flex items-center justify-between">
          <div className="h-4 w-28 bg-slate-800 rounded"></div>
          <div className="h-4 w-12 bg-slate-800 rounded-full"></div>
        </div>
        <div className="h-6 w-1/2 bg-slate-800 rounded"></div>
        <div className="h-3 w-3/4 bg-slate-800/60 rounded"></div>
      </div>
    )
  }

  if (variant === 'table') {
    return (
      <div className={`border border-slate-800/80 rounded-xl overflow-hidden bg-slate-950 p-4 space-y-3 animate-pulse ${className}`}>
        <div className="h-6 w-full bg-slate-900 rounded mb-4"></div>
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="flex items-center justify-between gap-4 border-b border-slate-800/40 pb-2">
            <div className="h-4 w-36 bg-slate-800/80 rounded"></div>
            <div className="h-4 w-24 bg-slate-800/80 rounded"></div>
            <div className="h-4 w-20 bg-slate-800/80 rounded"></div>
            <div className="h-4 w-16 bg-slate-800/80 rounded"></div>
          </div>
        ))}
      </div>
    )
  }

  if (variant === 'metric') {
    return (
      <div className={`p-4 bg-slate-900/60 border border-slate-800/80 rounded-xl space-y-2 animate-pulse ${className}`}>
        <div className="h-3 w-20 bg-slate-800 rounded"></div>
        <div className="h-7 w-24 bg-slate-800 rounded"></div>
        <div className="h-3 w-32 bg-slate-800/60 rounded"></div>
      </div>
    )
  }

  return <div className={`bg-slate-800/60 animate-pulse rounded-lg ${className}`}></div>
}
