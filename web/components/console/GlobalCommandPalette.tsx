'use client'

import React, { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'

interface SearchItem {
  id: string
  name: string
  category: 'RESOURCES' | 'PAGES' | 'ACTIONS' | 'DOCUMENTATION'
  type?: string
  region?: string
  href: string
}

interface GlobalCommandPaletteProps {
  isOpen: boolean
  onClose: () => void
}

export function GlobalCommandPalette({ isOpen, onClose }: GlobalCommandPaletteProps) {
  const router = useRouter()
  const [query, setQuery] = useState('')

  const items: SearchItem[] = [
    // Resources
    { id: 'res-1', name: 'production-db (arnv:db:ap-hyderabad-1:proj-default:database/production-db)', category: 'RESOURCES', type: 'DATABASE', region: 'ap-hyderabad-1', href: '/console/databases' },
    { id: 'res-2', name: 'analytics-db (arnv:db:ap-mumbai-1:proj-default:database/analytics-db)', category: 'RESOURCES', type: 'DATABASE', region: 'ap-mumbai-1', href: '/console/databases' },
    { id: 'res-3', name: 'anarva-media-assets (arnv:s3:ap-hyderabad-1:proj-default:storage/anarva-media-assets)', category: 'RESOURCES', type: 'STORAGE_BUCKET', region: 'ap-hyderabad-1', href: '/console/storage' },
    { id: 'res-4', name: 'ace-worker-node-01 (arnv:vm:ap-hyderabad-1:proj-default:compute/ace-worker-node-01)', category: 'RESOURCES', type: 'COMPUTE', region: 'ap-hyderabad-1', href: '/console/compute' },
    
    // Actions
    { id: 'act-1', name: 'Create Database Cluster', category: 'ACTIONS', href: '/console/databases' },
    { id: 'act-2', name: 'Create Object Storage Bucket', category: 'ACTIONS', href: '/console/storage' },
    { id: 'act-3', name: 'Deploy Compute Node (ACE)', category: 'ACTIONS', href: '/console/compute' },

    // Pages
    { id: 'page-1', name: 'Home Infrastructure Overview', category: 'PAGES', href: '/console' },
    { id: 'page-2', name: 'Managed Databases & SQL IDE', category: 'PAGES', href: '/console/databases' },
    { id: 'page-3', name: 'Anarva Object Storage (AOS)', category: 'PAGES', href: '/console/storage' },
    { id: 'page-4', name: 'Anarva Compute Engine (ACE)', category: 'PAGES', href: '/console/compute' },
    { id: 'page-5', name: 'IAM & Access Control', category: 'PAGES', href: '/console/iam' },
    { id: 'page-6', name: 'Platform Settings & Default Region', category: 'PAGES', href: '/console/settings' },
  ]

  const [userItems, setUserItems] = useState<SearchItem[]>([])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'lokeshashapu@gmail.com'
      
      const userNets = JSON.parse(localStorage.getItem(`anarva_user_networks_${email}`) || '[]')
      const userDbs = JSON.parse(localStorage.getItem(`anarva_user_databases_${email}`) || '[]')
      const userCompute = JSON.parse(localStorage.getItem(`anarva_user_compute_${email}`) || '[]')
      const userBuckets = JSON.parse(localStorage.getItem(`anarva_user_buckets_${email}`) || '[]')

      const dynamic: SearchItem[] = [
        ...userNets.map((n: any) => ({ id: n.id, name: `${n.name} (${n.cidr})`, category: 'RESOURCES' as const, type: 'NETWORK', region: n.regionId, href: '/console/networking' })),
        ...userDbs.map((d: any) => ({ id: d.id, name: d.name, category: 'RESOURCES' as const, type: 'DATABASE', region: d.regionId || 'us-east-1', href: '/console/databases' })),
        ...userCompute.map((c: any) => ({ id: c.id, name: c.name, category: 'RESOURCES' as const, type: 'COMPUTE', region: c.zone || 'us-east-1', href: '/console/compute' })),
        ...userBuckets.map((b: any) => ({ id: b.id, name: b.name, category: 'RESOURCES' as const, type: 'STORAGE_BUCKET', region: b.regionId || 'us-east-1', href: '/console/storage' })),
      ]
      setUserItems(dynamic)
    }
  }, [isOpen])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault()
        if (isOpen) onClose()
        else setQuery('')
      }
      if (e.key === 'Escape' && isOpen) {
        onClose()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  if (!isOpen) return null

  const allItems = [...userItems, ...items]

  const filteredItems = allItems.filter((item) => {
    if (!query) return true
    const q = query.toLowerCase()
    return (
      item.name.toLowerCase().includes(q) ||
      item.category.toLowerCase().includes(q) ||
      (item.type && item.type.toLowerCase().includes(q)) ||
      (item.region && item.region.toLowerCase().includes(q))
    )
  })

  const handleSelect = (href: string) => {
    router.push(href)
    onClose()
  }

  return (
    <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 flex items-start justify-center pt-16 sm:pt-24 p-4 animate-in fade-in">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-xl shadow-2xl overflow-hidden space-y-3 p-4">
        {/* Search Header */}
        <div className="flex items-center gap-3 border-b border-slate-800 pb-3 px-2">
          <svg className="w-4 h-4 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search resources, ARNV, pages, actions... (ESC to close)"
            className="w-full bg-transparent text-white placeholder-slate-500 text-sm focus:outline-none font-sans"
            autoFocus
          />
          <button onClick={onClose} className="px-2 py-0.5 bg-slate-800 rounded text-[10px] text-slate-400 font-mono">
            ESC
          </button>
        </div>

        {/* Results List */}
        <div className="max-h-80 overflow-y-auto space-y-1 text-xs">
          {filteredItems.length === 0 ? (
            <div className="p-6 text-center text-slate-500">
              No matching resources or actions found for &quot;{query}&quot;.
            </div>
          ) : (
            filteredItems.map((item) => (
              <button
                key={item.id}
                onClick={() => handleSelect(item.href)}
                className="w-full text-left px-3 py-2.5 rounded-xl hover:bg-slate-800 flex items-center justify-between transition group"
              >
                <div className="truncate">
                  <div className="font-semibold text-slate-200 group-hover:text-white truncate">{item.name}</div>
                  {item.type && (
                    <div className="text-[10px] text-slate-500 font-mono mt-0.5">
                      Type: {item.type} • Region: {item.region}
                    </div>
                  )}
                </div>
                <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-slate-950 border border-slate-800 text-slate-400 group-hover:border-blue-500/40 group-hover:text-blue-400">
                  {item.category}
                </span>
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
