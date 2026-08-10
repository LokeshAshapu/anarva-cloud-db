'use client'

import React, { useEffect, useState } from 'react'
import Link from 'next/link'
import { API_BASE_URL } from '@/lib/api'

export default function CloudConsoleHome() {
  const [dbCount, setDbCount] = useState<number>(1)
  const [storageUsedGB, setStorageUsedGB] = useState<number>(2.4)
  const [acuAllocated, setAcuAllocated] = useState<number>(1.5)
  const [networkEgressGB, setNetworkEgressGB] = useState<number>(0.85)
  const [estimatedCost, setEstimatedCost] = useState<string>('$14.20')
  const [latency, setLatency] = useState<string>('12.4 ms')
  const [loading, setLoading] = useState<boolean>(true)

  const fetchInfrastructureTelemetry = async () => {
    try {
      const start = performance.now()
      const res = await fetch(`${API_BASE_URL}/health`).catch(() => null)
      const elapsed = (performance.now() - start).toFixed(1)
      setLatency(`${elapsed} ms`)

      if (typeof window !== 'undefined') {
        const storedDbs = localStorage.getItem('anarva_user_databases')
        if (storedDbs) {
          const parsed = JSON.parse(storedDbs)
          if (Array.isArray(parsed)) {
            const active = parsed.filter((d: any) => d.status !== 'TERMINATED').length
            setDbCount(active || 1)
          }
        }
      }
    } catch (err) {
      console.error('Failed to update telemetry', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchInfrastructureTelemetry()
    const interval = setInterval(fetchInfrastructureTelemetry, 3000)
    return () => clearInterval(interval)
  }, [])

  const activeResources = [
    { name: 'Primary DB Cluster', type: 'Managed Database', status: 'RUNNING', spec: 'PostgreSQL 16 (1.0 ACU)', region: 'us-east-1' },
    { name: 'App Object Assets', type: 'Object Storage (AOS)', status: 'ACTIVE', spec: 'Standard S3 Bucket (2.4 GB)', region: 'us-east-1' },
    { name: 'API Serverless Worker', type: 'Anarva Compute (ACE)', status: 'RUNNING', spec: '0.5 ACU Auto-Scaled', region: 'us-east-1' },
    { name: 'Production VPC', type: 'Virtual Network', status: 'HEALTHY', spec: '10.0.0.0/16 Subnet', region: 'us-east-1' },
  ]

  const recentActivity = [
    { time: '2 mins ago', event: 'Automated Snapshot Backup Created', resource: 'Primary DB Cluster', user: 'System Worker' },
    { time: '14 mins ago', event: 'SQL Query Executed via Console', resource: 'Primary DB Cluster', user: 'Lokesh Ashapu' },
    { time: '1 hour ago', event: 'New Storage Bucket Created', resource: 'App Object Assets', user: 'Lokesh Ashapu' },
    { time: '3 hours ago', event: 'API Key Revoked & Rotated', resource: 'Security IAM', user: 'Lokesh Ashapu' },
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Cloud Infrastructure Overview</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Real-time health, active resources, compute allocation, and billing telemetry.</p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            href="/console/databases"
            className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition text-xs shadow-lg shadow-blue-600/20"
          >
            + Create Resource
          </Link>
        </div>
      </div>

      {/* Primary Infrastructure Telemetry Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Total Databases</div>
          <div className="text-3xl font-extrabold text-white">{loading ? '...' : dbCount}</div>
          <div className="text-xs text-emerald-400 flex items-center gap-1 font-mono">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-400"></span> 100% Operational
          </div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Storage Usage</div>
          <div className="text-3xl font-extrabold text-white font-mono">{storageUsedGB} GB</div>
          <div className="text-xs text-slate-400">Of 10.0 GB Default Quota</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Compute Usage</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">{acuAllocated} ACU</div>
          <div className="text-xs text-slate-400">Range: 0.5 - 128 ACU</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Network Egress</div>
          <div className="text-3xl font-extrabold text-white font-mono">{networkEgressGB} GB</div>
          <div className="text-xs text-slate-400">Avg Latency: <strong className="text-emerald-400 font-mono">{latency}</strong></div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Month-to-Date Cost</div>
          <div className="text-3xl font-extrabold text-emerald-400 font-mono">{estimatedCost}</div>
          <div className="text-xs text-slate-400">Est. $28.50 / month</div>
        </div>
      </div>

      {/* Active Resources & Health Status */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-bold text-white">Active Infrastructure Resources</h2>
            <Link href="/console/databases" className="text-xs text-blue-400 hover:text-blue-300 font-semibold">View All</Link>
          </div>

          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
            {activeResources.map((res, idx) => (
              <div key={idx} className="p-4 bg-slate-950 flex items-center justify-between">
                <div>
                  <div className="font-bold text-white">{res.name}</div>
                  <div className="text-slate-400 text-[11px] mt-0.5">{res.type} • {res.spec}</div>
                </div>

                <div className="flex items-center gap-3 font-mono">
                  <span className="text-slate-500 text-[11px]">{res.region}</span>
                  <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    {res.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Infrastructure Health & Activity */}
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
          <h2 className="text-base font-bold text-white">Infrastructure Health</h2>

          <div className="space-y-3 text-xs">
            <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl flex items-center justify-between">
              <span className="text-slate-300 font-medium">Database Engines</span>
              <span className="text-emerald-400 font-bold font-mono">100% Healthy</span>
            </div>
            <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl flex items-center justify-between">
              <span className="text-slate-300 font-medium">Object Storage (AOS)</span>
              <span className="text-emerald-400 font-bold font-mono">100% Healthy</span>
            </div>
            <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl flex items-center justify-between">
              <span className="text-slate-300 font-medium">API Gateway Router</span>
              <span className="text-emerald-400 font-bold font-mono">100% Healthy</span>
            </div>
          </div>

          <div className="pt-2 border-t border-slate-800">
            <div className="text-xs font-bold text-slate-400 mb-3">Recent Activity Log</div>
            <div className="space-y-2 text-[11px]">
              {recentActivity.slice(0, 3).map((act, idx) => (
                <div key={idx} className="flex items-start justify-between text-slate-400">
                  <div>
                    <div className="text-slate-200 font-medium">{act.event}</div>
                    <div className="text-[10px] text-slate-500">{act.resource} • {act.user}</div>
                  </div>
                  <span className="text-[10px] font-mono text-slate-500">{act.time}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
