'use client'

import React from 'react'
import Link from 'next/link'

export default function DashboardOverview() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-white tracking-tight">Platform Overview</h1>
        <p className="text-slate-400 mt-1">Real-time health, microservices status, and multi-tenant database infrastructure.</p>
      </div>

      {/* Metrics Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-2">
          <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Active Databases</div>
          <div className="text-3xl font-extrabold text-white">4 / 5</div>
          <div className="text-xs text-blue-400">80% quota utilized</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-2">
          <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Avg Latency</div>
          <div className="text-3xl font-extrabold text-white">1.4 ms</div>
          <div className="text-xs text-emerald-400">Optimal execution time</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-2">
          <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Storage Usage</div>
          <div className="text-3xl font-extrabold text-white">2.4 GB</div>
          <div className="text-xs text-slate-400">Of 10.0 GB quota</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-2">
          <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Audit Log Events</div>
          <div className="text-3xl font-extrabold text-white">1,248</div>
          <div className="text-xs text-emerald-400">Zero security anomalies</div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-4">
        <h2 className="text-lg font-bold text-white">Quick Actions</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Link
            href="/dashboard/databases"
            className="p-4 bg-slate-950 border border-slate-800 hover:border-blue-500/50 rounded-lg group transition"
          >
            <div className="font-semibold text-slate-100 group-hover:text-blue-400 transition">Provision Database</div>
            <div className="text-xs text-slate-400 mt-1">Spin up managed PostgreSQL or MySQL instances instantly.</div>
          </Link>

          <Link
            href="/dashboard/query"
            className="p-4 bg-slate-950 border border-slate-800 hover:border-blue-500/50 rounded-lg group transition"
          >
            <div className="font-semibold text-slate-100 group-hover:text-blue-400 transition">Execute SQL Query</div>
            <div className="text-xs text-slate-400 mt-1">Run queries with safety validation & timing metrics.</div>
          </Link>

          <Link
            href="/dashboard/backups"
            className="p-4 bg-slate-950 border border-slate-800 hover:border-blue-500/50 rounded-lg group transition"
          >
            <div className="font-semibold text-slate-100 group-hover:text-blue-400 transition">Create Snapshot</div>
            <div className="text-xs text-slate-400 mt-1">Backup database archives to object storage providers.</div>
          </Link>
        </div>
      </div>
    </div>
  )
}
