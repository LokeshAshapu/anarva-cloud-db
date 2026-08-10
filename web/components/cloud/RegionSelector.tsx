'use client'

import React, { useEffect, useState } from 'react'
import { RegionId } from '@/types/resource'

const REGION_KEY = 'anarva_selected_region'

export interface RegionOption {
  id: RegionId | string
  name: string
  displayName: string
  location: string
  status: 'AVAILABLE' | 'COMING_SOON'
}

export function RegionSelector() {
  const [selectedRegion, setSelectedRegion] = useState<string>('ap-hyderabad-1')

  const regions: RegionOption[] = [
    { id: 'ap-hyderabad-1', name: 'ap-hyderabad-1', displayName: 'Asia Pacific — Hyderabad', location: 'Hyderabad', status: 'AVAILABLE' },
    { id: 'ap-mumbai-1', name: 'ap-mumbai-1', displayName: 'Asia Pacific — Mumbai', location: 'Mumbai', status: 'AVAILABLE' },
    { id: 'ap-singapore-1', name: 'ap-singapore-1', displayName: 'Asia Pacific — Singapore', location: 'Singapore', status: 'AVAILABLE' },
    { id: 'us-east-1', name: 'us-east-1', displayName: 'US East — N. Virginia', location: 'Virginia', status: 'AVAILABLE' },
    { id: 'eu-west-1', name: 'eu-west-1', displayName: 'Europe West — Frankfurt', location: 'Frankfurt', status: 'AVAILABLE' },
    { id: 'sa-east-1', name: 'sa-east-1', displayName: 'South America — São Paulo', location: 'São Paulo', status: 'COMING_SOON' },
    { id: 'me-central-1', name: 'me-central-1', displayName: 'Middle East — UAE', location: 'UAE', status: 'COMING_SOON' },
  ]

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(REGION_KEY)
      if (stored) setSelectedRegion(stored)
    }
  }, [])

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const val = e.target.value
    setSelectedRegion(val)
    if (typeof window !== 'undefined') {
      localStorage.setItem(REGION_KEY, val)
    }
  }

  return (
    <div className="relative text-xs">
      <select
        value={selectedRegion}
        onChange={handleChange}
        className="bg-slate-900 border border-slate-800 text-slate-300 rounded-lg px-2.5 py-1.5 text-xs font-medium focus:outline-none focus:border-blue-500 cursor-pointer"
      >
        {regions.map((r) => (
          <option key={r.id} value={r.id} disabled={r.status === 'COMING_SOON'}>
            {r.displayName} {r.status === 'COMING_SOON' ? '(Coming Soon)' : ''}
          </option>
        ))}
      </select>
    </div>
  )
}
