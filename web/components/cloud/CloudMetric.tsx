'use client'

import React from 'react'

export interface CloudMetricProps {
  label: string
  value: string | number
  subtext?: string
  trend?: string
  trendType?: 'positive' | 'negative' | 'neutral'
  className?: string
}

export function CloudMetric({ label, value, subtext, trend, trendType = 'neutral', className = '' }: CloudMetricProps) {
  const trendColors = {
    positive: 'text-emerald-400',
    negative: 'text-red-400',
    neutral: 'text-slate-400',
  }

  return (
    <div className={`bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2 ${className}`}>
      <div className="flex items-center justify-between">
        <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">{label}</span>
        {trend && <span className={`text-xs font-mono font-bold ${trendColors[trendType]}`}>{trend}</span>}
      </div>
      <div className="text-2xl sm:text-3xl font-extrabold text-white font-mono tracking-tight">{value}</div>
      {subtext && <div className="text-xs text-slate-400 font-sans">{subtext}</div>}
    </div>
  )
}
