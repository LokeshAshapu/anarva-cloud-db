'use client'

import React, { useEffect, useState } from 'react'

const KEYS_STORAGE_KEY = 'anarva_user_apikeys'

export default function APIKeysPage() {
  const [apiKeys, setApiKeys] = useState<any[]>([])
  const [showModal, setShowModal] = useState(false)
  const [keyName, setKeyName] = useState('')
  const [scope, setScope] = useState('ADMIN_FULL_ACCESS')

  const [generatedKey, setGeneratedKey] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [copiedKeyId, setCopiedKeyId] = useState<string | null>(null)
  const [revealedKeyIds, setRevealedKeyIds] = useState<Record<string, boolean>>({})

  const defaultKeys = [
    {
      id: 'key-uuid-1',
      name: 'CLI Deployment Key',
      prefix: 'anarva_live_8f3a',
      token: 'anarva_live_8f3a92b1c4e7d5f0a6b8c2d4',
      scope: 'ADMIN_FULL_ACCESS',
      created_at: '2026-08-06',
      status: 'ACTIVE',
    },
    {
      id: 'key-uuid-2',
      name: 'Production SDK Token',
      prefix: 'anarva_live_11c4',
      token: 'anarva_live_11c4e9f2a7b5c3d1e0f8a6b4',
      scope: 'READ_WRITE',
      created_at: '2026-08-07',
      status: 'ACTIVE',
    },
  ]

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(KEYS_STORAGE_KEY)
      if (stored) {
        try {
          setApiKeys(JSON.parse(stored))
        } catch {
          setApiKeys(defaultKeys)
          localStorage.setItem(KEYS_STORAGE_KEY, JSON.stringify(defaultKeys))
        }
      } else {
        setApiKeys(defaultKeys)
        localStorage.setItem(KEYS_STORAGE_KEY, JSON.stringify(defaultKeys))
      }
    }
  }, [])

  const updateKeysState = (newKeys: any[]) => {
    setApiKeys(newKeys)
    if (typeof window !== 'undefined') {
      localStorage.setItem(KEYS_STORAGE_KEY, JSON.stringify(newKeys))
    }
  }

  const handleGenerateKey = (e: React.FormEvent) => {
    e.preventDefault()
    const prefixHex = Math.random().toString(36).substring(2, 6)
    const rawToken = `anarva_live_${prefixHex}${Math.random().toString(36).substring(2, 12)}`

    const newKeyObj = {
      id: `key-uuid-${Date.now()}`,
      name: keyName || 'Production API Token',
      prefix: `anarva_live_${prefixHex}`,
      token: rawToken,
      scope: scope,
      created_at: new Date().toISOString().substring(0, 10),
      status: 'ACTIVE',
    }

    updateKeysState([newKeyObj, ...apiKeys])
    setGeneratedKey(rawToken)
    setKeyName('')
    setShowModal(false)
  }

  const handleRevokeKey = (id: string) => {
    if (confirm('Revoke this API Key? Any application using this key will immediately be denied access.')) {
      const updated = apiKeys.filter((k) => k.id !== id)
      updateKeysState(updated)
    }
  }

  const copyToClipboard = () => {
    if (generatedKey) {
      navigator.clipboard.writeText(generatedKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const handleCopySingleKey = (token: string, id: string) => {
    navigator.clipboard.writeText(token)
    setCopiedKeyId(id)
    setTimeout(() => setCopiedKeyId(null), 2000)
  }

  const toggleRevealKey = (id: string) => {
    setRevealedKeyIds((prev) => ({
      ...prev,
      [id]: !prev[id],
    }))
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">API Keys & Security Audit</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Manage programmatic access tokens, permissions, and security audit logs.</p>
        </div>

        <button
          onClick={() => setShowModal(true)}
          className="w-full sm:w-auto px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25 text-sm"
        >
          + Generate New API Key
        </button>
      </div>

      {/* Active API Keys */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-4 shadow-xl">
        <h2 className="text-lg font-bold text-white">Active API Keys ({apiKeys.length})</h2>

        {apiKeys.length === 0 ? (
          <div className="p-8 text-center text-slate-400 border border-slate-800 rounded-lg">
            No active API keys found. Click "+ Generate New API Key" to provision a token.
          </div>
        ) : (
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-lg overflow-hidden font-mono text-xs">
            {apiKeys.map((k) => {
              const isRevealed = !!revealedKeyIds[k.id]
              return (
                <div key={k.id} className="p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 bg-slate-950">
                  <div>
                    <div className="font-bold text-white text-sm font-sans">{k.name}</div>
                    <div className="text-slate-400 mt-0.5 flex items-center gap-2">
                      <span>Token:</span>
                      <span className="text-emerald-400 font-mono select-all">
                        {isRevealed ? k.token : `${k.prefix}_********************`}
                      </span>
                      <button
                        type="button"
                        onClick={() => toggleRevealKey(k.id)}
                        className="text-xs text-blue-400 hover:underline ml-1"
                      >
                        {isRevealed ? 'Hide' : 'Reveal'}
                      </button>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <span className="px-2.5 py-1 bg-purple-500/10 text-purple-400 border border-purple-500/20 rounded-md font-semibold">
                      {k.scope || 'ADMIN_FULL_ACCESS'}
                    </span>

                    <button
                      onClick={() => handleCopySingleKey(k.token, k.id)}
                      className="px-3 py-1 bg-blue-600/10 hover:bg-blue-600/20 text-blue-400 text-xs font-semibold rounded-lg transition border border-blue-500/20"
                    >
                      {copiedKeyId === k.id ? '✔ Copied!' : 'Copy Key'}
                    </button>

                    <button
                      onClick={() => handleRevokeKey(k.id)}
                      className="px-3 py-1 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-xs font-semibold rounded-lg transition border border-red-500/20"
                    >
                      Revoke Key
                    </button>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>

      {/* Security Audit Feed */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 space-y-4 shadow-xl">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-bold text-white">Immutable Security Audit Log</h2>
          <span className="text-xs text-emerald-400 font-mono flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
            Zero Security Anomalies
          </span>
        </div>

        <div className="space-y-2 text-xs font-mono">
          <div className="p-3 bg-slate-950 border border-slate-800 rounded-lg flex items-center justify-between text-slate-300">
            <div>[AUDIT] Auth Token Pair Generated for user 'admin@anarva.io'</div>
            <div className="text-slate-500">Just now</div>
          </div>
          <div className="p-3 bg-slate-950 border border-slate-800 rounded-lg flex items-center justify-between text-slate-300">
            <div>[AUDIT] Multi-AZ High Availability Cluster Status Check (us-east-1a / us-east-1b)</div>
            <div className="text-slate-500">3 mins ago</div>
          </div>
          <div className="p-3 bg-slate-950 border border-slate-800 rounded-lg flex items-center justify-between text-slate-300">
            <div>[AUDIT] Point-in-time WAL Backup archive verified cleanly</div>
            <div className="text-slate-500">12 mins ago</div>
          </div>
        </div>
      </div>

      {/* Generate API Key Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Generate Programmatic API Key</h2>

            <form onSubmit={handleGenerateKey} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Key Identifier Name</label>
                <input
                  type="text"
                  required
                  value={keyName}
                  onChange={(e) => setKeyName(e.target.value)}
                  placeholder="e.g. Production CI/CD Token"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Permission Scope</label>
                <select
                  value={scope}
                  onChange={(e) => setScope(e.target.value)}
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                >
                  <option value="ADMIN_FULL_ACCESS">ADMIN (Full Cluster Management)</option>
                  <option value="READ_WRITE">READ_WRITE (Query & Insert Data)</option>
                  <option value="READ_ONLY">READ_ONLY (Select Queries Only)</option>
                </select>
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25"
                >
                  Generate Key
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Secret Token Reveal Modal */}
      {generatedKey && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-lg space-y-4">
            <h2 className="text-xl font-bold text-white">API Key Generated Successfully</h2>
            <p className="text-xs text-amber-400">
              ⚠️ Please copy your secret API key now. You can also view or copy it from your API Keys list anytime.
            </p>

            <div className="p-3 bg-slate-950 border border-slate-800 rounded-lg font-mono text-xs text-emerald-400 break-all select-all">
              {generatedKey}
            </div>

            <div className="flex gap-2 pt-2">
              <button
                onClick={copyToClipboard}
                className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg transition"
              >
                {copied ? '✔ Copied Key to Clipboard!' : 'Copy Secret API Key'}
              </button>
              <button
                onClick={() => setGeneratedKey(null)}
                className="px-4 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
              >
                Done
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
