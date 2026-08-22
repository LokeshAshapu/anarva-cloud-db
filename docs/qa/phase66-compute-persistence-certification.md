# ANARVA Cloud Phase 66 — Compute Control-Plane Persistence Certification Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Certification Date**: August 22, 2026  
**Auditor**: Principal Cloud Architect, Go Systems Engineer & SRE  
**Scope**: Compute Control-Plane Persistence (`internal/compute/`)  
**Baseline Commit**: `1df95bb` (Phase 65 Complete)  

---

## 1. Executive Summary & Before/After Architecture

Phase 66 eliminates the in-memory process map dependency for compute instances (`newMemComputeRepo`) and replaces it with a PostgreSQL-backed GORM repository (`PostgresComputeRepository` and `PostgresVolumeRepository`).

```
BEFORE PHASE 66:
Compute API Request -> ComputeUseCase -> newMemComputeRepo (Go Process RAM Map)
                                              │
                                   [ GATEWAY RESTART ]
                                              │
                                              v
                                   ALL COMPUTE METADATA LOST!

AFTER PHASE 66:
Compute API Request -> ComputeUseCase -> PostgresComputeRepository -> PostgreSQL Database
                                                                           │
                                                                   [ GATEWAY RESTART ]
                                                                           │
                                                                           v
                                                                 COMPUTE METADATA SURVIVES!
```

---

## 2. PostgreSQL Tables & Schema Migration

### Table: `compute_instances`
```sql
CREATE TABLE compute_instances (
    id VARCHAR(255) PRIMARY KEY,
    resource_id VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    region_id VARCHAR(100) NOT NULL,
    zone_id VARCHAR(100),
    status VARCHAR(50) NOT NULL,
    health VARCHAR(50) NOT NULL,
    plan_id VARCHAR(100) NOT NULL,
    acu NUMERIC(10,2) NOT NULL,
    vcpu NUMERIC(10,2) NOT NULL,
    memory_mb INTEGER NOT NULL,
    storage_gb INTEGER NOT NULL,
    image_id VARCHAR(100) NOT NULL,
    docker_image VARCHAR(255),
    network_id VARCHAR(255),
    subnet_id VARCHAR(255),
    private_ip VARCHAR(100),
    public_ip VARCHAR(100),
    provider VARCHAR(100) NOT NULL,
    provider_instance_id VARCHAR(255),
    security_json TEXT,
    env_vars_json TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_compute_instances_org_id ON compute_instances(organization_id);
CREATE INDEX idx_compute_instances_project_id ON compute_instances(project_id);
CREATE INDEX idx_compute_instances_provider_instance_id ON compute_instances(provider_instance_id);
CREATE INDEX idx_compute_instances_deleted_at ON compute_instances(deleted_at);
```

### Table: `compute_volumes`
```sql
CREATE TABLE compute_volumes (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    instance_id VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    size_gb INTEGER NOT NULL,
    region_id VARCHAR(100) NOT NULL,
    zone_id VARCHAR(100),
    type VARCHAR(50) NOT NULL,
    provider_volume_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_compute_volumes_org_id ON compute_volumes(organization_id);
CREATE INDEX idx_compute_volumes_project_id ON compute_volumes(project_id);
CREATE INDEX idx_compute_volumes_instance_id ON compute_volumes(instance_id);
```

---

## 3. Implementation Summary

1. **GORM Model Tags & JSON Serialization Hooks** (`internal/compute/domain/compute.go`):
   - Added GORM column tags and `TableName()` methods for `ComputeInstance` (`compute_instances`) and `Volume` (`compute_volumes`).
   - Added `BeforeSave` and `AfterFind` GORM hooks on `ComputeInstance` to transparently serialize/deserialize `Security` policies and `EnvVars` map to/from `security_json` and `env_vars_json`.
2. **PostgreSQL Repositories** (`internal/compute/repository/postgres_repository.go`):
   - Implemented `PostgresComputeRepository` and `PostgresVolumeRepository` satisfying `ComputeRepository` and `VolumeRepository` interfaces.
   - Enforced tenant isolation checks (`GetTenantScopedByID`, `GetTenantScopedVolumeByID`) returning `TENANT_ISOLATION_VIOLATION` on cross-tenant access.
3. **Gateway Production Wiring** (`cmd/gateway/main.go`):
   - Registered `&computeDomain.ComputeInstance{}` and `&computeDomain.Volume{}` in GORM `AutoMigrate(...)`.
   - Wired `PostgresComputeRepository` and `PostgresVolumeRepository` when `dbPool != nil`.
   - Production mode (`ANARVA_ENV=production`) **FAILS CLOSED** if `dbPool == nil`.

---

## 4. Verification Matrix & Test Results

| Verification Item | Details | Result | Evidence |
|:---|:---|:---:|:---|
| **PostgreSQL Persistence** | `PostgresComputeRepository` & `PostgresVolumeRepository` | 🟢 PASS | `internal/compute/repository/postgres_repository.go` |
| **GORM AutoMigrate** | Added models to gateway migration | 🟢 PASS | `cmd/gateway/main.go:356-357` |
| **JSON Hooks** | `Security` & `EnvVars` serialization | 🟢 PASS | `TestComputeDomain_JSONSerializationHooks` |
| **Tenant Isolation** | Scoped queries & cross-tenant rejection | 🟢 PASS | `GetTenantScopedByID` returns `TENANT_ISOLATION_VIOLATION` |
| **Restart Persistence** | Verified instance survival across repository reconstruction | 🟢 PASS | `TestPostgresComputeRepository_LiveDBOrSkip` step 2 |
| **Production Fail-Closed** | Forbids fallback to in-memory in production | 🟢 PASS | Gateway `log.Fatal` assertion in `main.go:412` |
| **Unit & Integration Suite**| `go test -v ./internal/compute/...` | 🟢 PASS | 100% test suite pass |
| **Gateway Package Suite** | `go test -v ./cmd/gateway/...` | 🟢 PASS | 100% test suite pass |
| **Production Binaries** | `go build ./cmd/gateway`, `./cmd/anarva` | 🟢 PASS | Binaries compiled cleanly |
| **Next.js Web App** | `npm run build` in `web/` | 🟢 PASS | 42/42 static & dynamic routes compiled |

---

## 5. Classification & Production Readiness

- **Compute Control-Plane Metadata**: **PRODUCTION_READY (DURABLE)**
- **Provider Resource Mapping**: **PRODUCTION_READY (DURABLE)**
- **Users & Authentication**: **PRODUCTION_READY (DURABLE)**
- **PostgreSQL Database Instance Metadata**: **PRODUCTION_READY (DURABLE)**
- **Backup Records & S3 Archives**: **PRODUCTION_READY (DURABLE)**
- **Networking Infrastructure (VPC/Subnets/SG)**: **PRODUCTION_READY (DURABLE)**
- **Reliability Operations & Leases**: **PRODUCTION_READY (DURABLE)**

---

## 6. Recommended Phase 67 Next Step

**RECOMMENDED PHASE 67 — PERSISTENT LOAD BALANCER & INGRESS CONTROL PLANE METADATA**
- **Objective**: Replace the in-memory `LoadBalancerRepository` with a PostgreSQL-backed GORM repository (`PostgresLoadBalancerRepository` -> `load_balancers`, `listeners`, `backend_pools`, `certificates`, `domains` tables) so load balancer rules and TLS certificates survive gateway restarts.
