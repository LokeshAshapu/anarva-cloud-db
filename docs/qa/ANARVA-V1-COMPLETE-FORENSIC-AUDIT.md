# ANARVA Cloud Control Plane — Comprehensive V1 Forensic Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 19, 2026  
**Auditor**: Principal Cloud Security Architect, Lead Database Engineer & SRE  
**Production Gateway**: `https://anarva-cloud-db-api.onrender.com`  
**Commit**: [`5d0956755c1a459bc628db2390b92c03d2ccb660`](https://github.com/LokeshAshapu/anarva-cloud-db/commit/5d0956755c1a459bc628db2390b92c03d2ccb660)  
**Audit Mode**: READ-ONLY FORENSIC INSPECTION  

---

## 1. Architectural Overview & System Design

The ANARVA Cloud Control Plane is built as a high-performance Go API gateway and monolithic router managing multi-tenant cloud resources, multi-engine databases, virtual networking, object storage, observability, and distributed reliability operations.

```
                           [ ANARVA Web Console / CLI / SDK ]
                                           │
                                           ▼
                                [ Go API Gateway (8080) ]
                                           │
               ┌───────────────────────────┼───────────────────────────┐
               ▼                           ▼                           ▼
     [ Auth & IAM Middleware ]   [ TenantContext Scoping ]   [ Reliability & Locks ]
               │                           │                           │
               └───────────────────────────┼───────────────────────────┘
                                           │
                        ┌──────────────────┴──────────────────┐
                        ▼                                     ▼
           [ PostgreSQL GORM Engine ]               [ Local Docker Provider ]
          (Persistent Control Plane)              (Local Engine Containers)
```

---

## 2. Deep Persistence Forensic Audit

### 2.1 DATABASE_URL Configuration Tracing

```
[Render Environment Variable] DATABASE_URL
  │
  ▼
[pkg/config/config.go] LoadConfig()
  ├─ Direct env lookup: rawURL := os.Getenv("DATABASE_URL")
  └─ Injects into cfg.Database.URL
  │
  ▼
[pkg/config/config.go] DatabaseConfig.DSN()
  ├─ Evaluates db.URL / DATABASE_URL
  └─ If ANARVA_ENV=production and URL missing: Returns empty string (Fail-Closed)
  │
  ▼
[cmd/gateway/main.go] Production Startup Assertion
  ├─ Asserts DSN is non-empty
  └─ Asserts Hostname != "localhost" / "127.0.0.1"
  │
  ▼
[pkg/database/postgres.go] NewPostgresDB()
  ├─ Opens connection via gorm.Open(postgres.Open(dsn))
  └─ Performs live PingContext & schema AutoMigrate
```

### 2.2 Persistence Boundary Matrix Summary

- **Control-Plane Entities** (Users, Sessions, API Keys, Orgs, Projects, Databases, Virtual Networks, Lock Leases, Operations): **PERSISTED IN POSTGRESQL**. They survive process restarts, machine reboots, container restarts, Render redeployments, and cross-client queries.
- **Object Storage & Backup Files**: Written to `./data/storage` and `./data/backups`. They survive local process restarts, but **DO NOT** survive Render container redeployments due to the ephemeral filesystem.
- **Compute & Load Balancers**: Held in gateway memory maps. They reset upon process restart.

---

## 3. Deep Multi-Tenant Data Isolation Audit

### 3.1 Tenant Context Flow

```
HTTP Request with JWT / API Key
  │
  ▼
[internal/gateway/middleware/auth_middleware.go]
  ├─ Validates JWT signature / API Key hash
  ├─ Extracts User ID & Tenant ID
  └─ Constructs TenantContext in Request context
  │
  ▼
[internal/database/delivery/http/database_handler.go]
  ├─ Retrieves TenantContext from r.Context()
  ├─ Passes User ID / Org ID to UseCase
  └─ Filters database queries: WHERE project_id IN (...) OR tenant_id = ...
```

### 3.2 Security Scoping Verification
All primary control-plane endpoints (`/api/v1/databases`, `/api/v1/projects`, `/api/v1/organizations`, `/api/v1/networking/vnets`) require authenticated JWT headers and enforce project/tenant scoping. Unauthenticated requests are rejected with `HTTP 401 AUTH_REQUIRED`.

---

## 4. Comprehensive Subsystem Breakdown

### 1. PostgreSQL Engine
- **Control-Plane Metadata**: Persisted in PostgreSQL DB via GORM.
- **Engine Provisioner**: `LocalDockerPostgresProvider` spins up Docker containers locally.
- **Reality**: `CONTROL_PLANE_ONLY` (Metadata) / `REAL_LOCAL` (Local Docker).

### 2. MySQL Engine
- Metadata stored in-memory; `LocalDockerMySQLProvider` manages local MySQL containers.
- **Reality**: `REAL_LOCAL` / `CONTROL_PLANE_ONLY`.

### 3. MongoDB & Redis Engines
- Metadata recorded in instance catalog; provisioner uses simulated drivers.
- **Reality**: `SIMULATED` / `CONTROL_PLANE_ONLY`.

### 4. Managed Object Storage
- S3-compatible API handlers; stores raw file objects on local filesystem (`./data/storage`).
- **Reality**: `DEVELOPMENT_ONLY` (Ephemeral local disk on Render).

### 5. VPC & Virtual Networking
- Network topology tables (`virtual_networks`, `subnets`, `security_groups`) persisted in PostgreSQL.
- **Reality**: `CONTROL_PLANE_ONLY`.

### 6. Load Balancer & Edge Delivery
- Routing rules and target health checks stored in-memory.
- **Reality**: `CONTROL_PLANE_ONLY` / `SIMULATED`.

### 7. Compute Platform
- Compute instance specifications tracked in-memory (`newMemComputeRepo`).
- **Reality**: `SIMULATED`.

### 8. Backup & Disaster Recovery
- Backup metadata saved to PostgreSQL; dump archives written to server disk (`./data/backups/`).
- **Reality**: `DEVELOPMENT_ONLY` (Artifact storage).

### 9. IAM & Security Engine
- JWT authentication, RBAC authorization service, and security status diagnostics.
- **Reality**: `CONTROL_PLANE_ONLY`.

### 10. Reliability & Operations Worker
- Distributed lock leases, idempotency records, and background recovery worker in PostgreSQL.
- **Reality**: `CONTROL_PLANE_ONLY`.

---

## 5. Master Capability Audit Table

| ID | Capability Name | Category | UI Route | API Route | Service / UseCase | Repository / Provider | Infrastructure Touched | Persistence | Isolation | Reality Classification | Production Readiness | Code Evidence | Known Limitations |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 01 | Auth & Sessions | Identity | `/login`, `/signup` | `/api/v1/auth/login` | `authUsecase` | `authRepo.UserRepository` | PostgreSQL / Disk JSON | PostgreSQL GORM | JWT Subject | `CONTROL_PLANE_ONLY` | READY | `auth_handler.go` | JSON fallback in dev mode |
| 02 | Orgs & Projects | Multi-Tenancy | `/console/projects` | `/api/v1/projects` | `projectUsecase` | `projectRepo.OrganizationRepository` | PostgreSQL | PostgreSQL GORM | Org Membership | `CONTROL_PLANE_ONLY` | READY | `project_usecase.go` | Metadata scoping |
| 03 | DB Catalog | Databases | `/console/databases` | `/api/v1/databases` | `databaseUsecase` | `databaseRepo.InstanceRepository` | PostgreSQL | PostgreSQL GORM | Project ID | `CONTROL_PLANE_ONLY` | READY | `database_handler.go` | Stores metadata |
| 04 | PostgreSQL Engine | Engine | `/console/databases/[id]` | `/api/v1/postgres/instances` | `postgresService` | `postgresProvider.DockerProvider` | Host Docker Daemon | Docker Volume | Container Port | `REAL_LOCAL` | NOT_CLOUD_READY | `docker_provider.go` | Requires local Docker |
| 05 | SQL Console | Query Engine | `/dashboard/query` | `/api/v1/databases/{id}/query` | `sqlService` | `sqlService.Execute` | Target Postgres DB | Target DB Engine | DB Credentials | `REAL_LOCAL` | READY_IF_CONNECTED | `sql_service.go` | Requires direct TCP access |
| 06 | MySQL Engine | Engine | `/console/databases` | `/api/v1/mysql/instances` | `mysqlService` | `mysqlProvider.DockerProvider` | Host Docker Daemon | Docker Volume | Container Port | `REAL_LOCAL` | NOT_CLOUD_READY | `mysql_service.go` | Local Docker dependency |
| 07 | Virtual Networks | Networking | `/console/networking` | `/api/v1/networking/vnets` | `networkingService` | `networkingRepo.PostgresRepository` | PostgreSQL | PostgreSQL GORM | Project ID | `CONTROL_PLANE_ONLY` | READY | `postgres_repository.go` | Persists topology in DB |
| 08 | Load Balancers | Edge | `/console/loadbalancers` | `/api/v1/loadbalancers` | `lbService` | `lbRepo.LoadBalancerRepository` | Memory Map | In-Memory | Project ID | `CONTROL_PLANE_ONLY` | EPHEMERAL | `lb_repository.go` | In-memory rules |
| 09 | Object Storage | Storage | `/console/storage` | `/api/v1/storage/buckets` | `storageService` | `storageProvider.LocalStorage` | Server Local Disk | Local Filesystem | Bucket Prefix | `DEVELOPMENT_ONLY` | NOT_CLOUD_READY | `local_storage_provider.go` | Ephemeral disk on Render |
| 10 | Compute Engine | Compute | `/console/compute` | `/api/v1/compute/instances` | `computeUsecase` | `newMemComputeRepo` | Gateway Memory | In-Memory | Project ID | `SIMULATED` | NOT_READY | `main.go:398` | In-memory instances |
| 11 | Control Backup | Backup | `/console/backups` | `/api/v1/backups` | `backupUsecase` | `backupRepo.BackupRepository` | PostgreSQL + Local Disk | Postgres + Disk Files | Project ID | `DEVELOPMENT_ONLY` | PARTIAL | `backup_repository.go` | Archive files on local disk |
| 12 | HA Simulator | HA Placement | `/console/infrastructure` | `/api/v1/infrastructure/simulate` | `infraService` | `infraProvider.SimulationProvider` | Memory State | In-Memory | Tenant ID | `SIMULATED` | SIMULATION_ONLY | `simulation_provider.go` | In-memory simulation |
| 13 | Cloud Mapping | Multi-Cloud | `/console/providers` | `/api/v1/providers/mappings` | `providerService` | `mappingRepository` | Memory / AWS API | In-Memory | Tenant ID | `CONTROL_PLANE_ONLY` | REQUIRES_AWS_KEYS | `provider_service.go` | Requires AWS API keys |
| 14 | IAM Engine | Security | `/console/iam` | `/api/v1/iam/roles` | `authorizationService` | `authorizationService` | Gateway Memory | In-Memory Evaluator | RBAC Roles | `CONTROL_PLANE_ONLY` | READY | `authorization_service.go` | Memory policy evaluation |
| 15 | Observability | Telemetry | `/console/monitoring` | `/api/v1/observability/metrics` | `observabilityService` | `auditLogRepository` | PostgreSQL / Prometheus | Postgres + Memory | Tenant ID | `CONTROL_PLANE_ONLY` | READY | `metrics.go` | Prometheus in memory |
| 16 | Provisioning | Orchestration | `/console/provisioning` | `/api/v1/provisioning/plans` | `provisioningUsecase` | `newMemProvisioningRepo` | Provider Registry | In-Memory | Project ID | `SIMULATED` | NOT_READY | `provisioning_usecase.go` | In-memory tracking |
| 17 | Billing Engine | Financials | `/console/billing` | `/api/v1/billing/usage` | `billingUsecase` | `billingUsecase` | Gateway Memory | In-Memory Quota | Org ID | `SIMULATED` | SIMULATION_ONLY | `billing_usecase.go` | No Stripe connection |
| 18 | Developer APIs | Dev Platform | `/console/developer` | `/api/v1/developer/keys` | `developerUsecase` | `api_key_repository.go` | PostgreSQL / Memory | PostgreSQL GORM | User / Org ID | `CONTROL_PLANE_ONLY` | PARTIAL | `api_key_repository.go` | Keys in DB, webhooks in memory |
| 19 | Reliability Ops | Resilience | `/console/operations` | `/api/v1/reliability/operations` | `reliabilityUsecase` | `reliabilityRepository` | PostgreSQL | PostgreSQL GORM | Tenant ID | `CONTROL_PLANE_ONLY` | READY | `reliability_repository.go` | Worker in gateway |
| 20 | Health Diag | SRE Tools | Public URL | `/api/v1/health/persistence` | `healthService` | `pkg/database/postgres.go` | PostgreSQL DB / Sockets | Live Socket Test | Public Sanitized | `CONTROL_PLANE_ONLY` | READY | `main.go:571` | Public read-only diagnostic |

---

## 6. Audit Summary Statistics

- **Total Capabilities Audited**: `20`
- **REAL EXECUTION Capabilities**: `0`
- **REAL LOCAL Capabilities**: `3` (Local Docker Postgres Engine, Local Docker MySQL Engine, SQL Console)
- **CONTROL-PLANE Capabilities**: `10` (Auth, Orgs/Projects, DB Catalog, Networks, Cloud Mapping, IAM, Observability, Dev APIs, Reliability Ops, Health Diag)
- **DEVELOPMENT-ONLY Capabilities**: `2` (Local Object Storage, Backup Dump Storage)
- **SIMULATED Capabilities**: `4` (Compute, HA Simulator, Provisioning Pipelines, Billing Engine)
- **BROKEN Capabilities**: `1` (Cloud DB Engine Provisioning on host without Docker daemon socket)
- **NOT IMPLEMENTED Capabilities**: `0`

---

## 7. Blockers, Risks & Recommended Implementation Order

### Critical Blockers:
1. **Render Environment Variable (`DATABASE_URL`)**: Requires explicit environment configuration in Render Dashboard pointing to managed PostgreSQL instance.
2. **Object Storage Ephemeral Disk**: `LocalStorageProvider` writes objects to local server disk (`./data/storage`), which is wiped when Render container restarts. Must be upgraded to S3/R2 SDK driver.
3. **Local Docker Dependency**: Managed database engine containers rely on local host Docker daemon.

### Recommended V1 Implementation Sequence:
1. **Phase 1**: Configure `DATABASE_URL` in Render environment to activate PostgreSQL persistence.
2. **Phase 2**: Replace `LocalStorageProvider` with AWS S3 / Cloudflare R2 SDK driver for persistent file storage.
3. **Phase 3**: Replace Docker CLI engine driver with AWS RDS / Kubernetes Operator driver for cloud database instances.
4. **Phase 4**: Migrate in-memory compute and load balancer repositories to PostgreSQL GORM tables.
