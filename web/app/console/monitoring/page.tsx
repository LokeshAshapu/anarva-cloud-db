'use client'

import React, { useState, useEffect } from 'react'

export default function MonitoringPage() {
  const [cpuUsage, setCpuUsage] = useState<number>(14.2)
  const [memoryUsage, setMemoryUsage] = useState<number>(38.5)
  const [connections, setConnections] = useState<number>(12)
  const [iops, setIops] = useState<number>(450)

  useEffect(() => {
    const interval = setInterval(() => {
      setCpuUsage(Number((Math.random() * 15 + 10).toFixed(1)))
      setMemoryUsage(Number((Math.random() * 5 + 36).toFixed(1)))
      setConnections(Math.floor(Math.random() * 4 + 10))
      setIops(Math.floor(Math.random() * 100 + 400))
    }, 2500)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Observability & Time-Series Metrics</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Real-time infrastructure performance telemetry, CPU, memory, IOPS, and connection health.</p>
        </div>

        <span className="flex items-center gap-2 text-xs font-mono text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-3 py-1 rounded-lg">
          <span className="h-2 w-2 rounded-full bg-emerald-400 animate-ping"></span>
          Live Stream Active
        </span>
      </div>

      {/* Real-time Telemetry Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">CPU Utilization</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">{cpuUsage}%</div>
          <div className="w-full bg-slate-950 h-1.5 rounded-full overflow-hidden">
            <div className="bg-blue-500 h-full transition-all duration-500" style={{ width: `${cpuUsage}%` }}></div>
          </div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Memory Allocation</div>
          <div className="text-3xl font-extrabold text-emerald-400 font-mono">{memoryUsage}%</div>
          <div className="w-full bg-slate-950 h-1.5 rounded-full overflow-hidden">
            <div className="bg-emerald-500 h-full transition-all duration-500" style={{ width: `${memoryUsage}%` }}></div>
          </div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Active Pool Connections</div>
          <div className="text-3xl font-extrabold text-white font-mono">{connections} / 100</div>
          <div className="text-xs text-slate-400">Peak 18 connections today</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Storage IOPS</div>
          <div className="text-3xl font-extrabold text-purple-400 font-mono">{iops} IOPS</div>
          <div className="text-xs text-slate-400">Max Provisioned: 3,000 IOPS</div>
        </div>
      </div>

      {/* Simulated Time-Series Performance Chart */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-bold text-white">System CPU & Memory Time-Series</h2>
          <span className="text-xs text-slate-400 font-mono">Last 1 hour</span>
        </div>

        <div className="h-48 w-full bg-slate-950 rounded-xl border border-slate-800 flex items-end p-4 gap-2">
          {[20, 35, 28, 42, 50, 38, 24, 30, 45, 60, 52, 40, 32, 28, 36, 44, 38, 25].map((val, idx) => (
            <div key={idx} className="flex-1 bg-blue-600/30 hover:bg-blue-500/50 rounded-t transition" style={{ height: `${val}%` }}></div>
          ))}
        </div>
      </div>
    </div>
  )
}
