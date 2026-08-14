import { NextResponse } from 'next/server'

export async function GET(request: Request) {
  return NextResponse.json({
    data: [
      {
        id: 'ank-101',
        name: 'Primary CLI Key',
        keyPrefix: 'ank_live_9f82...',
        status: 'ACTIVE',
        permissions: ['compute.read', 'compute.create', 'database.read', 'storage.read', 'network.read', 'provisioning.read'],
        createdBy: 'operator@anarva.internal',
        createdAt: new Date(Date.now() - 86400000).toISOString(),
        lastUsedAt: new Date().toISOString(),
      },
      {
        id: 'ank-102',
        name: 'GitHub Actions CI/CD Deployer Key',
        keyPrefix: 'ank_live_8f3c...',
        status: 'ACTIVE',
        permissions: ['compute.create', 'compute.update', 'database.create', 'provisioning.create'],
        createdBy: 'operator@anarva.internal',
        createdAt: new Date(Date.now() - 172800000).toISOString(),
        lastUsedAt: new Date(Date.now() - 3600000).toISOString(),
      },
      {
        id: 'ank-103',
        name: 'Staging Environment Test Key',
        keyPrefix: 'ank_test_4b19...',
        status: 'ACTIVE',
        permissions: ['compute.read', 'database.read', 'storage.read'],
        createdBy: 'operator@anarva.internal',
        createdAt: new Date(Date.now() - 259200000).toISOString(),
        lastUsedAt: new Date(Date.now() - 86400000).toISOString(),
      },
    ],
    meta: {
      requestId: `req_${Date.now()}`,
    },
  })
}

export async function POST(request: Request) {
  let body: any = {}
  try {
    body = await request.json()
  } catch (e) {}

  const isLive = body.isLive !== false
  const prefix = isLive ? 'ank_live_' : 'ank_test_'
  const randomHex = Math.random().toString(36).substring(2) + Math.random().toString(36).substring(2)
  const secretKey = `${prefix}${randomHex}`
  const keyPrefix = `${prefix}${randomHex.substring(0, 4)}...`

  const apiKey = {
    id: `ank-${Date.now()}`,
    name: body.name || 'New API Key',
    keyPrefix,
    status: 'ACTIVE',
    permissions: body.permissions || ['compute.read', 'compute.create', 'database.read', 'storage.read', 'network.read'],
    createdBy: 'user@anarva.io',
    createdAt: new Date().toISOString(),
  }

  return NextResponse.json({
    data: {
      apiKey,
      secretKey,
      warning: 'Store this secret key securely. It will never be displayed again.',
    },
    meta: {
      requestId: `req_${Date.now()}`,
    },
  })
}
