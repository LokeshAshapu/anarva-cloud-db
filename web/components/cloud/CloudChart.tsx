'use client'

import React from 'react'

export interface CloudChartProps {
  title: string
  data: number[]
  height?: number
}

export function CloudChart({ title, data, height = 180 }: CloudChartProps) {
  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-3">
      <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider">{title}</h4>
      <div className="w-full bg-slate-950 rounded-xl border border-slate-800 p-4 flex items-end gap-2" style={{ height: `${height}px` }}>
        {data.map((val, idx) => (
          <div
            key={idx}
            className="flex-1 bg-blue-600/40 hover:bg-blue-500 rounded-t transition-all duration-300"
            style={{ height: `${Math.min(Math.max(val, 10), 100)}%` }}
          ></div>
        ))}
      </div>
    </div>
  )
}
