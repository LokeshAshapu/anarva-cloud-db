'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { createClient } from '@/utils/supabase/client'

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
  createdAt: string
}

interface SubnetItem {
  id: string
  networkId: string
  name: string
  cidr: string
  type: 'PUBLIC' | 'PRIVATE' | 'INTERNAL'
  regionId: string
  zoneId: string
  gatewayIp: string
  status: string
}

interface SecurityGroupRuleItem {
  id: string
  direction: 'INGRESS' | 'EGRESS'
  protocol: 'TCP' | 'UDP' | 'ICMP'
  fromPort: number
  toPort: number
  sourceCidr: string
  action: 'ALLOW' | 'DENY'
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

  // Creation Wizard State
  const [isWizardOpen, setIsWizardOpen] = useState(false)
  const [wizardStep, setWizardStep] = useState(1)
  const [name, setName] = useState('')
  const [regionId, setRegionId] = useState('us-east-1')
  const [cidr, setCidr] = useState('10.0.0.0/16')
  const [enableIgw, setEnableIgw] = useState(true)
  const [isCreating, setIsCreating] = useState(false)

  // Subnet Modal State
  const [subnets, setSubnets] = useState<SubnetItem[]>([
    { id: 'sub-101', networkId: 'vpc-0a1b2c3d', name: 'public-subnet-1a', cidr: '10.0.1.0/24', type: 'PUBLIC', regionId: 'us-east-1', zoneId: 'us-east-1a', gatewayIp: '10.0.1.1', status: 'AVAILABLE' },
    { id: 'sub-102', networkId: 'vpc-0a1b2c3d', name: 'private-subnet-1b', cidr: '10.0.2.0/24', type: 'PRIVATE', regionId: 'us-east-1', zoneId: 'us-east-1b', gatewayIp: '10.0.2.1', status: 'AVAILABLE' },
  ])
  const [isSubnetModalOpen, setIsSubnetModalOpen] = useState(false)
  const [subName, setSubName] = useState('')
  const [subCidr, setSubCidr] = useState('10.0.3.0/24')
  const [subType, setSubType] = useState<'PUBLIC' | 'PRIVATE'>('PUBLIC')

