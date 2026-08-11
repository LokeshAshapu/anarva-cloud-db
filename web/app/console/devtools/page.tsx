'use client'

import React, { useState, useEffect } from 'react'
import { createClient } from '@/utils/supabase/client'
import { CloudButton } from '@/components/cloud/CloudButton'

export default function DevToolsPage() {
  const [activeTab, setActiveTab] = useState<'API_KEYS' | 'CLI' | 'SDK'>('CLI')
  const [osTab, setOsTab] = useState<'POWERSHELL' | 'CMD' | 'BASH'>('POWERSHELL')
  const [userEmail, setUserEmail] = useState('lokeshashapu@gmail.com')
  const [userName, setUserName] = useState('Lokesh Ashapu')
  const [copiedToken, setCopiedToken] = useState(false)

  useEffect(() => {
    async function loadUser() {
      if (typeof window !== 'undefined') {
        const storedEmail = localStorage.getItem('anarva_user_email')
        const storedName = localStorage.getItem('anarva_user_name')
        if (storedEmail) setUserEmail(storedEmail)
        if (storedName) setUserName(storedName)

        try {
          const supabase = createClient()
          const { data } = await supabase.auth.getUser()
          if (data?.user?.email) {
            setUserEmail(data.user.email)
            localStorage.setItem('anarva_user_email', data.user.email)
            const metaName = data.user.user_metadata?.full_name
            if (metaName) {
              setUserName(metaName)
              localStorage.setItem('anarva_user_name', metaName)
            }
          }
        } catch (e) {
          console.log('Supabase user load notice:', e)
        }
      }
    }
    loadUser()
  }, [])

  const samplePowershell = `# Install Anarva Cloud CLI on Windows PowerShell
iwr -useb https://cli.anarva.io/install.ps1 | iex

# Authenticate with credentials
anarva login --token anarva_live_8f3a921b

# Provision a new PostgreSQL cluster
anarva db create --name prod-db --engine postgres --acu 2.0 --region ap-hyderabad-1`

  const sampleCmd = `rem Install Anarva Cloud CLI on Windows Command Prompt (CMD)
powershell -Command "iwr -useb https://cli.anarva.io/install.ps1 | iex"

rem Authenticate with credentials
anarva login --token anarva_live_8f3a921b

rem Provision a new PostgreSQL cluster
anarva db create --name prod-db --engine postgres --acu 2.0 --region ap-hyderabad-1`

  const sampleBash = `# Install Anarva Cloud CLI on Linux/macOS
curl -sSL https://cli.anarva.io/install.sh | bash

# Authenticate with credentials
anarva login --token anarva_live_8f3a921b

# Provision a new PostgreSQL cluster
anarva db create --name prod-db --engine postgres --acu 2.0 --region ap-hyderabad-1`

  const sampleSdk = `import { AnarvaCloud } from '@anarva/sdk';

const cloud = new AnarvaCloud({
  apiKey: process.env.ANARVA_API_KEY,
  region: 'ap-hyderabad-1',
});

// Provision database
const cluster = await cloud.databases.create({
  name: 'prod-cluster',
  engine: 'postgres',
  acu: 2.0,
});`

  const handleCopyToken = () => {
    navigator.clipboard.writeText('anarva_live_8f3a921b')
    setCopiedToken(true)
    setTimeout(() => setCopiedToken(false), 2000)
  }

  return (
    <div className="space-y-8">
      {/* Header with Active User Badge */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Developer Tools & SDK Docs</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">API Key management, CLI installation for Windows & Linux/macOS, Node/Python SDKs.</p>
        </div>

        {/* Authenticated User Banner */}
        <div className="p-3 bg-slate-900 border border-slate-800 rounded-xl flex items-center gap-3 font-mono text-xs">
          <div className="h-8 w-8 rounded-full bg-blue-600 flex items-center justify-center font-bold text-white shrink-0">
            {userName.substring(0, 2).toUpperCase() || 'LA'}
          </div>
          <div>
            <div className="font-bold text-white font-sans">{userName}</div>
            <div className="text-slate-400 text-[11px]">{userEmail}</div>
          </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex flex-wrap items-center gap-2 border-b border-slate-800 pb-3 text-xs font-semibold">
        <button
          onClick={() => setActiveTab('CLI')}
          className={`px-4 py-2 rounded-xl transition ${
            activeTab === 'CLI' ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20' : 'text-slate-400 hover:bg-slate-900'
          }`}
        >
          Anarva CLI (`anarva`)
        </button>
        <button
          onClick={() => setActiveTab('API_KEYS')}
          className={`px-4 py-2 rounded-xl transition ${
            activeTab === 'API_KEYS' ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20' : 'text-slate-400 hover:bg-slate-900'
          }`}
        >
          API Keys & Tokens
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

      {/* CLI Tab Content */}
      {activeTab === 'CLI' && (
        <div className="space-y-6">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4 shadow-xl">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <h2 className="text-base font-bold text-white">Anarva Command Line Interface (`anarva`)</h2>

              {/* OS Selector Tabs */}
              <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-xl border border-slate-800 font-mono text-[11px]">
                <button
                  onClick={() => setOsTab('POWERSHELL')}
                  className={`px-3 py-1 rounded-lg transition ${osTab === 'POWERSHELL' ? 'bg-blue-600 text-white font-bold' : 'text-slate-400 hover:text-white'}`}
                >
                  Windows PowerShell
                </button>
                <button
                  onClick={() => setOsTab('CMD')}
                  className={`px-3 py-1 rounded-lg transition ${osTab === 'CMD' ? 'bg-blue-600 text-white font-bold' : 'text-slate-400 hover:text-white'}`}
                >
                  Windows CMD
                </button>
                <button
                  onClick={() => setOsTab('BASH')}
                  className={`px-3 py-1 rounded-lg transition ${osTab === 'BASH' ? 'bg-blue-600 text-white font-bold' : 'text-slate-400 hover:text-white'}`}
                >
                  Linux / macOS
                </button>
              </div>
            </div>

            {/* Windows Guidance Alert */}
            {osTab === 'CMD' && (
              <div className="p-3.5 bg-amber-500/10 border border-amber-500/20 rounded-xl text-amber-300 text-xs font-mono">
                <strong>💡 Windows Command Prompt Notice:</strong> In Windows CMD (`C:\Users\username`), `#` is not a comment character and `bash` is not built-in. Use <code>powershell -Command &quot;...&quot;</code> or run the commands directly in <strong>PowerShell</strong> instead.
              </div>
            )}

            <pre className="p-4 bg-slate-950 border border-slate-800 rounded-xl font-mono text-xs text-blue-300 overflow-x-auto select-all">
              {osTab === 'POWERSHELL' && samplePowershell}
              {osTab === 'CMD' && sampleCmd}
              {osTab === 'BASH' && sampleBash}
            </pre>
          </div>
        </div>
      )}

      {/* API Keys Tab Content */}
      {activeTab === 'API_KEYS' && (
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4 shadow-xl">
          <h2 className="text-base font-bold text-white">Authenticated User API Tokens</h2>
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs font-mono">
            <div className="p-4 bg-slate-950 flex items-center justify-between gap-4">
              <div>
                <div className="font-bold text-white font-sans">{userName} — Primary CLI Access Token</div>
                <div className="text-slate-400 text-[11px] mt-0.5">User: {userEmail} • Token: anarva_live_8f3a921b • Scope: ADMIN</div>
              </div>
              <CloudButton variant="secondary" size="sm" onClick={handleCopyToken}>
                {copiedToken ? '✓ Copied!' : 'Copy Token'}
              </CloudButton>
            </div>
          </div>
        </div>
      )}

      {/* SDK Tab Content */}
      {activeTab === 'SDK' && (
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4 shadow-xl">
          <h2 className="text-base font-bold text-white">Anarva Cloud Node.js & Python SDK</h2>
          <pre className="p-4 bg-slate-950 border border-slate-800 rounded-xl font-mono text-xs text-blue-300 overflow-x-auto select-all">
            {sampleSdk}
          </pre>
        </div>
      )}
    </div>
  )
}
