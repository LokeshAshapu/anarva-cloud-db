'use client'

import React, { useState } from 'react'

export default function APIKeysPage() {
  const [apiKeys] = useState([
    {
      id: 'key-uuid-1',
      name: 'CLI Deployment Key',
      prefix: 'anarva_live',
      created_at: '2026-08-06',
      status: 'ACTIVE',
    },
  ])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">API Keys & Security Audit</h1>
          <p className="text-slate-400 mt-1">Manage programmatic access tokens and security audit logs.</p>
        </div>

        <button className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25">
          + Generate New API Key
        </button>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-4">
        <h2 className="text-lg font-bold text-white">Active API Keys</h2>

        <div className="divide-y divide-slate-800 border border-slate-800 rounded-lg overflow-hidden">
          {apiKeys.map((k) => (
            <div key={k.id} className="p-4 flex items-center justify-between bg-slate-950">
              <div>
                <div className="font-semibold text-white">{k.name}</div>
                <div className="text-xs text-slate-400 font-mono mt-0.5">Prefix: {k.prefix}_***</div>
              </div>

              <div className="flex items-center gap-4">
                <span className="text-xs text-slate-500 font-mono">Created {k.created_at}</span>
                <button className="px-3 py-1 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-xs font-medium rounded-lg transition border border-red-500/20">
                  Revoke Key
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
