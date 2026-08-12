'use client'

import React, { useState, useEffect } from 'react'
import Link from 'next/link'

interface SearchItem {
  id: string
  name: string
  type: string
  region: string
  status: string
  href: string
}

interface GlobalSearchModalProps {
  isOpen: boolean
  onClose: () => void
}

export function GlobalSearchModal({ isOpen, onClose }: GlobalSearchModalProps) {
  const [query, setQuery] = useState('')
  const [items, setItems] = useState<SearchItem[]>([])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const sampleItems: SearchItem[] = [
        { id: 'res-db-prod-1', name: 'production-db', type: 'DATABASE', region: 'ap-hyderabad-1', status: 'AVAILABLE', href: '/console/databases' },
        { id: 'res-db-analytics-1', name: 'analytics-db', type: 'DATABASE', region: 'ap-mumbai-1', status: 'AVAILABLE', href: '/console/databases' },
        { id: 'res-s3-assets-1', name: 'anarva-media-assets', type: 'STORAGE', region: 'ap-hyderabad-1', status: 'AVAILABLE', href: '/console/storage' },
        { id: 'res-s3-logs-1', name: 'anarva-audit-logs', type: 'STORAGE', region: 'ap-mumbai-1', status: 'AVAILABLE', href: '/console/storage' },
        { id: 'res-ace-worker-1', name: 'ace-worker-node-01', type: 'COMPUTE', region: 'us-east-1', status: 'RUNNING', href: '/console/compute' },
        { id: 'res-vpc-prod-1', name: 'Primary Production VPC', type: 'NETWORK', region: 'us-east-1', status: 'AVAILABLE', href: '/console/networking' },
        { id: 'res-bak-snap-1', name: 'prod-daily-snap-01', type: 'BACKUP', region: 'ap-hyderabad-1', status: 'COMPLETED', href: '/console/backups' },
      ]

      const email = localStorage.getItem('anarva_user_email') || 'lokeshashapu@gmail.com'
      const userNetKey = `anarva_user_networks_${email}`
      const userNets = JSON.parse(localStorage.getItem(userNetKey) || '[]')
      const customItems = userNets.map((n: any) => ({
        id: n.id,
        name: n.name,
        type: 'NETWORK',
        region: n.regionId,
        status: n.status,
        href: '/console/networking',
      }))

      setItems([...customItems, ...sampleItems])
    }
  }, [isOpen])

  if (!isOpen) return null

  const filtered = items.filter(
    (item) =>
      item.name.toLowerCase().includes(query.toLowerCase()) ||
      item.type.toLowerCase().includes(query.toLowerCase()) ||
      item.id.toLowerCase().includes(query.toLowerCase())
  )

  return (
    <div className="fixed inset-0 bg-black/75 backdrop-blur-sm flex items-start justify-center pt-20 p-4 z-50">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-xl w-full p-4 space-y-4 shadow-2xl">
        <div className="flex items-center justify-between border-b border-slate-800 pb-3">
          <input
            type="text"
            autoFocus
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search databases, compute, buckets, networks, backups... (Ctrl+K)"
            className="w-full bg-transparent text-white font-mono text-sm focus:outline-none"
          />
          <button onClick={onClose} className="text-slate-400 text-xs font-mono px-2 py-1 bg-slate-800 rounded">
            ESC
          </button>
        </div>

        <div className="max-h-80 overflow-y-auto space-y-2 divide-y divide-slate-800">
          {filtered.length === 0 ? (
            <div className="text-center py-6 text-slate-500 font-mono text-xs">No resources matching '{query}'</div>
          ) : (
            filtered.map((item) => (
              <Link
                key={item.id}
                href={item.href}
                onClick={onClose}
                className="flex items-center justify-between p-3 hover:bg-slate-800 rounded-xl transition text-xs font-mono block"
              >
                <div>
                  <div className="font-bold text-white font-sans flex items-center gap-2">
                    {item.name}
                    <span className="text-[10px] px-2 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded">
                      {item.type}
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-0.5">
                    {item.id} • Region: {item.region}
                  </div>
                </div>

                <span className="text-[10px] px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded font-bold">
                  {item.status}
                </span>
              </Link>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
