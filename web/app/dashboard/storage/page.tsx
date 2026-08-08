'use client'

import React, { useState, useEffect } from 'react'

const STORAGE_KEY = 'anarva_unstructured_blobs'

export default function UnstructuredStoragePage() {
  const [buckets, setBuckets] = useState<string[]>(['media-assets', 'user-uploads', 'audio-vault'])
  const [activeBucket, setActiveBucket] = useState<string>('media-assets')

  const [files, setFiles] = useState<any[]>([])
  const [newBucketName, setNewBucketName] = useState('')
  const [showBucketModal, setShowBucketModal] = useState(false)

  const [uploadName, setUploadName] = useState('')
  const [fileType, setFileType] = useState<'IMAGE' | 'VIDEO' | 'AUDIO' | 'DOCUMENT'>('IMAGE')
  const [fileUrl, setFileUrl] = useState('')
  const [showUploadModal, setShowUploadModal] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  // Initial demo data
  const defaultFiles = [
    {
      id: 'blob-101',
      name: 'hero_banner.png',
      bucket: 'media-assets',
      type: 'IMAGE',
      size: '1.2 MB',
      mime: 'image/png',
      url: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=600&q=80',
      created_at: '2026-08-08 07:00',
    },
    {
      id: 'blob-102',
      name: 'product_demo.mp4',
      bucket: 'media-assets',
      type: 'VIDEO',
      size: '14.5 MB',
      mime: 'video/mp4',
      url: 'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4',
      created_at: '2026-08-08 07:15',
    },
    {
      id: 'blob-103',
      name: 'podcast_track.mp3',
      bucket: 'audio-vault',
      type: 'AUDIO',
      size: '4.8 MB',
      mime: 'audio/mp3',
      url: 'https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3',
      created_at: '2026-08-08 07:20',
    },
  ]

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        try {
          setFiles(JSON.parse(stored))
        } catch {
          setFiles(defaultFiles)
          localStorage.setItem(STORAGE_KEY, JSON.stringify(defaultFiles))
        }
      } else {
        setFiles(defaultFiles)
        localStorage.setItem(STORAGE_KEY, JSON.stringify(defaultFiles))
      }
    }
  }, [])

  const updateFiles = (newFiles: any[]) => {
    setFiles(newFiles)
    if (typeof window !== 'undefined') {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newFiles))
    }
  }

  const handleCreateBucket = (e: React.FormEvent) => {
    e.preventDefault()
    if (newBucketName && !buckets.includes(newBucketName)) {
      setBuckets([...buckets, newBucketName])
      setActiveBucket(newBucketName)
    }
    setNewBucketName('')
    setShowBucketModal(false)
  }

  const handleUploadFile = (e: React.FormEvent) => {
    e.preventDefault()
    const newFile = {
      id: `blob-${Date.now()}`,
      name: uploadName || 'unstructured_object',
      bucket: activeBucket,
      type: fileType,
      size: fileType === 'VIDEO' ? '12.4 MB' : fileType === 'AUDIO' ? '3.5 MB' : '850 KB',
      mime: fileType === 'IMAGE' ? 'image/png' : fileType === 'VIDEO' ? 'video/mp4' : fileType === 'AUDIO' ? 'audio/mp3' : 'application/pdf',
      url: fileUrl || 'https://images.unsplash.com/photo-1579546929518-9e396f3cc809?auto=format&fit=crop&w=600&q=80',
      created_at: new Date().toISOString().replace('T', ' ').substring(0, 16),
    }

    updateFiles([newFile, ...files])
    setUploadName('')
    setFileUrl('')
    setShowUploadModal(false)
  }

  const handleDeleteFile = (id: string) => {
    if (confirm('Delete this unstructured media object?')) {
      updateFiles(files.filter((f) => f.id !== id))
    }
  }

  const handleCopyUrl = (url: string, id: string) => {
    navigator.clipboard.writeText(url)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const filteredFiles = files.filter((f) => f.bucket === activeBucket)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Unstructured Media Storage</h1>
          <p className="text-slate-400 mt-1">
            Store and serve unstructured binary objects: Images, Videos, Audio, and Documents.
          </p>
        </div>

        <div className="flex gap-3">
          <button
            onClick={() => setShowBucketModal(true)}
            className="px-4 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-200 font-semibold rounded-lg transition border border-slate-700 text-sm"
          >
            + New Bucket
          </button>
          <button
            onClick={() => setShowUploadModal(true)}
            className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25 text-sm"
          >
            + Upload Object
          </button>
        </div>
      </div>

      {/* Bucket Selector Tabs */}
      <div className="flex items-center gap-2 overflow-x-auto border-b border-slate-800 pb-3">
        {buckets.map((b) => (
          <button
            key={b}
            onClick={() => setActiveBucket(b)}
            className={`px-4 py-2 text-sm font-semibold rounded-xl transition ${
              activeBucket === b
                ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20'
                : 'text-slate-400 hover:bg-slate-900 hover:text-slate-200'
            }`}
          >
            📁 {b}
          </button>
        ))}
      </div>

      {/* Media Objects Grid */}
      {filteredFiles.length === 0 ? (
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-12 text-center space-y-4">
          <div className="inline-flex items-center justify-center h-16 w-16 rounded-2xl bg-blue-600/10 text-blue-400 text-3xl font-bold border border-blue-500/20">
            📦
          </div>
          <h3 className="text-xl font-bold text-white">No Objects in Bucket '{activeBucket}'</h3>
          <p className="text-slate-400 text-sm max-w-md mx-auto">
            Upload images, videos, audio tracks, or documents to store unstructured data in this bucket.
          </p>
          <button
            onClick={() => setShowUploadModal(true)}
            className="px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition shadow-lg shadow-blue-600/25"
          >
            Upload Media Object
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredFiles.map((file) => (
            <div key={file.id} className="bg-slate-900 border border-slate-800 rounded-2xl overflow-hidden flex flex-col justify-between">
              {/* Media Preview Player */}
              <div className="h-48 bg-slate-950 flex items-center justify-center overflow-hidden relative">
                {file.type === 'IMAGE' && (
                  <img src={file.url} alt={file.name} className="w-full h-full object-cover" />
                )}

                {file.type === 'VIDEO' && (
                  <video src={file.url} controls className="w-full h-full object-cover" />
                )}

                {file.type === 'AUDIO' && (
                  <div className="w-full px-4 space-y-2 text-center">
                    <div className="text-3xl">🎵</div>
                    <audio src={file.url} controls className="w-full" />
                  </div>
                )}

                {file.type === 'DOCUMENT' && (
                  <div className="text-center space-y-2">
                    <div className="text-4xl">📄</div>
                    <div className="text-xs text-slate-400">{file.mime}</div>
                  </div>
                )}

                <span className="absolute top-3 right-3 px-2.5 py-1 text-xs font-bold bg-slate-950/80 backdrop-blur text-slate-200 border border-slate-800 rounded-full">
                  {file.type}
                </span>
              </div>

              {/* Object Details & Actions */}
              <div className="p-5 space-y-3">
                <div>
                  <h3 className="font-bold text-white text-base truncate">{file.name}</h3>
                  <div className="text-xs text-slate-400 font-mono mt-0.5">{file.id} • {file.size}</div>
                </div>

                <div className="flex gap-2 pt-2 border-t border-slate-800">
                  <button
                    onClick={() => handleCopyUrl(file.url, file.id)}
                    className="flex-1 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-lg transition"
                  >
                    {copiedId === file.id ? '✔ Copied URL!' : 'Copy Object URL'}
                  </button>
                  <button
                    onClick={() => handleDeleteFile(file.id)}
                    className="px-3 py-1.5 bg-red-500/10 hover:bg-red-500/20 text-red-400 text-xs font-semibold rounded-lg transition border border-red-500/20"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Bucket Modal */}
      {showBucketModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Create Storage Bucket</h2>
            <form onSubmit={handleCreateBucket} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Bucket Name</label>
                <input
                  type="text"
                  required
                  value={newBucketName}
                  onChange={(e) => setNewBucketName(e.target.value)}
                  placeholder="e.g. video-uploads"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowBucketModal(false)}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25"
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
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Upload Unstructured Media Object</h2>
            <form onSubmit={handleUploadFile} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Object Name</label>
                <input
                  type="text"
                  required
                  value={uploadName}
                  onChange={(e) => setUploadName(e.target.value)}
                  placeholder="e.g. promotional_video.mp4"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Media Type</label>
                <select
                  value={fileType}
                  onChange={(e) => setFileType(e.target.value as any)}
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                >
                  <option value="IMAGE">Image (PNG, JPG, WEBP)</option>
                  <option value="VIDEO">Video (MP4, WEBM)</option>
                  <option value="AUDIO">Audio (MP3, WAV)</option>
                  <option value="DOCUMENT">Document (PDF, DOCX)</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Media File URL / CDN Link</label>
                <input
                  type="text"
                  value={fileUrl}
                  onChange={(e) => setFileUrl(e.target.value)}
                  placeholder="e.g. https://... or leave blank for sample player"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500 text-xs font-mono"
                />
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowUploadModal(false)}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25"
                >
                  Upload to Bucket
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
