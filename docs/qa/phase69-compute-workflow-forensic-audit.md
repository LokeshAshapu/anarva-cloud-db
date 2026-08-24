# ANARVA Cloud Phase 69 — Compute Page One-Workflow Forensic Audit Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 23, 2026  
**Auditor**: Principal Cloud Infrastructure Architect & Security Lead  
**Scope**: Compute Workflow (`/console/compute`) Read-Only Forensic Audit  
**Baseline Commit**: `2a9b571` (Phase 68G Complete)  

---

## 1. Executive Summary

Phase 69 performs a read-only forensic architectural audit of the ANARVA Compute management workflow (`/console/compute`). Following the certified reference workflow established in Phase 68G (`/console/databases`), this audit investigates whether compute instance creation, lifecycle operations (Start, Stop, Restart), provider container execution, control-plane metadata persistence (`compute_instances`), provider resource mapping (`provider_resource_mappings`), tenant isolation, secret security, and gateway restart recovery are genuinely durable or simulated.

---

## 2. Frontend Architecture (`web/app/console/compute/page.tsx`)

- **Backend API Integration**: The page invokes `GET /api/v1/compute/instances?projectId=proj-default` on mount inside `fetchInstances()`.
- **Frontend Fallback Vulnerability**: When the backend API is unreachable or returns an empty payload, `handleCreateInstance` falls back to generating a client-side ephemeral object (`newInst`) with random IP addresses (`privateIp: 10.0.1.X`, `publicIp: 20.198.X.X`) and `providerInstanceId: docker-sim-${Date.now()}`.
- **Web Terminal Simulation**: The interactive container terminal in `handleExecuteCommand` uses `setTimeout` with hardcoded fake outputs (`uname -a`, `ps aux`, `cat /etc/env`) rather than streaming real stdout/stderr from `POST /api/v1/compute/instances/{id}/execute`.
- **LocalStorage State**: Compute instance objects are maintained in React `useState` (`instances`). Unlike the database page, compute instance metadata is not written to `localStorage`, but client-side creation fallbacks prevent the backend catalog from being strictly authoritative when backend errors occur.

---

## 3. API Request Flow

```
User Click (/console/compute)
             │
             v
   POST /api/v1/compute/instances
             │
             v
   ComputeHandler (internal/compute/delivery/http/compute_handler.go)
             │
             v
   ComputeUseCase (internal/compute/usecase/compute_usecase.go)
             │
     ┌───────┴─────────────────────────────┐
     │                                     │
     v                                     v
LocalDockerComputeProvider        PostgresComputeRepository
(internal/compute/provider)      (internal/compute/repository)
     │                                     │
     v                                     v
docker run -d --name container      GORM Save -> compute_instances
(Host Docker Daemon)               (ANARVA Control Plane DB)
```

---

## 4. Authentication Flow

- All frontend HTTP requests pass `getAuthHeaders()`, supplying `Authorization: Bearer <token>`.
- The gateway middleware validates JWT signatures and populates security context.

---

## 5. TenantContext Flow

- **Handler Extraction Gap**: `ComputeHandler.handleInstances` and `handleInstanceSubroutes` decode JSON payloads or URL parameters, defaulting `OrganizationID` to `"org-default"` and `ProjectID` to `"proj-default"` if missing from the request body.
- **Tenant Context Bypass**: Handlers do **NOT** invoke `security.GetTenantContext(r.Context())` to extract authenticated tenant parameters from JWT context.
- **Repository Isolation**: `PostgresComputeRepository` implements `GetTenantScopedByID(ctx, orgID, projID, id)`, which correctly rejects cross-tenant access with `TENANT_ISOLATION_VIOLATION`. However, HTTP handlers currently invoke `uc.GetInstance(ctx, id)` instead of `GetTenantScopedByID`, leaving the HTTP routing layer un-isolated!

---

## 6. Compute Control Plane

- **Table Schema**: `compute_instances` table defined in `internal/compute/domain/compute.go`.
- **Fields Persisted**: `id`, `resource_id` (ARNV format), `organization_id`, `project_id`, `name`, `slug`, `region_id`, `zone_id`, `status`, `health`, `plan_id`, `acu`, `vcpu`, `memory_mb`, `storage_gb`, `image_id`, `docker_image`, `network_id`, `subnet_id`, `private_ip`, `public_ip`, `provider`, `provider_instance_id`, `security_json`, `env_vars_json`, `created_at`, `updated_at`, `deleted_at`.
- **GORM AutoMigrate**: Registered in `cmd/gateway/main.go` line 359 (`&computeDomain.ComputeInstance{}`).
- **Durability Status**: **REAL DURABLE** when PostgreSQL control-plane DB is connected.

