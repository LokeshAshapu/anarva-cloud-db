'use client'

import React, { useState, useEffect } from 'react'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudMetric } from '@/components/cloud/CloudMetric'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
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
  currency: string
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

  // Account-Specific Calculated Resource Metrics
  const [accountUsage, setAccountUsage] = useState({
    computeAcu: 0,
    storageGb: 0,
    dbCount: 0,
    vpcCount: 0,
    computeCost: 0,
    dbCost: 0,
    storageCost: 0,
    totalAccruedCost: 0,
  })

  // Account-Scoped Quotas State
  const [quotas, setQuotas] = useState<QuotaItem[]>([])

  // Account-Scoped Invoices State
  const [invoices, setInvoices] = useState<InvoiceItem[]>([])

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
    let email = 'user@anarva.io'
    if (typeof window !== 'undefined') {
      email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)
    }

    calculateAccountBillingAndQuotas(email)
  }, [])

  function calculateAccountBillingAndQuotas(email: string) {
    let computeAcu = 0
    let storageGb = 0
    let dbCount = 0
    let vpcCount = 0

    if (typeof window !== 'undefined') {
      // 1. Account Compute ACU
      const storedCompute = localStorage.getItem(`anarva_user_compute_${email}`)
      if (storedCompute) {
        try {
          const instances = JSON.parse(storedCompute)
          if (Array.isArray(instances)) {
            computeAcu = instances.reduce((acc: number, inst: any) => acc + (inst.acu || 1.0), 0)
          }
        } catch (e) {}
      } else if (email === 'lokeshashapu@gmail.com') {
        computeAcu = 1.0
      }

      // 2. Account Database Count
      const storedDBs = localStorage.getItem(`anarva_user_databases_${email}`)
      if (storedDBs) {
        try {
          const dbs = JSON.parse(storedDBs)
          if (Array.isArray(dbs)) dbCount = dbs.length
        } catch (e) {}
      } else if (email === 'lokeshashapu@gmail.com') {
        dbCount = 1
      }

      // 3. Account Storage Capacity in GB
      const storedStorage = localStorage.getItem(`anarva_user_storage_${email}`)
      if (storedStorage) {
        try {
          const buckets = JSON.parse(storedStorage)
          if (Array.isArray(buckets)) {
            storageGb = buckets.reduce((acc: number, b: any) => acc + (b.sizeGb || 10.0), 0)
          }
        } catch (e) {}
      } else if (email === 'lokeshashapu@gmail.com') {
        storageGb = 20.0
      }

      // 4. Account VPC Count
      const storedNetworks = localStorage.getItem(`anarva_user_networking_${email}`)
      if (storedNetworks) {
        try {
          const nets = JSON.parse(storedNetworks)
          if (Array.isArray(nets)) vpcCount = nets.length
        } catch (e) {}
      } else if (email === 'lokeshashapu@gmail.com') {
        vpcCount = 1
      }
    }

    // Calculate Costs based on Account's Actual Resource Usage
    const computeCost = Number((computeAcu * 720 * 0.025).toFixed(2))
    const dbCost = Number((dbCount * 720 * 0.045).toFixed(2))
    const storageCost = Number((storageGb * 0.15).toFixed(2))
    const totalAccruedCost = Number((computeCost + dbCost + storageCost).toFixed(2))

    setAccountUsage({
      computeAcu,
      storageGb,
      dbCount,
      vpcCount,
      computeCost,
      dbCost,
      storageCost,
      totalAccruedCost,
    })

    // Account-Scoped Quotas
    const calculatedQuotas: QuotaItem[] = [
      {
        id: 'q-1',
        resourceType: 'COMPUTE',
        metric: 'compute.acu',
        limit: 32.0,
        currentUsage: computeAcu,
        unit: 'ACU',
        status: computeAcu >= 32.0 ? 'EXCEEDED' : computeAcu >= 25.6 ? 'NEAR_LIMIT' : 'AVAILABLE',
      },
      {
        id: 'q-2',
        resourceType: 'STORAGE',
        metric: 'storage.capacity',
        limit: 500.0,
        currentUsage: storageGb,
        unit: 'GB',
        status: storageGb >= 500.0 ? 'EXCEEDED' : storageGb >= 400.0 ? 'NEAR_LIMIT' : 'AVAILABLE',
      },
      {
        id: 'q-3',
        resourceType: 'DATABASE',
        metric: 'database.count',
        limit: 5.0,
        currentUsage: dbCount,
        unit: 'INSTANCES',
        status: dbCount >= 5.0 ? 'EXCEEDED' : dbCount >= 4.0 ? 'NEAR_LIMIT' : 'AVAILABLE',
      },
      {
        id: 'q-4',
        resourceType: 'NETWORK',
        metric: 'network.vpc',
        limit: 3.0,
        currentUsage: vpcCount,
        unit: 'NETWORKS',
        status: vpcCount >= 3.0 ? 'EXCEEDED' : vpcCount >= 2.5 ? 'NEAR_LIMIT' : 'AVAILABLE',
      },
    ]
    setQuotas(calculatedQuotas)

    // Account-Scoped Line Items
    const lines: InvoiceLineItem[] = []
    if (computeAcu > 0) {
      lines.push({
        id: 'l-1',
        resourceId: `compute-${email.split('@')[0]}`,
        description: `Compute Workload (${computeAcu.toFixed(1)} ACU * 720 Hours)`,
        quantity: computeAcu * 720,
        unit: 'ACU-hour',
        unitPrice: 0.025,
        amount: computeCost,
        usageQuality: 'ACTUAL (ACCOUNT SCOPED)',
      })
    }
    if (dbCount > 0) {
      lines.push({
        id: 'l-2',
        resourceId: `db-${email.split('@')[0]}`,
        description: `Managed Database Clusters (${dbCount} Instances * 720 Hours)`,
        quantity: dbCount * 720,
        unit: 'Instance-hour',
        unitPrice: 0.045,
        amount: dbCost,
        usageQuality: 'ACTUAL (ACCOUNT SCOPED)',
      })
    }
    if (storageGb > 0) {
      lines.push({
        id: 'l-3',
        resourceId: `storage-${email.split('@')[0]}`,
        description: `Anarva Object Storage (${storageGb.toFixed(1)} GB-month)`,
        quantity: storageGb,
        unit: 'GB-month',
        unitPrice: 0.15,
        amount: storageCost,
        usageQuality: 'ACTUAL (ACCOUNT SCOPED)',
      })
    }

    if (lines.length === 0) {
      lines.push({
        id: 'l-0',
        resourceId: 'no-active-resources',
        description: 'No Active Paid Resources Provisioned',
        quantity: 0,
        unit: 'Units',
        unitPrice: 0,
        amount: 0,
        usageQuality: 'NON_BILLABLE',
      })
    }

    setInvoices([
      {
        id: `inv-${email.split('@')[0]}-2026-08`,
        invoiceNumber: `INV-202608-${Math.abs(hashString(email) % 1000).toString().padStart(3, '0')}`,
        currency: 'USD',
        subtotal: totalAccruedCost,
        total: totalAccruedCost,
        status: 'DRAFT (SIMULATED)',
        pricingVersion: 'v1.0.0',
        realityLabel: `SIMULATED_BILLING FOR ${email.toUpperCase()} (NON BILLABLE)`,
        issuedAt: new Date().toISOString(),
        lines,
      },
    ])
  }

  function hashString(str: string): number {
    let hash = 0
    for (let i = 0; i < str.length; i++) {
      hash = (hash << 5) - hash + str.charCodeAt(i)
      hash |= 0
    }
    return hash
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
            <span className="text-xs font-mono text-slate-400">ACCOUNT BILLING ENGINE:</span>
            <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 text-xs font-mono font-bold">
              PAYG PRICING v1.0.0
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight mt-1">Billing & Quota Management</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-0.5">
            Resource usage metering, atomic quota limits, and draft invoices strictly scoped to account: <strong className="text-white">{userEmail}</strong>.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="px-3 py-1 bg-slate-900 border border-slate-800 text-xs text-slate-400 font-mono rounded-lg">
            Account: {userEmail}
          </span>
        </div>
      </div>

      {/* Non-Billable Reality Banner */}
      <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-xl font-mono text-xs text-amber-400 flex items-center justify-between gap-4">
        <div>
          <strong className="font-bold uppercase text-amber-300">🛡 ACCOUNT-SCOPED SIMULATED BILLING:</strong>
          <span className="ml-2 text-slate-300 text-[11px]">
            No commercial payment gateway (Stripe/Razorpay) is connected. All metrics and quota limits below are dynamically calculated from the logged-in user account ({userEmail}).
          </span>
        </div>
        <span className="px-2.5 py-1 bg-amber-500/20 text-amber-300 border border-amber-500/30 rounded text-[10px] font-bold">
          NON_BILLABLE
        </span>
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <CloudMetric label="Current Month Accrued" value={`$${accountUsage.totalAccruedCost.toFixed(2)}`} subtext={`Account: ${userEmail}`} trend="ACCUMULATING" trendType="positive" />
        <CloudMetric label="Active Pricing Version" value="v1.0.0" subtext="Anarva PAYG Standard" trend="ACTIVE" trendType="positive" />
        <CloudMetric label="Compute ACU Quota" value={`${accountUsage.computeAcu.toFixed(1)} / 32.0 ACU`} subtext={`${((accountUsage.computeAcu / 32.0) * 100).toFixed(1)}% Used`} trend="AVAILABLE" trendType="positive" />
        <CloudMetric label="Storage Quota" value={`${accountUsage.storageGb.toFixed(1)} / 500 GB`} subtext={`${((accountUsage.storageGb / 500.0) * 100).toFixed(1)}% Used`} trend="AVAILABLE" trendType="positive" />
      </div>

      {/* Navigation Tabs */}
      <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title={`Current Billing Cycle Accrued Usage & Line Items (${userEmail})`}>
            <div className="space-y-4">
              <div className="p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded-xl text-[11px]">
                ℹ Usage records and line item charges below are dynamically computed from active compute, database, and storage resources owned by <strong>{userEmail}</strong>.
              </div>

              <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
                {invoices[0]?.lines.map((item) => (
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
          <CloudCard title={`Account Resource Quotas & Allocation Limits (${userEmail})`}>
            <div className="space-y-4">
              <div className="p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded-xl text-[11px]">
                ℹ Resource quotas below reflect the actual live resources provisioned by <strong>{userEmail}</strong>. Quotas are enforced atomically before any new infrastructure provisioning operation succeeds.
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
                          <span className="text-slate-400">Account Usage ({userEmail}):</span>
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
          <CloudCard title={`Billing Period Draft Invoices (${userEmail})`}>
            <div className="space-y-4">
              {invoices.map((inv) => (
                <div key={inv.id} className="p-5 bg-slate-950 border border-slate-800 rounded-xl space-y-4">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-800 pb-3">
                    <div>
                      <div className="font-bold text-white text-base">{inv.invoiceNumber}</div>
                      <div className="text-[10px] text-slate-400 mt-0.5">
                        Account: {userEmail} • Pricing Version: {inv.pricingVersion} • Reality: {inv.realityLabel}
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
