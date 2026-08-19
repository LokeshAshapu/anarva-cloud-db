# ANARVA Cloud V1 — Control-Plane Persistence Migration Matrix

**Audit Date**: August 19, 2026  
**Auditor**: Database Architect & Storage Reliability Lead  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  

---

## 1. Persistence Target & Rule Definition

**ABSOLUTE RULE**: Production control-plane state must **NOT** depend on in-memory maps, local JSON files, local server filesystem (`./data`), or Render ephemeral disk.

---

## 2. Resource Persistence Audit & Migration Schedule

| Resource Type | Current Storage Engine | Production Safe? | Target Production Storage | Migration Required? | Priority | Implementation Strategy |
|:---|:---|:---:|:---|:---:|:---:|:---|
| **Users & Auth** | PostgreSQL GORM / Local JSON | ✅ YES (Prod DB) | Managed PostgreSQL | No | Done | Controlled via `DATABASE_URL` |
| **Sessions** | PostgreSQL GORM / Local JSON | ✅ YES (Prod DB) | Managed PostgreSQL | No | Done | Controlled via `DATABASE_URL` |
| **Organizations** | PostgreSQL GORM / Local JSON | ✅ YES (Prod DB) | Managed PostgreSQL | No | Done | Controlled via `DATABASE_URL` |
| **Projects** | PostgreSQL GORM / Local JSON | ✅ YES (Prod DB) | Managed PostgreSQL | No | Done | Controlled via `DATABASE_URL` |
| **Database Instance Catalog** | PostgreSQL GORM / Local JSON | ✅ YES (Prod DB) | Managed PostgreSQL | No | Done | Controlled via `DATABASE_URL` |
| **Virtual Networks** | PostgreSQL GORM | ✅ YES (Prod DB) | Managed PostgreSQL | No | Done | AutoMigrated in `cmd/gateway/main.go` |
| **Subnets & Security Groups**| PostgreSQL GORM | ✅ YES (Prod DB) | Managed PostgreSQL | No | Done | AutoMigrated in `cmd/gateway/main.go` |
| **Operations & Lock Leases** | PostgreSQL GORM | ✅ YES (Prod DB) | Managed PostgreSQL | No | Done | AutoMigrated in `cmd/gateway/main.go` |
| **Audit Logs** | PostgreSQL GORM | ✅ YES (Prod DB) | Managed PostgreSQL | No | Done | AutoMigrated in `cmd/gateway/main.go` |
| **Object Storage Payloads** | Local Disk (`./data/storage`) | ❌ NO | AWS S3 / Cloudflare R2 | **YES** | **P0 (CRITICAL)** | Upgrade `LocalStorageProvider` to S3 Driver |
| **Backup Dump Archives** | Local Disk (`./data/backups/`) | ❌ NO | AWS S3 / Cloudflare R2 | **YES** | **P0 (CRITICAL)** | Stream dumps to S3 Bucket Driver |
| **Compute Instance Metadata**| In-Memory Map (`newMemComputeRepo`) | ❌ NO | Managed PostgreSQL | **YES** | **P1 (HIGH)** | Create `GORM` Compute Repository |
| **Load Balancer Rules** | In-Memory Map (`lb_repository.go`) | ❌ NO | Managed PostgreSQL | **YES** | **P1 (HIGH)** | Create `GORM` Load Balancer Repository |
| **MySQL Instance Metadata** | In-Memory Map (`mysql_repository.go`) | ❌ NO | Managed PostgreSQL | **YES** | **P2 (MEDIUM)** | Create `GORM` MySQL Instance Repository |
| **Webhook Subscriptions** | Gateway Memory (`webhook_usecase.go`) | ❌ NO | Managed PostgreSQL | **YES** | **P2 (MEDIUM)** | Create `GORM` Webhook Repository |

---

## 3. Migration Execution Plan

1. **P0 (Object Storage & Backup Payload Ephemeral Disk Elimination)**:
   - Implement S3/R2 SDK driver (`internal/storage/provider/s3_storage_provider.go`) to replace local server filesystem writes (`./data/storage`).
   - Stream database dump streams directly to S3 storage bucket instead of local server disk (`./data/backups`).
2. **P1 (Compute & Load Balancer In-Memory Model Migration)**:
   - Create `GORM` schema models and GORM repositories for `compute_instances` and `load_balancers`.
   - Wire GORM repositories in `cmd/gateway/main.go` when `dbPool != nil`.