---

## 7. Compute Data Plane

- **Local Docker Containers**: Managed via `LocalDockerComputeProvider` (`internal/compute/provider/provider.go`).
- **Container Execution**: Runs `docker run -d --name anarva-acu-<slug> --cpus <vcpu> --memory <memory>m <docker_image>`.
- **Simulation Fallback**: If `docker` CLI is absent on host, `LocalDockerComputeProvider` sets `providerInstanceId = local-sim-<id>` and stores container metadata in an in-memory Go map (`map[string]*ComputeInstance`).

---

## 8. Provider Architecture (`internal/compute/provider/provider.go`)

- **Interface**: `ComputeProvider` defines `CreateInstance`, `GetInstance`, `ListInstances`, `StartInstance`, `StopInstance`, `RestartInstance`, `DeleteInstance`, `ResizeInstance`, `RebuildInstance`, `GetInstanceHealth`, `GetInstanceMetrics`, `ExecuteCommand`.
- **LocalDockerComputeProvider Implementation**:
  - Uses `os/exec.CommandContext` to invoke local `docker` CLI commands (`docker run`, `docker start`, `docker stop`, `docker rm`).
  - In-Memory State: Uses `p.instances map[string]*ComputeInstance` protected by `sync.RWMutex`.

---

## 9. Provider Resource Mapping (`provider_resource_mappings`)

- **Current Integration Gap**: Compute instance creation in `ComputeUseCase.CreateInstance` does **NOT** insert or query `provider_resource_mappings` table.
- **Provider ID Storage**: `provider_instance_id` is stored directly on the `compute_instances` row.
- **Restart Recovery Risk**: After a gateway restart, `LocalDockerComputeProvider` loses its in-memory `p.instances` map. Although the `compute_instances` database row retains `provider_instance_id`, `LocalDockerComputeProvider.GetInstance` searches `p.instances` map and returns `instance not found in local provider`.

---

## 10. Lifecycle State Machine

| Operation | Control Plane (`compute_instances`) | Provider (`LocalDockerComputeProvider`) | Durable | Survives Gateway Restart |
|:---|:---:|:---:|:---:|:---:|
| **Create** | `PROVISIONING` -> `RUNNING` (Saved to DB) | `docker run -d` / Memory Map | **YES** | **PARTIAL** (DB row persists, in-memory provider map lost) |
| **Start** | Updated to `RUNNING` | `docker start <container_id>` | **YES** | **NO** (Fails if provider map unpopulated) |
| **Stop** | Updated to `STOPPED` | `docker stop <container_id>` | **YES** | **NO** (Fails if provider map unpopulated) |
| **Restart** | Updated to `RUNNING` | `StopInstance` + `StartInstance` | **YES** | **NO** (Fails if provider map unpopulated) |
| **Delete** | Soft-deleted (`deleted_at = now()`) | `docker rm -f <container_id>` | **YES** | **YES** |

---

## 11. Persistence Boundary

```
[HTTP Request] -> ComputeHandler -> ComputeUseCase
                                       │
                    ┌──────────────────┴──────────────────┐
                    │                                     │
                    v                                     v
      PostgresComputeRepository              LocalDockerComputeProvider
      (Database: compute_instances)          (In-Memory Map + Docker CLI)
                    │                                     │
     DURABLE ACROSS PROCESS RESTART            EPHEMERAL IN-MEMORY MAP
```

---

## 12. Restart Behavior

1. **Gateway Process Restart**:
   - `compute_instances` metadata in PostgreSQL **SURVIVES**.
   - `LocalDockerComputeProvider.instances` in-memory map **IS ERASED**.
   - When calling `StartInstance` or `StopInstance` post-restart, `LocalDockerComputeProvider` fails with `instance not found in local provider` because `p.instances` map is empty.
2. **Docker Container Survival**:
   - Docker container on host machine **SURVIVES** (unless container was stopped/removed).

---

## 13. Render / Production Compatibility

- **Docker Socket Dependency**: `LocalDockerComputeProvider` relies on local `exec.CommandContext("docker", ...)` accessing host `/var/run/docker.sock` or Windows Docker named pipe.
- **Cloud Deployment Block**: On managed container platforms like Render, AWS Fargate, or GCP Cloud Run, `/var/run/docker.sock` is unavailable (no Docker-in-Docker permission).
- **Production Result**: On Render, `exec.LookPath("docker")` fails or container creation fails, causing compute provider to silently fall back to `local-sim-<id>` simulated metadata.

---

## 14. Tenant Isolation Audit

