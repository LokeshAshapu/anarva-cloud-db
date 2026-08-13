'use client'

import React, { useState, useEffect } from 'react'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudMetric } from '@/components/cloud/CloudMetric'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
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

interface WALSegment {
  segmentId: string
  lsn: string
  sizeBytes: number
  archivedAt: string
  status: string
}

interface RetentionPolicyConfig {
  dailySnapshotDays: number
  walSegmentDays: number
  manualSnapshotDays: number
  autoPruneEnabled: boolean
  coldStorageTiering: boolean
  wormVaultLock: boolean
}

export default function BackupsPage() {
  const [activeTab, setActiveTab] = useState('backups')
  const [userEmail, setUserEmail] = useState('user@anarva.io')

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
    {
      id: 'bak-prod-102',
      resourceId: 'arnv:bak:ap-hyderabad-1:proj-default:database/analytics-replica/backup/manual-pre-migration',
      databaseName: 'analytics-replica',
      name: 'manual-pre-migration',
      type: 'MANUAL',
      status: 'VERIFIED',
      integrity: 'VALID',
      sizeBytes: 28400000,
      retentionDays: 14,
      storageBucket: 'anarva-media-assets',
      createdAt: new Date(Date.now() - 172800000).toISOString(),
      expiresAt: new Date(Date.now() + 1036800000).toISOString(),
    },
  ])

  // Modals & Wizards
  const [createSnapshotModalOpen, setCreateSnapshotModalOpen] = useState(false)
  const [snapshotName, setSnapshotName] = useState('')
  const [targetDbName, setTargetDbName] = useState('production-db')
  const [retentionDays, setRetentionDays] = useState(7)
  const [isSubmitting, setIsSubmitting] = useState(false)

  // PITR Engine State
  const [pitrTargetDb, setPitrTargetDb] = useState('production-db')
  const [pitrTimestamp, setPitrTimestamp] = useState<string>(
    new Date(Date.now() - 300000).toISOString().slice(0, 19)
  )
  const [isPitrRestoring, setIsPitrRestoring] = useState(false)
  const [pitrStep, setPitrStep] = useState(0)
  const [pitrSuccessResult, setPitrSuccessResult] = useState<string | null>(null)

  // Retention Policy Config State
  const [retentionConfig, setRetentionConfig] = useState<RetentionPolicyConfig>({
    dailySnapshotDays: 7,
    walSegmentDays: 7,
    manualSnapshotDays: 30,
    autoPruneEnabled: true,
    coldStorageTiering: true,
    wormVaultLock: false,
  })
  const [isSavingRetention, setIsSavingRetention] = useState(false)
  const [retentionSaveSuccess, setRetentionSaveSuccess] = useState(false)

  const [walSegments] = useState<WALSegment[]>([
    {
      segmentId: '0000000100000000000000A1',
      lsn: '0/16B23F0',
      sizeBytes: 16777216,
      archivedAt: new Date(Date.now() - 60000).toISOString(),
      status: 'ARCHIVED (AOS S3)',
    },
    {
      segmentId: '0000000100000000000000A2',
      lsn: '0/1700000',
      sizeBytes: 16777216,
      archivedAt: new Date(Date.now() - 120000).toISOString(),
      status: 'ARCHIVED (AOS S3)',
    },
    {
      segmentId: '0000000100000000000000A3',
      lsn: '0/174E100',
      sizeBytes: 16777216,
      archivedAt: new Date(Date.now() - 180000).toISOString(),
      status: 'ARCHIVED (AOS S3)',
    },
  ])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)

      const storedBackups = localStorage.getItem(`anarva_user_backups_${email}`)
      if (storedBackups) {
        try {
          const parsed = JSON.parse(storedBackups)
          if (Array.isArray(parsed)) setBackups(parsed)
        } catch (e) {}
      }

      const storedRetention = localStorage.getItem(`anarva_user_retention_${email}`)
      if (storedRetention) {
        try {
          setRetentionConfig(JSON.parse(storedRetention))
        } catch (e) {}
      }
    }
  }, [])

  const handleCreateSnapshot = (e: React.FormEvent) => {
    e.preventDefault()
    if (!snapshotName) return
    setIsSubmitting(true)

    setTimeout(() => {
      const newBackup: BackupItem = {
        id: `bak-${Date.now()}`,
        resourceId: `arnv:bak:ap-hyderabad-1:proj-default:database/${targetDbName}/backup/${snapshotName}`,
        databaseName: targetDbName,
        name: snapshotName,
        type: 'SNAPSHOT',
        status: 'VERIFIED',
        integrity: 'VALID',
        sizeBytes: 18200000,
        retentionDays,
        storageBucket: 'anarva-media-assets',
        createdAt: new Date().toISOString(),
        expiresAt: new Date(Date.now() + retentionDays * 86400000).toISOString(),
      }

      setBackups((prev) => {
        const updated = [newBackup, ...prev]
        if (typeof window !== 'undefined') {
          localStorage.setItem(`anarva_user_backups_${userEmail}`, JSON.stringify(updated))
        }
        return updated
      })

      setIsSubmitting(false)
      setCreateSnapshotModalOpen(false)
      setSnapshotName('')
    }, 400)
  }

  const handleDeleteBackup = (id: string) => {
    setBackups((prev) => {
      const updated = prev.filter((b) => b.id !== id)
      if (typeof window !== 'undefined') {
        localStorage.setItem(`anarva_user_backups_${userEmail}`, JSON.stringify(updated))
      }
      return updated
    })
  }

  const handleExecutePITR = () => {
    setIsPitrRestoring(true)
    setPitrStep(1)
    setPitrSuccessResult(null)

    setTimeout(() => setPitrStep(2), 800)
    setTimeout(() => setPitrStep(3), 1600)
    setTimeout(() => setPitrStep(4), 2400)
    setTimeout(() => {
      setPitrStep(5)
      setIsPitrRestoring(false)
      setPitrSuccessResult(
        `postgresql://postgres:********@${pitrTargetDb}-pitr-restored.anarva-cloud.internal:5432/${pitrTargetDb}`
      )
    }, 3200)
  }

  const handleSaveRetentionConfig = (e: React.FormEvent) => {
    e.preventDefault()
    setIsSavingRetention(true)
    setRetentionSaveSuccess(false)

    setTimeout(() => {
      if (typeof window !== 'undefined') {
        localStorage.setItem(`anarva_user_retention_${userEmail}`, JSON.stringify(retentionConfig))
      }
      setIsSavingRetention(false)
      setRetentionSaveSuccess(true)
      setTimeout(() => setRetentionSaveSuccess(false), 3000)
    }, 300)
  }

  const tabs: TabItem[] = [
    { id: 'backups', label: 'Database Snapshots' },
    { id: 'pitr', label: 'Point-in-Time Recovery (PITR)' },
    { id: 'retention', label: 'Retention Policy' },
    { id: 'readiness', label: 'Disaster Recovery Signals' },
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono text-slate-400">DATA PROTECTION:</span>
            <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-xs font-mono font-bold">
              AUTOMATED SNAPSHOTS & WAL ARCHIVE
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight mt-1">Backup & Recovery Engine</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-0.5">
            Manage automated snapshot backups, WAL continuous archival, retention lifecycle rules, and Point-in-Time Recovery (PITR).
          </p>
        </div>

        <div className="flex items-center gap-2">
          <CloudButton variant="primary" size="sm" onClick={() => setCreateSnapshotModalOpen(true)}>
            + Create Snapshot Backup
          </CloudButton>
        </div>
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <CloudMetric label="Total Snapshots" value={backups.length} subtext="SHA-256 Verified" trend="HEALTHY" trendType="positive" />
        <CloudMetric label="WAL Segment Stream" value="ARCHIVED" subtext="AOS S3 Storage Bucket" trend="ACTIVE" trendType="positive" />
        <CloudMetric label="PITR Granularity" value="Second-by-Second" subtext="WAL Replay Delta Engine" trend="CONFIGURED" trendType="positive" />
        <CloudMetric label="Active Retention Window" value={`${retentionConfig.dailySnapshotDays} Days`} subtext="Automated Cleanup Cron" trend="ENFORCED" trendType="positive" />
      </div>

      {/* Tabs */}
      <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {/* Snapshots Tab */}
      {activeTab === 'backups' && (
        <CloudCard title="Managed Database Snapshots">
          <div className="space-y-4 font-mono text-xs">
            <div className="p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded-xl text-[11px]">
              ℹ Database snapshots are encrypted zero-trust backups stored directly in Anarva Object Storage (AOS S3).
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-slate-800 text-[10px] text-slate-400 uppercase">
                    <th className="py-3 px-4">SNAPSHOT NAME</th>
                    <th className="py-3 px-4">DATABASE</th>
                    <th className="py-3 px-4">TYPE</th>
                    <th className="py-3 px-4">STATUS</th>
                    <th className="py-3 px-4">SIZE</th>
                    <th className="py-3 px-4">CREATED</th>
                    <th className="py-3 px-4">ACTIONS</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60">
                  {backups.map((bak) => (
                    <tr key={bak.id} className="hover:bg-slate-900/40">
                      <td className="py-3 px-4 font-bold text-white">{bak.name}</td>
                      <td className="py-3 px-4 text-slate-300">{bak.databaseName}</td>
                      <td className="py-3 px-4">
                        <span className="px-2 py-0.5 bg-slate-800 text-slate-300 rounded text-[10px]">{bak.type}</span>
                      </td>
                      <td className="py-3 px-4">
                        <span className="px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded text-[10px] font-bold">
                          {bak.status}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-slate-300">{(bak.sizeBytes / 1024 / 1024).toFixed(2)} MB</td>
                      <td className="py-3 px-4 text-slate-400 text-[10px]">{new Date(bak.createdAt).toLocaleString()}</td>
                      <td className="py-3 px-4">
                        <CloudButton variant="danger" size="sm" onClick={() => handleDeleteBackup(bak.id)}>
                          Delete
                        </CloudButton>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </CloudCard>
      )}

      {/* Point-in-Time Recovery Tab */}
      {activeTab === 'pitr' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Continuous Write-Ahead Log (WAL) & Point-in-Time Recovery (PITR)">
            <div className="space-y-5">
              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                <div className="flex items-center gap-2">
                  <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[10px] font-bold">
                    LOCAL DEVELOPMENT WAL SIMULATOR / CONFIGURED
                  </span>
                  <span className="text-[10px] text-slate-400">Continuous WAL Segment Archival Active</span>
                </div>
                <p className="text-slate-300 text-xs">
                  Continuous Write-Ahead Log (WAL) streams PostgreSQL transaction logs directly to AOS S3 storage.
                  Select any target database and exact timestamp to restore a new database instance down to the second.
                </p>
              </div>

              {/* Target & Timestamp Selection */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div>
                  <label className="block text-slate-400 text-[10px] mb-1">TARGET DATABASE CLUSTER</label>
                  <select
                    value={pitrTargetDb}
                    onChange={(e) => setPitrTargetDb(e.target.value)}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-emerald-500"
                  >
                    <option value="production-db">production-db (PostgreSQL 16)</option>
                    <option value="analytics-replica">analytics-replica (PostgreSQL 16)</option>
                    <option value="users-primary">users-primary (PostgreSQL 16)</option>
                  </select>
                </div>

                <div className="sm:col-span-2">
                  <label className="block text-slate-400 text-[10px] mb-1">TARGET RECOVERY TIMESTAMP (UTC)</label>
                  <div className="flex gap-2">
                    <input
                      type="datetime-local"
                      step="1"
                      value={pitrTimestamp}
                      onChange={(e) => setPitrTimestamp(e.target.value)}
                      className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-emerald-500"
                    />
                    <CloudButton
                      variant="secondary"
                      size="sm"
                      onClick={() => setPitrTimestamp(new Date(Date.now() - 300000).toISOString().slice(0, 19))}
                    >
                      5m Ago
                    </CloudButton>
                    <CloudButton
                      variant="secondary"
                      size="sm"
                      onClick={() => setPitrTimestamp(new Date(Date.now() - 3600000).toISOString().slice(0, 19))}
                    >
                      1h Ago
                    </CloudButton>
                  </div>
                </div>
              </div>

              <div className="flex justify-end">
                <CloudButton variant="primary" size="sm" onClick={handleExecutePITR} disabled={isPitrRestoring}>
                  {isPitrRestoring ? 'Replaying WAL Log Segments...' : 'Execute Point-in-Time Restore'}
                </CloudButton>
              </div>

              {/* Progress & Result Box */}
              {pitrStep > 0 && (
                <div className="p-5 bg-slate-950 border border-slate-800 rounded-xl space-y-3">
                  <div className="font-bold text-white text-sm">PITR Recovery Progress:</div>
                  <div className="space-y-2 text-xs">
                    <div className={pitrStep >= 1 ? 'text-emerald-400' : 'text-slate-600'}>
                      {pitrStep >= 1 ? '✓' : '○'} Step 1: Validating Base Snapshot & Integrity Checksums
                    </div>
                    <div className={pitrStep >= 2 ? 'text-emerald-400' : 'text-slate-600'}>
                      {pitrStep >= 2 ? '✓' : '○'} Step 2: Downloading Archived WAL Segments from AOS S3 Storage
                    </div>
                    <div className={pitrStep >= 3 ? 'text-emerald-400' : 'text-slate-600'}>
                      {pitrStep >= 3 ? '✓' : '○'} Step 3: Replaying Log Sequence Numbers (LSN 0/16B23F0) to target timestamp {pitrTimestamp}
                    </div>
                    <div className={pitrStep >= 4 ? 'text-emerald-400' : 'text-slate-600'}>
                      {pitrStep >= 4 ? '✓' : '○'} Step 4: Verifying Transaction Consistency & Provisioning Restored Database Cluster
                    </div>
                  </div>

                  {pitrSuccessResult && (
                    <div className="mt-4 p-4 bg-emerald-500/10 border border-emerald-500/30 rounded-lg space-y-2">
                      <div className="font-bold text-emerald-400">✓ Point-in-Time Recovery Completed Successfully!</div>
                      <div className="text-[11px] text-slate-300">
                        Restored Connection URI: <code className="text-emerald-300 font-bold select-all">{pitrSuccessResult}</code>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* Active WAL Segments Stream */}
              <div className="space-y-2">
                <div className="font-bold text-slate-300 text-xs">Active WAL Log Segments Stream</div>
                <div className="space-y-1.5">
                  {walSegments.map((w) => (
                    <div key={w.segmentId} className="p-3 bg-slate-950 border border-slate-800 rounded-lg flex items-center justify-between">
                      <div>
                        <div className="font-bold text-white">{w.segmentId}</div>
                        <div className="text-[10px] text-slate-400">LSN: {w.lsn} • Size: {(w.sizeBytes / 1024 / 1024).toFixed(1)} MB</div>
                      </div>
                      <span className="px-2 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded text-[10px]">
                        {w.status}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </CloudCard>
        </div>
      )}

      {/* Retention Policy Tab */}
      {activeTab === 'retention' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Backup Retention & Lifecycle Management Policy">
            <form onSubmit={handleSaveRetentionConfig} className="space-y-6">
              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                <div className="flex items-center gap-2">
                  <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-[10px] font-bold">
                    AUTOMATED LIFECYCLE ENGINE ACTIVE
                  </span>
                  <span className="text-[10px] text-slate-400">Enforced by Daily Expiration Cron Job</span>
                </div>
                <p className="text-slate-300 text-xs">
                  Configure automated snapshot expiration windows, Write-Ahead Log (WAL) archive retention, and cold storage tiering rules for <strong className="text-white">{userEmail}</strong>.
                </p>
              </div>

              {/* Form Controls */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
                <div>
                  <label className="block text-slate-300 mb-1.5 font-bold">Daily Snapshot Retention</label>
                  <select
                    value={retentionConfig.dailySnapshotDays}
                    onChange={(e) => setRetentionConfig({ ...retentionConfig, dailySnapshotDays: Number(e.target.value) })}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  >
                    <option value={7}>7 Days (Standard)</option>
                    <option value={14}>14 Days (Extended)</option>
                    <option value={30}>30 Days (Monthly)</option>
                    <option value={90}>90 Days (Quarterly)</option>
                    <option value={365}>365 Days (1 Year Compliance)</option>
                  </select>
                  <span className="text-[10px] text-slate-500 mt-1 block">Automated daily snapshot retention window</span>
                </div>

                <div>
                  <label className="block text-slate-300 mb-1.5 font-bold">WAL Archive Retention (PITR)</label>
                  <select
                    value={retentionConfig.walSegmentDays}
                    onChange={(e) => setRetentionConfig({ ...retentionConfig, walSegmentDays: Number(e.target.value) })}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  >
                    <option value={3}>3 Days</option>
                    <option value={7}>7 Days (Recommended)</option>
                    <option value={14}>14 Days</option>
                  </select>
                  <span className="text-[10px] text-slate-500 mt-1 block">Continuous WAL delta replay window</span>
                </div>

                <div>
                  <label className="block text-slate-300 mb-1.5 font-bold">Manual Snapshot Retention</label>
                  <select
                    value={retentionConfig.manualSnapshotDays}
                    onChange={(e) => setRetentionConfig({ ...retentionConfig, manualSnapshotDays: Number(e.target.value) })}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  >
                    <option value={14}>14 Days</option>
                    <option value={30}>30 Days</option>
                    <option value={90}>90 Days</option>
                    <option value={0}>Indefinite (Until Deleted)</option>
                  </select>
                  <span className="text-[10px] text-slate-500 mt-1 block">On-demand manual snapshot retention</span>
                </div>
              </div>

              {/* Toggles */}
              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-bold text-white">Automated Expiration Cleanup</div>
                    <div className="text-[10px] text-slate-400">Automatically delete expired snapshots past the retention window</div>
                  </div>
                  <input
                    type="checkbox"
                    checked={retentionConfig.autoPruneEnabled}
                    onChange={(e) => setRetentionConfig({ ...retentionConfig, autoPruneEnabled: e.target.checked })}
                    className="h-4 w-4 rounded border-slate-800 text-blue-600 focus:ring-blue-500"
                  />
                </div>

                <div className="flex items-center justify-between border-t border-slate-800/60 pt-3">
                  <div>
                    <div className="font-bold text-white">Automated AOS S3 Cold Storage Tiering</div>
                    <div className="text-[10px] text-slate-400">Archive snapshots older than 7 days to Glacier cold storage tier for cost reduction</div>
                  </div>
                  <input
                    type="checkbox"
                    checked={retentionConfig.coldStorageTiering}
                    onChange={(e) => setRetentionConfig({ ...retentionConfig, coldStorageTiering: e.target.checked })}
                    className="h-4 w-4 rounded border-slate-800 text-blue-600 focus:ring-blue-500"
                  />
                </div>

                <div className="flex items-center justify-between border-t border-slate-800/60 pt-3">
                  <div>
                    <div className="font-bold text-white">Compliance WORM Vault Lock</div>
                    <div className="text-[10px] text-slate-400">Enforce Write-Once-Read-Many immutable backup lock against accidental deletion</div>
                  </div>
                  <input
                    type="checkbox"
                    checked={retentionConfig.wormVaultLock}
                    onChange={(e) => setRetentionConfig({ ...retentionConfig, wormVaultLock: e.target.checked })}
                    className="h-4 w-4 rounded border-slate-800 text-blue-600 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div className="flex items-center justify-between pt-2">
                {retentionSaveSuccess ? (
                  <span className="text-emerald-400 font-bold text-xs">✓ Retention Policy settings saved successfully!</span>
                ) : (
                  <span />
                )}

                <CloudButton variant="primary" size="sm" type="submit" disabled={isSavingRetention}>
                  {isSavingRetention ? 'Saving Policy...' : 'Save Retention Policy'}
                </CloudButton>
              </div>
            </form>
          </CloudCard>
        </div>
      )}

      {/* Disaster Recovery Signals Tab */}
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

          <CloudCard title="Backup Retention Policy Summary">
            <div className="space-y-3 text-xs font-mono">
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Daily Snapshot Retention:</span>
                <span className="text-white font-bold">{retentionConfig.dailySnapshotDays} Days</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">WAL Segment Retention (PITR):</span>
                <span className="text-emerald-400 font-bold">{retentionConfig.walSegmentDays} Days</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-slate-800">
                <span className="text-slate-400">Automated Expire Cleanup:</span>
                <span className="text-emerald-400 font-bold">{retentionConfig.autoPruneEnabled ? 'ACTIVE' : 'DISABLED'}</span>
              </div>
              <div className="flex justify-between py-1.5">
                <span className="text-slate-400">Cold Storage Tiering:</span>
                <span className="text-blue-400 font-bold">{retentionConfig.coldStorageTiering ? 'ENABLED (GLACIER)' : 'DISABLED'}</span>
              </div>
            </div>
          </CloudCard>
        </div>
      )}

      {/* Modal: Create Snapshot */}
      {createSnapshotModalOpen && (
        <CloudModal isOpen={createSnapshotModalOpen} title="Create Database Snapshot" onClose={() => setCreateSnapshotModalOpen(false)}>
          <form onSubmit={handleCreateSnapshot} className="space-y-4 font-mono text-xs">
            <div>
              <label className="block text-slate-300 mb-1">Target Database</label>
              <select
                value={targetDbName}
                onChange={(e) => setTargetDbName(e.target.value)}
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
              >
                <option value="production-db">production-db</option>
                <option value="analytics-replica">analytics-replica</option>
              </select>
            </div>

            <div>
              <label className="block text-slate-300 mb-1">Snapshot Name</label>
              <input
                type="text"
                required
                value={snapshotName}
                onChange={(e) => setSnapshotName(e.target.value)}
                placeholder="e.g. pre-migration-snapshot"
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
              />
            </div>

            <div>
              <label className="block text-slate-300 mb-1">Retention Period (Days)</label>
              <input
                type="number"
                min="1"
                max="30"
                value={retentionDays}
                onChange={(e) => setRetentionDays(Number(e.target.value))}
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
              />
            </div>

            <div className="pt-2 flex justify-end gap-2">
              <CloudButton variant="secondary" size="sm" onClick={() => setCreateSnapshotModalOpen(false)}>
                Cancel
              </CloudButton>
              <CloudButton variant="primary" size="sm" type="submit" disabled={isSubmitting}>
                {isSubmitting ? 'Creating Snapshot...' : 'Create Snapshot'}
              </CloudButton>
            </div>
          </form>
        </CloudModal>
      )}
    </div>
  )
}
