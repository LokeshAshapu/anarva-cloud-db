'use client'

import React from 'react'

export interface SelectOption {
  value: string
  label: string
}

export interface CloudSelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label?: string
  options: SelectOption[]
  error?: string
}

export function CloudSelect({ label, options, error, className = '', ...props }: CloudSelectProps) {
  return (
    <div className="space-y-1.5 w-full text-xs">
      {label && <label className="block text-slate-300 font-semibold">{label}</label>}
      <select
        className={`w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2.5 text-white focus:outline-none focus:border-blue-500 font-sans cursor-pointer transition ${
          error ? 'border-red-500/50' : ''
        } ${className}`}
        {...props}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
      {error && <p className="text-[11px] text-red-400 font-mono">{error}</p>}
    </div>
  )
}
