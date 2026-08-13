import { NextResponse } from 'next/server'

export async function GET(request: Request, { params }: { params: { id: string } }) {
  const id = params.id
  return NextResponse.json({
    data: {
      id,
      organizationId: 'org-default',
      projectId: 'proj-default',
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
  })
}

export async function DELETE(request: Request, { params }: { params: { id: string } }) {
  const id = params.id
  return NextResponse.json({
    status: 'DELETED',
    id,
  })
}
