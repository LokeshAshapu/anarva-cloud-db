'use client'

import React, { useState, useEffect } from 'react'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudButton } from '@/components/cloud/CloudButton'
import { API_BASE_URL } from '@/lib/api'

interface ActivityItem {
  id: string
  action: string
  resource: string
  actor: string
  time: string
  status: 'SUCCESS' | 'WARN' | 'INFO'
  requestId?: string
}

export default function AuditLogsPage() {
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [searchQuery, setSearchQuery] = useState('')
  const [actionFilter, setActionFilter] = useState('ALL')

  const [activities, setActivities] = useState<ActivityItem[]>([
    {
      id: 'act-101',
      action: 'NETWORK_DELETED',
      resource: 'Primary Production VPC',
      actor: 'lokeshashapu@gmail.com',
      time: 'Just now',
      status: 'INFO',
      requestId: 'req-net-99',
    },
    {
      id: 'act-102',
      action: 'COMPUTE_DELETED',
      resource: 'ace-worker-node-01',
      actor: 'lokeshashapu@gmail.com',
      time: 'Just now',
      status: 'INFO',
      requestId: 'req-cmp-88',
    },
    {
      id: 'act-103',
      action: 'RESOURCE_DELETED',
      resource: 'anarva-media-assets',
      actor: 'lokeshashapu@gmail.com',
      time: 'Just now',
      status: 'INFO',
      requestId: 'req-s3-77',
    },
    {
      id: 'act-104',
      action: 'RESOURCE_DELETED',
      resource: 'analytics-db',
      actor: 'lokeshashapu@gmail.com',
      time: 'Just now',
      status: 'INFO',
      requestId: 'req-db-66',
    },
    {
      id: 'act-105',
      action: 'RESOURCE_DELETED',
      resource: 'production-db',
      actor: 'lokeshashapu@gmail.com',
      time: 'Just now',
      status: 'INFO',
      requestId: 'req-db-55',
    },
    {
      id: 'act-106',
      action: 'RESOURCE_CREATED',
      resource: 'production-db',
      actor: 'lokeshashapu@gmail.com',
      time: '10 mins ago',
      status: 'SUCCESS',
      requestId: 'req-db-44',
    },
    {
      id: 'act-107',
      action: 'RESOURCE_CREATED',
      resource: 'anarva-media-assets',
      actor: 'lokeshashapu@gmail.com',
      time: '45 mins ago',
      status: 'SUCCESS',
      requestId: 'req-s3-33',
    },
    {
      id: 'act-108',
      action: 'RESOURCE_STARTED',
      resource: 'ace-worker-node-01',
      actor: 'lokeshashapu@gmail.com',
      time: '2 hours ago',
      status: 'SUCCESS',
      requestId: 'req-cmp-22',
    },
  ])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)
    }

    async function loadActivities() {
      try {
        const res = await fetch(`${API_BASE_URL}/api/v1/monitoring/overview`).catch(() => null)
        if (res && res.ok) {
          const body = await res.json()
          if (body.activityStream && Array.isArray(body.activityStream) && body.activityStream.length > 0) {
            setActivities(body.activityStream)
          }
        }
      } catch (e) {}
    }

    loadActivities()
  }, [])

  const filteredActivities = activities.filter((act) => {
    if (actionFilter !== 'ALL' && !act.action.includes(actionFilter)) return false
    if (searchQuery && !act.resource.toLowerCase().includes(searchQuery.toLowerCase()) && !act.action.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false
    }
    return true
  })

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono text-slate-400">SECURITY AUDIT TRAIL:</span>
            <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-xs font-mono font-bold">
              APPEND-ONLY AUDIT LOG
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight mt-1">Audit Logs & Activity History</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-0.5">
            Immutable, append-only security event history and resource mutation logs for account <strong className="text-white">{userEmail}</strong>.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <CloudButton
            variant="secondary"
            size="sm"
            onClick={() => {
              const csv = filteredActivities.map((a) => `${a.time},${a.action},${a.resource},${a.actor}`).join('\n')
              const blob = new Blob([csv], { type: 'text/csv' })
              const url = URL.createObjectURL(blob)
              const a = document.createElement('a')
              a.href = url
              a.download = `audit-logs-${userEmail.split('@')[0]}.csv`
              a.click()
            }}
          >
            Export Audit CSV
          </CloudButton>
        </div>
      </div>

      {/* Main Content */}
      <CloudCard title={`Activity Event History (${filteredActivities.length} Events)`}>
        <div className="space-y-4 font-mono text-xs">
          {/* Search & Filter Bar */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <input
              type="text"
              placeholder="Search resource or action..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="sm:col-span-2 px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
            />

            <select
              value={actionFilter}
              onChange={(e) => setActionFilter(e.target.value)}
              className="px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500 cursor-pointer"
            >
              <option value="ALL">All Event Types</option>
              <option value="CREATED">CREATED Events</option>
              <option value="DELETED">DELETED Events</option>
              <option value="STARTED">STARTED Events</option>
            </select>
          </div>

          {/* Activity Event Stream Table */}
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden">
            {filteredActivities.map((act) => (
              <div key={act.id} className="p-4 bg-slate-950 hover:bg-slate-900/50 transition flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span
                      className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                        act.action.includes('DELETED')
                          ? 'bg-red-500/10 text-red-400 border border-red-500/20'
                          : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                      }`}
                    >
                      {act.action}
                    </span>
                    <span className="font-bold text-white text-sm">{act.resource}</span>
                  </div>
                  <div className="text-[10px] text-slate-400">
                    Actor: <strong className="text-slate-200">{act.actor}</strong> • Time: {act.time}
                  </div>
                </div>

                <div className="text-right">
                  <span className="text-[10px] text-slate-500">ReqID: {act.requestId || 'req-trace'}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </CloudCard>
    </div>
  )
}
