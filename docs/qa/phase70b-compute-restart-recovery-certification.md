# ANARVA Cloud Phase 70B — Compute Gateway Restart Recovery Certification Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Certification Date**: August 23, 2026  
**Auditor**: Principal Cloud Security & Infrastructure Engineer  
**Scope**: Compute Gateway Restart Recovery & Re-hydration (`internal/compute/`)  
**Baseline Commit**: `2a9b571` (Phase 68G Complete) + Phase 70A (Compute Tenant Isolation)  

---

## 1. Executive Summary & Root Cause

Phase 70B fixes the gateway restart recovery boundary for ANARVA Compute.

### Original Problem
- **Root Cause**: `LocalDockerComputeProvider` held compute instance runtime references inside an in-memory Go `map[string]*ComputeInstance`. On gateway restart, this RAM map was wiped.
- **Symptom**: Although control-plane metadata (`compute_instances` and `provider_resource_mappings`) survived in PostgreSQL, subsequent `GET`, `Start`, `Stop`, `Restart`, `ExecuteCommand`, and `Delete` requests returned `instance not found in local provider`.

---

## 2. Recovery Architecture & Provider Resource Mapping

```
CONTROL PLANE SOURCE OF TRUTH:
PostgreSQL compute_instances table
             +
PostgreSQL provider_resource_mappings table (Phase 65 Mapping System)

RESTART RECOVERY WORKFLOW:
Gateway Restart (Provider RAM Map Empty)
       │
       v
HTTP Request (Authenticated TenantContext)
       │
       v
GetInstanceForTenant(ctx, orgID, projID, id)
       │
       v
Fetch durable ComputeInstance & ProviderResourceMapping from PostgreSQL
       │
       v
LocalDockerComputeProvider.RehydrateInstance(ctx, inst)
       ├─> Host Docker Daemon Present (hasDocker):
       │      Executes `docker inspect <provider_instance_id>`
       │      ├─> Container Exists: Synchronizes runtime status (RUNNING/STOPPED) & populates RAM cache
       │      └─> Container Missing: Returns error "provider resource no longer exists on host infrastructure"
       └─> Local / Test Simulation Mode:
              Populates RAM cache directly from durable persistent metadata
       │
       v
Compute Operation (GET / Start / Stop / Restart / Execute / Delete) -> SUCCESS!
```

---

## 3. Implementation Details

1. **`internal/compute/provider/provider.go`**:
   - Added `RehydratableProvider` interface with `RehydrateInstance(ctx context.Context, inst *domain.ComputeInstance) error`.
   - Implemented `RehydrateInstance` on `LocalDockerComputeProvider`. When re-hydrating, if host Docker daemon is present, inspects host container status via `docker inspect` and syncs status (`RUNNING` / `STOPPED`). Returns a controlled error if host container resource was destroyed externally.
2. **`internal/compute/usecase/compute_usecase.go`**:
   - Added `mappingRepo mapping.MappingRepository` integration to save `ProviderResourceMapping` records on `CreateInstance` and delete them on `DeleteInstance`.
   - Added `ensureProviderRehydrated(ctx, id)` helper. All compute lifecycle methods (`GetInstanceForTenant`, `StartInstance`, `StopInstance`, `RestartInstance`, `ExecuteCommand`, `GetInstanceMetrics`, `DeleteInstance`) invoke `ensureProviderRehydrated` prior to delegating to `uc.provider`.
3. **`cmd/gateway/main.go`**:
   - Wired `compUC.SetMappingRepository(prvMapRepo)` to bind PostgreSQL provider resource mappings to compute usecase.
4. **`internal/compute/phase70b_restart_recovery_test.go`**:
   - Comprehensive test suite proving instance recovery across gateway process restart, TenantContext isolation preservation post-restart, execution API post-restart, and missing provider resource error handling.

---

## 4. Security Review Results

```
COMPUTE METADATA PERSISTENCE:
    PASS (compute_instances table auto-migrated & persistent in PostgreSQL)

PROVIDER RESOURCE RECOVERY:
    PASS (LocalDockerComputeProvider re-hydrates state from persistent metadata & host Docker daemon)

GATEWAY RESTART RECOVERY:
    PASS (Start/Stop/Restart/Execute operations succeed post-restart)

TENANT ISOLATION:
    PASS (Phase 70A TenantContext validation enforced prior to any recovery or operation)

CROSS-PROJECT ISOLATION:
    PASS (Org & Project scoping enforced on recovery)

PROVIDER RESOURCE MAPPING:
    PASS (provider_resource_mappings populated on compute creation & deleted on termination)

MISSING PROVIDER RESOURCE HANDLING:
    SAFE (Controlled error returned if host container resource no longer exists)

RECOVERY AUTHORIZATION:
    PASS (Recovery requires authenticated TenantContext match)

PHASE 70A REGRESSION:
    PASS (All 9 Phase 70A tenant isolation tests remain 100% PASS)

BLOCKING FINDINGS:
    NO BLOCKING FINDINGS.
```

---

## 5. Remaining Compute Gaps (Future Phase Scope)

1. **Frontend Authority & Web Terminal**: Connect Next.js frontend web terminal to `POST /api/v1/compute/instances/{id}/execute`.
2. **Secret Encryption**: Encrypt `env_vars_json` column in PostgreSQL database.
3. **Cloud Infrastructure Integration**: Connect cloud providers (AWS EC2 / GCP Compute Engine / Kubernetes).
