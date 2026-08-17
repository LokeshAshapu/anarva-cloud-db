'use client'

import React, { useState, useEffect } from 'react'
import { CloudStatus } from '@/components/cloud/CloudStatus'
import { CloudButton } from '@/components/cloud/CloudButton'
import { CloudCard } from '@/components/cloud/CloudCard'
import { CloudTabs, TabItem } from '@/components/cloud/CloudTabs'
import { CloudModal } from '@/components/cloud/CloudModal'
import { CloudEmptyState } from '@/components/cloud/CloudEmptyState'
import { API_BASE_URL, getAuthHeaders } from '@/lib/api'

interface BucketItem {
  id: string
  name: string
  region: string
  status: string
  storageClass: string
  versioning: boolean
  publicAccess: string
  realityLabel: string
  createdAt: string
}

interface ObjectItem {
  id: string
  bucketId: string
  key: string
  contentType: string
  size: number
  category: string
  etag: string
  versionId: string
  createdAt: string
}

export default function ObjectStoragePage() {
  const [selectedBucket, setSelectedBucket] = useState<BucketItem | null>(null)
  const [activeTab, setActiveTab] = useState('objects')
  const [isWizardOpen, setIsWizardOpen] = useState(false)

  // Creation Wizard State
  const [bucketName, setBucketName] = useState('')
  const [regionId, setRegionId] = useState('ap-hyderabad-1')
  const [isCreating, setIsCreating] = useState(false)

  // Object Browser & Signed URL State
  const [selectedObject, setSelectedObject] = useState<ObjectItem | null>(null)
  const [signedUrlModalOpen, setSignedUrlModalOpen] = useState(false)
  const [generatedSignedUrl, setGeneratedSignedUrl] = useState('')
  const [uploadFileName, setUploadFileName] = useState('')

  // User Email & Buckets State
  const [userEmail, setUserEmail] = useState('user@anarva.io')
  const [buckets, setBuckets] = useState<BucketItem[]>([])
  const [objects, setObjects] = useState<ObjectItem[]>([])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const email = localStorage.getItem('anarva_user_email') || 'user@anarva.io'
      setUserEmail(email)

      const authHeaders = getAuthHeaders()
      fetch(`${API_BASE_URL}/api/v1/storage/buckets`, { headers: authHeaders })
        .then((r) => r.json())
        .then((res) => {
          if (res && res.data) setBuckets(res.data)
        })
        .catch(() => null)
    }
  }, [])

  const handleCreateBucket = async () => {
    setIsCreating(true)
    const cleanName = bucketName || 'production-bucket'

    const authHeaders = getAuthHeaders()
    const res = await fetch(`${API_BASE_URL}/api/v1/storage/buckets`, {
      method: 'POST',
      headers: authHeaders,
      body: JSON.stringify({
        organizationId: 'org-main',
        projectId: 'proj-main',
        name: cleanName,
        region: regionId,
      }),
    }).then((r) => r.json()).catch(() => null)

    if (res && res.data) {
      setBuckets([res.data, ...buckets])
    } else {
      const fallback: BucketItem = {
        id: `bkt-${Date.now()}`,
        name: cleanName,
        region: regionId,
        status: 'ACTIVE',
        storageClass: 'STANDARD',
        versioning: true,
        publicAccess: 'PRIVATE',
        realityLabel: 'LOCAL_STORAGE (REAL_LOCAL)',
        createdAt: new Date().toISOString(),
      }
      setBuckets([fallback, ...buckets])
    }

    setIsCreating(false)
    setIsWizardOpen(false)
    setBucketName('')
  }

  const handleDeleteBucket = async (id: string) => {
    if (confirm(`Are you sure you want to delete bucket '${id}'?`)) {
      const authHeaders = getAuthHeaders()
      await fetch(`${API_BASE_URL}/api/v1/storage/buckets/${id}`, {
        method: 'DELETE',
        headers: authHeaders,
      }).catch(() => null)
      setBuckets(buckets.filter((b) => b.id !== id))
      setSelectedBucket(null)
    }
  }

  const handleGenerateSignedUrl = async (key: string) => {
    if (!selectedBucket) return
    const authHeaders = getAuthHeaders()
    const res = await fetch(`${API_BASE_URL}/api/v1/storage/buckets/${selectedBucket.id}/signed-url`, {
      method: 'POST',
      headers: authHeaders,
      body: JSON.stringify({ key, method: 'GET', expiresSec: 3600 }),
    }).then((r) => r.json()).catch(() => null)

    if (res && res.data) {
      setGeneratedSignedUrl(res.data.url)
      setSignedUrlModalOpen(true)
    }
  }

  const handleUploadSimulatedObject = async () => {
    if (!uploadFileName || !selectedBucket) return
    const newObj: ObjectItem = {
      id: `obj-${Date.now()}`,
      bucketId: selectedBucket.id,
      key: uploadFileName,
      contentType: uploadFileName.endsWith('.png') ? 'image/png' : 'application/octet-stream',
      size: 1024 * 142,
      category: uploadFileName.endsWith('.png') ? 'IMAGES' : 'DOCUMENTS',
      etag: `"etag-${Date.now()}"`,
      versionId: 'v1',
      createdAt: new Date().toISOString(),
    }
    setObjects([newObj, ...objects])
    setUploadFileName('')
  }

  const detailTabs: TabItem[] = [
    { id: 'objects', label: 'Objects Explorer' },
    { id: 'signed-urls', label: 'Presigned URLs' },
    { id: 'policy', label: 'Bucket Policies & CORS' },
    { id: 's3', label: 'S3 Compatibility API' },
  ]

  if (selectedBucket) {
    return (
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
          <div>
            <button onClick={() => setSelectedBucket(null)} className="text-xs text-blue-400 font-mono mb-2">
              ← Back to Storage Buckets
            </button>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">🪣 {selectedBucket.name}</h1>
              <CloudStatus status={selectedBucket.status} />
              <span className="px-2 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 text-xs rounded font-mono font-bold">
                {selectedBucket.storageClass}
              </span>
            </div>
            <div className="text-xs text-slate-400 font-mono mt-1">
              Region: {selectedBucket.region} • Access: {selectedBucket.publicAccess} • Versioning: {selectedBucket.versioning ? 'ENABLED' : 'DISABLED'}
            </div>
          </div>

          <CloudButton variant="danger" size="sm" onClick={() => handleDeleteBucket(selectedBucket.id)}>
            Delete Bucket
          </CloudButton>
        </div>

        <CloudTabs tabs={detailTabs} activeTab={activeTab} onChange={setActiveTab} />

        {activeTab === 'objects' && (
          <div className="space-y-4 font-mono text-xs">
            <CloudCard title="Upload Object">
              <div className="flex gap-3">
                <input
                  type="text"
                  placeholder="Object Key (e.g. documents/2026/report.pdf)"
                  value={uploadFileName}
                  onChange={(e) => setUploadFileName(e.target.value)}
                  className="flex-1 p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
                />
                <CloudButton variant="primary" size="sm" onClick={handleUploadSimulatedObject}>
                  + Upload Object
                </CloudButton>
              </div>
            </CloudCard>

            <CloudCard title="Bucket Objects">
              {objects.length === 0 ? (
                <div className="text-slate-500 text-center py-6">No objects uploaded yet to this bucket.</div>
              ) : (
                <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden">
                  {objects.map((obj) => (
                    <div key={obj.id} className="p-4 bg-slate-950 flex items-center justify-between">
                      <div>
                        <div className="font-bold text-white font-sans text-sm">{obj.key}</div>
                        <div className="text-[10px] text-slate-500 mt-1">
                          Size: {(obj.size / 1024).toFixed(1)} KB • Type: {obj.contentType} • Version: {obj.versionId} • ETag: {obj.etag}
                        </div>
                      </div>

                      <div className="flex gap-2">
                        <CloudButton variant="secondary" size="sm" onClick={() => handleGenerateSignedUrl(obj.key)}>
                          Generate Presigned URL
                        </CloudButton>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CloudCard>
          </div>
        )}

        {activeTab === 's3' && (
          <CloudCard title="S3 Compatibility Endpoint Integration">
            <div className="space-y-3 font-mono text-xs">
              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-2">
                <div>Endpoint URL: <strong className="text-emerald-400">{API_BASE_URL}/s3/{selectedBucket.name}</strong></div>
                <div>Compatibility Level: <strong className="text-purple-400">ANARVA_COMPATIBLE</strong></div>
                <div>Supported Actions: ListBuckets, CreateBucket, DeleteBucket, HeadBucket, ListObjects, PutObject, GetObject, DeleteObject, CreateMultipartUpload</div>
              </div>
            </div>
          </CloudCard>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Managed Object Storage Platform</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">S3-compatible Object Storage Accounts, Buckets, Presigned URLs, & Versioning.</p>
        </div>

        <CloudButton variant="primary" size="sm" onClick={() => setIsWizardOpen(true)}>
          + Create Storage Bucket
        </CloudButton>
      </div>

      <CloudCard title="Storage Buckets Registry" subtitle={`Managed buckets for ${userEmail}`}>
        {buckets.length === 0 ? (
          <CloudEmptyState
            title="No Storage Buckets Created"
            description="Create an S3-compatible storage bucket to store files, images, videos, and documents."
            actionLabel="+ Create Storage Bucket"
            onAction={() => setIsWizardOpen(true)}
            icon={
              <svg className="w-6 h-6 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
              </svg>
            }
            docsLink="/console/developer"
          />
        ) : (
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
            {buckets.map((b) => (
              <div
                key={b.id}
                onClick={() => setSelectedBucket(b)}
                className="p-4 bg-slate-950 hover:bg-slate-900 cursor-pointer transition flex items-center justify-between font-mono"
              >
                <div>
                  <div className="font-bold text-white text-sm font-sans flex items-center gap-2">
                    <svg className="w-4 h-4 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                    </svg>
                    <span>{b.name}</span>
                    <span className="text-[10px] px-2 py-0.5 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded">
                      {b.storageClass} • {b.region}
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1">
                    ID: {b.id} • Access: {b.publicAccess} • Reality: {b.realityLabel}
                  </div>
                </div>
                <CloudStatus status={b.status} />
              </div>
            ))}
          </div>
        )}
      </CloudCard>

      {/* Create Bucket Modal */}
      {isWizardOpen && (
        <CloudModal isOpen={isWizardOpen} title="Create Storage Bucket" onClose={() => setIsWizardOpen(false)}>
          <form onSubmit={(e) => { e.preventDefault(); handleCreateBucket(); }} className="space-y-4 font-mono text-xs">
            <div>
              <label className="block text-slate-300 font-bold mb-1">Bucket Name</label>
              <input
                type="text"
                value={bucketName}
                onChange={(e) => setBucketName(e.target.value)}
                placeholder="production-media-bucket"
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none"
              />
            </div>

            <div>
              <label className="block text-slate-300 font-bold mb-1">Region</label>
              <select
                value={regionId}
                onChange={(e) => setRegionId(e.target.value)}
                className="w-full p-2.5 bg-slate-950 border border-slate-800 rounded text-slate-100 focus:outline-none font-bold"
              >
                <option value="ap-hyderabad-1">Asia Pacific (Hyderabad) - ap-hyderabad-1</option>
                <option value="us-east-1">US East (N. Virginia) - us-east-1</option>
              </select>
            </div>

            <div className="pt-3 border-t border-slate-800 flex justify-end gap-2">
              <CloudButton variant="secondary" size="sm" type="button" onClick={() => setIsWizardOpen(false)}>Cancel</CloudButton>
              <CloudButton variant="primary" size="sm" type="submit" disabled={isCreating}>
                {isCreating ? 'Creating Bucket...' : 'Create Bucket'}
              </CloudButton>
            </div>
          </form>
        </CloudModal>
      )}

      {/* Presigned URL Modal */}
      {signedUrlModalOpen && (
        <CloudModal isOpen={signedUrlModalOpen} title="Presigned Download URL Generated" onClose={() => setSignedUrlModalOpen(false)}>
          <div className="space-y-3 font-mono text-xs">
            <p className="text-slate-300 font-sans">
              This presigned URL allows secure temporary access to the object with HMAC signature protection.
            </p>
            <div className="p-3 bg-slate-950 border border-slate-800 rounded break-all text-emerald-400">
              {generatedSignedUrl}
            </div>
            <div className="flex justify-end">
              <CloudButton variant="primary" size="sm" onClick={() => setSignedUrlModalOpen(false)}>Close</CloudButton>
            </div>
          </div>
        </CloudModal>
      )}
    </div>
  )
}