  // Security Group Rules State
  const [sgRules, setSgRules] = useState<SecurityGroupRuleItem[]>([
    { id: 'rule-1', direction: 'INGRESS', protocol: 'TCP', fromPort: 5432, toPort: 5432, sourceCidr: '0.0.0.0/0', action: 'ALLOW' },
    { id: 'rule-2', direction: 'INGRESS', protocol: 'TCP', fromPort: 443, toPort: 443, sourceCidr: '0.0.0.0/0', action: 'ALLOW' },
    { id: 'rule-3', direction: 'INGRESS', protocol: 'TCP', fromPort: 80, toPort: 80, sourceCidr: '0.0.0.0/0', action: 'ALLOW' },
  ])

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
        setNetworks(JSON.parse(stored))
      } else if (email === 'lokeshashapu@gmail.com') {
        const defaults: NetworkItem[] = [
          {
            id: 'vpc-0a1b2c3d',
            resourceId: 'arnv:vpc:us-east-1:proj-default:vpc/primary-production-vpc',
            name: 'Primary Production VPC',
            slug: 'primary-production-vpc',
            regionId: 'us-east-1',
            cidr: '10.0.0.0/16',
            subnetsCount: 2,
            status: 'AVAILABLE',
            provider: 'LOCAL_DOCKER',
            createdAt: new Date().toISOString(),
          },
        ]
        setNetworks(defaults)
        localStorage.setItem(netKey, JSON.stringify(defaults))
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

  const handleCreateNetwork = () => {
    setIsCreating(true)
    setTimeout(() => {
      const vpcName = name || 'custom-cloud-vpc'
      const newNet: NetworkItem = {
        id: `vpc-${Math.floor(Math.random() * 90000 + 10000)}`,
        resourceId: `arnv:vpc:${regionId}:proj-default:vpc/${vpcName}`,
        name: vpcName,
        slug: vpcName.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
        regionId,
        cidr: cidr || '10.0.0.0/16',
        subnetsCount: 2,
        status: 'AVAILABLE',
        provider: 'LOCAL_DOCKER',
        createdAt: new Date().toISOString(),
      }

      const updated = [newNet, ...networks]
      saveUserNetworks(updated)

      // Add default subnets for the new network
      setSubnets((prev) => [
        ...prev,
        { id: `sub-${Date.now()}-1`, networkId: newNet.id, name: `${vpcName}-public-1a`, cidr: '10.0.1.0/24', type: 'PUBLIC', regionId, zoneId: `${regionId}a`, gatewayIp: '10.0.1.1', status: 'AVAILABLE' },
        { id: `sub-${Date.now()}-2`, networkId: newNet.id, name: `${vpcName}-private-1b`, cidr: '10.0.2.0/24', type: 'PRIVATE', regionId, zoneId: `${regionId}b`, gatewayIp: '10.0.2.1', status: 'AVAILABLE' },
      ])

      // Record activity event
      if (typeof window !== 'undefined') {
        const actKey = `anarva_user_activities_${userEmail}`
        const existingActs = JSON.parse(localStorage.getItem(actKey) || '[]')
        const newAct = {
          id: `act-${Date.now()}`,
          action: 'NETWORK_CREATED',
          resource: vpcName,
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

  const handleDeleteNetwork = (id: string, netName: string) => {
    if (confirm(`Are you sure you want to delete VPC network '${netName}'?`)) {
      const updated = networks.filter((n) => n.id !== id)
      saveUserNetworks(updated)

      if (typeof window !== 'undefined') {
        const actKey = `anarva_user_activities_${userEmail}`
        const existingActs = JSON.parse(localStorage.getItem(actKey) || '[]')
        const newAct = {
          id: `act-${Date.now()}`,
          action: 'NETWORK_DELETED',
          resource: netName,
          actor: userEmail,
          time: 'Just now',
        }
        localStorage.setItem(actKey, JSON.stringify([newAct, ...existingActs]))
      }

      setSelectedNetwork(null)
    }
  }

  const handleAddSubnet = () => {
    if (!selectedNetwork) return
    const newSub: SubnetItem = {
      id: `sub-${Date.now()}`,
      networkId: selectedNetwork.id,
      name: subName || 'new-subnet',
      cidr: subCidr || '10.0.3.0/24',
      type: subType,
      regionId: selectedNetwork.regionId,
      zoneId: `${selectedNetwork.regionId}a`,
      gatewayIp: subCidr.replace(/\.0\/24$/, '.1'),
      status: 'AVAILABLE',
    }
    setSubnets([...subnets, newSub])
    setIsSubnetModalOpen(false)
    setSubName('')
  }

  const detailTabs: TabItem[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'topology', label: 'Network Topology' },
    { id: 'subnets', label: 'Subnets' },
    { id: 'routes', label: 'Route Tables' },
    { id: 'security', label: 'Security Groups' },
    { id: 'ipam', label: 'IPAM & Addresses' },
    { id: 'dns', label: 'Private DNS' },
    { id: 'loadbalancers', label: 'Load Balancers' },
    { id: 'danger', label: 'Danger Zone' },
  ]

  // NETWORK DETAIL VIEW
  if (selectedNetwork) {
    const netSubnets = subnets.filter((s) => s.networkId === selectedNetwork.id)

    return (
      <div className="space-y-6">
        {/* Header */}
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
              <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-xs font-mono font-bold">
                {selectedNetwork.cidr}
              </span>
            </div>
            <div className="text-xs text-slate-400 font-mono flex items-center gap-2">
              <span className="text-emerald-400 font-bold bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
                {selectedNetwork.resourceId}
              </span>
              <span>•</span>
              <span>Provider: LOCAL DEVELOPMENT PROVIDER (Docker Bridge)</span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <CloudButton variant="secondary" size="sm" onClick={() => setIsSubnetModalOpen(true)}>
              + Create Subnet
            </CloudButton>
            <CloudButton variant="danger" size="sm" onClick={() => handleDeleteNetwork(selectedNetwork.id, selectedNetwork.name)}>
              Delete VPC
            </CloudButton>
          </div>
        </div>

        {/* Tabs */}
        <CloudTabs tabs={detailTabs} activeTab={activeTab} onChange={setActiveTab} />

        {/* Tab Contents */}
        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <CloudCard title="VPC Network Configuration">
              <div className="space-y-3 text-xs font-mono">
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">IPv4 CIDR Block:</span>
                  <span className="text-blue-400 font-bold">{selectedNetwork.cidr}</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Associated Region:</span>
                  <span className="text-white font-bold">{selectedNetwork.regionId}</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Total Provisioned Subnets:</span>
                  <span className="text-white font-bold">{netSubnets.length > 0 ? netSubnets.length : 2} Subnets</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Internet Gateway (IGW):</span>
                  <span className="text-emerald-400 font-bold">ATTACHED (igw-0a1b2c3d)</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">DNS Resolution:</span>
                  <span className="text-white font-bold">ENABLED (anarva.internal)</span>
                </div>
              </div>
            </CloudCard>

            <CloudCard title="IPAM & Address Summary">
              <div className="space-y-3 text-xs font-mono">
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Available IPv4 Addresses:</span>
                  <span className="text-emerald-400 font-bold">65,531 IPs Available</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Allocated Private IPs:</span>
                  <span className="text-white font-bold">4 Allocated</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Reserved Gateway IPs:</span>
                  <span className="text-slate-300 font-bold">10.0.0.1, 10.0.1.1, 10.0.2.1</span>
                </div>
              </div>
            </CloudCard>
          </div>
        )}

        {activeTab === 'topology' && (
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-6 shadow-2xl">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3 font-mono text-xs">
              <span className="font-bold text-white">Interactive Network Topology Diagram</span>
              <span className="text-blue-400 font-bold">VPC: {selectedNetwork.cidr}</span>
            </div>

            {/* Architecture Map */}
            <div className="p-6 bg-slate-950 border border-slate-800 rounded-xl space-y-6 text-xs font-mono text-center">
              {/* Internet Gateway */}
              <div className="inline-block px-4 py-2 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 font-bold rounded-xl shadow-lg">
                🌐 Public Internet & Gateway (igw-0a1b2c3d)
              </div>

              <div className="text-slate-500 font-bold">↓</div>

              {/* VPC Box */}
              <div className="border border-blue-500/30 bg-blue-500/5 rounded-2xl p-6 space-y-6 text-left">
                <div className="text-blue-400 font-bold flex items-center justify-between">
                  <span>ANARVA VPC: {selectedNetwork.name} ({selectedNetwork.cidr})</span>
                  <span className="text-[10px] bg-blue-500/10 px-2 py-0.5 rounded border border-blue-500/20">ISOLATED</span>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {/* Public Subnet */}
                  <div className="bg-slate-900 border border-emerald-500/30 rounded-xl p-4 space-y-3">
                    <div className="font-bold text-emerald-400 text-xs flex items-center justify-between">
                      <span>PUBLIC SUBNET (10.0.1.0/24)</span>
                      <span className="text-[10px] text-slate-400">IGW Route</span>
                    </div>
                    <div className="space-y-2">
                      <div className="p-2 bg-slate-950 border border-slate-800 rounded-lg text-white font-bold">
                        ⚡ Application Load Balancer (lb-app-01)
                      </div>
                      <div className="p-2 bg-slate-950 border border-slate-800 rounded-lg text-white font-bold">
                        💻 ACE Compute Instance (ace-worker-node-01)
                      </div>
                    </div>
                  </div>

                  {/* Private Subnet */}
                  <div className="bg-slate-900 border border-purple-500/30 rounded-xl p-4 space-y-3">
                    <div className="font-bold text-purple-400 text-xs flex items-center justify-between">
                      <span>PRIVATE SUBNET (10.0.2.0/24)</span>
                      <span className="text-[10px] text-slate-400">Internal Only</span>
                    </div>
                    <div className="space-y-2">
                      <div className="p-2 bg-slate-950 border border-slate-800 rounded-lg text-white font-bold">
                        🗄️ PostgreSQL Managed Cluster (production-db)
                      </div>
                      <div className="p-2 bg-slate-950 border border-slate-800 rounded-lg text-white font-bold">
                        🔒 Private DNS Endpoint (db.anarva.internal)
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'subnets' && (
          <CloudCard title="VPC Subnets Registry">
            <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
              {netSubnets.map((sub) => (
                <div key={sub.id} className="p-4 bg-slate-950 flex items-center justify-between">
                  <div>
                    <div className="font-bold text-white font-sans flex items-center gap-2">
                      {sub.name}
                      <span
                        className={`text-[10px] px-2 py-0.5 rounded font-normal border ${
                          sub.type === 'PUBLIC'
                            ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                            : 'bg-purple-500/10 text-purple-400 border-purple-500/20'
                        }`}
                      >
                        {sub.type}
                      </span>
                    </div>
                    <div className="text-[10px] text-slate-500 mt-0.5">
                      CIDR: {sub.cidr} • Gateway IP: {sub.gatewayIp} • Zone: {sub.zoneId}
                    </div>
                  </div>
                  <CloudStatus status={sub.status} />
                </div>
              ))}
            </div>
          </CloudCard>
        )}

        {activeTab === 'routes' && (
          <CloudCard title="Route Tables">
            <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
              <div className="p-4 bg-slate-950 flex items-center justify-between">
                <div>
                  <div className="font-bold text-white">10.0.0.0/16</div>
                  <div className="text-[10px] text-slate-500">Destination CIDR</div>
                </div>
                <div className="text-emerald-400 font-bold">LOCAL</div>
              </div>
              <div className="p-4 bg-slate-950 flex items-center justify-between">
                <div>
                  <div className="font-bold text-white">0.0.0.0/0</div>
                  <div className="text-[10px] text-slate-500">Destination CIDR</div>
                </div>
                <div className="text-blue-400 font-bold">INTERNET_GATEWAY (igw-0a1b2c3d)</div>
              </div>
            </div>
          </CloudCard>
        )}

        {activeTab === 'security' && (
          <CloudCard title="Security Groups Firewall Rules">
            <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
              {sgRules.map((rule) => (
                <div key={rule.id} className="p-4 bg-slate-950 flex items-center justify-between">
                  <div>
                    <div className="font-bold text-white">
                      {rule.direction} {rule.protocol} Port {rule.fromPort}
                    </div>
                    <div className="text-[10px] text-slate-500 mt-0.5">Source: {rule.sourceCidr}</div>
                  </div>
                  <span className="px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded font-bold">
                    {rule.action}
                  </span>
                </div>
              ))}
            </div>
          </CloudCard>
        )}

        {/* Add Subnet Modal */}
        {isSubnetModalOpen && (
          <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
            <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 space-y-4">
              <h3 className="text-base font-bold text-white">Create New Subnet</h3>
              <div className="space-y-3 text-xs">
                <div className="space-y-1">
                  <label className="font-semibold text-slate-300">Subnet Name</label>
                  <input
                    type="text"
                    value={subName}
                    onChange={(e) => setSubName(e.target.value)}
                    placeholder="e.g. public-subnet-1c"
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono"
                  />
                </div>
                <div className="space-y-1">
                  <label className="font-semibold text-slate-300">Subnet CIDR Block</label>
                  <input
                    type="text"
                    value={subCidr}
                    onChange={(e) => setSubCidr(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono"
                  />
                </div>
                <div className="space-y-1">
                  <label className="font-semibold text-slate-300">Subnet Access Type</label>
                  <select
                    value={subType}
                    onChange={(e) => setSubType(e.target.value as any)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono"
                  >
                    <option value="PUBLIC">PUBLIC (Internet Gateway Route)</option>
                    <option value="PRIVATE">PRIVATE (Internal Workloads Only)</option>
                  </select>
                </div>
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <CloudButton variant="outline" size="sm" onClick={() => setIsSubnetModalOpen(false)}>
                  Cancel
                </CloudButton>
                <CloudButton variant="primary" size="sm" onClick={handleAddSubnet}>
                  Create Subnet
                </CloudButton>
              </div>
            </div>
          </div>
        )}
      </div>
    )
  }

  // NETWORK LIST VIEW
  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Virtual Private Cloud (VPC) & Networking</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Configure isolated cloud networks, subnets, firewalls, and load balancers.</p>
        </div>

        <CloudButton variant="primary" size="sm" onClick={() => setIsWizardOpen(true)}>
          + Create VPC Network
        </CloudButton>
      </div>

      {/* Overview Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Active VPC Networks</div>
          <div className="text-3xl font-extrabold text-white font-mono">{networks.length}</div>
          <div className="text-xs text-slate-400">Total Subnets: {subnets.length}</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Security Group Rules</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">{sgRules.length} Active Rules</div>
          <div className="text-xs text-slate-400">Ingress: Port 5432, 443, 80</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Application Load Balancers</div>
          <div className="text-3xl font-extrabold text-emerald-400 font-mono">1 Active</div>
          <div className="text-xs text-slate-400">TLS 1.3 Termination Active</div>
        </div>
      </div>

      {/* VPC Table */}
      <CloudCard title="VPC Networks Registry" subtitle={`Account networks for ${userEmail}`}>
        {networks.length === 0 ? (
          <CloudEmptyState
            title="No VPC networks created yet"
            description="You currently have 0 isolated VPC networks. Click '+ Create VPC Network' to set up your network boundary."
            actionLabel="+ Create VPC Network"
            onAction={() => setIsWizardOpen(true)}
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
                    <span className="text-[10px] px-2 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded font-normal">
                      {vpc.cidr}
                    </span>
                  </div>
                  <div className="text-slate-400 text-[11px] mt-1">
                    ID: {vpc.id} • Region: {vpc.regionId} • Subnets: {subnets.filter((s) => s.networkId === vpc.id).length}
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  <CloudStatus status={vpc.status} />
                  <span className="text-slate-400 text-xs font-sans">Manage →</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CloudCard>

      {/* Creation Wizard Modal */}
      {isWizardOpen && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 space-y-6 shadow-2xl">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <h3 className="text-base font-bold text-white">7-Step VPC Creation Wizard</h3>
              <span className="text-xs font-mono text-blue-400">Step {wizardStep} of 2</span>
            </div>

            {wizardStep === 1 && (
              <div className="space-y-4 text-xs">
                <div className="space-y-1">
                  <label className="font-semibold text-slate-300">1. VPC Network Name</label>
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g. Primary Production VPC"
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono"
                  />
                </div>
                <div className="space-y-1">
                  <label className="font-semibold text-slate-300">2. Target Region</label>
                  <select
                    value={regionId}
                    onChange={(e) => setRegionId(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono"
                  >
                    <option value="us-east-1">us-east-1 (N. Virginia)</option>
                    <option value="ap-hyderabad-1">ap-hyderabad-1 (Hyderabad)</option>
                    <option value="ap-mumbai-1">ap-mumbai-1 (Mumbai)</option>
                  </select>
                </div>
                <div className="space-y-1">
                  <label className="font-semibold text-slate-300">3. IPv4 CIDR Block</label>
                  <input
                    type="text"
                    value={cidr}
                    onChange={(e) => setCidr(e.target.value)}
                    placeholder="10.0.0.0/16"
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono"
                  />
                </div>
              </div>
            )}

            {wizardStep === 2 && (
              <div className="space-y-4 text-xs font-mono">
                <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                  <div className="font-bold text-white">VPC Summary</div>
                  <div>Network Name: {name || 'custom-cloud-vpc'}</div>
                  <div>Region: {regionId}</div>
                  <div>IPv4 CIDR: {cidr}</div>
                  <div>Default Subnets: Public (10.0.1.0/24), Private (10.0.2.0/24)</div>
                  <div>Provider: LOCAL DEVELOPMENT PROVIDER (Docker Bridge)</div>
                </div>
              </div>
            )}

            <div className="flex items-center justify-between pt-2">
              <CloudButton
                variant="outline"
                size="sm"
                onClick={() => (wizardStep > 1 ? setWizardStep(wizardStep - 1) : setIsWizardOpen(false))}
              >
                {wizardStep > 1 ? 'Back' : 'Cancel'}
              </CloudButton>

              {wizardStep < 2 ? (
                <CloudButton variant="primary" size="sm" onClick={() => setWizardStep(wizardStep + 1)}>
                  Next Step →
                </CloudButton>
              ) : (
                <CloudButton variant="primary" size="sm" onClick={handleCreateNetwork} disabled={isCreating}>
                  {isCreating ? 'Provisioning...' : 'Provision VPC Network'}
                </CloudButton>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
