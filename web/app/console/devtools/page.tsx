'use client'

import React, { useState } from 'react'

export default function DevToolsPage() {
  const [activeTab, setActiveTab] = useState<'API_KEYS' | 'CLI' | 'SDK'>('API_KEYS')

  const sampleCli = `# Install Anarva Cloud CLI
curl -sSL https://cli.anarva.io/install.sh | bash

# Authenticate with credentials
anarva login --token anarva_live_8f3a921b

# Provision a new PostgreSQL cluster
anarva db create --name prod-db --engine postgres --acu 2.0 --region us-east-1`

  const sampleSdk = `import { AnarvaCloud } from '@anarva/sdk';

const cloud = new AnarvaCloud({
  apiKey: process.env.ANARVA_API_KEY,
  region: 'us-east-1',
});

// Provision database
const cluster = await cloud.databases.create({
  name: 'prod-cluster',
  engine: 'postgres',
  acu: 1.0,
});`

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Developer Tools & SDK Docs</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">API Key management, CLI architecture, Node/Python SDKs, and gRPC schemas.</p>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex items-center gap-2 border-b border-slate-800 pb-3 text-xs font-semibold">
        <button
          onClick={() => setActiveTab('API_KEYS')}
          className={`px-4 py-2 rounded-xl transition ${
            activeTab === 'API_KEYS' ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20' : 'text-slate-400 hover:bg-slate-900'
          }`}
        >
          API Keys & Service Accounts
        </button>
        <button
          onClick={() => setActiveTab('CLI')}
          className={`px-4 py-2 rounded-xl transition ${
            activeTab === 'CLI' ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20' : 'text-slate-400 hover:bg-slate-900'
          }`}
        >
          Anarva CLI (`anarva`)
        </button>
        <button
          onClick={() => setActiveTab('SDK')}
          className={`px-4 py-2 rounded-xl transition ${
            activeTab === 'SDK' ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20' : 'text-slate-400 hover:bg-slate-900'
          }`}
        >
          Node.js & Python SDKs
        </button>
      </div>

      {/* Tab Contents */}
      {activeTab === 'API_KEYS' && (
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
          <h2 className="text-base font-bold text-white">Active API Secret Keys</h2>
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
            <div className="p-4 bg-slate-950 flex items-center justify-between">
              <div>
                <div className="font-bold text-white font-sans">Production CLI Deployment Key</div>
                <div className="text-slate-400 text-[11px] mt-0.5">Prefix: anarva_live_8f3a_*** • Scope: ADMIN_FULL_ACCESS</div>
              </div>
              <button className="px-3 py-1 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg text-[11px] font-sans">
                Copy Token
              </button>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'CLI' && (
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
          <h2 className="text-base font-bold text-white">Anarva Command Line Interface (CLI)</h2>
          <pre className="p-4 bg-slate-950 border border-slate-800 rounded-xl font-mono text-xs text-blue-300 overflow-x-auto">
            {sampleCli}
          </pre>
        </div>
      )}

      {activeTab === 'SDK' && (
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
          <h2 className="text-base font-bold text-white">Anarva Cloud SDK Integration</h2>
          <pre className="p-4 bg-slate-950 border border-slate-800 rounded-xl font-mono text-xs text-blue-300 overflow-x-auto">
            {sampleSdk}
          </pre>
        </div>
      )}
    </div>
  )
}
