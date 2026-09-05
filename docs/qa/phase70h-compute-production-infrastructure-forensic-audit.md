# ANARVA Cloud Phase 70H — Compute Production Infrastructure Forensic Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 24, 2026  
**Auditor**: Principal Cloud Infrastructure Architect & Security Lead  
**Scope**: Compute Infrastructure Deployment Architecture, Docker Dependency Analysis & Production Readiness Audit  
**Baseline Commit**: Phase 70G Certified (`2a9b571` + 70A–70G)  

---

## 1. Executive Summary

A forensic infrastructure audit was conducted on ANARVA Compute (`LocalDockerComputeProvider`, `ComputeUseCase`, `ComputeHandler`, `cmd/gateway/main.go`) to evaluate its compatibility with production cloud deployment environments (such as Render, containerized PaaS platforms, or managed cloud services).

### Primary Audit Finding
`LocalDockerComputeProvider` relies directly on invoking the host `docker` CLI (`exec.Command("docker", ...)`) against a local host Docker daemon via `/var/run/docker.sock`. Standard production PaaS web services (like Render, Railway, or AWS App Runner) run in restricted sandbox microVMs that **DO NOT** mount `/var/run/docker.sock` or provide host Docker daemon access.

Consequently, running the gateway on Render causes `LocalDockerComputeProvider` to fall back to simulation mode (`docker-sim-...`), rendering real customer compute workload execution impossible on standard PaaS web services.

---

## 2. Infrastructure Audit Matrix

```
PHASE 70H:

COMPUTE PROVIDER:
    DEVELOPMENT_ONLY

LOCAL DOCKER:
    SUPPORTED (Requires host Docker CLI & active daemon)

RENDER COMPATIBILITY:
    FAIL (Render web services do not provide Docker socket or host daemon access)

DOCKER SOCKET:
    REQUIRED (LocalDockerComputeProvider requires /var/run/docker.sock for host container management)

DOCKER CLI:
    REQUIRED (LocalDockerComputeProvider executes 'docker run', 'docker inspect', 'docker exec', 'docker rm')

GATEWAY RESTART RECOVERY:
    PASS (Control-plane metadata persists in PostgreSQL; re-hydration works on persistent host Docker)

RENDER REDEPLOYMENT RECOVERY:
    NOT POSSIBLE (Redeploying gateway container on new Render host destroys local Docker container state)

RUNTIME PERSISTENCE:
    EPHEMERAL (Host Docker container state is tied to local host daemon lifetime)

TENANT RUNTIME ISOLATION:
    NEEDS HARDENING (Containers share host Docker bridge network and host kernel)

CONTAINER SECURITY:
    NEEDS HARDENING (Docker socket access gives root-equivalent host privileges)

MULTI-TENANT PRODUCTION:
    NOT READY (Single host Docker daemon lacks multi-tenant network/VM security boundaries)

PRODUCTION COMPUTE:
    NOT READY (Requires decoupled remote compute provider or cloud orchestration API)

SMALLEST NEXT ARCHITECTURE:
    Decoupled Remote Compute Provider Abstraction (Remote Worker API / AWS / Kubernetes) separating Gateway Control Plane from Compute Workload Execution.

BLOCKING FINDINGS:
    1. RENDER INCOMPATIBILITY: Standard PaaS container platforms (Render, Railway) do not mount /var/run/docker.sock or run a host Docker daemon, causing LocalDockerComputeProvider to fall back to simulation mode in production.
    2. HOST DAEMON COUPLING: Customer containers are created on the gateway's local host, making compute workloads ephemeral and lost during gateway container redeployments.
    3. DOCKER SOCKET PRIVILEGE RISK: Mounting /var/run/docker.sock into gateway container grants root-equivalent control over host node.
    4. SINGLE-HOST MULTI-TENANT NETWORK RISKS: Containers share default host Docker bridge network without strict per-tenant virtual network isolation.
```

---

## 3. Forensic Infrastructure & Deployment Analysis

