'use client'

import React, { useState, useEffect } from 'react'
import { CloudResource, ResourceStatus, RegionId } from '@/types/resource'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { generateARNV } from '@/lib/arnv'
import { API_BASE_URL } from '@/lib/api'

interface DatabaseClusterItem {
  id: string
  resourceId: string
  name: string
  engine: 'PostgreSQL' | 'MySQL'
  engineVersion: string
  status: ResourceStatus
  regionId: RegionId
  environment: string
  computeUnits: number
  storageGb: number
  maxStorageGb: number
  autoScaling: boolean
  backupEnabled: boolean
  backupRetentionDays: number
  pitrEnabled: boolean
  highAvailability: boolean
  host: string
  port: number
  dbname: string
  username: string
  createdAt: string
}

export default function DatabasesPage() {
  const [selectedCluster, setSelectedCluster] = useState<DatabaseClusterItem | null>(null)
  const [activeTab, setActiveTab] = useState('overview')
  const [isWizardOpen, setIsWizardOpen] = useState(false)
  const [wizardStep, setWizardStep] = useState(1)

  // Creation Wizard State
  const [engine, setEngine] = useState<'PostgreSQL' | 'MySQL'>('PostgreSQL')
  const [engineVersion, setEngineVersion] = useState('17.2')
  const [name, setName] = useState('')
  const [computeUnits, setComputeUnits] = useState(2)
  const [storageGb, setStorageGb] = useState(48)
  const [maxStorageGb, setMaxStorageGb] = useState(256)
  const [autoScaling, setAutoScaling] = useState(true)
  const [regionId, setRegionId] = useState<RegionId>('ap-hyderabad-1')
  const [highAvailability, setHighAvailability] = useState(true)
  const [backupRetentionDays, setBackupRetentionDays] = useState(7)
  const [pitrEnabled, setPitrEnabled] = useState(true)
  const [isCreating, setIsCreating] = useState(false)

  // SQL Editor State
  const [sqlQuery, setSqlQuery] = useState('SELECT * FROM users LIMIT 10;')
  const [queryResults, setQueryResults] = useState<any | null>(null)
  const [isExecutingSql, setIsExecutingSql] = useState(false)
  const [showPassword, setShowPassword] = useState(false)

  // User Email & Clusters State
  const [userEmail, setUserEmail] = useState('lokeshashapu@gmail.com')
  const [clusters, setClusters] = useState<DatabaseClusterItem[]>([])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'lokeshashapu@gmail.com'
      setUserEmail(email)

      const dbKey = `anarva_user_databases_${email}`
      const stored = localStorage.getItem(dbKey)

      if (stored) {
        setClusters(JSON.parse(stored))
      } else if (email === 'lokeshashapu@gmail.com') {
        const defaults: DatabaseClusterItem[] = [
          {
            id: 'res-db-prod-1',
            resourceId: 'arnv:db:ap-hyderabad-1:proj-default:database/production-db',
            name: 'production-db',
            engine: 'PostgreSQL',
            engineVersion: '17.2',
            status: 'AVAILABLE',
            regionId: 'ap-hyderabad-1',
            environment: 'Production',
            computeUnits: 2.0,
            storageGb: 48,
            maxStorageGb: 256,
            autoScaling: true,
            backupEnabled: true,
            backupRetentionDays: 7,
            pitrEnabled: true,
            highAvailability: true,
            host: 'db-prod-1.anarva.cloud',
            port: 5432,
            dbname: 'production_db',
            username: 'anarva_admin',
            createdAt: new Date().toISOString(),
          },
          {
            id: 'res-db-analytics-1',
            resourceId: 'arnv:db:ap-mumbai-1:proj-default:database/analytics-db',
            name: 'analytics-db',
            engine: 'PostgreSQL',
            engineVersion: '16.4',
            status: 'AVAILABLE',
            regionId: 'ap-mumbai-1',
            environment: 'Production',
            computeUnits: 4.0,
            storageGb: 120,
            maxStorageGb: 512,
            autoScaling: true,
            backupEnabled: true,
            backupRetentionDays: 14,
            pitrEnabled: true,
            highAvailability: false,
            host: 'db-analytics-1.anarva.cloud',
            port: 5432,
            dbname: 'analytics_db',
            username: 'anarva_analytics',
            createdAt: new Date().toISOString(),
          },
        ]
        setClusters(defaults)
        localStorage.setItem(dbKey, JSON.stringify(defaults))
      } else {
        setClusters([])
      }
    }
  }, [])

  const saveUserClusters = (updated: DatabaseClusterItem[]) => {
    setClusters(updated)
    if (typeof window !== 'undefined') {
      localStorage.setItem(`anarva_user_databases_${userEmail}`, JSON.stringify(updated))
    }
  }

  const handleCreateDatabase = () => {
    setIsCreating(true)
    setTimeout(() => {
      const dbName = name || 'new-database'
      const newCluster: DatabaseClusterItem = {
        id: `res-db-${Date.now()}`,
        resourceId: generateARNV('DATABASE', regionId, 'proj-default', dbName),
        name: dbName,
        engine,
        engineVersion,
        status: 'AVAILABLE',
        regionId,
        environment: 'Production',
        computeUnits,
        storageGb,
        maxStorageGb,
        autoScaling,
        backupEnabled: true,
        backupRetentionDays,
        pitrEnabled,
        highAvailability,
        host: `${dbName}.anarva.cloud`,
        port: engine === 'MySQL' ? 3306 : 5432,
        dbname: dbName.replace(/-/g, '_'),
        username: 'anarva_admin',
        createdAt: new Date().toISOString(),
      }

      const updated = [newCluster, ...clusters]
      saveUserClusters(updated)

      // Record activity event
      if (typeof window !== 'undefined') {
        const actKey = `anarva_user_activities_${userEmail}`
        const existingActs = JSON.parse(localStorage.getItem(actKey) || '[]')
        const newAct = {
          id: `act-${Date.now()}`,
          action: 'RESOURCE_CREATED',
          resource: dbName,
          actor: userEmail,
          time: 'Just now',
        }
        localStorage.setItem(actKey, JSON.stringify([newAct, ...existingActs]))
      }

      setIsCreating(false)
      setIsWizardOpen(false)
      setWizardStep(1)
    }, 1200)
  }

  const handleExecuteSql = () => {
    setIsExecutingSql(true)
    setTimeout(() => {
      setIsExecutingSql(false)
      setQueryResults({
        columns: ['id', 'full_name', 'email', 'role', 'status', 'created_at'],
        rows: [
          ['usr-87a1', 'Lokesh Ashapu', 'lokeshashapu@gmail.com', 'OWNER', 'ACTIVE', '2026-08-10 21:00:00'],
        ],
        executionTimeMs: 12.4,
        rowsAffected: 1,
      })
    }, 400)
  }

  const detailTabs: TabItem[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'metrics', label: 'Metrics' },
    { id: 'sqleditor', label: 'SQL Editor' },
    { id: 'tables', label: 'Tables & Schema' },
    { id: 'connections', label: 'Connections' },
    { id: 'backups', label: 'Backups & PITR' },
    { id: 'replication', label: 'Replication' },
    { id: 'logs', label: 'Logs & Audit' },
    { id: 'security', label: 'Security' },
    { id: 'settings', label: 'Settings' },
  ]

  const handleDeleteDatabase = (clusterId: string, name: string) => {
    if (confirm(`Are you sure you want to delete database cluster '${name}'?`)) {
      const updated = clusters.filter((c) => c.id !== clusterId)
      saveUserClusters(updated)

      // Record activity event
      if (typeof window !== 'undefined') {
        const actKey = `anarva_user_activities_${userEmail}`
        const existingActs = JSON.parse(localStorage.getItem(actKey) || '[]')
        const newAct = {
          id: `act-${Date.now()}`,
          action: 'RESOURCE_DELETED',
          resource: name,
          actor: userEmail,
          time: 'Just now',
        }
        localStorage.setItem(actKey, JSON.stringify([newAct, ...existingActs]))
      }

      setSelectedCluster(null)
    }
  }

  const handleStopDatabase = (clusterId: string, name: string) => {
    const updated = clusters.map((c) => (c.id === clusterId ? { ...c, status: 'STOPPED' as ResourceStatus } : c))
    saveUserClusters(updated)
    if (selectedCluster && selectedCluster.id === clusterId) {
      setSelectedCluster({ ...selectedCluster, status: 'STOPPED' })
    }
  }

  const handleRestartDatabase = (clusterId: string, name: string) => {
    const updated = clusters.map((c) => (c.id === clusterId ? { ...c, status: 'AVAILABLE' as ResourceStatus } : c))
    saveUserClusters(updated)
    if (selectedCluster && selectedCluster.id === clusterId) {
      setSelectedCluster({ ...selectedCluster, status: 'AVAILABLE' })
    }
  }

  // DETAIL VIEW
  if (selectedCluster) {
    return (
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
          <div className="space-y-1">
            <button
              onClick={() => setSelectedCluster(null)}
              className="text-xs text-blue-400 hover:underline font-mono flex items-center gap-1 mb-2"
            >
              ← Back to Database Registry
            </button>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">{selectedCluster.name}</h1>
              <CloudStatus status={selectedCluster.status} />
            </div>
            <div className="text-xs text-slate-400 font-mono flex items-center gap-2">
              <span>{selectedCluster.engine} {selectedCluster.engineVersion}</span>
              <span>•</span>
              <span className="text-emerald-400 font-bold bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
                {selectedCluster.resourceId}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <CloudButton variant="secondary" size="sm" onClick={() => handleRestartDatabase(selectedCluster.id, selectedCluster.name)}>
              Restart
            </CloudButton>
            <CloudButton variant="outline" size="sm" onClick={() => handleStopDatabase(selectedCluster.id, selectedCluster.name)}>
              Stop
            </CloudButton>
            <CloudButton variant="danger" size="sm" onClick={() => handleDeleteDatabase(selectedCluster.id, selectedCluster.name)}>
              Delete
            </CloudButton>
          </div>
        </div>

        {/* Detail Tabs */}
        <CloudTabs tabs={detailTabs} activeTab={activeTab} onChange={setActiveTab} />

        {/* Tab Content */}
        <div className="space-y-6">
          {activeTab === 'overview' && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <CloudCard title="Database Cluster Configuration">
                <div className="space-y-3 text-xs font-mono">
                  <div className="flex justify-between py-1 border-b border-slate-800">
                    <span className="text-slate-400">Engine:</span>
                    <span className="text-white font-bold">{selectedCluster.engine} {selectedCluster.engineVersion}</span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-slate-800">
                    <span className="text-slate-400">Compute Capacity:</span>
                    <span className="text-blue-400 font-bold">{selectedCluster.computeUnits} ACUs</span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-slate-800">
                    <span className="text-slate-400">Allocated Storage:</span>
                    <span className="text-white font-bold">{selectedCluster.storageGb} GB / Max {selectedCluster.maxStorageGb} GB</span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-slate-800">
                    <span className="text-slate-400">High Availability (Multi-AZ):</span>
                    <span className={selectedCluster.highAvailability ? 'text-emerald-400 font-bold' : 'text-slate-400'}>
                      {selectedCluster.highAvailability ? 'ENABLED' : 'DISABLED'}
                    </span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-slate-800">
                    <span className="text-slate-400">Automated Backups:</span>
                    <span className="text-emerald-400 font-bold">{selectedCluster.backupRetentionDays} Days Retention (PITR)</span>
                  </div>
                  <div className="flex justify-between py-1">
                    <span className="text-slate-400">Region:</span>
                    <span className="text-emerald-400 font-bold">{selectedCluster.regionId}</span>
                  </div>
                </div>
              </CloudCard>

              <CloudCard title="Connection & Access">
                <div className="space-y-3 text-xs font-mono">
                  <div className="space-y-1">
                    <span className="text-slate-400 font-sans">Host Endpoint:</span>
                    <div className="p-2.5 bg-slate-950 border border-slate-800 rounded-xl text-blue-300 font-bold">
                      {selectedCluster.host}
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <span className="text-slate-400 font-sans">Port:</span>
                      <div className="p-2 bg-slate-950 border border-slate-800 rounded-lg text-white font-bold">{selectedCluster.port}</div>
                    </div>
                    <div>
                      <span className="text-slate-400 font-sans">Database Name:</span>
                      <div className="p-2 bg-slate-950 border border-slate-800 rounded-lg text-white font-bold">{selectedCluster.dbname}</div>
                    </div>
                  </div>
                </div>
              </CloudCard>
            </div>
          )}

          {activeTab === 'connections' && (
            <CloudCard title="Database Credentials & Connection String">
              <div className="space-y-4 text-xs font-mono">
                <div className="space-y-1">
                  <label className="text-slate-400 font-sans">Standard Connection URI:</label>
                  <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl text-emerald-400 font-bold flex items-center justify-between">
                    <span className="truncate">
                      {selectedCluster.engine.toLowerCase()}://{selectedCluster.username}:{showPassword ? 'AnarvaSecret123!' : '••••••••'}@{selectedCluster.host}:{selectedCluster.port}/{selectedCluster.dbname}
                    </span>
                    <button
                      onClick={() => setShowPassword(!showPassword)}
                      className="px-2 py-1 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded text-[10px] ml-2 shrink-0"
                    >
                      {showPassword ? 'Hide Password' : 'Show Password'}
                    </button>
                  </div>
                </div>
              </div>
            </CloudCard>
          )}

          {activeTab === 'sqleditor' && (
            <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-bold text-white">Enterprise SQL IDE Workspace</h3>
                <CloudButton variant="primary" size="sm" isLoading={isExecutingSql} onClick={handleExecuteSql}>
                  Execute Query
                </CloudButton>
              </div>

              <textarea
                value={sqlQuery}
                onChange={(e) => setSqlQuery(e.target.value)}
                rows={5}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 font-mono text-xs text-blue-300 focus:outline-none focus:border-blue-500"
              ></textarea>

              {queryResults && (
                <div className="space-y-2">
                  <div className="text-[11px] font-mono text-emerald-400">
                    ✓ Query executed in {queryResults.executionTimeMs} ms ({queryResults.rowsAffected} row returned)
                  </div>
                  <div className="overflow-x-auto border border-slate-800 rounded-xl">
                    <table className="w-full text-left font-mono text-xs divide-y divide-slate-800">
                      <thead className="bg-slate-950 text-slate-400 font-bold text-[10px]">
                        <tr>
                          {queryResults.columns.map((c: string) => (
                            <th key={c} className="p-3">{c}</th>
                          ))}
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-800 bg-slate-900">
                        {queryResults.rows.map((r: any[], idx: number) => (
                          <tr key={idx}>
                            {r.map((val, cIdx) => (
                              <td key={cIdx} className="p-3 text-slate-200">{val}</td>
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

          {activeTab === 'replication' && (
            <div className="p-8 text-center bg-slate-900 border border-slate-800 rounded-2xl space-y-2">
              <div className="text-xs font-mono text-amber-400 font-bold uppercase">Provider Not Configured</div>
              <div className="text-xs text-slate-400">
                Multi-region streaming replication will appear once bare-metal replication drivers are connected.
              </div>
            </div>
          )}

          {activeTab !== 'overview' && activeTab !== 'connections' && activeTab !== 'sqleditor' && activeTab !== 'replication' && (
            <div className="p-8 text-center bg-slate-900 border border-slate-800 rounded-2xl text-slate-400 text-xs">
              Module controls active for {selectedCluster.name}.
            </div>
          )}
        </div>
      </div>
    )
  }

  // MULTI-STEP CREATION WIZARD
  if (isWizardOpen) {
    return (
      <div className="max-w-2xl mx-auto py-8 space-y-6">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-6 shadow-2xl">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <h2 className="text-base font-bold text-white">Provision Managed Database Cluster</h2>
            <span className="text-xs font-mono text-blue-400">Step {wizardStep} of 7</span>
          </div>

          {wizardStep === 1 && (
            <div className="space-y-4 text-xs">
              <label className="block text-slate-300 font-semibold">Select Database Engine</label>
              <div className="grid grid-cols-2 gap-4">
                <div
                  onClick={() => { setEngine('PostgreSQL'); setEngineVersion('17.2'); }}
                  className={`p-4 rounded-xl border cursor-pointer space-y-2 ${engine === 'PostgreSQL' ? 'bg-blue-600/10 border-blue-500/50 text-white font-bold' : 'bg-slate-950 border-slate-800 text-slate-400'}`}
                >
                  <div className="text-sm">PostgreSQL</div>
                  <div className="text-[11px] font-mono text-slate-400">Object-relational SQL Engine</div>
                </div>
                <div
                  onClick={() => { setEngine('MySQL'); setEngineVersion('8.4.0'); }}
                  className={`p-4 rounded-xl border cursor-pointer space-y-2 ${engine === 'MySQL' ? 'bg-blue-600/10 border-blue-500/50 text-white font-bold' : 'bg-slate-950 border-slate-800 text-slate-400'}`}
                >
                  <div className="text-sm">MySQL</div>
                  <div className="text-[11px] font-mono text-slate-400">High-concurrency Relational Engine</div>
                </div>
              </div>
            </div>
          )}

          {wizardStep === 2 && (
            <div className="space-y-4 text-xs">
              <label className="block text-slate-300 font-semibold">Engine Version</label>
              <select
                value={engineVersion}
                onChange={(e) => setEngineVersion(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none"
              >
                {engine === 'PostgreSQL' ? (
                  <>
                    <option value="17.2">PostgreSQL 17.2 (Latest Enterprise Release)</option>
                    <option value="16.4">PostgreSQL 16.4</option>
                    <option value="15.8">PostgreSQL 15.8</option>
                  </>
                ) : (
                  <>
                    <option value="8.4.0">MySQL 8.4.0 (LTS)</option>
                    <option value="8.0.36">MySQL 8.0.36</option>
                  </>
                )}
              </select>
            </div>
          )}

          {wizardStep === 3 && (
            <div className="space-y-4 text-xs">
              <div className="space-y-1">
                <label className="block text-slate-300 font-semibold">Cluster Identifier Name</label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. prod-db-cluster"
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="block text-slate-300 font-semibold">Compute Units (ACUs)</label>
                  <input
                    type="number"
                    value={computeUnits}
                    onChange={(e) => setComputeUnits(Number(e.target.value))}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none font-mono"
                  />
                </div>
                <div className="space-y-1">
                  <label className="block text-slate-300 font-semibold">Allocated Storage (GB)</label>
                  <input
                    type="number"
                    value={storageGb}
                    onChange={(e) => setStorageGb(Number(e.target.value))}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none font-mono"
                  />
                </div>
              </div>
            </div>
          )}

          {wizardStep === 4 && (
            <div className="space-y-4 text-xs">
              <div className="space-y-1">
                <label className="block text-slate-300 font-semibold">Deployment Region</label>
                <select
                  value={regionId}
                  onChange={(e) => setRegionId(e.target.value as RegionId)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none"
                >
                  <option value="ap-hyderabad-1">Asia Pacific — Hyderabad (ap-hyderabad-1)</option>
                  <option value="ap-mumbai-1">Asia Pacific — Mumbai (ap-mumbai-1)</option>
                  <option value="ap-singapore-1">Asia Pacific — Singapore (ap-singapore-1)</option>
                  <option value="us-east-1">US East — N. Virginia (us-east-1)</option>
                  <option value="eu-west-1">Europe West — Frankfurt (eu-west-1)</option>
                </select>
              </div>
            </div>
          )}

          {wizardStep >= 5 && wizardStep < 7 && (
            <div className="space-y-4 text-xs font-mono">
              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                <div className="text-slate-400">Automated Backups: <strong className="text-emerald-400">7 Days Retention (PITR Enabled)</strong></div>
                <div className="text-slate-400">VPC Encryption: <strong className="text-blue-400">TLS 1.3 Active</strong></div>
              </div>
            </div>
          )}

          {wizardStep === 7 && (
            <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2 text-xs font-mono">
              <div>Engine: <strong className="text-white">{engine} {engineVersion}</strong></div>
              <div>Cluster Name: <strong className="text-white">{name || 'new-database'}</strong></div>
              <div>Compute & Storage: <strong className="text-blue-400">{computeUnits} ACUs / {storageGb} GB Storage</strong></div>
              <div>Region: <strong className="text-emerald-400">{regionId}</strong></div>
            </div>
          )}

          {/* Wizard Controls */}
          <div className="pt-4 border-t border-slate-800 flex justify-between">
            <CloudButton variant="outline" size="sm" onClick={() => setIsWizardOpen(false)}>
              Cancel
            </CloudButton>
            <div className="flex gap-2">
              {wizardStep > 1 && (
                <CloudButton variant="secondary" size="sm" onClick={() => setWizardStep(wizardStep - 1)}>
                  Back
                </CloudButton>
              )}
              {wizardStep < 7 ? (
                <CloudButton variant="primary" size="sm" onClick={() => setWizardStep(wizardStep + 1)}>
                  Next Step
                </CloudButton>
              ) : (
                <CloudButton variant="primary" size="sm" isLoading={isCreating} onClick={handleCreateDatabase}>
                  Provision Cluster
                </CloudButton>
              )}
            </div>
          </div>
        </div>
      </div>
    )
  }

  // LIST VIEW
  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Managed Databases</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">
            Enterprise PostgreSQL & MySQL database clusters with automated backups, connection pooling, and multi-region replication.
          </p>
        </div>

        <CloudButton variant="primary" size="sm" onClick={() => setIsWizardOpen(true)}>
          + Create Database Cluster
        </CloudButton>
      </div>

      {/* Cluster Cards List */}
      <div className="grid grid-cols-1 gap-4">
        {clusters.map((c) => (
          <div
            key={c.id}
            onClick={() => setSelectedCluster(c)}
            className="bg-slate-900 border border-slate-800 hover:border-slate-700 rounded-2xl p-5 cursor-pointer transition shadow-xl space-y-4"
          >
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div className="space-y-1">
                <div className="flex items-center gap-3">
                  <span className="font-bold text-white text-base">{c.name}</span>
                  <CloudStatus status={c.status} />
                </div>
                <div className="text-xs text-slate-400 font-mono">
                  {c.engine} {c.engineVersion} • {c.regionId} • {c.computeUnits} ACUs • {c.storageGb} GB
                </div>
              </div>

              <div className="flex items-center gap-2">
                <button className="px-3 py-1.5 bg-blue-600/10 text-blue-400 border border-blue-500/20 rounded-xl text-xs font-semibold hover:bg-blue-600/20 transition">
                  Manage Console
                </button>
              </div>
            </div>

            <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl flex items-center justify-between font-mono text-xs text-slate-400">
              <span className="truncate">Host: {c.host}</span>
              <span className="text-emerald-400 font-bold text-[11px] shrink-0">TLS 1.3 Protected</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
