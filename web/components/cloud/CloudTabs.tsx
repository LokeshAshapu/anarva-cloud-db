'use client'

import React from 'react'

export interface TabItem {
  id: string
  label: string
  count?: number
}

export interface CloudTabsProps {
  tabs: TabItem[]
  activeTab: string
  onChange: (id: string) => void
  className?: string
}

export function CloudTabs({ tabs, activeTab, onChange, className = '' }: CloudTabsProps) {
  return (
    <div className={`flex items-center gap-2 border-b border-slate-800 pb-3 text-xs font-semibold overflow-x-auto max-w-full flex-nowrap scrollbar-none select-none ${className}`}>
      {tabs.map((tab) => {
        const isActive = activeTab === tab.id
        return (
          <button
            key={tab.id}
            onClick={() => onChange(tab.id)}
            className={`px-3.5 py-2 rounded-xl transition flex items-center gap-2 flex-shrink-0 whitespace-nowrap ${
              isActive
                ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20 font-bold shadow-sm'
                : 'text-slate-400 hover:bg-slate-900 hover:text-slate-200'
            }`}
          >
            <span>{tab.label}</span>
            {tab.count !== undefined && (
              <span className={`px-1.5 py-0.5 rounded text-[10px] font-mono ${isActive ? 'bg-blue-500/20 text-blue-300' : 'bg-slate-800 text-slate-400'}`}>
                {tab.count}
              </span>
            )}
          </button>
        )
      })}
    </div>
  )
}
