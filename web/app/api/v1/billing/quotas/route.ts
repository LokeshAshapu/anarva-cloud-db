import { NextResponse } from 'next/server'

export async function GET(request: Request) {
  return NextResponse.json({
    data: [
      {
        id: 'q-acu-101',
        organizationId: 'org-default',
        projectId: 'proj-default',
        resourceType: 'COMPUTE',
        metric: 'compute.acu',
        limit: 8.0,
        currentUsage: 1.0,
        unit: 'ACU',
        period: 'PERPETUAL',
        status: 'AVAILABLE',
      },
      {
        id: 'q-sto-102',
        organizationId: 'org-default',
        projectId: 'proj-default',
        resourceType: 'STORAGE',
        metric: 'storage.capacity',
        limit: 25.0,
        currentUsage: 2.5,
        unit: 'GB',
        period: 'PERPETUAL',
        status: 'AVAILABLE',
      },
      {
        id: 'q-db-103',
        organizationId: 'org-default',
        projectId: 'proj-default',
        resourceType: 'DATABASE',
        metric: 'database.count',
        limit: 2.0,
        currentUsage: 1.0,
        unit: 'INSTANCES',
        period: 'PERPETUAL',
        status: 'AVAILABLE',
      },
      {
        id: 'q-net-104',
        organizationId: 'org-default',
        projectId: 'proj-default',
        resourceType: 'NETWORK',
        metric: 'network.vpc',
        limit: 2.0,
        currentUsage: 1.0,
        unit: 'NETWORKS',
        period: 'PERPETUAL',
        status: 'AVAILABLE',
      },
    ],
    meta: {
      requestId: `req_${Date.now()}`,
    },
  })
}
