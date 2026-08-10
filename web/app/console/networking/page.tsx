'use client'

import React from 'react'

export default function NetworkingPage() {
  const vpcs = [
    { id: 'vpc-0a1b2c3d', name: 'Primary Production VPC', cidr: '10.0.0.0/16', region: 'us-east-1', subnets: 3, status: 'Active (Simulation Layer)' },
    { id: 'vpc-4e5f6a7b', name: 'Staging & Sandbox VPC', cidr: '172.16.0.0/16', region: 'us-east-1', subnets: 2, status: 'Active (Simulation Layer)' },
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Virtual Private Cloud (VPC) & Networking</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Configure isolated cloud networks, subnets, firewalls, and load balancers.</p>
        </div>

        <span className="px-3 py-1 bg-amber-500/10 text-amber-400 border border-amber-500/20 text-xs font-mono rounded-lg">
          Networking Control Layer • Planned Simulation Mode
        </span>
      </div>

      {/* Networking Overview Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Active VPC Networks</div>
          <div className="text-3xl font-extrabold text-white font-mono">2</div>
          <div className="text-xs text-slate-400">Default Subnets: 5</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Security Groups</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">4 Rules</div>
          <div className="text-xs text-slate-400">Ingress: Port 5432, 443, 80</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Application Load Balancers</div>
          <div className="text-3xl font-extrabold text-emerald-400 font-mono">1 Active</div>
          <div className="text-xs text-slate-400">TLS 1.3 Termination Active</div>
        </div>
      </div>

      {/* VPC Table */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
        <h2 className="text-base font-bold text-white">VPC Networks</h2>

        <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
          {vpcs.map((vpc) => (
            <div key={vpc.id} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3 font-mono">
              <div>
                <div className="font-bold text-white font-sans">{vpc.name}</div>
                <div className="text-slate-400 text-[11px] mt-0.5">{vpc.id} • CIDR: {vpc.cidr}</div>
              </div>

              <div className="flex items-center gap-4">
                <span className="text-slate-400">{vpc.subnets} Subnets ({vpc.region})</span>
                <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
                  {vpc.status}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
