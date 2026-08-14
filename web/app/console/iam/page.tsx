'use client'

import React, { useState, useEffect } from 'react'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudMetric } from '@/components/cloud/CloudMetric'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudModal } from '@/components/cloud/CloudModal'
import { API_BASE_URL } from '@/lib/api'
import { createClient } from '@/utils/supabase/client'

interface MemberItem {
  id: string
  name: string
  email: string
  role: string
  status: string
}

export default function IAMPage() {
  const [userEmail, setUserEmail] = useState('operator@anarva.internal')
  const [userName, setUserName] = useState('Cloud Operator')
  const [inviteModalOpen, setInviteModalOpen] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('DEVELOPER')
  const [isInviting, setIsInviting] = useState(false)

  const [members, setMembers] = useState<MemberItem[]>([
    { id: 'usr-87a1', name: 'Cloud Operator', email: 'operator@anarva.internal', role: 'OWNER', status: 'ACTIVE' },
  ])

  useEffect(() => {
    let email = 'user@anarva.io'
    let name = 'Account User'

    if (typeof window !== 'undefined') {
      const storedEmail = localStorage.getItem('anarva_user_email')
      const storedName = localStorage.getItem('anarva_user_name')
      if (storedEmail) email = storedEmail
      if (storedName) name = storedName
    }

    try {
      const supabase = createClient()
      supabase.auth.getUser().then(({ data }) => {
        if (data?.user?.email) {
          email = data.user.email
          localStorage.setItem('anarva_user_email', email)
          if (data.user.user_metadata?.full_name) {
            name = data.user.user_metadata.full_name
            localStorage.setItem('anarva_user_name', name)
          }
          setUserEmail(email)
          setUserName(name)
          setMembers([{ id: 'usr-current', name, email, role: 'OWNER', status: 'ACTIVE' }])
        }
      })
    } catch (e) {}

    setUserEmail(email)
    setUserName(name)
    setMembers([{ id: 'usr-current', name, email, role: 'OWNER', status: 'ACTIVE' }])
  }, [])

  const handleInviteMember = () => {
    if (!inviteEmail) return
    setIsInviting(true)
    setTimeout(() => {
      const newMem: MemberItem = {
        id: `mem-${Date.now()}`,
        name: inviteEmail.split('@')[0],
        email: inviteEmail,
        role: inviteRole,
        status: 'INVITED',
      }
      setMembers([...members, newMem])
      setIsInviting(false)
      setInviteModalOpen(false)
      setInviteEmail('')
    }, 1000)
  }

  const [policyJson, setPolicyJson] = useState(`{
  "Version": "2026-08-11",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["database:*", "storage:*", "compute:*", "network:*"],
      "Resource": "arnv:*:ap-hyderabad-1:proj-default:*"
    }
  ]
}`)

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Identity & Access Management (IAM)</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">
            Manage organization members, team roles, zero-trust RBAC access models, and JSON permissions.
          </p>
        </div>

        <CloudButton variant="primary" size="sm" onClick={() => setInviteModalOpen(true)}>
          + Invite Team Member
        </CloudButton>
      </div>

      {/* Summary Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <CloudMetric label="Organization Members" value={`${members.length} Members`} subtext="Active Owner & Team" trend="RBAC" trendType="positive" />
        <CloudMetric label="Access Control Model" value="RBAC + JSON" subtext="Zero-Trust Enforced" trend="ACTIVE" trendType="positive" />
        <CloudMetric label="Active Service Accounts" value="1 Account" subtext="CI/CD Deployer Token" trend="CONFIGURED" trendType="positive" />
      </div>

      {/* Active Members Table */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4 shadow-xl">
        <h2 className="text-base font-bold text-white">Organization Account Members</h2>

        <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
          {members.map((mem) => (
            <div key={mem.id} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <div className="font-bold text-white text-sm">{mem.name}</div>
                <div className="text-slate-400 text-[11px] mt-0.5">{mem.email} • ID: {mem.id}</div>
              </div>

              <div className="flex items-center gap-3">
                <span className="px-2.5 py-0.5 rounded text-[10px] font-bold bg-blue-600/10 text-blue-400 border border-blue-500/20">
                  ROLE: {mem.role}
                </span>
                <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold border ${mem.status === 'ACTIVE' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border-amber-500/20'}`}>
                  {mem.status}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* JSON Access Policy Editor */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4 shadow-xl">
        <h2 className="text-base font-bold text-white">Granular IAM JSON Access Policy Editor</h2>
        <textarea
          value={policyJson}
          onChange={(e) => setPolicyJson(e.target.value)}
          rows={7}
          className="w-full bg-slate-950 border border-slate-800 rounded-xl p-4 font-mono text-xs text-blue-300 focus:outline-none focus:border-blue-500"
        ></textarea>
      </div>

      {/* Member Invite Modal */}
      {inviteModalOpen && (
        <CloudModal
          isOpen={inviteModalOpen}
          onClose={() => setInviteModalOpen(false)}
          title="Invite Organization Member"
          subtitle="Send an invitation with specific role and project scope"
        >
          <div className="space-y-4 text-xs">
            <div className="space-y-1">
              <label className="block text-slate-300 font-semibold">User Email Address</label>
              <input
                type="email"
                value={inviteEmail}
                onChange={(e) => setInviteEmail(e.target.value)}
                placeholder="colleague@company.com"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none"
              />
            </div>
            <div className="space-y-1">
              <label className="block text-slate-300 font-semibold">Assign Role</label>
              <select
                value={inviteRole}
                onChange={(e) => setInviteRole(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none cursor-pointer"
              >
                <option value="ADMIN">ADMIN (Full Project Access)</option>
                <option value="DEVELOPER">DEVELOPER (Database & Storage Read/Write)</option>
                <option value="DATABASE_ADMIN">DATABASE_ADMIN (Databases Only)</option>
                <option value="STORAGE_ADMIN">STORAGE_ADMIN (Object Storage Only)</option>
                <option value="VIEWER">VIEWER (Read-Only Observer)</option>
              </select>
            </div>
            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2">
              <CloudButton variant="outline" size="sm" onClick={() => setInviteModalOpen(false)}>
                Cancel
              </CloudButton>
              <CloudButton variant="primary" size="sm" isLoading={isInviting} onClick={handleInviteMember}>
                Send Invitation
              </CloudButton>
            </div>
          </div>
        </CloudModal>
      )}
    </div>
  )
}
