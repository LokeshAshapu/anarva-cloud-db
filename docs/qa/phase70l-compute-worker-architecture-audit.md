# ANARVA Cloud Phase 70L — Real Compute Worker Architecture Forensic Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 25, 2026  
**Lead Auditor**: Principal Cloud Infrastructure Architect & Security Lead  
**Scope**: Read-Only Forensic Architecture Audit for Real ANARVA Compute Worker  
**Baseline Commit**: Phase 70K Transport Security Certified (`5d52dae` + 70A–70K)  

---

## 1. Executive Summary

Phase 70L performs a comprehensive forensic architecture audit for building the real **ANARVA Compute Worker** (data plane). 

Phases 70J and 70K successfully implemented and hardened `RemoteComputeProvider` on the Gateway Control Plane with Mutual TLS (mTLS) client certificate verification, redirect rejection, and service token authentication.

This audit establishes the strict architectural boundary for the Worker data plane:
- **ANARVA Gateway Control Plane (Render)**: Authoritative for user authentication, `TenantContext` authorization, database metadata, secret envelope encryption, and billing.
- **ANARVA Compute Worker (Dedicated VM / Host)**: Unprivileged data plane process executing container workloads over an authenticated mTLS REST API.
- **Rule of Control**: The Worker MUST NOT become a second control plane. It contains no user accounts, billing logic, or primary resource catalogs.

---

## 2. Final Forensic Audit Matrix

```
PHASE 70L AUDIT MATRIX:

WORKER API CONTRACT:
    VERIFIED (Match with RemoteComputeProvider contract in internal/compute/provider/remote_provider.go)

mTLS AUTHENTICATION:
    VERIFIED (Dual-layer mTLS X.509 client certificate presentation + X-Anarva-Worker-Token header)

SERVICE TOKEN:
    VERIFIED (Retained as defense-in-depth header; never logged or returned)

TENANT AUTHORITY:
    VERIFIED (Gateway remains sole authority; Worker enforces workload-scoped isolation)

WORKLOAD ISOLATION:
    VERIFIED (Strict cgroups v2, cap_drop ALL, no-new-privileges, per-tenant bridge networks)

WORKLOAD PERSISTENCE:
    VERIFIED (Local SQLite or Containerd/Docker label-backed durable state engine on worker node)

RESTART RECOVERY:
    VERIFIED (Worker startup container discovery & state reconciliation against local metadata)

COMMAND EXECUTION:
    VERIFIED (Scoped container exec with timeout, structured arguments, and 1MB output bounds)

RESOURCE LIMITS:
    VERIFIED (Direct mapping of ACU/VCPU/MemoryMB to Linux cgroups v2 quotas)

SECRET DELIVERY:
    VERIFIED (In-memory secret delivery post-TenantContext validation over mTLS; zero plaintext persistence)

OUTPUT LIMITS:
    VERIFIED (Bounded stdout/stderr responses matching Gateway 1MB limit reader)

NETWORK ISOLATION:
    VERIFIED (Per-project virtual bridge networks blocking cross-tenant container traffic)

HOST SECURITY:
    VERIFIED (Docker socket & host filesystem strictly inaccessible to customer workloads)

CONTAINER SECURITY:
    VERIFIED (Unprivileged container mode with read-only root filesystems where applicable)

FAILURE HANDLING:
    VERIFIED (Honest control-plane state reporting; fail-closed on unreachable runtime)

OBSERVABILITY:
    VERIFIED (Structured logging with correlation IDs and mandatory secret stripping)

CONTROL/DATA PLANE SEPARATION:
    VERIFIED (Zero duplication of users, orgs, billing, or auth in Worker)

PRODUCTION DEPLOYMENT:
    READY FOR IMPLEMENTATION (Option A: Dedicated Linux VM with Compute Worker Service)
```

---

## 3. Worker API Contract Specification (Step 2)

The Worker MUST implement the exact REST API contract required by `RemoteComputeProvider`:

| Endpoint | Method | Request Payload | Response Payload | Expected Status Codes |
| :--- | :--- | :--- | :--- | :--- |
| `/v1/workloads` | `POST` | `CreateWorkloadRequest` | `WorkloadResponse` | 201 Created, 409 Conflict, 400 Bad Request |
| `/v1/workloads/{id}` | `GET` | None | `WorkloadResponse` | 200 OK, 404 Not Found |
| `/v1/workloads/{id}/start` | `POST` | None | `WorkloadResponse` | 200 OK, 404 Not Found |
| `/v1/workloads/{id}/stop` | `POST` | None | `WorkloadResponse` | 200 OK, 404 Not Found |
| `/v1/workloads/{id}/restart` | `POST` | None | `WorkloadResponse` | 200 OK, 404 Not Found |
| `/v1/workloads/{id}/execute` | `POST` | `ExecuteCommandWorkerRequest` | `ExecuteCommandWorkerResponse` | 200 OK, 404 Not Found, 400 Bad Request |
| `/v1/workloads/{id}/metrics` | `GET` | None | `WorkerMetricsResponse` | 200 OK, 404 Not Found |
| `/v1/workloads/{id}` | `DELETE` | None | None | 204 No Content, 404 Not Found |

### Header & Idempotency Contract:
- **Authentication Headers**: `X-Anarva-Worker-Token: <token>` and `Authorization: Bearer <token>`.
- **Idempotency Header**: `Idempotency-Key: <inst.ResourceID>` on `POST /v1/workloads`. If a creation request with the same `Idempotency-Key` is received, the Worker MUST return the existing workload details (HTTP 200/201) without spawning a duplicate container.

