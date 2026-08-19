# ANARVA Cloud V1 — Security & Multi-Tenant Isolation Boundaries

**Audit Date**: August 19, 2026  
**Auditor**: Principal Cloud Security Architect & Security Engineer  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  

---

## 1. Multi-Tenant Scoping Pipeline

```
Client HTTP Request
  │
  ▼
[internal/gateway/middleware/auth_middleware.go]
  ├─ Validates JWT Signature / API Key Hash
  ├─ Extracts User ID, Organization ID, Tenant ID
  └─ Injects TenantContext into Request Context
  │
  ▼
[Controller / Handler Layer]
  ├─ Calls middleware.GetTenantContext(r.Context())
  └─ Passes Authorized Org/Project IDs to UseCase
  │
  ▼
[Repository Layer (GORM)]
  └─ SELECT * FROM resource_table WHERE project_id IN (?) AND organization_id = ?
```

---

## 2. Security Scoping Audit & IDOR Prevention Matrix

| Endpoint Route | Auth Required? | Tenant Scoping Method | Server-Side Validation | IDOR Protection Status | Evidence |
|:---|:---:|:---|:---|:---:|:---|
| `POST /api/v1/auth/register` | No | Public Endpoint | Password hashing (Bcrypt) | N/A | `auth_handler.go` |
| `POST /api/v1/auth/login` | No | Public Endpoint | Password check | N/A | `auth_handler.go` |
| `GET /api/v1/auth/me` | Yes | JWT `sub` Claim | Token validation | 🟢 PASS | `auth_handler.go` |
| `GET /api/v1/organizations` | Yes | JWT TenantContext | Org Membership Query | 🟢 PASS | `project_usecase.go` |
| `POST /api/v1/projects` | Yes | JWT TenantContext | Org ID Verification | 🟢 PASS | `project_repository.go` |
| `GET /api/v1/databases` | Yes | JWT TenantContext | Project ID Array Filter | 🟢 PASS | `database_handler.go` |
| `GET /api/v1/databases/{id}` | Yes | JWT TenantContext | Instance Ownership Check | 🟢 PASS | `database_handler.go` |
| `POST /api/v1/databases/{id}/query` | Yes | JWT TenantContext | Database Ownership Verification | 🟢 PASS | `sql_service.go` |
| `GET /api/v1/networking/vnets` | Yes | JWT TenantContext | Project ID Filter | 🟢 PASS | `postgres_repository.go` |
| `GET /api/v1/backups` | Yes | JWT TenantContext | Project ID Filter | 🟢 PASS | `backup_repository.go` |
| `GET /api/v1/reliability/operations` | Yes | JWT TenantContext | Tenant ID Filter | 🟢 PASS | `reliability_repository.go` |
| `GET /api/v1/health/persistence` | No | Public Diagnostic | Sanitized Output (No Secrets) | 🟢 PASS | `main.go:571` |

---

## 3. Strict Server-Side Authorization Mandate

- Client-supplied resource IDs (`{id}`) must **NEVER** be trusted directly.
- Handlers must retrieve `TenantContext` from `r.Context()` and verify that the target resource belongs to an organization/project owned by the caller. Rejects unauthorized access with `HTTP 404 Not Found` or `HTTP 403 Forbidden`.
