'use client'

import React, { useState } from 'react'

export default function DatabasesPage() {
  const [databases, setDatabases] = useState([
    {
      id: 'db-uuid-1',
      name: 'Primary Application Database',
      engine: 'postgres',
      status: 'RUNNING',
      host: 'localhost',
      port: 15432,
      db_name: 'db_prod_main',
      storage_size_gb: 10,
    },
    {
      id: 'db-uuid-2',
      name: 'Analytics Data Warehouse',
      engine: 'postgres',
      status: 'RUNNING',
      host: 'localhost',
      port: 15433,
      db_name: 'db_analytics',
      storage_size_gb: 20,
    },
  ])

  const [name, setName] = useState('')
  const [engine, setEngine] = useState('postgres')
  const [showModal, setShowModal] = useState(false)

  const handleProvision = (e: React.FormEvent) => {
    e.preventDefault()
    const newDb = {
      id: `db-uuid-${Date.now()}`,
      name: name || 'New Managed Instance',
      engine: engine,
      status: 'RUNNING',
      host: 'localhost',
      port: 15000 + Math.floor(Math.random() * 5000),
      db_name: `db_${Math.random().toString(36).substring(7)}`,
      storage_size_gb: 10,
    }
    setDatabases([...databases, newDb])
    setName('')
    setShowModal(false)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Managed Databases</h1>
          <p className="text-slate-400 mt-1">Provision and manage serverless database instances across global regions.</p>
        </div>

        <button
          onClick={() => setShowModal(true)}
          className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25"
        >
          + Provision Database
        </button>
      </div>

      {/* Database Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {databases.map((db) => (
          <div key={db.id} className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-bold text-white text-lg">{db.name}</h3>
                <div className="text-xs text-slate-400 font-mono mt-0.5">{db.id}</div>
              </div>

              <span className="px-2.5 py-1 text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full flex items-center gap-1.5">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                {db.status}
              </span>
            </div>

            <div className="grid grid-cols-2 gap-3 text-xs bg-slate-950 p-3 rounded-lg border border-slate-800 font-mono text-slate-300">
              <div>Engine: <span className="text-blue-400">{db.engine.toUpperCase()}</span></div>
              <div>Port: <span className="text-white">{db.port}</span></div>
              <div>Database: <span className="text-white">{db.db_name}</span></div>
              <div>Storage: <span className="text-white">{db.storage_size_gb} GB</span></div>
            </div>

            <div className="flex items-center gap-2 pt-2">
              <button className="flex-1 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium rounded-lg transition">
                Connection String
              </button>
              <button className="py-1.5 px-3 bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 text-xs font-medium rounded-lg transition border border-amber-500/20">
                Stop
              </button>
              <button className="py-1.5 px-3 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-xs font-medium rounded-lg transition border border-red-500/20">
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

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
    </div>
  )
}
