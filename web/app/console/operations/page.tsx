'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudModal } from '@/components/cloud/CloudModal'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { API_BASE_URL } from '@/lib/api'

interface TimelineEvent {
  stepNumber: number
  name: string
  description: string
  status: string
  timestamp: string
}

interface RecoveryInfo {
  attempted: boolean
  attempt: number
  status: string
  reason?: string
}

interface OperationItem {
  id: string
  organizationId: string
  projectId: string
  resourceId: string
  resourceType?: string
  type: string
  status: string
  progress: number
  createdAt: string
  startedAt?: string
  completedAt?: string
  updatedAt: string
  heartbeatAt?: string
  leaseExpiresAt?: string
  retryCount?: number
  errorCode?: string
  errorMessage?: string
  requestId: string
  actorId?: string
  recovery?: RecoveryInfo
  timeline?: TimelineEvent[]
}

interface SystemComponent {
  name: string
  key: string
  status: string
  description: string
}

interface SystemStatusData {
  status: string
  components: SystemComponent[]
  requestId: string
}

export default function OperationsConsolePage() {
  const [operations, setOperations] = useState<OperationItem[]>([])
  const [systemStatus, setSystemStatus] = useState<SystemStatusData | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  // Filtering state
  const [statusFilter, setStatusFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [resourceFilter, setResourceFilter] = useState('')

  // Detail Modal state
  const [selectedOp, setSelectedOp] = useState<OperationItem | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [timeline, setTimeline] = useState<TimelineEvent[]>([])

  useEffect(() => {
    fetchData()
  }, [statusFilter, typeFilter, resourceFilter])

  const fetchData = async () => {
    setIsLoading(true)
    try {
      // Build Operations URL with query params
      const params = new URLSearchParams()
      if (statusFilter) params.set('status', statusFilter)
      if (typeFilter) params.set('operationType', typeFilter)
      if (resourceFilter) params.set('resourceId', resourceFilter)

      const [opsRes, sysRes] = await Promise.all([
        fetch(`${API_BASE_URL}/api/v1/operations?${params.toString()}`).then((r) => r.json()),
        fetch(`${API_BASE_URL}/api/v1/system/status`).then((r) => r.json()),
      ])

      if (opsRes && opsRes.data) setOperations(opsRes.data)
      if (sysRes && sysRes.data) setSystemStatus(sysRes.data)
    } catch {
      // Ignore network errors in dev
    } finally {
      setIsLoading(false)
    }
  }

  const handleOpenDetail = async (op: OperationItem) => {
    setSelectedOp(op)
    setIsModalOpen(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/operations/${op.id}/timeline`).then((r) => r.json())
      if (res && res.data) {
        setTimeline(res.data)
      } else if (op.timeline) {
        setTimeline(op.timeline)
      } else {
        setTimeline([])
      }
    } catch {
      setTimeline(op.timeline || [])
    }
  }

  // Summary Metrics calculation
  const totalOps = operations.length
  const activeOps = operations.filter((o) => o.status === 'RUNNING' || o.status === 'PENDING' || o.status === 'QUEUED').length
  const succeededOps = operations.filter((o) => o.status === 'SUCCEEDED').length
  const failedOps = operations.filter((o) => o.status === 'FAILED').length
  const timedOutOps = operations.filter((o) => o.status === 'TIMED_OUT').length
  const recoveringOps = operations.filter((o) => o.status === 'RECOVERING' || (o.recovery && o.recovery.attempted)).length

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-gray-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Anarva Operations Center</h1>
          <p className="text-sm text-gray-400 mt-1">
            Real-time control plane operation lifecycle management, system status, timeout detection & failure recovery.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <CloudButton variant="secondary" onClick={fetchData}>
            Refresh Operations
          </CloudButton>
        </div>
      </div>

      {/* Anarva System Status Grid */}
      <CloudCard>
        <div className="flex justify-between items-center mb-4">
          <div>
            <h2 className="text-sm font-semibold text-gray-200 uppercase tracking-wider">Anarva Platform System Status</h2>
            <p className="text-xs text-gray-400">Live readiness status across Anarva Control Plane subsystems</p>
          </div>
          {systemStatus && (
            <CloudStatus status={systemStatus.status === 'READY' ? 'AVAILABLE' : systemStatus.status} />
          )}
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {systemStatus?.components.map((comp) => (
            <div key={comp.key} className="bg-gray-900/60 p-3 rounded border border-gray-800 flex flex-col justify-between">
              <div className="flex justify-between items-start">
                <span className="text-xs font-semibold text-white">{comp.name}</span>
                <span
                  className={`text-[10px] font-mono font-bold px-1.5 py-0.5 rounded ${
                    comp.status === 'READY'
                      ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                      : comp.status === 'DEGRADED'
                      ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                      : 'bg-red-500/10 text-red-400 border border-red-500/20'
                  }`}
                >
                  {comp.status}
                </span>
              </div>
              <span className="text-[11px] text-gray-400 mt-2 block">{comp.description}</span>
            </div>
          ))}
        </div>
      </CloudCard>

      {/* Operational Metrics Cards */}
      <div className="grid grid-cols-2 md:grid-cols-6 gap-3">
        <CloudCard>
          <span className="text-[11px] text-gray-400 uppercase font-semibold">Total</span>
          <div className="text-xl font-bold text-white mt-1">{totalOps}</div>
        </CloudCard>
        <CloudCard>
          <span className="text-[11px] text-gray-400 uppercase font-semibold">Active</span>
          <div className="text-xl font-bold text-cyan-400 mt-1">{activeOps}</div>
        </CloudCard>
        <CloudCard>
          <span className="text-[11px] text-gray-400 uppercase font-semibold">Succeeded</span>
          <div className="text-xl font-bold text-emerald-400 mt-1">{succeededOps}</div>
        </CloudCard>
        <CloudCard>
          <span className="text-[11px] text-gray-400 uppercase font-semibold">Failed</span>
          <div className="text-xl font-bold text-red-400 mt-1">{failedOps}</div>
        </CloudCard>
        <CloudCard>
          <span className="text-[11px] text-gray-400 uppercase font-semibold">Timed Out</span>
          <div className="text-xl font-bold text-amber-400 mt-1">{timedOutOps}</div>
        </CloudCard>
        <CloudCard>
          <span className="text-[11px] text-gray-400 uppercase font-semibold">Recovering</span>
          <div className="text-xl font-bold text-purple-400 mt-1">{recoveringOps}</div>
        </CloudCard>
      </div>

      {/* Filters Bar */}
      <div className="flex flex-wrap gap-4 bg-gray-900/40 p-4 border border-gray-800 rounded-lg">
        <div className="w-48">
          <label className="block text-[11px] font-medium text-gray-400 uppercase mb-1">Status</label>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="w-full bg-gray-900 border border-gray-700 rounded p-1.5 text-xs text-white focus:outline-none"
          >
            <option value="">All Statuses</option>
            <option value="RUNNING">RUNNING</option>
            <option value="SUCCEEDED">SUCCEEDED</option>
            <option value="FAILED">FAILED</option>
            <option value="TIMED_OUT">TIMED_OUT</option>
            <option value="RECOVERING">RECOVERING</option>
          </select>
        </div>
        <div className="w-48">
          <label className="block text-[11px] font-medium text-gray-400 uppercase mb-1">Operation Type</label>
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="w-full bg-gray-900 border border-gray-700 rounded p-1.5 text-xs text-white focus:outline-none"
          >
            <option value="">All Types</option>
            <option value="CREATE_VPC">CREATE_VPC</option>
            <option value="CREATE_DATABASE">CREATE_DATABASE</option>
            <option value="CREATE_COMPUTE">CREATE_COMPUTE</option>
            <option value="DELETE_VPC">DELETE_VPC</option>
          </select>
        </div>
        <div className="w-48">
          <label className="block text-[11px] font-medium text-gray-400 uppercase mb-1">Resource ID</label>
          <input
            type="text"
            value={resourceFilter}
            onChange={(e) => setResourceFilter(e.target.value)}
            placeholder="Filter by Resource..."
            className="w-full bg-gray-900 border border-gray-700 rounded p-1.5 text-xs text-white focus:outline-none"
          />
        </div>
      </div>

      {/* Operations Table */}
      {operations.length === 0 && !isLoading ? (
        <CloudEmptyState
          title="No Operations Found"
          description="Control-plane operations will appear here when infrastructure resources are provisioned or updated."
        />
      ) : (
        <div className="overflow-x-auto border border-gray-800 rounded-lg">
          <table className="w-full text-left text-sm text-gray-300">
            <thead className="bg-gray-900/60 text-gray-400 uppercase text-xs">
              <tr>
                <th className="px-4 py-3">Operation ID</th>
                <th className="px-4 py-3">Resource</th>
                <th className="px-4 py-3">Type</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Actor</th>
                <th className="px-4 py-3">Request ID</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800">
              {operations.map((op) => (
                <tr key={op.id} className="hover:bg-gray-800/40">
                  <td className="px-4 py-3 font-mono text-cyan-400 text-xs">{op.id}</td>
                  <td className="px-4 py-3 font-mono text-xs text-gray-300">{op.resourceId}</td>
                  <td className="px-4 py-3 text-xs font-semibold text-white">{op.type}</td>
                  <td className="px-4 py-3">
                    <CloudStatus status={op.status} />
                  </td>
                  <td className="px-4 py-3 text-xs text-gray-400">{op.actorId || 'SYSTEM'}</td>
                  <td className="px-4 py-3 font-mono text-xs text-gray-500">{op.requestId}</td>
                  <td className="px-4 py-3">
                    <CloudButton variant="secondary" onClick={() => handleOpenDetail(op)}>
                      Timeline & Detail
                    </CloudButton>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Operation Detail Modal */}
      {selectedOp && (
        <CloudModal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          title={`Operation ${selectedOp.id} Lifecycle`}
        >
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-3 text-xs bg-gray-900/60 p-3 rounded border border-gray-800">
              <div>
                <span className="text-gray-400 block uppercase">Resource ID</span>
                <span className="font-mono text-cyan-400 font-medium">{selectedOp.resourceId}</span>
              </div>
              <div>
                <span className="text-gray-400 block uppercase">Operation Type</span>
                <span className="font-semibold text-white">{selectedOp.type}</span>
              </div>
              <div>
                <span className="text-gray-400 block uppercase">Status</span>
                <CloudStatus status={selectedOp.status} />
              </div>
              <div>
                <span className="text-gray-400 block uppercase">Request ID Trace</span>
                <span className="font-mono text-gray-300">{selectedOp.requestId}</span>
              </div>
            </div>

            {selectedOp.errorMessage && (
              <div className="bg-red-500/10 border border-red-500/30 p-3 rounded text-xs text-red-400 font-mono">
                <span className="font-bold block uppercase mb-1">Failure Reason</span>
                {selectedOp.errorCode ? `[${selectedOp.errorCode}] ` : ''}{selectedOp.errorMessage}
              </div>
            )}

            {selectedOp.recovery && selectedOp.recovery.attempted && (
              <div className="bg-purple-500/10 border border-purple-500/30 p-3 rounded text-xs text-purple-300">
                <span className="font-bold block uppercase mb-1">Failure Recovery Info</span>
                <div>Status: <span className="font-semibold">{selectedOp.recovery.status}</span></div>
                <div>Attempt: <span className="font-semibold">#{selectedOp.recovery.attempt}</span></div>
                {selectedOp.recovery.reason && <div>Reason: {selectedOp.recovery.reason}</div>}
              </div>
            )}

            {/* Lifecycle Timeline */}
            <div>
              <h3 className="text-xs font-semibold text-gray-400 uppercase mb-2">Lifecycle Timeline Events</h3>
              {timeline.length === 0 ? (
                <p className="text-xs text-gray-500 italic">No timeline events recorded.</p>
              ) : (
                <div className="space-y-2">
                  {timeline.map((evt, idx) => (
                    <div key={idx} className="flex items-start gap-3 bg-gray-900/40 p-2.5 rounded border border-gray-800 text-xs">
                      <span className="bg-cyan-500/20 text-cyan-400 font-bold px-2 py-0.5 rounded text-[10px]">
                        #{evt.stepNumber || idx + 1}
                      </span>
                      <div className="flex-1">
                        <div className="flex justify-between items-center">
                          <span className="font-semibold text-white">{evt.name}</span>
                          <span className="text-[10px] text-gray-400 font-mono">
                            {new Date(evt.timestamp).toLocaleTimeString()}
                          </span>
                        </div>
                        <p className="text-gray-400 text-[11px] mt-0.5">{evt.description}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </CloudModal>
      )}
    </div>
  )
}
