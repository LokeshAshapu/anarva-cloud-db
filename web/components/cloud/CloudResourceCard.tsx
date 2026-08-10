'use client'

import React from 'react'
import { CloudResource } from '@/types/resource'
import { CloudStatus } from './CloudStatus'

export interface CloudResourceCardProps {
  resource: CloudResource
  action?: React.ReactNode
  onClick?: () => void
}

export function CloudResourceCard({ resource, action, onClick }: CloudResourceCardProps) {
  return (
    <div
      onClick={onClick}
      className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-3 hover:border-slate-700 transition shadow-xl cursor-pointer flex flex-col justify-between"
    >
      <div className="space-y-2">
        <div className="flex items-start justify-between">
          <div>
            <h4 className="font-bold text-white text-sm">{resource.name}</h4>
            <div className="text-[11px] text-slate-400 font-mono mt-0.5">{resource.type} • {resource.region}</div>
          </div>
          <CloudStatus status={resource.status} />
        </div>

        {resource.tags && resource.tags.length > 0 && (
          <div className="flex flex-wrap gap-1 pt-1">
            {resource.tags.map((t, idx) => (
              <span key={idx} className="px-2 py-0.5 bg-slate-950 border border-slate-800 text-[10px] text-slate-400 font-mono rounded">
                {t.key}:{t.value}
              </span>
            ))}
          </div>
        )}
      </div>

      {action && <div className="pt-3 border-t border-slate-800 flex justify-end">{action}</div>}
    </div>
  )
}
