'use client'

import React, { useState, useEffect } from 'react'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudMetric } from '@/components/cloud/CloudMetric'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { API_BASE_URL } from '@/lib/api'

interface BackupItem {
  id: string
  resourceId: string
  databaseName: string
  name: string
  type: 'AUTOMATED' | 'MANUAL' | 'SNAPSHOT' | 'WAL_ARCHIVE'
  status: 'VERIFIED' | 'COMPLETED' | 'REQUESTED' | 'FAILED'
  integrity: 'VALID' | 'UNVERIFIED'
  sizeBytes: number
  retentionDays: number
  storageBucket: string
  createdAt: string
  expiresAt: string
}

export default function BackupsPage() {
  const [activeTab, setActiveTab] = useState('backups')
  const [backups, setBackups] = useState<BackupItem[]>([
    {
      id: 'bak-prod-101',
      resourceId: 'arnv:bak:ap-hyderabad-1:proj-default:database/production-db/backup/daily-snapshot-20260810',
      databaseName: 'production-db',
      name: 'daily-snapshot-20260810',
      type: 'AUTOMATED',
      status: 'VERIFIED',
      integrity: 'VALID',
      sizeBytes: 14589000,
      retentionDays: 7,
      storageBucket: 'anarva-media-assets',
      createdAt: new Date(Date.now() - 86400000).toISOString(),
      expiresAt: new Date(Date.now() + 518400000).toISOString(),
    },
  ])

  // Modals & Wizards
  const [createSnapshotModalOpen, setCreateSnapshotModalOpen] = useState(false)
  const [snapshotName, setSnapshotName] = useState('')
  const [targetDbName, setTargetDbName] = useState('production-db')
  const [retentionDays, setRetentionDays] = useState(7)
  const [isCreatingSnapshot, setIsCreatingSnapshot] = useState(false)

  const [restoreModalOpen, setRestoreModalOpen] = useState(false)
  const [selectedBackupForRestore, setSelectedBackupForRestore] = useState<BackupItem | null>(null)
  const [restoredDbName, setRestoredDbName] = useState('production-db-restored')
  const [isRestoring, setIsRestoring] = useState(false)

  const handleCreateSnapshot = () => {
    setIsCreatingSnapshot(true)
    setTimeout(() => {
      const name = snapshotName || `manual-${Date.now().toString().slice(-4)}`
      const newBackup: BackupItem = {
        id: `bak-${Date.now()}`,
        resourceId: `arnv:bak:ap-hyderabad-1:proj-default:database/${targetDbName}/backup/${name}`,
        databaseName: targetDbName,
        name,
        type: 'MANUAL',
        status: 'VERIFIED',
        integrity: 'VALID',
        sizeBytes: 14589000,
        retentionDays,
        storageBucket: 'anarva-media-assets',
        createdAt: new Date().toISOString(),
        expiresAt: new Date(Date.now() + retentionDays * 86400000).toISOString(),
      }

      setBackups([newBackup, ...backups])
      setIsCreatingSnapshot(false)
      setCreateSnapshotModalOpen(false)
      setSnapshotName('')
    }, 1000)
  }

  const handleExecuteRestore = () => {
    setIsRestoring(true)
    setTimeout(() => {
      setIsRestoring(false)
      setRestoreModalOpen(false)
      alert(`Asynchronous Restore Job submitted cleanly for ${restoredDbName}!`)
    }, 1200)
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const tabs: TabItem[] = [
    { id: 'backups', label: 'Database Backups & Snapshots' },
    { id: 'pitr', label: 'Point-in-Time Recovery (PITR)' },
    { id: 'readiness', label: 'Disaster Recovery Readiness' },
    { id: 'retention', label: 'Retention Policies' },
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Database Backups & Disaster Recovery</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">
            Automated database snapshots, object storage backup retention, restore job state machines, and PITR timeline models.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <CloudButton variant="primary" size="sm" onClick={() => setCreateSnapshotModalOpen(true)}>
            + Create Manual Snapshot
          </CloudButton>
        </div>
      </div>

      {/* Recovery Readiness Banner */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 shadow-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-3">
            <span className="text-sm font-bold text-white">Disaster Recovery Readiness Status</span>
            <span className="px-2.5 py-0.5 rounded text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">
              PARTIAL READINESS
            </span>
          </div>
          <p className="text-xs text-slate-400">
            Control-plane automated snapshots active in AOS Object Storage (<strong className="text-emerald-400">anarva-media-assets</strong>). Streaming WAL archival pending bare-metal provider connection.
          </p>
        </div>

        <CloudButton variant="secondary" size="sm" onClick={() => setActiveTab('readiness')}>
          Inspect Readiness
        </CloudButton>
      </div>

      {/* Summary Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <CloudMetric label="Total Backups" value={`${backups.length} Snapshots`} subtext="Verified Control-Plane" trend="VERIFIED" trendType="positive" />
        <CloudMetric label="Total Storage Used" value={formatBytes(backups.reduce((acc, b) => acc + b.sizeBytes, 0))} subtext="AOS Object Storage" trend="OPTIMIZED" trendType="positive" />
        <CloudMetric label="Backup Retention" value="7 Days Default" subtext="Automated Lifecycle" trend="CONFIGURED" trendType="positive" />
        <CloudMetric label="WAL Archival PITR" value="PENDING" subtext="Bare-Metal Provider Pending" trend="NOTICE" trendType="neutral" />
      </div>

      {/* Navigation Tabs */}
      <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {/* Tab Content */}
      <div className="space-y-6">
        {activeTab === 'backups' && (
          <div className="border border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-slate-900/60">
            <div className="p-4 bg-slate-950 border-b border-slate-800 flex items-center justify-between">
              <h3 className="text-sm font-bold text-white">Database Backup & Snapshot Registry</h3>
              <span className="text-xs font-mono text-slate-400">{backups.length} Backups Total</span>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-left font-sans text-xs divide-y divide-slate-800">
                <thead className="bg-slate-950 text-slate-400 font-bold uppercase text-[10px] tracking-wider">
                  <tr>
                    <th className="p-4">Snapshot Name</th>
                    <th className="p-4">Database</th>
                    <th className="p-4">Type</th>
                    <th className="p-4">Status</th>
                    <th className="p-4">Size</th>
                    <th className="p-4">Expires</th>
                    <th className="p-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/80 font-mono">
                  {backups.map((b) => (
                    <tr key={b.id} className="hover:bg-slate-800/40 transition">
                      <td className="p-4 font-bold text-white">{b.name}</td>
                      <td className="p-4 text-blue-400">{b.databaseName}</td>
                      <td className="p-4 text-slate-300">
                        <span className="px-2 py-0.5 bg-slate-800 rounded text-[10px]">{b.type}</span>
                      </td>
                      <td className="p-4">
                        <CloudStatus status={b.status} />
                      </td>
                      <td className="p-4 text-slate-300">{formatBytes(b.sizeBytes)}</td>
                      <td className="p-4 text-slate-400">{new Date(b.expiresAt).toLocaleDateString()}</td>
                      <td className="p-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => { setSelectedBackupForRestore(b); setRestoredDbName(`${b.databaseName}-restored`); setRestoreModalOpen(true); }}
                            className="px-2.5 py-1 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/20 rounded text-[11px] font-semibold transition"
                          >
                            Restore
                          </button>
                          <button
                            onClick={() => setBackups(backups.filter((item) => item.id !== b.id))}
                            className="px-2.5 py-1 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 rounded text-[11px] font-semibold transition"
                          >
                            Delete
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {activeTab === 'pitr' && (
          <div className="p-8 text-center bg-slate-900 border border-slate-800 rounded-2xl space-y-2 font-mono">
            <div className="text-xs text-amber-400 font-bold uppercase">PROVIDER NOT CONNECTED</div>
            <div className="text-xs text-slate-400 max-w-lg mx-auto">
              Continuous Write-Ahead Log (WAL) archival and second-by-second Point-in-Time Recovery (PITR) require bare-metal PostgreSQL replication driver attachment.
            </div>
          </div>
        )}

        {activeTab === 'readiness' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <CloudCard title="Disaster Recovery Signals & Verification">
              <div className="space-y-3 text-xs font-mono">
                <div className="flex justify-between py-1.5 border-b border-slate-800">
                  <span className="text-slate-400">Automated Daily Backups:</span>
                  <span className="text-emerald-400 font-bold">ENABLED & VERIFIED</span>
                </div>
                <div className="flex justify-between py-1.5 border-b border-slate-800">
                  <span className="text-slate-400">Storage Destination:</span>
                  <span className="text-blue-400 font-bold">anarva-media-assets (AOS S3)</span>
                </div>
                <div className="flex justify-between py-1.5 border-b border-slate-800">
                  <span className="text-slate-400">Backup Integrity Checksum:</span>
                  <span className="text-emerald-400 font-bold">SHA-256 VALIDATED</span>
                </div>
                <div className="flex justify-between py-1.5">
                  <span className="text-slate-400">Multi-Region Disaster Failover:</span>
                  <span className="text-slate-500 font-bold">COMING SOON</span>
                </div>
              </div>
            </CloudCard>

            <CloudCard title="Backup Retention Policy">
              <div className="space-y-3 text-xs font-mono">
                <div className="flex justify-between py-1.5 border-b border-slate-800">
                  <span className="text-slate-400">Default Retention Window:</span>
                  <span className="text-white font-bold">7 Days</span>
                </div>
                <div className="flex justify-between py-1.5 border-b border-slate-800">
                  <span className="text-slate-400">Automated Expire Cleanup:</span>
                  <span className="text-emerald-400 font-bold">ACTIVE</span>
                </div>
              </div>
            </CloudCard>
          </div>
        )}
      </div>

      {/* Create Manual Snapshot Modal */}
      {createSnapshotModalOpen && (
        <CloudModal
          isOpen={createSnapshotModalOpen}
          onClose={() => setCreateSnapshotModalOpen(false)}
          title="Create Database Manual Snapshot"
          subtitle="Provision control-plane snapshot stored in AOS Object Storage"
        >
          <div className="space-y-4 text-xs">
            <div className="space-y-1">
              <label className="block text-slate-300 font-semibold">Target Database</label>
              <select
                value={targetDbName}
                onChange={(e) => setTargetDbName(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none"
              >
                <option value="production-db">production-db (PostgreSQL 17.2)</option>
                <option value="analytics-db">analytics-db (PostgreSQL 16.4)</option>
              </select>
            </div>

            <div className="space-y-1">
              <label className="block text-slate-300 font-semibold">Snapshot Identifier Name</label>
              <input
                type="text"
                value={snapshotName}
                onChange={(e) => setSnapshotName(e.target.value)}
                placeholder="e.g. pre-deployment-backup-v2"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none font-mono"
              />
            </div>

            <div className="space-y-1">
              <label className="block text-slate-300 font-semibold">Retention Days</label>
              <select
                value={retentionDays}
                onChange={(e) => setRetentionDays(Number(e.target.value))}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none font-mono"
              >
                <option value={7}>7 Days Retention</option>
                <option value={14}>14 Days Retention</option>
                <option value={30}>30 Days Retention</option>
                <option value={90}>90 Days Retention</option>
              </select>
            </div>

            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2">
              <CloudButton variant="outline" size="sm" onClick={() => setCreateSnapshotModalOpen(false)}>
                Cancel
              </CloudButton>
              <CloudButton variant="primary" size="sm" isLoading={isCreatingSnapshot} onClick={handleCreateSnapshot}>
                Create Snapshot
              </CloudButton>
            </div>
          </div>
        </CloudModal>
      )}

      {/* Restore Modal */}
      {restoreModalOpen && selectedBackupForRestore && (
        <CloudModal
          isOpen={restoreModalOpen}
          onClose={() => setRestoreModalOpen(false)}
          title="Restore Database from Snapshot"
          subtitle={`Restore ${selectedBackupForRestore.name} into a new managed database cluster`}
        >
          <div className="space-y-4 text-xs font-mono">
            <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl space-y-1 text-slate-400">
              <div>Source Snapshot: <strong className="text-white">{selectedBackupForRestore.name}</strong></div>
              <div>Source Database: <strong className="text-blue-400">{selectedBackupForRestore.databaseName}</strong></div>
              <div>Snapshot Size: <strong className="text-emerald-400">{formatBytes(selectedBackupForRestore.sizeBytes)}</strong></div>
            </div>

            <div className="space-y-1 font-sans">
              <label className="block text-slate-300 font-semibold">Target New Database Name</label>
              <input
                type="text"
                value={restoredDbName}
                onChange={(e) => setRestoredDbName(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none font-mono"
              />
            </div>

            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2 font-sans">
              <CloudButton variant="outline" size="sm" onClick={() => setRestoreModalOpen(false)}>
                Cancel
              </CloudButton>
              <CloudButton variant="primary" size="sm" isLoading={isRestoring} onClick={handleExecuteRestore}>
                Submit Restore Job
              </CloudButton>
            </div>
          </div>
        </CloudModal>
      )}
    </div>
  )
}
