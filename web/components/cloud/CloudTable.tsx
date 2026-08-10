'use client'

import React from 'react'

export interface ColumnDef<T> {
  header: string
  accessorKey?: keyof T
  cell?: (row: T) => React.ReactNode
  className?: string
}

export interface CloudTableProps<T> {
  columns: ColumnDef<T>[]
  data: T[]
  emptyMessage?: string
  isLoading?: boolean
  className?: string
}

export function CloudTable<T extends { id?: string | number }>({
  columns,
  data,
  emptyMessage = 'No resources found.',
  isLoading = false,
  className = '',
}: CloudTableProps<T>) {
  if (isLoading) {
    return (
      <div className="p-8 text-center text-xs text-slate-500 bg-slate-950 rounded-2xl border border-slate-800 space-y-2">
        <span className="inline-block h-4 w-4 rounded-full border-2 border-blue-500 border-t-transparent animate-spin"></span>
        <div>Loading cloud infrastructure resources...</div>
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="p-8 text-center text-xs text-slate-500 bg-slate-950 rounded-2xl border border-slate-800">
        {emptyMessage}
      </div>
    )
  }

  return (
    <div className={`overflow-x-auto border border-slate-800 rounded-2xl ${className}`}>
      <table className="w-full text-left font-sans text-xs divide-y divide-slate-800">
        <thead className="bg-slate-950 text-slate-400 font-bold uppercase text-[10px] tracking-wider">
          <tr>
            {columns.map((col, idx) => (
              <th key={idx} className={`p-4 ${col.className || ''}`}>
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-800/80 bg-slate-900/60 font-mono">
          {data.map((row, rIdx) => (
            <tr key={row.id || rIdx} className="hover:bg-slate-800/40 transition">
              {columns.map((col, cIdx) => (
                <td key={cIdx} className={`p-4 text-slate-200 ${col.className || ''}`}>
                  {col.cell ? col.cell(row) : (col.accessorKey ? String(row[col.accessorKey] ?? '') : null)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
