# ANARVA Cloud Phase 70O — Production Compute Worker Deployment Readiness Audit Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 25, 2026  
**Lead Auditor**: Principal Cloud Infrastructure Architect & Security Lead  
**Scope**: Read-Only Forensic Audit for Production Linux Compute Worker Deployment Readiness  
**Baseline Commit**: Phase 70N End-to-End Certified (`5d52dae` + 70A–70N)  

---

## 1. Executive Summary

Phase 70O performs a strict, read-only forensic audit of the current **ANARVA Compute Worker** implementation to determine whether it is ready for deployment on a production Linux VM host.

While Phases 70M and 70N certified the core data-plane foundation, mTLS transport security, REST API contract, secret redaction, and local container execution, deploying to a real cloud infrastructure environment requires resolving explicit operational and security requirements.

**Final Scorecard Summary**:
- **SECURITY**: `NOT READY` (Requires non-root worker process isolation and metadata file permissions hardening)
- **NETWORK**: `NOT READY` (Requires Cloud Metadata Endpoint `169.254.169.254` firewall blocking)
- **SECRETS**: `READY` (AES-256-GCM encrypted at rest in Gateway, in-memory delivery over mTLS, zero persistence)
- **RUNTIME**: `READY` (DockerContainerRuntime with `cap_drop=ALL`, `no-new-privileges`, and cgroup quotas)
- **RESOURCE GOVERNANCE**: `NOT READY` (Requires `--pids-limit` and execution rate limiting)
- **OBSERVABILITY**: `NOT READY` (Requires Prometheus metric exporter `/metrics` endpoints)
- **DISASTER RECOVERY**: `NOT READY` (Requires SQLite WAL atomic file locking & automated backup sequence)
- **DEPLOYMENT**: `NOT READY` (Requires systemd unit template, automated certificate rotation, and installer script)
- **MULTI-WORKER**: `NOT READY` (Requires control-plane node scheduling & worker health heartbeats)
- **OVERALL SCORECARD**: **`NOT READY`**

---

## 2. Current Architecture Topology

```
                  CONTROL PLANE (Render)

      ┌──────────────────────────────────────────────┐
      │               ANARVA GATEWAY                 │
      │                                              │
      │ • Authoritative TenantContext Authorization  │
      │ • Managed PostgreSQL Database               │
      │ • AES-256-GCM Envelope Secret Encryption    │
      │ • RemoteComputeProvider HTTP Client          │
      └──────────────────────┬───────────────────────┘
                             │
                  mTLS + Service Token
                   Port 8443 / HTTPS
                             │
                             ▼

                  DATA PLANE (Dedicated VM)

      ┌──────────────────────────────────────────────┐
      │            ANARVA COMPUTE WORKER             │
      │                                              │
      │ • Worker REST API Server (`WorkerServer`)    │
      │ • Local Metadata Store (`FileMetadataStore`) │
      │ • Container Runtime (`DockerContainerRuntime`)│
      └──────────────────────┬───────────────────────┘
                             │
                             ▼
                     Docker Engine / Socket
                             │
                             ▼
                    Customer Containers
```

---

## 3. Host Requirements & Implicit Assumptions (Step 2)

### 3.1 Host Environment Requirements:
- **Operating System**: Linux (Ubuntu 24.04 LTS / Debian 12 / Rocky Linux 9).
- **Kernel Version**: Linux Kernel $\ge 5.15$ with cgroups v2 enabled (`systemd.unified_cgroup_hierarchy=1`).
- **Container Engine**: Docker Engine $\ge 24.0.0$ or `containerd` $\ge 1.7.0$.
- **Ports**: Inbound TCP Port 8443 (Worker REST API listening on mTLS).
- **Storage**: Dedicated NVMe volume mounted at `/var/lib/anarva-worker/` for container layers and metadata storage.

### 3.2 Implicit Host Assumptions (Code Inspection):
1. `exec.LookPath("docker")`: Assumes `docker` CLI binary is present in host `PATH`. [STATICALLY VERIFIED]
2. `/var/run/docker.sock`: Assumes default Linux Docker Unix domain socket path. [STATICALLY VERIFIED]
3. File Store: Defaults to local path `worker_data.json` if `WORKER_DB_PATH` is unspecified. [STATICALLY VERIFIED]

---

## 4. Privilege Model & Host Security Audit (Step 3)

1. **Worker Process Privileges**:
   - Currently, if the Worker binary is launched as `root`, container management operates directly.
   - **Production Requirement**: The Worker daemon MUST run as a dedicated, unprivileged system user `anarva-worker` assigned to the `docker` group.
2. **Host Socket Isolation**:
   - Customer workloads MUST NEVER mount `/var/run/docker.sock`. [PREVIOUSLY VERIFIED in 70M/70N]
   - The Worker process holds socket access, but API endpoints strictly forbid arbitrary volume mounts or socket passthrough.

---

## 5. Docker Runtime & Container Sandboxing Audit (Step 4)

| Container Security Control | Current Implementation | Production Classification |
| :--- | :--- | :--- |
| `--cap-drop=ALL` | Enforced in `runtime.go` | `VERIFIED` |
| `--security-opt=no-new-privileges:true` | Enforced in `runtime.go` | `VERIFIED` |
| Privileged Mode | Disabled (`privileged = false`) | `VERIFIED` |
| Host Network Sharing | Disabled (Uses custom bridge `anarva-project-<id>-net`) | `VERIFIED` |
| Host PID/IPC Sharing | Disabled | `VERIFIED` |
| CPU Quotas | Enforced (`--cpus`) | `VERIFIED` |
| Memory Limits | Enforced (`--memory`) | `VERIFIED` |
| Process Limits (`--pids-limit`) | Not configured in `runtime.go` | `MISSING` (Blocker against fork-bombs) |
| Read-Only Root Filesystem | Optional | `PARTIAL` |

