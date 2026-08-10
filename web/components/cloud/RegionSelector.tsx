'use client'

import React, { useEffect, useState } from 'react'
import { CloudRegion } from '@/types/resource'

const REGION_KEY = 'anarva_selected_region'

export interface RegionOption {
  id: CloudRegion | string
  name: string
  location: string
  status: 'AVAILABLE' | 'COMING_SOON'
}

export function RegionSelector() {
  const [selectedRegion, setSelectedRegion] = useState<string>('ap-south-1')

  const regions: RegionOption[] = [
    { id: 'ap-south-2', name: 'Asia Pacific (Hyderabad)', location: 'hyderabad', status: 'AVAILABLE' },
    { id: 'ap-south-1', name: 'Asia Pacific (Mumbai)', location: 'mumbai', status: 'AVAILABLE' },
    { id: 'ap-southeast-1', name: 'Asia Pacific (Singapore)', location: 'singapore', status: 'AVAILABLE' },
    { id: 'us-east-1', name: 'US East (N. Virginia)', location: 'virginia', status: 'AVAILABLE' },
    { id: 'eu-west-1', name: 'Europe West (Frankfurt)', location: 'frankfurt', status: 'AVAILABLE' },
    { id: 'sa-east-1', name: 'South America (São Paulo)', location: 'sao-paulo', status: 'COMING_SOON' },
    { id: 'me-central-1', name: 'Middle East (UAE)', location: 'uae', status: 'COMING_SOON' },
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
            {r.name} {r.status === 'COMING_SOON' ? '(Coming Soon)' : ''}
          </option>
        ))}
      </select>
    </div>
  )
}
