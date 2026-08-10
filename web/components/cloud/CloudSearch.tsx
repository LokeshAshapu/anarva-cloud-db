'use client'

import React from 'react'

export interface CloudSearchProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
}

export function CloudSearch({ value, onChange, placeholder = 'Search resources...' }: CloudSearchProps) {
  return (
    <div className="relative w-full max-w-xs text-xs">
      <svg className="w-4 h-4 text-slate-400 absolute left-3 top-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full bg-slate-950 border border-slate-800 rounded-xl pl-9 pr-3 py-2 text-white focus:outline-none focus:border-blue-500 font-sans"
      />
    </div>
  )
}
