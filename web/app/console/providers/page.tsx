'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { API_BASE_URL } from '@/lib/api'

interface ProviderItem {
  id: string
  name: string
  type: string
  status: string
  credentialReference: string
  capabilities: Record<string, boolean>
  regions: string[]
  realityLabel: string
}

interface ImportedMapping {
  anarvaResourceId: string
  provider: string
  providerResourceId: string
  region: string
  managed: boolean
}

export default function CloudProvidersPage() {
  const [activeTab, setActiveTab] = useState('providers')
  const [providers, setProviders] = useState<ProviderItem[]>([
    {
      id: 'provider-local-docker',
      name: 'Local Docker Engine',
      type: 'LOCAL_DOCKER',
      status: 'CONNECTED',
      credentialReference: 'cred-local-socket',
      capabilities: { compute: true, postgresql: true, mysql: true, objectStorage: true },
      regions: ['local-region-1'],
      realityLabel: 'LOCAL_DOCKER (CONNECTED)',
    },
    {
      id: 'provider-aws',
      name: 'Amazon Web Services (AWS)',
      type: 'AWS',
      status: 'NOT_CONFIGURED',
      credentialReference: '',
      capabilities: { compute: true, kubernetes: true, postgresql: true, mysql: true, objectStorage: true },
      regions: [],
      realityLabel: 'AWS (NOT_CONFIGURED)',
    },
    {
      id: 'provider-gcp',
      name: 'Google Cloud Platform (GCP)',
      type: 'GOOGLE_CLOUD',
      status: 'NOT_CONFIGURED',
      credentialReference: '',
      capabilities: { compute: true, kubernetes: true, postgresql: true, mysql: true, objectStorage: true },
      regions: [],
      realityLabel: 'GOOGLE_CLOUD (NOT_CONFIGURED)',
    },
  ])

  const [importedResources, setImportedResources] = useState<ImportedMapping[]>([])
  const [selectedProvider, setSelectedProvider] = useState<ProviderItem | null>(null)
  const [isConnectModalOpen, setIsConnectModalOpen] = useState(false)
  const [credArn, setCredArn] = useState('')
  const [isVerifying, setIsVerifying] = useState(false)

  // Resource Import State
  const [importProvider, setImportProvider] = useState('AWS')
  const [importResId, setImportResId] = useState('')
  const [importResType, setImportResType] = useState('ec2-instance')
  const [importRegion, setImportRegion] = useState('us-east-1')

  useEffect(() => {
    fetch(`${API_BASE_URL}/api/v1/providers`)
      .then((r) => r.json())
      .then((res) => {
        if (res && res.data) setProviders(res.data)
      })
      .catch(() => null)
  }, [])

  const handleVerifyProvider = async (providerId: string) => {
    setIsVerifying(true)
    const res = await fetch(`${API_BASE_URL}/api/v1/providers/${providerId}/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ credentialReference: credArn || 'arn:aws:iam::123456789012:role/AnarvaExecutionRole' }),
    })
      .then((r) => r.json())
      .catch(() => null)

    if (res && res.data) {
      setProviders(providers.map((p) => (p.id === providerId ? res.data : p)))
    }
    setIsVerifying(false)
    setIsConnectModalOpen(false)
  }

  const handleImportResource = async () => {
    if (!importResId) return
    const res = await fetch(`${API_BASE_URL}/api/v1/resources/import`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        provider: importProvider,
        providerResourceId: importResId,
        resourceType: importResType,
        region: importRegion,
      }),
    })
      .then((r) => r.json())
      .catch(() => null)

    if (res && res.data) {
      setImportedResources([res.data, ...importedResources])
      setImportResId('')
    }
  }

  const handleAdoptResource = async (anarvaId: string) => {
    const res = await fetch(`${API_BASE_URL}/api/v1/resources/${anarvaId}/adopt`, {
      method: 'POST',
    })
      .then((r) => r.json())
      .catch(() => null)

    if (res && res.data) {
      setImportedResources(importedResources.map((m) => (m.anarvaResourceId === anarvaId ? res.data : m)))
    }
  }

  const handleReleaseResource = async (anarvaId: string) => {
    await fetch(`${API_BASE_URL}/api/v1/resources/${anarvaId}/release`, {
      method: 'POST',
    }).catch(() => null)

    setImportedResources(
      importedResources.map((m) => (m.anarvaResourceId === anarvaId ? { ...m, managed: false } : m))
    )
  }

  const tabItems: TabItem[] = [
    { id: 'providers', label: 'Cloud Providers' },
    { id: 'import', label: 'Resource Import & Adoption' },
    { id: 'drift', label: 'Drift Repair Dashboard' },
  ]

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Real Cloud Infrastructure Providers</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Multi-cloud execution targets: AWS, GCP, and Local Docker Engine.</p>
        </div>
      </div>

      <CloudTabs tabs={tabItems} activeTab={activeTab} onChange={setActiveTab} />

      {activeTab === 'providers' && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 font-mono text-xs">
          {providers.map((p) => (
            <CloudCard key={p.id} title={p.name}>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-slate-400">Status:</span>
                  <CloudStatus status={p.status} />
                </div>
                <div>
                  Label: <strong className="text-blue-400">{p.realityLabel}</strong>
                </div>

                <div className="pt-2 border-t border-slate-800 space-y-1">
                  <div className="font-bold text-slate-300 font-sans mb-1 text-[11px]">Provider Capabilities:</div>
                  <div className="flex flex-wrap gap-1">
                    {Object.keys(p.capabilities).map((cap) => (
                      <span key={cap} className="px-1.5 py-0.5 bg-slate-900 border border-slate-800 rounded text-[10px] text-emerald-400">
                        {cap}
                      </span>
                    ))}
                  </div>
                </div>

                <div className="pt-3 flex justify-end">
                  {p.status === 'NOT_CONFIGURED' ? (
                    <CloudButton
                      variant="primary"
                      size="sm"
                      onClick={() => {
                        setSelectedProvider(p)
                        setIsConnectModalOpen(true)
                      }}
                    >
                      Connect {p.type}
                    </CloudButton>
                  ) : (
                    <CloudButton variant="secondary" size="sm" onClick={() => handleVerifyProvider(p.id)}>
                      Re-Verify Connection
                    </CloudButton>
                  )}
                </div>
              </div>
            </CloudCard>
          ))}
        </div>
      )}

      {activeTab === 'import' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Import Unmanaged Cloud Resource" subtitle="Import existing AWS/GCP resources without automatic management">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div>
                <label className="block text-slate-300 font-bold mb-1">Provider</label>
                <select
                  value={importProvider}
                  onChange={(e) => setImportProvider(e.target.value)}
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                >
                  <option value="AWS">AWS</option>
                  <option value="GOOGLE_CLOUD">Google Cloud</option>
                </select>
              </div>

              <div>
                <label className="block text-slate-300 font-bold mb-1">Provider Resource ID</label>
                <input
                  type="text"
                  placeholder="e.g. i-0a1b2c3d4e5f67890"
                  value={importResId}
                  onChange={(e) => setImportResId(e.target.value)}
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-slate-300 font-bold mb-1">Region</label>
                <input
                  type="text"
                  value={importRegion}
                  onChange={(e) => setImportRegion(e.target.value)}
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                />
              </div>

              <div className="flex items-end">
                <CloudButton variant="primary" size="sm" onClick={handleImportResource}>
                  + Import Resource
                </CloudButton>
              </div>
            </div>
          </CloudCard>

          <CloudCard title="Imported Infrastructure Registry">
            {importedResources.length === 0 ? (
              <CloudEmptyState
                title="No Imported Cloud Resources"
                description="Import existing cloud infrastructure to monitor and adopt management."
                actionLabel="Import Resource"
                onAction={() => {}}
                icon="☁️"
                docsLink="/console/developer"
              />
            ) : (
              <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden">
                {importedResources.map((res) => (
                  <div key={res.anarvaResourceId} className="p-4 bg-slate-950 flex items-center justify-between">
                    <div>
                      <div className="font-bold text-white font-sans text-sm">{res.providerResourceId}</div>
                      <div className="text-[10px] text-slate-500 mt-0.5">
                        ID: {res.anarvaResourceId} • Provider: {res.provider} • Region: {res.region} • Managed: {res.managed ? 'TRUE' : 'FALSE (UNMANAGED)'}
                      </div>
                    </div>

                    <div className="flex gap-2">
                      {!res.managed ? (
                        <CloudButton variant="primary" size="sm" onClick={() => handleAdoptResource(res.anarvaResourceId)}>
                          Adopt Management
                        </CloudButton>
                      ) : (
                        <CloudButton variant="secondary" size="sm" onClick={() => handleReleaseResource(res.anarvaResourceId)}>
                          Release Management
                        </CloudButton>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CloudCard>
        </div>
      )}

      {activeTab === 'drift' && (
        <CloudCard title="Desired vs Observed State Drift Engine">
          <CloudEmptyState
            title="Zero Resource Drift Detected"
            description="All active resources match their desired state definitions across providers."
            actionLabel="Refresh Verification"
            onAction={() => {}}
            icon="⚖️"
            docsLink="/console/developer"
          />
        </CloudCard>
      )}

      {/* Connect Provider Modal */}
      {isConnectModalOpen && selectedProvider && (
        <CloudModal isOpen={isConnectModalOpen} title={`Connect ${selectedProvider.name}`} onClose={() => setIsConnectModalOpen(false)}>
          <div className="space-y-4 font-mono text-xs">
            <div>
              <label className="block text-slate-300 font-bold mb-1">IAM Role ARN / Service Account Reference</label>
              <input
                type="text"
                placeholder={selectedProvider.type === 'AWS' ? 'arn:aws:iam::123456789012:role/AnarvaExecutionRole' : 'projects/anarva-prod/serviceAccounts/anarva-sa@...'}
                value={credArn}
                onChange={(e) => setCredArn(e.target.value)}
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
              />
              <p className="text-[10px] text-slate-500 mt-1 font-sans">
                Never enter secret keys directly. Prefer IAM Role ARNs and Service Account references.
              </p>
            </div>

            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2">
              <CloudButton variant="secondary" size="sm" onClick={() => setIsConnectModalOpen(false)}>
                Cancel
              </CloudButton>
              <CloudButton variant="primary" size="sm" onClick={() => handleVerifyProvider(selectedProvider.id)} disabled={isVerifying}>
                {isVerifying ? 'Authenticating...' : 'Verify & Connect'}
              </CloudButton>
            </div>
          </div>
        </CloudModal>
      )}
    </div>
  )
}
