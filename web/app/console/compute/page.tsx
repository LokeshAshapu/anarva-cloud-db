'use client'

import React, { useState, useEffect } from 'react'

export default function ComputeEnginePage() {
  const [acu, setAcu] = useState<number>(1.0)
  const [region, setRegion] = useState<string>('us-east-1')
  const [autoScaling, setAutoScaling] = useState<boolean>(true)
  const [isProvisioning, setIsProvisioning] = useState<boolean>(false)

  // User Email & Compute State
  const [userEmail, setUserEmail] = useState('lokeshashapu@gmail.com')
  const [instances, setInstances] = useState<any[]>([])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'lokeshashapu@gmail.com'
      setUserEmail(email)

      const computeKey = `anarva_user_compute_${email}`
      const stored = localStorage.getItem(computeKey)

      if (stored) {
        setInstances(JSON.parse(stored))
      } else if (email === 'lokeshashapu@gmail.com') {
        const defaults = [
          { id: 'acu-instance-8f12', name: 'API Gateway Worker', acu: 1.0, memory: '2 GB', vCPU: 1, status: 'RUNNING', zone: 'us-east-1a' },
        ]
        setInstances(defaults)
        localStorage.setItem(computeKey, JSON.stringify(defaults))
      } else {
        setInstances([])
      }
    }
  }, [])

  const saveUserCompute = (updated: any[]) => {
    setInstances(updated)
    if (typeof window !== 'undefined') {
      localStorage.setItem(`anarva_user_compute_${userEmail}`, JSON.stringify(updated))
    }
  }

  const handleLaunchCompute = (e: React.FormEvent) => {
    e.preventDefault()
    setIsProvisioning(true)
    setTimeout(() => {
      const instName = `Worker Instance ${Math.floor(Math.random() * 900 + 100)}`
      const newInst = {
        id: `acu-instance-${Math.floor(Math.random() * 9000 + 1000)}`,
        name: instName,
        acu: acu,
        memory: `${acu * 2} GB`,
        vCPU: acu,
        status: 'RUNNING',
        zone: `${region}a`,
      }
      const updated = [newInst, ...instances]
      saveUserCompute(updated)

      // Record activity event
      if (typeof window !== 'undefined') {
        const actKey = `anarva_user_activities_${userEmail}`
        const existingActs = JSON.parse(localStorage.getItem(actKey) || '[]')
        const newAct = {
          id: `act-${Date.now()}`,
          action: 'RESOURCE_STARTED',
          resource: instName,
          actor: userEmail,
          time: 'Just now',
        }
        localStorage.setItem(actKey, JSON.stringify([newAct, ...existingActs]))
      }

      setIsProvisioning(false)
    }, 1200)
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Anarva Compute Engine (ACE)</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Scale workloads dynamically with Anarva Compute Units (0.5 – 128 ACUs).</p>
        </div>
      </div>

      {/* ACU Overview Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Active Compute Capacity</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">1.5 / 128 ACU</div>
          <div className="text-xs text-slate-400">Total 2 Active Workloads</div>
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

      {/* Compute Launch Wizard Form */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-6">
        <h2 className="text-base font-bold text-white">Launch Anarva Compute Instance</h2>

        <form onSubmit={handleLaunchCompute} className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 text-xs">
          <div className="space-y-1">
            <label className="text-slate-400 font-semibold">Instance Name</label>
            <input
              type="text"
              defaultValue="Serverless Worker Task"
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500"
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-slate-400 font-semibold">Target Region</label>
            <select
              value={region}
              onChange={(e) => setRegion(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500"
            >
              <option value="us-east-1">us-east-1 (N. Virginia)</option>
              <option value="eu-west-1">eu-west-1 (Frankfurt)</option>
              <option value="ap-south-1">ap-south-1 (Mumbai)</option>
            </select>
          </div>

          <div className="space-y-1">
            <label className="text-slate-400 font-semibold">Anarva Compute Units (ACU)</label>
            <select
              value={acu}
              onChange={(e) => setAcu(Number(e.target.value))}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500"
            >
              <option value={0.5}>0.5 ACU (0.5 vCPU, 1 GB RAM)</option>
              <option value={1.0}>1.0 ACU (1.0 vCPU, 2 GB RAM)</option>
              <option value={2.0}>2.0 ACU (2.0 vCPU, 4 GB RAM)</option>
              <option value={4.0}>4.0 ACU (4.0 vCPU, 8 GB RAM)</option>
              <option value={8.0}>8.0 ACU (8.0 vCPU, 16 GB RAM)</option>
              <option value={16.0}>16.0 ACU (16.0 vCPU, 32 GB RAM)</option>
            </select>
          </div>

          <div className="flex items-end">
            <button
              type="submit"
              disabled={isProvisioning}
              className="w-full py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition shadow-lg shadow-blue-600/20 disabled:opacity-50"
            >
              {isProvisioning ? 'Provisioning ACU...' : 'Launch Compute Unit'}
            </button>
          </div>
        </form>
      </div>

      {/* Active Instances Table */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
        <h2 className="text-base font-bold text-white">Active Compute Workloads</h2>

        <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
          {instances.map((inst) => (
            <div key={inst.id} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3 font-mono">
              <div>
                <div className="font-bold text-white font-sans">{inst.name}</div>
                <div className="text-slate-400 text-[11px] mt-0.5">{inst.id} • Zone: {inst.zone}</div>
              </div>

              <div className="flex items-center gap-4">
                <span className="text-blue-400 font-bold">{inst.acu} ACU ({inst.memory})</span>
                <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  {inst.status}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
