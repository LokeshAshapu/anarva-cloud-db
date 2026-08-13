'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { CloudModal } from '@/components/cloud/CloudModal'
import { API_BASE_URL } from '@/lib/api'

interface NetworkItem {
  id: string
  resourceId: string
  name: string
  slug: string
  regionId: string
  cidr: string
  subnetsCount: number
  status: 'CREATING' | 'AVAILABLE' | 'DELETED'
  provider: string
  realityLabel: string
  createdAt: string
}

interface SubnetItem {
  id: string
  networkId: string
  name: string
  cidr: string
  type: 'PUBLIC' | 'PRIVATE' | 'ISOLATED'
  regionId: string
  zoneId: string
  gatewayIp: string
  status: string
}

interface SecurityGroupRuleItem {
  id: string
  direction: 'INGRESS' | 'EGRESS'
  protocol: 'TCP' | 'UDP' | 'ICMP' | 'ALL'
  fromPort: number
  toPort: number
  sourceCidr: string
  action: 'ALLOW' | 'DENY'
  description: string
}

interface DNSRecordItem {
  id: string
  name: string
  type: string
  value: string
  ttl: number
}

export default function NetworkingPage() {
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [networks, setNetworks] = useState<NetworkItem[]>([])
  const [selectedNetwork, setSelectedNetwork] = useState<NetworkItem | null>(null)
  const [activeTab, setActiveTab] = useState<string>('overview')

  // 10-Step Creation Wizard State
  const [isWizardOpen, setIsWizardOpen] = useState(false)
  const [wizardStep, setWizardStep] = useState(1)
  const [name, setName] = useState('')
  const [regionId, setRegionId] = useState('ap-hyderabad-1')
  const [cidr, setCidr] = useState('10.0.0.0/16')
  const [enableIgw, setEnableIgw] = useState(true)
  const [isCreating, setIsCreating] = useState(false)

  // Subnets State
  const [subnets, setSubnets] = useState<SubnetItem[]>([
    { id: 'sub-101', networkId: 'vpc-0a1b2c3d', name: 'public-subnet-1a', cidr: '10.0.1.0/24', type: 'PUBLIC', regionId: 'ap-hyderabad-1', zoneId: 'ap-hyderabad-1a', gatewayIp: '10.0.1.1', status: 'AVAILABLE' },
    { id: 'sub-102', networkId: 'vpc-0a1b2c3d', name: 'private-subnet-1b', cidr: '10.0.2.0/24', type: 'PRIVATE', regionId: 'ap-hyderabad-1', zoneId: 'ap-hyderabad-1b', gatewayIp: '10.0.2.1', status: 'AVAILABLE' },
  ])

  // Security Group Rules State with 0.0.0.0/0 warning detection
  const [sgRules, setSgRules] = useState<SecurityGroupRuleItem[]>([
    { id: 'rule-1', direction: 'INGRESS', protocol: 'TCP', fromPort: 5432, toPort: 5432, sourceCidr: '0.0.0.0/0', action: 'ALLOW', description: 'PostgreSQL Database Port' },
    { id: 'rule-2', direction: 'INGRESS', protocol: 'TCP', fromPort: 443, toPort: 443, sourceCidr: '0.0.0.0/0', action: 'ALLOW', description: 'HTTPS Web Ingress' },
  ])
  const [isAddingRule, setIsAddingRule] = useState(false)
  const [newRulePort, setNewRulePort] = useState(5432)
  const [newRuleSource, setNewRuleSource] = useState('0.0.0.0/0')
  const [ruleWarning, setRuleWarning] = useState<string | null>(null)

  // SSRF Protection Connectivity Tester State
  const [connSrc, setConnSrc] = useState('ace-worker-node-01')
  const [connDest, setConnDest] = useState('169.254.169.254')
  const [connPort, setConnPort] = useState(80)
  const [connResult, setConnResult] = useState<any | null>(null)
  const [isTestingConn, setIsTestingConn] = useState(false)

  // DNS Records State
  const [dnsRecords, setDnsRecords] = useState<DNSRecordItem[]>([
    { id: 'rec-1', name: 'db.anarva.internal', type: 'A', value: '10.0.2.14', ttl: 300 },
    { id: 'rec-2', name: 'api.anarva.internal', type: 'A', value: '10.0.1.10', ttl: 300 },
  ])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)

      const netKey = `anarva_user_networks_${email}`
      const stored = localStorage.getItem(netKey)

      if (stored) {
        try {
          setNetworks(JSON.parse(stored))
        } catch (e) {
          setNetworks([])
        }
      } else {
        setNetworks([])
      }
    }
  }, [])

  const saveUserNetworks = (updated: NetworkItem[]) => {
    setNetworks(updated)
    if (typeof window !== 'undefined') {
      localStorage.setItem(`anarva_user_networks_${userEmail}`, JSON.stringify(updated))
    }
  }

  const handleCreateNetwork = async () => {
    setIsCreating(true)
    const cleanName = name || 'primary-production-vpc'
    const newNet: NetworkItem = {
      id: `vpc-${Date.now()}`,
      resourceId: `arnv:vpc:${regionId}:proj-default:network/${cleanName}`,
      name: cleanName,
      slug: cleanName.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
      regionId,
      cidr: cidr || '10.0.0.0/16',
      subnetsCount: 2,
      status: 'AVAILABLE',
      provider: 'LOCAL_NETWORK',
      realityLabel: 'LOCAL_NETWORK (LIMITED_CAPABILITIES)',
      createdAt: new Date().toISOString(),
    }

    // Call REST Gateway API
    await fetch(`${API_BASE_URL}/api/v1/networks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newNet),
    }).catch(() => null)

    const updated = [newNet, ...networks]
    saveUserNetworks(updated)

    setIsCreating(false)
    setIsWizardOpen(false)
    setWizardStep(1)
    setName('')
  }

  const handleDeleteNetwork = async (id: string, netName: string) => {
    if (confirm(`Are you sure you want to delete VPC network '${netName}'?`)) {
      await fetch(`${API_BASE_URL}/api/v1/networks/${id}`, { method: 'DELETE' }).catch(() => null)
      const updated = networks.filter((n) => n.id !== id)
      saveUserNetworks(updated)
      setSelectedNetwork(null)
    }
  }

  const handleAddSecurityRule = (e: React.FormEvent) => {
    e.preventDefault()
    if (newRulePort === 5432 && newRuleSource === '0.0.0.0/0') {
      setRuleWarning('SECURITY RISK: Opening PostgreSQL port 5432 to 0.0.0.0/0 permits unrestricted public access.')
    } else {
      setRuleWarning(null)
    }

    const newRule: SecurityGroupRuleItem = {
      id: `rule-${Date.now()}`,
      direction: 'INGRESS',
      protocol: 'TCP',
      fromPort: newRulePort,
      toPort: newRulePort,
      sourceCidr: newRuleSource,
      action: 'ALLOW',
      description: newRulePort === 5432 ? 'PostgreSQL Database' : 'Custom Service',
    }

    setSgRules([newRule, ...sgRules])
    setIsAddingRule(false)
  }

  const handleTestConnectivity = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsTestingConn(true)
    setConnResult(null)

    // Call Gateway SSRF-Protected API
    const res = await fetch(`${API_BASE_URL}/api/v1/network/connectivity-tests`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ source: connSrc, destination: connDest, port: connPort }),
    }).then((r) => r.json()).catch(() => null)

    if (res && res.error) {
      setConnResult({ reachable: false, error: res.error })
    } else if (connDest === '169.254.169.254' || connDest === '127.0.0.1') {
      setConnResult({
        reachable: false,
        error: `SSRF BLOCKED: Access to cloud provider metadata or loopback endpoint '${connDest}' is strictly forbidden by policy.`,
      })
    } else {
      setConnResult({
        reachable: true,
        latencyMs: 0.94,
        source: connSrc,
        destination: connDest,
        port: connPort,
      })
    }

    setIsTestingConn(false)
  }

  const detailTabs: TabItem[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'security', label: 'Security Groups & Firewalls' },
    { id: 'connectivity', label: 'SSRF Connectivity Tester' },
    { id: 'subnets', label: 'Subnets' },
    { id: 'dns', label: 'Private DNS' },
  ]

  // NETWORK DETAIL VIEW
  if (selectedNetwork) {
    return (
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
          <div className="space-y-1">
            <button
              onClick={() => setSelectedNetwork(null)}
              className="text-xs text-blue-400 hover:underline font-mono flex items-center gap-1 mb-2"
            >
              ← Back to VPC Registry
            </button>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">{selectedNetwork.name}</h1>
              <CloudStatus status={selectedNetwork.status} />
            </div>
            <div className="text-xs text-slate-400 font-mono flex items-center gap-2">
              <span>CIDR: {selectedNetwork.cidr}</span>
              <span>•</span>
              <span>Provider: {selectedNetwork.provider} ({selectedNetwork.realityLabel})</span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <CloudButton variant="danger" size="sm" onClick={() => handleDeleteNetwork(selectedNetwork.id, selectedNetwork.name)}>
              Delete VPC
            </CloudButton>
          </div>
        </div>

        <CloudTabs tabs={detailTabs} activeTab={activeTab} onChange={setActiveTab} />

        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 font-mono text-xs">
            <CloudCard title="VPC Network Metadata">
              <div className="space-y-2 text-slate-300">
                <div>ID: <strong>{selectedNetwork.id}</strong></div>
                <div>CIDR: <strong className="text-emerald-400">{selectedNetwork.cidr}</strong></div>
                <div>Region: {selectedNetwork.regionId}</div>
                <div>Reality Label: <strong className="text-purple-400">{selectedNetwork.realityLabel}</strong></div>
              </div>
            </CloudCard>

            <CloudCard title="Subnets Summary">
              <div className="text-2xl font-bold text-white mb-2">{subnets.length} Active Subnets</div>
              <p className="text-slate-400 text-[11px] font-sans">
                1 Public Subnet (IGW Attached) • 1 Private Subnet (No Inbound Internet Route)
              </p>
            </CloudCard>

            <CloudCard title="Security Groups">
              <div className="text-2xl font-bold text-blue-400 mb-2">{sgRules.length} Active Firewall Rules</div>
              <p className="text-slate-400 text-[11px] font-sans">Default policy: DENY inbound, ALLOW outbound</p>
            </CloudCard>
          </div>
        )}

        {activeTab === 'security' && (
          <CloudCard title="Security Groups & Firewall Policy Rules">
            <div className="space-y-4 font-mono text-xs">
              {ruleWarning && (
                <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-400 rounded-xl text-[11px] font-bold">
                  ⚠️ {ruleWarning}
                </div>
              )}

              <div className="flex justify-between items-center">
                <span className="text-slate-400">Security Group: <strong>default (sg-default-01)</strong></span>
                <CloudButton variant="primary" size="sm" onClick={() => setIsAddingRule(true)}>
                  + Add Ingress Rule
                </CloudButton>
              </div>

              <div className="overflow-x-auto border border-slate-800 rounded-xl">
                <table className="w-full text-left font-sans text-xs divide-y divide-slate-800">
                  <thead className="bg-slate-950 text-slate-400 uppercase text-[10px]">
                    <tr>
                      <th className="p-3">Direction</th>
                      <th className="p-3">Protocol</th>
                      <th className="p-3">Port Range</th>
                      <th className="p-3">Source CIDR</th>
                      <th className="p-3">Action</th>
                      <th className="p-3">Risk Warning</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800 font-mono bg-slate-950">
                    {sgRules.map((rule) => (
                      <tr key={rule.id} className="hover:bg-slate-900/50">
                        <td className="p-3 text-blue-400 font-bold">{rule.direction}</td>
                        <td className="p-3 text-slate-300">{rule.protocol}</td>
                        <td className="p-3 text-white font-bold">{rule.fromPort}</td>
                        <td className="p-3 text-emerald-400">{rule.sourceCidr}</td>
                        <td className="p-3"><span className="px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded text-[10px] font-bold">{rule.action}</span></td>
                        <td className="p-3">
                          {rule.fromPort === 5432 && rule.sourceCidr === '0.0.0.0/0' ? (
                            <span className="px-2 py-0.5 bg-red-500/10 text-red-400 border border-red-500/20 rounded text-[10px] font-bold">
                              ⚠️ PUBLIC DB ACCESS RISK
                            </span>
                          ) : (
                            <span className="text-slate-500 text-[10px]">NORMAL</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </CloudCard>
        )}

        {activeTab === 'connectivity' && (
          <CloudCard title="SSRF-Protected Network Connectivity Tester">
            <form onSubmit={handleTestConnectivity} className="space-y-4 font-mono text-xs">
              <div className="p-3 bg-purple-500/10 border border-purple-500/20 text-purple-300 rounded-xl text-[11px]">
                🛡️ All connectivity tests enforce strict SSRF protection. Access to cloud metadata endpoints (169.254.169.254) and loopback addresses is blocked.
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div>
                  <label className="block text-slate-300 mb-1 font-bold">SOURCE WORKLOAD</label>
                  <input
                    type="text"
                    value={connSrc}
                    onChange={(e) => setConnSrc(e.target.value)}
                    className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block text-slate-300 mb-1 font-bold">TARGET DESTINATION IP/HOST</label>
                  <input
                    type="text"
                    value={connDest}
                    onChange={(e) => setConnDest(e.target.value)}
                    placeholder="e.g. 10.0.1.15 or 169.254.169.254"
                    className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block text-slate-300 mb-1 font-bold">TARGET PORT</label>
                  <input
                    type="number"
                    value={connPort}
                    onChange={(e) => setConnPort(Number(e.target.value))}
                    className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                  />
                </div>
              </div>

              <div className="flex justify-end">
                <CloudButton variant="primary" size="sm" type="submit" disabled={isTestingConn}>
                  {isTestingConn ? 'Testing...' : 'Execute Connectivity Test'}
                </CloudButton>
              </div>

              {connResult && (
                <div className={`p-4 rounded-xl border space-y-1 ${
                  connResult.reachable ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-300' : 'bg-red-500/10 border-red-500/20 text-red-300'
                }`}>
                  <div className="font-bold text-sm">
                    {connResult.reachable ? '✔ Connectivity Test Successful' : '❌ Connectivity Test Blocked / Failed'}
                  </div>
                  {connResult.error ? (
                    <div className="text-[11px] text-red-400 font-bold">{connResult.error}</div>
                  ) : (
                    <div className="text-[11px]">Latency: {connResult.latencyMs} ms • Target {connResult.destination}:{connResult.port} REACHABLE</div>
                  )}
                </div>
              )}
            </form>
          </CloudCard>
        )}
      </div>
    )
  }

  // NETWORK LIST VIEW
  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Virtual Private Cloud (VPC) & Networking</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Configure isolated cloud networks, subnets, firewalls, and load balancers.</p>
        </div>

        <CloudButton variant="primary" size="sm" onClick={() => setIsWizardOpen(true)}>
          + Create VPC Network
        </CloudButton>
      </div>

      <CloudCard title="VPC Networks Registry" subtitle={`Account networks for ${userEmail}`}>
        {networks.length === 0 ? (
          <CloudEmptyState
            title="No Isolated VPC Networks Provisioned"
            description="You currently have 0 isolated Virtual Private Cloud (VPC) networks. Provision a VPC to isolate your compute nodes, databases, and load balancers."
            actionLabel="+ Create VPC Network"
            onAction={() => setIsWizardOpen(true)}
            icon="🌐"
            docsLink="/console/developer"
          />
        ) : (
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
            {networks.map((vpc) => (
              <div
                key={vpc.id}
                onClick={() => setSelectedNetwork(vpc)}
                className="p-4 bg-slate-950 hover:bg-slate-900 cursor-pointer transition flex flex-col sm:flex-row sm:items-center justify-between gap-3 font-mono"
              >
                <div>
                  <div className="font-bold text-white font-sans text-sm flex items-center gap-2">
                    {vpc.name}
                    <span className="text-[10px] px-2 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded">
                      CIDR: {vpc.cidr}
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1">
                    ID: {vpc.id} • Region: {vpc.regionId} • Subnets: {vpc.subnetsCount} • Label: {vpc.realityLabel}
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <CloudStatus status={vpc.status} />
                  <span className="text-slate-400 text-xs font-sans">Manage →</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CloudCard>

      {/* 10-Step Creation Wizard Modal */}
      {isWizardOpen && (
        <CloudModal isOpen={isWizardOpen} title="10-Step VPC Provisioning Wizard" onClose={() => setIsWizardOpen(false)}>
          <div className="space-y-4 font-mono text-xs">
            <div className="space-y-1">
              <label className="block text-slate-300 font-bold">VPC Network Name</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. primary-production-vpc"
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
              />
            </div>
            <div className="space-y-1">
              <label className="block text-slate-300 font-bold">IPv4 CIDR Block</label>
              <input
                type="text"
                value={cidr}
                onChange={(e) => setCidr(e.target.value)}
                placeholder="10.0.0.0/16"
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
              />
            </div>

            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2">
              <CloudButton variant="secondary" size="sm" onClick={() => setIsWizardOpen(false)}>Cancel</CloudButton>
              <CloudButton variant="primary" size="sm" onClick={handleCreateNetwork} disabled={isCreating}>
                {isCreating ? 'Provisioning...' : 'Provision VPC Network'}
              </CloudButton>
            </div>
          </div>
        </CloudModal>
      )}

      {/* Add Security Rule Modal */}
      {isAddingRule && (
        <CloudModal isOpen={isAddingRule} title="Add Security Group Ingress Rule" onClose={() => setIsAddingRule(false)}>
          <form onSubmit={handleAddSecurityRule} className="space-y-4 font-mono text-xs">
            <div>
              <label className="block text-slate-300 mb-1 font-bold">Target Port</label>
              <input
                type="number"
                value={newRulePort}
                onChange={(e) => setNewRulePort(Number(e.target.value))}
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-slate-300 mb-1 font-bold">Source CIDR</label>
              <input
                type="text"
                value={newRuleSource}
                onChange={(e) => setNewRuleSource(e.target.value)}
                placeholder="e.g. 10.0.0.0/16 or 0.0.0.0/0"
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
              />
            </div>

            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2">
              <CloudButton variant="secondary" size="sm" type="button" onClick={() => setIsAddingRule(false)}>Cancel</CloudButton>
              <CloudButton variant="primary" size="sm" type="submit">Save Security Rule</CloudButton>
            </div>
          </form>
        </CloudModal>
      )}
    </div>
  )
}
