'use client'

import React from 'react'
import Link from 'next/link'

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gray-950 text-gray-100 font-sans selection:bg-cyan-500 selection:text-white">
      {/* Navigation Header */}
      <header className="h-16 border-b border-gray-800/80 px-6 max-w-7xl mx-auto flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center text-white font-bold shadow-lg shadow-cyan-500/20">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
            </svg>
          </div>
          <span className="font-mono text-lg font-extrabold tracking-tight text-white">ANARVA</span>
        </div>

        <nav className="hidden md:flex items-center gap-8 text-xs font-mono text-gray-400 uppercase tracking-wider">
          <a href="#overview" className="hover:text-white transition-colors">Overview</a>
          <a href="#architecture" className="hover:text-white transition-colors">Architecture</a>
          <a href="#security" className="hover:text-white transition-colors">Security</a>
          <a href="#reliability" className="hover:text-white transition-colors">Reliability</a>
          <a href="#developer" className="hover:text-white transition-colors">Developer API</a>
        </nav>

        <div className="flex items-center gap-3">
          <Link href="/login" className="px-4 py-2 text-xs font-mono text-gray-300 hover:text-white transition-colors">
            Sign In
          </Link>
          <Link href="/console" className="px-4 py-2 text-xs font-mono font-bold text-gray-950 bg-cyan-400 hover:bg-cyan-300 rounded-lg shadow-lg shadow-cyan-500/20 transition-all">
            Launch Console →
          </Link>
        </div>
      </header>

      {/* Hero Section */}
      <section className="py-24 px-6 max-w-6xl mx-auto text-center space-y-8">
        <div className="inline-flex items-center gap-2 px-3 py-1 bg-cyan-500/10 border border-cyan-500/20 rounded-full text-xs font-mono text-cyan-400">
          <span className="w-2 h-2 rounded-full bg-cyan-400 animate-pulse"></span>
          <span>ANARVA Platform v0.1.0 Released</span>
        </div>

        <h1 className="text-4xl sm:text-6xl font-extrabold text-white tracking-tight leading-tight max-w-4xl mx-auto">
          Your Infrastructure Control Plane
        </h1>

        <p className="text-lg sm:text-xl text-gray-400 max-w-2xl mx-auto font-light leading-relaxed">
          Provision, manage, observe, secure, and automate cloud infrastructure through one developer-first control plane.
        </p>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
          <Link href="/console" className="w-full sm:w-auto px-6 py-3 text-sm font-mono font-bold text-gray-950 bg-cyan-400 hover:bg-cyan-300 rounded-xl shadow-xl shadow-cyan-500/20 transition-all">
            Launch Console →
          </Link>
          <a href="#overview" className="w-full sm:w-auto px-6 py-3 text-sm font-mono text-gray-300 bg-gray-900 border border-gray-800 hover:bg-gray-800 rounded-xl transition-all">
            Explore Platform
          </a>
        </div>

        {/* Control Plane Architecture Visualization */}
        <div className="mt-16 p-6 sm:p-8 bg-gray-900/60 border border-gray-800 rounded-2xl max-w-4xl mx-auto space-y-6">
          <div className="flex items-center justify-between border-b border-gray-800 pb-4 text-xs font-mono text-gray-400">
            <span>ANARVA CONTROL PLANE FLOW</span>
            <span className="text-emerald-400 font-bold">● SYSTEM HEALTHY</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-5 gap-3 text-xs font-mono">
            <div className="p-4 bg-gray-950 border border-gray-800 rounded-xl text-cyan-400 text-center">
              <span className="block font-bold mb-1">1. CONSOLE</span>
              <span className="text-[10px] text-gray-500">UI / CLI / SDK</span>
            </div>
            <div className="p-4 bg-gray-950 border border-gray-800 rounded-xl text-cyan-400 text-center">
              <span className="block font-bold mb-1">2. CONTROL PLANE</span>
              <span className="text-[10px] text-gray-500">API Gateway & IAM</span>
            </div>
            <div className="p-4 bg-gray-950 border border-gray-800 rounded-xl text-cyan-400 text-center">
              <span className="block font-bold mb-1">3. RESOURCES</span>
              <span className="text-[10px] text-gray-500">Compute / DB / S3</span>
            </div>
            <div className="p-4 bg-gray-950 border border-gray-800 rounded-xl text-cyan-400 text-center">
              <span className="block font-bold mb-1">4. OPERATIONS</span>
              <span className="text-[10px] text-gray-500">Locks & Recovery</span>
            </div>
            <div className="p-4 bg-gray-950 border border-gray-800 rounded-xl text-cyan-400 text-center">
              <span className="block font-bold mb-1">5. OBSERVABILITY</span>
              <span className="text-[10px] text-gray-500">Metrics & Audit</span>
            </div>
          </div>
        </div>
      </section>

      {/* Overview Grid Section */}
      <section id="overview" className="py-20 px-6 max-w-6xl mx-auto space-y-12 border-t border-gray-800/80">
        <div className="text-center space-y-3">
          <h2 className="text-2xl sm:text-3xl font-extrabold text-white">Unified Infrastructure Control</h2>
          <p className="text-sm text-gray-400 max-w-xl mx-auto">
            Everything required to operate production compute, database, and storage resources in one unified platform.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="p-6 bg-gray-900/60 border border-gray-800 rounded-xl space-y-3">
            <div className="w-10 h-10 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 flex items-center justify-center font-bold">
              ⚡
            </div>
            <h3 className="text-base font-bold text-white">Anarva Compute</h3>
            <p className="text-xs text-gray-400 leading-relaxed">
              Virtual compute instances with instant provisioning, automated health monitoring, and complete lifecycle controls.
            </p>
          </div>

          <div className="p-6 bg-gray-900/60 border border-gray-800 rounded-xl space-y-3">
            <div className="w-10 h-10 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 flex items-center justify-center font-bold">
              🗄️
            </div>
            <h3 className="text-base font-bold text-white">Managed Databases</h3>
            <p className="text-xs text-gray-400 leading-relaxed">
              PostgreSQL and MySQL database clusters featuring automated PITR backups, Multi-AZ failover, and connection pooling.
            </p>
          </div>

          <div className="p-6 bg-gray-900/60 border border-gray-800 rounded-xl space-y-3">
            <div className="w-10 h-10 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 flex items-center justify-center font-bold">
              📦
            </div>
            <h3 className="text-base font-bold text-white">Object Storage</h3>
            <p className="text-xs text-gray-400 leading-relaxed">
              High-availability object storage with presigned URLs, path traversal protection, and tenant-level access isolation.
            </p>
          </div>
        </div>
      </section>

      {/* Developer Tooling Section */}
      <section id="developer" className="py-20 px-6 max-w-6xl mx-auto space-y-12 border-t border-gray-800/80">
        <div className="text-center space-y-3">
          <h2 className="text-2xl sm:text-3xl font-extrabold text-white">Developer-First Platform</h2>
          <p className="text-sm text-gray-400 max-w-xl mx-auto">
            Manage infrastructure using the Anarva CLI, TypeScript SDK, or official Terraform Provider.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 font-mono text-xs">
          <div className="p-6 bg-gray-900/80 border border-gray-800 rounded-xl space-y-3">
            <div className="flex justify-between text-gray-400 text-[10px] uppercase border-b border-gray-800 pb-2">
              <span>ANARVA CLI</span>
              <span className="text-cyan-400">v0.1.0</span>
            </div>
            <pre className="text-cyan-300 overflow-x-auto p-3 bg-gray-950 rounded border border-gray-800/60">
{`$ anarva db list
$ anarva compute launch --type c1.large
$ anarva backup create --db-id db-101`}
            </pre>
          </div>

          <div className="p-6 bg-gray-900/80 border border-gray-800 rounded-xl space-y-3">
            <div className="flex justify-between text-gray-400 text-[10px] uppercase border-b border-gray-800 pb-2">
              <span>TERRAFORM PROVIDER</span>
              <span className="text-cyan-400">provider.anarva</span>
            </div>
            <pre className="text-cyan-300 overflow-x-auto p-3 bg-gray-950 rounded border border-gray-800/60">
{`resource "anarva_database" "primary" {
  name   = "prod-db"
  engine = "postgres"
}`}
            </pre>
          </div>
        </div>
      </section>

      {/* CTA Footer */}
      <footer className="py-16 px-6 border-t border-gray-800/80 text-center space-y-6">
        <h2 className="text-2xl font-extrabold text-white">Ready to operate your infrastructure?</h2>
        <div>
          <Link href="/console" className="px-6 py-3 text-sm font-mono font-bold text-gray-950 bg-cyan-400 hover:bg-cyan-300 rounded-xl shadow-xl shadow-cyan-500/20 transition-all">
            Launch ANARVA Console →
          </Link>
        </div>
        <p className="text-xs font-mono text-gray-500 pt-8">
          © 2026 ANARVA Cloud Platform. All rights reserved.
        </p>
      </footer>
    </div>
  )
}
