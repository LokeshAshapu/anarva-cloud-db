'use client'

import React from 'react'

export interface CloudSkeletonProps {
  className?: string
}

export function CloudSkeleton({ className = 'h-4 w-full' }: CloudSkeletonProps) {
  return <div className={`bg-slate-800/60 animate-pulse rounded-lg ${className}`}></div>
}
