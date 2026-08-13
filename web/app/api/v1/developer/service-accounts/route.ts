import { NextResponse } from 'next/server'

export async function GET(request: Request) {
  return NextResponse.json({
    data: [
      {
        id: 'sa-101',
        name: 'GitHub Actions CI/CD Deployer',
        description: 'Automated deployment service account for GitHub repository',
        status: 'ACTIVE',
        role: 'ADMIN',
        createdBy: 'lokeshashapu@gmail.com',
        createdAt: new Date(Date.now() - 172800000).toISOString(),
      },
      {
        id: 'sa-102',
        name: 'Kubernetes Worker Node Autoscaler',
        description: 'Managed ACU capacity autoscaling automation identity',
        status: 'ACTIVE',
        role: 'DEVELOPER',
        createdBy: 'lokeshashapu@gmail.com',
        createdAt: new Date(Date.now() - 86400000).toISOString(),
      },
      {
        id: 'sa-103',
        name: 'Prometheus Telemetry Collector',
        description: 'Read-only metrics and observability collector service account',
        status: 'ACTIVE',
        role: 'AUDITOR',
        createdBy: 'lokeshashapu@gmail.com',
        createdAt: new Date(Date.now() - 43200000).toISOString(),
      },
    ],
    meta: {
      requestId: `req_${Date.now()}`,
    },
  })
}
