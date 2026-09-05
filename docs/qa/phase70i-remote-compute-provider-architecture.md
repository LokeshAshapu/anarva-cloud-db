# ANARVA Cloud Phase 70I — Remote Compute Provider Architecture Design

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Design Date**: August 24, 2026  
**Architect**: Principal Cloud Infrastructure Architect & Security Lead  
**Scope**: Control-Plane / Data-Plane Decoupling & Remote Compute Architecture Spec  
**Baseline Commit**: Phase 70H Audit Certified (`5d52dae` + 70A–70H)  

---

## 1. Executive Summary

Phase 70I establishes the official architectural specification for **ANARVA Remote Compute Provider (`RemoteComputeProvider`)**.

Phase 70H confirmed that `LocalDockerComputeProvider` requires direct access to `/var/run/docker.sock` and host `docker` CLI commands, making it incompatible with standard PaaS deployment environments like Render.

Phase 70I decouples the **ANARVA Gateway Control Plane** (running on Render) from the **Compute Data Plane** (running on dedicated remote worker infrastructure). The Gateway control plane handles authentication, tenant context authorization, database persistence, and secret envelope encryption, while remote worker nodes execute container workloads via an authenticated REST API.

---

## 2. Final Architecture Classification Matrix

```
PHASE 70I:

CONTROL/DATA PLANE SEPARATION:
    PASS (Control plane on Render decoupled from Remote Worker Data Plane)

REMOTE PROVIDER DESIGN:
    APPROVED (RemoteComputeProvider REST client spec complete)

GATEWAY DOCKER DEPENDENCY:
    REMOVED BY DESIGN (Gateway requires zero docker CLI or host docker.sock access)

GATEWAY DOCKER SOCKET:
    NOT REQUIRED (Gateway invokes HTTPS REST API on remote worker)

WORKLOAD DURABILITY:
    PASS (Workloads run on dedicated worker node; survive gateway restarts & redeployments)

PROVIDER MAPPING:
    PASS (provider_resource_mappings maps instance ID to remote workload ID)

SERVICE AUTHENTICATION:
    mTLS (Mutual TLS with X.509 certificates) or X-Anarva-Worker-Token Bearer Token

SECRET DELIVERY:
    mTLS Payload Transmission (Gateway decrypts secrets post-authorization and sends over TLS)

IDEMPOTENCY:
    PASS (Idempotency-Key header using anarva_resource_id prevents duplicate containers)

MULTI-TENANT BOUNDARY:
    HARDENED_CONTAINER (cgroups v2, cap_drop ALL, per-tenant virtual networks)

LOCAL PROVIDER:
    RETAINED (LocalDockerComputeProvider preserved for local development & regression testing)

RECOMMENDED V1:
    OPTION A: Dedicated ANARVA Compute Worker Node with Authenticated TLS REST API

PRODUCTION COMPUTE ARCHITECTURE:
    APPROVED

PHASE 70J:
    Implement RemoteComputeProvider client struct & Mock/Remote Worker API contracts.

BLOCKING FINDINGS:
    NO BLOCKING FINDINGS. Architectural design approved for implementation in Phase 70J.
```

---

## 3. Control-Plane / Data-Plane Decoupling

```
                     Internet / HTTP Clients
                                │
                                ▼
         ┌─────────────────────────────────────────────┐
         │ ANARVA Gateway (Control Plane on Render)    │
         │  - JWT Authentication & TenantContext       │
         │  - PostgreSQL Metadata (compute_instances)  │
         │  - Secret Envelope Encryption (pkg/crypto)  │
         │  - provider_resource_mappings               │
         └──────────────────────┬──────────────────────┘
                                │
                        Authenticated HTTPS / mTLS
                        Service-to-Service REST API
                                │
                                ▼
         ┌─────────────────────────────────────────────┐
         │ ANARVA Compute Worker Node (Data Plane)     │
         │  - Worker REST API Listener                │
         │  - Container Runtime (Containerd / Docker) │
         │  - Isolated Tenant Workload Containers      │
         └─────────────────────────────────────────────┘
```

---

## 4. Remote Worker API Contract Specification

The internal REST API exposed by the ANARVA Compute Worker Node:

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/v1/workloads` | Create container workload with cgroup limits & env vars |
| `GET` | `/v1/workloads/{workloadID}` | Inspect workload runtime state and health |
| `POST` | `/v1/workloads/{workloadID}/start` | Start stopped container workload |
| `POST` | `/v1/workloads/{workloadID}/stop` | Stop running container workload |
| `POST` | `/v1/workloads/{workloadID}/restart` | Cleanly restart container workload |
| `POST` | `/v1/workloads/{workloadID}/execute` | Execute command via stdin/stdout stream |
| `GET` | `/v1/workloads/{workloadID}/metrics` | Query cgroup CPU, memory, network telemetry |
| `DELETE`| `/v1/workloads/{workloadID}` | Terminate and destroy container workload |

---

## 5. Security & Trust Boundary

1. **Authentication & Authorization Boundary**:
   - The Gateway remains the sole authority for validating user JWTs, `TenantContext` (`org_id`, `project_id`), and instance ownership.
   - The Worker API requires service-to-service authentication via `mTLS` (Mutual TLS) or secret header token (`X-Anarva-Worker-Token`).
2. **Secret Delivery**:
   - `ComputeUseCase` decrypts `EnvVars` using `pkg/crypto/secret_cipher.go` ONLY AFTER validating tenant context authorization.
   - Decrypted secrets are transmitted to the Worker over encrypted TLS and injected into container memory. Plaintext secrets are **NEVER** stored in `provider_resource_mappings` or logged.
3. **Multi-Tenant Container Hardening**:
   - Drop all Linux capabilities (`--cap-drop=ALL`).
   - Disable root privilege escalation (`--no-new-privileges`).
   - Isolate tenant containers on dedicated bridge networks (`proj-<id>-net`).

---

## 6. Provider Selection Strategy

- **Development Mode (`ANARVA_ENV=development`)**:
  - Gateway uses `LocalDockerComputeProvider` for local testing on host Docker.
- **Production Mode (`ANARVA_ENV=production`)**:
  - Gateway uses `RemoteComputeProvider` connecting to `COMPUTE_WORKER_URL`.
  - Gateway runs on Render with **ZERO** host `/var/run/docker.sock` or `docker` CLI dependencies.

---

## 7. Next Steps (Phase 70J Target Scope)

1. Implement `RemoteComputeProvider` struct in `internal/compute/provider/remote_provider.go`.
2. Define configuration parameters (`COMPUTE_WORKER_URL`, `COMPUTE_WORKER_TOKEN`) in `pkg/config/config.go`.
3. Add provider selection switch in `cmd/gateway/main.go`.
4. Create test suite verifying provider contract & remote API translation.
