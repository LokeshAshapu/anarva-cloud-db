# ANARVA Cloud Phase 68B — Control-Plane Persistence Certification Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Certification Date**: August 23, 2026  
**Auditor**: Principal Cloud Architect & Security Engineer  
**Scope**: Database Control-Plane Metadata Persistence & Tenant Authorization (`internal/postgres/`)  
**Baseline Commit**: `51cea88` (Phase 66 Complete)  

---

## 1. Executive Summary & Before/After Architecture

Phase 68B makes PostgreSQL database instance metadata (`PostgresInstance`) durable by persisting instance records into the PostgreSQL control-plane database (`postgres_instances` table) using GORM. It also secures all database API subroutes (including `POST /api/v1/databases/{id}/query`) by enforcing tenant ownership validation (`TenantContext`) prior to query execution.

```
BEFORE PHASE 68B:
Create Instance -> LocalDockerPostgresProvider (Go RAM map) -> GATEWAY RESTART -> METADATA LOST!
Query Endpoint -> POST /api/v1/databases/{id}/query -> NO TENANT CHECK -> IDOR VULNERABILITY!

AFTER PHASE 68B:
Create Instance -> GormPostgresInstanceRepository -> postgres_instances table -> DURABLE!
Query Endpoint -> TenantContext Validation -> GetInstanceForTenant -> 403 TENANT_ISOLATION_VIOLATION on unauthorized access -> SECURE!
```

---

## 2. PostgreSQL Table & Schema Migration

### Table: `postgres_instances`
```sql
CREATE TABLE postgres_instances (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) DEFAULT 'LOCAL_POSTGRES',
    version VARCHAR(50) DEFAULT '17',
    status VARCHAR(50) DEFAULT 'CREATING',
    region_id VARCHAR(100) NOT NULL,
    zone_id VARCHAR(100),
    cpu NUMERIC(10,2) DEFAULT 1.0,
    memory_mb INTEGER DEFAULT 1024,
    storage_gb INTEGER DEFAULT 25,
    storage_type VARCHAR(50) DEFAULT 'SSD',
    network_id VARCHAR(255) NOT NULL,
    subnet_id VARCHAR(255),
    availability_mode VARCHAR(50) DEFAULT 'SINGLE',
    backup_mode VARCHAR(50) DEFAULT 'DAILY_SNAPSHOT',
    maintenance_window VARCHAR(100) DEFAULT 'Sun:03:00',
    provider_resource_id VARCHAR(255),
    host VARCHAR(255),
    port INTEGER DEFAULT 5432,
    public_access BOOLEAN DEFAULT FALSE,
    reality_label VARCHAR(100) DEFAULT 'LOCAL_POSTGRES',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_postgres_instances_org_id ON postgres_instances(organization_id);
CREATE INDEX idx_postgres_instances_project_id ON postgres_instances(project_id);
CREATE INDEX idx_postgres_instances_deleted_at ON postgres_instances(deleted_at);
```

---

## 3. Implementation Summary

1. **GORM Model Tags & TableName** (`internal/postgres/domain/postgres.go`):
   - Configured `PostgresInstance`, `PostgresDatabase`, `PostgresUser` with GORM tags and `TableName()` methods. Added `DeletedAt *time.Time`.
2. **PostgreSQL Control-Plane Repository** (`internal/postgres/repository/postgres_repository.go`):
   - Created `GormPostgresInstanceRepository` implementing `PostgresInstanceRepository` interface.
   - Enforced tenant isolation checks (`GetByIDForTenant`, `ListByProject`) returning `TENANT_ISOLATION_VIOLATION` on cross-tenant access.
3. **Service Layer** (`internal/postgres/service/postgres_service.go`):
   - Updated `PostgresService` to integrate with `PostgresInstanceRepository` (`NewPostgresServiceWithRepo`).
4. **API Handler & Tenant Authorization** (`internal/postgres/handler/postgres_handler.go`):
   - Enforced `security.GetTenantContext(r.Context())` and `GetInstanceForTenant(...)` validation across all database endpoints.
   - Unauthorized access attempts on `/api/v1/databases/{id}/query` return HTTP 403 `TENANT_ISOLATION_VIOLATION`.
5. **Gateway Wiring & AutoMigrate** (`cmd/gateway/main.go`):
   - Added `&postgresDomain.PostgresInstance{}` to `AutoMigrate(...)`.
   - Production mode (`ANARVA_ENV=production`) **FAILS CLOSED** (`log.Fatal`) if `dbPool == nil`.
6. **Frontend Integration** (`web/app/console/databases/page.tsx`):
   - Updated `useEffect` to fetch authoritative database list from `GET /api/v1/databases`.

---

## 4. Verification Matrix & Test Results

| Verification Item | Details | Result | Evidence |
|:---|:---|:---:|:---|
| **Database Instance Durability** | `GormPostgresInstanceRepository` & `postgres_instances` | 🟢 PASS | `internal/postgres/repository/postgres_repository.go` |
| **GORM AutoMigrate** | Added `PostgresInstance` to gateway migration | 🟢 PASS | `cmd/gateway/main.go:361` |
| **Tenant Isolation** | Scoped queries & cross-tenant rejection | 🟢 PASS | `TestPostgresQueryEndpoint_TenantAuthorization` (Tenant B rejected with 403) |
| **Restart Persistence** | Verified instance survival across repository reconstruction | 🟢 PASS | `TestPostgresInstanceRepository_LiveDBOrSkip` |
| **Production Fail-Closed** | Forbids fallback to in-memory in production | 🟢 PASS | Gateway `log.Fatal` assertion in `main.go:447` |
| **Unit & Integration Suite**| `go test -v ./internal/postgres/...` | 🟢 PASS | 100% test suite pass |
| **Gateway Package Suite** | `go test -v ./cmd/gateway/...` | 🟢 PASS | 100% test suite pass |
| **Production Binaries** | `go build ./cmd/gateway`, `./cmd/anarva` | 🟢 PASS | Binaries compiled cleanly |
| **Next.js Web App** | `npm run build` in `web/` | 🟢 PASS | 42/42 static & dynamic routes compiled |

---

## 5. Intentionally Unresolved Data-Plane Limitations

The following items belong to future data-plane phases and remain intentionally unresolved in Phase 68B:
- Customer SQL table schema durability (`CREATE TABLE`)
- Customer SQL row data durability (`INSERT INTO`)
- Real PostgreSQL engine execution
- Real Docker container / volume mounting
- Removal of `SQLService` simulation

---

## FINAL CERTIFICATION

PHASE 68B:

Database Instance Metadata:
    DURABLE

Control Plane PostgreSQL:
    PASS

Tenant Isolation:
    PASS

Cross-Project Isolation:
    PASS

Production Memory Fallback:
    BLOCKED

Backend Database Listing:
    AUTHORITATIVE

SQL Data Plane:
    UNCHANGED

Customer Row Persistence:
    NOT YET IMPLEMENTED

NO BLOCKING FINDINGS:
    YES
