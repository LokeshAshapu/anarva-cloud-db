'use client'

import React, { useState } from 'react'

export default function ShareViewPage({ params }: { params: { token: string } }) {
  const [copied, setCopied] = useState(false)
  const shareToken = params.token || 'demo-token'

  const mockExportData = {
    tableName: 'customer_orders',
    databaseName: 'Primary Application Database',
    format: 'CSV / JSON / SQL',
    accessLevel: 'ANYONE_WITH_LINK',
    created_at: '2026-08-08',
    columns: ['id', 'customer_name', 'amount', 'status', 'created_at'],
    rows: [
      { id: 1, customer_name: 'Lokesh Ashapu', amount: '299.99', status: 'COMPLETED', created_at: '2026-08-08 07:00:00' },
      { id: 2, customer_name: 'Enterprise Client', amount: '1499.00', status: 'PROCESSING', created_at: '2026-08-08 07:05:00' },
      { id: 3, customer_name: 'Acme Corp', amount: '850.50', status: 'PAID', created_at: '2026-08-08 07:10:00' },
    ],
  }

  const handleDownloadCSV = () => {
    const headers = mockExportData.columns.join(',')
    const rows = mockExportData.rows.map((r) => `${r.id},"${r.customer_name}",${r.amount},${r.status},"${r.created_at}"`).join('\n')
    const csvContent = `data:text/csv;charset=utf-8,${headers}\n${rows}`
    const encodedUri = encodeURI(csvContent)
    const link = document.createElement('a')
    link.setAttribute('href', encodedUri)
    link.setAttribute('download', `${mockExportData.tableName}_export.csv`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const handleDownloadJSON = () => {
    const jsonStr = JSON.stringify(mockExportData.rows, null, 2)
    const blob = new Blob([jsonStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${mockExportData.tableName}_export.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const handleCopyLink = () => {
    if (typeof window !== 'undefined') {
      navigator.clipboard.writeText(window.location.href)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-6 md:p-12">
      <div className="max-w-4xl mx-auto space-y-8">
        {/* Header */}
        <header className="flex items-center justify-between border-b border-slate-800 pb-6">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-gradient-to-tr from-blue-600 to-cyan-400 flex items-center justify-center font-bold text-white shadow-lg text-xl">
              A
            </div>
            <div>
              <h1 className="text-xl font-bold text-white">Anarva Shared Database Snapshot</h1>
              <p className="text-xs text-slate-400 font-mono">Token: {shareToken}</p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <span className="px-3 py-1 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold rounded-full flex items-center gap-1.5">
              <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
              Google Drive Access: Anyone with Link
            </span>

            <button
              onClick={handleCopyLink}
              className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-lg transition border border-slate-700"
            >
              {copied ? '✔ Link Copied!' : 'Copy Share Link'}
            </button>
          </div>
        </header>

        {/* Database Export Overview */}
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <span className="text-xs text-slate-400 uppercase font-semibold">Shared Table</span>
              <h2 className="text-2xl font-extrabold text-white">{mockExportData.tableName}</h2>
              <p className="text-xs text-slate-400 mt-0.5">Source Database: {mockExportData.databaseName}</p>
            </div>

            <div className="flex gap-2">
              <button
                onClick={handleDownloadCSV}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold text-xs rounded-lg transition shadow-lg shadow-blue-600/25"
              >
                Download CSV
              </button>
              <button
                onClick={handleDownloadJSON}
                className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 font-semibold text-xs rounded-lg transition border border-slate-700"
              >
                Download JSON
              </button>
            </div>
          </div>

          {/* Table Data Preview */}
          <div className="overflow-x-auto border border-slate-800 rounded-xl mt-4">
            <table className="w-full text-left text-xs font-mono">
              <thead className="bg-slate-950 text-slate-300 border-b border-slate-800">
                <tr>
                  {mockExportData.columns.map((col, idx) => (
                    <th key={idx} className="p-3.5 border-r border-slate-800 last:border-r-0">
                      {col}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800 text-slate-200">
                {mockExportData.rows.map((row, rIdx) => (
                  <tr key={rIdx} className="hover:bg-slate-800/40 transition">
                    <td className="p-3.5 border-r border-slate-800 text-white font-semibold">{row.id}</td>
                    <td className="p-3.5 border-r border-slate-800 text-blue-400">{row.customer_name}</td>
                    <td className="p-3.5 border-r border-slate-800">{row.amount}</td>
                    <td className="p-3.5 border-r border-slate-800">
                      <span className="px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full font-semibold">
                        {row.status}
                      </span>
                    </td>
                    <td className="p-3.5 text-slate-400">{row.created_at}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
