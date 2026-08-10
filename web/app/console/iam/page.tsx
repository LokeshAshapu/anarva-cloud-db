'use client'

import React, { useState, useEffect } from 'react'

export default function IAMPage() {
  const [userEmail, setUserEmail] = useState('lokeshashapu@gmail.com')
  const [userName, setUserName] = useState('Lokesh Ashapu')

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email')
      const name = localStorage.getItem('anarva_user_name')
      if (email) setUserEmail(email)
      if (name) setUserName(name)
    }
  }, [])

  const members = [
    { id: 'usr-87a1', name: userName, email: userEmail, role: 'OWNER', status: 'ACTIVE' },
  ]

  const [policyJson, setPolicyJson] = useState(`{
  "Version": "2026-08-10",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["database:*", "storage:*", "compute:*"],
      "Resource": "arn:anarva:cloud:us-east-1:org-default:*"
    }
  ]
}`)

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Identity and Access Management (IAM)</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Manage organization team members, roles, fine-grained access policies, and service accounts.</p>
        </div>
      </div>

      {/* IAM Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Organization Members</div>
          <div className="text-3xl font-extrabold text-white font-mono">1 Active Owner</div>
          <div className="text-xs text-slate-400">Owner Access Granted to {userEmail}</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Access Control Model</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">RBAC + JSON Policy</div>
          <div className="text-xs text-slate-400">Zero-Trust Isolation Active</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Service Accounts</div>
          <div className="text-3xl font-extrabold text-emerald-400 font-mono">1 Active</div>
          <div className="text-xs text-slate-400">CLI & SDK Automated Tokens</div>
        </div>
      </div>

      {/* Team Members List */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
        <h2 className="text-base font-bold text-white">Active Organization Account</h2>

        <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
          {members.map((mem) => (
            <div key={mem.id} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <div className="font-bold text-white text-sm">{mem.name}</div>
                <div className="text-slate-400 text-[11px] font-mono mt-0.5">{mem.email} • ID: {mem.id}</div>
              </div>

              <div className="flex items-center gap-4 font-mono">
                <span className="px-2.5 py-0.5 rounded text-[10px] font-bold bg-blue-600/10 text-blue-400 border border-blue-500/20">
                  {mem.role}
                </span>
                <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  {mem.status}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* JSON Access Policy Editor */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
        <h2 className="text-base font-bold text-white">Fine-Grained IAM Policy Editor</h2>
        <textarea
          value={policyJson}
          onChange={(e) => setPolicyJson(e.target.value)}
          rows={8}
          className="w-full bg-slate-950 border border-slate-800 rounded-xl p-4 font-mono text-xs text-blue-300 focus:outline-none focus:border-blue-500"
        ></textarea>
      </div>
    </div>
  )
}