- **Repository Layer**: `PostgresComputeRepository` has `GetTenantScopedByID`, which checks `OrganizationID` and `ProjectID` and returns `TENANT_ISOLATION_VIOLATION`.
- **HTTP Handler Layer Gap**: `ComputeHandler` in `internal/compute/delivery/http/compute_handler.go` does **NOT** extract `TenantContext` from JWT context and does **NOT** call `GetTenantScopedByID`.
- **Vulnerability**: Tenant B can query or stop Tenant A's compute instance if Tenant B knows Tenant A's compute instance ID.

---

## 15. Secret Security

- **Environment Variables (`EnvVars`)**: Stored in `compute_instances.env_vars_json` as JSON string via GORM hooks.
- **Security Policies (`Security`)**: Stored in `compute_instances.security_json` as JSON string via GORM hooks.
- **Secret Encryption**: Environment variables are stored as **PLAIN TEXT JSON** in PostgreSQL database column `env_vars_json`.
- **API Leakage**: Plaintext environment variables are returned in API GET responses (`json:"envVars,omitempty"`).

---

## 16. Current Limitations

1. HTTP handlers do not enforce `TenantContext` isolation.
2. `LocalDockerComputeProvider` relies on an in-memory map that vanishes on gateway restart.
3. Web terminal in frontend `/console/compute` uses hardcoded mock strings instead of real execution API.
4. Client-side creation in frontend falls back to mock object when API fails.
5. Plaintext storage of `envVars` in PostgreSQL database.

---

## 17. Exact Data Loss Boundary

- **Control-Plane Metadata**: `compute_instances` database rows are persistent.
- **Provider State Loss**: Process restart wipes `LocalDockerComputeProvider.instances` in-memory map, causing subsequent `Start`, `Stop`, `Restart`, and `ExecuteCommand` calls to fail until provider state is re-hydrated from `compute_instances` or `provider_resource_mappings`.

---

## 18. Smallest Required Fix (Phase 70 Scope Preview)

1. **Enforce Tenant Context in Compute Handler**:
   Update `ComputeHandler` to extract `tc := security.GetTenantContext(r.Context())` and invoke `GetTenantScopedByID(ctx, tc.OrganizationID, tc.ProjectID, id)` for all instance operations.
2. **Re-hydrate Provider State or Recover from DB**:
   Update `LocalDockerComputeProvider.GetInstance` to recover container metadata from `compute_instances` repository when in-memory map lookup misses.
3. **Wire Web Terminal API**:
   Connect frontend Web Terminal to `POST /api/v1/compute/instances/{id}/execute`.

---

## 19. Verification Plan

1. Verify `PostgresComputeRepository` unit and integration tests pass cleanly.
2. Verify tenant isolation rejection on `/api/v1/compute/instances/{id}` subroutes.
3. Verify compute instance metadata persistence across gateway process restart.

---

```
COMPUTE WORKFLOW FORENSIC CLASSIFICATION:

CONTROL-PLANE PERSISTENCE:
    PASS (GORM compute_instances table auto-migrated & persistent in PostgreSQL)

COMPUTE DATA-PLANE REALITY:
    PARTIAL (Executes real host Docker containers when docker CLI present; falls back to simulation when absent)

COMPUTE RESOURCE PERSISTENCE:
    PARTIAL (Control-plane DB row persists; in-memory provider map lost on process restart)

PROVIDER RESOURCE MAPPING:
    FAIL (provider_resource_mappings table not integrated into compute workflow)

TENANT ISOLATION:
    FAIL (HTTP handler does not extract TenantContext; uses hardcoded defaults)

CROSS-PROJECT ISOLATION:
    FAIL (HTTP handler list & subroute endpoints omit tenant context checks)

SECRET SECURITY:
    NEEDS REVIEW (Environment variables stored in plain text JSON column env_vars_json)

GATEWAY RESTART RECOVERY:
    FAIL (In-memory provider map erased on process restart, breaking Start/Stop/Restart calls)

PRODUCTION COMPATIBILITY:
    FAIL (LocalDockerComputeProvider depends on host Docker socket unavailable on Render/Fargate)

FRONTEND AUTHORITY:
    PARTIAL (Attempts API call first, but falls back to mock client-side state and fake web terminal)

BLOCKING FINDINGS:
    1. HTTP handler (compute_handler.go) omits TenantContext extraction, allowing cross-tenant compute access.
    2. LocalDockerComputeProvider loses in-memory container state on gateway restart, causing lifecycle actions (Start/Stop/Restart) to fail post-restart.
    3. Frontend web terminal uses hardcoded mock strings instead of calling execution API endpoint.

SMALLEST NEXT FIX:
    Phase 70 — Enforce TenantContext isolation and Gateway Restart Recovery in internal/compute/.
```
