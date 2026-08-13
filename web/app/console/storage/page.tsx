'use client'

import React, { useState, useEffect } from 'react'
import { CloudResource, ResourceStatus, RegionId } from '@/types/resource'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { generateARNV } from '@/lib/arnv'

interface BucketItem {
  id: string
  resourceId: string
  name: string
  regionId: RegionId
  status: ResourceStatus
  storageClass: 'STANDARD' | 'INFREQUENT_ACCESS' | 'ARCHIVE'
  versioningEnabled: boolean
  publicAccessBlocked: boolean
  objectCount: number
  sizeBytes: number
  createdAt: string
}

interface ObjectItem {
  id: string
  bucketId: string
  objectKey: string
  contentType: string
  sizeBytes: number
  etag: string
  versionId: string
  checksum: string
  createdAt: string
}

export default function StoragePage() {
  const [selectedBucket, setSelectedBucket] = useState<BucketItem | null>(null)
  const [activeTab, setActiveTab] = useState('objects')
  const [isWizardOpen, setIsWizardOpen] = useState(false)
  const [wizardStep, setWizardStep] = useState(1)

  // Creation Wizard State
  const [bucketName, setBucketName] = useState('')
  const [regionId, setRegionId] = useState<RegionId>('ap-hyderabad-1')
  const [storageClass, setStorageClass] = useState<'STANDARD' | 'INFREQUENT_ACCESS' | 'ARCHIVE'>('STANDARD')
  const [versioningEnabled, setVersioningEnabled] = useState(true)
  const [publicAccessBlocked, setPublicAccessBlocked] = useState(true)
  const [isCreating, setIsCreating] = useState(false)

  // Object Browser & Signed URL State
  const [currentPrefix, setCurrentPrefix] = useState('')
  const [selectedObject, setSelectedObject] = useState<ObjectItem | null>(null)
  const [signedUrlModalOpen, setSignedUrlModalOpen] = useState(false)
  const [generatedSignedUrl, setGeneratedSignedUrl] = useState('')
  const [signedUrlExpiry, setSignedUrlExpiry] = useState('3600')
  const [uploadFileName, setUploadFileName] = useState('')
  const [uploadFileType, setUploadFileType] = useState('image/png')
  const [isUploading, setIsUploading] = useState(false)

  // User Email & Buckets State
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [buckets, setBuckets] = useState<BucketItem[]>([])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)

      const bucketKey = `anarva_user_storage_${email}`
      const stored = localStorage.getItem(bucketKey)

      if (stored) {
        try {
          setBuckets(JSON.parse(stored))
        } catch (e) {
          setBuckets([])
        }
      } else {
        setBuckets([])
      }
    }
  }, [])

  const saveUserBuckets = (updated: BucketItem[]) => {
    setBuckets(updated)
    if (typeof window !== 'undefined') {
      localStorage.setItem(`anarva_user_storage_${userEmail}`, JSON.stringify(updated))
    }
  }

  // Objects List
  const [objects, setObjects] = useState<ObjectItem[]>([
    {
      id: 'obj-101',
      bucketId: 'res-s3-assets-1',
      objectKey: 'avatars/lokesh/profile.png',
      contentType: 'image/png',
      sizeBytes: 512000,
      etag: '"a3b2c1d4e5"',
      versionId: 'v1.0',
      checksum: 'sha256-e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
      createdAt: new Date().toISOString(),
    },
    {
      id: 'obj-102',
      bucketId: 'res-s3-assets-1',
      objectKey: 'documents/contracts/2026/service-agreement.pdf',
      contentType: 'application/pdf',
      sizeBytes: 2400000,
      etag: '"f9e8d7c6b5"',
      versionId: 'v1.0',
      checksum: 'sha256-1c88846c8793b8e788c2ad16a695b28d7085c9eb92f44005b4b1a45a1c726ee8',
      createdAt: new Date().toISOString(),
    },
  ])

  const handleCreateBucket = () => {
    setIsCreating(true)
    setTimeout(() => {
      const cleanName = bucketName.toLowerCase().replace(/[^a-z0-9-]/g, '-') || 'new-bucket'
      const newBucket: BucketItem = {
        id: `res-s3-${Date.now()}`,
        resourceId: generateARNV('STORAGE_BUCKET', regionId, 'proj-default', cleanName),
        name: cleanName,
        regionId,
        status: 'AVAILABLE',
        storageClass,
        versioningEnabled,
        publicAccessBlocked,
        objectCount: 0,
        sizeBytes: 0,
        createdAt: new Date().toISOString(),
      }

      const updated = [newBucket, ...buckets]
      saveUserBuckets(updated)

      // Record activity event
      if (typeof window !== 'undefined') {
        const actKey = `anarva_user_activities_${userEmail}`
        const existingActs = JSON.parse(localStorage.getItem(actKey) || '[]')
        const newAct = {
          id: `act-${Date.now()}`,
          action: 'RESOURCE_CREATED',
          resource: cleanName,
          actor: userEmail,
          time: 'Just now',
        }
        localStorage.setItem(actKey, JSON.stringify([newAct, ...existingActs]))
      }

      setIsCreating(false)
      setIsWizardOpen(false)
      setWizardStep(1)
    }, 1000)
  }

  const handleUploadObject = () => {
    if (!uploadFileName || !selectedBucket) return
    setIsUploading(true)
    setTimeout(() => {
      const keyPrefix = currentPrefix ? `${currentPrefix}/` : ''
      const newObj: ObjectItem = {
        id: `obj-${Date.now()}`,
        bucketId: selectedBucket.id,
        objectKey: `${keyPrefix}${uploadFileName}`,
        contentType: uploadFileType,
        sizeBytes: 1024 * 350,
        etag: `"${Math.random().toString(36).substring(7)}"`,
        versionId: 'v1.0',
        checksum: `sha256-${Math.random().toString(36).substring(2)}`,
        createdAt: new Date().toISOString(),
      }

      setObjects([newObj, ...objects])
      setIsUploading(false)
      setUploadFileName('')
    }, 1000)
  }

  const handleGenerateSignedUrl = (obj: ObjectItem) => {
    setSelectedObject(obj)
    const token = `sig-${Date.now()}-${obj.id}`
    const exp = Date.now() + Number(signedUrlExpiry) * 1000
    const signed = `https://aos.anarva.cloud/${selectedBucket?.name}/${obj.objectKey}?token=${token}&expires=${exp}`
    setGeneratedSignedUrl(signed)
    setSignedUrlModalOpen(true)
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const detailTabs: TabItem[] = [
    { id: 'objects', label: 'Object Browser' },
    { id: 'overview', label: 'Overview' },
    { id: 'metrics', label: 'Metrics' },
    { id: 'permissions', label: 'Permissions & CORS' },
    { id: 'lifecycle', label: 'Lifecycle Rules' },
    { id: 'versions', label: 'Versioning' },
    { id: 'person', label: 'Person Entity View' },
    { id: 'settings', label: 'Settings' },
  ]

  const handleDeleteBucket = (bucketId: string, name: string) => {
    if (confirm(`Are you sure you want to delete storage bucket '${name}'?`)) {
      const updated = buckets.filter((b) => b.id !== bucketId)
      saveUserBuckets(updated)

      // Record activity event
      if (typeof window !== 'undefined') {
        const actKey = `anarva_user_activities_${userEmail}`
        const existingActs = JSON.parse(localStorage.getItem(actKey) || '[]')
        const newAct = {
          id: `act-${Date.now()}`,
          action: 'RESOURCE_DELETED',
          resource: name,
          actor: userEmail,
          time: 'Just now',
        }
        localStorage.setItem(actKey, JSON.stringify([newAct, ...existingActs]))
      }

      setSelectedBucket(null)
    }
  }

  // BUCKET DETAIL VIEW
  if (selectedBucket) {
    const bucketObjects = objects.filter((o) => o.bucketId === selectedBucket.id)

    return (
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
          <div className="space-y-1">
            <button
              onClick={() => setSelectedBucket(null)}
              className="text-xs text-blue-400 hover:underline font-mono flex items-center gap-1 mb-2"
            >
              ← Back to AOS Bucket Registry
            </button>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">{selectedBucket.name}</h1>
              <CloudStatus status={selectedBucket.status} />
            </div>
            <div className="text-xs text-slate-400 font-mono flex items-center gap-2">
              <span>{selectedBucket.storageClass} Storage</span>
              <span>•</span>
              <span className="text-emerald-400 font-bold bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
                {selectedBucket.resourceId}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <CloudButton variant="outline" size="sm" onClick={() => setActiveTab('permissions')}>
              Bucket Policy
            </CloudButton>
            <CloudButton variant="danger" size="sm" onClick={() => handleDeleteBucket(selectedBucket.id, selectedBucket.name)}>
              Delete Bucket
            </CloudButton>
          </div>
        </div>

        {/* Tabs */}
        <CloudTabs tabs={detailTabs} activeTab={activeTab} onChange={setActiveTab} />

        {/* Tab Content */}
        <div className="space-y-6">
          {activeTab === 'objects' && (
            <div className="space-y-6">
              {/* File Upload Section */}
              <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-4">
                <h3 className="text-sm font-bold text-white">Upload New Object</h3>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                  <input
                    type="text"
                    value={uploadFileName}
                    onChange={(e) => setUploadFileName(e.target.value)}
                    placeholder="Object filename e.g. avatar.png"
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white text-xs focus:outline-none"
                  />
                  <select
                    value={uploadFileType}
                    onChange={(e) => setUploadFileType(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white text-xs focus:outline-none"
                  >
                    <option value="image/png">image/png</option>
                    <option value="image/jpeg">image/jpeg</option>
                    <option value="application/pdf">application/pdf</option>
                    <option value="video/mp4">video/mp4</option>
                    <option value="text/plain">text/plain</option>
                  </select>
                  <CloudButton variant="primary" size="sm" isLoading={isUploading} onClick={handleUploadObject}>
                    Upload to AOS
                  </CloudButton>
                </div>
              </div>

              {/* Objects Table */}
              <div className="border border-slate-800 rounded-2xl overflow-hidden shadow-xl bg-slate-900/60">
                <div className="p-4 bg-slate-950 border-b border-slate-800 flex items-center justify-between font-mono text-xs">
                  <span className="text-slate-400">Prefix: <strong className="text-white">s3://{selectedBucket.name}/{currentPrefix}</strong></span>
                  <span className="text-slate-500">{bucketObjects.length} Objects Total</span>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-left font-sans text-xs divide-y divide-slate-800">
                    <thead className="bg-slate-950 text-slate-400 font-bold uppercase text-[10px] tracking-wider">
                      <tr>
                        <th className="p-4">Object Key</th>
                        <th className="p-4">Content Type</th>
                        <th className="p-4">Size</th>
                        <th className="p-4">ETag Checksum</th>
                        <th className="p-4 text-right">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/80 font-mono">
                      {bucketObjects.map((obj) => (
                        <tr key={obj.id} className="hover:bg-slate-800/40 transition">
                          <td className="p-4 font-bold text-white">{obj.objectKey}</td>
                          <td className="p-4 text-blue-400">{obj.contentType}</td>
                          <td className="p-4 text-slate-300">{formatBytes(obj.sizeBytes)}</td>
                          <td className="p-4 text-slate-500 text-[10px]">{obj.etag}</td>
                          <td className="p-4 text-right">
                            <div className="flex items-center justify-end gap-2">
                              <button
                                onClick={() => handleGenerateSignedUrl(obj)}
                                className="px-2 py-1 bg-blue-600/10 hover:bg-blue-600/20 text-blue-400 border border-blue-500/20 rounded text-[11px] font-semibold transition"
                              >
                                Signed URL
                              </button>
                              <button
                                onClick={() => setObjects(objects.filter((o) => o.id !== obj.id))}
                                className="px-2 py-1 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 rounded text-[11px] font-semibold transition"
                              >
                                Delete
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'person' && (
            <CloudCard title="Person Entity Storage Mapping Layer">
              <div className="space-y-4 text-xs font-mono">
                <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                  <div className="font-bold text-white">Entity Scope: Person (Lokesh Ashapu)</div>
                  <div className="text-slate-400">Mapped Bucket: <strong className="text-blue-400">{selectedBucket.name}</strong></div>
                  <div className="text-slate-400">Object Key Namespace: <strong className="text-emerald-400">avatars/user-001/</strong></div>
                </div>
              </div>
            </CloudCard>
          )}

          {activeTab !== 'objects' && activeTab !== 'person' && (
            <div className="p-8 text-center bg-slate-900 border border-slate-800 rounded-2xl text-slate-400 text-xs">
              Module controls active for bucket {selectedBucket.name}.
            </div>
          )}
        </div>

        {/* Signed URL Modal */}
        {signedUrlModalOpen && selectedObject && (
          <CloudModal
            isOpen={signedUrlModalOpen}
            onClose={() => setSignedUrlModalOpen(false)}
            title="Generate Signed URL"
            subtitle={`Temporary download URL for ${selectedObject.objectKey}`}
          >
            <div className="space-y-4 text-xs font-mono">
              <div className="space-y-1">
                <label className="text-slate-400 font-sans">Generated Temporary Signed URL:</label>
                <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl text-emerald-400 break-all select-all font-bold">
                  {generatedSignedUrl}
                </div>
              </div>
              <div className="flex justify-end">
                <CloudButton variant="primary" size="sm" onClick={() => setSignedUrlModalOpen(false)}>
                  Close
                </CloudButton>
              </div>
            </div>
          </CloudModal>
        )}
      </div>
    )
  }

  // MULTI-STEP BUCKET CREATION WIZARD
  if (isWizardOpen) {
    return (
      <div className="max-w-xl mx-auto py-8 space-y-6">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-6 shadow-2xl">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <h2 className="text-base font-bold text-white">Create AOS Storage Bucket</h2>
            <span className="text-xs font-mono text-blue-400">Step {wizardStep} of 6</span>
          </div>

          {wizardStep === 1 && (
            <div className="space-y-2 text-xs">
              <label className="block text-slate-300 font-semibold">Bucket Name</label>
              <input
                type="text"
                value={bucketName}
                onChange={(e) => setBucketName(e.target.value)}
                placeholder="e.g. user-media-assets"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none"
              />
              <p className="text-[11px] text-slate-500">Bucket names must be lowercase letters, numbers, and hyphens.</p>
            </div>
          )}

          {wizardStep === 2 && (
            <div className="space-y-2 text-xs">
              <label className="block text-slate-300 font-semibold">Deployment Region</label>
              <select
                value={regionId}
                onChange={(e) => setRegionId(e.target.value as RegionId)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none"
              >
                <option value="ap-hyderabad-1">Asia Pacific — Hyderabad (ap-hyderabad-1)</option>
                <option value="ap-mumbai-1">Asia Pacific — Mumbai (ap-mumbai-1)</option>
                <option value="ap-singapore-1">Asia Pacific — Singapore (ap-singapore-1)</option>
                <option value="us-east-1">US East — N. Virginia (us-east-1)</option>
                <option value="eu-west-1">Europe West — Frankfurt (eu-west-1)</option>
              </select>
            </div>
          )}

          {wizardStep === 3 && (
            <div className="space-y-2 text-xs">
              <label className="block text-slate-300 font-semibold">Default Storage Class</label>
              <select
                value={storageClass}
                onChange={(e) => setStorageClass(e.target.value as any)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 text-white focus:outline-none"
              >
                <option value="STANDARD">STANDARD (High Throughput & Low Latency)</option>
                <option value="INFREQUENT_ACCESS">INFREQUENT_ACCESS (Lower Cost Storage)</option>
                <option value="ARCHIVE">ARCHIVE (Long-Term Vault)</option>
              </select>
            </div>
          )}

          {wizardStep >= 4 && wizardStep < 6 && (
            <div className="space-y-3 text-xs font-mono">
              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                <div className="text-slate-400">Versioning: <strong className="text-emerald-400">ENABLED</strong></div>
                <div className="text-slate-400">Public Access Block: <strong className="text-blue-400">BLOCKED (Private Default)</strong></div>
              </div>
            </div>
          )}

          {wizardStep === 6 && (
            <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2 text-xs font-mono">
              <div>Bucket Name: <strong className="text-white">{bucketName || 'new-bucket'}</strong></div>
              <div>Region: <strong className="text-emerald-400">{regionId}</strong></div>
              <div>Storage Class: <strong className="text-blue-400">{storageClass}</strong></div>
            </div>
          )}

          {/* Wizard Controls */}
          <div className="pt-4 border-t border-slate-800 flex justify-between">
            <CloudButton variant="outline" size="sm" onClick={() => setIsWizardOpen(false)}>
              Cancel
            </CloudButton>
            <div className="flex gap-2">
              {wizardStep > 1 && (
                <CloudButton variant="secondary" size="sm" onClick={() => setWizardStep(wizardStep - 1)}>
                  Back
                </CloudButton>
              )}
              {wizardStep < 6 ? (
                <CloudButton variant="primary" size="sm" onClick={() => setWizardStep(wizardStep + 1)}>
                  Next Step
                </CloudButton>
              ) : (
                <CloudButton variant="primary" size="sm" isLoading={isCreating} onClick={handleCreateBucket}>
                  Provision Bucket
                </CloudButton>
              )}
            </div>
          </div>
        </div>
      </div>
    )
  }

  // BUCKET LIST VIEW
  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Anarva Object Storage (AOS)</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">
            Scalable S3-compatible object storage buckets with signed URL architecture and multi-region replication specs.
          </p>
        </div>

        <CloudButton variant="primary" size="sm" onClick={() => setIsWizardOpen(true)}>
          + Create Storage Bucket
        </CloudButton>
      </div>

      {/* Bucket List Cards / Enterprise Empty State */}
      {buckets.length === 0 ? (
        <CloudEmptyState
          title="No Storage Buckets Provisioned"
          description="You currently have 0 S3-compatible storage buckets. Provision a storage bucket to upload assets, media, or database backups."
          actionLabel="+ Create Storage Bucket"
          onAction={() => setIsWizardOpen(true)}
          icon="📦"
          docsLink="/console/developer"
        />
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {buckets.map((b) => (
            <div
              key={b.id}
              onClick={() => setSelectedBucket(b)}
              className="bg-slate-900 border border-slate-800 hover:border-slate-700 rounded-2xl p-5 cursor-pointer transition shadow-xl space-y-4"
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="space-y-1">
                  <div className="flex items-center gap-3">
                    <span className="font-bold text-white text-base">{b.name}</span>
                    <CloudStatus status={b.status} />
                  </div>
                  <div className="text-xs text-slate-400 font-mono">
                    {b.regionId} • {b.storageClass} • {b.objectCount} Objects ({formatBytes(b.sizeBytes)})
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <button className="px-3 py-1.5 bg-blue-600/10 text-blue-400 border border-blue-500/20 rounded-xl text-xs font-semibold hover:bg-blue-600/20 transition">
                    Browse Bucket
                  </button>
                </div>
              </div>

              <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl flex items-center justify-between font-mono text-xs text-slate-400">
                <span className="truncate">Resource ID: {b.resourceId}</span>
                <span className="text-emerald-400 font-bold text-[11px] shrink-0">Versioning Active</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
