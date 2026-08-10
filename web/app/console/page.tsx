'use client'

import React, { useState, useEffect } from 'react'
import Link from 'next/link'
import { CloudMetric } from '@/components/cloud/CloudMetric'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { API_BASE_URL } from '@/lib/api'

export default function CloudConsoleHome() {
  const [resourceCount, setResourceCount] = useState({
    total: 5,
    databases: 2,
    storage: 1,
    compute: 1,
    network: 1,
    backups: 2,
  })

  const [activities, setActivities] = useState([
    { id: 'act-1', action: 'RESOURCE_CREATED', resource: 'production-db', actor: 'lokeshashapu@gmail.com', time: '10 mins ago' },
    { id: 'act-2', action: 'RESOURCE_CREATED', resource: 'anarva-media-assets', actor: 'lokeshashapu@gmail.com', time: '45 mins ago' },
    { id: 'act-3', action: 'RESOURCE_STARTED', resource: 'ace-worker-node-01', actor: 'lokeshashapu@gmail.com', time: '2 hours ago' },
  ])

  useEffect(() => {
    // Fetch live backend resource statistics if API gateway is online
    async function fetchStats() {
      try {
        const res = await fetch(`${API_BASE_URL}/api/v1/resources`).catch(() => null)
        if (res && res.ok) {
          const list = await res.json()
          if (Array.isArray(list)) {
            const dbs = list.filter((r) => r.type === 'DATABASE').length
            const s3s = list.filter((r) => r.type === 'STORAGE_BUCKET').length
            const vms = list.filter((r) => r.type === 'COMPUTE').length
            const vpcs = list.filter((r) => r.type === 'NETWORK').length
            setResourceCount({
              total: list.length,
              databases: dbs,
              storage: s3s,
              compute: vms,
              network: vpcs,
              backups: 2,
            })
          }
        }
      } catch (e) {
        console.log('Backend stats fetch notice:', e)
      }
    }
    fetchStats()
  }, [])

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono text-slate-400">ORGANIZATION:</span>
            <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-xs font-mono font-bold">
              org-default (Anarva Systems)
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight mt-1">Cloud Infrastructure Overview</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-0.5">
            Real-time infrastructure health, metered telemetry, and resource hierarchy monitoring.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Link href="/console/databases">
            <CloudButton variant="primary" size="sm">
              + Deploy Database
            </CloudButton>
          </Link>
          <Link href="/console/storage">
            <CloudButton variant="secondary" size="sm">
              + Create Bucket
            </CloudButton>
          </Link>
        </div>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
        <CloudMetric label="Total Resources" value={resourceCount.total} subtext="Across 5 Regions" trend="ACTIVE" trendType="positive" />
        <CloudMetric label="Databases" value={resourceCount.databases} subtext="PostgreSQL / MySQL" trend="100% UP" trendType="positive" />
        <CloudMetric label="Storage Buckets" value={resourceCount.storage} subtext="Anarva AOS S3" trend="NORMAL" trendType="positive" />
        <CloudMetric label="Compute ACUs" value={resourceCount.compute} subtext="0.5 - 128 ACUs" trend="ACTIVE" trendType="positive" />
        <CloudMetric label="Virtual Networks" value={resourceCount.network} subtext="10.0.0.0/16 VPC" trend="ISOLATED" trendType="positive" />
        <CloudMetric label="Backups & PITR" value={resourceCount.backups} subtext="Point-in-Time Active" trend="PROTECTED" trendType="positive" />
      </div>

      {/* Grid Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Resources */}
        <div className="lg:col-span-2 space-y-4">
          <CloudCard title="Managed Cloud Resource Registry" subtitle="Active resources under project 'Anarva Cloud Platform'">
            <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
              <div className="p-4 bg-slate-950 flex items-center justify-between font-mono">
                <div>
                  <div className="font-bold text-white">production-db</div>
                  <div className="text-[10px] text-slate-500">arnv:db:ap-hyderabad-1:proj-default:database/production-db</div>
                </div>
                <CloudStatus status="AVAILABLE" />
              </div>

              <div className="p-4 bg-slate-950 flex items-center justify-between font-mono">
                <div>
                  <div className="font-bold text-white">analytics-db</div>
                  <div className="text-[10px] text-slate-500">arnv:db:ap-mumbai-1:proj-default:database/analytics-db</div>
                </div>
                <CloudStatus status="AVAILABLE" />
              </div>

              <div className="p-4 bg-slate-950 flex items-center justify-between font-mono">
                <div>
                  <div className="font-bold text-white">anarva-media-assets</div>
                  <div className="text-[10px] text-slate-500">arnv:s3:ap-hyderabad-1:proj-default:storage/anarva-media-assets</div>
                </div>
                <CloudStatus status="AVAILABLE" />
              </div>

              <div className="p-4 bg-slate-950 flex items-center justify-between font-mono">
                <div>
                  <div className="font-bold text-white">ace-worker-node-01</div>
                  <div className="text-[10px] text-slate-500">arnv:vm:ap-hyderabad-1:proj-default:compute/ace-worker-node-01</div>
                </div>
                <CloudStatus status="AVAILABLE" />
              </div>
            </div>
          </CloudCard>
        </div>

        {/* Audit Activity Stream */}
        <div className="space-y-4">
          <CloudCard title="Recent Activity Stream" subtitle="Organization event log">
            <div className="space-y-3 text-xs">
              {activities.map((act) => (
                <div key={act.id} className="p-3 bg-slate-950 border border-slate-800 rounded-xl space-y-1">
                  <div className="flex items-center justify-between">
                    <span className="font-mono font-bold text-blue-400 text-[10px]">{act.action}</span>
                    <span className="text-[10px] text-slate-500">{act.time}</span>
                  </div>
                  <div className="font-bold text-white">{act.resource}</div>
                  <div className="text-[10px] text-slate-400 font-mono">Actor: {act.actor}</div>
                </div>
              ))}
            </div>
          </CloudCard>
        </div>
      </div>
    </div>
  )
}
