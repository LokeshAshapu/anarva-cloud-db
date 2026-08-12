'use client'

import React, { useState, useEffect } from 'react'
import Link from 'next/link'
import { CloudMetric } from '@/components/cloud/CloudMetric'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { createClient } from '@/utils/supabase/client'

interface ActivityItem {
  id: string
  action: string
  resource: string
  actor: string
  time: string
}

interface ResourceItem {
  id: string
  name: string
  resourceId: string
  type: 'DATABASE' | 'STORAGE_BUCKET' | 'COMPUTE' | 'NETWORK'
  status: string
}

export default function CloudConsoleHome() {
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [userResources, setUserResources] = useState<ResourceItem[]>([])
  const [activities, setActivities] = useState<ActivityItem[]>([])
  const [isLoaded, setIsLoaded] = useState(false)

  useEffect(() => {
    async function loadDashboard() {
      let email = 'user@anarva.io'
      if (typeof window !== 'undefined') {
        const storedEmail = localStorage.getItem('anarva_user_email')
        if (storedEmail) email = storedEmail

        try {
          const supabase = createClient()
          const { data } = await supabase.auth.getUser()
          if (data?.user?.email) {
            email = data.user.email
            localStorage.setItem('anarva_user_email', email)
          }
        } catch (e) {
          console.log('Supabase user check:', e)
        }
      }
      setUserEmail(email)

      // Load user-specific resources
      const dbKey = `anarva_user_databases_${email}`
      const bucketKey = `anarva_user_buckets_${email}`
      const computeKey = `anarva_user_compute_${email}`
      const actKey = `anarva_user_activities_${email}`

      let dbs: any[] = []
      let buckets: any[] = []
      let compute: any[] = []
      let acts: any[] = []

      const storedDbs = localStorage.getItem(dbKey)
      const storedBuckets = localStorage.getItem(bucketKey)
      const storedCompute = localStorage.getItem(computeKey)
      const storedActs = localStorage.getItem(actKey)

      if (storedDbs) {
        dbs = JSON.parse(storedDbs)
      } else if (email === 'lokeshashapu@gmail.com') {
        // Initial seed only for primary admin
        dbs = [
          { id: 'res-db-prod-1', name: 'production-db', resourceId: 'arnv:db:ap-hyderabad-1:proj-default:database/production-db', type: 'DATABASE', status: 'AVAILABLE' },
          { id: 'res-db-analytics-1', name: 'analytics-db', resourceId: 'arnv:db:ap-mumbai-1:proj-default:database/analytics-db', type: 'DATABASE', status: 'AVAILABLE' },
        ]
        localStorage.setItem(dbKey, JSON.stringify(dbs))
      }

      if (storedBuckets) {
        buckets = JSON.parse(storedBuckets)
      } else if (email === 'lokeshashapu@gmail.com') {
        buckets = [
          { id: 'res-s3-assets-1', name: 'anarva-media-assets', resourceId: 'arnv:s3:ap-hyderabad-1:proj-default:storage/anarva-media-assets', type: 'STORAGE_BUCKET', status: 'AVAILABLE' },
        ]
        localStorage.setItem(bucketKey, JSON.stringify(buckets))
      }

      if (storedCompute) {
        compute = JSON.parse(storedCompute)
      } else if (email === 'lokeshashapu@gmail.com') {
        compute = [
          { id: 'acu-1', name: 'ace-worker-node-01', resourceId: 'arnv:vm:ap-hyderabad-1:proj-default:compute/ace-worker-node-01', type: 'COMPUTE', status: 'AVAILABLE' },
        ]
        localStorage.setItem(computeKey, JSON.stringify(compute))
      }

      if (storedActs) {
        acts = JSON.parse(storedActs)
      } else if (email === 'lokeshashapu@gmail.com') {
        acts = [
          { id: 'act-1', action: 'RESOURCE_CREATED', resource: 'production-db', actor: email, time: '10 mins ago' },
          { id: 'act-2', action: 'RESOURCE_CREATED', resource: 'anarva-media-assets', actor: email, time: '45 mins ago' },
          { id: 'act-3', action: 'RESOURCE_STARTED', resource: 'ace-worker-node-01', actor: email, time: '2 hours ago' },
        ]
        localStorage.setItem(actKey, JSON.stringify(acts))
      }

      const combined: ResourceItem[] = [
        ...dbs.map((d: any) => ({ id: d.id, name: d.name, resourceId: d.resourceId || `arnv:db:ap-hyderabad-1:proj-default:database/${d.name}`, type: 'DATABASE' as const, status: d.status || 'AVAILABLE' })),
        ...buckets.map((b: any) => ({ id: b.id, name: b.name, resourceId: b.resourceId || `arnv:s3:ap-hyderabad-1:proj-default:storage/${b.name}`, type: 'STORAGE_BUCKET' as const, status: b.status || 'AVAILABLE' })),
        ...compute.map((c: any) => ({ id: c.id, name: c.name, resourceId: c.resourceId || `arnv:vm:ap-hyderabad-1:proj-default:compute/${c.name}`, type: 'COMPUTE' as const, status: c.status || 'AVAILABLE' })),
      ]

      setUserResources(combined)
      setActivities(acts)
      setIsLoaded(true)
    }

    loadDashboard()
  }, [])

  const dbCount = userResources.filter((r) => r.type === 'DATABASE').length
  const bucketCount = userResources.filter((r) => r.type === 'STORAGE_BUCKET').length
  const computeCount = userResources.filter((r) => r.type === 'COMPUTE').length

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono text-slate-400">ORGANIZATION / ACCOUNT:</span>
            <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-xs font-mono font-bold">
              {userEmail}
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
        <CloudMetric label="Total Resources" value={userResources.length} subtext="Account Isolated" trend="ACTIVE" trendType="positive" />
        <CloudMetric label="Databases" value={dbCount} subtext="PostgreSQL / MySQL" trend="100% UP" trendType="positive" />
        <CloudMetric label="Storage Buckets" value={bucketCount} subtext="Anarva AOS S3" trend="NORMAL" trendType="positive" />
        <CloudMetric label="Compute ACUs" value={computeCount} subtext="0.5 - 128 ACUs" trend="ACTIVE" trendType="positive" />
        <CloudMetric label="Virtual Networks" value={1} subtext="10.0.0.0/16 VPC" trend="ISOLATED" trendType="positive" />
        <CloudMetric label="Backups & PITR" value={dbCount > 0 ? 1 : 0} subtext="Point-in-Time Active" trend="PROTECTED" trendType="positive" />
      </div>

      {/* Grid Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Resources */}
        <div className="lg:col-span-2 space-y-4">
          <CloudCard title="Managed Cloud Resource Registry" subtitle={`Active resources under account '${userEmail}'`}>
            {userResources.length === 0 ? (
              <CloudEmptyState
                title="No cloud resources created yet"
                description="Your account currently has 0 active database clusters or storage buckets. Use the action buttons above to provision your first resource."
                actionLabel="+ Deploy Managed Database"
                onAction={() => (window.location.href = '/console/databases')}
              />
            ) : (
              <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
                {userResources.map((res) => (
                  <div key={res.id} className="p-4 bg-slate-950 flex items-center justify-between">
                    <div>
                      <div className="font-bold text-white flex items-center gap-2">
                        {res.name}
                        <span className="text-[10px] px-1.5 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded font-normal">
                          {res.type}
                        </span>
                      </div>
                      <div className="text-[10px] text-slate-500 mt-0.5">{res.resourceId}</div>
                    </div>
                    <CloudStatus status={res.status} />
                  </div>
                ))}
              </div>
            )}
          </CloudCard>
        </div>

        {/* Audit Activity Stream */}
        <div className="space-y-4">
          <CloudCard title="Recent Activity Stream" subtitle={`Activity event log for ${userEmail}`}>
            {activities.length === 0 ? (
              <div className="p-6 text-center text-xs text-slate-500 font-mono">No recent activity events recorded.</div>
            ) : (
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
            )}
          </CloudCard>
        </div>
      </div>
    </div>
  )
}
