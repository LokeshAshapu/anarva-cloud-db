'use client'

import React from 'react'

export interface CloudInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  helperText?: string
}

export function CloudInput({ label, error, helperText, className = '', ...props }: CloudInputProps) {
  return (
    <div className="space-y-1.5 w-full text-xs">
      {label && <label className="block text-slate-300 font-semibold">{label}</label>}
      <input
        className={`w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2.5 text-white focus:outline-none focus:border-blue-500 font-sans transition ${
          error ? 'border-red-500/50' : ''
        } ${className}`}
        {...props}
      />
      {error && <p className="text-[11px] text-red-400 font-mono">{error}</p>}
      {helperText && !error && <p className="text-[11px] text-slate-500">{helperText}</p>}
    </div>
  )
}
