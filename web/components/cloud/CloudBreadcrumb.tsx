'use client'

import React from 'react'
import Link from 'next/link'

export interface BreadcrumbItem {
  label: string
  href?: string
}

export interface CloudBreadcrumbProps {
  items: BreadcrumbItem[]
}

export function CloudBreadcrumb({ items }: CloudBreadcrumbProps) {
  return (
    <nav className="flex items-center gap-2 text-xs font-mono text-slate-400">
      {items.map((item, idx) => {
        const isLast = idx === items.length - 1
        return (
          <React.Fragment key={idx}>
            {idx > 0 && <span className="text-slate-600">/</span>}
            {item.href && !isLast ? (
              <Link href={item.href} className="hover:text-blue-400 transition">
                {item.label}
              </Link>
            ) : (
              <span className={isLast ? 'text-white font-bold' : ''}>{item.label}</span>
            )}
          </React.Fragment>
        )
      })}
    </nav>
  )
}
