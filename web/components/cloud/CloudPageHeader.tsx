'use client'

import React from 'react'
import { CloudBreadcrumb, BreadcrumbItem } from './CloudBreadcrumb'

export interface CloudPageHeaderProps {
  title: string
  description?: string
  breadcrumbs?: BreadcrumbItem[]
  action?: React.ReactNode
}

export function CloudPageHeader({ title, description, breadcrumbs, action }: CloudPageHeaderProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
      <div className="space-y-1">
        {breadcrumbs && <CloudBreadcrumb items={breadcrumbs} />}
        <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">{title}</h1>
        {description && <p className="text-slate-400 text-xs sm:text-sm">{description}</p>}
      </div>
      {action && <div className="flex items-center gap-3">{action}</div>}
    </div>
  )
}
