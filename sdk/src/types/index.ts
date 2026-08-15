export type APIKeyStatus = 'ACTIVE' | 'REVOKED' | 'EXPIRED' | 'SUSPENDED';
export type APIKeyEnvironment = 'LIVE' | 'TEST';

export interface APIKey {
  id: string;
  organizationId: string;
  projectId: string;
  name: string;
  keyPrefix: string;
  environment: APIKeyEnvironment;
  status: APIKeyStatus;
  permissions: string[];
  createdBy: string;
  createdAt: string;
  expiresAt?: string;
  lastUsedAt?: string;
  revokedAt?: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  status: 'ACTIVE' | 'SUSPENDED';
  createdAt: string;
}

export interface Project {
  id: string;
  organizationId: string;
  name: string;
  slug: string;
  region: string;
  status: 'ACTIVE' | 'SUSPENDED';
  createdAt: string;
}

export interface ComputeInstance {
  id: string;
  organizationId: string;
  projectId: string;
  name: string;
  instanceType: string;
  status: 'RUNNING' | 'STOPPED' | 'TERMINATED';
  acuUnits: number;
  regionId: string;
  createdAt: string;
}

export interface CreateComputeParams {
  name: string;
  projectId: string;
  acuUnits: number;
  regionId?: string;
}

export interface DatabaseInstance {
  id: string;
  organizationId: string;
  projectId: string;
  name: string;
  engine: 'POSTGRESQL';
  engineVersion: string;
  status: 'AVAILABLE' | 'MODIFYING' | 'REBOOTING' | 'FAILING_OVER' | 'STOPPED';
  port: number;
  cpuUnits: number;
  memoryMb: number;
  storageGb: number;
  multiAz: boolean;
  primaryAz: string;
  secondaryAz?: string;
  publiclyAccessible: boolean;
  createdAt: string;
}

export interface CreateDatabaseParams {
  name: string;
  projectId: string;
  engine?: 'POSTGRESQL';
  version?: string;
  storageGb?: number;
  acuUnits?: number;
  multiAz?: boolean;
}

export interface DatabaseBackup {
  id: string;
  organizationId: string;
  projectId: string;
  resourceId: string;
  snapshotName: string;
  status: 'AVAILABLE' | 'CREATING' | 'DELETING';
  sizeGb: number;
  createdAt: string;
}

export interface RecoveryWindowInfo {
  resourceId: string;
  earliestRestorableTime: string;
  latestRestorableTime: string;
  pitrEnabled: boolean;
}

export interface MultiAZConfig {
  resourceId: string;
  organizationId: string;
  projectId: string;
  multiAz: boolean;
  primaryAvailabilityZone: string;
  secondaryAvailabilityZone?: string;
  haStatus: 'HA_ENABLED' | 'HA_DISABLED' | 'HA_MODIFYING' | 'HA_DEGRADED';
  updatedAt: string;
}

export interface FailoverJob {
  id: string;
  resourceId: string;
  status: 'FAILOVER_INITIATED' | 'PRIMARY_CHANGED' | 'COMPLETED' | 'FAILED';
  previousPrimaryAz: string;
  newPrimaryAz: string;
  requestedBy: string;
  requestedAt: string;
  completedAt?: string;
}

export interface StorageBucket {
  id: string;
  organizationId: string;
  projectId: string;
  name: string;
  regionId: string;
  encryption: 'SSE-S3';
  publicAccessBlock: boolean;
  createdAt: string;
}

export interface CloudWatchMetrics {
  resourceId: string;
  source: 'AWS CloudWatch';
  cpuUtilization?: string;
  networkInBytes?: string;
  networkOutBytes?: string;
  databaseConnections?: number;
  freeStorageSpaceBytes?: string;
  status: 'OK' | 'NO_DATA' | 'NOT_CONFIGURED';
  lastUpdated: string;
}

export interface BillingInvoice {
  id: string;
  organizationId: string;
  billingPeriod: string;
  subtotalUsd: string;
  taxUsd: string;
  totalUsd: string;
  status: 'DRAFT' | 'FINALIZED' | 'PAID';
  pricingVersion: string;
  createdAt: string;
}

export interface ControlPlaneOperation {
  id: string;
  organizationId: string;
  projectId: string;
  resourceId: string;
  type: string;
  status: 'QUEUED' | 'IN_PROGRESS' | 'COMPLETED' | 'FAILED';
  failureReason?: string;
  createdAt: string;
  completedAt?: string;
}
