# ANARVA Cloud V1 — Control Plane vs Data Plane Architecture Audit

**Audit Date**: August 19, 2026  
**Auditor**: Multi-Tenant SaaS Architect & Distributed Systems Lead  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  

---

## 1. Architectural Boundary Definitions

In cloud infrastructure engineering, a strict separation must exist between the **Control Plane** and the **Data Plane**:

- **Control Plane**: Manages identity, tenancy, billing, metadata, resource configuration, orchestration, policies, and API routing. It never touches raw customer payload traffic or customer data directly.
- **Data Plane**: Executes the actual customer workloads (storing PostgreSQL table rows, executing queries, handling S3 object data, routing live network packets).

---

## 2. Subsystem Control Plane / Data Plane Boundary Matrix

| Subsystem / Capability | Control Plane Component | Data Plane Component | Current Classification | Architectural Health |
|:---|:---|:---|:---|:---|
| **Users & Auth** | `users`, `sessions`, `api_keys` in Postgres DB | N/A | `CONTROL_PLANE` | 🟢 HEALTHY |
| **Organizations & Projects** | `organizations`, `projects` in Postgres DB | N/A | `CONTROL_PLANE` | 🟢 HEALTHY |
| **Database Instance Metadata** | `database_instances` catalog table | Customer PostgreSQL / MySQL engine | `BOTH` | 🟡 PARTIAL (Engine runs locally) |
| **SQL Console Query Engine** | `POST /api/v1/databases/{id}/query` | Target PostgreSQL instance connection | `DATA_PLANE` | 🟢 HEALTHY |
| **Managed Object Storage** | `buckets` metadata in storage service | `./data/storage` local server disk | `BOTH` | 🔴 VIOLATION (Payload stored on API server disk) |
| **VPC & Subnets** | `virtual_networks`, `subnets` in Postgres DB | Virtual interface / packet router | `CONTROL_PLANE` | 🟢 HEALTHY |
| **Load Balancer Rules** | In-memory `map[string]*LoadBalancer` | Local Docker NGINX container | `BOTH` | 🟡 PARTIAL (In-memory metadata) |
| **Compute Workloads** | In-memory compute repository | Local Docker containers | `BOTH` | 🟡 PARTIAL (In-memory metadata) |
| **Backups & Recovery** | `backup_records` in Postgres DB | `./data/backups/` dump archive files | `BOTH` | 🔴 VIOLATION (Dumps stored on API server disk) |
| **Reliability & Locks** | `resource_lock_leases`, `anarva_operations` | Recovery worker background loop | `CONTROL_PLANE` | 🟢 HEALTHY |
| **Observability & Metrics** | `audit_logs` in Postgres DB, `/metrics` | Prometheus exporter | `CONTROL_PLANE` | 🟢 HEALTHY |

---

## 3. Critical Architectural Boundary Violations Identified

### Violation 1: Direct File Object Data Plane on Control Plane Disk
- **Location**: `internal/storage/provider/local_storage_provider.go`
- **Issue**: The API server process writes customer uploaded raw file objects directly to `./data/storage` on the local server filesystem.
- **Risk**: In a containerized cloud deployment (e.g. Render Web Services), the API server container disk is ephemeral. Storing customer payload data on the control-plane server filesystem causes immediate data loss when the container restarts or redeploys.
- **Remediation**: The control plane must delegate object uploads to an isolated external S3 / Cloudflare R2 bucket.

### Violation 2: Direct Backup Archive File Storage on Control Plane Disk
- **Location**: `internal/backup/repository/backup_repository.go`
- **Issue**: Database backup archives are written to `./data/backups/` on the API server local disk.
- **Risk**: Backup dump archives disassociate from metadata upon API container redeployment.
- **Remediation**: Stream database dump streams directly to remote persistent object storage.

---

## 4. Summary Matrix of Plane Scoping

- **Pure Control-Plane Subsystems**: `12`
- **Data-Plane Subsystems**: `2` (SQL Execution, Docker Volume Payload)
- **Hybrid Control/Data Subsystems**: `6`
- **Architectural Boundary Violations**: `2` (Object Storage disk writes, Backup archive disk writes)
