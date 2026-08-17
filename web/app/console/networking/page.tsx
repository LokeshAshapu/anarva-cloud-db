'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { API_BASE_URL, getAuthHeaders } from '@/lib/api'

interface VPCItem {
  id: string
  name: string
  cidr: string
  regionId: string
  status: string
  dnsEnabled: boolean
  createdAt: string
}

interface SubnetItem {
  id: string
  networkId: string
  name: string
  cidr: string
  availabilityZone: string
  type: string
  status: string
}

interface SecurityGroupItem {
  id: string
  networkId: string
  name: string
  description: string
  status: string
  rules: {
    id: string
    direction: string
    protocol: string
    fromPort: number
    toPort: number
    cidr: string
    action: string
  }[]
}

interface RouteTableItem {
  id: string
  networkId: string
  name: string
  status: string
  routes: {
    id: string
    destination: string
    target: string
    targetType: string
  }[]
}

export default function NetworkingConsolePage() {
  const [activeTab, setActiveTab] = useState('vpcs')
  const [vpcs, setVpcs] = useState<VPCItem[]>([])
  const [subnets, setSubnets] = useState<SubnetItem[]>([])
  const [securityGroups, setSecurityGroups] = useState<SecurityGroupItem[]>([])
  const [routeTables, setRouteTables] = useState<RouteTableItem[]>([])
  const [isLoading, setIsLoading] = useState(true)

  // VPC Modal State
  const [isVpcModalOpen, setIsVpcModalOpen] = useState(false)
  const [newVpcName, setNewVpcName] = useState('')
  const [newVpcCidr, setNewVpcCidr] = useState('10.0.0.0/16')
  const [newVpcRegion, setNewVpcRegion] = useState('us-east-1')
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Subnet Modal State
  const [isSubnetModalOpen, setIsSubnetModalOpen] = useState(false)
  const [newSubnetVpcId, setNewSubnetVpcId] = useState('')
  const [newSubnetName, setNewSubnetName] = useState('')
  const [newSubnetCidr, setNewSubnetCidr] = useState('10.0.1.0/24')
  const [newSubnetZone, setNewSubnetZone] = useState('us-east-1a')
  const [newSubnetType, setNewSubnetType] = useState('PRIVATE')

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    setIsLoading(true)
    try {
      const authHeaders = getAuthHeaders()
      const [vpcRes, subRes, sgRes, rtRes] = await Promise.all([
        fetch(`${API_BASE_URL}/api/v1/networks`, { headers: authHeaders }).then((r) => r.json()),
        fetch(`${API_BASE_URL}/api/v1/subnets`, { headers: authHeaders }).then((r) => r.json()),
        fetch(`${API_BASE_URL}/api/v1/security-groups`, { headers: authHeaders }).then((r) => r.json()),
        fetch(`${API_BASE_URL}/api/v1/route-tables`, { headers: authHeaders }).then((r) => r.json()),
      ])

      if (vpcRes && vpcRes.data) setVpcs(vpcRes.data)
      if (subRes && subRes.data) setSubnets(subRes.data)
      if (sgRes && sgRes.data) setSecurityGroups(sgRes.data)
      if (rtRes && rtRes.data) setRouteTables(rtRes.data)
    } catch {
      // Ignore network errors in dev mode
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreateVpc = async () => {
    setIsSubmitting(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/networks`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({
          organizationId: 'org-default',
          projectId: 'proj-default',
          name: newVpcName || 'anarva-vpc-prod',
          regionId: newVpcRegion,
          cidr: newVpcCidr,
        }),
      }).then((r) => r.json())

      if (res && res.data) {
        setVpcs([...vpcs, res.data])
        setIsVpcModalOpen(false)
        setNewVpcName('')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleCreateSubnet = async () => {
    setIsSubmitting(true)
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/subnets`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({
          organizationId: 'org-default',
          projectId: 'proj-default',
          vpcId: newSubnetVpcId || vpcs[0]?.id || 'vpc-101',
          name: newSubnetName || 'subnet-private-1',
          cidr: newSubnetCidr,
          zone: newSubnetZone,
          type: newSubnetType,
        }),
      }).then((r) => r.json())

      if (res && res.data) {
        setSubnets([...subnets, res.data])
        setIsSubnetModalOpen(false)
        setNewSubnetName('')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const tabs: TabItem[] = [
    { id: 'vpcs', label: `Anarva VPCs (${vpcs.length})` },
    { id: 'subnets', label: `Subnets (${subnets.length})` },
    { id: 'security-groups', label: `Security Groups (${securityGroups.length})` },
    { id: 'route-tables', label: `Route Tables (${routeTables.length})` },
    { id: 'ipam', label: 'IPAM & CIDR Allocations' },
  ]

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-gray-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Anarva Networking Control Plane</h1>
          <p className="text-sm text-gray-400 mt-1">
            Manage tenant-isolated Anarva VPCs, Subnets, Security Groups, Routing and IP Address allocations.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <CloudButton variant="secondary" onClick={fetchData}>
            Refresh
          </CloudButton>
          <CloudButton variant="primary" onClick={() => setIsVpcModalOpen(true)}>
            + Create Anarva VPC
          </CloudButton>
        </div>
      </div>

      {/* Summary Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <CloudCard>
          <span className="text-xs text-gray-400 uppercase font-semibold">Active VPCs</span>
          <div className="text-2xl font-bold text-white mt-1">{vpcs.length}</div>
          <span className="text-xs text-emerald-400 mt-1 block">100% Tenant Isolated</span>
        </CloudCard>
        <CloudCard>
          <span className="text-xs text-gray-400 uppercase font-semibold">Provisioned Subnets</span>
          <div className="text-2xl font-bold text-white mt-1">{subnets.length}</div>
          <span className="text-xs text-blue-400 mt-1 block">Public & Private Ranges</span>
        </CloudCard>
        <CloudCard>
          <span className="text-xs text-gray-400 uppercase font-semibold">Security Groups</span>
          <div className="text-2xl font-bold text-white mt-1">{securityGroups.length}</div>
          <span className="text-xs text-purple-400 mt-1 block">Firewall Rules Active</span>
        </CloudCard>
        <CloudCard>
          <span className="text-xs text-gray-400 uppercase font-semibold">Route Tables</span>
          <div className="text-2xl font-bold text-white mt-1">{routeTables.length}</div>
          <span className="text-xs text-gray-400 mt-1 block">IGW & Local Routes</span>
        </CloudCard>
      </div>

      {/* Main Tabs */}
      <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {/* Tab 1: VPCs */}
      {activeTab === 'vpcs' && (
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <h2 className="text-lg font-semibold text-white">Anarva Virtual Private Clouds</h2>
            <CloudButton variant="secondary" onClick={() => setIsVpcModalOpen(true)}>
              + New VPC
            </CloudButton>
          </div>
          {vpcs.length === 0 && !isLoading ? (
            <CloudEmptyState
              title="No Anarva VPCs Found"
              description="Create a tenant-isolated Anarva VPC to begin provisioning subnets and resources."
              actionLabel="Create Anarva VPC"
              onAction={() => setIsVpcModalOpen(true)}
            />
          ) : (
            <div className="overflow-x-auto border border-gray-800 rounded-lg">
              <table className="w-full text-left text-sm text-gray-300">
                <thead className="bg-gray-900/60 text-gray-400 uppercase text-xs">
                  <tr>
                    <th className="px-4 py-3">VPC ID</th>
                    <th className="px-4 py-3">Name</th>
                    <th className="px-4 py-3">CIDR Block</th>
                    <th className="px-4 py-3">Region</th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3">DNS</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-800">
                  {vpcs.map((v) => (
                    <tr key={v.id} className="hover:bg-gray-800/40">
                      <td className="px-4 py-3 font-mono text-cyan-400 text-xs">{v.id}</td>
                      <td className="px-4 py-3 font-medium text-white">{v.name}</td>
                      <td className="px-4 py-3 font-mono text-xs text-emerald-400">{v.cidr}</td>
                      <td className="px-4 py-3 text-xs">{v.regionId}</td>
                      <td className="px-4 py-3">
                        <CloudStatus status={v.status} />
                      </td>
                      <td className="px-4 py-3 text-xs text-gray-400">{v.dnsEnabled ? 'Enabled' : 'Disabled'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Tab 2: Subnets */}
      {activeTab === 'subnets' && (
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <h2 className="text-lg font-semibold text-white">VPC Subnets</h2>
            <CloudButton variant="secondary" onClick={() => setIsSubnetModalOpen(true)}>
              + Create Subnet
            </CloudButton>
          </div>
          {subnets.length === 0 ? (
            <CloudEmptyState
              title="No Subnets Found"
              description="Subnets partition your Anarva VPC CIDR into public and private network segments."
              actionLabel="Create Subnet"
              onAction={() => setIsSubnetModalOpen(true)}
            />
          ) : (
            <div className="overflow-x-auto border border-gray-800 rounded-lg">
              <table className="w-full text-left text-sm text-gray-300">
                <thead className="bg-gray-900/60 text-gray-400 uppercase text-xs">
                  <tr>
                    <th className="px-4 py-3">Subnet ID</th>
                    <th className="px-4 py-3">VPC ID</th>
                    <th className="px-4 py-3">Name</th>
                    <th className="px-4 py-3">Subnet CIDR</th>
                    <th className="px-4 py-3">Type</th>
                    <th className="px-4 py-3">Zone</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-800">
                  {subnets.map((s) => (
                    <tr key={s.id} className="hover:bg-gray-800/40">
                      <td className="px-4 py-3 font-mono text-cyan-400 text-xs">{s.id}</td>
                      <td className="px-4 py-3 font-mono text-xs text-gray-400">{s.networkId}</td>
                      <td className="px-4 py-3 font-medium text-white">{s.name}</td>
                      <td className="px-4 py-3 font-mono text-xs text-emerald-400">{s.cidr}</td>
                      <td className="px-4 py-3 text-xs font-semibold">
                        <span className={s.type === 'PUBLIC' ? 'text-amber-400' : 'text-blue-400'}>{s.type}</span>
                      </td>
                      <td className="px-4 py-3 text-xs">{s.availabilityZone}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Create VPC Modal */}
      <CloudModal
        isOpen={isVpcModalOpen}
        onClose={() => setIsVpcModalOpen(false)}
        title="Create Anarva VPC"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase mb-1">VPC Name</label>
            <input
              type="text"
              value={newVpcName}
              onChange={(e) => setNewVpcName(e.target.value)}
              placeholder="e.g. anarva-vpc-prod"
              className="w-full bg-gray-900 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:border-cyan-500"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase mb-1">IPv4 CIDR Block</label>
            <input
              type="text"
              value={newVpcCidr}
              onChange={(e) => setNewVpcCidr(e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded p-2 text-sm text-white font-mono focus:outline-none focus:border-cyan-500"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase mb-1">Region</label>
            <select
              value={newVpcRegion}
              onChange={(e) => setNewVpcRegion(e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none"
            >
              <option value="us-east-1">us-east-1 (N. Virginia)</option>
              <option value="ap-south-1">ap-south-1 (Mumbai)</option>
              <option value="eu-west-1">eu-west-1 (Ireland)</option>
            </select>
          </div>
          <div className="flex justify-end gap-3 mt-6">
            <CloudButton variant="secondary" onClick={() => setIsVpcModalOpen(false)}>
              Cancel
            </CloudButton>
            <CloudButton variant="primary" onClick={handleCreateVpc} disabled={isSubmitting}>
              {isSubmitting ? 'Provisioning...' : 'Provision VPC'}
            </CloudButton>
          </div>
        </div>
      </CloudModal>

      {/* Create Subnet Modal */}
      <CloudModal
        isOpen={isSubnetModalOpen}
        onClose={() => setIsSubnetModalOpen(false)}
        title="Create Subnet"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase mb-1">Target VPC</label>
            <select
              value={newSubnetVpcId}
              onChange={(e) => setNewSubnetVpcId(e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded p-2 text-sm text-white font-mono focus:outline-none"
            >
              {vpcs.map((v) => (
                <option key={v.id} value={v.id}>
                  {v.name} ({v.id}) - {v.cidr}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase mb-1">Subnet Name</label>
            <input
              type="text"
              value={newSubnetName}
              onChange={(e) => setNewSubnetName(e.target.value)}
              placeholder="e.g. subnet-private-1a"
              className="w-full bg-gray-900 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none focus:border-cyan-500"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase mb-1">Subnet CIDR Block</label>
            <input
              type="text"
              value={newSubnetCidr}
              onChange={(e) => setNewSubnetCidr(e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded p-2 text-sm text-white font-mono focus:outline-none focus:border-cyan-500"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-400 uppercase mb-1">Subnet Type</label>
            <select
              value={newSubnetType}
              onChange={(e) => setNewSubnetType(e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded p-2 text-sm text-white focus:outline-none"
            >
              <option value="PRIVATE">PRIVATE (Internal Application & Database)</option>
              <option value="PUBLIC">PUBLIC (Internet Facing & Gateway)</option>
              <option value="ISOLATED">ISOLATED (Air-gapped Storage)</option>
            </select>
          </div>
          <div className="flex justify-end gap-3 mt-6">
            <CloudButton variant="secondary" onClick={() => setIsSubnetModalOpen(false)}>
              Cancel
            </CloudButton>
            <CloudButton variant="primary" onClick={handleCreateSubnet} disabled={isSubmitting}>
              {isSubmitting ? 'Creating...' : 'Create Subnet'}
            </CloudButton>
          </div>
        </div>
      </CloudModal>
    </div>
  )
}
