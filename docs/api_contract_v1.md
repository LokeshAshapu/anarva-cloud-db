# ANARVA CLOUD PLATFORM — PUBLIC API CONTRACT v1

This document outlines the authoritative REST API contract for **Anarva Cloud Platform API (v1)** consumed by the official Anarva CLI (`anarva`) and TypeScript SDK (`@anarva/sdk`).

---

## 🔑 Authentication & Authorization

- **Header**: `Authorization: Bearer <API_KEY>`
- **API Key Formats**: `anarva_live_<secret>` / `anarva_test_<secret>`
- **Error Status**: `401 Unauthorized` for invalid, missing, or revoked API keys.
- **Tenant Isolation**: Resolved server-side from API Key context. Cross-organization access attempts return `404 Not Found` or `403 Forbidden`.

---

## 📡 Endpoints Inventory

### 1. System Health
- **GET** `/api/v1/health`
  - **Auth Required**: No
  - **Response**: `{"status": "OK", "version": "v1.0.0", "timestamp": "..."}`

### 2. Organizations
- **GET** `/api/v1/organizations`
  - **Auth Required**: Yes (`anarva_live_...`)
  - **Response**: `{"data": [{"id": "org-default", "name": "...", "slug": "...", "status": "ACTIVE"}], "requestId": "..."}`
- **GET** `/api/v1/organizations/:id`
  - **Auth Required**: Yes

### 3. Projects
- **GET** `/api/v1/projects`
  - **Auth Required**: Yes
  - **Response**: `{"data": [{"id": "proj-default", "organizationId": "org-default", "name": "...", "slug": "..."}]}`

### 4. Cloud Resources Observability
- **GET** `/api/v1/resources?resourceType=EC2|RDS|S3`
  - **Auth Required**: Yes
  - **Query Parameters**: `resourceType` (optional), `projectId` (optional)
  - **Response**: `{"data": [...], "requestId": "..."}`

### 5. Managed PostgreSQL Platform (RDS)
- **GET** `/api/v1/databases/:id`
  - **Auth Required**: Yes
  - **Permission Required**: `database.read`
- **POST** `/api/v1/databases`
  - **Auth Required**: Yes
  - **Permission Required**: `database.create`
  - **Body**: `{"name": "string", "projectId": "string", "storageGb": 25, "acuUnits": 2.0, "multiAz": true}`
- **POST** `/api/v1/databases/:id/failover`
  - **Auth Required**: Yes
  - **Permission Required**: `database.failover`
  - **Response**: `{"data": {"id": "job-failover-123", "status": "COMPLETED", "previousPrimaryAz": "ap-south-1a", "newPrimaryAz": "ap-south-1b"}}`

### 6. Backups & Snapshot Management
- **GET** `/api/v1/backups?resourceId=:id`
  - **Auth Required**: Yes
- **POST** `/api/v1/backups`
  - **Auth Required**: Yes
  - **Body**: `{"resourceId": "string", "snapshotName": "string"}`

### 7. CloudWatch Infrastructure Metrics
- **GET** `/api/v1/metrics/:resourceId`
  - **Auth Required**: Yes
  - **Permission Required**: `metrics.read`
  - **Response**: `{"data": {"resourceId": "...", "source": "AWS CloudWatch", "cpuUtilization": "15.1%", "status": "OK"}}`

### 8. Billing Engine & Invoices
- **GET** `/api/v1/billing/invoices`
  - **Auth Required**: Yes
  - **Permission Required**: `billing.read`
  - **Response**: `{"data": [{"id": "inv-2026-08", "subtotalUsd": "43.49", "status": "FINALIZED"}]}`

### 9. Asynchronous Control-Plane Operations
- **GET** `/api/v1/operations/:id`
  - **Auth Required**: Yes
  - **Response**: `{"data": {"id": "op-101", "status": "COMPLETED", "resourceId": "anarva-rds-prod-01"}}`

### 10. Developer API Keys Management
- **GET** `/api/v1/developer/keys`
  - **Auth Required**: Yes (`iam.read` or `OWNER` / `ADMIN`)
- **POST** `/api/v1/developer/keys`
  - **Auth Required**: Yes (`iam.write`)
  - **Body**: `{"name": "string", "projectId": "string", "permissions": ["compute.read", "database.read"], "isLive": true}`
