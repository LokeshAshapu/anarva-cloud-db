import { NextResponse } from 'next/server'

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const orgId = searchParams.get('organizationId') || 'org-default'
  const projectId = searchParams.get('projectId') || 'proj-default'

  return NextResponse.json({
    data: [
      {
        id: 'res-db-prod-1',
        organizationId: orgId,
        projectId: projectId,
        name: 'production-db',
        provider: 'LOCAL_POSTGRES',
        version: '17.2',
        status: 'AVAILABLE',
        regionId: 'ap-hyderabad-1',
        cpu: 2.0,
        memoryMb: 2048,
        storageGb: 48,
        networkId: 'vpc-net-1',
        availabilityMode: 'SINGLE',
        host: 'localhost',
        port: 5432,
        publicAccess: false,
        realityLabel: 'LOCAL_POSTGRES (DOCKER_SIM)',
        createdAt: new Date().toISOString(),
      },
    ],
    meta: {
      count: 1,
      requestId: `req_${Date.now()}`,
    },
  })
}

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const newDb = {
      id: `db_${Date.now()}`,
      organizationId: body.organizationId || 'org-default',
      projectId: body.projectId || 'proj-default',
      name: body.name || 'new-postgres-db',
      provider: 'LOCAL_POSTGRES',
      version: body.version || '17.2',
      status: 'AVAILABLE',
      regionId: body.regionId || 'ap-hyderabad-1',
      cpu: body.cpu || 2.0,
      memoryMb: body.memoryMb || 2048,
      storageGb: body.storageGb || 25,
      networkId: body.networkId || 'vpc-net-1',
      availabilityMode: 'SINGLE',
      host: 'localhost',
      port: 5433,
      publicAccess: body.publicAccess || false,
      realityLabel: 'LOCAL_POSTGRES (DOCKER_SIM)',
      createdAt: new Date().toISOString(),
    }
    return NextResponse.json({ data: newDb }, { status: 201 })
  } catch (error: any) {
    return NextResponse.json({ error: { message: error.message } }, { status: 400 })
  }
}
