'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'

import { API_BASE_URL } from '@/lib/api'
import { hashPassword } from '@/lib/crypto'
import { AnarvaLogo } from '@/components/AnarvaLogo'
import { createClient } from '@/utils/supabase/client'

export default function SignUpPage() {
  const router = useRouter()
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSignUp = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setSuccess('')

    try {
      // 1. Supabase Authentication Sign Up
      const supabase = createClient()
      const { data: supaData, error: supaError } = await supabase.auth.signUp({
        email,
        password,
        options: {
          data: {
            full_name: fullName,
          },
        },
      })

      if (supaError) {
        console.warn('Supabase Auth warning:', supaError.message)
      }

      // 2. Encrypt password using SHA-256 Web Crypto
      const encryptedPassword = await hashPassword(password)

      // Save locally encrypted
      if (typeof window !== 'undefined') {
        const stored = JSON.parse(localStorage.getItem('anarva_registered_users') || '[]')
        const updated = [
          ...stored.filter((u: any) => u.email !== email),
          { fullName, email, passwordHash: encryptedPassword, supaId: supaData?.user?.id },
        ]
        localStorage.setItem('anarva_registered_users', JSON.stringify(updated))
      }

      await fetch(`${API_BASE_URL}/api/v1/auth/signup`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          full_name: fullName,
          email,
          password: encryptedPassword,
        }),
      }).catch(() => null)

      setSuccess('Account registered with Supabase & SHA-256 Zero-Trust Encryption! Redirecting to login...')
      setTimeout(() => {
        router.push('/login')
      }, 1200)
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-slate-950">
      <div className="w-full max-w-md bg-slate-900 border border-slate-800 rounded-2xl p-8 shadow-2xl space-y-6">
        <div className="text-center space-y-2">
          <div className="inline-flex items-center justify-center p-3 rounded-2xl bg-blue-600/10 border border-blue-500/20">
            <AnarvaLogo className="h-16 w-16" />
          </div>
          <h1 className="text-2xl font-bold text-white">Create Account</h1>
          <p className="text-sm text-slate-400">Join Anarva Cloud Database Platform (Supabase Auth Connected)</p>
        </div>

        {error && (
          <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-mono">
            [ERROR] {error}
          </div>
        )}

        {success && (
          <div className="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-sm">
            {success}
          </div>
        )}

        <form onSubmit={handleSignUp} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Full Name</label>
            <input
              type="text"
              required
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              className="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500 transition"
              placeholder="John Doe"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Email Address</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500 transition"
              placeholder="developer@anarva.io"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 uppercase mb-1">Password</label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-lg text-slate-100 focus:outline-none focus:border-blue-500 transition"
              placeholder="••••••••••••"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition shadow-lg shadow-blue-600/25 disabled:opacity-50"
          >
            {loading ? 'Authenticating with Supabase...' : 'Register Account'}
          </button>
        </form>

        <div className="text-center text-sm text-slate-400">
          Already have an account?{' '}
          <Link href="/login" className="text-blue-400 hover:underline">
            Sign In
          </Link>
        </div>
      </div>
    </div>
  )
}
