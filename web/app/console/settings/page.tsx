'use client'

import React from 'react'

export default function SettingsPage() {
  return (
    <div className="space-y-8">
      <div className="border-b border-slate-800 pb-5">
        <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Platform Settings</h1>
        <p className="text-slate-400 text-xs sm:text-sm mt-1">Configure organization parameters, default region preferences, and environment notifications.</p>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4 max-w-2xl">
        <h2 className="text-base font-bold text-white">Organization Configuration</h2>
        <div className="space-y-4 text-xs">
          <div className="space-y-1">
            <label className="text-slate-400 font-semibold">Organization Name</label>
            <input
              type="text"
              defaultValue="Default Organization"
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500"
            />
          </div>
          <div className="space-y-1">
            <label className="text-slate-400 font-semibold">Default Deployment Region</label>
            <select className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500">
              <option value="us-east-1">us-east-1 (N. Virginia)</option>
              <option value="eu-west-1">eu-west-1 (Frankfurt)</option>
              <option value="ap-south-1">ap-south-1 (Mumbai)</option>
            </select>
          </div>
          <button className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition shadow-lg shadow-blue-600/20">
            Save Preferences
          </button>
        </div>
      </div>
    </div>
  )
}
