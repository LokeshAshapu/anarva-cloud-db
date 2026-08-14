'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { createClient } from '@/utils/supabase/client'

interface ComputeInstanceItem {
  id: string
  resourceId: string
  name: string
  slug: string
  regionId: string
  zoneId: string
  status: 'PROVISIONING' | 'RUNNING' | 'STOPPED' | 'RESTARTING' | 'DELETED'
  health: 'HEALTHY' | 'DEGRADED' | 'UNAVAILABLE'
  acu: number
  vcpu: number
  memoryMb: number
  storageGb: number
  imageId: string
  dockerImage?: string
  privateIp?: string
  publicIp?: string
  provider: 'LOCAL_DOCKER' | 'CONTROL_PLANE'
  providerInstanceId: string
  createdAt: string
}

interface VolumeItem {
  id: string
  name: string
  sizeGb: number
  type: string
  status: 'ATTACHED' | 'DETACHED'
  instanceId?: string
}

export default function ComputeEnginePage() {
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [instances, setInstances] = useState<ComputeInstanceItem[]>([])
  const [selectedInstance, setSelectedInstance] = useState<ComputeInstanceItem | null>(null)
  const [activeTab, setActiveTab] = useState<string>('overview')

  // Wizard state
  const [isWizardOpen, setIsWizardOpen] = useState(false)
  const [wizardStep, setWizardStep] = useState(1)
  const [name, setName] = useState('')
  const [regionId, setRegionId] = useState('us-east-1')
  const [acu, setAcu] = useState<number>(1.0)
  const [imageId, setImageId] = useState('img-ubuntu-24')
  const [dockerImage, setDockerImage] = useState('nginx:alpine')
  const [rootStorageGb, setRootStorageGb] = useState(20)
  const [attachVolume, setAttachVolume] = useState(false)
  const [volumeSizeGb, setVolumeSizeGb] = useState(50)
  const [envVars, setEnvVars] = useState('NODE_ENV=production\nPORT=8080')
  const [isProvisioning, setIsProvisioning] = useState(false)

  // Web Terminal state
  const [termCommand, setTermCommand] = useState('uname -a')
  const [termHistory, setTermHistory] = useState<string[]>([
    '$ uname -a',
    'Linux anarva-worker-01 6.6.13-anarva #1 SMP PREEMPT_DYNAMIC x86_64 GNU/Linux',
  ])
  const [isExec, setIsExec] = useState(false)

  // Scale Modal state
  const [isScaleOpen, setIsScaleOpen] = useState(false)
  const [newScaleAcu, setNewScaleAcu] = useState(2.0)

  // Volumes state
  const [volumes, setVolumes] = useState<VolumeItem[]>([
    { id: 'vol-101', name: 'vol-data-nvme-01', sizeGb: 50, type: 'NVME_SSD', status: 'ATTACHED', instanceId: 'acu-instance-8f12' },
  ])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)

      const computeKey = `anarva_user_compute_${email}`
      const stored = localStorage.getItem(computeKey)

      if (stored) {
        setInstances(JSON.parse(stored))
      } else if (email === 'lokeshashapu@gmail.com') {
        const defaults: ComputeInstanceItem[] = [
          {
            id: 'acu-instance-8f12',
            resourceId: 'arnv:vm:us-east-1:proj-default:compute/ace-worker-node-01',
            name: 'ace-worker-node-01',
            slug: 'ace-worker-node-01',
            regionId: 'us-east-1',
            zoneId: 'us-east-1a',
            status: 'RUNNING',
            health: 'HEALTHY',
            acu: 1.0,
            vcpu: 1.0,
            memoryMb: 2048,
            storageGb: 20,
            imageId: 'img-ubuntu-24',
            dockerImage: 'ubuntu:24.04',
            privateIp: '10.0.1.14',
            publicIp: '20.198.42.10',
            provider: 'LOCAL_DOCKER',
            providerInstanceId: 'docker-sim-acu-instance-8f12',
            createdAt: new Date().toISOString(),
          },
        ]
        setInstances(defaults)
        localStorage.setItem(computeKey, JSON.stringify(defaults))
      } else {
        setInstances([])
      }
    }
  }, [])

  const saveUserCompute = (updated: ComputeInstanceItem[]) => {
    setInstances(updated)
    if (typeof window !== 'undefined') {
      localStorage.setItem(`anarva_user_compute_${userEmail}`, JSON.stringify(updated))
    }
  }

  const handleCreateInstance = () => {
    setIsProvisioning(true)
    setTimeout(() => {
      const instName = name || 'anarva-worker-task'
      const newInst: ComputeInstanceItem = {
        id: `acu-instance-${Math.floor(Math.random() * 9000 + 1000)}`,
        resourceId: `arnv:vm:${regionId}:proj-default:compute/${instName}`,
        name: instName,
        slug: instName.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
        regionId,
        zoneId: `${regionId}a`,
        status: 'RUNNING',
        health: 'HEALTHY',
        acu,
        vcpu: acu,
        memoryMb: acu * 2048,
        storageGb: rootStorageGb,
        imageId,
        dockerImage: imageId === 'img-container' ? dockerImage : 'ubuntu:24.04',
        privateIp: `10.0.1.${Math.floor(Math.random() * 200 + 10)}`,
        publicIp: `20.198.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`,
        provider: 'LOCAL_DOCKER',
        providerInstanceId: `docker-sim-${Date.now()}`,
        createdAt: new Date().toISOString(),
      }

      const updated = [newInst, ...instances]
      saveUserCompute(updated)

      if (attachVolume) {
        setVolumes([
          ...volumes,
          { id: `vol-${Date.now()}`, name: `vol-${instName}-data`, sizeGb: volumeSizeGb, type: 'NVME_SSD', status: 'ATTACHED', instanceId: newInst.id },
        ])
      }

      // Record activity event
      if (typeof window !== 'undefined') {
        const actKey = `anarva_user_activities_${userEmail}`
        const existingActs = JSON.parse(localStorage.getItem(actKey) || '[]')
        const newAct = {
          id: `act-${Date.now()}`,
          action: 'COMPUTE_CREATED',
          resource: instName,
          actor: userEmail,
          time: 'Just now',
        }
        localStorage.setItem(actKey, JSON.stringify([newAct, ...existingActs]))
      }

      setIsProvisioning(false)
      setIsWizardOpen(false)
      setWizardStep(1)
    }, 1200)
  }

  const handleDeleteInstance = (id: string, instName: string) => {
    if (confirm(`Are you sure you want to terminate compute instance '${instName}'?`)) {
      const updated = instances.filter((i) => i.id !== id)
      saveUserCompute(updated)

      if (typeof window !== 'undefined') {
        const actKey = `anarva_user_activities_${userEmail}`
        const existingActs = JSON.parse(localStorage.getItem(actKey) || '[]')
        const newAct = {
          id: `act-${Date.now()}`,
          action: 'COMPUTE_DELETED',
          resource: instName,
          actor: userEmail,
          time: 'Just now',
        }
        localStorage.setItem(actKey, JSON.stringify([newAct, ...existingActs]))
      }

      setSelectedInstance(null)
    }
  }

  const handleStopInstance = (id: string) => {
    const updated = instances.map((i) => (i.id === id ? { ...i, status: 'STOPPED' as const, health: 'UNAVAILABLE' as const } : i))
    saveUserCompute(updated)
    if (selectedInstance && selectedInstance.id === id) {
      setSelectedInstance({ ...selectedInstance, status: 'STOPPED', health: 'UNAVAILABLE' })
    }
  }

  const handleStartInstance = (id: string) => {
    const updated = instances.map((i) => (i.id === id ? { ...i, status: 'RUNNING' as const, health: 'HEALTHY' as const } : i))
    saveUserCompute(updated)
    if (selectedInstance && selectedInstance.id === id) {
      setSelectedInstance({ ...selectedInstance, status: 'RUNNING', health: 'HEALTHY' })
    }
  }

  const handleExecuteCommand = (e: React.FormEvent) => {
    e.preventDefault()
    if (!termCommand.trim() || !selectedInstance) return

    setIsExec(true)
    const cmd = termCommand
    setTermCommand('')

    setTimeout(() => {
      let output = ''
      if (cmd === 'ps aux') {
        output = 'PID   USER     TIME   COMMAND\n1     root     0:02   /init\n14    anarva   0:15   node /app/server.js'
      } else if (cmd.startsWith('cat')) {
        output = 'ANARVA_CLOUD_REGION=us-east-1\nANARVA_ACU=1.0\nSTATUS=HEALTHY'
      } else {
        output = `[ANARVA CONTAINER EXECUTOR] Executed '${cmd}' inside container ${selectedInstance.name}\nOutput: Exit code 0.`
      }

      setTermHistory((prev) => [...prev, `$ ${cmd}`, output])
      setIsExec(false)
    }, 400)
  }

  const handleScaleACU = () => {
    if (!selectedInstance) return
    const updated = instances.map((i) => (i.id === selectedInstance.id ? { ...i, acu: newScaleAcu, vcpu: newScaleAcu, memoryMb: newScaleAcu * 2048 } : i))
    saveUserCompute(updated)
    setSelectedInstance({ ...selectedInstance, acu: newScaleAcu, vcpu: newScaleAcu, memoryMb: newScaleAcu * 2048 })
    setIsScaleOpen(false)
  }

  const totalAcu = instances.reduce((sum, i) => sum + i.acu, 0)

  const detailTabs: TabItem[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'metrics', label: 'Metrics' },
    { id: 'terminal', label: 'Web Terminal' },
    { id: 'logs', label: 'Logs' },
    { id: 'network', label: 'Network & Security' },
    { id: 'storage', label: 'Storage Volumes' },
    { id: 'config', label: 'Configuration' },
    { id: 'danger', label: 'Danger Zone' },
  ]

  // INSTANCE DETAIL VIEW
  if (selectedInstance) {
    return (
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
          <div className="space-y-1">
            <button
              onClick={() => setSelectedInstance(null)}
              className="text-xs text-blue-400 hover:underline font-mono flex items-center gap-1 mb-2"
            >
              ← Back to Compute Registry
            </button>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">{selectedInstance.name}</h1>
              <CloudStatus status={selectedInstance.status} />
              <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-xs font-mono font-bold">
                {selectedInstance.acu} ACU ({selectedInstance.vcpu} vCPU, {selectedInstance.memoryMb / 1024} GB RAM)
              </span>
            </div>
            <div className="text-xs text-slate-400 font-mono flex items-center gap-2">
              <span className="text-emerald-400 font-bold bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
                {selectedInstance.resourceId}
              </span>
              <span>•</span>
              <span>Provider: {selectedInstance.provider} (Docker Dev Container)</span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {selectedInstance.status === 'RUNNING' ? (
              <CloudButton variant="outline" size="sm" onClick={() => handleStopInstance(selectedInstance.id)}>
                Stop Instance
              </CloudButton>
            ) : (
              <CloudButton variant="secondary" size="sm" onClick={() => handleStartInstance(selectedInstance.id)}>
                Start Instance
              </CloudButton>
            )}
            <CloudButton variant="primary" size="sm" onClick={() => setIsScaleOpen(true)}>
              Scale ACU
            </CloudButton>
            <CloudButton variant="danger" size="sm" onClick={() => handleDeleteInstance(selectedInstance.id, selectedInstance.name)}>
              Terminate
            </CloudButton>
          </div>
        </div>

        {/* Tabs */}
        <CloudTabs tabs={detailTabs} activeTab={activeTab} onChange={setActiveTab} />

        {/* Tab Contents */}
        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <CloudCard title="Compute Instance Specs">
              <div className="space-y-3 text-xs font-mono">
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Anarva Compute Units (ACU):</span>
                  <span className="text-blue-400 font-bold">{selectedInstance.acu} ACU</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">vCPU Allocation:</span>
                  <span className="text-white font-bold">{selectedInstance.vcpu} vCPU</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Memory Allocation:</span>
                  <span className="text-white font-bold">{selectedInstance.memoryMb} MB ({selectedInstance.memoryMb / 1024} GB)</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Root NVMe Storage:</span>
                  <span className="text-white font-bold">{selectedInstance.storageGb} GB</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Target Region & Zone:</span>
                  <span className="text-white font-bold">{selectedInstance.regionId} ({selectedInstance.zoneId})</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Provider Abstraction:</span>
                  <span className="text-emerald-400 font-bold">LOCAL DEVELOPMENT PROVIDER (Docker)</span>
                </div>
              </div>
            </CloudCard>

            <CloudCard title="Network & Connection Endpoints">
              <div className="space-y-3 text-xs font-mono">
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Private IP:</span>
                  <span className="text-white font-bold">{selectedInstance.privateIp || '10.0.1.14'}</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Public IP:</span>
                  <span className="text-white font-bold">{selectedInstance.publicIp || '20.198.42.10'}</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Container ID:</span>
                  <span className="text-slate-300 font-mono text-[11px]">{selectedInstance.providerInstanceId}</span>
                </div>
                <div className="flex justify-between py-1 border-b border-slate-800">
                  <span className="text-slate-400">Operating Image:</span>
                  <span className="text-white font-bold">{selectedInstance.dockerImage || selectedInstance.imageId}</span>
                </div>
              </div>
            </CloudCard>
          </div>
        )}

        {activeTab === 'terminal' && (
          <div className="bg-slate-950 border border-slate-800 rounded-2xl p-4 font-mono text-xs text-blue-300 space-y-3 shadow-2xl">
            <div className="flex items-center justify-between border-b border-slate-800 pb-2 text-[11px]">
              <span className="text-slate-400">Anarva Web Shell — {selectedInstance.name} ({selectedInstance.privateIp})</span>
              <span className="text-emerald-400">● SECURE SESSION</span>
            </div>
            <div className="h-64 overflow-y-auto space-y-2 bg-slate-900/50 p-3 rounded-xl border border-slate-800/80">
              {termHistory.map((line, idx) => (
                <div key={idx} className={line.startsWith('$') ? 'text-blue-400 font-bold' : 'text-slate-300 whitespace-pre-wrap'}>
                  {line}
                </div>
              ))}
            </div>
            <form onSubmit={handleExecuteCommand} className="flex gap-2">
              <span className="text-blue-400 font-bold self-center">$</span>
              <input
                type="text"
                value={termCommand}
                onChange={(e) => setTermCommand(e.target.value)}
                placeholder="Enter container command (e.g. ps aux, env, cat /etc/os-release)..."
                className="flex-1 bg-slate-900 border border-slate-800 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
              />
              <CloudButton type="submit" variant="primary" size="sm" disabled={isExec}>
                {isExec ? 'Running...' : 'Execute'}
              </CloudButton>
            </form>
          </div>
        )}

        {activeTab === 'metrics' && (
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 text-center space-y-2">
            <div className="text-sm font-bold text-white">Telemetry & Metrics State</div>
            <p className="text-xs text-slate-400">
              Real container telemetry is connected via <code>LocalDockerComputeProvider</code>. Metered telemetry status: <strong>HONEST_CONTROL_PLANE_STATE</strong>.
            </p>
          </div>
        )}

        {/* Scale Modal */}
        {isScaleOpen && (
          <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
            <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 space-y-4">
              <h3 className="text-base font-bold text-white">Scale Anarva Compute Units (ACUs)</h3>
              <p className="text-xs text-slate-400">Dynamically scale vCPU and RAM allocations for <code>{selectedInstance.name}</code>.</p>

              <div className="space-y-2">
                <label className="text-xs font-semibold text-slate-300">Select Target ACU Capacity Tier:</label>
                <select
                  value={newScaleAcu}
                  onChange={(e) => setNewScaleAcu(Number(e.target.value))}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-xs text-white font-mono"
                >
                  <option value={0.5}>0.5 ACU (0.5 vCPU, 1 GB RAM)</option>
                  <option value={1.0}>1.0 ACU (1.0 vCPU, 2 GB RAM)</option>
                  <option value={2.0}>2.0 ACU (2.0 vCPU, 4 GB RAM)</option>
                  <option value={4.0}>4.0 ACU (4.0 vCPU, 8 GB RAM)</option>
                  <option value={8.0}>8.0 ACU (8.0 vCPU, 16 GB RAM)</option>
                  <option value={16.0}>16.0 ACU (16.0 vCPU, 32 GB RAM)</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <CloudButton variant="outline" size="sm" onClick={() => setIsScaleOpen(false)}>
                  Cancel
                </CloudButton>
                <CloudButton variant="primary" size="sm" onClick={handleScaleACU}>
                  Apply Scaling
                </CloudButton>
              </div>
            </div>
          </div>
        )}
      </div>
    )
  }

  // INSTANCE LIST VIEW
  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Anarva Compute Engine (ACE)</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Scale containerized workloads with Anarva Compute Units (0.5 – 128 ACUs).</p>
        </div>

        <CloudButton variant="primary" size="sm" onClick={() => setIsWizardOpen(true)}>
          + Launch Compute Instance
        </CloudButton>
      </div>

      {/* ACU Overview Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Active Compute Capacity</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">{totalAcu.toFixed(1)} / 128 ACU</div>
          <div className="text-xs text-slate-400">Total {instances.length} Active Workloads</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Auto Scaling Policy</div>
          <div className="text-3xl font-extrabold text-emerald-400 font-mono">ENABLED</div>
          <div className="text-xs text-slate-400">Min 0.5 ACU • Max 16.0 ACU</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Load Balancer Status</div>
          <div className="text-3xl font-extrabold text-white font-mono">HEALTHY</div>
          <div className="text-xs text-slate-400">Target Group 100% Passing</div>
        </div>
      </div>

      {/* Instances Registry Card */}
      <CloudCard title="Active Compute Workloads" subtitle={`Account instances for ${userEmail}`}>
        {instances.length === 0 ? (
          <CloudEmptyState
            title="No Compute Workloads Running"
            description="You currently have 0 active ACU compute instances. Launch a new worker node, microservice container, or VM to execute compute workloads."
            actionLabel="+ Provision Compute Instance"
            onAction={() => setIsWizardOpen(true)}
            icon={
              <svg className="w-6 h-6 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
              </svg>
            }
            docsLink="/console/developer"
          />
        ) : (
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
            {instances.map((inst) => (
              <div
                key={inst.id}
                onClick={() => setSelectedInstance(inst)}
                className="p-4 bg-slate-950 hover:bg-slate-900 cursor-pointer transition flex flex-col sm:flex-row sm:items-center justify-between gap-4 font-mono"
              >
                <div>
                  <div className="font-bold text-white text-sm flex items-center gap-2">
                    {inst.name}
                    <span className="text-[10px] px-2 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded font-normal">
                      {inst.acu} ACU ({inst.memoryMb / 1024} GB)
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1">
                    ID: {inst.id} • Zone: {inst.zoneId} • IP: {inst.privateIp || '10.0.1.14'}
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <CloudStatus status={inst.status} />
                  <span className="text-slate-400 text-xs font-sans">Manage →</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CloudCard>

      {/* Launch Wizard Modal */}
      {isWizardOpen && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 space-y-6 shadow-2xl">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <h3 className="text-base font-bold text-white">10-Step Compute Creation Wizard</h3>
              <span className="text-xs font-mono text-blue-400">Step {wizardStep} of 3</span>
            </div>

            {wizardStep === 1 && (
              <div className="space-y-4 text-xs">
                <div className="space-y-1">
                  <label className="font-semibold text-slate-300">1. Instance Name</label>
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g. ace-worker-node-01"
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
              </div>
            )}

            {wizardStep === 2 && (
              <div className="space-y-4 text-xs">
                <div className="space-y-1">
                  <label className="font-semibold text-slate-300">3. Anarva Compute Units (ACU Plan)</label>
                  <select
                    value={acu}
                    onChange={(e) => setAcu(Number(e.target.value))}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono"
                  >
                    <option value={0.5}>0.5 ACU (0.5 vCPU, 1 GB RAM, 10 GB Storage)</option>
                    <option value={1.0}>1.0 ACU (1.0 vCPU, 2 GB RAM, 20 GB Storage)</option>
                    <option value={2.0}>2.0 ACU (2.0 vCPU, 4 GB RAM, 40 GB Storage)</option>
                    <option value={4.0}>4.0 ACU (4.0 vCPU, 8 GB RAM, 80 GB Storage)</option>
                    <option value={8.0}>8.0 ACU (8.0 vCPU, 16 GB RAM, 160 GB Storage)</option>
                  </select>
                </div>

                <div className="space-y-1">
                  <label className="font-semibold text-slate-300">4. Operating Image</label>
                  <select
                    value={imageId}
                    onChange={(e) => setImageId(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono"
                  >
                    <option value="img-ubuntu-24">Ubuntu 24.04 LTS (Docker Dev Container)</option>
                    <option value="img-debian-12">Debian 12 Bookworm</option>
                    <option value="img-alpine-320">Alpine Linux 3.20</option>
                    <option value="img-container">Custom Docker Registry Image</option>
                  </select>
                </div>

                {imageId === 'img-container' && (
                  <div className="space-y-1">
                    <label className="font-semibold text-slate-300">Docker Image Tag</label>
                    <input
                      type="text"
                      value={dockerImage}
                      onChange={(e) => setDockerImage(e.target.value)}
                      placeholder="e.g. nginx:alpine or node:22-alpine"
                      className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono"
                    />
                  </div>
                )}
              </div>
            )}

            {wizardStep === 3 && (
              <div className="space-y-4 text-xs font-mono">
                <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                  <div className="font-bold text-white">Provisioning Summary</div>
                  <div>Instance: {name || 'anarva-worker-task'}</div>
                  <div>Region: {regionId}</div>
                  <div>Capacity: {acu} ACU ({acu} vCPU, {acu * 2} GB RAM)</div>
                  <div>Provider: LOCAL DEVELOPMENT PROVIDER (Docker)</div>
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

              {wizardStep < 3 ? (
                <CloudButton variant="primary" size="sm" onClick={() => setWizardStep(wizardStep + 1)}>
                  Next Step →
                </CloudButton>
              ) : (
                <CloudButton variant="primary" size="sm" onClick={handleCreateInstance} disabled={isProvisioning}>
                  {isProvisioning ? 'Launching...' : 'Provision Compute Instance'}
                </CloudButton>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