---

## 6. Network Security & SSRF Prevention (Step 5)

1. **Ingress Filtering**:
   - Inbound TCP 8443 MUST be restricted by cloud firewall / security group to Gateway IP addresses ONLY.
2. **Egress & Cloud Metadata Protection**:
   - Customer containers running on cloud VMs (AWS EC2 / GCP Compute) could attempt SSRF to `169.254.169.254` (Instance Metadata Service).
   - **Production Requirement**: Worker MUST insert host `iptables` / `nftables` rules blocking container bridge traffic to `169.254.169.254` and host loopback `127.0.0.1:8443`.

---

## 7. mTLS & Worker Token Security Review (Steps 6 & 7)

- **mTLS Validation**: `BuildTLSConfig` enforces `MinVersion = tls.VersionTLS12` and `InsecureSkipVerify = false`. [STATICALLY VERIFIED]
- **Dual-Layer Authentication**: `X-Anarva-Worker-Token` header + mTLS client certificate validation. [PREVIOUSLY VERIFIED]
- **Certificate Rotation**: Production deployment requires automated zero-downtime certificate reloading (via SIGHUP or in-memory dynamic cert reloader).

---

## 8. Secret Lifecycle & Non-Persistence Audit (Step 8)

- **In-Memory Transport**: Plaintext secrets (`EnvVars`) are decrypted in Gateway memory and sent over mTLS.
- **Worker Non-Persistence**:
  - `worker_metadata.json`: **Secrets ABSENT**. [VERIFIED in 70N]
  - Container Labels (`com.anarva.*`): **Secrets ABSENT**. [VERIFIED in 70M]
  - GET API Responses: **Redacted (`EnvVars = nil`)**. [VERIFIED in 70N]

---

## 9. Resource Governance & Anti-Exhaustion (Step 10)

| Risk Factor | Defense Mechanism | Status |
| :--- | :--- | :--- |
| CPU Starvation | cgroups v2 `--cpus` | `VERIFIED` |
| Memory Exhaustion | cgroups v2 `--memory` | `VERIFIED` |
| Fork Bomb | Linux `--pids-limit 512` | `MISSING` |
| Large Command Output | 1MB `LimitReader` and buffer cap | `VERIFIED` |
| Concurrent Execution Exhaustion | Rate limiter /Semaphore for `docker exec` | `MISSING` |

---

## 10. Multi-Worker Architecture Audit (Step 15)

The current `provider_resource_mappings` table tracks `ProviderResourceID` and `Region`.

To support **Multi-Worker Clusters** (Worker A, Worker B, Worker C):
1. Gateway requires a `compute_worker_nodes` catalog table (`id`, `endpoint`, `status`, `available_acu`, `heartbeat_at`).
2. Placement Engine in `ComputeUseCase` to select worker nodes based on capacity and region.

---

## 11. Threat Model & Risk Matrix (Step 24)

| Threat | Attack Vector | Current Defense | Residual Risk | Required Mitigation |
| :--- | :--- | :--- | :--- | :--- |
| Container Escape | Compromised workload | `--cap-drop=ALL`, `--no-new-privileges` | Zero-day kernel vulnerability | Seccomp default profile + gVisor/Kata option |
| Cloud Metadata SSRF | Workload querying `169.254.169.254` | None on bridge | Token theft via IAM role | `iptables` drop rule for `169.254.169.254` |
| Fork Bomb | Workload spawning infinite sub-processes | None | Host PID exhaustion | Set `--pids-limit 512` on `docker run` |
| Metadata File Corruption | Worker crash during JSON write | `FileMetadataStore` overwrite | Partial write on power loss | Atomic write (`.tmp` + rename) or SQLite WAL |

---

## 12. Production Readiness Scorecard (Step 25)

```
SECURITY:            NOT READY
NETWORK:             NOT READY
SECRETS:             READY
RUNTIME:             READY
RESOURCE GOVERNANCE: NOT READY
OBSERVABILITY:       NOT READY
DISASTER RECOVERY:   NOT READY
DEPLOYMENT:          NOT READY
MULTI-WORKER:        NOT READY

OVERALL SCORECARD:   NOT READY
```

---

## 13. Blocking vs. Non-Blocking Findings

### Blocking Findings (Must be resolved before Phase 70P production VM deployment):
1. **PIDs Limit Missing**: `runtime.go` must add `--pids-limit 512` to prevent container fork-bombs.
2. **Cloud Metadata Firewall**: Host iptables rule required to block container traffic to `169.254.169.254`.
3. **Atomic File Storage**: `FileMetadataStore` must use atomic write-rename pattern (`.tmp` $\to$ target) or SQLite WAL mode to protect against power-loss corruption.
4. **Non-Root Execution**: Worker systemd unit template must specify unprivileged `User=anarva-worker`.

### Non-Blocking Findings:
1. Multi-Worker scheduling table (`compute_worker_nodes`) for multi-node clusters.
2. Prometheus `/metrics` endpoint for operational telemetry.

---

## 14. Recommended Phase 70P Scope

**Phase 70P — Production Compute Worker Host Installer & Security Hardening**:
1. Implement `--pids-limit` and atomic file write pattern in `internal/worker/`.
2. Create `scripts/install-worker.sh` provisioning `anarva-worker` user, systemd unit, and `iptables` metadata blocking rules.
3. Build production deployment certification suite.
