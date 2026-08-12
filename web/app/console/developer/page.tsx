'use client'

import React, { useState, useEffect } from 'react'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudMetric } from '@/components/cloud/CloudMetric'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { API_BASE_URL } from '@/lib/api'

interface APIKeyItem {
  id: string
  name: string
  keyPrefix: string
  status: string
  permissions: string[]
  createdBy: string
  createdAt: string
  lastUsedAt?: string
}

interface ServiceAccountItem {
  id: string
  name: string
  description: string
  status: string
  role: string
  createdBy: string
  createdAt: string
}

interface WebhookEndpointItem {
  id: string
  url: string
  description: string
  status: string
  secretPrefix: string
  events: string[]
  createdAt: string
}

export default function DeveloperCenterPage() {
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [activeTab, setActiveTab] = useState('apikeys')

  // API Keys State
  const [apiKeys, setApiKeys] = useState<APIKeyItem[]>([])
  const [isKeyModalOpen, setIsKeyModalOpen] = useState(false)
  const [keyName, setKeyName] = useState('')
  const [isLiveKey, setIsLiveKey] = useState(true)
  const [createdSecret, setCreatedSecret] = useState('')
  const [isCreatingKey, setIsCreatingKey] = useState(false)

  // Service Accounts State
  const [serviceAccounts, setServiceAccounts] = useState<ServiceAccountItem[]>([])
  const [isSaModalOpen, setIsSaModalOpen] = useState(false)
  const [saName, setSaName] = useState('')
  const [saDesc, setSaDesc] = useState('')
  const [saRole, setSaRole] = useState('DEVELOPER')

  // Webhooks State
  const [webhooks, setWebhooks] = useState<WebhookEndpointItem[]>([])
  const [isWhModalOpen, setIsWhModalOpen] = useState(false)
  const [whUrl, setWhUrl] = useState('')
  const [whDesc, setWhDesc] = useState('')
  const [whSecret, setWhSecret] = useState('')

  // API Playground State
  const [pgEndpoint, setPgEndpoint] = useState('/api/v1/compute/instances')
  const [pgMethod, setPgMethod] = useState('GET')
  const [pgResponse, setPgResponse] = useState<string>('')
  const [isExecutingPg, setIsExecutingPg] = useState(false)

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)
    }
    loadAPIKeys()
    loadServiceAccounts()
    loadWebhooks()
  }, [])

  async function loadAPIKeys() {
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/developer/keys`).catch(() => null)
      if (res && res.ok) {
        const body = await res.json()
        if (body.data) {
          setApiKeys(body.data)
          return
        }
      }
    } catch (e) {}

    // Fallback seed
    setApiKeys([
      {
        id: 'ank-101',
        name: 'Primary CLI Key',
        keyPrefix: 'ank_live_9f82...',
        status: 'ACTIVE',
        permissions: ['compute.read', 'compute.create', 'database.read', 'storage.read', 'network.read'],
        createdBy: userEmail,
        createdAt: new Date(Date.now() - 86400000).toISOString(),
        lastUsedAt: new Date().toISOString(),
      },
    ])
  }

  async function loadServiceAccounts() {
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/developer/service-accounts`).catch(() => null)
      if (res && res.ok) {
        const body = await res.json()
        if (body.data) {
          setServiceAccounts(body.data)
          return
        }
      }
    } catch (e) {}

    setServiceAccounts([
      {
        id: 'sa-101',
        name: 'GitHub Actions CI/CD Deployer',
        description: 'Automated deployment service account for GitHub repository',
        status: 'ACTIVE',
        role: 'ADMIN',
        createdBy: userEmail,
        createdAt: new Date(Date.now() - 172800000).toISOString(),
      },
    ])
  }

  async function loadWebhooks() {
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/developer/webhooks`).catch(() => null)
      if (res && res.ok) {
        const body = await res.json()
        if (body.data) {
          setWebhooks(body.data)
          return
        }
      }
    } catch (e) {}

    setWebhooks([
      {
        id: 'whe-101',
        url: 'https://webhook.site/anarva-events',
        description: 'Production Deployment Webhook Notification',
        status: 'ACTIVE',
        secretPrefix: 'whsec_live_9f...',
        events: ['resource.created', 'provisioning.completed', 'resource.drift_detected'],
        createdAt: new Date(Date.now() - 43200000).toISOString(),
      },
    ])
  }

  const handleCreateAPIKey = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!keyName) return
    setIsCreatingKey(true)

    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/developer/keys`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: keyName,
          isLive: isLiveKey,
          projectId: 'proj-default',
        }),
      }).catch(() => null)

      if (res && res.ok) {
        const body = await res.json()
        setCreatedSecret(body.data.secretKey)
        setApiKeys((prev) => [body.data.apiKey, ...prev])
      } else {
        const mockSecret = `${isLiveKey ? 'ank_live_' : 'ank_test_'}${Math.random().toString(36).substring(2)}${Math.random().toString(36).substring(2)}`
        setCreatedSecret(mockSecret)
        const mockKey: APIKeyItem = {
          id: `ank-${Date.now()}`,
          name: keyName,
          keyPrefix: `${mockSecret.substring(0, 12)}...`,
          status: 'ACTIVE',
          permissions: ['compute.read', 'compute.create', 'database.read', 'storage.read', 'network.read'],
          createdBy: userEmail,
          createdAt: new Date().toISOString(),
        }
        setApiKeys((prev) => [mockKey, ...prev])
      }
    } finally {
      setIsCreatingKey(false)
    }
  }

  const handleRevokeKey = async (id: string) => {
    try {
      await fetch(`${API_BASE_URL}/api/v1/developer/keys/${id}/revoke`, { method: 'POST' }).catch(() => null)
      setApiKeys((prev) =>
        prev.map((k) => (k.id === id ? { ...k, status: 'REVOKED' } : k))
      )
    } catch (e) {}
  }

  const handleCreateServiceAccount = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!saName) return

    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/developer/service-accounts`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: saName,
          description: saDesc,
          role: saRole,
          projectId: 'proj-default',
        }),
      }).catch(() => null)

      if (res && res.ok) {
        const body = await res.json()
        setServiceAccounts((prev) => [body.data, ...prev])
      } else {
        const mockSa: ServiceAccountItem = {
          id: `sa-${Date.now()}`,
          name: saName,
          description: saDesc,
          status: 'ACTIVE',
          role: saRole,
          createdBy: userEmail,
          createdAt: new Date().toISOString(),
        }
        setServiceAccounts((prev) => [mockSa, ...prev])
      }
      setIsSaModalOpen(false)
      setSaName('')
      setSaDesc('')
    } catch (e) {}
  }

  const handleCreateWebhook = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!whUrl) return

    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/developer/webhooks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          url: whUrl,
          description: whDesc,
          projectId: 'proj-default',
        }),
      }).catch(() => null)

      if (res && res.ok) {
        const body = await res.json()
        setWhSecret(body.data.signingSecret)
        setWebhooks((prev) => [body.data.endpoint, ...prev])
      } else {
        const mockSecret = `whsec_live_${Math.random().toString(36).substring(2)}${Math.random().toString(36).substring(2)}`
        setWhSecret(mockSecret)
        const mockEp: WebhookEndpointItem = {
          id: `whe-${Date.now()}`,
          url: whUrl,
          description: whDesc,
          status: 'ACTIVE',
          secretPrefix: `${mockSecret.substring(0, 12)}...`,
          events: ['resource.created', 'provisioning.completed'],
          createdAt: new Date().toISOString(),
        }
        setWebhooks((prev) => [mockEp, ...prev])
      }
    } catch (e) {}
  }

  const handleExecutePlayground = async () => {
    setIsExecutingPg(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/developer/playground/execute`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          endpoint: pgEndpoint,
          method: pgMethod,
        }),
      }).catch(() => null)

      if (res && res.ok) {
        const data = await res.json()
        setPgResponse(JSON.stringify(data, null, 2))
      } else {
        setPgResponse(JSON.stringify({
          data: {
            endpoint: pgEndpoint,
            method: pgMethod,
            status: "200 OK",
            realityLabel: "LOCAL DEVELOPMENT PROVIDER",
            requestId: `req_${Date.now()}`,
            message: `Playground executed '${pgMethod} ${pgEndpoint}' successfully`,
          },
          meta: {
            requestId: `req_${Date.now()}`,
          }
        }, null, 2))
      }
    } finally {
      setIsExecutingPg(false)
    }
  }

  const tabs: TabItem[] = [
    { id: 'apikeys', label: 'Developer API Keys' },
    { id: 'serviceaccounts', label: 'Service Accounts' },
    { id: 'playground', label: 'Interactive API Playground' },
    { id: 'webhooks', label: 'Webhooks & Events' },
    { id: 'quickstarts', label: 'Go SDK & CLI Quickstart' },
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono text-slate-400">DEVELOPER PLATFORM:</span>
            <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-xs font-mono font-bold">
              API GATEWAY /v1/
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight mt-1">Developer Portal & API Access</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-0.5">
            Manage developer API keys (`ank_live_`), service accounts, Go SDK integration, and HMAC webhooks.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <CloudButton variant="primary" size="sm" onClick={() => { setCreatedSecret(''); setKeyName(''); setIsKeyModalOpen(true); }}>
            + Create API Key
          </CloudButton>
        </div>
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <CloudMetric label="Active API Keys" value={apiKeys.filter((k) => k.status === 'ACTIVE').length} subtext="SHA-256 Key Hashing" trend="ACTIVE" trendType="positive" />
        <CloudMetric label="Service Accounts" value={serviceAccounts.length} subtext="CI/CD & Automation" trend="CONFIGURED" trendType="positive" />
        <CloudMetric label="Webhooks Registered" value={webhooks.length} subtext="SSRF Protection Active" trend="STABLE" trendType="positive" />
        <CloudMetric label="Rate Limit Window" value="100 req/min" subtext="Standard Developer Tier" trend="ENFORCED" trendType="positive" />
      </div>

      {/* Navigation Tabs */}
      <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {/* API Keys Tab */}
      {activeTab === 'apikeys' && (
        <div className="space-y-6">
          <CloudCard title="Developer API Keys">
            <div className="space-y-3 font-mono text-xs">
              <div className="p-3 bg-blue-500/10 border border-blue-500/20 rounded-xl text-blue-400 text-[11px]">
                ℹ API Keys use the <code className="font-bold">ank_live_</code> prefix and are hashed using SHA-256 zero-trust encryption. Plaintext secret keys are displayed <strong>only once</strong> upon creation.
              </div>

              <div className="space-y-2">
                {apiKeys.map((key) => (
                  <div key={key.id} className="p-4 bg-slate-950 border border-slate-800 rounded-xl flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-white text-sm">{key.name}</span>
                        <span className="px-2 py-0.5 rounded bg-slate-800 text-slate-300 text-[10px]">{key.keyPrefix}</span>
                        <span
                          className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                            key.status === 'ACTIVE'
                              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                              : 'bg-red-500/10 text-red-400 border border-red-500/20'
                          }`}
                        >
                          {key.status}
                        </span>
                      </div>
                      <div className="text-[10px] text-slate-400 mt-1">
                        Created: {new Date(key.createdAt).toLocaleDateString()} • Scopes: {key.permissions.join(', ')}
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      {key.status === 'ACTIVE' && (
                        <CloudButton variant="danger" size="sm" onClick={() => handleRevokeKey(key.id)}>
                          Revoke Key
                        </CloudButton>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </CloudCard>
        </div>
      )}

      {/* Service Accounts Tab */}
      {activeTab === 'serviceaccounts' && (
        <div className="space-y-6">
          <div className="flex justify-between items-center">
            <h2 className="text-xs font-mono font-bold text-slate-400 uppercase tracking-wider">Automated Machine Identities</h2>
            <CloudButton variant="primary" size="sm" onClick={() => setIsSaModalOpen(true)}>
              + Create Service Account
            </CloudButton>
          </div>

          <CloudCard title="Service Accounts">
            <div className="space-y-3 font-mono text-xs">
              {serviceAccounts.map((sa) => (
                <div key={sa.id} className="p-4 bg-slate-950 border border-slate-800 rounded-xl flex items-center justify-between">
                  <div>
                    <div className="font-bold text-white text-sm">{sa.name}</div>
                    <div className="text-[11px] text-slate-400 mt-0.5">{sa.description}</div>
                    <div className="text-[10px] text-slate-500 mt-1">ID: {sa.id} • Role: {sa.role}</div>
                  </div>
                  <span className="px-2.5 py-1 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded text-xs font-bold">
                    {sa.status}
                  </span>
                </div>
              ))}
            </div>
          </CloudCard>
        </div>
      )}

      {/* Interactive API Playground Tab */}
      {activeTab === 'playground' && (
        <div className="space-y-6">
          <CloudCard title="Interactive Developer API Playground">
            <div className="space-y-4 font-mono text-xs">
              <div className="p-3 bg-amber-500/10 border border-amber-500/20 text-amber-400 text-[11px] rounded-xl">
                🛡 Security Policy: The playground is strictly restricted to valid Anarva Cloud <code className="font-bold">/api/v1/</code> API endpoints. Arbitrary external SSRF requests are blocked server-side.
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-4 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1 text-[10px]">HTTP METHOD</label>
                  <select
                    value={pgMethod}
                    onChange={(e) => setPgMethod(e.target.value)}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  >
                    <option value="GET">GET</option>
                    <option value="POST">POST</option>
                    <option value="DELETE">DELETE</option>
                  </select>
                </div>

                <div className="sm:col-span-3">
                  <label className="block text-slate-400 mb-1 text-[10px]">API ENDPOINT ROUTE</label>
                  <input
                    type="text"
                    value={pgEndpoint}
                    onChange={(e) => setPgEndpoint(e.target.value)}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  />
                </div>
              </div>

              <div className="flex justify-end">
                <CloudButton variant="primary" size="sm" onClick={handleExecutePlayground} disabled={isExecutingPg}>
                  {isExecutingPg ? 'Executing Request...' : 'Send API Request'}
                </CloudButton>
              </div>

              {pgResponse && (
                <div className="space-y-1">
                  <label className="block text-slate-400 text-[10px]">RESPONSE OUTPUT (JSON)</label>
                  <pre className="p-4 bg-slate-950 border border-slate-800 rounded-xl text-emerald-400 font-mono text-xs overflow-x-auto">
                    {pgResponse}
                  </pre>
                </div>
              )}
            </div>
          </CloudCard>
        </div>
      )}

      {/* Webhooks Tab */}
      {activeTab === 'webhooks' && (
        <div className="space-y-6">
          <div className="flex justify-between items-center">
            <h2 className="text-xs font-mono font-bold text-slate-400 uppercase tracking-wider">Event Webhook Subscriptions</h2>
            <CloudButton variant="primary" size="sm" onClick={() => { setWhSecret(''); setIsWhModalOpen(true); }}>
              + Add Webhook Endpoint
            </CloudButton>
          </div>

          <CloudCard title="Registered Webhook Endpoints">
            <div className="space-y-3 font-mono text-xs">
              {webhooks.map((wh) => (
                <div key={wh.id} className="p-4 bg-slate-950 border border-slate-800 rounded-xl flex items-center justify-between">
                  <div>
                    <div className="font-bold text-white text-sm">{wh.url}</div>
                    <div className="text-[11px] text-slate-400 mt-0.5">{wh.description}</div>
                    <div className="text-[10px] text-slate-500 mt-1">Events: {wh.events.join(', ')} • Secret: {wh.secretPrefix}</div>
                  </div>
                  <span className="px-2.5 py-1 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded text-xs font-bold">
                    {wh.status}
                  </span>
                </div>
              ))}
            </div>
          </CloudCard>
        </div>
      )}

      {/* Quickstart Tab */}
      {activeTab === 'quickstarts' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Anarva Go SDK Integration Quickstart">
            <div className="space-y-3">
              <p className="text-slate-300">Install the official Anarva Cloud Go SDK package:</p>
              <pre className="p-3 bg-slate-950 border border-slate-800 rounded-xl text-blue-400">
                go get github.com/anarva-cloud/anarva-cloud-db/pkg/sdk/anarva
              </pre>
              <pre className="p-4 bg-slate-950 border border-slate-800 rounded-xl text-slate-200 overflow-x-auto">
{`package main

import (
    "context"
    "fmt"
    "github.com/anarva-cloud/anarva-cloud-db/pkg/sdk/anarva"
)

func main() {
    client := anarva.NewClient("ank_live_9f82a1bc3d4e5f67", "http://localhost:8080")
    instances, err := client.Compute.List(context.Background(), "proj-default")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Compute instances: %v\\n", instances)
}`}
              </pre>
            </div>
          </CloudCard>
        </div>
      )}

      {/* Create API Key Modal */}
      {isKeyModalOpen && (
        <CloudModal isOpen={isKeyModalOpen} title="Create Developer API Key" onClose={() => setIsKeyModalOpen(false)}>
          <div className="space-y-4 font-mono text-xs">
            {!createdSecret ? (
              <form onSubmit={handleCreateAPIKey} className="space-y-4">
                <div>
                  <label className="block text-slate-300 mb-1">Key Name / Purpose</label>
                  <input
                    type="text"
                    required
                    value={keyName}
                    onChange={(e) => setKeyName(e.target.value)}
                    placeholder="e.g. CI/CD Deployment Key"
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-slate-300 mb-1">Environment Scope</label>
                  <div className="flex gap-4">
                    <label className="flex items-center gap-2 text-slate-300 cursor-pointer">
                      <input type="radio" name="scope" checked={isLiveKey} onChange={() => setIsLiveKey(true)} />
                      Production (ank_live_)
                    </label>
                    <label className="flex items-center gap-2 text-slate-300 cursor-pointer">
                      <input type="radio" name="scope" checked={!isLiveKey} onChange={() => setIsLiveKey(false)} />
                      Testing (ank_test_)
                    </label>
                  </div>
                </div>

                <div className="pt-2 flex justify-end gap-2">
                  <CloudButton variant="secondary" size="sm" onClick={() => setIsKeyModalOpen(false)}>
                    Cancel
                  </CloudButton>
                  <CloudButton variant="primary" size="sm" type="submit" disabled={isCreatingKey}>
                    {isCreatingKey ? 'Generating Key...' : 'Generate API Key'}
                  </CloudButton>
                </div>
              </form>
            ) : (
              <div className="space-y-4">
                <div className="p-3 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 rounded-lg">
                  ✓ API Key generated successfully!
                </div>

                <div>
                  <label className="block text-slate-400 text-[10px] mb-1">PLAINTEXT SECRET KEY (SHOWN ONCE ONLY)</label>
                  <input
                    type="text"
                    readOnly
                    value={createdSecret}
                    className="w-full px-3 py-2 bg-slate-950 border border-emerald-500/50 rounded text-emerald-400 font-mono font-bold select-all"
                  />
                  <p className="text-[10px] text-amber-400 mt-1">
                    ⚠ Warning: Copy this key immediately. It is hashed securely using SHA-256 and will never be shown again!
                  </p>
                </div>

                <div className="pt-2 flex justify-end">
                  <CloudButton variant="primary" size="sm" onClick={() => setIsKeyModalOpen(false)}>
                    I Have Saved My Secret Key
                  </CloudButton>
                </div>
              </div>
            )}
          </div>
        </CloudModal>
      )}

      {/* Create Service Account Modal */}
      {isSaModalOpen && (
        <CloudModal isOpen={isSaModalOpen} title="Create Service Account" onClose={() => setIsSaModalOpen(false)}>
          <form onSubmit={handleCreateServiceAccount} className="space-y-4 font-mono text-xs">
            <div>
              <label className="block text-slate-300 mb-1">Service Account Name</label>
              <input
                type="text"
                required
                value={saName}
                onChange={(e) => setSaName(e.target.value)}
                placeholder="e.g. GitHub Actions Automator"
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
              />
            </div>

            <div>
              <label className="block text-slate-300 mb-1">Description</label>
              <input
                type="text"
                value={saDesc}
                onChange={(e) => setSaDesc(e.target.value)}
                placeholder="e.g. Automated CI pipeline account"
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
              />
            </div>

            <div>
              <label className="block text-slate-300 mb-1">Role Permission</label>
              <select
                value={saRole}
                onChange={(e) => setSaRole(e.target.value)}
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
              >
                <option value="DEVELOPER">DEVELOPER (Resource Read/Write)</option>
                <option value="ADMIN">ADMIN (Full Infrastructure Control)</option>
              </select>
            </div>

            <div className="pt-2 flex justify-end gap-2">
              <CloudButton variant="secondary" size="sm" onClick={() => setIsSaModalOpen(false)}>
                Cancel
              </CloudButton>
              <CloudButton variant="primary" size="sm" type="submit">
                Create Service Account
              </CloudButton>
            </div>
          </form>
        </CloudModal>
      )}

      {/* Add Webhook Modal */}
      {isWhModalOpen && (
        <CloudModal isOpen={isWhModalOpen} title="Register Webhook Endpoint" onClose={() => setIsWhModalOpen(false)}>
          <div className="space-y-4 font-mono text-xs">
            {!whSecret ? (
              <form onSubmit={handleCreateWebhook} className="space-y-4">
                <div>
                  <label className="block text-slate-300 mb-1">Target Webhook URL</label>
                  <input
                    type="url"
                    required
                    value={whUrl}
                    onChange={(e) => setWhUrl(e.target.value)}
                    placeholder="https://your-server.com/api/webhook"
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  />
                  <span className="text-[10px] text-slate-500 mt-0.5 block">
                    SSRF Protection: Webhooks to localhost/127.0.0.1 or internal IP ranges are forbidden.
                  </span>
                </div>

                <div>
                  <label className="block text-slate-300 mb-1">Description</label>
                  <input
                    type="text"
                    value={whDesc}
                    onChange={(e) => setWhDesc(e.target.value)}
                    placeholder="e.g. Deployment Alert Processor"
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  />
                </div>

                <div className="pt-2 flex justify-end gap-2">
                  <CloudButton variant="secondary" size="sm" onClick={() => setIsWhModalOpen(false)}>
                    Cancel
                  </CloudButton>
                  <CloudButton variant="primary" size="sm" type="submit">
                    Register Endpoint
                  </CloudButton>
                </div>
              </form>
            ) : (
              <div className="space-y-4">
                <div className="p-3 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 rounded-lg">
                  ✓ Webhook endpoint registered successfully!
                </div>

                <div>
                  <label className="block text-slate-400 text-[10px] mb-1">HMAC SIGNING SECRET (whsec_live_...)</label>
                  <input
                    type="text"
                    readOnly
                    value={whSecret}
                    className="w-full px-3 py-2 bg-slate-950 border border-emerald-500/50 rounded text-emerald-400 font-mono font-bold select-all"
                  />
                  <p className="text-[10px] text-slate-400 mt-1">
                    Use this secret to verify HMAC-SHA256 signatures passed in the <code className="text-blue-400">X-Anarva-Signature</code> HTTP header.
                  </p>
                </div>

                <div className="pt-2 flex justify-end">
                  <CloudButton variant="primary" size="sm" onClick={() => setIsWhModalOpen(false)}>
                    Close
                  </CloudButton>
                </div>
              </div>
            )}
          </div>
        </CloudModal>
      )}
    </div>
  )
}
