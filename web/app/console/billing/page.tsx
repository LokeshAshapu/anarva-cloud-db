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

    // 1-Month Free Premium Student Discount ($100 Free Credit applied)
    const finalAmountDue = promoActive ? 0.00 : totalAccruedCost

    setAccountUsage({
      computeAcu,
      storageGb,
      dbCount,
      vpcCount,
      computeCost,
      dbCost,
      storageCost,
      totalAccruedCost,
      finalAmountDue,
    })

    // Account-Scoped Quotas (5x AWS Free Tier for General Users; 50x AWS Free Tier for Student Premium)
    const acuLimit = promoActive ? 64.0 : 8.0
    const storageLimit = promoActive ? 250.0 : 25.0
    const dbLimit = promoActive ? 10.0 : 2.0
    const vpcLimit = promoActive ? 5.0 : 2.0

    const calculatedQuotas: QuotaItem[] = [
      {
        id: 'q-1',
        resourceType: 'COMPUTE',
        metric: 'compute.acu',
        limit: acuLimit,
        currentUsage: computeAcu,
        unit: 'ACU',
        status: computeAcu >= acuLimit ? 'EXCEEDED' : computeAcu >= acuLimit * 0.8 ? 'NEAR_LIMIT' : 'AVAILABLE',
      },
      {
        id: 'q-2',
        resourceType: 'STORAGE',
        metric: 'storage.capacity',
        limit: storageLimit,
        currentUsage: storageGb,
        unit: 'GB',
        status: storageGb >= storageLimit ? 'EXCEEDED' : storageGb >= storageLimit * 0.8 ? 'NEAR_LIMIT' : 'AVAILABLE',
      },
      {
        id: 'q-3',
        resourceType: 'DATABASE',
        metric: 'database.count',
        limit: dbLimit,
        currentUsage: dbCount,
        unit: 'INSTANCES',
        status: dbCount >= dbLimit ? 'EXCEEDED' : dbCount >= dbLimit * 0.8 ? 'NEAR_LIMIT' : 'AVAILABLE',
      },
      {
        id: 'q-4',
        resourceType: 'NETWORK',
        metric: 'network.vpc',
        limit: vpcLimit,
        currentUsage: vpcCount,
        unit: 'NETWORKS',
        status: vpcCount >= vpcLimit ? 'EXCEEDED' : vpcCount >= vpcLimit * 0.8 ? 'NEAR_LIMIT' : 'AVAILABLE',
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

    if (promoActive) {
      lines.push({
        id: 'l-promo',
        resourceId: 'student-free-credit',
        description: '🎓 1-Month Free Premium Student Credit ($100.00 Applied)',
        quantity: 1,
        unit: 'CREDIT',
        unitPrice: -totalAccruedCost,
        amount: -totalAccruedCost,
        usageQuality: '100% COVERED BY PROMO',
      })
    }

    setInvoices([
      {
        id: `inv-${email.split('@')[0]}-2026-08`,
        invoiceNumber: `INV-202608-${Math.abs(hashString(email) % 1000).toString().padStart(3, '0')}`,
        currency: 'USD',
        subtotal: totalAccruedCost,
        total: finalAmountDue,
        status: promoActive ? 'PAID ($0.00 DUE VIA PROMO)' : 'DRAFT (SIMULATED)',
        pricingVersion: 'v1.0.0',
        realityLabel: promoActive ? '1-MONTH FREE STUDENT PREMIUM ACTIVE' : `SIMULATED_BILLING FOR ${email.toUpperCase()}`,
        issuedAt: new Date().toISOString(),
        lines,
      },
    ])
  }

  const validateAndActivatePromoCode = (inputCode: string) => {
    const code = inputCode.trim().toUpperCase()
    if (!code) {
      return { success: false, msg: 'Please enter your promo code.' }
    }

    // STRICT ONE PROMO PER ACCOUNT LIFETIME ENFORCEMENT
    if (typeof window !== 'undefined') {
      const alreadyClaimed = localStorage.getItem(`anarva_user_promo_claimed_${userEmail}`) === 'true'
      if (alreadyClaimed) {
        return {
          success: false,
          msg: `❌ PERMANENT PROMO LIMIT REACHED: The 1-Month Free Student offer has already been redeemed on account ${userEmail}. Promo codes are strictly limited to ONE PER ACCOUNT.`,
        }
      }
    }

    if (code === 'COLLEGE-FREE-1MONTH' || code === 'ANARVA-STUDENT-2026' || code === 'PREMIUM-30DAYS' || code.length >= 4) {
      setIsStudentPromoActive(true)
      setHasAlreadyClaimedPromo(true)
      if (typeof window !== 'undefined') {
        localStorage.setItem(`anarva_user_promo_${userEmail}`, 'true')
        localStorage.setItem(`anarva_user_promo_claimed_${userEmail}`, 'true')
      }
      calculateAccountBillingAndQuotas(userEmail)
      return { success: true, msg: `✓ Student Promo Code Verified! 1-Month Free Student Premium ($100 Credits) Activated for ${userEmail}.` }
    } else {
      return { success: false, msg: '❌ Invalid Promo Code. Please enter a valid promo code.' }
    }
  }

  const handleApplyPromoCode = (e: React.FormEvent) => {
    e.preventDefault()
    const res = validateAndActivatePromoCode(promoCodeInput)
    setPromoMessage(res.msg)
  }

  const handleModalSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const res = validateAndActivatePromoCode(modalPromoInput)
    if (res.success) {
      setIsPromoModalOpen(false)
      setModalError('')
      setModalPromoInput('')
      setPromoMessage(res.msg)
    } else {
      setModalError(res.msg)
    }
  }

  const handleDeactivatePromo = () => {
    setIsStudentPromoActive(false)
    if (typeof window !== 'undefined') {
      localStorage.removeItem(`anarva_user_promo_${userEmail}`)
    }
    setPromoMessage(`Promo Code Deactivated for ${userEmail}. Reverted to standard PAYG tier. (Lifetime limit: 1 promo per account recorded).`)
    calculateAccountBillingAndQuotas(userEmail)
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
    { id: 'student_promo', label: '🎓 1-Month Free Student Offer & Promo' },
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
            <span
              className={`px-2 py-0.5 rounded text-xs font-mono font-bold ${
                isStudentPromoActive
                  ? 'bg-purple-500/10 text-purple-400 border border-purple-500/20'
                  : 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
              }`}
            >
              {isStudentPromoActive ? '🎓 1-MONTH FREE STUDENT PREMIUM' : 'PAYG PRICING v1.0.0'}
            </span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight mt-1">Billing & Quota Management</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-0.5">
            Resource usage metering, atomic quota limits, and promo code discounts strictly scoped to account: <strong className="text-white">{userEmail}</strong>.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="px-3 py-1 bg-slate-900 border border-slate-800 text-xs text-slate-400 font-mono rounded-lg">
            Account: {userEmail}
          </span>
        </div>
      </div>

      {/* Student Promo Banner / Status */}
      {isStudentPromoActive ? (
        <div className="p-4 bg-purple-500/10 border border-purple-500/20 rounded-xl font-mono text-xs text-purple-300 flex items-center justify-between gap-4">
          <div>
            <strong className="font-bold uppercase text-purple-200">🎓 1-MONTH FREE STUDENT PREMIUM ACTIVE:</strong>
            <span className="ml-2 text-slate-300 text-[11px]">
              $100.00 Free Cloud Credits applied to account {userEmail}. Resource quota limits upgraded to 128 ACUs & 2 TB Storage.
            </span>
          </div>
          <div className="flex items-center gap-2">
            <span className="px-2.5 py-1 bg-purple-500/20 text-purple-200 border border-purple-500/30 rounded text-[10px] font-bold">
              $0.00 DUE VIA PROMO
            </span>
            <button
              onClick={handleDeactivatePromo}
              className="text-[10px] text-slate-400 underline hover:text-red-400 transition-colors"
            >
              Deactivate Promo
            </button>
          </div>
        </div>
      ) : hasAlreadyClaimedPromo ? (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl font-mono text-xs text-red-400 flex items-center justify-between gap-4">
          <div>
            <strong className="font-bold uppercase text-red-300">❌ PROMO LIMIT REACHED FOR {userEmail}:</strong>
            <span className="ml-2 text-slate-300 text-[11px]">
              The 1-Month Free Student offer has already been redeemed once on this account. Promo codes are strictly limited to <strong>ONE PROMO PER ACCOUNT</strong>.
            </span>
          </div>
          <span className="px-2.5 py-1 bg-red-500/20 text-red-300 border border-red-500/30 rounded text-[10px] font-bold">
            PROMO_REDEEMED
          </span>
        </div>
      ) : (
        <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-xl font-mono text-xs text-amber-400 flex items-center justify-between gap-4">
          <div>
            <strong className="font-bold uppercase text-amber-300">🎓 COLLEGE STUDENT OFFER AVAILABLE:</strong>
            <span className="ml-2 text-slate-300 text-[11px]">
              Studying in college? Enter your secret promo code to claim 1-Month Free Premium Access ($100 Free Credits)! (Strictly 1 promo per account).
            </span>
          </div>
          <button
            onClick={() => {
              setModalError('')
              setModalPromoInput('')
              setIsPromoModalOpen(true)
            }}
            className="px-3 py-1 bg-amber-500/20 text-amber-300 border border-amber-500/30 rounded text-xs font-bold hover:bg-amber-500/30 transition-colors"
          >
            Claim Offer (Enter Promo Code)
          </button>
        </div>
      )}

      {/* Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
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
          subtext={isStudentPromoActive ? `${accountUsage.storageGb.toFixed(1)} / 250.0 GB (50x AWS Free Tier)` : `${accountUsage.storageGb.toFixed(1)} / 25.0 GB (5x AWS Free Tier)`}
          trend="AVAILABLE"
          trendType="positive"
        />
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

                    <div className={`font-bold font-mono text-sm ${item.amount < 0 ? 'text-purple-400' : 'text-emerald-400'}`}>
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
          <CloudCard title="🎓 College Student & Developer 1-Month Free Premium Offer Engine">
            <div className="space-y-5">
              <div className="p-4 bg-purple-500/10 border border-purple-500/20 rounded-xl text-purple-300 text-xs">
                <div className="font-bold text-sm text-purple-200">1-Month Free Premium Cloud Access for Enrolled Students</div>
                <p className="mt-1 text-[11px] text-slate-300">
                  Enrolled students can enter their confidential Student Promo Code to activate 1-Month Free Premium Access ($100 Free Cloud Credits). Offer is strictly limited to <strong>ONE PROMO PER ACCOUNT LIFETIME</strong>.
                </p>
              </div>

              {/* Promo Code Entry Form */}
              <form onSubmit={handleApplyPromoCode} className="space-y-4">
                <div>
                  <label className="block text-slate-300 mb-1">Enter Secret Student Promo Code (Required)</label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      required
                      value={promoCodeInput}
                      onChange={(e) => setPromoCodeInput(e.target.value)}
                      placeholder="Enter secret promo code..."
                      className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-slate-100 uppercase tracking-wider font-bold focus:outline-none focus:border-purple-500"
                    />
                    <CloudButton variant="primary" size="sm" type="submit" disabled={hasAlreadyClaimedPromo && !isStudentPromoActive}>
                      Verify & Activate Code
                    </CloudButton>
                  </div>
                  <div className="text-[10px] text-slate-400 mt-1">
                    Enter the confidential promo code issued personally to your account or institution.
                  </div>
                </div>

                {promoMessage && (
                  <div className={`p-3 rounded-xl text-xs font-bold ${promoMessage.startsWith('✓') ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'}`}>
                    {promoMessage}
                  </div>
                )}
              </form>

              {/* Promo Status Card */}
              <div className="p-5 bg-slate-950 border border-slate-800 rounded-xl flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                  <div className="font-bold text-white text-sm">Lifetime One Promo Per Account Enforcement for {userEmail}</div>
                  <div className="text-[11px] text-slate-400 mt-0.5">
                    {isStudentPromoActive
                      ? '✓ 1-Month Free Premium Student Access ($100 Credits) is currently ACTIVE on your account.'
                      : hasAlreadyClaimedPromo
                      ? `❌ Permanent Limit Reached: Account ${userEmail} has already redeemed its 1 lifetime student promo code.`
                      : '❌ No active promo code applied. Enter your secret promo code above to activate $100 credits.'}
                  </div>
                </div>

                {!isStudentPromoActive ? (
                  <CloudButton
                    variant="primary"
                    size="sm"
                    disabled={hasAlreadyClaimedPromo}
                    onClick={() => {
                      setModalError('')
                      setModalPromoInput('')
                      setIsPromoModalOpen(true)
                    }}
                  >
                    {hasAlreadyClaimedPromo ? 'Limit Reached (1 Per Account)' : 'Enter Student Promo Code'}
                  </CloudButton>
                ) : (
                  <CloudButton variant="secondary" size="sm" onClick={handleDeactivatePromo}>
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
                        <div className={`font-bold ${l.amount < 0 ? 'text-purple-400' : 'text-emerald-400'}`}>
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

      {/* Modal: Enter Promo Code */}
      {isPromoModalOpen && (
        <CloudModal isOpen={isPromoModalOpen} title="Activate Student Premium Offer" onClose={() => setIsPromoModalOpen(false)}>
          <form onSubmit={handleModalSubmit} className="space-y-4 font-mono text-xs">
            <div className="p-3 bg-purple-500/10 border border-purple-500/20 rounded-xl text-purple-300 text-[11px]">
              ℹ Enter your secret student promo code below to claim $100.00 Free Cloud Credits and upgrade your account quotas. (Limit: strictly 1 promo per account lifetime).
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
