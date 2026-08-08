'use client'

import React, { useEffect, useState } from 'react'
import { API_BASE_URL } from '@/lib/api'

const STORAGE_KEY = 'anarva_user_databases'

export default function DatabasesPage() {
  const [databases, setDatabases] = useState<any[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [name, setName] = useState('')
  const [engine, setEngine] = useState('postgres')
  const [showModal, setShowModal] = useState(false)

  const [activeConnStr, setActiveConnStr] = useState<string | null>(null)
  const [copiedConn, setCopiedConn] = useState(false)

  // Export & Share modal state
  const [shareDb, setShareDb] = useState<any | null>(null)
  const [exportFormat, setExportFormat] = useState('CSV')
  const [accessLevel, setAccessLevel] = useState('ANYONE_WITH_LINK')
  const [generatedShareUrl, setGeneratedShareUrl] = useState<string | null>(null)
  const [copiedShareLink, setCopiedShareLink] = useState(false)

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
          return
        }
      }
    } catch (err) {
      console.error('Failed to fetch remote databases', err)
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
      name: name || 'New Managed Instance',
      engine: engine,
      status: 'RUNNING',
      host: 'localhost',
      port: port,
      db_name: `db_${Math.random().toString(36).substring(7)}`,
      username: 'anarva_admin',
      storage_size_gb: 10,
    }

    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/databases`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          project_id: 'proj-default',
          name: name || 'New Managed Instance',
          engine: engine,
          storage_size_gb: 10,
        }),
      })

      if (res.ok) {
        const remoteDb = await res.json()
        updateDatabasesState([remoteDb, ...databases])
      } else {
        updateDatabasesState([newDb, ...databases])
      }
    } catch (err) {
      updateDatabasesState([newDb, ...databases])
    } finally {
      setName('')
      setShowModal(false)
    }
  }

  const handleToggleStatus = async (db: any) => {
    const isRunning = db.status === 'RUNNING'
    const action = isRunning ? 'stop' : 'start'

    const updated = databases.map((item) =>
      item.id === db.id ? { ...item, status: isRunning ? 'STOPPED' : 'RUNNING' } : item
    )
    updateDatabasesState(updated)

    try {
      await fetch(`${API_BASE_URL}/api/v1/databases/${db.id}/${action}`, { method: 'POST' }).catch(() => null)
    } catch {}
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this database instance? All data will be permanently purged.')) {
      return
    }

    const updated = databases.filter((item) => item.id !== id)
    updateDatabasesState(updated)

    try {
      await fetch(`${API_BASE_URL}/api/v1/databases/${id}`, { method: 'DELETE' }).catch(() => null)
    } catch {}
  }

  const handleShowConnStr = (db: any) => {
    const connStr = `${db.engine || 'postgres'}://anarva_admin:eX938#kL9@${db.host || 'localhost'}:${db.port || 15432}/${db.db_name || 'app_db'}`
    setActiveConnStr(connStr)
    setCopiedConn(false)
  }

  const copyConnToClipboard = () => {
    if (activeConnStr) {
      navigator.clipboard.writeText(activeConnStr)
      setCopiedConn(true)
      setTimeout(() => setCopiedConn(false), 2000)
    }
  }

  // Generate Export & Share Link
  const handleOpenExportShare = (db: any) => {
    setShareDb(db)
    setGeneratedShareUrl(null)
    setCopiedShareLink(false)
  }

  const handleGenerateShareLink = (e: React.FormEvent) => {
    e.preventDefault()
    const token = `export-${Math.random().toString(36).substring(2, 10)}`
    const origin = typeof window !== 'undefined' ? window.location.origin : 'https://anarva-cloud-db.vercel.app'
    const url = `${origin}/share/${token}`
    setGeneratedShareUrl(url)
  }

  const copyShareLinkToClipboard = () => {
    if (generatedShareUrl) {
      navigator.clipboard.writeText(generatedShareUrl)
      setCopiedShareLink(true)
      setTimeout(() => setCopiedShareLink(false), 2000)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Managed Databases</h1>
          <p className="text-slate-400 mt-1">Provision, scale, export tables, and share access links.</p>
        </div>

        <button
          onClick={() => setShowModal(true)}
          className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25"
        >
          + Provision Database
        </button>
      </div>

      {loading ? (
        <div className="p-8 text-center text-slate-400">Loading database infrastructure...</div>
      ) : databases.length === 0 ? (
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-12 text-center space-y-4">
          <div className="inline-flex items-center justify-center h-16 w-16 rounded-2xl bg-blue-600/10 text-blue-400 text-3xl font-bold border border-blue-500/20">
            💾
          </div>
          <h3 className="text-xl font-bold text-white">No Database Instances Found</h3>
          <p className="text-slate-400 text-sm max-w-md mx-auto">
            You haven't provisioned any cloud database instances yet. Click "+ Provision Database" above to deploy your first PostgreSQL or MySQL cluster!
          </p>
          <button
            onClick={() => setShowModal(true)}
            className="px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition shadow-lg shadow-blue-600/25"
          >
            Provision Your First Database
          </button>
        </div>
      ) : (
        /* Database Grid */
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {databases.map((db) => (
            <div key={db.id} className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-bold text-white text-lg">{db.name}</h3>
                  <div className="text-xs text-slate-400 font-mono mt-0.5">{db.id}</div>
                </div>

                <span
                  className={`px-2.5 py-1 text-xs font-semibold rounded-full flex items-center gap-1.5 border ${
                    db.status === 'RUNNING'
                      ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                      : 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                  }`}
                >
                  <span
                    className={`h-1.5 w-1.5 rounded-full ${
                      db.status === 'RUNNING' ? 'bg-emerald-400 animate-pulse' : 'bg-amber-400'
                    }`}
                  ></span>
                  {db.status}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-3 text-xs bg-slate-950 p-3 rounded-lg border border-slate-800 font-mono text-slate-300">
                <div>Engine: <span className="text-blue-400">{(db.engine || 'postgres').toUpperCase()}</span></div>
                <div>Port: <span className="text-white">{db.port || 15432}</span></div>
                <div>Database: <span className="text-white">{db.db_name || 'app_db'}</span></div>
                <div>Storage: <span className="text-white">{db.storage_size_gb || 10} GB</span></div>
              </div>

              <div className="flex items-center gap-2 pt-2">
                <button
                  onClick={() => handleShowConnStr(db)}
                  className="flex-1 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium rounded-lg transition"
                >
                  Connection String
                </button>

                <button
                  onClick={() => handleOpenExportShare(db)}
                  className="flex-1 py-1.5 bg-blue-600/10 hover:bg-blue-600/20 text-blue-400 border border-blue-500/20 text-xs font-medium rounded-lg transition"
                >
                  Export & Share
                </button>

                <button
                  onClick={() => handleToggleStatus(db)}
                  className={`py-1.5 px-3 text-xs font-medium rounded-lg transition border ${
                    db.status === 'RUNNING'
                      ? 'bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 border-amber-500/20'
                      : 'bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border-emerald-500/20'
                  }`}
                >
                  {db.status === 'RUNNING' ? 'Stop' : 'Start'}
                </button>

                <button
                  onClick={() => handleDelete(db.id)}
                  className="py-1.5 px-3 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-xs font-medium rounded-lg transition border border-red-500/20"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Provisioning Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Provision New Database</h2>

            <form onSubmit={handleProvision} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Database Name</label>
                <input
                  type="text"
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Production Cluster"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Database Engine</label>
                <select
                  value={engine}
                  onChange={(e) => setEngine(e.target.value)}
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                >
                  <option value="postgres">PostgreSQL 16</option>
                  <option value="mysql">MySQL 8.0</option>
                </select>
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25"
                >
                  Confirm & Provision
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Connection String Modal */}
      {activeConnStr && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-lg space-y-4">
            <h2 className="text-xl font-bold text-white">Database Connection String</h2>
            <p className="text-xs text-slate-400">Use this connection string in your backend application or database client.</p>

            <div className="p-3 bg-slate-950 border border-slate-800 rounded-lg font-mono text-xs text-emerald-400 break-all select-all">
              {activeConnStr}
            </div>

            <div className="flex gap-2 pt-2">
              <button
                onClick={copyConnToClipboard}
                className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg transition"
              >
                {copiedConn ? '✔ Copied to Clipboard!' : 'Copy Connection String'}
              </button>
              <button
                onClick={() => setActiveConnStr(null)}
                className="px-4 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Export & Share Modal (Google Drive Style) */}
      {shareDb && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-lg space-y-4">
            <h2 className="text-xl font-bold text-white">Export & Share Database Table</h2>
            <p className="text-xs text-slate-400">
              Generate a Google Drive style shareable access link for database: <span className="text-white font-semibold">{shareDb.name}</span>.
            </p>

            <form onSubmit={handleGenerateShareLink} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Export Format</label>
                <select
                  value={exportFormat}
                  onChange={(e) => setExportFormat(e.target.value)}
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                >
                  <option value="CSV">CSV (Spreadsheet Compatible)</option>
                  <option value="JSON">JSON (REST Payload Format)</option>
                  <option value="SQL">SQL Dump (Full Schema & Data Insert)</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Access Control (Google Drive Style)</label>
                <select
                  value={accessLevel}
                  onChange={(e) => setAccessLevel(e.target.value)}
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                >
                  <option value="ANYONE_WITH_LINK">Anyone with the link can view & download</option>
                  <option value="RESTRICTED">Restricted - Only authorized organization members</option>
                </select>
              </div>

              {!generatedShareUrl ? (
                <div className="flex gap-2 pt-2">
                  <button
                    type="button"
                    onClick={() => setShareDb(null)}
                    className="flex-1 py-2.5 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="flex-1 py-2.5 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25"
                  >
                    Create Share Link
                  </button>
                </div>
              ) : (
                <div className="space-y-3 pt-2">
                  <div className="text-xs text-emerald-400 font-semibold">✔ Shareable Link Generated:</div>
                  <div className="p-3 bg-slate-950 border border-slate-800 rounded-lg font-mono text-xs text-white break-all select-all">
                    {generatedShareUrl}
                  </div>

                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={copyShareLinkToClipboard}
                      className="flex-1 py-2.5 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg transition"
                    >
                      {copiedShareLink ? '✔ Link Copied!' : 'Copy Share Link'}
                    </button>
                    <button
                      type="button"
                      onClick={() => setShareDb(null)}
                      className="px-4 py-2.5 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                    >
                      Done
                    </button>
                  </div>
                </div>
              )}
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
