# ANARVA Cloud V1 — System Architecture & Design Specification

**Phase**: Architecture Hardening Audit  
**Audit Date**: August 19, 2026  
**Auditor**: Principal Cloud Architect, Go Backend Engineer & Security Lead  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  
**Production API**: `https://anarva-cloud-db-api.onrender.com`  

---

## 1. Executive System Architecture

The ANARVA Cloud platform is architected as an API gateway and monolithic control-plane router managing multi-tenant cloud resources, multi-engine databases, virtual networks, object storage, telemetry, and distributed operations.

```
                           [ ANARVA Console / CLI / SDK ]
                                         │
                                         ▼
                            [ Go API Gateway (8080) ]
                                         │
              ┌──────────────────────────┼──────────────────────────┐
              ▼                          ▼                          ▼
    [ Auth Middleware ]        [ TenantContext Scoping ]     [ Reliability & Locks ]
              │                          │                          │
              └──────────────────────────┼──────────────────────────┘
                                         │
                      ┌──────────────────┴──────────────────┐
                      ▼                                     ▼
         [ PostgreSQL GORM Engine ]               [ Pluggable Data Plane ]
       (Control Plane Metadata)               (S3 / RDS / Docker Hosts)
```

---

## 2. Control Plane vs Data Plane Separation

### Control Plane (ANARVA API Gateway + Control Database):
- Manages users, authentication, sessions, organizations, projects, IAM/RBAC, API keys, database instance catalog, network topologies, provider mappings, operation lifecycle states, audit logs, billing metadata, and resource lock leases.
- Persisted in managed PostgreSQL via GORM when `DATABASE_URL` is set.

### Data Plane (Managed Workloads & External Providers):
- Executes actual customer workloads: PostgreSQL/MySQL database table rows, S3 object storage payloads, backup dump archives, compute container instances, live network packets, and load balancer rules.
- Must be decoupled from the control plane so that destroying or restarting an API server container never causes customer data payload loss.

---

## 3. Core Architectural Subsystems

### 1. Authentication & Security Engine
- JWT Bearer Tokens (HMAC SHA-256) and SHA-256 API Key verification.
- Enforces `TenantContext` injection via `internal/gateway/middleware/auth_middleware.go`.

### 2. Multi-Tenant Resource Scoping
- All primary resource queries scope through `organization_id` and `project_id`.
- Rejects unauthenticated or cross-tenant resource manipulation server-side (`HTTP 401` / `HTTP 403`).

### 3. Distributed Reliability & Operation Worker
- Operation lifecycle states (`REQUESTED`, `QUEUED`, `PROVISIONING`, `READY`, `FAILED`, `RECOVERING`).
- Resource lock leases (`resource_lock_leases`) prevent concurrent modification race conditions.
- Background `RecoveryWorker` (`internal/reliability/usecase/recovery_worker.go`) automatically reconciles stale operations.

### 4. Pluggable Infrastructure Provider Abstraction
- Defined interfaces for `DatabaseProvider`, `StorageProvider`, `NetworkProvider`, `LoadBalancerProvider`, and `BackupProvider`.
- Enables clean separation between local development implementations (`LocalDockerPostgresProvider`, `LocalStorageProvider`) and cloud implementations (`AWSDatabaseProvider`, `S3StorageProvider`).

---

## 4. Production Deployment & Render Compatibility Strategy

1. **Fail-Closed Enforcement**: In `ANARVA_ENV=production`, the Go gateway requires a valid `DATABASE_URL` pointing to managed PostgreSQL. If missing or resolving to `localhost`, startup fails immediately (`log.Fatal`).
2. **Commit & Build Versioning**: `/api/v1/health` and `/api/v1/health/persistence` expose `build` metadata (`gitCommit: "phase-60-cd8ca2a"`), providing transparent deployment verification.
