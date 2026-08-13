'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { API_BASE_URL } from '@/lib/api'

interface RegionItem {
  id: string
  name: string
  code: string
  provider: string
  status: string
  latitudeReference: number
  countryCode: string
  capacityStatus: string
  realityLabel: string
}

interface IncidentItem {
  id: string
  severity: string
  regionId: string
  type: string
  status: string
  summary: string
  startedAt: string
}

export default function GlobalInfrastructurePage() {
  const [activeTab, setActiveTab] = useState('regions')
  const [regions, setRegions] = useState<RegionItem[]>([
    {
      id: 'ap-hyderabad-1',
      name: 'Asia Pacific (Hyderabad)',
      code: 'ap-hyderabad-1',
      provider: 'LOCAL_SIMULATION',
      status: 'ACTIVE',
      latitudeReference: 17.385,
      countryCode: 'IN',
      capacityStatus: 'OPTIMAL',
      realityLabel: 'LOCAL_SIMULATION (LIMITED_CAPABILITIES)',
    },
    {
      id: 'us-east-1',
      name: 'US East (N. Virginia)',
      code: 'us-east-1',
      provider: 'LOCAL_SIMULATION',
      status: 'ACTIVE',
      latitudeReference: 38.907,
      countryCode: 'US',
      capacityStatus: 'OPTIMAL',
      realityLabel: 'LOCAL_SIMULATION (LIMITED_CAPABILITIES)',
    },
  ])

  const [incidents, setIncidents] = useState<IncidentItem[]>([])
  const [isSimulating, setIsSimulating] = useState(false)
  const [isFailoverOpen, setIsFailoverOpen] = useState(false)
  const [generationLock, setGenerationLock] = useState(1)
  const [failoverResult, setFailoverResult] = useState<any | null>(null)

  useEffect(() => {
    fetch(`${API_BASE_URL}/api/v1/regions`)
      .then((r) => r.json())
      .then((res) => {
        if (res && res.data) setRegions(res.data)
      })
      .catch(() => null)

    fetch(`${API_BASE_URL}/api/v1/incidents`)
      .then((r) => r.json())
      .then((res) => {
        if (res && res.data) setIncidents(res.data)
      })
      .catch(() => null)
  }, [])

  const handleSimulateOutage = async (regionId: string) => {
    setIsSimulating(true)
    const res = await fetch(`${API_BASE_URL}/api/v1/infrastructure/simulate-outage`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ regionId }),
    })
      .then((r) => r.json())
      .catch(() => null)

    if (res && res.data) {
      setIncidents([res.data, ...incidents])
    }
    setIsSimulating(false)
  }

  const handleExecuteFailover = async () => {
    setFailoverResult(null)
    const payload = {
      id: `pol-${Date.now()}`,
      resourceId: 'prod-database-cluster',
      primary: 'ap-hyderabad-1',
      secondary: 'us-east-1',
      healthThreshold: 3,
      mode: 'AUTOMATIC',
      generationLock,
    }

    const res = await fetch(`${API_BASE_URL}/api/v1/failover/execute`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    }).then((r) => r.json()).catch(() => null)

    setFailoverResult(res)
    setGenerationLock(generationLock + 1)
  }

  const tabItems: TabItem[] = [
    { id: 'regions', label: 'Regions & Zones' },
    { id: 'failover', label: 'HA & Failover Engine' },
    { id: 'incidents', label: 'Incident Log' },
    { id: 'simulator', label: 'Outage Simulator' },
  ]

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Global Infrastructure Control Plane</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Multi-Region placement, Availability Zones, Data Residency & Failover Engine.</p>
        </div>

        <div className="flex items-center gap-2">
          <span className="px-3 py-1 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-xs rounded font-mono font-bold">
            GLOBAL HEALTH: HEALTHY
          </span>
        </div>
      </div>

      <CloudTabs tabs={tabItems} activeTab={activeTab} onChange={setActiveTab} />

      {activeTab === 'regions' && (
        <CloudCard title="Global Active Regions Registry" subtitle="Provider-aware Regions and Availability Zones">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 font-mono text-xs">
            {regions.map((reg) => (
              <div key={reg.id} className="p-5 bg-slate-950 border border-slate-800 rounded-xl space-y-3">
                <div className="flex items-center justify-between">
                  <div className="font-bold text-white text-sm font-sans flex items-center gap-2">
                    🚩 {reg.name} ({reg.code})
                  </div>
                  <CloudStatus status={reg.status} />
                </div>

                <div className="space-y-1 text-slate-400">
                  <div>Country: <strong className="text-slate-200">{reg.countryCode}</strong></div>
                  <div>Capacity: <strong className="text-emerald-400">{reg.capacityStatus}</strong></div>
                  <div>Reality Label: <strong className="text-blue-400">{reg.realityLabel}</strong></div>
                </div>

                <div className="pt-2 border-t border-slate-800 flex justify-between items-center text-[10px]">
                  <span className="text-slate-500">Availability Zones: 2 ({reg.code}a, {reg.code}b)</span>
                  <button onClick={() => handleSimulateOutage(reg.id)} className="text-xs text-orange-400 hover:text-orange-300">
                    Simulate Outage
                  </button>
                </div>
              </div>
            ))}
          </div>
        </CloudCard>
      )}

      {activeTab === 'failover' && (
        <CloudCard title="High-Availability & Failover Control Engine" subtitle="Distributed lock generation & Split-brain protection">
          <div className="space-y-6 font-mono text-xs">
            <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
              <div className="font-bold text-white text-sm font-sans">Active Target: Primary DB Cluster (prod-database-cluster)</div>
              <div className="text-slate-400">Primary Region: <strong className="text-emerald-400">ap-hyderabad-1</strong></div>
              <div className="text-slate-400">Secondary Standby: <strong className="text-purple-400">us-east-1</strong></div>
              <div className="text-slate-400">Current Generation Lock: <strong className="text-blue-400">Gen #{generationLock}</strong></div>
            </div>

            <div className="flex gap-3">
              <CloudButton variant="primary" size="sm" onClick={handleExecuteFailover}>
                ⚡ Trigger Failover (Gen #{generationLock})
              </CloudButton>
            </div>

            {failoverResult && (
              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                {failoverResult.error ? (
                  <div className="text-red-400 font-bold">❌ {failoverResult.error}</div>
                ) : (
                  <div>
                    <div className="text-emerald-400 font-bold mb-2">✅ Failover Executed Successfully (Plan ID: {failoverResult.data.id})</div>
                    <ul className="space-y-1 text-slate-300 list-disc list-inside">
                      {failoverResult.data.steps.map((step: string, idx: number) => (
                        <li key={idx}>{step}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            )}
          </div>
        </CloudCard>
      )}

      {activeTab === 'incidents' && (
        <CloudCard title="Infrastructure Incident & Alert Log">
          {incidents.length === 0 ? (
            <CloudEmptyState
              title="No Active Infrastructure Incidents"
              description="All global regions, zones, networks, and databases are operating normally."
              actionLabel="View Regions"
              onAction={() => setActiveTab('regions')}
              icon="🛡️"
              docsLink="/console/developer"
            />
          ) : (
            <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
              {incidents.map((inc) => (
                <div key={inc.id} className="p-4 bg-slate-950 flex items-center justify-between">
                  <div>
                    <div className="font-bold text-orange-400 font-sans">{inc.summary}</div>
                    <div className="text-[10px] text-slate-500 mt-1">ID: {inc.id} • Region: {inc.regionId} • Type: {inc.type}</div>
                  </div>
                  <span className="px-2 py-0.5 bg-red-500/10 text-red-400 border border-red-500/20 rounded font-bold">
                    {inc.severity}
                  </span>
                </div>
              ))}
            </div>
          )}
        </CloudCard>
      )}

      {activeTab === 'simulator' && (
        <CloudCard title="Safe Development Outage Simulator" subtitle="LOCAL_SIMULATION mode only">
          <div className="space-y-4 font-mono text-xs">
            <p className="text-slate-400 font-sans">
              Test multi-region failover and incident alerts safely without impacting production infrastructure.
            </p>
            <div className="flex gap-3">
              <CloudButton variant="secondary" size="sm" onClick={() => handleSimulateOutage('ap-hyderabad-1')} disabled={isSimulating}>
                Simulate ap-hyderabad-1 Outage
              </CloudButton>
              <CloudButton variant="secondary" size="sm" onClick={() => handleSimulateOutage('us-east-1')} disabled={isSimulating}>
                Simulate us-east-1 Outage
              </CloudButton>
            </div>
          </div>
        </CloudCard>
      )}
    </div>
  )
}
