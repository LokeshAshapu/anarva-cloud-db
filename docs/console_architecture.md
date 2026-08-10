# Anarva Cloud Console Architecture Documentation

## 1. System Overview

The **Anarva Cloud Console** is an enterprise-grade control plane interface for managing cloud infrastructure resources including Databases, Object Storage (AOS), Compute Engine (ACE), Networking (VPC), IAM, Observability, Backups (PITR), Developer Tools, and Billing.

---

## 2. Resource Model (`web/types/resource.ts`)

Every cloud infrastructure object implements the standardized `CloudResource` interface:

```typescript
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
```

Specific resource models (`DatabaseResource`, `StorageResource`, `ComputeResource`, `NetworkResource`, `BackupResource`) extend this base model.

---

## 3. Design Tokens & Component Library (`web/components/cloud/`)

Design tokens are defined in `web/lib/tokens.ts` using the Anarva visual identity:
- **Color Palette**: Dark Navy (`#020617`), Slate (`#0f172a`), Electric Blue (`#3b82f6`), Violet (`#8b5cf6`), Cyan (`#06b6d4`), Emerald (`#10b981`).
- **Reusable Component Suite**:
  - `CloudButton`, `CloudCard`, `CloudMetric`, `CloudTable`, `CloudBadge`, `CloudStatus`
  - `CloudModal`, `CloudDrawer`, `CloudInput`, `CloudSelect`, `CloudTabs`, `CloudBreadcrumb`
  - `CloudSearch`, `CloudAlert`, `CloudEmptyState`, `CloudSkeleton`, `CloudChart`, `CloudPageHeader`
  - `CloudResourceCard`, `CloudResourceStatus`, `CloudResourceCreationWizard`

---

## 4. Global Shell & Routing

- `/console` / `/console/home` — Infrastructure Overview Dashboard
- `/console/compute` — Anarva Compute Engine (ACE)
- `/console/databases` — Managed Databases & SQL IDE
- `/console/storage` — Anarva Object Storage (AOS)
- `/console/networking` — Virtual Private Cloud (VPC)
- `/console/security` — Security & Audit Stream
- `/console/monitoring` — Observability & Time-Series Metrics
- `/console/backups` — Snapshots & Point-in-Time Recovery
- `/console/billing` — Usage Metering & Cost Analytics
- `/console/iam` — Identity, Team Roles & JSON Policies
- `/console/developer` / `/console/devtools` — API Keys, CLI & SDK Docs
- `/console/settings` — Platform Preferences & Default Region Selector

---

## 5. Region Infrastructure Mapping

Primary supported regions:
- `ap-south-2` — Asia Pacific (Hyderabad) [Available]
- `ap-south-1` — Asia Pacific (Mumbai) [Available]
- `ap-southeast-1` — Asia Pacific (Singapore) [Available]
- `us-east-1` — US East (N. Virginia) [Available]
- `eu-west-1` — Europe West (Frankfurt) [Available]

---

## 6. Future Integration Points

- **Data Plane Connectors**: Future integration with Docker Engine API, Kubernetes CRDs, and bare-metal provisioners via `pkg/providers/`.
- **Backend Search Integration**: Global Command Palette (`⌘K`) hook points for server-side resource indexing.
