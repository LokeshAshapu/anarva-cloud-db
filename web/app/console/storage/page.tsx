'use client'

import React, { useEffect, useState } from 'react'

const BUCKET_KEY = 'anarva_aos_buckets'

interface ObjectItem {
  id: string
  name: string
  bucket: string
  sizeBytes: number
  contentType: string
  url: string
  isPublic: boolean
  lastModified: string
}

export default function ObjectStoragePage() {
  const [activeTab, setActiveTab] = useState<'BUCKETS' | 'ENTITY_GROUPS'>('BUCKETS')
  const [buckets, setBuckets] = useState<string[]>(['prod-app-assets', 'user-avatars-public', 'backups-archive-private'])
  const [selectedBucket, setSelectedBucket] = useState<string>('prod-app-assets')
  const [newBucketName, setNewBucketName] = useState<string>('')
  const [showBucketModal, setShowBucketModal] = useState<boolean>(false)

  // Objects state
  const [objects, setObjects] = useState<ObjectItem[]>([
    { id: 'obj-1', name: 'logo-brand-trident.png', bucket: 'prod-app-assets', sizeBytes: 245000, contentType: 'image/png', url: '/anarva-trident.png', isPublic: true, lastModified: '2026-08-10 18:30:00' },
    { id: 'obj-2', name: 'user_profile_export.csv', bucket: 'prod-app-assets', sizeBytes: 12400, contentType: 'text/csv', url: '#', isPublic: false, lastModified: '2026-08-10 19:12:00' },
    { id: 'obj-3', name: 'system_log_dump.json', bucket: 'backups-archive-private', sizeBytes: 458000, contentType: 'application/json', url: '#', isPublic: false, lastModified: '2026-08-10 20:00:00' },
  ])

  // Upload modal state
  const [showUploadModal, setShowUploadModal] = useState(false)
  const [uploadFileName, setUploadFileName] = useState('')
  const [isPublicUpload, setIsPublicUpload] = useState(true)

  // Signed URL modal
  const [signedUrlModalObj, setSignedUrlModalObj] = useState<ObjectItem | null>(null)
  const [generatedSignedUrl, setGeneratedSignedUrl] = useState<string | null>(null)
  const [copiedSigned, setCopiedSigned] = useState(false)

  const handleCreateBucket = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newBucketName) return
    const formatted = newBucketName.toLowerCase().replace(/[^a-z0-9-]/g, '-')
    if (!buckets.includes(formatted)) {
      setBuckets([...buckets, formatted])
      setSelectedBucket(formatted)
    }
    setNewBucketName('')
    setShowBucketModal(false)
  }

  const handleUploadObject = (e: React.FormEvent) => {
    e.preventDefault()
    if (!uploadFileName) return
    const newObj: ObjectItem = {
      id: `obj-${Date.now()}`,
      name: uploadFileName,
      bucket: selectedBucket,
      sizeBytes: Math.floor(Math.random() * 500000 + 10000),
      contentType: uploadFileName.endsWith('.png') ? 'image/png' : uploadFileName.endsWith('.csv') ? 'text/csv' : 'application/octet-stream',
      url: '#',
      isPublic: isPublicUpload,
      lastModified: new Date().toISOString().replace('T', ' ').substring(0, 19),
    }
    setObjects([newObj, ...objects])
    setShowUploadModal(false)
    setUploadFileName('')
  }

  const handleGenerateSignedUrl = (obj: ObjectItem) => {
    setSignedUrlModalObj(obj)
    const token = Math.random().toString(36).substring(2, 15)
    setGeneratedSignedUrl(`https://anarva-cloud-db.vercel.app/share/signed-${token}?bucket=${obj.bucket}&key=${encodeURIComponent(obj.name)}&expires=3600`)
  }

  const filteredObjects = objects.filter((o) => o.bucket === selectedBucket)

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Anarva Object Storage (AOS)</h1>
          <p className="text-slate-400 text-xs sm:text-sm mt-1">Scalable S3-compatible object storage for images, videos, audio, documents, and datasets.</p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => setShowBucketModal(true)}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition text-xs shadow-lg shadow-blue-600/20"
          >
            + Create Bucket
          </button>
        </div>
      </div>

      {/* Storage Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Total Active Buckets</div>
          <div className="text-3xl font-extrabold text-white font-mono">{buckets.length}</div>
          <div className="text-xs text-slate-400">Default Region: us-east-1</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Total Stored Objects</div>
          <div className="text-3xl font-extrabold text-blue-400 font-mono">{objects.length} Objects</div>
          <div className="text-xs text-slate-400">Total Size: 2.4 GB</div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-2">
          <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider">AOS Storage Policy</div>
          <div className="text-3xl font-extrabold text-emerald-400 font-mono">S3-COMPATIBLE</div>
          <div className="text-xs text-slate-400">AES-256 Server Encryption</div>
        </div>
      </div>

      {/* Bucket Selector & Actions Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-slate-900 border border-slate-800 rounded-2xl p-4">
        <div className="flex items-center gap-3">
          <span className="text-xs font-bold text-slate-400">Selected Bucket:</span>
          <select
            value={selectedBucket}
            onChange={(e) => setSelectedBucket(e.target.value)}
            className="bg-slate-950 border border-slate-800 text-white rounded-xl px-3 py-1.5 text-xs font-mono focus:outline-none"
          >
            {buckets.map((b) => (
              <option key={b} value={b}>
                {b}
              </option>
            ))}
          </select>
        </div>

        <button
          onClick={() => setShowUploadModal(true)}
          className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-xl transition"
        >
          + Upload Object to {selectedBucket}
        </button>
      </div>

      {/* Objects Table */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 space-y-4">
        <h2 className="text-base font-bold text-white">Objects in bucket: <span className="text-blue-400 font-mono">{selectedBucket}</span></h2>

        {filteredObjects.length === 0 ? (
          <div className="p-8 text-center text-xs text-slate-500 bg-slate-950 rounded-xl border border-slate-800">
            No objects uploaded to this bucket yet. Click "+ Upload Object" above.
          </div>
        ) : (
          <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden text-xs">
            {filteredObjects.map((obj) => (
              <div key={obj.id} className="p-4 bg-slate-950 flex flex-col sm:flex-row sm:items-center justify-between gap-3 font-mono">
                <div>
                  <div className="font-bold text-white font-sans text-sm">{obj.name}</div>
                  <div className="text-slate-400 text-[11px] mt-0.5">
                    {(obj.sizeBytes / 1024).toFixed(1)} KB • {obj.contentType} • Last Modified: {obj.lastModified}
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <span className={`px-2.5 py-0.5 rounded text-[10px] font-bold ${obj.isPublic ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-slate-800 text-slate-400'}`}>
                    {obj.isPublic ? 'Public Read' : 'Private'}
                  </span>
                  <button
                    onClick={() => handleGenerateSignedUrl(obj)}
                    className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-blue-400 rounded-xl text-[11px] font-sans font-semibold transition"
                  >
                    Generate Signed URL
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create Bucket Modal */}
      {showBucketModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 space-y-6">
            <h2 className="text-lg font-bold text-white">Create Anarva Storage Bucket</h2>

            <form onSubmit={handleCreateBucket} className="space-y-4 text-xs">
              <div className="space-y-1">
                <label className="text-slate-400 font-semibold">Bucket Name</label>
                <input
                  type="text"
                  value={newBucketName}
                  onChange={(e) => setNewBucketName(e.target.value)}
                  placeholder="e.g. company-media-assets"
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono focus:outline-none focus:border-blue-500"
                  required
                />
              </div>

              <div className="pt-4 border-t border-slate-800 flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setShowBucketModal(false)}
                  className="px-4 py-2 bg-slate-800 text-slate-300 rounded-xl"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl shadow-lg shadow-blue-600/20"
                >
                  Create Bucket
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Upload Object Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 space-y-6">
            <h2 className="text-lg font-bold text-white">Upload Object to <span className="text-blue-400 font-mono">{selectedBucket}</span></h2>

            <form onSubmit={handleUploadObject} className="space-y-4 text-xs">
              <div className="space-y-1">
                <label className="text-slate-400 font-semibold">Object Key / File Name</label>
                <input
                  type="text"
                  value={uploadFileName}
                  onChange={(e) => setUploadFileName(e.target.value)}
                  placeholder="e.g. documents/reports_2026.pdf"
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-white font-mono focus:outline-none focus:border-blue-500"
                  required
                />
              </div>

              <div className="flex items-center gap-3 pt-2">
                <input
                  type="checkbox"
                  id="publicUploadCheck"
                  checked={isPublicUpload}
                  onChange={(e) => setIsPublicUpload(e.target.checked)}
                  className="rounded border-slate-800 bg-slate-950 text-blue-600 focus:ring-0"
                />
                <label htmlFor="publicUploadCheck" className="text-slate-300 font-medium cursor-pointer">
                  Grant Public Read Access
                </label>
              </div>

              <div className="pt-4 border-t border-slate-800 flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setShowUploadModal(false)}
                  className="px-4 py-2 bg-slate-800 text-slate-300 rounded-xl"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-5 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl shadow-lg shadow-blue-600/20"
                >
                  Upload Object
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Signed URL Modal */}
      {signedUrlModalObj && generatedSignedUrl && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 space-y-4">
            <h3 className="text-base font-bold text-white">Generated Signed Access URL</h3>
            <p className="text-xs text-slate-400">Valid for 60 minutes. Grant temporary access to private object: <strong className="text-white">{signedUrlModalObj.name}</strong></p>

            <div className="p-3 bg-slate-950 border border-slate-800 rounded-xl font-mono text-xs text-blue-300 break-all">
              {generatedSignedUrl}
            </div>

            <div className="flex justify-end gap-3">
              <button
                onClick={() => {
                  navigator.clipboard.writeText(generatedSignedUrl)
                  setCopiedSigned(true)
                  setTimeout(() => setCopiedSigned(false), 2000)
                }}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold rounded-xl"
              >
                {copiedSigned ? 'Copied Signed URL!' : 'Copy Signed URL'}
              </button>
              <button
                onClick={() => setSignedUrlModalObj(null)}
                className="px-4 py-2 bg-slate-800 text-slate-300 text-xs rounded-xl"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
