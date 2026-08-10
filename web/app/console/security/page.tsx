'use client'

import React from 'react'

export default function SecurityPage() {
  const auditLogs = [
    { time: '2026-08-10 21:15:20', action: 'API_KEY_CREATED', ip: '157.48.22.10', ua: 'Anarva-CLI/1.0', status: 'SUCCESS' },
    { time: '2026-08-10 20:44:12', action: 'DATABASE_PROVISIONED', ip: '157.48.22.10', ua: 'Mozilla/5.0 (Windows NT 10.0)', status: 'SUCCESS' },
    { time: '2026-08-10 19:22:04', action: 'SQL_QUERY_EXECUTED', ip: '157.48.22.10', ua: 'AnarvaConsole/1.0', status: 'SUCCESS' },
  ]

  return (
    <div className="space-y-8">
      <div className="border-b border-slate-800 pb-5">
        <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Security & Audit Log Stream</h1>
        <p className="text-slate-400 text-xs sm:text-sm mt-1">Immutable security event logs, zero-trust authentication events, and authorization records.</p>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
        <h2 className="text-base font-bold text-white">Security Audit Log</h2>
        <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
          {auditLogs.map((log, idx) => (
            <div key={idx} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <div className="font-bold text-white font-sans">{log.action}</div>
                <div className="text-slate-400 text-[11px] mt-0.5">{log.time} • IP: {log.ip} ({log.ua})</div>
              </div>
              <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                {log.status}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
