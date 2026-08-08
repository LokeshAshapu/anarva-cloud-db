'use client'

import React, { useEffect, useState } from 'react'
import { API_BASE_URL } from '@/lib/api'

const ORGS_KEY = 'anarva_user_orgs'
const PROJS_KEY = 'anarva_user_projects'

export default function ProjectsPage() {
  const [organizations, setOrganizations] = useState<any[]>([])
  const [projects, setProjects] = useState<any[]>([])
  const [loading, setLoading] = useState<boolean>(true)

  const [showOrgModal, setShowOrgModal] = useState(false)
  const [showProjModal, setShowProjModal] = useState(false)

  const [orgName, setOrgName] = useState('')
  const [projName, setProjName] = useState('')
  const [region, setRegion] = useState('us-east-1')

  const updateOrgsState = (newOrgs: any[]) => {
    setOrganizations(newOrgs)
    if (typeof window !== 'undefined') {
      localStorage.setItem(ORGS_KEY, JSON.stringify(newOrgs))
    }
  }

  const updateProjsState = (newProjs: any[]) => {
    setProjects(newProjs)
    if (typeof window !== 'undefined') {
      localStorage.setItem(PROJS_KEY, JSON.stringify(newProjs))
    }
  }

  const fetchData = async () => {
    let localOrgs: any[] = []
    let localProjs: any[] = []

    if (typeof window !== 'undefined') {
      try {
        localOrgs = JSON.parse(localStorage.getItem(ORGS_KEY) || '[]')
        localProjs = JSON.parse(localStorage.getItem(PROJS_KEY) || '[]')
      } catch {}
    }

    if (localOrgs.length === 0) {
      localOrgs = [{ id: 'org-default', name: 'My Cloud Organization', slug: 'my-cloud-org', role: 'OWNER' }]
    }

    if (localProjs.length === 0) {
      localProjs = [
        { id: 'proj-default', name: 'Main Production Environment', slug: 'main-prod', region: 'us-east-1', dbCount: 0, maxDbs: 5 },
      ]
    }

    try {
      const orgRes = await fetch(`${API_BASE_URL}/api/v1/organizations/org-default`).catch(() => null)
      if (orgRes && orgRes.ok) {
        const oData = await orgRes.json()
        localOrgs = [oData, ...localOrgs.filter((o) => o.id !== oData.id)]
      }

      const projRes = await fetch(`${API_BASE_URL}/api/v1/organizations/org-default/projects`).catch(() => null)
      if (projRes && projRes.ok) {
        const pData = await projRes.json()
        if (Array.isArray(pData) && pData.length > 0) {
          localProjs = pData
        }
      }
    } catch (err) {
      console.error('Failed to load project data', err)
    } finally {
      updateOrgsState(localOrgs)
      updateProjsState(localProjs)
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const handleCreateOrg = (e: React.FormEvent) => {
    e.preventDefault()
    const newOrg = {
      id: `org-${Date.now()}`,
      name: orgName || 'New Organization',
      slug: (orgName || 'new-org').toLowerCase().replace(/\s+/g, '-'),
      role: 'OWNER',
    }
    updateOrgsState([newOrg, ...organizations])
    setOrgName('')
    setShowOrgModal(false)
  }

  const handleCreateProj = (e: React.FormEvent) => {
    e.preventDefault()
    const newProj = {
      id: `proj-${Date.now()}`,
      name: projName || 'New Project',
      slug: (projName || 'new-project').toLowerCase().replace(/\s+/g, '-'),
      region: region,
      dbCount: 0,
      maxDbs: 5,
    }
    updateProjsState([...projects, newProj])
    setProjName('')
    setShowProjModal(false)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Projects & Organizations</h1>
          <p className="text-slate-400 mt-1">Multi-tenant organization isolation and project quota management.</p>
        </div>

        <button
          onClick={() => setShowProjModal(true)}
          className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25"
        >
          + Create Project
        </button>
      </div>

      {/* Organization Header */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 flex items-center justify-between">
        <div>
          <div className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Active Organization</div>
          <h2 className="text-xl font-bold text-white">{organizations[0]?.name || 'My Cloud Organization'}</h2>
          <div className="text-xs text-slate-400 font-mono">
            ID: {organizations[0]?.id || 'org-default'} • Role: {organizations[0]?.role || 'OWNER'}
          </div>
        </div>

        <button
          onClick={() => setShowOrgModal(true)}
          className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 text-sm font-semibold rounded-lg transition border border-slate-700"
        >
          + New Organization
        </button>
      </div>

      {/* Projects List */}
      <div className="space-y-4">
        <h2 className="text-lg font-bold text-white">Projects ({projects.length})</h2>

        {loading ? (
          <div className="p-8 text-center text-slate-400">Loading projects...</div>
        ) : projects.length === 0 ? (
          <div className="bg-slate-900 border border-slate-800 rounded-xl p-8 text-center text-slate-400">
            No projects created yet. Click "+ Create Project" to organize your database instances!
          </div>
        ) : (
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
                  <span className="font-semibold text-emerald-400">
                    {proj.dbCount || 0} / {proj.maxDbs || 5} Databases
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create Org Modal */}
      {showOrgModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Create Organization</h2>
            <form onSubmit={handleCreateOrg} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Organization Name</label>
                <input
                  type="text"
                  required
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  placeholder="e.g. Acme Corp"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>
              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowOrgModal(false)}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25"
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Create Project Modal */}
      {showProjModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Create Project</h2>
            <form onSubmit={handleCreateProj} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Project Name</label>
                <input
                  type="text"
                  required
                  value={projName}
                  onChange={(e) => setProjName(e.target.value)}
                  placeholder="e.g. Production Cluster"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Deployment Region</label>
                <select
                  value={region}
                  onChange={(e) => setRegion(e.target.value)}
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                >
                  <option value="us-east-1">US East (N. Virginia)</option>
                  <option value="eu-central-1">EU Central (Frankfurt)</option>
                  <option value="ap-southeast-1">Asia Pacific (Singapore)</option>
                </select>
              </div>
              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowProjModal(false)}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25"
                >
                  Create Project
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
