'use client'

import React, { useState, useEffect } from 'react'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudButton } from '@/components/cloud/CloudButton'

interface FeedbackItem {
  feedbackId: string
  id: string
  userId: string
  userEmail: string
  userName?: string
  organizationId: string
  projectId?: string
  rating: number
  category: string
  subject: string
  message: string
  status: 'NEW' | 'REVIEWING' | 'PLANNED' | 'IN_PROGRESS' | 'RESOLVED' | 'CLOSED'
  targetEmail: string
  createdAt: string
  updatedAt: string
  requestId: string
}

interface FeedbackAnalytics {
  totalFeedback: number
  averageRating: number
  ratingDistribution: Record<number, number>
  statusCounts: Record<string, number>
}

export default function FeedbackManagementPage() {
  const [items, setItems] = useState<FeedbackItem[]>([
    {
      feedbackId: 'fb-101',
      id: 'fb-101',
      userId: 'usr-operator-01',
      userEmail: 'lokeshashapu@gmail.com',
      userName: 'Cloud Operator',
      organizationId: 'org-default',
      projectId: 'proj-default',
      rating: 5,
      category: 'GENERAL',
      subject: 'Database provisioning speed is excellent',
      message: 'The RDS PostgreSQL cluster provisioning and failover orchestration experience is extremely smooth.',
      status: 'NEW',
      targetEmail: '23w61a0506@gmail.com',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      requestId: 'req_fb_101',
    },
    {
      feedbackId: 'fb-102',
      id: 'fb-102',
      userId: 'usr-dev-02',
      userEmail: 'developer@anarva.io',
      userName: 'Senior Engineer',
      organizationId: 'org-default',
      projectId: 'proj-default',
      rating: 4,
      category: 'FEATURE_REQUEST',
      subject: 'Add auto-scaling bounds for Database ACU capacity',
      message: 'Would love to see automated ACU scaling policies based on CloudWatch CPU metrics.',
      status: 'REVIEWING',
      targetEmail: '23w61a0506@gmail.com',
      createdAt: new Date(Date.now() - 86400000).toISOString(),
      updatedAt: new Date(Date.now() - 43200000).toISOString(),
      requestId: 'req_fb_102',
    },
  ])

  const [analytics, setAnalytics] = useState<FeedbackAnalytics>({
    totalFeedback: 2,
    averageRating: 4.5,
    ratingDistribution: { 5: 1, 4: 1, 3: 0, 2: 0, 1: 0 },
    statusCounts: { NEW: 1, REVIEWING: 1, PLANNED: 0, IN_PROGRESS: 0, RESOLVED: 0, CLOSED: 0 },
  })

  const [statusFilter, setStatusFilter] = useState('ALL')
  const [minRatingFilter, setMinRatingFilter] = useState(0)
  const [page, setPage] = useState(1)
  const [selectedFeedback, setSelectedFeedback] = useState<FeedbackItem | null>(null)
  const [updating, setUpdating] = useState(false)

  const fetchFeedback = async () => {
    try {
      const query = new URLSearchParams()
      if (statusFilter !== 'ALL') query.set('status', statusFilter)
      if (minRatingFilter > 0) query.set('minRating', minRatingFilter.toString())
      query.set('page', page.toString())
      query.set('pageSize', '10')

      const res = await fetch(`/api/v1/feedback?${query.toString()}`)
      if (res.ok) {
        const body = await res.json()
        if (body.data?.items) {
          setItems(body.data.items)
        }
      }
    } catch (e) {
      console.log('Feedback fetch notice:', e)
    }
  }

  const fetchAnalytics = async () => {
    try {
      const res = await fetch('/api/v1/feedback/analytics')
      if (res.ok) {
        const body = await res.json()
        if (body.data) {
          setAnalytics(body.data)
        }
      }
    } catch (e) {
      console.log('Analytics fetch notice:', e)
    }
  }

  useEffect(() => {
    fetchFeedback()
    fetchAnalytics()
  }, [statusFilter, minRatingFilter, page])

  const handleUpdateStatus = async (fbId: string, newStatus: string) => {
    setUpdating(true)
    try {
      const res = await fetch(`/api/v1/feedback/${fbId}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      })
      if (res.ok) {
        setItems((prev) =>
          prev.map((item) => (item.feedbackId === fbId ? { ...item, status: newStatus as any } : item))
        )
        if (selectedFeedback && selectedFeedback.feedbackId === fbId) {
          setSelectedFeedback({ ...selectedFeedback, status: newStatus as any })
        }
        fetchAnalytics()
      }
    } catch (e) {
      console.log('Status update notice:', e)
    } finally {
      setUpdating(false)
    }
  }

  const renderStars = (count: number) => (
    <div className="flex items-center gap-0.5 text-amber-400">
      {[1, 2, 3, 4, 5].map((s) => (
        <span key={s} className={s <= count ? 'text-amber-400' : 'text-slate-700'}>★</span>
      ))}
    </div>
  )

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'NEW':
        return 'bg-blue-500/10 text-blue-400 border-blue-500/30'
      case 'REVIEWING':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/30'
      case 'PLANNED':
        return 'bg-purple-500/10 text-purple-400 border-purple-500/30'
      case 'IN_PROGRESS':
        return 'bg-cyan-500/10 text-cyan-400 border-cyan-500/30'
      case 'RESOLVED':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
      case 'CLOSED':
        return 'bg-slate-800 text-slate-400 border-slate-700'
      default:
        return 'bg-slate-800 text-slate-400 border-slate-700'
    }
  }

  return (
    <div className="space-y-6 font-sans text-slate-100 max-w-7xl mx-auto pb-12">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-extrabold text-white tracking-tight">Anarva Feedback Intelligence</h1>
            <span className="px-2.5 py-0.5 text-[10px] font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-full">
              Phase 40 Real
            </span>
          </div>
          <p className="text-xs text-slate-400 mt-1">
            Tenant-isolated feedback management, status lifecycles, and aggregate ratings
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-slate-400">Target Dispatch Email:</span>
          <span className="font-mono text-xs font-bold text-blue-400 bg-blue-500/10 px-2.5 py-1 rounded-lg border border-blue-500/20">
            23w61a0506@gmail.com
          </span>
        </div>
      </div>

      {/* Analytics KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <CloudCard className="p-4 border-slate-800 bg-slate-900/60">
          <div className="text-xs font-medium text-slate-400">Total Submissions</div>
          <div className="text-2xl font-bold text-white mt-1">{analytics.totalFeedback}</div>
          <div className="text-[11px] text-slate-500 mt-1">Persisted in Anarva store</div>
        </CloudCard>

        <CloudCard className="p-4 border-slate-800 bg-slate-900/60">
          <div className="text-xs font-medium text-slate-400">Average Rating</div>
          <div className="text-2xl font-bold text-amber-400 mt-1 flex items-center gap-2">
            <span>{analytics.averageRating.toFixed(1)}</span>
            <span className="text-xs text-slate-400 font-normal">/ 5.0</span>
          </div>
          <div className="text-[11px] text-slate-500 mt-1">From verified console users</div>
        </CloudCard>

        <CloudCard className="p-4 border-slate-800 bg-slate-900/60">
          <div className="text-xs font-medium text-slate-400">New & Reviewing</div>
          <div className="text-2xl font-bold text-blue-400 mt-1">
            {(analytics.statusCounts.NEW || 0) + (analytics.statusCounts.REVIEWING || 0)}
          </div>
          <div className="text-[11px] text-slate-500 mt-1">Awaiting roadmap triage</div>
        </CloudCard>

        <CloudCard className="p-4 border-slate-800 bg-slate-900/60">
          <div className="text-xs font-medium text-slate-400">Resolved Items</div>
          <div className="text-2xl font-bold text-emerald-400 mt-1">
            {analytics.statusCounts.RESOLVED || 0}
          </div>
          <div className="text-[11px] text-slate-500 mt-1">Addressed in releases</div>
        </CloudCard>
      </div>

      {/* Controls & Filter Bar */}
      <CloudCard className="p-4 border-slate-800 bg-slate-900/40 space-y-4">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
          <div className="flex items-center gap-3 w-full sm:w-auto">
            {/* Status Filter */}
            <div className="flex items-center gap-2 text-xs">
              <span className="text-slate-400 font-medium">Status:</span>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="px-3 py-1.5 bg-slate-950 border border-slate-800 rounded-xl text-slate-200 focus:outline-none focus:border-blue-500/60"
              >
                <option value="ALL">All Statuses</option>
                <option value="NEW">NEW</option>
                <option value="REVIEWING">REVIEWING</option>
                <option value="PLANNED">PLANNED</option>
                <option value="IN_PROGRESS">IN_PROGRESS</option>
                <option value="RESOLVED">RESOLVED</option>
                <option value="CLOSED">CLOSED</option>
              </select>
            </div>

            {/* Min Rating Filter */}
            <div className="flex items-center gap-2 text-xs">
              <span className="text-slate-400 font-medium">Min Rating:</span>
              <select
                value={minRatingFilter}
                onChange={(e) => setMinRatingFilter(Number(e.target.value))}
                className="px-3 py-1.5 bg-slate-950 border border-slate-800 rounded-xl text-slate-200 focus:outline-none focus:border-blue-500/60"
              >
                <option value={0}>All Ratings</option>
                <option value={5}>5 Stars Only</option>
                <option value={4}>4+ Stars</option>
                <option value={3}>3+ Stars</option>
              </select>
            </div>
          </div>

          <div className="text-xs text-slate-400 font-mono">
            Showing {items.length} feedback entries
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto border border-slate-800 rounded-xl">
          <table className="w-full text-left text-xs text-slate-300">
            <thead className="bg-slate-950/80 text-slate-400 border-b border-slate-800 uppercase font-mono text-[10px]">
              <tr>
                <th className="p-3">Rating</th>
                <th className="p-3">Subject & Category</th>
                <th className="p-3">Submitter</th>
                <th className="p-3">Status</th>
                <th className="p-3">Created Date</th>
                <th className="p-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {items.length === 0 ? (
                <tr>
                  <td colSpan={6} className="p-6 text-center text-slate-500">
                    No feedback entries match the selected filters.
                  </td>
                </tr>
              ) : (
                items.map((fb) => (
                  <tr key={fb.feedbackId} className="hover:bg-slate-800/40 transition">
                    <td className="p-3 font-semibold">{renderStars(fb.rating)}</td>
                    <td className="p-3">
                      <div className="font-bold text-white max-w-xs truncate">{fb.subject}</div>
                      <div className="text-[10px] text-slate-500 font-mono mt-0.5">{fb.category}</div>
                    </td>
                    <td className="p-3">
                      <div className="font-medium text-slate-200">{fb.userEmail}</div>
                      <div className="text-[10px] font-mono text-slate-500">{fb.organizationId}</div>
                    </td>
                    <td className="p-3">
                      <span className={`px-2 py-0.5 text-[10px] font-mono font-bold rounded border ${getStatusBadgeClass(fb.status)}`}>
                        {fb.status}
                      </span>
                    </td>
                    <td className="p-3 text-slate-400 font-mono text-[11px]">
                      {new Date(fb.createdAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                    </td>
                    <td className="p-3 text-right">
                      <button
                        onClick={() => setSelectedFeedback(fb)}
                        className="px-3 py-1 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg text-xs font-medium border border-slate-700 transition"
                      >
                        View Detail
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </CloudCard>

      {/* Feedback Detail Drawer / Modal */}
      {selectedFeedback && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-in fade-in">
          <div className="w-full max-w-2xl bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl overflow-hidden font-sans text-slate-100">
            <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-950">
              <div className="flex items-center gap-3">
                <span className={`px-2.5 py-0.5 text-xs font-mono font-bold rounded border ${getStatusBadgeClass(selectedFeedback.status)}`}>
                  {selectedFeedback.status}
                </span>
                <span className="font-mono text-xs text-slate-400">{selectedFeedback.feedbackId}</span>
              </div>
              <button
                onClick={() => setSelectedFeedback(null)}
                className="p-1.5 text-slate-400 hover:text-white rounded-lg hover:bg-slate-800 transition"
              >
                ✕
              </button>
            </div>

            <div className="p-6 space-y-4">
              <div>
                <div className="flex items-center gap-3 mb-1">
                  {renderStars(selectedFeedback.rating)}
                  <span className="text-xs font-mono font-bold text-amber-400">{selectedFeedback.rating} / 5 Stars</span>
                </div>
                <h3 className="text-lg font-bold text-white">{selectedFeedback.subject}</h3>
                <p className="text-xs text-slate-400 font-mono mt-0.5">Category: {selectedFeedback.category}</p>
              </div>

              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl text-xs text-slate-200 whitespace-pre-wrap font-sans leading-relaxed">
                {selectedFeedback.message}
              </div>

              {/* Context metadata */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 p-3 bg-slate-950/60 border border-slate-800/60 rounded-xl text-xs font-mono">
                <div>
                  <div className="text-[10px] text-slate-500">SUBMITTER EMAIL</div>
                  <div className="text-slate-300 truncate">{selectedFeedback.userEmail}</div>
                </div>
                <div>
                  <div className="text-[10px] text-slate-500">ORGANIZATION</div>
                  <div className="text-slate-300 truncate">{selectedFeedback.organizationId}</div>
                </div>
                <div>
                  <div className="text-[10px] text-slate-500">PROJECT</div>
                  <div className="text-slate-300 truncate">{selectedFeedback.projectId || 'N/A'}</div>
                </div>
                <div>
                  <div className="text-[10px] text-slate-500">REQUEST ID</div>
                  <div className="text-slate-300 truncate">{selectedFeedback.requestId}</div>
                </div>
              </div>

              {/* Status transition controls */}
              <div className="border-t border-slate-800 pt-4">
                <div className="text-xs font-semibold text-slate-300 mb-2">Update Feedback Status:</div>
                <div className="flex flex-wrap gap-2">
                  {(['NEW', 'REVIEWING', 'PLANNED', 'IN_PROGRESS', 'RESOLVED', 'CLOSED'] as const).map((st) => (
                    <button
                      key={st}
                      disabled={updating || selectedFeedback.status === st}
                      onClick={() => handleUpdateStatus(selectedFeedback.feedbackId, st)}
                      className={`px-3 py-1.5 rounded-lg text-xs font-mono font-bold border transition ${
                        selectedFeedback.status === st
                          ? 'bg-blue-600 text-white border-blue-500 opacity-50 cursor-default'
                          : 'bg-slate-950 text-slate-300 hover:text-white border-slate-800 hover:border-slate-700'
                      }`}
                    >
                      {st}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
