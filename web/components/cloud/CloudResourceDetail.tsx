'use client'

import React, { useState } from 'react'
import { CloudResource } from '@/types/resource'
import { CloudStatus } from './CloudStatus'
import { CloudTabs, TabItem } from './CloudTabs'
import { CloudCard } from './CloudCard'
import { CloudButton } from './CloudButton'

export interface CloudResourceDetailProps {
  resource: CloudResource
  onBack?: () => void
}

export function CloudResourceDetail({ resource, onBack }: CloudResourceDetailProps) {
  const [activeTab, setActiveTab] = useState('overview')

  const tabs: TabItem[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'metrics', label: 'Metrics' },
    { id: 'configuration', label: 'Configuration' },
    { id: 'activity', label: 'Activity Log' },
    { id: 'security', label: 'Security' },
    { id: 'networking', label: 'Networking' },
    { id: 'backups', label: 'Backups' },
    { id: 'settings', label: 'Settings' },
  ]

  return (
    <div className="space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div className="space-y-1">
          {onBack && (
            <button
              onClick={onBack}
              className="text-xs text-blue-400 hover:underline font-mono flex items-center gap-1 mb-2"
            >
              ← Back to Resources
            </button>
          )}
          <div className="flex items-center gap-3">
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">{resource.name}</h1>
            <CloudStatus status={resource.status} />
          </div>
          <div className="text-xs text-slate-400 font-mono flex items-center gap-2">
            <span>ARNV:</span>
            <span className="text-emerald-400 font-bold bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
              {resource.resourceId}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <CloudButton variant="secondary" size="sm">
            Restart Instance
          </CloudButton>
          <CloudButton variant="outline" size="sm">
            Edit Tags
          </CloudButton>
        </div>
      </div>

      {/* Tabs */}
      <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {/* Tab Content */}
      <div className="space-y-6">
        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <CloudCard title="Resource Hierarchy & Scope">
              <div className="space-y-3 text-xs font-mono">
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Organization:</span>
                  <span className="text-white font-bold">{resource.organizationId} (Anarva Systems)</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Project:</span>
                  <span className="text-white font-bold">{resource.projectId} (Anarva Cloud Platform)</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Environment:</span>
                  <span className="text-blue-400 font-bold">{resource.environment}</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Region:</span>
                  <span className="text-emerald-400 font-bold">{resource.regionId}</span>
                </div>
                <div className="flex justify-between py-1">
                  <span className="text-slate-400">Owner:</span>
                  <span className="text-white font-bold">{resource.ownerId}</span>
                </div>
              </div>
            </CloudCard>

            <CloudCard title="Resource Metadata & Tags">
              <div className="space-y-3 text-xs">
                <div className="text-slate-400">Associated Environment Tags:</div>
                <div className="flex flex-wrap gap-2">
                  {resource.tags && resource.tags.length > 0 ? (
                    resource.tags.map((t, idx) => (
                      <span key={idx} className="px-2.5 py-1 bg-slate-950 border border-slate-800 rounded font-mono text-slate-300">
                        {t.key}: <strong className="text-blue-400">{t.value}</strong>
                      </span>
                    ))
                  ) : (
                    <span className="text-slate-500 italic">No tags assigned.</span>
                  )}
                </div>
                <div className="pt-4 border-t border-slate-800 font-mono text-[11px] text-slate-500">
                  Created At: {new Date(resource.createdAt).toLocaleString()}
                </div>
              </div>
            </CloudCard>
          </div>
        )}

        {activeTab === 'metrics' && (
          <div className="p-8 text-center bg-slate-900 border border-slate-800 rounded-2xl text-slate-400 text-xs space-y-2">
            <div className="font-bold text-white">Live Time-Series Observability Metrics</div>
            <div>CPU, Memory, and IOPS streaming pipeline active.</div>
          </div>
        )}

        {activeTab !== 'overview' && activeTab !== 'metrics' && (
          <div className="p-8 text-center bg-slate-900 border border-slate-800 rounded-2xl text-slate-400 text-xs space-y-2">
            <div className="font-bold text-white capitalize">{activeTab} Module</div>
            <div>Configuration parameters and controls for {resource.name}.</div>
          </div>
        )}
      </div>
    </div>
  )
}
