'use client'

import React, { useState, useEffect } from 'react'

const PERSON_STORAGE_KEY = 'anarva_person_profiles'
const BLOB_STORAGE_KEY = 'anarva_person_blobs'

interface Person {
  id: string
  name: string
  email: string
  role: string
}

interface PersonBlob {
  id: string
  person_id: string
  name: string
  bucket: 'Images' | 'Videos' | 'Audio' | 'Documents' | 'Links'
  type: 'IMAGE' | 'VIDEO' | 'AUDIO' | 'DOCUMENT' | 'LINK'
  size: string
  mime: string
  url: string
  created_at: string
}

export default function UnstructuredStoragePage() {
  // Current user auth session state
  const [currentUser, setCurrentUser] = useState<Person>({
    id: 'usr-87a1',
    name: 'Lokesh Ashapu',
    email: 'lokesh@anarva.io',
    role: 'Database Admin',
  })
  const [isAdmin, setIsAdmin] = useState<boolean>(true)

  // Person profiles state
  const [persons, setPersons] = useState<Person[]>([
    { id: 'usr-87a1', name: 'Lokesh Ashapu', email: 'lokesh@anarva.io', role: 'Database Admin' },
    { id: 'usr-92c4', name: 'Enterprise Client', email: 'enterprise@acme.com', role: 'Org Member' },
    { id: 'usr-11f8', name: 'Dev Team Lead', email: 'devlead@anarva.io', role: 'Engineer' },
  ])
  const [activePersonId, setActivePersonId] = useState<string>('usr-87a1')

  // Buckets state
  const [buckets, setBuckets] = useState<string[]>(['Images', 'Videos', 'Audio', 'Documents', 'Links'])
  const [activeBucket, setActiveBucket] = useState<string>('Images')

  const [files, setFiles] = useState<PersonBlob[]>([])

  // Modal states
  const [showPersonModal, setShowPersonModal] = useState(false)
  const [newPersonName, setNewPersonName] = useState('')
  const [newPersonEmail, setNewPersonEmail] = useState('')
  const [newPersonRole, setNewPersonRole] = useState('Member')

  const [showUploadModal, setShowUploadModal] = useState(false)
  const [uploadMode, setUploadMode] = useState<'FILE' | 'LINK'>('FILE')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [linkTitle, setLinkTitle] = useState('')
  const [linkUrl, setLinkUrl] = useState('')

  const [copiedId, setCopiedId] = useState<string | null>(null)

  // Default initial demo objects linked to persons
  const defaultBlobs: PersonBlob[] = [
    {
      id: 'blob-101',
      person_id: 'usr-87a1',
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
      person_id: 'usr-87a1',
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
      person_id: 'usr-87a1',
      name: 'voice_memo_instructions.mp3',
      bucket: 'Audio',
      type: 'AUDIO',
      size: '4.8 MB',
      mime: 'audio/mp3',
      url: 'https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3',
      created_at: '2026-08-08 07:20',
    },
    {
      id: 'blob-104',
      person_id: 'usr-87a1',
      name: 'GitHub Repository Codebase',
      bucket: 'Links',
      type: 'LINK',
      size: 'URL Link',
      mime: 'text/html',
      url: 'https://github.com/LokeshAshapu/anarva-cloud-db',
      created_at: '2026-08-08 07:30',
    },
    {
      id: 'blob-105',
      person_id: 'usr-92c4',
      name: 'enterprise_contract.pdf',
      bucket: 'Documents',
      type: 'DOCUMENT',
      size: '2.4 MB',
      mime: 'application/pdf',
      url: 'https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf',
      created_at: '2026-08-08 07:35',
    },
  ]

  useEffect(() => {
    if (typeof window !== 'undefined') {
      // Check logged in user session
      let userList: any[] = []
      try {
        userList = JSON.parse(localStorage.getItem('anarva_registered_users') || '[]')
      } catch {}

      if (userList.length > 0) {
        const lastUser = userList[userList.length - 1]
        const userRole: string = lastUser.email.includes('admin') || lastUser.email.includes('owner') ? 'Database Admin' : 'Member'
        const userObj: Person = {
          id: `usr-${lastUser.email.substring(0, 4)}`,
          name: lastUser.fullName || lastUser.email.split('@')[0],
          email: lastUser.email,
          role: userRole,
        }

        setCurrentUser(userObj)
        const isUserAdmin = userRole === 'Database Admin' || userRole === 'OWNER' || userRole === 'ADMIN'
        setIsAdmin(isUserAdmin)

        if (!isUserAdmin) {
          setActivePersonId(userObj.id)
        }
      }

      const storedPersons = localStorage.getItem(PERSON_STORAGE_KEY)
      if (storedPersons) {
        try {
          setPersons(JSON.parse(storedPersons))
        } catch {}
      }

      const storedBlobs = localStorage.getItem(BLOB_STORAGE_KEY)
      if (storedBlobs) {
        try {
          setFiles(JSON.parse(storedBlobs))
        } catch {
          setFiles(defaultBlobs)
          localStorage.setItem(BLOB_STORAGE_KEY, JSON.stringify(defaultBlobs))
        }
      } else {
        setFiles(defaultBlobs)
        localStorage.setItem(BLOB_STORAGE_KEY, JSON.stringify(defaultBlobs))
      }
    }
  }, [])

  const updateFiles = (newFiles: PersonBlob[]) => {
    setFiles(newFiles)
    if (typeof window !== 'undefined') {
      localStorage.setItem(BLOB_STORAGE_KEY, JSON.stringify(newFiles))
    }
  }

  const updatePersons = (newPersons: Person[]) => {
    setPersons(newPersons)
    if (typeof window !== 'undefined') {
      localStorage.setItem(PERSON_STORAGE_KEY, JSON.stringify(newPersons))
    }
  }

  const categorizeFile = (fileName: string): { bucket: 'Images' | 'Videos' | 'Audio' | 'Documents' | 'Links'; type: 'IMAGE' | 'VIDEO' | 'AUDIO' | 'DOCUMENT' | 'LINK' } => {
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

  const handleAddPerson = (e: React.FormEvent) => {
    e.preventDefault()
    if (!isAdmin) return
    const newPerson: Person = {
      id: `usr-${Math.random().toString(36).substring(2, 6)}`,
      name: newPersonName || 'New User Profile',
      email: newPersonEmail || 'user@anarva.io',
      role: newPersonRole,
    }
    updatePersons([...persons, newPerson])
    setActivePersonId(newPerson.id)
    setNewPersonName('')
    setNewPersonEmail('')
    setShowPersonModal(false)
  }

  const handleUploadFileOrLink = (e: React.FormEvent) => {
    e.preventDefault()
    const targetPersonId = isAdmin ? activePersonId : currentUser.id

    if (uploadMode === 'FILE' && selectedFile) {
      const { bucket, type } = categorizeFile(selectedFile.name)
      const localUrl = URL.createObjectURL(selectedFile)

      const fileSizeFormatted =
        selectedFile.size > 1024 * 1024
          ? `${(selectedFile.size / (1024 * 1024)).toFixed(1)} MB`
          : `${(selectedFile.size / 1024).toFixed(0)} KB`

      const newBlob: PersonBlob = {
        id: `blob-${Date.now()}`,
        person_id: targetPersonId,
        name: selectedFile.name,
        bucket: bucket,
        type: type,
        size: fileSizeFormatted,
        mime: selectedFile.type || 'application/octet-stream',
        url: localUrl,
        created_at: new Date().toISOString().replace('T', ' ').substring(0, 16),
      }

      updateFiles([newBlob, ...files])
      setActiveBucket(bucket)
      setSelectedFile(null)
    } else if (uploadMode === 'LINK' && linkUrl) {
      const newBlob: PersonBlob = {
        id: `blob-${Date.now()}`,
        person_id: targetPersonId,
        name: linkTitle || linkUrl,
        bucket: 'Links',
        type: 'LINK',
        size: 'URL Link',
        mime: 'text/html',
        url: linkUrl.startsWith('http') ? linkUrl : `https://${linkUrl}`,
        created_at: new Date().toISOString().replace('T', ' ').substring(0, 16),
      }

      updateFiles([newBlob, ...files])
      setActiveBucket('Links')
      setLinkTitle('')
      setLinkUrl('')
    }

    setShowUploadModal(false)
  }

  const handleDeleteFile = (id: string) => {
    if (confirm('Delete this media object/link from this person record?')) {
      updateFiles(files.filter((f) => f.id !== id))
    }
  }

  const handleCopyUrl = (url: string, id: string) => {
    navigator.clipboard.writeText(url)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  // Active Person container selection logic
  const effectiveActivePersonId = isAdmin ? activePersonId : currentUser.id
  const activePerson = isAdmin
    ? persons.find((p) => p.id === effectiveActivePersonId) || currentUser
    : currentUser

  const personFiles = files.filter((f) => f.person_id === effectiveActivePersonId)
  const filteredFiles = personFiles.filter((f) => f.bucket === activeBucket)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Person-Centric Database & Storage</h1>
          <p className="text-slate-400 mt-1">
            Unified entity database: Every person record holds all their structured profile data, files, and links.
          </p>
        </div>

        <div className="flex gap-3">
          {isAdmin && (
            <button
              onClick={() => setShowPersonModal(true)}
              className="px-4 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-200 font-semibold rounded-lg transition border border-slate-700 text-sm"
            >
              + Create Person Profile
            </button>
          )}
          <button
            onClick={() => setShowUploadModal(true)}
            className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25 text-sm"
          >
            + Push File / Link to Container
          </button>
        </div>
      </div>

      {/* Person Selector Bar (Only Visible for Admins / Owners) */}
      {isAdmin ? (
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 space-y-3">
          <div className="flex items-center justify-between">
            <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
              Select Person Record (Admin Multi-Tenant View)
            </div>
            <span className="text-xs text-blue-400 font-mono">Person ID: {activePerson.id}</span>
          </div>

          <div className="flex items-center gap-3 overflow-x-auto">
            {persons.map((p) => {
              const pFileCount = files.filter((f) => f.person_id === p.id).length
              const isSelected = p.id === activePersonId
              return (
                <button
                  key={p.id}
                  onClick={() => setActivePersonId(p.id)}
                  className={`px-4 py-3 rounded-xl border text-left transition min-w-[200px] flex items-center justify-between ${
                    isSelected
                      ? 'bg-blue-600/10 border-blue-500/40 text-white shadow-lg shadow-blue-600/10'
                      : 'bg-slate-950 border-slate-800 text-slate-400 hover:border-slate-700'
                  }`}
                >
                  <div>
                    <div className="font-bold text-sm">{p.name}</div>
                    <div className="text-xs text-slate-400 truncate max-w-[140px]">{p.email}</div>
                  </div>
                  <span className="px-2 py-0.5 text-xs bg-slate-800 text-slate-300 rounded-full font-mono font-semibold">
                    {pFileCount}
                  </span>
                </button>
              )
            })}
          </div>
        </div>
      ) : (
        <div className="p-3 bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-semibold rounded-xl flex items-center gap-2">
          <span>🔒</span>
          <span>Zero-Trust Access Isolation Enabled: You are logged in as standard user <strong className="text-white">{currentUser.name}</strong> ({currentUser.email}). Other users' databases and content are restricted.</span>
        </div>
      )}

      {/* Person Container Info */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 flex items-center justify-between text-xs text-slate-300 font-mono">
        <div>
          Active Container: <span className="text-white font-bold">{activePerson.name}</span> ({activePerson.email}) • Role: <span className="text-blue-400">{activePerson.role}</span>
        </div>
        <div>
          Total Attached Media & Links: <span className="text-emerald-400 font-bold">{personFiles.length} items</span>
        </div>
      </div>

      {/* Bucket Category Tabs */}
      <div className="flex items-center gap-2 overflow-x-auto border-b border-slate-800 pb-3">
        {buckets.map((b) => {
          const count = personFiles.filter((f) => f.bucket === b).length
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
              <span>{b === 'Links' ? '🔗' : '📁'} {b}</span>
              <span className="px-2 py-0.5 text-xs bg-slate-800 text-slate-300 rounded-full font-mono">
                {count}
              </span>
            </button>
          )
        })}
      </div>

      {/* Media & Links Grid */}
      {filteredFiles.length === 0 ? (
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-12 text-center space-y-4">
          <div className="inline-flex items-center justify-center h-16 w-16 rounded-2xl bg-blue-600/10 text-blue-400 text-3xl font-bold border border-blue-500/20">
            {activeBucket === 'Links' ? '🔗' : '📦'}
          </div>
          <h3 className="text-xl font-bold text-white">
            No {activeBucket} Pushed for {activePerson.name}
          </h3>
          <p className="text-slate-400 text-sm max-w-md mx-auto">
            Push computer files or external URLs/links directly into your personal record container!
          </p>
          <button
            onClick={() => setShowUploadModal(true)}
            className="px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition shadow-lg shadow-blue-600/25"
          >
            Push File / Link to Container
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredFiles.map((file) => (
            <div key={file.id} className="bg-slate-900 border border-slate-800 rounded-2xl overflow-hidden flex flex-col justify-between">
              {/* Preview Box */}
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

                {file.type === 'LINK' && (
                  <div className="text-center space-y-2 p-4">
                    <div className="text-5xl">🔗</div>
                    <div className="text-sm font-bold text-blue-400 truncate max-w-xs">{file.name}</div>
                    <div className="text-xs text-slate-400 font-mono truncate max-w-xs">{file.url}</div>
                  </div>
                )}

                <span className="absolute top-3 right-3 px-2.5 py-1 text-xs font-bold bg-slate-950/80 backdrop-blur text-slate-200 border border-slate-800 rounded-full">
                  {file.type}
                </span>
              </div>

              {/* Details & Actions */}
              <div className="p-5 space-y-3">
                <div>
                  <h3 className="font-bold text-white text-base truncate" title={file.name}>
                    {file.name}
                  </h3>
                  <div className="text-xs text-slate-400 font-mono mt-0.5">
                    Owner: <span className="text-blue-400">{activePerson.name}</span> • {file.size}
                  </div>
                </div>

                <div className="flex gap-2 pt-2 border-t border-slate-800">
                  {file.type === 'LINK' ? (
                    <a
                      href={file.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex-1 py-1.5 bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold rounded-lg transition text-center"
                    >
                      Open Link ↗
                    </a>
                  ) : (
                    <button
                      onClick={() => handleCopyUrl(file.url, file.id)}
                      className="flex-1 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-lg transition"
                    >
                      {copiedId === file.id ? '✔ Copied URL!' : 'Copy Object URL'}
                    </button>
                  )}

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

      {/* Push File or Link Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Push to {activePerson.name}'s Container</h2>

            {/* Mode Switcher */}
            <div className="flex bg-slate-950 p-1 rounded-xl border border-slate-800">
              <button
                type="button"
                onClick={() => setUploadMode('FILE')}
                className={`flex-1 py-1.5 text-xs font-semibold rounded-lg transition ${
                  uploadMode === 'FILE' ? 'bg-blue-600 text-white' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                📁 Upload Computer File
              </button>
              <button
                type="button"
                onClick={() => setUploadMode('LINK')}
                className={`flex-1 py-1.5 text-xs font-semibold rounded-lg transition ${
                  uploadMode === 'LINK' ? 'bg-blue-600 text-white' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                🔗 Add Web / App Link
              </button>
            </div>

            <form onSubmit={handleUploadFileOrLink} className="space-y-4">
              {uploadMode === 'FILE' ? (
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
                        Auto-categorized into:{' '}
                        <span className="font-bold uppercase">{categorizeFile(selectedFile.name).bucket}</span>
                      </div>
                    )}
                    <div className="text-xs text-slate-500">Supports PNG, JPG, MP4, MP3, PDF & more</div>
                  </label>
                </div>
              ) : (
                <div className="space-y-3">
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Link Title / Label</label>
                    <input
                      type="text"
                      required
                      value={linkTitle}
                      onChange={(e) => setLinkTitle(e.target.value)}
                      placeholder="e.g. YouTube Video / Drive File Link"
                      className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">URL Address</label>
                    <input
                      type="text"
                      required
                      value={linkUrl}
                      onChange={(e) => setLinkUrl(e.target.value)}
                      placeholder="e.g. https://youtube.com/watch?v=..."
                      className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500 text-xs font-mono"
                    />
                  </div>
                </div>
              )}

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
                  disabled={uploadMode === 'FILE' ? !selectedFile : !linkUrl}
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25 disabled:opacity-50"
                >
                  Push to Container
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Create Person Profile Modal */}
      {showPersonModal && (
        <div className="fixed inset-0 bg-slate-950/80 backdrop-blur flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 w-full max-w-md space-y-4">
            <h2 className="text-xl font-bold text-white">Create Person Profile Container</h2>
            <form onSubmit={handleAddPerson} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Full Name</label>
                <input
                  type="text"
                  required
                  value={newPersonName}
                  onChange={(e) => setNewPersonName(e.target.value)}
                  placeholder="e.g. Sarah Jenkins"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Email Address</label>
                <input
                  type="email"
                  required
                  value={newPersonEmail}
                  onChange={(e) => setNewPersonEmail(e.target.value)}
                  placeholder="sarah@acme.com"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Role / Category</label>
                <input
                  type="text"
                  value={newPersonRole}
                  onChange={(e) => setNewPersonRole(e.target.value)}
                  placeholder="Member / Admin / Customer"
                  className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowPersonModal(false)}
                  className="flex-1 py-2 bg-slate-800 text-slate-300 text-sm font-semibold rounded-lg"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-lg shadow-lg shadow-blue-600/25"
                >
                  Create Person Container
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
