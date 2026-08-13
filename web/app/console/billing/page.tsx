'use client'

import React, { useState, useEffect } from 'react'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudMetric } from '@/components/cloud/CloudMetric'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { API_BASE_URL } from '@/lib/api'

interface QuotaItem {
  id: string
  resourceType: string
  metric: string
  limit: number
  currentUsage: number
  unit: string
  status: 'AVAILABLE' | 'NEAR_LIMIT' | 'EXCEEDED' | 'UNLIMITED'
}

interface InvoiceLineItem {
  id: string
  resourceId: string
  description: string
  quantity: number
  unit: string
  unitPrice: number
  amount: number
  usageQuality: string
}

interface InvoiceItem {
  id: string
  invoiceNumber: string
  subtotal: number
  total: number
  status: string
  pricingVersion: string
  realityLabel: string
  issuedAt: string
  lines: InvoiceLineItem[]
}

interface PricingComponentItem {
  id: string
  resourceType: string
  metric: string
  unit: string
  unitPrice: number
  includedQuantity: number
}

export default function BillingPage() {
  const [activeTab, setActiveTab] = useState('overview')
  const [userEmail, setUserEmail] = useState('user@anarva.io')

  // Quotas State
  const [quotas, setQuotas] = useState<QuotaItem[]>([
    { id: 'q-1', resourceType: 'COMPUTE', metric: 'compute.acu', limit: 32.0, currentUsage: 4.0, unit: 'ACU', status: 'AVAILABLE' },
    { id: 'q-2', resourceType: 'STORAGE', metric: 'storage.capacity', limit: 500.0, currentUsage: 28.5, unit: 'GB', status: 'AVAILABLE' },
    { id: 'q-3', resourceType: 'DATABASE', metric: 'database.count', limit: 5.0, currentUsage: 2.0, unit: 'INSTANCES', status: 'AVAILABLE' },
    { id: 'q-4', resourceType: 'NETWORK', metric: 'network.vpc', limit: 3.0, currentUsage: 1.0, unit: 'NETWORKS', status: 'AVAILABLE' },
  ])

  // Invoices State
  const [invoices, setInvoices] = useState<InvoiceItem[]>([
    {
      id: 'inv-2026-08',
      invoiceNumber: 'INV-202608-001',
      subtotal: 21.48,
      total: 21.48,
      status: 'DRAFT (SIMULATED)',
      pricingVersion: 'v1.0.0',
      realityLabel: 'SIMULATED_BILLING (NOT BILLABLE)',
      issuedAt: new Date().toISOString(),
      lines: [
        { id: 'l-1', resourceId: 'ace-worker-node-01', description: 'Compute Instance ace-worker-node-01 (1.0 ACU * 720 Hours)', quantity: 720, unit: 'ACU-hour', unitPrice: 0.025, amount: 18.00, usageQuality: 'SIMULATED' },
        { id: 'l-2', resourceId: 'anarva-media-assets', description: 'Object Storage Bucket (23.2 GB-month)', quantity: 23.2, unit: 'GB-month', unitPrice: 0.15, amount: 3.48, usageQuality: 'SIMULATED' },
      ],
    },
  ])

  // Pricing Catalog
  const [pricingComponents] = useState<PricingComponentItem[]>([
    { id: 'p-1', resourceType: 'COMPUTE', metric: 'compute.runtime', unit: 'ACU-hour', unitPrice: 0.025, includedQuantity: 0 },
    { id: 'p-2', resourceType: 'DATABASE', metric: 'database.runtime', unit: 'Instance-hour', unitPrice: 0.045, includedQuantity: 0 },
    { id: 'p-3', resourceType: 'STORAGE', metric: 'storage.capacity', unit: 'GB-month', unitPrice: 0.15, includedQuantity: 0 },
    { id: 'p-4', resourceType: 'NETWORK', metric: 'network.egress', unit: 'GB', unitPrice: 0.09, includedQuantity: 10 },
  ])

  // Cost Estimator Wizard State
  const [estResourceType, setEstResourceType] = useState('COMPUTE')
  const [estAcu, setEstAcu] = useState(2.0)
  const [estHours, setEstHours] = useState(720)
  const [estOutput, setEstOutput] = useState<{ cost: number; rate: string; text: string } | null>(null)

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)
    }

    loadBillingData()
  }, [])

  async function loadBillingData() {
    try {
      const qRes = await fetch(`${API_BASE_URL}/api/v1/billing/quotas`).catch(() => null)
      if (qRes && qRes.ok) {
        const body = await qRes.json()
        if (body.data) setQuotas(body.data)
      }
    } catch (e) {}
  }

  const handleCalculateEstimate = (e: React.FormEvent) => {
    e.preventDefault()
    let rate = 0.025
    if (estResourceType === 'DATABASE') rate = 0.045
    if (estResourceType === 'STORAGE') rate = 0.015

    const cost = estAcu * rate * estHours
    setEstOutput({
      cost: Number(cost.toFixed(2)),
      rate: `$${rate} / unit-hour`,
      text: `${estAcu} Units * ${estHours} Hours @ $${rate}/hr = $${cost.toFixed(2)} USD`,
    })
  }

  const tabs: TabItem[] = [
    { id: 'overview', label: 'Billing Overview & Accrued Usage' },
    { id: 'quotas', label: 'Resource Quota Manager' },
    { id: 'estimator', label: 'Cost Estimator Wizard' },
    { id: 'invoices', label: 'Invoices & Line Items' },
    { id: 'pricing', label: 'Pricing Catalog (v1.0.0)' },
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono text-slate-400">BILLING ENGINE:</span>
            <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-xs font-mono font-bold">
              PAYG PRICING v1.0.0
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight mt-1">Billing & Quota Management</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-0.5">
            Resource usage metering, atomic quota limits, pre-provisioning cost estimation, and draft invoice generation for <strong className="text-slate-200">{userEmail}</strong>.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="px-3 py-1 bg-slate-900 border border-slate-800 text-xs text-slate-400 font-mono rounded-lg">
            Billing Period: August 2026
          </span>
        </div>
      </div>

      {/* Non-Billable Reality Banner */}
      <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-xl font-mono text-xs text-amber-400 flex items-center justify-between gap-4">
        <div>
          <strong className="font-bold uppercase text-amber-300">🛡 SIMULATED BILLING & LOCAL DEVELOPMENT MODE:</strong>
          <span className="ml-2 text-slate-300 text-[11px]">
            No commercial payment gateway (Stripe/Razorpay) is connected. All accrued usage is local development & simulated billing estimation.
          </span>
        </div>
        <span className="px-2.5 py-1 bg-amber-500/20 text-amber-300 border border-amber-500/30 rounded text-[10px] font-bold">
          NON_BILLABLE
        </span>
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <CloudMetric label="Current Month Accrued" value="$21.48" subtext="SIMULATED_BILLING (NON_BILLABLE)" trend="ACCUMULATING" trendType="positive" />
        <CloudMetric label="Active Pricing Version" value="v1.0.0" subtext="Anarva PAYG Standard" trend="ACTIVE" trendType="positive" />
        <CloudMetric label="Compute ACU Quota" value="4.0 / 32.0 ACU" subtext="12.5% Allocation Used" trend="AVAILABLE" trendType="positive" />
        <CloudMetric label="Storage Quota" value="28.5 / 500 GB" subtext="5.7% Allocation Used" trend="AVAILABLE" trendType="positive" />
      </div>

      {/* Navigation Tabs */}
      <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Current Billing Cycle Accrued Usage & Line Items">
            <div className="space-y-4">
              <div className="p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded-xl text-[11px]">
                ℹ Usage records are metered continuously from local Docker container lifecycle events and storage allocation.
              </div>

              <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
                {invoices[0].lines.map((item) => (
                  <div key={item.id} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-white text-sm">{item.description}</span>
                        <span className="px-2 py-0.5 rounded bg-slate-800 text-slate-300 text-[10px]">{item.usageQuality}</span>
                      </div>
                      <div className="text-slate-400 text-[11px] font-mono mt-1">
                        Resource: {item.resourceId} • {item.quantity} {item.unit} @ ${item.unitPrice}/{item.unit}
                      </div>
                    </div>

                    <div className="font-bold text-emerald-400 font-mono text-sm">${item.amount.toFixed(2)} USD</div>
                  </div>
                ))}
              </div>
            </div>
          </CloudCard>
        </div>
      )}

      {/* Quotas Tab */}
      {activeTab === 'quotas' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Organization & Project Resource Quotas (Atomic Concurrency Engine)">
            <div className="space-y-4">
              <div className="p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded-xl text-[11px]">
                ℹ Resource quotas are enforced atomically using mutex-locked concurrency validation before any infrastructure provisioning operation succeeds.
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {quotas.map((q) => {
                  const percent = Math.min(100, Math.round((q.currentUsage / q.limit) * 100))
                  return (
                    <div key={q.id} className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-3">
                      <div className="flex items-center justify-between">
                        <div>
                          <div className="font-bold text-white text-sm">{q.metric}</div>
                          <div className="text-[10px] text-slate-400 uppercase">Resource: {q.resourceType}</div>
                        </div>
                        <span
                          className={`px-2.5 py-1 rounded text-[10px] font-bold ${
                            q.status === 'AVAILABLE'
                              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                              : 'bg-red-500/10 text-red-400 border border-red-500/20'
                          }`}
                        >
                          {q.status}
                        </span>
                      </div>

                      <div className="space-y-1">
                        <div className="flex justify-between text-[11px]">
                          <span className="text-slate-400">Current Usage:</span>
                          <span className="text-white font-bold">
                            {q.currentUsage} / {q.limit} {q.unit} ({percent}%)
                          </span>
                        </div>
                        <div className="w-full h-2 bg-slate-900 rounded-full overflow-hidden">
                          <div
                            className={`h-full ${percent > 85 ? 'bg-red-500' : 'bg-emerald-500'}`}
                            style={{ width: `${percent}%` }}
                          />
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          </CloudCard>
        </div>
      )}

      {/* Cost Estimator Tab */}
      {activeTab === 'estimator' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Pre-Provisioning Cost Estimator Wizard">
            <form onSubmit={handleCalculateEstimate} className="space-y-5">
              <div className="p-3 bg-amber-500/10 border border-amber-500/20 text-amber-400 rounded-xl text-[11px]">
                ℹ Calculate projected infrastructure costs before provisioning workloads. Estimates use Pricing Plan v1.0.0.
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div>
                  <label className="block text-slate-400 mb-1 text-[10px]">RESOURCE TYPE</label>
                  <select
                    value={estResourceType}
                    onChange={(e) => setEstResourceType(e.target.value)}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  >
                    <option value="COMPUTE">COMPUTE (ACU Workload)</option>
                    <option value="DATABASE">DATABASE (PostgreSQL Cluster)</option>
                    <option value="STORAGE">STORAGE (Object Bucket)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 text-[10px]">ACU / CAPACITY UNITS</label>
                  <input
                    type="number"
                    step="0.5"
                    min="0.5"
                    max="64"
                    value={estAcu}
                    onChange={(e) => setEstAcu(Number(e.target.value))}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 text-[10px]">EXPECTED RUNTIME (HOURS)</label>
                  <input
                    type="number"
                    step="1"
                    min="1"
                    max="8760"
                    value={estHours}
                    onChange={(e) => setEstHours(Number(e.target.value))}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none focus:border-blue-500"
                  />
                </div>
              </div>

              <div className="flex justify-end">
                <CloudButton variant="primary" size="sm" type="submit">
                  Calculate Estimated Cost
                </CloudButton>
              </div>

              {estOutput && (
                <div className="p-5 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-400 text-xs">PROJECTED ESTIMATED COST:</span>
                    <span className="text-2xl font-extrabold text-emerald-400 font-mono">${estOutput.cost.toFixed(2)} USD</span>
                  </div>
                  <div className="text-[11px] text-slate-300">{estOutput.text}</div>
                  <div className="text-[10px] text-amber-400 font-bold mt-1">
                    Reality Label: NOT_BILLABLE (PRE-PROVISIONING ESTIMATE ONLY)
                  </div>
                </div>
              )}
            </form>
          </CloudCard>
        </div>
      )}

      {/* Invoices Tab */}
      {activeTab === 'invoices' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Billing Period Draft Invoices">
            <div className="space-y-4">
              {invoices.map((inv) => (
                <div key={inv.id} className="p-5 bg-slate-950 border border-slate-800 rounded-xl space-y-4">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-800 pb-3">
                    <div>
                      <div className="font-bold text-white text-base">{inv.invoiceNumber}</div>
                      <div className="text-[10px] text-slate-400 mt-0.5">
                        Pricing Version: {inv.pricingVersion} • Reality: {inv.realityLabel}
                      </div>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="px-2.5 py-1 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded text-xs font-bold">
                        {inv.status}
                      </span>
                      <span className="text-xl font-extrabold text-white font-mono">${inv.total.toFixed(2)} {inv.currency}</span>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="text-[11px] font-bold text-slate-400">INVOICE LINE ITEMS</div>
                    {inv.lines.map((l) => (
                      <div key={l.id} className="p-3 bg-slate-900/60 border border-slate-800/80 rounded-lg flex items-center justify-between">
                        <div>
                          <div className="font-bold text-slate-200">{l.description}</div>
                          <div className="text-[10px] text-slate-400">Quality: {l.usageQuality}</div>
                        </div>
                        <div className="font-bold text-emerald-400">${l.amount.toFixed(2)}</div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </CloudCard>
        </div>
      )}

      {/* Pricing Catalog Tab */}
      {activeTab === 'pricing' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Anarva PAYG Pricing Catalog (Version v1.0.0)">
            <div className="space-y-4">
              <div className="p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded-xl text-[11px]">
                ℹ Historical invoices are immutable and always preserve the pricing version active during that billing cycle.
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {pricingComponents.map((c) => (
                  <div key={c.id} className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="font-bold text-white text-sm">{c.resourceType}</div>
                      <span className="px-2 py-0.5 bg-slate-800 text-slate-300 rounded text-[10px]">{c.unit}</span>
                    </div>
                    <div className="text-2xl font-extrabold text-emerald-400 font-mono">${c.unitPrice} / {c.unit}</div>
                    <div className="text-[10px] text-slate-400">
                      Metric: {c.metric} • Free Tier: {c.includedQuantity} {c.unit} included
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </CloudCard>
        </div>
      )}
    </div>
  )
}
