'use client'

import React, { useState } from 'react'

export default function BackupsPage() {
  const [backups, setBackups] = useState([
    { id: 'snap-01a9', name: 'Automated Daily Snapshot', cluster: 'Primary Application Cluster', size: '210 MB', status: 'COMPLETED', date: '2026-08-10 02:00:00' },
    { id: 'snap-88c2', name: 'Pre-Deployment Manual Backup', cluster: 'Primary Application Cluster', size: '204 MB', status: 'COMPLETED', date: '2026-08-09 18:45:00' },
  ])

  const [isTriggering, setIsTriggering] = useState(false)
  const [restoreModalSnap, setRestoreModalSnap] = useState<any | null>(null)
  const [restoreSuccess, setRestoreSuccess] = useState(false)

  const handleTriggerBackup = () => {
    setIsTriggering(true)
    setTimeout(() => {
      const newSnap = {
        id: `snap-${Math.floor(Math.random() * 9000 + 1000)}`,
        name: 'Manual Instant Snapshot',
        cluster: 'Primary Application Cluster',
        size: '212 MB',
        status: 'COMPLETED',
        date: new Date().toISOString().replace('T', ' ').substring(0, 19),
      }
      setBackups([newSnap, ...backups])
      setIsTriggering(false)
    }, 1500)
  }

  const handleRestore = () => {
    setRestoreSuccess(true)
    setTimeout(() => {
      setRestoreSuccess(false)
      setRestoreModalSnap(null)
    }, 2000)
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Point-in-Time Backups & Disaster Recovery</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Automated daily snapshots, streaming WAL archives, and 1-click database cluster recovery.</p>
        </div>

        <button
          onClick={handleTriggerBackup}
          disabled={isTriggering}
          className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition text-xs shadow-lg shadow-blue-600/20 disabled:opacity-50"
        >
          {isTriggering ? 'Creating Snapshot...' : '+ Trigger Manual Snapshot'}
        </button>
      </div>

      {/* Backup Strategy Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Automated Backup Schedule</div>
          <div className="text-3xl font-extrabold text-emerald-400 font-mono">DAILY @ 02:00 UTC</div>
          <div className="text-xs text-slate-400">7-Day Retention Window</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">WAL Streaming Archival</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">ACTIVE</div>
          <div className="text-xs text-slate-400">Recovery RPO: &lt; 5 Seconds</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Total Stored Snapshots</div>
          <div className="text-3xl font-extrabold text-white font-mono">{backups.length}</div>
          <div className="text-xs text-slate-400">Total Backup Size: 414 MB</div>
        </div>
      </div>

      {/* Snapshots Table */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
        <h2 className="text-base font-bold text-white">Database Snapshot Archives</h2>

        <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
          {backups.map((snap) => (
            <div key={snap.id} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <div className="font-bold text-white font-sans text-sm">{snap.name}</div>
                <div className="text-slate-400 text-[11px] mt-0.5">{snap.id} • {snap.cluster} • {snap.size} • Created: {snap.date}</div>
              </div>

              <div className="flex items-center gap-3">
                <span className="px-2.5 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  {snap.status}
                </span>
                <button
                  onClick={() => setRestoreModalSnap(snap)}
                  className="px-3 py-1.5 bg-blue-600/10 hover:bg-blue-600/20 text-blue-400 border border-blue-500/20 rounded-xl text-[11px] font-sans font-semibold transition"
                >
                  Restore PITR
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Restore PITR Modal */}
      {restoreModalSnap && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 space-y-4">
            <h3 className="text-base font-bold text-white">Confirm Point-in-Time Recovery</h3>
            <p className="text-xs text-slate-400">
              Are you sure you want to restore cluster <strong className="text-white">{restoreModalSnap.cluster}</strong> from snapshot <strong className="text-white">{restoreModalSnap.name}</strong> ({restoreModalSnap.date})?
            </p>

            {restoreSuccess ? (
              <div className="p-3 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-bold rounded-xl text-center">
                Database Cluster Restored Successfully!
              </div>
            ) : (
              <div className="flex justify-end gap-3 pt-2">
                <button
                  onClick={() => setRestoreModalSnap(null)}
                  className="px-4 py-2 bg-slate-800 text-slate-300 text-xs rounded-xl"
                >
                  Cancel
                </button>
                <button
                  onClick={handleRestore}
                  className="px-5 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold rounded-xl"
                >
                  Confirm Restore
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
