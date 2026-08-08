'use client'

import React, { useState, useEffect } from 'react'

const STORAGE_KEY = 'anarva_unstructured_blobs'

export default function UnstructuredStoragePage() {
  const [buckets, setBuckets] = useState<string[]>(['Images', 'Videos', 'Audio', 'Documents'])
  const [activeBucket, setActiveBucket] = useState<string>('Images')

  const [files, setFiles] = useState<any[]>([])
  const [newBucketName, setNewBucketName] = useState('')
  const [showBucketModal, setShowBucketModal] = useState(false)
  const [showUploadModal, setShowUploadModal] = useState(false)

  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  // Default initial demo objects
  const defaultFiles = [
    {
      id: 'blob-101',
      name: 'architecture_diagram.png',
      bucket: 'Images',
      type: 'IMAGE',
      size: '1.2 MB',
      mime: 'image/png',
      url: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=600&q=80',
      created_at: '2026-08-08 07:00',
    },
    {
      id: 'blob-102',
      name: 'platform_demo.mp4',
      bucket: 'Videos',
      type: 'VIDEO',
      size: '14.5 MB',
      mime: 'video/mp4',
      url: 'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4',
      created_at: '2026-08-08 07:15',
    },
    {
      id: 'blob-103',
      name: 'podcast_track.mp3',
      bucket: 'Audio',
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

  // Automatic bucket classification based on file extension
  const categorizeFile = (fileName: string): { bucket: string; type: 'IMAGE' | 'VIDEO' | 'AUDIO' | 'DOCUMENT' } => {
    const ext = fileName.split('.').pop()?.toLowerCase() || ''

    if (['png', 'jpg', 'jpeg', 'webp', 'gif', 'svg', 'bmp', 'ico'].includes(ext)) {
      return { bucket: 'Images', type: 'IMAGE' }
    } else if (['mp4', 'webm', 'mov', 'avi', 'mkv', 'flv'].includes(ext)) {
      return { bucket: 'Videos', type: 'VIDEO' }
    } else if (['mp3', 'wav', 'aac', 'ogg', 'flac', 'm4a'].includes(ext)) {
      return { bucket: 'Audio', type: 'AUDIO' }
    } else {
      return { bucket: 'Documents', type: 'DOCUMENT' }
    }
  }

  const handleLocalFileUpload = (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedFile) return

    const { bucket, type } = categorizeFile(selectedFile.name)
    const localUrl = URL.createObjectURL(selectedFile)

    // Ensure bucket exists in tabs
    if (!buckets.includes(bucket)) {
      setBuckets([...buckets, bucket])
    }

    const fileSizeFormatted =
      selectedFile.size > 1024 * 1024
        ? `${(selectedFile.size / (1024 * 1024)).toFixed(1)} MB`
        : `${(selectedFile.size / 1024).toFixed(0)} KB`

    const newFileObj = {
      id: `blob-${Date.now()}`,
      name: selectedFile.name,
      bucket: bucket,
      type: type,
      size: fileSizeFormatted,
      mime: selectedFile.type || 'application/octet-stream',
      url: localUrl,
      created_at: new Date().toISOString().replace('T', ' ').substring(0, 16),
    }

    updateFiles([newFileObj, ...files])
    setActiveBucket(bucket) // Auto-switch tab to relevant bucket!
    setSelectedFile(null)
    setShowUploadModal(false)
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
            Store, categorize, and serve unstructured binary objects: Images, Videos, Audio, and Documents.
          </p>
        </div>

        <div className="flex gap-3">
          <button
            onClick={() => setShowBucketModal(true)}
            className="px-4 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-200 font-semibold rounded-lg transition border border-slate-700 text-sm"
          >
            + Custom Bucket
          </button>
          <button
            onClick={() => setShowUploadModal(true)}
            className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25 text-sm"
          >
            + Upload File from Computer
          </button>
        </div>
      </div>

      {/* Bucket Selector Tabs */}
      <div className="flex items-center gap-2 overflow-x-auto border-b border-slate-800 pb-3">
        {buckets.map((b) => {
          const count = files.filter((f) => f.bucket === b).length
          return (
            <button
              key={b}
              onClick={() => setActiveBucket(b)}
              className={`px-4 py-2 text-sm font-semibold rounded-xl transition flex items-center gap-2 ${
                activeBucket === b
                  ? 'bg-blue-600/10 text-blue-400 border border-blue-500/20'
                  : 'text-slate-400 hover:bg-slate-900 hover:text-slate-200'
              }`}
            >
              <span>📁 {b}</span>
              <span className="px-2 py-0.5 text-xs bg-slate-800 text-slate-300 rounded-full font-mono">
                {count}
              </span>
            </button>
          )
        })}
      </div>

      {/* Media Objects Grid */}
      {filteredFiles.length === 0 ? (
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-12 text-center space-y-4">
          <div className="inline-flex items-center justify-center h-16 w-16 rounded-2xl bg-blue-600/10 text-blue-400 text-3xl font-bold border border-blue-500/20">
            📦
          </div>
          <h3 className="text-xl font-bold text-white">No Objects in Bucket '{activeBucket}'</h3>
          <p className="text-slate-400 text-sm max-w-md mx-auto">
            Upload files directly from your computer. Based on the file extension, objects are automatically categorized into their relevant bucket!
          </p>
          <button
            onClick={() => setShowUploadModal(true)}
            className="px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition shadow-lg shadow-blue-600/25"
          >
            Upload File from Computer
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
                  <div className="text-center space-y-2 p-4">
                    <div className="text-4xl">📄</div>
                    <div className="text-xs text-slate-300 font-mono truncate max-w-xs">{file.name}</div>
                    <div className="text-xs text-slate-500">{file.mime}</div>
                  </div>
                )}

                <span className="absolute top-3 right-3 px-2.5 py-1 text-xs font-bold bg-slate-950/80 backdrop-blur text-slate-200 border border-slate-800 rounded-full">
                  {file.type}
                </span>
              </div>

              {/* Object Details & Actions */}
              <div className="p-5 space-y-3">
                <div>
                  <h3 className="font-bold text-white text-base truncate" title={file.name}>
                    {file.name}
                  </h3>
                  <div className="text-xs text-slate-400 font-mono mt-0.5">
                    {file.id} • {file.size}
                  </div>
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

      {/* Upload Computer File Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Upload File from Computer</h2>
            <p className="text-xs text-slate-400">
              Files are automatically categorized into <span className="text-blue-400 font-semibold">Images</span>, <span className="text-blue-400 font-semibold">Videos</span>, <span className="text-blue-400 font-semibold">Audio</span>, or <span className="text-blue-400 font-semibold">Documents</span> based on extension.
            </p>

            <form onSubmit={handleLocalFileUpload} className="space-y-4">
              <div className="border-2 border-dashed border-slate-800 rounded-2xl p-6 text-center hover:border-blue-500/50 transition">
                <input
                  type="file"
                  required
                  id="file-input"
                  onChange={(e) => {
                    if (e.target.files && e.target.files[0]) {
                      setSelectedFile(e.target.files[0])
                    }
                  }}
                  className="hidden"
                />
                <label htmlFor="file-input" className="cursor-pointer space-y-2 block">
                  <div className="text-4xl">📁</div>
                  <div className="text-sm font-semibold text-blue-400 hover:underline">
                    {selectedFile ? selectedFile.name : 'Choose file from your computer'}
                  </div>
                  {selectedFile && (
                    <div className="text-xs text-emerald-400 font-mono">
                      Auto-categorized into bucket:{' '}
                      <span className="font-bold uppercase">{categorizeFile(selectedFile.name).bucket}</span>
                    </div>
                  )}
                  <div className="text-xs text-slate-500">
                    Supports PNG, JPG, MP4, MP3, PDF, DOCX, ZIP & more
                  </div>
                </label>
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => {
                    setSelectedFile(null)
                    setShowUploadModal(false)
                  }}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={!selectedFile}
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25 disabled:opacity-50"
                >
                  Upload & Categorize
                </button>
              </div>
            </form>
          </div>
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
                  placeholder="e.g. Archive-Vault"
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
    </div>
  )
}