```
User Create Compute Request (POST /api/v1/compute/instances)
       │
       v
ANARVA Gateway (Hosted on Render / Container PaaS)
       │
       v
LocalDockerComputeProvider (`exec.LookPath("docker")`)
       │
       ├───────────────────────────────┐
       │ (If Host Docker Available)    │ (In Render / PaaS Environment)
       v                               v
Access /var/run/docker.sock     Docker Socket Missing / Permission Denied
       │                               │
       v                               v
Execute CLI Commands            Fallback to Simulation Mode
(`docker run`, `docker exec`)   (`docker-sim-<id>`) — NO REAL WORKLOAD
```

### 3.1 Docker Dependency & Command Map
`LocalDockerComputeProvider` (`internal/compute/provider/provider.go`) executes the following system CLI commands:
- `docker run -d --name anarva-acu-<slug> --cpus <cpus> --memory <mem> <image>`: Container provisioning.
- `docker inspect --format {{.State.Status}} <containerID>`: Status check and restart re-hydration.
- `docker stop <containerID>`: Halting container execution.
- `docker start <containerID>`: Resuming container execution.
- `docker exec <containerID> sh -c <cmd>`: Command execution for Web Terminal.
- `docker rm -f <containerID>`: Terminating and destroying container.

### 3.2 Render & Managed Container PaaS Compatibility
- Managed container platforms (Render, Railway, Heroku) run application code inside unprivileged containers without Docker socket forwarding (`/var/run/docker.sock`).
- Without `/var/run/docker.sock`, the `docker` CLI cannot communicate with a daemon, triggering fallback to `docker-sim-...`. Real container creation is blocked on Render.

### 3.3 Gateway Redeployment & Host Coupling
- On Render or Kubernetes, redeploying the gateway creates a new gateway container on a new host instance.
- Because `LocalDockerComputeProvider` provisions container workloads on the gateway's immediate local host daemon, redeploying the gateway leaves containers behind on the old host or loses runtime state entirely.

### 3.4 Multi-Tenant Security & Network Isolation Boundary
- Operating multiple customer containers on a single host Docker daemon shares the host Linux kernel and standard Docker bridge network (`172.17.0.0/16`).
- Without custom per-tenant overlay networks, VPC isolation, or dedicated microVM sandboxes (e.g. AWS Firecracker / gVisor), containers on the same host Docker daemon can scan host network interfaces or communicate across container boundaries.

---

## 4. Provider Classification

`LocalDockerComputeProvider` is classified as **`DEVELOPMENT_ONLY`**.

**Justification**:
1. Requires local host Docker daemon access (`/var/run/docker.sock`).
2. Couples customer compute workloads directly to the gateway host filesystem and process lifecycle.
3. Cannot execute real container workloads inside standard managed PaaS platforms like Render.

---

## 5. Smallest Safe Production Architecture Recommendation

To transition ANARVA Compute from local development to production readiness on Render:

1. **Control-Plane / Data-Plane Decoupling**:
   - The Gateway on Render remains the central control-plane API router and PostgreSQL metadata manager.
   - Compute workload execution is delegated to dedicated, remote Compute Worker Nodes or cloud orchestration providers.
2. **Target Production Provider Abstractions**:
   - **Remote Worker Agent**: Gateway communicates via TLS-authenticated gRPC/REST API with dedicated worker instances running Docker/Containerd.
   - **AWS Provider (`AWSComputeProvider`)**: Provisions EC2 instances or ECS tasks via AWS SDK.
   - **Kubernetes Provider (`KubernetesComputeProvider`)**: Provisions pods / CRDs via Kubernetes API (`client-go`).

---

## 6. Strict Compliance Acknowledgment
- **No Code Modifications**: Zero code changes were made during Phase 70H.
- **No Infrastructure Created**: No cloud SDKs or remote worker agents were installed.
- **No Commits/Pushes**: No git commit or push commands were executed.
- **Scope Integrity**: Maintained 100% read-only infrastructure forensic audit boundaries.
