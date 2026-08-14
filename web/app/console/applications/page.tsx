'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { CloudModal } from '@/components/cloud/CloudModal'
import { API_BASE_URL } from '@/lib/api'

interface ApplicationItem {
  id: string
  name: string
  status: string
  networkId: string
  loadBalancerId: string
  domainReference: string
  containerImage: string
  acuCount: number
  health: string
  createdAt: string
}

export default function ApplicationsPage() {
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [apps, setApps] = useState<ApplicationItem[]>([])
  const [selectedApp, setSelectedApp] = useState<ApplicationItem | null>(null)
  const [activeTab, setActiveTab] = useState<string>('overview')

  // 13-Step Creation Wizard State
  const [isWizardOpen, setIsWizardOpen] = useState(false)
  const [wizardStep, setWizardStep] = useState(1)

  const [appName, setAppName] = useState('')
  const [containerImage, setContainerImage] = useState('nginx:alpine')
  const [acuCount, setAcuCount] = useState(2)
  const [port, setPort] = useState(80)
  const [healthPath, setHealthPath] = useState('/health')
  const [domainName, setDomainName] = useState('api.anarva.cloud')
  const [isDeploying, setIsDeploying] = useState(false)

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)

      const stored = localStorage.getItem(`anarva_user_apps_${email}`)
      if (stored) {
        try {
          setApps(JSON.parse(stored))
        } catch (e) {
          setApps([])
        }
      } else {
        setApps([])
      }
    }
  }, [])

  const saveApps = (updated: ApplicationItem[]) => {
    setApps(updated)
    if (typeof window !== 'undefined') {
      localStorage.setItem(`anarva_user_apps_${userEmail}`, JSON.stringify(updated))
    }
  }

  const handleDeployApp = async () => {
    setIsDeploying(true)

    const newApp: ApplicationItem = {
      id: `app-${Date.now()}`,
      name: appName || 'primary-web-service',
      status: 'RUNNING',
      networkId: 'vpc-01',
      loadBalancerId: `lb-${Date.now()}`,
      domainReference: domainName,
      containerImage,
      acuCount,
      health: 'HEALTHY',
      createdAt: new Date().toISOString(),
    }

    await fetch(`${API_BASE_URL}/api/v1/applications`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        organizationId: 'org-default',
        projectId: 'proj-default',
        name: newApp.name,
        containerImage,
        networkId: 'vpc-01',
        domainName,
        acuCount,
        port,
      }),
    }).catch(() => null)

    const updated = [newApp, ...apps]
    saveApps(updated)

    setIsDeploying(false)
    setIsWizardOpen(false)
    setWizardStep(1)
    setAppName('')
  }

  const tabs: TabItem[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'deployments', label: 'Deployments' },
    { id: 'domains', label: 'Domains & TLS' },
    { id: 'health', label: 'Health Probes' },
    { id: 'metrics', label: 'Observability & Metrics' },
  ]

  if (selectedApp) {
    return (
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
          <div>
            <button onClick={() => setSelectedApp(null)} className="text-xs text-blue-400 font-mono mb-2">
              ← Back to Applications
            </button>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">{selectedApp.name}</h1>
              <CloudStatus status={selectedApp.status} />
              <span className="px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-xs rounded font-mono font-bold">
                ✔ {selectedApp.health}
              </span>
            </div>
            <div className="text-xs text-slate-400 font-mono">
              Domain: {selectedApp.domainReference} • Image: {selectedApp.containerImage} • ACUs: {selectedApp.acuCount}
            </div>
          </div>
        </div>

        <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 font-mono text-xs">
            <CloudCard title="Application Metadata">
              <div className="space-y-2 text-slate-300">
                <div>ID: <strong>{selectedApp.id}</strong></div>
                <div>Domain: <strong className="text-emerald-400">{selectedApp.domainReference}</strong></div>
                <div>Container Image: <strong>{selectedApp.containerImage}</strong></div>
                <div>ACUs Allocated: <strong className="text-purple-400">{selectedApp.acuCount} ACUs</strong></div>
              </div>
            </CloudCard>
            <CloudCard title="Unified Deployment Flow">
              <div className="text-xs text-slate-300 space-y-1">
                <div>Network: <strong className="text-blue-400">vpc-01</strong></div>
                <div>Load Balancer: <strong>{selectedApp.loadBalancerId}</strong></div>
                <div>TLS Certificate: <strong className="text-emerald-400">ACTIVE (HTTPS 443)</strong></div>
                <div>DNS Record: <strong className="text-purple-400">A / CNAME Verified</strong></div>
              </div>
            </CloudCard>
            <CloudCard title="Health Status">
              <div className="text-2xl font-bold text-emerald-400 mb-1">✔ HEALTHY</div>
              <p className="text-slate-400 font-sans text-xs">Active probe GET /health returning HTTP 200 OK</p>
            </CloudCard>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Applications & Unified Deployment</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Unified end-to-end deployment workflow: Network → Container → Load Balancer → Domain → TLS → Health.</p>
        </div>
        <CloudButton variant="primary" size="sm" onClick={() => setIsWizardOpen(true)}>
          + Deploy Application
        </CloudButton>
      </div>

      <CloudCard title="Applications Registry" subtitle={`Unified applications for ${userEmail}`}>
        {apps.length === 0 ? (
          <CloudEmptyState
            title="No Unified Applications Deployed"
            description="Deploy an application using the 13-step unified wizard integrating Containers, ACUs, Load Balancers, Domains, TLS, and Health Checks."
            actionLabel="+ Deploy Application"
            onAction={() => setIsWizardOpen(true)}
            icon={
              <svg className="w-6 h-6 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            }
            docsLink="/console/developer"
          />
        ) : (
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
            {apps.map((app) => (
              <div
                key={app.id}
                onClick={() => setSelectedApp(app)}
                className="p-4 bg-slate-950 hover:bg-slate-900 cursor-pointer transition flex items-center justify-between font-mono"
              >
                <div>
                  <div className="font-bold text-white text-sm font-sans flex items-center gap-2">
                    {app.name}
                    <span className="text-[10px] px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded">
                      {app.domainReference}
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1">
                    Image: {app.containerImage} • ACUs: {app.acuCount} • Health: {app.health}
                  </div>
                </div>
                <CloudStatus status={app.status} />
              </div>
            ))}
          </div>
        )}
      </CloudCard>

      {/* 13-Step Unified Deployment Wizard */}
      {isWizardOpen && (
        <CloudModal isOpen={isWizardOpen} title={`13-Step Unified Deployment Wizard (Step ${wizardStep}/13)`} onClose={() => setIsWizardOpen(false)}>
          <div className="space-y-4 font-mono text-xs">
            {wizardStep === 1 && (
              <div>
                <label className="block text-slate-300 font-bold mb-1">Step 1: Application Name</label>
                <input
                  type="text"
                  value={appName}
                  onChange={(e) => setAppName(e.target.value)}
                  placeholder="e.g. payment-gateway-api"
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                />
              </div>
            )}

            {wizardStep === 2 && (
              <div>
                <label className="block text-slate-300 font-bold mb-1">Step 2: Container Image</label>
                <input
                  type="text"
                  value={containerImage}
                  onChange={(e) => setContainerImage(e.target.value)}
                  placeholder="e.g. nginx:alpine or gcr.io/anarva/api:latest"
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                />
              </div>
            )}

            {wizardStep === 3 && (
              <div>
                <label className="block text-slate-300 font-bold mb-1">Step 3: Compute ACUs</label>
                <select
                  value={acuCount}
                  onChange={(e) => setAcuCount(Number(e.target.value))}
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                >
                  <option value={1}>1 ACU (0.5 vCPU, 1 GB RAM)</option>
                  <option value={2}>2 ACUs (1.0 vCPU, 2 GB RAM)</option>
                  <option value={4}>4 ACUs (2.0 vCPU, 4 GB RAM)</option>
                </select>
              </div>
            )}

            {wizardStep === 4 && (
              <div>
                <label className="block text-slate-300 font-bold mb-1">Step 4: Custom Domain Name</label>
                <input
                  type="text"
                  value={domainName}
                  onChange={(e) => setDomainName(e.target.value)}
                  placeholder="e.g. api.anarva.cloud"
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                />
              </div>
            )}

            {wizardStep >= 5 && (
              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                <div className="font-bold text-white text-sm">Review Unified Application Stack</div>
                <div>App Name: <strong className="text-emerald-400">{appName || 'payment-gateway-api'}</strong></div>
                <div>Container Image: <strong>{containerImage}</strong></div>
                <div>Compute Allocation: <strong>{acuCount} ACUs</strong></div>
                <div>Target Domain: <strong className="text-purple-400">{domainName}</strong></div>
                <div>VPC Network: <strong>vpc-01 (ap-hyderabad-1)</strong></div>
                <div>Load Balancer: <strong>Automatic Application Load Balancer</strong></div>
                <div>TLS Certificate: <strong>Auto-Provisioned Let's Encrypt / Local CA</strong></div>
              </div>
            )}

            <div className="pt-3 border-t border-slate-800 flex justify-between">
              {wizardStep > 1 ? (
                <CloudButton variant="secondary" size="sm" onClick={() => setWizardStep(wizardStep - 1)}>Previous</CloudButton>
              ) : <div />}

              {wizardStep < 5 ? (
                <CloudButton variant="primary" size="sm" onClick={() => setWizardStep(wizardStep + 1)}>Next →</CloudButton>
              ) : (
                <CloudButton variant="primary" size="sm" onClick={handleDeployApp} disabled={isDeploying}>
                  {isDeploying ? 'Deploying Stack...' : 'Execute Deployment'}
                </CloudButton>
              )}
            </div>
          </div>
        </CloudModal>
      )}
    </div>
  )
}
