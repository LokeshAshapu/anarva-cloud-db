# ANARVA Cloud Phase 70A — Compute Tenant Isolation Certification Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Certification Date**: August 23, 2026  
**Auditor**: Principal Cloud Security Engineer  
**Scope**: Compute HTTP API Tenant Isolation Enforcement (`internal/compute/delivery/http/compute_handler.go`)  
**Baseline Commit**: `2a9b571` (Phase 68G Complete)  

---

## 1. Executive Summary & Root Cause

Phase 70A fixes the compute tenant isolation vulnerability identified in Phase 69 (`docs/qa/phase69-compute-workflow-forensic-audit.md`).

### Original Vulnerability & Root Cause
- **Vulnerability**: `ComputeHandler` in `internal/compute/delivery/http/compute_handler.go` did **NOT** extract `TenantContext` from authenticated JWT contexts (`security.GetTenantContext(r.Context())`).
- **Authorization Bypass**: Missing parameter fallbacks automatically defaulted `OrganizationID` to `"org-default"` and `ProjectID` to `"proj-default"`. An attacker knowing another tenant's compute instance ID could query (`GET`), start (`POST /start`), stop (`POST /stop`), restart (`POST /restart`), or delete (`DELETE`) another tenant's compute instance.

---

## 2. Authorization Flow Transformation

```
BEFORE PHASE 70A (UNPROTECTED):
HTTP Request -> ComputeHandler -> Hardcoded Fallbacks ("org-default", "proj-default")
             -> GetInstance(id) -> Resource Expose / Mutate (NO TENANT CHECK!)

AFTER PHASE 70A (PROTECTED):
HTTP Request -> JWT Middleware -> TenantContext (OrganizationID, ProjectID)
             -> ComputeHandler -> security.GetTenantContext(r.Context())
             -> GetInstanceForTenant(ctx, orgID, projID, id)
             -> GetTenantScopedByID(ctx, orgID, projID, id)
             -> OrganizationID / ProjectID Match Check
                  ├─> Match: Proceed to Operation
                  └─> Mismatch: HTTP 403 TENANT_ISOLATION_VIOLATION
```

---

## 3. Implementation Details

1. **`internal/compute/delivery/http/compute_handler.go`**:
   - `handleInstances`: Extracts `tc := security.GetTenantContext(r.Context())`. On creation (`POST`), validates that client payload does not conflict with non-empty `tc.OrganizationID` / `tc.ProjectID`. Rejects conflicts with HTTP 403 `TENANT_ISOLATION_VIOLATION`. Uses `tc.OrganizationID` / `tc.ProjectID` as authoritative identity.
   - `handleInstanceSubroutes`: Invokes `GetInstanceForTenant(r.Context(), tc.OrganizationID, tc.ProjectID, id)` **BEFORE** executing any subroute operation (`GET`, `DELETE`, `start`, `stop`, `restart`, `execute`, `metrics`). Unauthorized requests return HTTP 403 `TENANT_ISOLATION_VIOLATION`.
2. **`internal/compute/usecase/compute_usecase.go`**:
   - Added `GetInstanceForTenant(ctx, orgID, projID, id)` and `ListInstancesForTenant(ctx, orgID, projID)`. Delegates to `repo.GetTenantScopedByID` when repository is present, and enforces context checks when repository is absent.
3. **`internal/compute/phase70a_tenant_isolation_test.go`**:
   - Comprehensive unit test suite proving Tenant B rejection (HTTP 403 `TENANT_ISOLATION_VIOLATION`) across all read, write, lifecycle, listing, and creation endpoints.

---

## 4. Security Verification Results

```
COMPUTE TENANT ISOLATION:
    PASS

CROSS-PROJECT ISOLATION:
    PASS

AUTHENTICATED TENANT CONTEXT:
    PASS

DEFAULT TENANT FALLBACK:
    REMOVED

LIST ENDPOINT ISOLATION:
    PASS

CREATE TENANT AUTHORITY:
    PASS

MUTATION ENDPOINT ISOLATION:
    PASS

INFORMATION DISCLOSURE:
    SAFE

REGRESSION TESTS:
    PASS

BLOCKING FINDINGS:
    NO BLOCKING FINDINGS.
```

---

## 5. Remaining Compute Gaps (Future Phase Scope)

1. **Provider Restart Recovery**: `LocalDockerComputeProvider` in-memory container state is lost on gateway restart. (Target: Phase 70B).
2. **Frontend Authority & Web Terminal Integration**: Connect Next.js frontend web terminal to `POST /api/v1/compute/instances/{id}/execute`.
3. **Secret Encryption**: Encrypt `env_vars_json` column in PostgreSQL database.
