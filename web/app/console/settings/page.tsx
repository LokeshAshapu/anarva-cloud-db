'use client'

import React, { useState, useEffect } from 'react'

const SETTINGS_KEY = 'anarva_platform_settings'

export default function SettingsPage() {
  const [orgName, setOrgName] = useState('Default Organization')
  const [defaultRegion, setDefaultRegion] = useState('ap-south-1')
  const [notificationEmail, setNotificationEmail] = useState('lokesh@anarva.io')
  const [savedSuccess, setSavedSuccess] = useState(false)

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(SETTINGS_KEY)
      if (stored) {
        try {
          const parsed = JSON.parse(stored)
          if (parsed.orgName) setOrgName(parsed.orgName)
          if (parsed.defaultRegion) setDefaultRegion(parsed.defaultRegion)
          if (parsed.notificationEmail) setNotificationEmail(parsed.notificationEmail)
        } catch (e) {
          console.error('Failed to parse settings', e)
        }
      }
    }
  }, [])

  const handleSaveSettings = (e: React.FormEvent) => {
    e.preventDefault()
    const settingsObj = {
      orgName,
      defaultRegion,
      notificationEmail,
      updatedAt: new Date().toISOString(),
    }

    if (typeof window !== 'undefined') {
      localStorage.setItem(SETTINGS_KEY, JSON.stringify(settingsObj))
    }

    setSavedSuccess(true)
    setTimeout(() => {
      setSavedSuccess(false)
    }, 3000)
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="border-b border-slate-800 pb-5">
        <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Platform Settings</h1>
        <p className="text-slate-400 text-xs sm:text-sm mt-1">Configure organization parameters, default region preferences, and environment notifications.</p>
      </div>

      {/* Save Success Notification Alert */}
      {savedSuccess && (
        <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold rounded-2xl flex items-center justify-between animate-in fade-in">
          <div className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full bg-emerald-400 animate-ping"></span>
            <span>Platform preferences saved successfully!</span>
          </div>
          <span className="font-mono text-[10px]">VERIFIED</span>
        </div>
      )}

      {/* Organization Settings Form */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-6 max-w-2xl shadow-xl">
        <h2 className="text-base font-bold text-white">Organization Configuration</h2>

        <form onSubmit={handleSaveSettings} className="space-y-4 text-xs">
          <div className="space-y-1">
            <label className="text-slate-400 font-semibold">Organization Name</label>
            <input
              type="text"
              value={orgName}
              onChange={(e) => setOrgName(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500 font-medium"
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-slate-400 font-semibold">Default Deployment Region</label>
            <select
              value={defaultRegion}
              onChange={(e) => setDefaultRegion(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500 font-medium cursor-pointer"
            >
              <option value="us-east-1">us-east-1 (N. Virginia)</option>
              <option value="us-west-2">us-west-2 (Oregon)</option>
              <option value="eu-west-1">eu-west-1 (Frankfurt)</option>
              <option value="ap-south-1">ap-south-1 (Mumbai)</option>
              <option value="ap-southeast-1">ap-southeast-1 (Singapore)</option>
            </select>
          </div>

          <div className="space-y-1">
            <label className="text-slate-400 font-semibold">System Notification Recipient Email</label>
            <input
              type="email"
              value={notificationEmail}
              onChange={(e) => setNotificationEmail(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500 font-mono"
              required
            />
          </div>

          <div className="pt-4 border-t border-slate-800 flex justify-end">
            <button
              type="submit"
              className="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition shadow-lg shadow-blue-600/20"
            >
              Save Preferences
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
