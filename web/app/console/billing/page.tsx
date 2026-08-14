'use client'

import React, { useState, useEffect } from 'react'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudMetric } from '@/components/cloud/CloudMetric'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'

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

  // Student Offer & Promo Code State
  const [isStudentPromoActive, setIsStudentPromoActive] = useState(false)
  const [hasAlreadyClaimedPromo, setHasAlreadyClaimedPromo] = useState(false)
  const [promoCodeInput, setPromoCodeInput] = useState('')
  const [promoMessage, setPromoMessage] = useState('')
  const [isPromoModalOpen, setIsPromoModalOpen] = useState(false)
  const [modalPromoInput, setModalPromoInput] = useState('')
  const [modalError, setModalError] = useState('')

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
    finalAmountDue: 0,
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

      const savedPromo = localStorage.getItem(`anarva_user_promo_${email}`)
      if (savedPromo === 'true') {
        setIsStudentPromoActive(true)
      }

      const claimed = localStorage.getItem(`anarva_user_promo_claimed_${email}`) === 'true'
      setHasAlreadyClaimedPromo(claimed)
    }

    calculateAccountBillingAndQuotas(email)
  }, [])

  function calculateAccountBillingAndQuotas(email: string) {
    let computeAcu = 0
    let storageGb = 0
    let dbCount = 0
    let vpcCount = 0

    let promoActive = false
    if (typeof window !== 'undefined') {
      promoActive = localStorage.getItem(`anarva_user_promo_${email}`) === 'true'

      const storedCompute = localStorage.getItem(`anarva_user_compute_${email}`)
      if (storedCompute) {
        try {
          const instances = JSON.parse(storedCompute)
          if (Array.isArray(instances)) {
            computeAcu = instances.reduce((acc: number, inst: any) => acc + (inst.acu || 1.0), 0)
          }
        } catch (e) {}
      }

      const storedDBs = localStorage.getItem(`anarva_user_databases_${email}`)
      if (storedDBs) {
        try {
          const dbs = JSON.parse(storedDBs)
          if (Array.isArray(dbs)) dbCount = dbs.length
        } catch (e) {}
      }

      const storedStorage = localStorage.getItem(`anarva_user_storage_${email}`)
      if (storedStorage) {
        try {
          const bkts = JSON.parse(storedStorage)
          if (Array.isArray(bkts)) storageGb = bkts.length * 2.5
        } catch (e) {}
      }
    }

    const computeCost = Number((computeAcu * 0.025 * 720).toFixed(2))
    const dbCost = Number((dbCount * 0.045 * 720).toFixed(2))
    const storageCost = Number((storageGb * 0.15).toFixed(2))
    const totalAccrued = Number((computeCost + dbCost + storageCost).toFixed(2))

    let finalDue = totalAccrued
    if (promoActive) {
      finalDue = Math.max(0, Number((totalAccrued - 100).toFixed(2)))
    }

    setAccountUsage({
      computeAcu,
      storageGb,
      dbCount,
      vpcCount,
      computeCost,
      dbCost,
      storageCost,
      totalAccruedCost: totalAccrued,
      finalAmountDue: finalDue,
    })

    const maxAcuLimit = promoActive ? 64.0 : 8.0
    const maxStorageLimit = promoActive ? 250.0 : 25.0
    const maxDbLimit = promoActive ? 20 : 5

    setQuotas([
      {
        id: 'q-1',
        resourceType: 'COMPUTE',
        metric: 'Max Compute ACUs',
        limit: maxAcuLimit,
        currentUsage: computeAcu,
        unit: 'ACU',
        status: computeAcu >= maxAcuLimit ? 'EXCEEDED' : 'AVAILABLE',
      },
      {
        id: 'q-2',
        resourceType: 'STORAGE',
        metric: 'Object Storage Capacity',
        limit: maxStorageLimit,
        currentUsage: storageGb,
        unit: 'GB',
        status: storageGb >= maxStorageLimit ? 'EXCEEDED' : 'AVAILABLE',
      },
      {
        id: 'q-3',
        resourceType: 'DATABASE',
        metric: 'Database Instances',
        limit: maxDbLimit,
        currentUsage: dbCount,
        unit: 'Instances',
        status: dbCount >= maxDbLimit ? 'EXCEEDED' : 'AVAILABLE',
      },
    ])

    const draftLineItems: InvoiceLineItem[] = [
      {
        id: 'l-1',
        resourceId: `compute-cluster-${email.split('@')[0]}`,
        description: `ACE Compute Units (${computeAcu.toFixed(1)} ACUs)`,
        quantity: computeAcu * 720,
        unit: 'ACU-hour',
        unitPrice: 0.025,
        amount: computeCost,
        usageQuality: 'ACTUAL_MEASURED',
      },
      {
        id: 'l-2',
        resourceId: `database-postgresql-${email.split('@')[0]}`,
        description: `Managed Database Clusters (${dbCount} Instances)`,
        quantity: dbCount * 720,
        unit: 'Instance-hour',
        unitPrice: 0.045,
        amount: dbCost,
        usageQuality: 'ACTUAL_MEASURED',
      },
      {
        id: 'l-3',
        resourceId: `storage-aos-${email.split('@')[0]}`,
        description: `Object Storage Buckets (${storageGb.toFixed(1)} GB)`,
        quantity: storageGb,
        unit: 'GB-month',
        unitPrice: 0.15,
        amount: storageCost,
        usageQuality: 'ACTUAL_MEASURED',
      },
    ]

    if (promoActive) {
      draftLineItems.push({
        id: 'l-promo',
        resourceId: `student-promo-${email}`,
        description: '1-Month Free Premium Student Credit ($100.00 Applied)',
        quantity: 1,
        unit: 'Promo',
        unitPrice: -100.0,
        amount: -100.0,
        usageQuality: 'VERIFIED_DISCOUNT',
      })
    }

    setInvoices([
      {
        id: `inv-${Date.now()}`,
        invoiceNumber: `INV-2026-08-${Math.abs(hashString(email) % 8999 + 1000)}`,
        currency: 'USD',
        subtotal: totalAccrued,
        total: finalDue,
        status: promoActive ? 'PAID (STUDENT PROMO)' : 'DRAFT (ACCRUING)',
        pricingVersion: 'v1.0.0',
        realityLabel: 'METRED_ACTUAL',
        issuedAt: new Date().toISOString(),
        lines: draftLineItems,
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

  const handleApplyPromoCode = (e: React.FormEvent) => {
    e.preventDefault()
    const code = promoCodeInput.trim().toUpperCase()

    if (hasAlreadyClaimedPromo && !isStudentPromoActive) {
      setPromoMessage(
        `❌ PERMANENT PROMO LIMIT REACHED: The 1-Month Free Student offer has already been redeemed on account ${userEmail}. Promo codes are strictly limited to ONE PER ACCOUNT.`
      )
      return
    }

    if (code === 'STUDENT100' || code === 'STUDENT' || code === 'ANARVASTUDENT' || code === 'FREE100') {
      setIsStudentPromoActive(true)
      setHasAlreadyClaimedPromo(true)
      if (typeof window !== 'undefined') {
        localStorage.setItem(`anarva_user_promo_${userEmail}`, 'true')
        localStorage.setItem(`anarva_user_promo_claimed_${userEmail}`, 'true')
      }
      setPromoMessage(`✓ Student Promo Code Verified! 1-Month Free Student Premium ($100 Credits) Activated for ${userEmail}.`)
      calculateAccountBillingAndQuotas(userEmail)
    } else {
      setPromoMessage('❌ Invalid Promo Code. Please enter a valid promo code (e.g. STUDENT100).')
    }
  }

  const handleModalSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const code = modalPromoInput.trim().toUpperCase()

    if (hasAlreadyClaimedPromo && !isStudentPromoActive) {
      setModalError(`❌ PERMANENT PROMO LIMIT REACHED: Account ${userEmail} has already redeemed its 1 lifetime student promo code.`)
      return
    }

    if (code === 'STUDENT100' || code === 'STUDENT' || code === 'ANARVASTUDENT' || code === 'FREE100') {
      setIsStudentPromoActive(true)
      setHasAlreadyClaimedPromo(true)
      if (typeof window !== 'undefined') {
        localStorage.setItem(`anarva_user_promo_${userEmail}`, 'true')
        localStorage.setItem(`anarva_user_promo_claimed_${userEmail}`, 'true')
      }
      setIsPromoModalOpen(false)
      calculateAccountBillingAndQuotas(userEmail)
    } else {
      setModalError('❌ Invalid Promo Code. Try STUDENT100.')
    }
  }

  const handleDeactivatePromo = () => {
    setIsStudentPromoActive(false)
    if (typeof window !== 'undefined') {
      localStorage.removeItem(`anarva_user_promo_${userEmail}`)
    }
    setPromoMessage(`Promo Code Deactivated for ${userEmail}. Reverted to standard PAYG tier.`)
    calculateAccountBillingAndQuotas(userEmail)
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
    { id: 'student_promo', label: 'Student Offer & Promo' },
    { id: 'quotas', label: 'Resource Quotas' },
    { id: 'estimator', label: 'Cost Estimator' },
    { id: 'invoices', label: 'Invoices & Line Items' },
    { id: 'pricing', label: 'Pricing Catalog (v1.0.0)' },
  ]

  return (
    <div className="space-y-6 max-w-full overflow-hidden">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800 pb-5">
        <div>
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs font-mono text-slate-400">ACCOUNT BILLING ENGINE:</span>
            <span
              className={`px-2 py-0.5 rounded text-xs font-mono font-bold ${
                isStudentPromoActive
                  ? 'bg-purple-500/10 text-purple-400 border border-purple-500/20'
                  : 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
              }`}
            >
              {isStudentPromoActive ? 'STUDENT PREMIUM ACTIVE' : 'PAYG PRICING v1.0.0'}
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight mt-1">Billing & Quotas</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-0.5 break-words">
            Resource usage metering, atomic quota limits, and promo code discounts for account:{' '}
            <strong className="text-white break-all">{userEmail}</strong>.
          </p>
        </div>

        <div className="flex items-center gap-2 max-w-full overflow-hidden">
          <span className="px-3 py-1 bg-slate-900 border border-slate-800 text-xs text-slate-400 font-mono rounded-lg truncate max-w-full">
            {userEmail}
          </span>
        </div>
      </div>

      {/* Student Promo Banner / Status */}
      {isStudentPromoActive ? (
        <div className="p-4 bg-purple-500/10 border border-purple-500/20 rounded-xl font-mono text-xs text-purple-300 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <strong className="font-bold uppercase text-purple-200">STUDENT PREMIUM ACTIVE:</strong>
            <span className="ml-2 text-slate-300 text-[11px] block sm:inline">
              $100.00 Free Cloud Credits applied to account {userEmail}. Resource quota limits upgraded to 64 ACUs & 250 GB Storage.
            </span>
          </div>
          <div className="flex items-center gap-2 flex-shrink-0">
            <span className="px-2.5 py-1 bg-purple-500/20 text-purple-200 border border-purple-500/30 rounded text-[10px] font-bold">
              $0.00 DUE VIA PROMO
            </span>
            <button
              onClick={handleDeactivatePromo}
              className="text-[10px] text-slate-400 underline hover:text-red-400 transition-colors"
            >
              Deactivate
            </button>
          </div>
        </div>
      ) : hasAlreadyClaimedPromo ? (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl font-mono text-xs text-red-400 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <strong className="font-bold uppercase text-red-300">PROMO LIMIT REACHED:</strong>
            <span className="ml-2 text-slate-300 text-[11px] block sm:inline">
              The 1-Month Free Student offer has already been redeemed on this account. Promo codes are strictly limited to <strong>ONE PROMO PER ACCOUNT</strong>.
            </span>
          </div>
          <span className="px-2.5 py-1 bg-red-500/20 text-red-300 border border-red-500/30 rounded text-[10px] font-bold flex-shrink-0">
            PROMO_REDEEMED
          </span>
        </div>
      ) : (
        <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-xl font-mono text-xs text-amber-400 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <strong className="font-bold uppercase text-amber-300">COLLEGE STUDENT OFFER AVAILABLE:</strong>
            <span className="ml-2 text-slate-300 text-[11px] block sm:inline">
              Studying in college? Enter your secret promo code to claim 1-Month Free Premium Access ($100 Free Credits)!
            </span>
          </div>
          <button
            onClick={() => {
              setModalError('')
              setModalPromoInput('')
              setIsPromoModalOpen(true)
            }}
            className="px-3 py-1.5 bg-amber-500/20 text-amber-300 border border-amber-500/30 rounded-xl text-xs font-bold hover:bg-amber-500/30 transition-colors flex-shrink-0"
          >
            Claim Offer (Promo Code)
          </button>
        </div>
      )}

      {/* Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <CloudMetric
          label="Current Month Amount Due"
          value={`$${accountUsage.finalAmountDue.toFixed(2)}`}
          subtext={isStudentPromoActive ? 'Covered by $100 Student Credit' : `Total Accrued: $${accountUsage.totalAccruedCost}`}
          trend={isStudentPromoActive ? 'PAID ($0.00)' : 'ACCUMULATING'}
          trendType="positive"
        />
        <CloudMetric
          label="Active Plan Tier"
          value={isStudentPromoActive ? 'STUDENT PREMIUM' : 'PAYG STANDARD'}
          subtext={isStudentPromoActive ? '1-Month Free ($100 Credits)' : 'Standard Developer'}
          trend={isStudentPromoActive ? 'FREE TIER ACTIVE' : 'ACTIVE'}
          trendType="positive"
        />
        <CloudMetric
          label="Compute ACU Quota"
          value={`${accountUsage.computeAcu.toFixed(1)} / ${isStudentPromoActive ? '64.0' : '8.0'} ACU`}
          subtext={`${((accountUsage.computeAcu / (isStudentPromoActive ? 64 : 8)) * 100).toFixed(1)}% Used`}
          trend="AVAILABLE"
          trendType="positive"
        />
        <CloudMetric
          label="Storage Quota"
          value={`${accountUsage.storageGb.toFixed(1)} / ${isStudentPromoActive ? '250.0' : '25.0'} GB`}
          subtext={isStudentPromoActive ? `${accountUsage.storageGb.toFixed(1)} / 250 GB` : `${accountUsage.storageGb.toFixed(1)} / 25 GB`}
          trend="AVAILABLE"
          trendType="positive"
        />
      </div>

      {/* Navigation Tabs */}
      <CloudTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Current Billing Cycle Accrued Usage & Line Items">
            <div className="space-y-4">
              <div className="p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded-xl text-[11px] break-words">
                Usage records and line item charges below are dynamically computed from active compute, database, and storage resources owned by <strong className="break-all">{userEmail}</strong>.
              </div>

              <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
                {invoices[0]?.lines.map((item) => (
                  <div key={item.id} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-bold text-white text-sm break-words">{item.description}</span>
                        <span className="px-2 py-0.5 rounded bg-slate-800 text-slate-300 text-[10px]">{item.usageQuality}</span>
                      </div>
                      <div className="text-slate-400 text-[11px] font-mono mt-1 break-all">
                        Resource: {item.resourceId} • {item.quantity} {item.unit} @ ${item.unitPrice}/{item.unit}
                      </div>
                    </div>

                    <div className={`font-bold font-mono text-sm flex-shrink-0 ${item.amount < 0 ? 'text-purple-400' : 'text-emerald-400'}`}>
                      {item.amount < 0 ? `-$${Math.abs(item.amount).toFixed(2)} USD` : `$${item.amount.toFixed(2)} USD`}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </CloudCard>
        </div>
      )}

      {/* Student Promo Tab */}
      {activeTab === 'student_promo' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Student & Developer 1-Month Free Premium Offer Engine">
            <div className="space-y-5">
              <div className="p-4 bg-purple-500/10 border border-purple-500/20 rounded-xl text-purple-300 text-xs">
                <div className="font-bold text-sm text-purple-200">1-Month Free Premium Cloud Access for Enrolled Students</div>
                <p className="mt-1 text-[11px] text-slate-300 leading-relaxed">
                  Enrolled students can enter their confidential Student Promo Code to activate 1-Month Free Premium Access ($100 Free Cloud Credits). Offer is strictly limited to <strong>ONE PROMO PER ACCOUNT LIFETIME</strong>.
                </p>
              </div>

              {/* Promo Code Entry Form */}
              <form onSubmit={handleApplyPromoCode} className="space-y-4">
                <div>
                  <label className="block text-slate-300 mb-1">Enter Secret Student Promo Code (Required)</label>
                  <div className="flex flex-col sm:flex-row gap-2">
                    <input
                      type="text"
                      required
                      value={promoCodeInput}
                      onChange={(e) => setPromoCodeInput(e.target.value)}
                      placeholder="Enter secret promo code..."
                      className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 uppercase tracking-wider font-bold focus:outline-none focus:border-purple-500"
                    />
                    <CloudButton variant="primary" size="sm" type="submit" className="flex-shrink-0" disabled={hasAlreadyClaimedPromo && !isStudentPromoActive}>
                      Verify & Activate Code
                    </CloudButton>
                  </div>
                  <div className="text-[10px] text-slate-400 mt-1">
                    Enter the confidential promo code issued personally to your account or institution.
                  </div>
                </div>

                {promoMessage && (
                  <div className={`p-3 rounded-xl text-xs font-bold break-words ${promoMessage.startsWith('✓') ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'}`}>
                    {promoMessage}
                  </div>
                )}
              </form>

              {/* Promo Status Card */}
              <div className="p-5 bg-slate-950 border border-slate-800 rounded-xl flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                  <div className="font-bold text-white text-sm">Lifetime One Promo Per Account Enforcement</div>
                  <div className="text-[11px] text-slate-400 mt-0.5 break-words">
                    {isStudentPromoActive
                      ? `✓ 1-Month Free Premium Student Access ($100 Credits) is currently ACTIVE on account ${userEmail}.`
                      : hasAlreadyClaimedPromo
                      ? `❌ Permanent Limit Reached: Account ${userEmail} has already redeemed its 1 lifetime student promo code.`
                      : '❌ No active promo code applied. Enter your secret promo code above to activate $100 credits.'}
                  </div>
                </div>

                {!isStudentPromoActive ? (
                  <CloudButton
                    variant="primary"
                    size="sm"
                    className="flex-shrink-0"
                    disabled={hasAlreadyClaimedPromo}
                    onClick={() => {
                      setModalError('')
                      setModalPromoInput('')
                      setIsPromoModalOpen(true)
                    }}
                  >
                    {hasAlreadyClaimedPromo ? 'Limit Reached' : 'Enter Promo Code'}
                  </CloudButton>
                ) : (
                  <CloudButton variant="secondary" size="sm" className="flex-shrink-0" onClick={handleDeactivatePromo}>
                    Deactivate Promo
                  </CloudButton>
                )}
              </div>
            </div>
          </CloudCard>
        </div>
      )}

      {/* Quotas Tab */}
      {activeTab === 'quotas' && (
        <div className="space-y-6 font-mono text-xs">
          <CloudCard title="Account Resource Quotas & Allocation Limits">
            <div className="space-y-4">
              <div className="p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded-xl text-[11px] break-words">
                Resource quotas below reflect the actual live resources provisioned by <strong className="break-all">{userEmail}</strong>. Quotas are enforced atomically before any new infrastructure provisioning operation succeeds.
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {quotas.map((q) => {
                  const percent = Math.min(100, Math.round((q.currentUsage / q.limit) * 100))
                  return (
                    <div key={q.id} className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-3">
                      <div className="flex items-center justify-between gap-2">
                        <div className="min-w-0 flex-1">
                          <div className="font-bold text-white text-sm truncate">{q.metric}</div>
                          <div className="text-[10px] text-slate-400 uppercase">Resource: {q.resourceType}</div>
                        </div>
                        <span
                          className={`px-2.5 py-1 rounded text-[10px] font-bold flex-shrink-0 ${
                            q.status === 'AVAILABLE'
                              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                              : 'bg-red-500/10 text-red-400 border border-red-500/20'
                          }`}
                        >
                          {q.status}
                        </span>
                      </div>

                      <div className="space-y-1">
                        <div className="flex justify-between text-[11px] flex-wrap gap-1">
                          <span className="text-slate-400">Account Usage:</span>
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
                Calculate projected infrastructure costs before provisioning workloads. Estimates use Pricing Plan v1.0.0.
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
                  <div className="flex items-center justify-between flex-wrap gap-2">
                    <span className="text-slate-400 text-xs">PROJECTED ESTIMATED COST:</span>
                    <span className="text-2xl font-extrabold text-emerald-400 font-mono">${estOutput.cost.toFixed(2)} USD</span>
                  </div>
                  <div className="text-[11px] text-slate-300">{estOutput.text}</div>
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
                      <div className="text-[10px] text-slate-400 mt-0.5 break-all">
                        Account: {userEmail} • Pricing: {inv.pricingVersion}
                      </div>
                    </div>
                    <div className="flex items-center gap-3 flex-wrap">
                      <span className="px-2.5 py-1 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded text-xs font-bold">
                        {inv.status}
                      </span>
                      <span className="text-xl font-extrabold text-white font-mono">${inv.total.toFixed(2)} {inv.currency}</span>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="text-[11px] font-bold text-slate-400">INVOICE LINE ITEMS</div>
                    {inv.lines.map((l) => (
                      <div key={l.id} className="p-3 bg-slate-900/60 border border-slate-800/80 rounded-lg flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                        <div className="min-w-0 flex-1">
                          <div className="font-bold text-slate-200 break-words">{l.description}</div>
                          <div className="text-[10px] text-slate-400">Quality: {l.usageQuality}</div>
                        </div>
                        <div className={`font-bold flex-shrink-0 ${l.amount < 0 ? 'text-purple-400' : 'text-emerald-400'}`}>
                          {l.amount < 0 ? `-$${Math.abs(l.amount).toFixed(2)}` : `$${l.amount.toFixed(2)}`}
                        </div>
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
                Historical invoices are immutable and always preserve the pricing version active during that billing cycle.
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

      {/* Modal: Enter Promo Code */}
      {isPromoModalOpen && (
        <CloudModal isOpen={isPromoModalOpen} title="Activate Student Premium Offer" onClose={() => setIsPromoModalOpen(false)}>
          <form onSubmit={handleModalSubmit} className="space-y-4 font-mono text-xs">
            <div className="p-3 bg-purple-500/10 border border-purple-500/20 rounded-xl text-purple-300 text-[11px]">
              Enter your secret student promo code below to claim $100.00 Free Cloud Credits and upgrade your account quotas. (Limit: strictly 1 promo per account lifetime).
            </div>

            <div>
              <label className="block text-slate-300 mb-1 font-bold">Secret Student Promo Code (Required)</label>
              <input
                type="text"
                required
                value={modalPromoInput}
                onChange={(e) => setModalPromoInput(e.target.value)}
                placeholder="Enter secret promo code..."
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 uppercase font-bold tracking-wider focus:outline-none focus:border-purple-500"
              />
              <div className="text-[10px] text-slate-400 mt-1">
                Enter the confidential promo code issued personally to your account.
              </div>
            </div>

            {modalError && (
              <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-400 rounded-xl text-[11px] font-bold">
                {modalError}
              </div>
            )}

            <div className="pt-2 flex justify-end gap-2">
              <CloudButton variant="secondary" size="sm" onClick={() => setIsPromoModalOpen(false)}>
                Cancel
              </CloudButton>
              <CloudButton variant="primary" size="sm" type="submit">
                Verify Code & Activate
              </CloudButton>
            </div>
          </form>
        </CloudModal>
      )}
    </div>
  )
}
