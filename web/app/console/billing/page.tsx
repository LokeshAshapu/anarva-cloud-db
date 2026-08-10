'use client'

import React from 'react'

export default function BillingPage() {
  const lineItems = [
    { service: 'Anarva Compute Units (ACU)', usage: '1.5 ACU • 720 Hours', rate: '$0.025 / ACU-hr', cost: '$18.00' },
    { service: 'Managed PostgreSQL Storage', usage: '20 GB Storage Allocation', rate: '$0.15 / GB-mo', cost: '$3.00' },
    { service: 'Anarva Object Storage (AOS)', usage: '2.4 GB Objects + 1,200 Requests', rate: '$0.02 / GB-mo', cost: '$0.48' },
    { service: 'Network Egress Bandwidth', usage: '0.85 GB Egress Transfer', rate: '$0.09 / GB', cost: '$0.08' },
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Billing & Cost Management</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Real-time resource usage breakdown, cost estimation, invoices, and budget alerts.</p>
        </div>

        <span className="px-3 py-1 bg-slate-900 border border-slate-800 text-xs text-slate-400 font-mono rounded-lg">
          Invoice Period: August 2026
        </span>
      </div>

      {/* Cost Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Current Month Accrued</div>
          <div className="text-3xl font-extrabold text-emerald-400 font-mono">$21.56</div>
          <div className="text-xs text-slate-400">Est. Month End: $28.50</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Active Plan Tier</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">ENTERPRISE PAYG</div>
          <div className="text-xs text-slate-400">Pay-As-You-Go Metering</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Budget Threshold Alert</div>
          <div className="text-3xl font-extrabold text-white font-mono">$50.00 / mo</div>
          <div className="text-xs text-emerald-400">Within Safe Limits (43%)</div>
        </div>
      </div>

      {/* Usage Line Items */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
        <h2 className="text-base font-bold text-white">Current Billing Cycle Line Items</h2>

        <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
          {lineItems.map((item, idx) => (
            <div key={idx} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <div className="font-bold text-white">{item.service}</div>
                <div className="text-slate-400 text-[11px] font-mono mt-0.5">{item.usage} • {item.rate}</div>
              </div>

              <div className="font-bold text-emerald-400 font-mono text-sm">{item.cost}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
