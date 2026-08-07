'use client'

import React, { useState } from 'react'

export default function BackupsPage() {
  const [backups] = useState([
    {
      id: 'snap-uuid-101',
      name: 'Automated Daily Snapshot',
      db: 'Primary Application Database',
      type: 'SNAPSHOT',
      status: 'COMPLETED',
      size: '108 B',
      created_at: '2026-08-06 18:00:00',
    },
  ])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Backups & Point-In-Time Restore</h1>
          <p className="text-slate-400 mt-1">Managed automated snapshot dumps and object storage archives.</p>
        </div>

        <button className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25">
          + Trigger Snapshot Backup
        </button>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
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
                <td className="p-4">{b.db}</td>
                <td className="p-4"><span className="px-2 py-0.5 bg-slate-800 text-slate-300 rounded">{b.type}</span></td>
                <td className="p-4">{b.size}</td>
                <td className="p-4"><span className="text-emerald-400 font-semibold">{b.status}</span></td>
                <td className="p-4 text-slate-400">{b.created_at}</td>
                <td className="p-4 text-right space-x-2">
                  <button className="px-3 py-1 bg-blue-600/20 text-blue-400 hover:bg-blue-600/30 rounded transition">
                    Restore
                  </button>
                  <button className="px-3 py-1 bg-red-600/20 text-red-400 hover:bg-red-600/30 rounded transition">
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
