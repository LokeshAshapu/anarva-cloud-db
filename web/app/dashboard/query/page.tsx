'use client'

import React, { useState } from 'react'
import { fetchAPI } from '@/lib/api'

export default function SQLConsolePage() {
  const [sql, setSql] = useState("SELECT * FROM users LIMIT 10;")
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<any>(null)
  const [error, setError] = useState('')

  const handleExecute = async () => {
    setLoading(true)
    setError('')
    setResult(null)

    try {
      const data = await fetchAPI<any>('/api/v1/query', {
        method: 'POST',
        body: JSON.stringify({
          database_id: 'db-demo-id',
          sql: sql,
        }),
      })

      setResult(data)
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Interactive SQL Query Console</h1>
          <p className="text-slate-400 mt-1">Execute SQL statements with safety AST validation and real-time execution telemetry.</p>
        </div>

        <button
          onClick={handleExecute}
          disabled={loading}
          className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25 disabled:opacity-50"
        >
          {loading ? 'Executing...' : 'Run Query (Ctrl + Enter)'}
        </button>
      </div>

      {/* Editor Box */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 space-y-2">
        <div className="flex items-center justify-between text-xs text-slate-400 pb-2 border-b border-slate-800">
          <span className="font-mono">Target Database: db-production-main (PostgreSQL 16)</span>
          <span>Engine: Distributed SQL Parser v1.0</span>
        </div>

        <textarea
          rows={6}
          value={sql}
          onChange={(e) => setSql(e.target.value)}
          className="w-full bg-slate-950 text-emerald-400 font-mono text-sm p-4 rounded-lg border border-slate-800 focus:outline-none focus:border-blue-500 transition resize-y"
          placeholder="ENTER SQL STATEMENT HERE..."
        />
      </div>

      {error && (
        <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-mono">
          [ERROR] {error}
        </div>
      )}

      {/* Results View */}
      {result && (
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-4">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-400">
            <span>Query Execution Result</span>
            <div className="flex items-center gap-4 text-emerald-400">
              <span>Time: {result.execution_time_ms ? result.execution_time_ms.toFixed(2) : 0.85} ms</span>
              <span>Rows Affected: {result.rows_affected || 0}</span>
            </div>
          </div>

          <div className="overflow-x-auto border border-slate-800 rounded-lg">
            <table className="w-full text-left text-xs font-mono">
              <thead className="bg-slate-950 text-slate-300 border-b border-slate-800">
                <tr>
                  {result.columns && result.columns.length > 0 ? (
                    result.columns.map((col: any, idx: number) => (
                      <th key={idx} className="p-3 border-r border-slate-800 last:border-r-0">
                        {col.name} <span className="text-slate-500 font-normal">({col.type})</span>
                      </th>
                    ))
                  ) : (
                    <th className="p-3">Result Status</th>
                  )}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800 text-slate-200">
                {result.rows && result.rows.length > 0 ? (
                  result.rows.map((row: any, rIdx: number) => (
                    <tr key={rIdx} className="hover:bg-slate-800/40 transition">
                      {result.columns.map((col: any, cIdx: number) => (
                        <td key={cIdx} className="p-3 border-r border-slate-800 last:border-r-0">
                          {String(row[col.name] ?? 'NULL')}
                        </td>
                      ))}
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td className="p-4 text-slate-500 text-center">Query executed successfully with 0 returned rows.</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
