# ANARVA Cloud V1 — Multi-Tenant Security & Isolation Audit

**Audit Date**: August 19, 2026  
**Auditor**: Principal Cloud Security Architect & Security Engineer  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  

---

## 1. Security Architecture & Identity Scoping

The ANARVA Cloud Control Plane uses JWT authentication, API Key verification, and an explicit `TenantContext` injected into HTTP request context to enforce multi-tenant isolation across all control-plane APIs.

```
Incoming HTTP Request
  │
  ▼
[internal/gateway/middleware/auth_middleware.go]
  ├─ Authenticate JWT / API Key
  ├─ Extract User ID, Organization ID, Tenant ID
  └─ Build TenantContext -> r.WithContext(...)
  │
  ▼
[Controller / Handler]
  ├─ Retrieve TenantContext via middleware.GetTenantContext(r.Context())
  └─ Pass Tenant ID / Project ID to UseCase
  │
  ▼
[Repository GORM Query]
  └─ SELECT * FROM table WHERE project_id = ? AND organization_id = ?
```

---

## 2. Security Subsystem Audit Matrix

| Security Domain | Implementation Location | Authentication Mechanism | Tenant Scoping Method | Database Enforcement | Security Rating | Evidence |
|:---|:---|:---|:---|:---|:---|:---|
| **User Auth** | `internal/auth/delivery/http/auth_handler.go` | JWT Bearer Token / Bcrypt | User ID Claim (`sub`) | `WHERE id = ?` | 🟢 PASS | `auth_handler.go` |
| **Organization Access** | `internal/project/usecase/project_usecase.go` | JWT Bearer Token | Org Membership Check | `WHERE user_id = ? AND org_id = ?` | 🟢 PASS | `project_usecase.go` |
| **Project Access** | `internal/project/repository/project_repository.go` | JWT Bearer Token | TenantContext Org ID | `WHERE org_id = ?` | 🟢 PASS | `project_repository.go` |
| **Database Instance Catalog** | `internal/database/delivery/http/database_handler.go` | JWT Bearer Token | Project ID Scoping | `WHERE project_id IN (...)` | 🟢 PASS | `database_handler.go` |
| **SQL Console Queries** | `internal/postgres/service/sql_service.go` | JWT + DB Ownership Check | Database ID Ownership | Database Instance Check | 🟢 PASS | `sql_service.go` |
| **Virtual Networks** | `internal/networking/repository/postgres_repository.go` | JWT Bearer Token | Project ID Scoping | `WHERE project_id = ?` | 🟢 PASS | `postgres_repository.go` |
| **Backup Records** | `internal/backup/repository/backup_repository.go` | JWT Bearer Token | Project ID Scoping | `WHERE project_id = ?` | 🟢 PASS | `backup_repository.go` |
| **Reliability Ops & Locks** | `internal/reliability/repository/reliability_repository.go` | JWT Bearer Token | Tenant ID Scoping | `WHERE tenant_id = ?` | 🟢 PASS | `reliability_repository.go` |
| **Object Storage Buckets** | `internal/storage/service/storage_service.go` | JWT Bearer Token | Bucket Path Prefix | Path Prefixing | 🟡 PARTIAL (File path) | `storage_service.go` |
| **Public Diagnostics** | `cmd/gateway/main.go:571` | Public Read-Only | None (Sanitized Output) | Diagnostic Ping Only | 🟢 PASS | `main.go` |

---

## 3. IDOR (Insecure Direct Object Reference) Vulnerability Forensic Audit

1. **Database Instance Fetch (`GET /api/v1/databases/{id}`)**:
   - **Verification**: Handlers call `TenantContext` to fetch authorized project IDs first before querying the database repository. Rejects unauthorized access with `HTTP 404 / HTTP 403`.
2. **SQL Query Execution (`POST /api/v1/databases/{id}/query`)**:
   - **Verification**: Verifies that the targeted database instance ID belongs to an authorized project owned by the user before opening a connection.
3. **Public Endpoint Sanitization (`GET /api/v1/health/persistence`)**:
   - **Verification**: Scoped as public read-only without requiring Authorization header. Returns high-level connectivity diagnostics (`PASS`/`FAIL`) and build commit hash without exposing passwords, connection credentials, or tokens.
