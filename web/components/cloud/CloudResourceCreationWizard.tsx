'use client'

import React, { useState } from 'react'
import { CloudButton } from './CloudButton'

export interface WizardStep {
  id: string
  title: string
}

export interface CloudResourceCreationWizardProps {
  title: string
  resourceType: 'DATABASE' | 'STORAGE' | 'COMPUTE' | 'NETWORK'
  onComplete: (data: any) => void
  onCancel: () => void
}

export function CloudResourceCreationWizard({
  title,
  resourceType,
  onComplete,
  onCancel,
}: CloudResourceCreationWizardProps) {
  const [currentStepIndex, setCurrentStepIndex] = useState(0)
  const [name, setName] = useState('')
  const [region, setRegion] = useState('ap-south-1')
  const [isProvisioning, setIsProvisioning] = useState(false)

  const steps: WizardStep[] = [
    { id: 'choose', title: 'Choose Resource' },
    { id: 'configure', title: 'Configure Parameters' },
    { id: 'review', title: 'Review & Verify' },
    { id: 'provision', title: 'Provision Resource' },
  ]

  const handleNext = () => {
    if (currentStepIndex < steps.length - 1) {
      setCurrentStepIndex(currentStepIndex + 1)
    } else {
      setIsProvisioning(true)
      setTimeout(() => {
        setIsProvisioning(false)
        onComplete({ name, region, resourceType })
      }, 1500)
    }
  }

  const handleBack = () => {
    if (currentStepIndex > 0) {
      setCurrentStepIndex(currentStepIndex - 1)
    }
  }

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-6 shadow-2xl max-w-xl mx-auto">
      <div className="flex items-center justify-between border-b border-slate-800 pb-3">
        <h3 className="text-base font-bold text-white tracking-tight">{title}</h3>
        <span className="text-xs font-mono text-blue-400">
          Step {currentStepIndex + 1} of {steps.length}
        </span>
      </div>

      {/* Progress Bar */}
      <div className="flex items-center justify-between gap-2 text-xs font-mono">
        {steps.map((s, idx) => (
          <div
            key={s.id}
            className={`flex-1 text-center py-1.5 rounded-lg border transition ${
              idx === currentStepIndex
                ? 'bg-blue-600/10 text-blue-400 border-blue-500/30 font-bold'
                : idx < currentStepIndex
                ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                : 'bg-slate-950 text-slate-500 border-slate-800'
            }`}
          >
            {s.title}
          </div>
        ))}
      </div>

      {/* Step Contents */}
      <div className="space-y-4 text-xs">
        {currentStepIndex === 0 && (
          <div className="space-y-3">
            <label className="block text-slate-300 font-semibold">Resource Specification Type</label>
            <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-1">
              <div className="font-bold text-white">{resourceType} Instance</div>
              <div className="text-slate-400 text-[11px]">Standard Enterprise Cloud Provisioning Spec</div>
            </div>
          </div>
        )}

        {currentStepIndex === 1 && (
          <div className="space-y-4">
            <div className="space-y-1">
              <label className="block text-slate-300 font-semibold">Resource Name</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. prod-service-cluster"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-blue-500"
                required
              />
            </div>
            <div className="space-y-1">
              <label className="block text-slate-300 font-semibold">Target Region</label>
              <select
                value={region}
                onChange={(e) => setRegion(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white focus:outline-none"
              >
                <option value="ap-south-2">Asia Pacific — Hyderabad (ap-south-2)</option>
                <option value="ap-south-1">Asia Pacific — Mumbai (ap-south-1)</option>
                <option value="ap-southeast-1">Asia Pacific — Singapore (ap-southeast-1)</option>
                <option value="us-east-1">US East — N. Virginia (us-east-1)</option>
                <option value="eu-west-1">Europe West — Frankfurt (eu-west-1)</option>
              </select>
            </div>
          </div>
        )}

        {currentStepIndex === 2 && (
          <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2 font-mono">
            <div className="text-slate-400">Resource Type: <strong className="text-white">{resourceType}</strong></div>
            <div className="text-slate-400">Resource Name: <strong className="text-white">{name || 'Default Resource'}</strong></div>
            <div className="text-slate-400">Region: <strong className="text-white">{region}</strong></div>
          </div>
        )}

        {currentStepIndex === 3 && (
          <div className="p-6 text-center space-y-2">
            <span className="h-6 w-6 rounded-full border-2 border-blue-500 border-t-transparent animate-spin inline-block"></span>
            <div className="text-slate-300 font-semibold">Provisioning {name || resourceType}...</div>
          </div>
        )}
      </div>

      {/* Buttons */}
      <div className="pt-4 border-t border-slate-800 flex justify-between">
        <CloudButton variant="outline" size="sm" onClick={onCancel}>
          Cancel
        </CloudButton>
        <div className="flex gap-2">
          {currentStepIndex > 0 && (
            <CloudButton variant="secondary" size="sm" onClick={handleBack}>
              Back
            </CloudButton>
          )}
          <CloudButton variant="primary" size="sm" isLoading={isProvisioning} onClick={handleNext}>
            {currentStepIndex === steps.length - 1 ? 'Provision' : 'Next Step'}
          </CloudButton>
        </div>
      </div>
    </div>
  )
}
