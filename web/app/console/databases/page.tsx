'use client'

import React, { useState } from 'react'
import { CloudResource } from '@/types/resource'
import { CloudResourceList } from '@/components/cloud/CloudResourceList'
import { CloudResourceDetail } from '@/components/cloud/CloudResourceDetail'
import { CloudResourceCreationWizard } from '@/components/cloud/CloudResourceCreationWizard'
import { generateARNV } from '@/lib/arnv'

export default function DatabasesPage() {
  const [selectedResource, setSelectedResource] = useState<CloudResource | null>(null)
  const [isWizardOpen, setIsWizardOpen] = useState(false)

  const [dbResources, setDbResources] = useState<CloudResource[]>([
    {
      id: 'res-db-prod-1',
      resourceId: 'arnv:db:ap-hyderabad-1:proj-default:database/production-db',
      name: 'production-db',
      type: 'DATABASE',
      status: 'AVAILABLE',
      organizationId: 'org-default',
      projectId: 'proj-default',
      environment: 'Production',
      regionId: 'ap-hyderabad-1',
      ownerId: 'lokeshashapu@gmail.com',
      tags: [{ key: 'Environment', value: 'Production' }, { key: 'Team', value: 'Engineering' }],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
    {
      id: 'res-db-analytics-1',
      resourceId: 'arnv:db:ap-mumbai-1:proj-default:database/analytics-db',
      name: 'analytics-db',
      type: 'DATABASE',
      status: 'AVAILABLE',
      organizationId: 'org-default',
      projectId: 'proj-default',
      environment: 'Production',
      regionId: 'ap-mumbai-1',
      ownerId: 'lokeshashapu@gmail.com',
      tags: [{ key: 'Environment', value: 'Production' }, { key: 'Application', value: 'Analytics' }],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
  ])

  const handleCreateComplete = (data: any) => {
    const newRes: CloudResource = {
      id: `res-db-${Date.now()}`,
      resourceId: generateARNV('DATABASE', data.region, 'proj-default', data.name || 'new-database'),
      name: data.name || 'new-database',
      type: 'DATABASE',
      status: 'AVAILABLE',
      organizationId: 'org-default',
      projectId: 'proj-default',
      environment: 'Production',
      regionId: data.region,
      ownerId: 'lokeshashapu@gmail.com',
      tags: [{ key: 'Environment', value: 'Production' }],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    setDbResources([newRes, ...dbResources])
    setIsWizardOpen(false)
  }

  const handleDelete = (id: string) => {
    setDbResources(dbResources.filter((r) => r.id !== id))
  }

  if (selectedResource) {
    return <CloudResourceDetail resource={selectedResource} onBack={() => setSelectedResource(null)} />
  }

  if (isWizardOpen) {
    return (
      <div className="py-8">
        <CloudResourceCreationWizard
          title="Provision Managed Database Cluster"
          resourceType="DATABASE"
          onComplete={handleCreateComplete}
          onCancel={() => setIsWizardOpen(false)}
        />
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="border-b border-slate-800 pb-5">
        <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Managed Databases</h1>
        <p className="text-slate-400 text-xs sm:text-sm mt-1">
          High-performance distributed database clusters (PostgreSQL / MySQL) with automated failover and encryption.
        </p>
      </div>

      <CloudResourceList
        title="Managed Database Registry"
        resources={dbResources}
        onCreateResource={() => setIsWizardOpen(true)}
        onDeleteResource={handleDelete}
        onViewResource={(res) => setSelectedResource(res)}
      />
    </div>
  )
}