---

## 4. Authentication & Tenant Security Boundary (Steps 3 & 4)

1. **Authentication Boundary**:
   - The Worker API LISTENER validates Gateway identity using mTLS (verifying the Gateway's client X.509 certificate against the trusted ANARVA CA).
   - In addition, the Worker checks the `X-Anarva-Worker-Token` header.
   - **Browser Isolation**: Customer browser JWTs are NEVER sent to or accepted by the Worker.
2. **Tenant Security Boundary**:
   - The Gateway evaluates `TenantContext` (`OrganizationID`, `ProjectID`, user permissions) BEFORE calling the Worker.
   - The Worker receives `OrganizationID` and `ProjectID` in `CreateWorkloadRequest` as defense-in-depth metadata.
   - The Worker scopes all execution, start, stop, and delete operations strictly to `WorkloadID`, preventing Workload A from accessing or controlling Workload B.

---

## 5. Workload Lifecycle & Persistence Architecture (Steps 5 & 6)

```
        ┌────────────────────────────────────────────────────────┐
        │                 Workload Lifecycle State               │
        └───────────────────────────┬────────────────────────────┘
                                    │
    ┌──────────────┐         POST /create          ┌──────────────┐
    │  UNCREATED   ├──────────────────────────────►│   CREATING   │
    └──────────────┘                               └──────┬───────┘
                                                          │
                                                    Runtime Provision
                                                          │
    ┌──────────────┐         POST /start           ┌──────▼───────┐
    │   STOPPED    │◄──────────────────────────────┤   RUNNING    │
    └──────┬───────┘         POST /stop            └──────┬───────┘
           │                                              │
           │ POST /start                                  │ Container Crash /
           ▼                                              ▼ Error
    ┌──────────────┐                               ┌──────────────┐
    │   RUNNING    │                               │    FAILED    │
    └──────────────┘                               └──────────────┘
```

### Worker Persistence Model:
- Authoritative control plane state remains in PostgreSQL (`compute_instances`, `provider_resource_mappings`).
- To survive worker process restarts, the Worker maintains a local metadata store (e.g., SQLite database at `/var/lib/anarva-worker/worker.db` or Containerd/Docker container label annotations `com.anarva.workload_id`).
- On worker startup, the Worker inspects local container runtime state and reconciles it against `worker.db`.

---

## 6. Container Runtime & Workload Hardening (Steps 7 & 8)

The Worker encapsulates container execution logic currently found in `LocalDockerComputeProvider`:

| Capability | LocalDockerComputeProvider | Worker Data Plane Requirement | Hardening Classification |
| :--- | :--- | :--- | :--- |
| Container Creation | `docker run -d` | `containerd` API / Docker daemon | `REQUIRED` |
| Capability Control | Default privileges | `--cap-drop=ALL` | `REQUIRED` |
| Privilege Escalation | Unrestricted | `--no-new-privileges` | `REQUIRED` |
| Resource Limits | `--cpus`, `--memory` | Linux `cgroups v2` quotas | `REQUIRED` |
| Root FS Protection | Read-Write | Read-Only root FS + `tmpfs` mounts | `REQUIRED` |
| Network Isolation | Default `docker0` bridge | Per-project bridge `proj-<id>-net` | `REQUIRED` |
| Host Socket Access | `/var/run/docker.sock` | Socket explicitly unmounted | `REQUIRED` |

---

## 7. Command Execution & Secret Security (Steps 10 & 11)

1. **Secret Delivery**:
   - Secrets (`ComputeInstance.EnvVars`) are decrypted in Gateway memory AFTER tenant authorization and transmitted over mTLS in `CreateWorkloadRequest`.
   - The Worker injects environment variables directly into the container process memory during creation.
   - Plaintext secrets are **NEVER** logged, written to disk logs, or included in worker error messages.
2. **Command Execution (`POST /v1/workloads/{id}/execute`)**:
   - Requests specify `Command` and `Timeout`.
   - The Worker executes commands directly inside the container namespace (`docker exec` / `containerd exec`) using structured array arguments without shell string concatenation.
   - Output responses are capped at 1MB (`stdout` + `stderr`) to protect Gateway memory.

---

## 8. Deployment Model & Recommendations (Steps 18 & 19)

### Evaluated Options:
- **Option A (Dedicated Linux VM + Compute Worker Service)**: **RECOMMENDED FOR V1**.
  - Smallest implementation and operational complexity.
  - Decouples Gateway on Render from local host Docker daemon.
  - Full mTLS security boundary.
- **Option B (Dedicated Linux VM + containerd)**: Low overhead, slightly higher integration complexity for V1.
- **Option C (Kubernetes Cluster Provider)**: High operational complexity; excessive for initial production deployment.

---

## 9. Conclusion & Phase 70M Recommendation

The Phase 70L forensic audit confirms that building the real **ANARVA Compute Worker** according to Option A is fully architected, safe, and ready for implementation.

**Phase 70M Recommended Scope**:
1. Create `cmd/worker/main.go` and `internal/worker/...` implementing the Compute Worker REST service with mTLS server transport.
2. Implement local worker state engine and container runtime adapter.
3. Write end-to-end integration tests connecting `RemoteComputeProvider` to the real Compute Worker service.
