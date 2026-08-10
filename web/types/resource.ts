// Anarva Cloud Platform — Phase 4 Resource Hierarchy & Registry Model

export type EnvironmentType = 'Development' | 'Staging' | 'Production'

export type RegionId =
  | 'ap-hyderabad-1'
  | 'ap-mumbai-1'
  | 'ap-singapore-1'
  | 'us-east-1'
  | 'eu-west-1'

export type ResourceStatus =
  | 'CREATING'
  | 'AVAILABLE'
  | 'UPDATING'
  | 'DELETING'
  | 'DELETED'
  | 'FAILED'
  | 'STOPPED'
  | 'COMING_SOON'
  | 'MAINTENANCE'

export type ResourceType =
  | 'DATABASE'
  | 'STORAGE_BUCKET'
  | 'COMPUTE'
  | 'NETWORK'
  | 'BACKUP'
  | 'REPLICA'

export interface ResourceTag {
  key: string
  value: string
}

export interface Organization {
  id: string
  name: string
  slug: string
  ownerId: string
  createdAt: string
  updatedAt: string
}

export interface Project {
  id: string
  organizationId: string
  name: string
  slug: string
  description: string
  environment: EnvironmentType
  defaultRegion: RegionId
  createdAt: string
  updatedAt: string
}

export interface Region {
  id: RegionId
  name: string
  displayName: string
  location: string
  status: 'AVAILABLE' | 'COMING_SOON' | 'MAINTENANCE'
}

export interface CloudResource {
  id: string
  resourceId: string // ARNV string e.g. arnv:db:ap-hyderabad-1:proj-default:database/production-db
  name: string
  type: ResourceType
  status: ResourceStatus
  organizationId: string
  projectId: string
  environment: EnvironmentType
  regionId: RegionId
  ownerId: string
  tags: ResourceTag[]
  createdAt: string
  updatedAt: string
}

export interface ActivityEvent {
  id: string
  organizationId: string
  projectId: string
  resourceId?: string
  actorId: string
  action:
    | 'RESOURCE_CREATED'
    | 'RESOURCE_UPDATED'
    | 'RESOURCE_DELETED'
    | 'RESOURCE_STARTED'
    | 'RESOURCE_STOPPED'
    | 'RESOURCE_CONFIGURATION_CHANGED'
    | 'USER_LOGIN'
    | 'API_KEY_CREATED'
    | 'BACKUP_CREATED'
  timestamp: string
  metadata?: Record<string, string>
}
