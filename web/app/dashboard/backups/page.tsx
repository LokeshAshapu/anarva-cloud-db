'use client'

import React, { useEffect, useState } from 'react'

const BACKUPS_STORAGE_KEY = 'anarva_user_backups'

export default function BackupsPage() {
  const [backups, setBackups] = useState<any[]>([])
  const [showTriggerModal, setShowTriggerModal] = useState(false)
  const [snapshotName, setSnapshotName] = useState('')
  const [targetDb, setTargetDb] = useState('Primary Application Database')

  const [restoreModalBackup, setRestoreModalBackup] = useState<any | null>(null)
  const [restoring, setRestoring] = useState(false)
  const [restoreSuccess, setRestoreSuccess] = useState('')

  const defaultBackups = [
    {
      id: 'snap-uuid-101',
      name: 'Automated Daily Snapshot',
      db: 'Primary Application Database',
      type: 'SNAPSHOT',
      status: 'COMPLETED',
      size: '2.4 GB',
      created_at: '2026-08-06 18:00:00',
    },
    {
      id: 'snap-uuid-102',
      name: 'Pre-Deployment Safety Backup',
      db: 'Analytics Data Warehouse',
      type: 'FULL_PITR',
      status: 'COMPLETED',
      size: '5.1 GB',
      created_at: '2026-08-07 12:30:00',
    },
  ]

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(BACKUPS_STORAGE_KEY)
      if (stored) {
        try {
          setBackups(JSON.parse(stored))
        } catch {
          setBackups(defaultBackups)
          localStorage.setItem(BACKUPS_STORAGE_KEY, JSON.stringify(defaultBackups))
        }
      } else {
        setBackups(defaultBackups)
        localStorage.setItem(BACKUPS_STORAGE_KEY, JSON.stringify(defaultBackups))
      }
    }
  }, [])

  const updateBackupsState = (newBackups: any[]) => {
    setBackups(newBackups)
    if (typeof window !== 'undefined') {
      localStorage.setItem(BACKUPS_STORAGE_KEY, JSON.stringify(newBackups))
    }
  }

  // Trigger snapshot backup
  const handleTriggerBackup = (e: React.FormEvent) => {
    e.preventDefault()
    const newSnap = {
      id: `snap-uuid-${Date.now()}`,
      name: snapshotName || 'Manual On-Demand Snapshot',
      db: targetDb,
      type: 'SNAPSHOT',
      status: 'COMPLETED',
      size: `${(1 + Math.random() * 4).toFixed(1)} GB`,
      created_at: new Date().toISOString().replace('T', ' ').substring(0, 19),
    }

    updateBackupsState([newSnap, ...backups])
    setSnapshotName('')
    setShowTriggerModal(false)
  }

  // Delete Backup Snapshot
  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to permanently delete this snapshot backup from cloud object storage?')) {
      const updated = backups.filter((b) => b.id !== id)
      updateBackupsState(updated)
    }
  }

  // Restore Point-In-Time Backup
  const handleConfirmRestore = () => {
    if (!restoreModalBackup) return
    setRestoring(true)
    setRestoreSuccess('')

    setTimeout(() => {
      setRestoring(false)
      setRestoreSuccess(`✔ Database '${restoreModalBackup.db}' successfully restored to point-in-time snapshot state (${restoreModalBackup.created_at})!`)
      setTimeout(() => {
        setRestoreModalBackup(null)
        setRestoreSuccess('')
      }, 2500)
    }, 1200)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Backups & Point-In-Time Restore</h1>
          <p className="text-slate-400 mt-1">Managed automated snapshot dumps and object storage archives.</p>
        </div>

        <button
          onClick={() => setShowTriggerModal(true)}
          className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25"
        >
          + Trigger Snapshot Backup
        </button>
      </div>

      {backups.length === 0 ? (
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-12 text-center space-y-4">
          <div className="inline-flex items-center justify-center h-16 w-16 rounded-2xl bg-blue-600/10 text-blue-400 text-3xl font-bold border border-blue-500/20">
            📦
          </div>
          <h3 className="text-xl font-bold text-white">No Snapshot Backups Created</h3>
          <p className="text-slate-400 text-sm max-w-md mx-auto">
            Click "+ Trigger Snapshot Backup" above to create an automated point-in-time backup snapshot of your cloud database!
          </p>
        </div>
      ) : (
        <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-xl">
          <table className="w-full text-left text-xs font-mono">
            <thead className="bg-slate-950 text-slate-400 border-b border-slate-800">
              <tr>
                <th className="p-4">Snapshot Name</th>
                <th className="p-4">Database</th>
                <th className="p-4">Type</th>
                <th className="p-4">Size</th>
                <th className="p-4">Status</th>
                <th className="p-4">Created At</th>
                <th className="p-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800 text-slate-200">
              {backups.map((b) => (
                <tr key={b.id} className="hover:bg-slate-800/40 transition">
                  <td className="p-4 font-semibold text-white">{b.name}</td>
                  <td className="p-4 text-blue-400">{b.db}</td>
                  <td className="p-4">
                    <span className="px-2.5 py-0.5 bg-slate-800 text-slate-300 rounded-full font-semibold">
                      {b.type}
                    </span>
                  </td>
                  <td className="p-4">{b.size}</td>
                  <td className="p-4">
                    <span className="px-2.5 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full font-semibold">
                      {b.status}
                    </span>
                  </td>
                  <td className="p-4 text-slate-400">{b.created_at}</td>
                  <td className="p-4 text-right space-x-2">
                    <button
                      onClick={() => setRestoreModalBackup(b)}
                      className="px-3 py-1.5 bg-blue-600/20 text-blue-400 hover:bg-blue-600/40 border border-blue-500/20 rounded-lg transition font-semibold"
                    >
                      Restore
                    </button>
                    <button
                      onClick={() => handleDelete(b.id)}
                      className="px-3 py-1.5 bg-red-500/20 text-red-400 hover:bg-red-500/40 border border-red-500/20 rounded-lg transition font-semibold"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Trigger Snapshot Backup Modal */}
      {showTriggerModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Trigger Snapshot Backup</h2>

            <form onSubmit={handleTriggerBackup} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Snapshot Label</label>
                <input
                  type="text"
                  required
                  value={snapshotName}
                  onChange={(e) => setSnapshotName(e.target.value)}
                  placeholder="e.g. Pre-Migration Backup"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Target Database</label>
                <select
                  value={targetDb}
                  onChange={(e) => setTargetDb(e.target.value)}
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                >
                  <option value="Primary Application Database">Primary Application Database</option>
                  <option value="Analytics Data Warehouse">Analytics Data Warehouse</option>
                  <option value="Production Aurora Cluster">Production Aurora Cluster</option>
                </select>
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowTriggerModal(false)}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25"
                >
                  Start Backup
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Restore Confirmation Modal */}
      {restoreModalBackup && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Restore Point-In-Time Backup</h2>
            <p className="text-xs text-slate-400">
              You are about to restore <span className="text-white font-semibold">{restoreModalBackup.db}</span> to snapshot state from <span className="text-blue-400 font-mono">{restoreModalBackup.created_at}</span>.
            </p>

            {restoreSuccess ? (
              <div className="p-3 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold rounded-lg">
                {restoreSuccess}
              </div>
            ) : (
              <div className="p-3 bg-amber-500/10 border border-amber-500/20 text-amber-400 text-xs rounded-lg">
                ⚠️ Caution: This will replace current database tables with the point-in-time snapshot state.
              </div>
            )}

            {!restoreSuccess && (
              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setRestoreModalBackup(null)}
                  disabled={restoring}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg disabled:opacity-50"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleConfirmRestore}
                  disabled={restoring}
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25 disabled:opacity-50"
                >
                  {restoring ? 'Restoring Snapshot...' : 'Confirm & Restore'}
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
