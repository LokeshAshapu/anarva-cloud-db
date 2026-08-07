'use client'

import React, { useState } from 'react'

export default function ProjectsPage() {
  const [organizations] = useState([
    { id: 'org-1', name: 'Acme Enterprises', slug: 'acme-enterprises', role: 'OWNER' },
  ])

  const [projects] = useState([
    { id: 'proj-1', name: 'Production Main', slug: 'acme-prod', region: 'us-east-1', dbCount: 2, maxDbs: 5 },
    { id: 'proj-2', name: 'Staging Environment', slug: 'acme-staging', region: 'eu-central-1', dbCount: 1, maxDbs: 5 },
  ])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-white tracking-tight">Projects & Organizations</h1>
        <p className="text-slate-400 mt-1">Multi-tenant organization isolation and project quota management.</p>
      </div>

      {/* Organization Header */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 flex items-center justify-between">
        <div>
          <div className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Active Organization</div>
          <h2 className="text-xl font-bold text-white">{organizations[0].name}</h2>
          <div className="text-xs text-slate-400 font-mono">ID: {organizations[0].id} • Role: {organizations[0].role}</div>
        </div>

        <button className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 text-sm font-semibold rounded-lg transition">
          + Create Organization
        </button>
      </div>

      {/* Projects List */}
      <div className="space-y-4">
        <h2 className="text-lg font-bold text-white">Projects ({projects.length})</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {projects.map((proj) => (
            <div key={proj.id} className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="font-bold text-white text-lg">{proj.name}</h3>
                <span className="px-2.5 py-0.5 text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-full">
                  {proj.region}
                </span>
              </div>

              <div className="text-xs text-slate-400 font-mono">Slug: {proj.slug}</div>

              <div className="pt-2 flex items-center justify-between text-xs text-slate-300 border-t border-slate-800">
                <span>Managed DB Quota:</span>
                <span className="font-semibold text-emerald-400">{proj.dbCount} / {proj.maxDbs} Databases</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
