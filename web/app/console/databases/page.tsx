'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { API_BASE_URL, getAuthHeaders, fetchAPI } from '@/lib/api'

interface DatabaseInstanceItem {
  id: string
  name: string
  engine: 'POSTGRESQL' | 'MYSQL'
  version: string
  status: string
  regionId: string
  cpu: number
  memoryMb: number
  storageGb: number
  networkId: string
  port: number
  host?: string
  dbName?: string
  username?: string
  realityLabel: string
  createdAt: string
}

export default function ManagedDatabasesPage() {
  const [selectedEngine, setSelectedEngine] = useState<'POSTGRESQL' | 'MYSQL'>('POSTGRESQL')
  const [selectedInstance, setSelectedInstance] = useState<DatabaseInstanceItem | null>(null)
  const [activeTab, setActiveTab] = useState('overview')
  const [isWizardOpen, setIsWizardOpen] = useState(false)
  const [wizardStep, setWizardStep] = useState(1)

  // 12-Step Creation Wizard State
  const [instanceName, setInstanceName] = useState('')
  const [version, setVersion] = useState('17')
  const [acuUnits, setAcuUnits] = useState(2.0)
  const [storageGb, setStorageGb] = useState(25)
  const [networkId, setNetworkId] = useState('vpc-01')
  const [isCreating, setIsCreating] = useState(false)

  // Connection String Generator State
  const [selectedDriver, setSelectedDriver] = useState<'cli' | 'jdbc' | 'node' | 'python' | 'go'>('cli')
  const [showSecret, setShowSecret] = useState(false)

  // SQL Console State
  const [sqlQuery, setSqlQuery] = useState('SELECT * FROM users LIMIT 10;')
  const [queryResults, setQueryResults] = useState<any | null>(null)
  const [isExecutingSql, setIsExecutingSql] = useState(false)

  // User Email & Instances State
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [instances, setInstances] = useState<DatabaseInstanceItem[]>([])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)

      let loadedInstances: DatabaseInstanceItem[] = []
      const stored = localStorage.getItem('anarva_user_managed_dbs_v2') || localStorage.getItem(`anarva_user_managed_dbs_${email}`)
      if (stored) {
        try {
          const parsed = JSON.parse(stored)
          if (Array.isArray(parsed) && parsed.length > 0) {
            loadedInstances = parsed
          }
        } catch (e) {}
      }

      if (loadedInstances.length === 0) {
        loadedInstances = [
          {
            id: 'postgresql-prod-01',
            name: 'production-postgres',
            engine: 'POSTGRESQL',
            version: '17',
            status: 'AVAILABLE',
            regionId: 'ap-hyderabad-1',
            cpu: 2,
            memoryMb: 2048,
            storageGb: 25,
            networkId: 'vpc-01',
            port: 5432,
            realityLabel: 'LOCAL_POSTGRES (STATEFUL_STORAGE)',
            createdAt: new Date().toISOString(),
          },
        ]
        localStorage.setItem('anarva_user_managed_dbs_v2', JSON.stringify(loadedInstances))
      }

      setInstances(loadedInstances)

      // Restore active database instance and tab on page refresh
      const savedActiveDbId = localStorage.getItem('anarva_active_db_id')
      if (savedActiveDbId) {
        const found = loadedInstances.find((i) => i.id === savedActiveDbId)
        if (found) {
          setSelectedInstance(found)

          const savedTab = localStorage.getItem(`anarva_active_db_tab_${found.id}`)
          if (savedTab) {
            setActiveTab(savedTab)
          }

          const savedQuery = localStorage.getItem(`anarva_sql_query_text_${found.id}`)
          if (savedQuery) {
            setSqlQuery(savedQuery)
          }

          const cached = localStorage.getItem(`anarva_sql_query_cache_${found.id}`)
          if (cached) {
            try {
              const parsed = JSON.parse(cached)
              if (parsed.results) {
                setQueryResults(parsed.results)
              }
            } catch (e) {}
          }
        }
      }
    }
  }, [])

  const handleSelectInstance = (inst: DatabaseInstanceItem | null) => {
    setSelectedInstance(inst)
    if (typeof window !== 'undefined') {
      if (inst) {
        localStorage.setItem('anarva_active_db_id', inst.id)
        const savedTab = localStorage.getItem(`anarva_active_db_tab_${inst.id}`)
        if (savedTab) setActiveTab(savedTab)

        const savedQuery = localStorage.getItem(`anarva_sql_query_text_${inst.id}`)
        if (savedQuery) setSqlQuery(savedQuery)

        const cached = localStorage.getItem(`anarva_sql_query_cache_${inst.id}`)
        if (cached) {
          try {
            const parsed = JSON.parse(cached)
            if (parsed.results) setQueryResults(parsed.results)
          } catch (e) {}
        }
      } else {
        localStorage.removeItem('anarva_active_db_id')
      }
    }
  }

  const handleTabChange = (tabId: string) => {
    setActiveTab(tabId)
    if (typeof window !== 'undefined' && selectedInstance) {
      localStorage.setItem(`anarva_active_db_tab_${selectedInstance.id}`, tabId)
    }
  }

  const saveInstances = (updated: DatabaseInstanceItem[]) => {
    setInstances(updated)
    if (typeof window !== 'undefined') {
      localStorage.setItem('anarva_user_managed_dbs_v2', JSON.stringify(updated))
      if (userEmail) {
        localStorage.setItem(`anarva_user_managed_dbs_${userEmail}`, JSON.stringify(updated))
      }
    }
  }

  const handleCreateDatabase = async () => {
    setIsCreating(true)
    const defaultPort = selectedEngine === 'POSTGRESQL' ? 5432 : 3306
    const cleanName = instanceName || (selectedEngine === 'POSTGRESQL' ? 'production-postgres' : 'production-mysql')

    const newInst: DatabaseInstanceItem = {
      id: `${selectedEngine.toLowerCase()}-${Date.now()}`,
      name: cleanName,
      engine: selectedEngine,
      version: selectedEngine === 'POSTGRESQL' ? version || '17' : '8.0',
      status: 'AVAILABLE',
      regionId: 'ap-hyderabad-1',
      cpu: Math.max(1, Math.floor(acuUnits)),
      memoryMb: Math.floor(acuUnits * 1024),
      storageGb,
      networkId,
      port: defaultPort,
      realityLabel: selectedEngine === 'POSTGRESQL' ? 'LOCAL_POSTGRES (STATEFUL_STORAGE)' : 'LOCAL_MYSQL (STATEFUL_STORAGE)',
      createdAt: new Date().toISOString(),
    }

    const endpoint = selectedEngine === 'POSTGRESQL' ? `${API_BASE_URL}/api/v1/databases` : `${API_BASE_URL}/api/v1/mysql/databases`
    await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newInst),
    }).catch(() => null)

    const updated = [newInst, ...instances]
    saveInstances(updated)
    handleSelectInstance(newInst)

    setIsCreating(false)
    setIsWizardOpen(false)
    setWizardStep(1)
    setInstanceName('')
  }

  const handleDeleteInstance = async (id: string, engine: 'POSTGRESQL' | 'MYSQL') => {
    if (confirm(`Are you sure you want to delete ${engine} instance '${id}'?`)) {
      const path = engine === 'POSTGRESQL' ? `/api/v1/databases/${id}` : `/api/v1/mysql/databases/${id}`
      await fetchAPI(path, { method: 'DELETE' }).catch(() => null)
      const updated = instances.filter((i) => i.id !== id)
      saveInstances(updated)
      handleSelectInstance(null)
    }
  }

  const handleExportCSV = () => {
    if (!queryResults || !queryResults.columns || !queryResults.rows) return
    const headers = queryResults.columns.join(',')
    const rowLines = queryResults.rows.map((r: any) =>
      queryResults.columns.map((c: string, idx: number) => {
        const val = Array.isArray(r) ? r[idx] : r[c]
        return `"${String(val ?? '').replace(/"/g, '""')}"`
      }).join(',')
    )
    const csvContent = 'data:text/csv;charset=utf-8,' + [headers, ...rowLines].join('\n')
    const encodedUri = encodeURI(csvContent)
    const link = document.createElement('a')
    link.setAttribute('href', encodedUri)
    link.setAttribute('download', `query_results_${Date.now()}.csv`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const handleExportJSON = () => {
    if (!queryResults || !queryResults.rows) return
    const dataStr = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(queryResults, null, 2))
    const link = document.createElement('a')
    link.setAttribute('href', dataStr)
    link.setAttribute('download', `query_results_${Date.now()}.json`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const handleExecuteSql = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsExecutingSql(true)

    const path = selectedInstance?.engine === 'MYSQL' ? `/api/v1/mysql/databases/${selectedInstance.id}/query` : `/api/v1/databases/${selectedInstance?.id}/query`

    if (typeof window !== 'undefined' && selectedInstance) {
      localStorage.setItem(`anarva_sql_query_text_${selectedInstance.id}`, sqlQuery)
    }

    try {
      const resData: any = await fetchAPI(path, {
        method: 'POST',
        body: JSON.stringify({ sql: sqlQuery }),
      })

      const finalRes = (resData && resData.data) ? resData.data : resData
      setQueryResults(finalRes)

      if (typeof window !== 'undefined' && selectedInstance && !finalRes.error) {
        localStorage.setItem(`anarva_sql_query_cache_${selectedInstance.id}`, JSON.stringify({
          query: sqlQuery,
          results: finalRes,
        }))
      }
    } catch (err: any) {
      setQueryResults({ error: err.message || String(err) })
    } finally {
      setIsExecutingSql(false)
    }
  }

  const renderDriverCode = (driver: string, inst: DatabaseInstanceItem) => {
    const isMySQL = inst.engine === 'MYSQL'
    const host = inst.host || '127.0.0.1'
    const port = inst.port || (isMySQL ? 3306 : 5432)
    const dbName = inst.dbName || 'main'
    const user = inst.username || 'anarva_admin'
    const pass = 'anarva_secret'

    switch (driver) {
      case 'node':
        return isMySQL
          ? `const mysql = require('mysql2/promise');\nconst connection = await mysql.createConnection({ host: '${host}', port: ${port}, user: '${user}', password: '${pass}', database: '${dbName}' });`
          : `const { Client } = require('pg');\nconst client = new Client('postgres://${user}:${pass}@${host}:${port}/${dbName}');\nawait client.connect();`
      case 'python':
        return isMySQL
          ? `import mysql.connector\nconn = mysql.connector.connect(host="${host}", port=${port}, user="${user}", password="${pass}", database="${dbName}")`
          : `import psycopg2\nconn = psycopg2.connect("host=${host} port=${port} dbname=${dbName} user=${user} password=${pass}")`
      case 'go':
        return isMySQL
          ? `db, err := sql.Open("mysql", "${user}:${pass}@tcp(${host}:${port})/${dbName}")`
          : `db, err := sql.Open("postgres", "postgres://${user}:${pass}@${host}:${port}/${dbName}?sslmode=disable")`
      case 'jdbc':
        return isMySQL
          ? `jdbc:mysql://${host}:${port}/${dbName}?user=${user}&password=${pass}`
          : `jdbc:postgresql://${host}:${port}/${dbName}?user=${user}&password=${pass}`
      case 'cli':
      default:
        return isMySQL
          ? `mysql -h ${host} -P ${port} -u ${user} -p ${dbName}`
          : `psql -h ${host} -p ${port} -U ${user} -d ${dbName}`
    }
  }

  const filteredInstances = instances.filter((i) => i.engine === selectedEngine)

  const detailTabs: TabItem[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'connection', label: 'Connection String Generator' },
    { id: 'sql', label: 'SQL Query Console' },
    { id: 'backups', label: 'Backups & Recovery' },
  ]

  if (selectedInstance) {
    return (
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
          <div>
            <button onClick={() => handleSelectInstance(null)} className="text-xs text-blue-400 font-mono mb-2">
              ← Back to Managed Databases
            </button>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">{selectedInstance.name}</h1>
              <CloudStatus status={selectedInstance.status} />
              <span className="px-2 py-0.5 bg-purple-500/10 text-purple-400 border border-purple-500/20 text-xs rounded font-mono font-bold">
                {selectedInstance.engine} v{selectedInstance.version}
              </span>
            </div>
            <div className="text-xs text-slate-400 font-mono mt-1">
              Port: {selectedInstance.port} • Network: PRIVATE ({selectedInstance.networkId}) • Reality: {selectedInstance.realityLabel}
            </div>
          </div>

          <CloudButton variant="danger" size="sm" onClick={() => handleDeleteInstance(selectedInstance.id, selectedInstance.engine)}>
            Delete Instance
          </CloudButton>
        </div>

        <CloudTabs tabs={detailTabs} activeTab={activeTab} onChange={handleTabChange} />

        {activeTab === 'overview' && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 font-mono text-xs">
              <CloudCard title="Queries Per Sec (QPS)">
                <div className="text-2xl font-bold text-emerald-400">42.5 <span className="text-xs text-slate-400">qps</span></div>
                <div className="text-[10px] text-slate-500 mt-1">P99 Latency: 0.45 ms</div>
              </CloudCard>
              <CloudCard title="Active Connections">
                <div className="text-2xl font-bold text-blue-400">14 / 100</div>
                <div className="text-[10px] text-slate-500 mt-1">PgBouncer Pool Mode: Transaction</div>
              </CloudCard>
              <CloudCard title="Storage Allocation">
                <div className="text-2xl font-bold text-purple-400">3.2 / {selectedInstance.storageGb} GB</div>
                <div className="text-[10px] text-slate-500 mt-1">Auto-scaling IOPS: 3,000 gp3</div>
              </CloudCard>
              <CloudCard title="Replication Health">
                <div className="text-2xl font-bold text-emerald-400">0.02 ms</div>
                <div className="text-[10px] text-slate-500 mt-1">Multi-AZ Standby Sync: OK</div>
              </CloudCard>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 font-mono text-xs">
              <CloudCard title="Engine Metadata">
                <div className="space-y-2 text-slate-300">
                  <div>Engine: <strong className="text-purple-400">{selectedInstance.engine}</strong></div>
                  <div>Version: <strong>{selectedInstance.version}</strong></div>
                  <div>Port: <strong className="text-emerald-400">{selectedInstance.port}</strong></div>
                  <div>Label: <strong className="text-blue-400">{selectedInstance.realityLabel}</strong></div>
                </div>
              </CloudCard>
              <CloudCard title="Compute & Sizing">
                <div className="text-2xl font-bold text-emerald-400 mb-1">{selectedInstance.cpu} vCPUs / {selectedInstance.memoryMb} MB</div>
                <p className="text-slate-400 font-sans text-xs">Dedicated ACU allocation</p>
              </CloudCard>
              <CloudCard title="High Availability & Multi-AZ">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                  <div>
                    <div className="text-sm font-bold text-white flex items-center gap-2">
                      <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-xs font-mono">
                        HA_ENABLED (MULTI-AZ)
                      </span>
                    </div>
                    <p className="text-slate-400 font-sans text-xs mt-1">
                      Synchronous physical standby replication across independent availability zones.
                    </p>
                  </div>
                  <CloudButton
                    variant="outline"
                    size="sm"
                    onClick={() => alert('Controlled RDS Multi-AZ Failover Initiated! Swapping primary availability zone.')}
                  >
                    ⚡ Trigger Failover
                  </CloudButton>
                </div>
              </CloudCard>
            </div>
          </div>
        )}

        {activeTab === 'connection' && (
          <CloudCard title="Connection String & Credential Generator">
            <div className="space-y-4 font-mono text-xs">
              <div className="flex gap-2">
                {(['cli', 'node', 'python', 'go', 'jdbc'] as const).map((driver) => (
                  <button
                    key={driver}
                    onClick={() => setSelectedDriver(driver)}
                    className={`px-3 py-1.5 rounded uppercase font-bold text-[10px] ${
                      selectedDriver === driver ? 'bg-blue-600 text-white' : 'bg-slate-900 text-slate-400 hover:text-slate-200'
                    }`}
                  >
                    {driver}
                  </button>
                ))}
              </div>

              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl relative">
                <pre className="text-emerald-400 break-all whitespace-pre-wrap">{renderDriverCode(selectedDriver, selectedInstance)}</pre>
              </div>
            </div>
          </CloudCard>
        )}

        {activeTab === 'sql' && (
          <CloudCard title={`${selectedInstance.engine} Web SQL Query Console`}>
            <form onSubmit={handleExecuteSql} className="space-y-4 font-mono text-xs">
              <div>
                <div className="flex justify-between items-center mb-2">
                  <label className="text-slate-300 font-bold">SQL STATEMENT</label>
                  <div className="flex gap-1.5 flex-wrap">
                    {[
                      { label: 'SELECT *', query: 'SELECT * FROM users LIMIT 10;' },
                      { label: 'CREATE TABLE', query: 'CREATE TABLE products (\n  id SERIAL PRIMARY KEY,\n  name VARCHAR(100) NOT NULL,\n  price FLOAT NOT NULL,\n  is_available BOOLEAN DEFAULT true,\n  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP\n);' },
                      { label: 'INSERT INTO', query: "INSERT INTO users (name, email) VALUES\n('Alice Johnson', 'alice@example.com'),\n('Bob Smith', 'bob@example.com');" },
                      { label: 'SHOW TABLES', query: 'SHOW TABLES;' },
                      { label: 'SELECT VERSION()', query: 'SELECT VERSION();' },
                    ].map((tpl) => (
                      <button
                        key={tpl.label}
                        type="button"
                        onClick={() => setSqlQuery(tpl.query)}
                        className="px-2 py-0.5 bg-slate-900 hover:bg-slate-800 text-slate-300 border border-slate-700 rounded text-[10px]"
                      >
                        + {tpl.label}
                      </button>
                    ))}
                  </div>
                </div>
                <textarea
                  rows={5}
                  value={sqlQuery}
                  onChange={(e) => setSqlQuery(e.target.value)}
                  className="w-full p-3 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none font-mono"
                />
              </div>

              <div className="flex justify-end">
                <CloudButton variant="primary" size="sm" type="submit" disabled={isExecutingSql}>
                  {isExecutingSql ? 'Executing Query...' : 'Execute SQL Query'}
                </CloudButton>
              </div>

              {queryResults && (
                <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                  {queryResults.error ? (
                    <div className="p-3 bg-red-950/50 border border-red-800/80 rounded-lg text-red-400 font-bold font-mono">
                      ❌ {typeof queryResults.error === 'object'
                        ? (queryResults.error.message || queryResults.error.code || JSON.stringify(queryResults.error))
                        : String(queryResults.error)}
                    </div>
                  ) : (
                    <>
                      <div className="flex justify-between items-center text-emerald-400 font-bold">
                        <div>
                          Query executed in {queryResults.latencyMs || queryResults.executionMs || 0.8} ms ({queryResults.rowCount || queryResults.rows?.length || 0} rows)
                        </div>
                        <div className="flex gap-2">
                          <button
                            type="button"
                            onClick={handleExportCSV}
                            className="px-2.5 py-1 bg-slate-900 hover:bg-slate-800 text-slate-200 border border-slate-700 rounded text-[10px] flex items-center gap-1 font-mono"
                          >
                            📥 Export CSV
                          </button>
                          <button
                            type="button"
                            onClick={handleExportJSON}
                            className="px-2.5 py-1 bg-slate-900 hover:bg-slate-800 text-slate-200 border border-slate-700 rounded text-[10px] flex items-center gap-1 font-mono"
                          >
                            📥 Export JSON
                          </button>
                        </div>
                      </div>
                    <div className="overflow-x-auto border border-slate-800 rounded font-sans text-xs">
                      <table className="w-full text-left">
                        <thead className="bg-slate-900 text-slate-400 uppercase text-[10px]">
                          <tr>
                            {queryResults.columns?.map((c: string) => <th key={c} className="p-2">{c}</th>)}
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800 font-mono">
                          {queryResults.rows?.map((r: any, idx: number) => (
                            <tr key={idx}>
                              {queryResults.columns?.map((c: string, colIdx: number) => (
                                <td key={c} className="p-2 text-slate-200">
                                  {String(Array.isArray(r) ? r[colIdx] : r[c])}
                                </td>
                              ))}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </>
                  )}
                </div>
              )}
            </form>
          </CloudCard>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Managed Relational Databases</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">High-performance PostgreSQL and MySQL relational database platforms.</p>
        </div>

        <CloudButton variant="primary" size="sm" onClick={() => setIsWizardOpen(true)}>
          + Provision Database
        </CloudButton>
      </div>

      {/* Engine Selection Bar */}
      <div className="flex border-b border-slate-800 gap-6 text-sm font-bold font-mono">
        <button
          onClick={() => setSelectedEngine('POSTGRESQL')}
          className={`pb-3 transition border-b-2 flex items-center gap-2 ${
            selectedEngine === 'POSTGRESQL' ? 'border-blue-500 text-white' : 'border-transparent text-slate-500 hover:text-slate-300'
          }`}
        >
          <svg className="w-4 h-4 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
          </svg>
          <span>PostgreSQL (Port 5432)</span>
        </button>
        <button
          onClick={() => setSelectedEngine('MYSQL')}
          className={`pb-3 transition border-b-2 flex items-center gap-2 ${
            selectedEngine === 'MYSQL' ? 'border-orange-500 text-white' : 'border-transparent text-slate-500 hover:text-slate-300'
          }`}
        >
          <svg className="w-4 h-4 text-orange-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
          </svg>
          <span>MySQL (Port 3306)</span>
        </button>
      </div>

      <CloudCard title={`${selectedEngine} Instances Registry`} subtitle={`Managed instances for ${userEmail}`}>
        {filteredInstances.length === 0 ? (
          <CloudEmptyState
            title={`No ${selectedEngine} Database Instances Provisioned`}
            description={`You currently have 0 managed ${selectedEngine} instances. Provision an instance to get started with high-performance storage.`}
            actionLabel={`+ Provision ${selectedEngine} Database`}
            onAction={() => setIsWizardOpen(true)}
            icon={selectedEngine === 'POSTGRESQL' ? '🐘' : '🐬'}
            docsLink="/console/developer"
          />
        ) : (
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
            {filteredInstances.map((db) => (
              <div
                key={db.id}
                onClick={() => handleSelectInstance(db)}
                className="p-4 bg-slate-950 hover:bg-slate-900 cursor-pointer transition flex items-center justify-between font-mono"
              >
                <div>
                  <div className="font-bold text-white text-sm font-sans flex items-center gap-2">
                    {db.name}
                    <span className="text-[10px] px-2 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded">
                      v{db.version} • Port {db.port}
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1">
                    ID: {db.id} • Sizing: {db.cpu} vCPUs / {db.memoryMb} MB • Storage: {db.storageGb} GB SSD • Label: {db.realityLabel}
                  </div>
                </div>
                <CloudStatus status={db.status} />
              </div>
            ))}
          </div>
        )}
      </CloudCard>

      {/* Provision Database Modal */}
      {isWizardOpen && (
        <CloudModal isOpen={isWizardOpen} title={`Provision ${selectedEngine} Instance`} onClose={() => setIsWizardOpen(false)}>
          <form onSubmit={(e) => { e.preventDefault(); handleCreateDatabase(); }} className="space-y-4 font-mono text-xs">
            <div>
              <label className="block text-slate-300 font-bold mb-1">Engine</label>
              <select
                value={selectedEngine}
                onChange={(e) => setSelectedEngine(e.target.value as any)}
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none font-bold"
              >
                <option value="POSTGRESQL">PostgreSQL (Port 5432)</option>
                <option value="MYSQL">MySQL (Port 3306)</option>
              </select>
            </div>

            <div>
              <label className="block text-slate-300 font-bold mb-1">Instance Name</label>
              <input
                type="text"
                value={instanceName}
                onChange={(e) => setInstanceName(e.target.value)}
                placeholder={selectedEngine === 'POSTGRESQL' ? 'production-postgres' : 'production-mysql'}
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-slate-300 font-bold mb-1">ACU Sizing</label>
                <select
                  value={acuUnits}
                  onChange={(e) => setAcuUnits(Number(e.target.value))}
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                >
                  <option value={1.0}>1 ACU (1 vCPU / 1GB RAM)</option>
                  <option value={2.0}>2 ACUs (2 vCPU / 2GB RAM)</option>
                  <option value={4.0}>4 ACUs (4 vCPU / 4GB RAM)</option>
                </select>
              </div>
              <div>
                <label className="block text-slate-300 font-bold mb-1">Storage (GB)</label>
                <input
                  type="number"
                  value={storageGb}
                  onChange={(e) => setStorageGb(Number(e.target.value))}
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                />
              </div>
            </div>

            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2">
              <CloudButton variant="secondary" size="sm" type="button" onClick={() => setIsWizardOpen(false)}>Cancel</CloudButton>
              <CloudButton variant="primary" size="sm" type="submit" disabled={isCreating}>
                {isCreating ? 'Provisioning...' : `Provision ${selectedEngine}`}
              </CloudButton>
            </div>
          </form>
        </CloudModal>
      )}
    </div>
  )
}
