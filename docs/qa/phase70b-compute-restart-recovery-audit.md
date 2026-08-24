# ANARVA Cloud Phase 70B — Compute Gateway Restart Recovery Audit Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 23, 2026  
**Auditor**: Principal Cloud Infrastructure Architect & Security Lead  
**Scope**: Compute Gateway Restart Recovery (`internal/compute/`) Forensic Audit  
**Baseline Commit**: `2a9b571` (Phase 68G Complete) + Phase 70A (Compute Tenant Isolation)  

---

## 1. Executive Summary

Phase 70B investigates and fixes the gateway restart recovery gap in ANARVA Compute. While control-plane metadata (`compute_instances` and `provider_resource_mappings`) is durable in PostgreSQL, `LocalDockerComputeProvider` relies on an in-memory Go map (`instances map[string]*ComputeInstance`). On gateway restart, this map is erased, causing subsequent `Start`, `Stop`, `Restart`, and `Execute` operations to fail with `instance not found in local provider`.

---

## 2. 10-Point Forensic Inspection

### 1. What is stored in `compute_instances`?
Durable PostgreSQL control-plane metadata: `id`, `resource_id`, `organization_id`, `project_id`, `name`, `slug`, `region_id`, `zone_id`, `status`, `health`, `plan_id`, `acu`, `vcpu`, `memory_mb`, `storage_gb`, `image_id`, `docker_image`, `network_id`, `subnet_id`, `private_ip`, `public_ip`, `provider`, `provider_instance_id`, `security_json`, `env_vars_json`, `created_at`, `updated_at`, `deleted_at`.

### 2. What is stored in `provider_resource_mappings`?
Durable mapping table (`ProviderResourceMapping` model) linking `anarva_resource_id` (`inst.ID`) to `provider_resource_id` (`inst.ProviderInstanceID`), along with `organization_id`, `project_id`, `provider`, `provider_resource_type` (`COMPUTE_INSTANCE`), `region`, `zone`, `status`, `managed`.

### 3. What is stored only in `LocalDockerComputeProvider.instances`?
An in-memory Go `map[string]*ComputeInstance` in RAM storing transient pointers to `ComputeInstance` structs.

### 4. What information is required to reconstruct provider state?
`ID`, `ProviderInstanceID` (container ID/name or simulation ID), `Slug`, `DockerImage`, `VCPU`, `MemoryMB`, `OrganizationID`, `ProjectID`, `Status`, `Health`.

### 5. Does the actual Docker container survive gateway restart?
**YES**. Docker containers run inside the host OS Docker daemon process (`dockerd`). Restarting the ANARVA gateway process does NOT kill host Docker containers (`anarva-acu-<slug>`).

### 6. Is `provider_instance_id` sufficient to reconnect to an existing container?
**YES**. Executing `docker inspect <provider_instance_id>` or `docker inspect anarva-acu-<slug>` verifies container existence and queries runtime status (`running`, `exited`).

### 7. Can the provider discover existing containers?
**YES**. Via `docker inspect` for real containers, or via lazy hydration from durable control-plane metadata for simulated references.

### 8. Can provider state be reconstructed from PostgreSQL?
**YES**. When a provider lookup misses in RAM, loading the durable `ComputeInstance` from `PostgresComputeRepository` / `provider_resource_mappings` allows `LocalDockerComputeProvider` to lazily re-hydrate its runtime map.

### 9. Do Start/Stop/Restart/Execute require the in-memory map?
Currently, yes. If `p.instances[id]` is missing, `LocalDockerComputeProvider` returns `"instance not found"`. Lazy hydration or direct DB fallback resolves this issue.

### 10. Should gateway restart recreate provider state or discover existing resources?
It must **discover and reconnect** to existing provider resources using `provider_instance_id` and durable PostgreSQL metadata.

---

## 3. Data Loss & Recovery Boundary

```
[Gateway Restart]
       │
       v
PostgreSQL DB (compute_instances & provider_resource_mappings) -> SURVIVES RESTART
Host Docker Daemon (containers anarva-acu-*)                   -> SURVIVES RESTART
LocalDockerComputeProvider.instances RAM map                   -> ERASED ON RESTART

RECOVERY PATH:
HTTP Request (Authenticated TenantContext)
       │
       v
GetInstanceForTenant (Validates Tenant A vs Tenant B)
       │
       v
PostgresComputeRepository (Fetch durable ComputeInstance)
       │
       v
LocalDockerComputeProvider.RehydrateInstance(inst)
       ├─> Real Docker: exec docker inspect <provider_instance_id>
       │      ├─> Container Exists: Sync status (RUNNING / STOPPED) & Re-hydrate RAM map
       │      └─> Container Missing: Return error "provider resource unavailable"
       └─> Simulation: Re-hydrate RAM map from persistent metadata
       │
       v
Execute Operation (Start / Stop / Restart / Execute) -> SUCCESS!
```

---

## 4. Architectural Choice & Plan

We choose **Lazy Hydration with Provider Discovery**:
1. Wire `MappingRepository` into `ComputeUseCase` to create/delete `provider_resource_mappings` on compute creation/deletion.
2. Add `RehydrateInstance(inst *domain.ComputeInstance)` and `GetInstanceWithRecovery(ctx, id)` to `LocalDockerComputeProvider`.
3. In `LocalDockerComputeProvider`:
   - If `id` is not in RAM map `p.instances`:
     - If `inst` is provided from DB, check Docker host container status via `docker inspect` (if `p.hasDocker` and not simulated ID).
     - If real container exists (or simulated reference), populate `p.instances[id] = inst` and return it.
     - If container was deleted externally from Docker, return error `compute instance provider container no longer exists`.
4. Enforce Phase 70A `TenantContext` authorization before any recovery or provider operation.
