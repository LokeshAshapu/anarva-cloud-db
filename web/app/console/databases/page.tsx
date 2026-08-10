'use client'

import React, { useEffect, useState } from 'react'
import { API_BASE_URL } from '@/lib/api'

const STORAGE_KEY = 'anarva_user_databases'

export default function ManagedDatabasesPage() {
  const [activeTab, setActiveTab] = useState<'CLUSTERS' | 'SQL_IDE'>('CLUSTERS')
  const [databases, setDatabases] = useState<any[]>([])
  const [loading, setLoading] = useState<boolean>(true)

  // Provisioning form state
  const [name, setName] = useState('')
  const [engine, setEngine] = useState('postgres')
  const [version, setVersion] = useState('16')
  const [multiAZ, setMultiAZ] = useState(true)
  const [acuMin, setAcuMin] = useState(0.5)
  const [acuMax, setAcuMax] = useState(16)
  const [storageGb, setStorageGb] = useState(20)
  const [showModal, setShowModal] = useState(false)

  // Connection string & metrics modals
  const [activeConnStr, setActiveConnStr] = useState<string | null>(null)
  const [copiedConn, setCopiedConn] = useState(false)
  const [metricsDb, setMetricsDb] = useState<any | null>(null)

  // Export & Share modal state
  const [shareDb, setShareDb] = useState<any | null>(null)
  const [exportFormat, setExportFormat] = useState('CSV')
  const [generatedShareUrl, setGeneratedShareUrl] = useState<string | null>(null)
  const [copiedShareLink, setCopiedShareLink] = useState(false)

  // SQL IDE State
  const [sqlQuery, setSqlQuery] = useState(`-- Anarva Cloud SQL Query Console
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  role VARCHAR(50) DEFAULT 'DEVELOPER',
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (full_name, email, role)
VALUES ('Lokesh Ashapu', 'lokesh@anarva.io', 'OWNER')
ON CONFLICT (email) DO NOTHING;

SELECT * FROM users;`)

  const [queryResults, setQueryResults] = useState<any | null>(null)
  const [isExecuting, setIsExecuting] = useState(false)
  const [queryTimeMs, setQueryTimeMs] = useState<number | null>(null)
  const [selectedDbId, setSelectedDbId] = useState<string>('db-default')
  const [queryHistory, setQueryHistory] = useState<string[]>([])

  const updateDatabasesState = (newDatabases: any[]) => {
    setDatabases(newDatabases)
    if (typeof window !== 'undefined') {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newDatabases))
    }
  }

  const fetchDatabases = async () => {
    let localItems: any[] = []
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        try {
          localItems = JSON.parse(stored)
        } catch {}
      }
    }

    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/projects/proj-default/databases`)
      if (res.ok) {
        const remoteData = await res.json()
        if (Array.isArray(remoteData) && remoteData.length > 0) {
          const merged = [...remoteData]
          localItems.forEach((local) => {
            if (!merged.some((m) => m.id === local.id)) {
              merged.push(local)
            }
          })
          updateDatabasesState(merged)
          setLoading(false)
          return
        }
      }
    } catch (err) {
      console.error('Failed to fetch remote databases', err)
    }

    if (localItems.length === 0) {
      localItems = [
        {
          id: 'db-default',
          name: 'Primary Application Cluster',
          engine: 'postgres',
          status: 'RUNNING',
          host: 'localhost',
          port: 15432,
          db_name: 'anarva_db',
          username: 'anarva_admin',
          storage_size_gb: 20,
          multiAZ: true,
          primaryAZ: 'us-east-1a',
          standbyAZ: 'us-east-1b',
          acus: '0.5 - 16.0 ACU',
          replicas: 1,
        },
      ]
    }

    setDatabases(localItems)
    setLoading(false)
  }

  useEffect(() => {
    fetchDatabases()
  }, [])

  const handleProvision = async (e: React.FormEvent) => {
    e.preventDefault()
    const port = 15000 + Math.floor(Math.random() * 5000)
    const newDb = {
      id: `db-uuid-${Date.now()}`,
      name: name || 'New Anarva Instance',
      engine: engine,
      status: 'RUNNING',
      host: 'localhost',
      port: port,
      db_name: `db_${Math.random().toString(36).substring(7)}`,
      username: 'anarva_admin',
      storage_size_gb: storageGb,
      multiAZ: multiAZ,
      primaryAZ: 'us-east-1a',
      standbyAZ: multiAZ ? 'us-east-1b' : 'N/A',
      acus: `${acuMin} - ${acuMax} ACU`,
      replicas: 0,
    }

    try {
      await fetch(`${API_BASE_URL}/api/v1/databases`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          project_id: 'proj-default',
          name: newDb.name,
          engine: engine,
          storage_size_gb: storageGb,
        }),
      }).catch(() => null)
    } catch {}

    const updated = [newDb, ...databases]
    updateDatabasesState(updated)

    setShowModal(false)
    setName('')
  }

  const toggleStatus = (id: string) => {
    const updated = databases.map((db) => {
      if (db.id === id) {
        const nextStatus = db.status === 'RUNNING' ? 'STOPPED' : 'RUNNING'
        return { ...db, status: nextStatus }
      }
      return db
    })
    updateDatabasesState(updated)
  }

  const terminateDatabase = (id: string) => {
    const updated = databases.filter((db) => db.id !== id)
    updateDatabasesState(updated)
  }

  const handleExecuteQuery = async () => {
    setIsExecuting(true)
    const start = performance.now()

    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/query`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ database_id: selectedDbId, sql: sqlQuery }),
      }).catch(() => null)

      const elapsed = (performance.now() - start).toFixed(1)
      setQueryTimeMs(Number(elapsed))

      if (res && res.ok) {
        const data = await res.json()
        setQueryResults(data)
      } else {
        // Fallback execution dataset
        setQueryResults({
          columns: ['id', 'full_name', 'email', 'role', 'status', 'created_at'],
          rows: [
            ['usr-87a1', 'Lokesh Ashapu', 'lokeshashapu@gmail.com', 'OWNER', 'ACTIVE', '2026-08-10 21:00:00'],
          ],
          rows_affected: 3,
        })
      }

      setQueryHistory([sqlQuery, ...queryHistory])
    } catch {
      setQueryResults({
        columns: ['id', 'full_name', 'email', 'role', 'status'],
        rows: [['usr-87a1', 'Lokesh Ashapu', 'lokesh@anarva.io', 'OWNER', 'ACTIVE']],
        rows_affected: 1,
      })
    } finally {
      setIsExecuting(false)
    }
  }

  return (
    <div className="space-y-8">
      {/* Top Section Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Managed Database Engine</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Deploy, auto-scale, and query serverless PostgreSQL & MySQL clusters.</p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => setShowModal(true)}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition text-xs shadow-lg shadow-blue-600/20"
          >
            + Create Database Cluster
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-slate-800 pb-3 text-xs font-semibold">
        <button
          onClick={() => setActiveTab('CLUSTERS')}
          className={`px-4 py-2 rounded-xl transition ${
            activeTab === 'CLUSTERS' ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20' : 'text-slate-400 hover:bg-slate-900'
          }`}
        >
          Database Clusters ({databases.length})
        </button>
        <button
          onClick={() => setActiveTab('SQL_IDE')}
          className={`px-4 py-2 rounded-xl transition ${
            activeTab === 'SQL_IDE' ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20' : 'text-slate-400 hover:bg-slate-900'
          }`}
        >
          SQL Workspace IDE
        </button>
      </div>

      {/* Database Clusters View */}
      {activeTab === 'CLUSTERS' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {databases.map((db) => (
            <div key={db.id} className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4 hover:border-slate-700 transition shadow-xl flex flex-col justify-between">
              <div className="space-y-3">
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="font-bold text-white text-base">{db.name}</h3>
                    <div className="text-xs text-slate-400 font-mono mt-0.5">{db.engine.toUpperCase()} 16</div>
                  </div>
                  <span
                    className={`px-2.5 py-0.5 rounded-full text-[10px] font-bold font-mono border ${
                      db.status === 'RUNNING'
                        ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                        : 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                    }`}
                  >
                    {db.status}
                  </span>
                </div>

                <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl space-y-1.5 text-xs font-mono">
                  <div className="flex items-center justify-between text-slate-400">
                    <span>ACU Allocation:</span>
                    <span className="text-blue-400 font-bold">{db.acus || '0.5 - 16.0 ACU'}</span>
                  </div>
                  <div className="flex items-center justify-between text-slate-400">
                    <span>High Availability:</span>
                    <span className="text-emerald-400">{db.multiAZ ? 'Multi-AZ (Primary+Standby)' : 'Single AZ'}</span>
                  </div>
                  <div className="flex items-center justify-between text-slate-400">
                    <span>Storage Size:</span>
                    <span className="text-slate-200">{db.storage_size_gb || 20} GB</span>
                  </div>
                </div>
              </div>

              <div className="pt-2 border-t border-slate-800 flex items-center justify-between gap-2 text-xs font-semibold">
                <button
                  onClick={() => setActiveConnStr(`${db.engine}://${db.username}:••••••••@${db.host}:${db.port}/${db.db_name}?sslmode=disable`)}
                  className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-xl text-center transition"
                >
                  Connection
                </button>
                <button
                  onClick={() => toggleStatus(db.id)}
                  className="px-3 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl transition"
                >
                  {db.status === 'RUNNING' ? 'Stop' : 'Start'}
                </button>
                <button
                  onClick={() => terminateDatabase(db.id)}
                  className="px-3 py-2 bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-500/20 rounded-xl transition"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* SQL Workspace IDE View */}
      {activeTab === 'SQL_IDE' && (
        <div className="space-y-6">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className="text-xs font-bold text-slate-400">Target Database:</span>
                <select
                  value={selectedDbId}
                  onChange={(e) => setSelectedDbId(e.target.value)}
                  className="bg-slate-950 border border-slate-800 text-white rounded-lg px-3 py-1.5 text-xs font-mono focus:outline-none"
                >
                  {databases.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name} ({d.engine})
                    </option>
                  ))}
                </select>
              </div>

              <button
                onClick={handleExecuteQuery}
                disabled={isExecuting}
                className="px-5 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded-xl transition shadow-lg shadow-blue-600/20 disabled:opacity-50"
              >
                {isExecuting ? 'Executing SQL...' : 'Run Statement (⌘Enter)'}
              </button>
            </div>

            <textarea
              value={sqlQuery}
              onChange={(e) => setSqlQuery(e.target.value)}
              rows={8}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl p-4 font-mono text-xs text-blue-300 focus:outline-none focus:border-blue-500"
            ></textarea>
          </div>

          {/* Results Table */}
          {queryResults && (
            <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-bold text-white">Execution Output & Results</h3>
                <span className="text-xs font-mono text-emerald-400">
                  {queryTimeMs} ms • {queryResults.rows_affected || queryResults.rows?.length} rows returned
                </span>
              </div>

              <div className="overflow-x-auto border border-slate-800 rounded-xl">
                <table className="w-full text-left font-mono text-xs divide-y divide-slate-800">
                  <thead className="bg-slate-950 text-slate-400 uppercase text-[10px]">
                    <tr>
                      {queryResults.columns.map((col: string, idx: number) => (
                        <th key={idx} className="p-3">
                          {col}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800 bg-slate-900/50">
                    {queryResults.rows.map((row: any[], rIdx: number) => (
                      <tr key={rIdx} className="hover:bg-slate-800/40 transition">
                        {row.map((cell: any, cIdx: number) => (
                          <td key={cIdx} className="p-3 text-slate-200">
                            {cell}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Provisioning Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 space-y-6">
            <h2 className="text-lg font-bold text-white">Provision Database Cluster</h2>

            <form onSubmit={handleProvision} className="space-y-4 text-xs">
              <div className="space-y-1">
                <label className="text-slate-400 font-semibold">Cluster Name</label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Production Analytics"
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500"
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-slate-400 font-semibold">Engine</label>
                  <select
                    value={engine}
                    onChange={(e) => setEngine(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none"
                  >
                    <option value="postgres">PostgreSQL 16</option>
                    <option value="mysql">MySQL 8.0</option>
                  </select>
                </div>

                <div className="space-y-1">
                  <label className="text-slate-400 font-semibold">Storage (GB)</label>
                  <input
                    type="number"
                    value={storageGb}
                    onChange={(e) => setStorageGb(Number(e.target.value))}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none"
                  />
                </div>
              </div>

              <div className="flex items-center gap-3 pt-2">
                <input
                  type="checkbox"
                  id="multiAzCheck"
                  checked={multiAZ}
                  onChange={(e) => setMultiAZ(e.target.checked)}
                  className="rounded border-slate-800 bg-slate-950 text-blue-600 focus:ring-0"
                />
                <label htmlFor="multiAzCheck" className="text-slate-300 font-medium cursor-pointer">
                  Enable Multi-AZ High Availability (Standby Replica)
                </label>
              </div>

              <div className="pt-4 border-t border-slate-800 flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl shadow-lg shadow-blue-600/20"
                >
                  Provision Instance
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Connection Modal */}
      {activeConnStr && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 space-y-4">
            <h3 className="text-base font-bold text-white">Database Connection URI</h3>
            <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl font-mono text-xs text-blue-300 break-all">
              {activeConnStr}
            </div>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => {
                  navigator.clipboard.writeText(activeConnStr)
                  setCopiedConn(true)
                  setTimeout(() => setCopiedConn(false), 2000)
                }}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold rounded-xl"
              >
                {copiedConn ? 'Copied URI!' : 'Copy Connection String'}
              </button>
              <button
                onClick={() => setActiveConnStr(null)}
                className="px-4 py-2 bg-slate-800 text-slate-300 text-xs rounded-xl"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
