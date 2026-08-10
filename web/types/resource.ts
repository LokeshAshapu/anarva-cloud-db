// Anarva Cloud Platform — Resource Model Abstraction

export type ResourceStatus =
  | 'AVAILABLE'
  | 'PROVISIONING'
  | 'DEGRADED'
  | 'STOPPED'
  | 'FAILED'
  | 'TERMINATED'
  | 'MAINTENANCE'

export type CloudRegion =
  | 'ap-south-2' // Asia Pacific — Hyderabad
  | 'ap-south-1' // Asia Pacific — Mumbai
  | 'ap-southeast-1' // Asia Pacific — Singapore
  | 'us-east-1' // US East — N. Virginia
  | 'eu-west-1' // Europe West — Frankfurt

export interface ResourceTag {
  key: string
  value: string
}

export interface CloudResource {
  id: string
  name: string
  type: 'DATABASE' | 'STORAGE' | 'COMPUTE' | 'NETWORK' | 'BACKUP'
  status: ResourceStatus
  region: CloudRegion
  projectId: string
  ownerId: string
  createdAt: string
  updatedAt: string
  tags?: ResourceTag[]
}

export interface DatabaseResource extends CloudResource {
  type: 'DATABASE'
  engine: 'postgres' | 'mysql'
  version: string
  acuAllocated: number
  storageSizeGb: number
  multiAZ: boolean
  connectionUri: string
}

export interface StorageResource extends CloudResource {
  type: 'STORAGE'
  bucketName: string
  objectCount: number
  totalSizeBytes: number
  isPublic: boolean
  versioning: boolean
}

export interface ComputeResource extends CloudResource {
  type: 'COMPUTE'
  acu: number
  vCPU: number
  memoryGb: number
  availabilityZone: string
  publicIp?: string
}

export interface NetworkResource extends CloudResource {
  type: 'NETWORK'
  cidrBlock: string
  subnetCount: number
  securityRulesCount: number
}

export interface BackupResource extends CloudResource {
  type: 'BACKUP'
  targetResourceId: string
  sizeBytes: number
  backupType: 'AUTOMATED' | 'MANUAL'
  retentionDays: number
}
