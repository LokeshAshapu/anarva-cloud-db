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
    positive: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20',
    negative: 'text-red-400 bg-red-500/10 border-red-500/20',
    neutral: 'text-slate-400 bg-slate-800/50 border-slate-700/50',
  }

  return (
    <div className={`bg-gradient-to-b from-slate-900/90 to-slate-950/90 backdrop-blur-xl border border-slate-800/80 hover:border-slate-700/80 rounded-2xl p-5 space-y-2 shadow-xl hover:shadow-2xl transition-all duration-300 relative overflow-hidden group ${className}`}>
      <div className="flex items-center justify-between">
        <span className="text-[10px] font-mono font-bold text-slate-400 uppercase tracking-widest">{label}</span>
        {trend && (
          <span className={`text-[10px] font-mono font-bold px-2 py-0.5 rounded-full border ${trendColors[trendType]}`}>
            {trend}
          </span>
        )}
      </div>
      <div className="text-2xl sm:text-3xl font-extrabold text-white font-mono tracking-tight drop-shadow-sm">{value}</div>
      {subtext && <div className="text-xs text-slate-400 font-sans">{subtext}</div>}
      
      {/* Decorative sparkline graphic line at bottom */}
      <div className="pt-2 flex items-end gap-1 h-3 opacity-30 group-hover:opacity-70 transition-opacity">
        <div className="w-1/6 bg-blue-500 rounded-t h-1.5" />
        <div className="w-1/6 bg-blue-500 rounded-t h-2.5" />
        <div className="w-1/6 bg-blue-500 rounded-t h-2" />
        <div className="w-1/6 bg-blue-500 rounded-t h-3" />
        <div className="w-1/6 bg-blue-500 rounded-t h-2.5" />
        <div className="w-1/6 bg-blue-500 rounded-t h-3.5" />
      </div>
    </div>
  )
}
