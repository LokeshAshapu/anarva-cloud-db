'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { CloudModal } from '@/components/cloud/CloudModal'
import { API_BASE_URL } from '@/lib/api'

interface LoadBalancerItem {
  id: string
  name: string
  provider: string
  type: string
  scheme: string
  networkId: string
  status: string
  ipReference: string
  hostnameReference: string
  realityLabel: string
  createdAt: string
}

export default function LoadBalancersPage() {
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [lbs, setLbs] = useState<LoadBalancerItem[]>([])
  const [selectedLB, setSelectedLB] = useState<LoadBalancerItem | null>(null)
  const [activeTab, setActiveTab] = useState<string>('overview')

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [lbName, setLbName] = useState('')
  const [lbType, setLbType] = useState('APPLICATION')
  const [lbScheme, setLbScheme] = useState('PUBLIC')
  const [isCreating, setIsCreating] = useState(false)

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)

      const stored = localStorage.getItem(`anarva_user_lbs_${email}`)
      if (stored) {
        try {
          setLbs(JSON.parse(stored))
        } catch (e) {
          setLbs([])
        }
      } else {
        setLbs([])
      }
    }
  }, [])

  const saveLBs = (updated: LoadBalancerItem[]) => {
    setLbs(updated)
    if (typeof window !== 'undefined') {
      localStorage.setItem(`anarva_user_lbs_${userEmail}`, JSON.stringify(updated))
    }
  }

  const handleCreateLB = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsCreating(true)

    const newLb: LoadBalancerItem = {
      id: `lb-${Date.now()}`,
      name: lbName || 'primary-alb',
      provider: 'LOCAL_LOAD_BALANCER',
      type: lbType,
      scheme: lbScheme,
      networkId: 'vpc-01',
      status: 'ACTIVE',
      ipReference: '127.0.0.1',
      hostnameReference: `lb-${Date.now()}.anarva.local`,
      realityLabel: 'LOCAL_LOAD_BALANCER (LIMITED_CAPABILITIES)',
      createdAt: new Date().toISOString(),
    }

    await fetch(`${API_BASE_URL}/api/v1/load-balancers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newLb),
    }).catch(() => null)

    const updated = [newLb, ...lbs]
    saveLBs(updated)
    setIsCreating(false)
    setIsModalOpen(false)
    setLbName('')
  }

  const handleDeleteLB = async (id: string) => {
    if (confirm('Delete load balancer?')) {
      await fetch(`${API_BASE_URL}/api/v1/load-balancers/${id}`, { method: 'DELETE' }).catch(() => null)
      const updated = lbs.filter((l) => l.id !== id)
      saveLBs(updated)
      setSelectedLB(null)
    }
  }

  const tabs: TabItem[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'listeners', label: 'Listeners & Rules' },
    { id: 'pools', label: 'Backend Pools' },
    { id: 'certs', label: 'TLS Certificates' },
  ]

  if (selectedLB) {
    return (
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
          <div>
            <button onClick={() => setSelectedLB(null)} className="text-xs text-blue-400 font-mono mb-2">
              ← Back to Load Balancers
            </button>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">{selectedLB.name}</h1>
              <CloudStatus status={selectedLB.status} />
            </div>
            <div className="text-xs text-slate-400 font-mono">
              Scheme: {selectedLB.scheme} • Type: {selectedLB.type} • IP: {selectedLB.ipReference}
            </div>
          </div>
          <CloudButton variant="danger" size="sm" onClick={() => handleDeleteLB(selectedLB.id)}>
            Delete Load Balancer
          </CloudButton>
        </div>

        <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 font-mono text-xs">
            <CloudCard title="Endpoint Metadata">
              <div className="space-y-2 text-slate-300">
                <div>Hostname: <strong>{selectedLB.hostnameReference}</strong></div>
                <div>IP Address: <strong>{selectedLB.ipReference}</strong></div>
                <div>Label: <strong className="text-purple-400">{selectedLB.realityLabel}</strong></div>
              </div>
            </CloudCard>
            <CloudCard title="Listeners">
              <div className="text-2xl font-bold text-emerald-400">2 Active</div>
              <p className="text-slate-400 font-sans text-xs">Port 80 (HTTP) • Port 443 (HTTPS)</p>
            </CloudCard>
            <CloudCard title="Backend Targets">
              <div className="text-2xl font-bold text-blue-400">4 Targets</div>
              <p className="text-slate-400 font-sans text-xs">Round-Robin Algorithm Active</p>
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
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Application Load Balancers</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">High-performance HTTP/HTTPS application delivery & traffic distribution.</p>
        </div>
        <CloudButton variant="primary" size="sm" onClick={() => setIsModalOpen(true)}>
          + Create Load Balancer
        </CloudButton>
      </div>

      <CloudCard title="Load Balancers Registry" subtitle={`Account endpoints for ${userEmail}`}>
        {lbs.length === 0 ? (
          <CloudEmptyState
            title="No Application Load Balancers Configured"
            description="Provision a Load Balancer to distribute inbound HTTP/HTTPS traffic across compute ACUs and container instances."
            actionLabel="+ Create Load Balancer"
            onAction={() => setIsModalOpen(true)}
            icon="⚖️"
            docsLink="/console/developer"
          />
        ) : (
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
            {lbs.map((lb) => (
              <div
                key={lb.id}
                onClick={() => setSelectedLB(lb)}
                className="p-4 bg-slate-950 hover:bg-slate-900 cursor-pointer transition flex items-center justify-between font-mono"
              >
                <div>
                  <div className="font-bold text-white text-sm font-sans flex items-center gap-2">
                    {lb.name}
                    <span className="text-[10px] px-2 py-0.5 bg-purple-500/10 text-purple-400 border border-purple-500/20 rounded">
                      {lb.type} ({lb.scheme})
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1">
                    ID: {lb.id} • Host: {lb.hostnameReference} • Label: {lb.realityLabel}
                  </div>
                </div>
                <CloudStatus status={lb.status} />
              </div>
            ))}
          </div>
        )}
      </CloudCard>

      {isModalOpen && (
        <CloudModal isOpen={isModalOpen} title="Provision Load Balancer" onClose={() => setIsModalOpen(false)}>
          <form onSubmit={handleCreateLB} className="space-y-4 font-mono text-xs">
            <div>
              <label className="block text-slate-300 font-bold mb-1">Load Balancer Name</label>
              <input
                type="text"
                value={lbName}
                onChange={(e) => setLbName(e.target.value)}
                placeholder="e.g. primary-alb"
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-slate-300 font-bold mb-1">Type</label>
                <select
                  value={lbType}
                  onChange={(e) => setLbType(e.target.value)}
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                >
                  <option value="APPLICATION">APPLICATION (Layer 7)</option>
                  <option value="NETWORK">NETWORK (Layer 4)</option>
                </select>
              </div>
              <div>
                <label className="block text-slate-300 font-bold mb-1">Scheme</label>
                <select
                  value={lbScheme}
                  onChange={(e) => setLbScheme(e.target.value)}
                  className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                >
                  <option value="PUBLIC">PUBLIC (Internet Facing)</option>
                  <option value="INTERNAL">INTERNAL (VPC Private)</option>
                </select>
              </div>
            </div>

            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2">
              <CloudButton variant="secondary" size="sm" type="button" onClick={() => setIsModalOpen(false)}>Cancel</CloudButton>
              <CloudButton variant="primary" size="sm" type="submit" disabled={isCreating}>
                {isCreating ? 'Provisioning...' : 'Provision Load Balancer'}
              </CloudButton>
            </div>
          </form>
        </CloudModal>
      )}
    </div>
  )
}
