'use client'

import React, { useState } from 'react'
import { CloudResource, ResourceStatus, RegionId, EnvironmentType, ResourceType } from '@/types/resource'
import { CloudStatus } from './CloudStatus'
import { CloudButton } from './CloudButton'
import { CloudModal } from './CloudModal'

export interface CloudResourceListProps {
  resources: CloudResource[]
  title: string
  onCreateResource?: () => void
  onDeleteResource?: (id: string) => void
  onViewResource?: (resource: CloudResource) => void
}

export function CloudResourceList({
  resources,
  title,
  onCreateResource,
  onDeleteResource,
  onViewResource,
}: CloudResourceListProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState<string>('ALL')
  const [selectedRegion, setSelectedRegion] = useState<string>('ALL')
  const [selectedEnv, setSelectedEnv] = useState<string>('ALL')
  const [deleteTarget, setDeleteTarget] = useState<CloudResource | null>(null)
  const [deleteConfirmInput, setDeleteConfirmInput] = useState('')

  const filteredResources = resources.filter((res) => {
    if (selectedStatus !== 'ALL' && res.status !== selectedStatus) return false
    if (selectedRegion !== 'ALL' && res.regionId !== selectedRegion) return false
    if (selectedEnv !== 'ALL' && res.environment !== selectedEnv) return false
    if (searchQuery.trim() !== '') {
      const q = searchQuery.toLowerCase()
      const matchName = res.name.toLowerCase().includes(q)
      const matchArn = res.resourceId.toLowerCase().includes(q)
      const matchId = res.id.toLowerCase().includes(q)
      const matchTag = res.tags?.some((t) => `${t.key}:${t.value}`.toLowerCase().includes(q))
      if (!matchName && !matchArn && !matchId && !matchTag) return false
    }
    return true
  })

  const handleDeleteConfirm = () => {
    if (deleteTarget && deleteConfirmInput === deleteTarget.name) {
      if (onDeleteResource) onDeleteResource(deleteTarget.id)
      setDeleteTarget(null)
      setDeleteConfirmInput('')
    }
  }

  return (
    <div className="space-y-6">
      {/* Header & Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold text-white tracking-tight">{title}</h2>
          <p className="text-xs text-slate-400 mt-0.5">
            Showing {filteredResources.length} of {resources.length} active resources
          </p>
        </div>
        {onCreateResource && (
          <CloudButton variant="primary" size="sm" onClick={onCreateResource}>
            + Create Resource
          </CloudButton>
        )}
      </div>

      {/* Filter Bar */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-4 flex flex-col md:flex-row items-center justify-between gap-3 text-xs">
        {/* Search */}
        <div className="relative w-full md:w-72">
          <svg className="w-4 h-4 text-slate-400 absolute left-3 top-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search name, ARNV, tags..."
            className="w-full bg-slate-950 border border-slate-800 rounded-xl pl-9 pr-3 py-2 text-white focus:outline-none focus:border-blue-500 font-sans"
          />
        </div>

        {/* Dropdowns */}
        <div className="flex flex-wrap items-center gap-2 w-full md:w-auto">
          <select
            value={selectedStatus}
            onChange={(e) => setSelectedStatus(e.target.value)}
            className="bg-slate-950 border border-slate-800 text-slate-300 rounded-xl px-3 py-2 text-xs focus:outline-none cursor-pointer"
          >
            <option value="ALL">All Statuses</option>
            <option value="AVAILABLE">AVAILABLE</option>
            <option value="CREATING">CREATING</option>
            <option value="DEGRADED">DEGRADED</option>
            <option value="STOPPED">STOPPED</option>
          </select>

          <select
            value={selectedRegion}
            onChange={(e) => setSelectedRegion(e.target.value)}
            className="bg-slate-950 border border-slate-800 text-slate-300 rounded-xl px-3 py-2 text-xs focus:outline-none cursor-pointer"
          >
            <option value="ALL">All Regions</option>
            <option value="ap-hyderabad-1">Hyderabad (ap-hyderabad-1)</option>
            <option value="ap-mumbai-1">Mumbai (ap-mumbai-1)</option>
            <option value="ap-singapore-1">Singapore (ap-singapore-1)</option>
            <option value="us-east-1">N. Virginia (us-east-1)</option>
            <option value="eu-west-1">Frankfurt (eu-west-1)</option>
          </select>

          <select
            value={selectedEnv}
            onChange={(e) => setSelectedEnv(e.target.value)}
            className="bg-slate-950 border border-slate-800 text-slate-300 rounded-xl px-3 py-2 text-xs focus:outline-none cursor-pointer"
          >
            <option value="ALL">All Environments</option>
            <option value="Production">Production</option>
            <option value="Staging">Staging</option>
            <option value="Development">Development</option>
          </select>
        </div>
      </div>

      {/* Resource Table */}
      {filteredResources.length === 0 ? (
        <div className="p-8 text-center bg-slate-900 border border-slate-800 rounded-2xl text-slate-400 text-xs space-y-1">
          <div className="font-bold text-white">No matching resources found</div>
          <div>Try adjusting your filters or search terms.</div>
        </div>
      ) : (
        <div className="border border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-slate-900/60">
          <div className="overflow-x-auto">
            <table className="w-full text-left font-sans text-xs divide-y divide-slate-800">
              <thead className="bg-slate-950 text-slate-400 font-bold uppercase text-[10px] tracking-wider">
                <tr>
                  <th className="p-4">Name & ARNV Identifier</th>
                  <th className="p-4">Status</th>
                  <th className="p-4">Region</th>
                  <th className="p-4">Type</th>
                  <th className="p-4">Environment</th>
                  <th className="p-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/80 font-mono">
                {filteredResources.map((res) => (
                  <tr key={res.id} className="hover:bg-slate-800/40 transition">
                    <td className="p-4">
                      <button
                        onClick={() => onViewResource && onViewResource(res)}
                        className="font-bold text-white hover:text-blue-400 text-left transition"
                      >
                        {res.name}
                      </button>
                      <div className="text-[10px] text-slate-500 font-mono truncate max-w-xs sm:max-w-sm mt-0.5">
                        {res.resourceId}
                      </div>
                    </td>
                    <td className="p-4">
                      <CloudStatus status={res.status} />
                    </td>
                    <td className="p-4 text-slate-300">{res.regionId}</td>
                    <td className="p-4">
                      <span className="px-2 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded text-[10px] font-bold">
                        {res.type}
                      </span>
                    </td>
                    <td className="p-4 text-slate-400">{res.environment}</td>
                    <td className="p-4 text-right">
                      <div className="flex items-center justify-end gap-2">
                        {onViewResource && (
                          <button
                            onClick={() => onViewResource(res)}
                            className="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg text-[11px] font-semibold transition"
                          >
                            View
                          </button>
                        )}
                        {onDeleteResource && (
                          <button
                            onClick={() => setDeleteTarget(res)}
                            className="px-2.5 py-1 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 rounded-lg text-[11px] font-semibold transition"
                          >
                            Delete
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Safe Resource Deletion Confirmation Modal */}
      {deleteTarget && (
        <CloudModal
          isOpen={!!deleteTarget}
          onClose={() => setDeleteTarget(null)}
          title="Safe Resource Deletion"
          subtitle={`Confirmation required for ${deleteTarget.name}`}
        >
          <div className="space-y-4 text-xs">
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 space-y-1">
              <div className="font-bold">⚠️ Destructive Operation</div>
              <div>
                You are preparing to delete resource <strong className="text-white font-mono">{deleteTarget.name}</strong> ({deleteTarget.type}) in region <strong className="text-white font-mono">{deleteTarget.regionId}</strong>.
              </div>
            </div>

            <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl font-mono text-[11px] text-slate-300">
              ARNV: {deleteTarget.resourceId}
            </div>

            <div className="space-y-1">
              <label className="block text-slate-300 font-semibold">
                To confirm deletion, type <span className="font-mono text-red-400 font-bold">{deleteTarget.name}</span> below:
              </label>
              <input
                type="text"
                value={deleteConfirmInput}
                onChange={(e) => setDeleteConfirmInput(e.target.value)}
                placeholder={deleteTarget.name}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono focus:outline-none focus:border-red-500"
              />
            </div>

            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2">
              <CloudButton variant="outline" size="sm" onClick={() => setDeleteTarget(null)}>
                Cancel
              </CloudButton>
              <CloudButton
                variant="danger"
                size="sm"
                disabled={deleteConfirmInput !== deleteTarget.name}
                onClick={handleDeleteConfirm}
              >
                Confirm Delete
              </CloudButton>
            </div>
          </div>
        </CloudModal>
      )}
    </div>
  )
}
