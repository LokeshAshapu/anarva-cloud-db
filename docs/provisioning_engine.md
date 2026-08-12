# Anarva Cloud — Phase 13: Infrastructure Provisioning Engine Architecture

## 1. Overview

The **Anarva Infrastructure Provisioning Engine** is the centralized control-plane orchestrator responsible for translating high-level control-plane resource requests (`DATABASE`, `COMPUTE`, `STORAGE`, `NETWORK`, `SUBNET`, `VOLUME`, `LOAD_BALANCER`, `DNS`, `BACKUP`) into validated, step-by-step infrastructure operations executed across infrastructure providers.

It enforces strict multi-tenant isolation, idempotency tracking, operation concurrency locks, automated rollback on failure, drift detection, and state reconciliation.

---

## 2. Infrastructure Reality Classification

To maintain complete architectural honesty, every provider, capability, and operation in the Provisioning Engine is explicitly categorized with a reality status label:

| Reality Status | Description |
| :--- | :--- |
| **REAL** | Real physical/logical resources managed directly by real driver APIs. |
| **LOCAL DEVELOPMENT PROVIDER** | Docker container tasks spawned with cgroup CPU/Memory limits, local PostgreSQL databases, and local file storage. |
| **CONFIGURED** | Provider configurations registered and validated in control plane state. |
| **PROVIDER NOT CONNECTED** | Driver interfaces registered but remote API endpoints not linked. |
| **SIMULATED** | Control plane state machine simulating provider operations. |
| **PLANNED** | Future cloud drivers (AWS EC2, GCP Compute Engine, Kubernetes CRDs). |

---

## 3. Provisioning Request State Machine

```
   [REQUESTED] ──► [VALIDATING] ──► [PLANNING] ──► [QUEUED] 
                                                        │
   ┌────────────────────────────────────────────────────┘
   ▼
[PROVISIONING] ──► [CONFIGURING] ──► [VERIFYING] ──► [COMPLETED]
       │                 │                │
       └─────────────────┴────────────────┴──► [ROLLING_BACK] ──► [ROLLED_BACK]
```

- **REQUESTED**: Request received with `Idempotency-Key` and IAM authorization.
- **VALIDATING**: Organization, project, region capacity bounds, and provider capabilities checked.
- **PLANNING**: Step-by-step Execution Plan generated with estimated action steps.
- **QUEUED**: Resource lock acquired (`LOCKED`).
- **PROVISIONING**: Provider task initiated (`docker run --cpus=1.0 --memory=2g ...`).
- **CONFIGURING**: Networking bridge attachment, security group rules, and NVMe mounts applied.
- **VERIFYING**: Health check verified.
- **COMPLETED**: Resource marked available.
- **ROLLING_BACK**: Automatic reverse teardown executed if any step fails.
- **ROLLED_BACK**: Teardown complete, lock released.

---

## 4. Docker Provider & Security Model

The `DockerInfrastructureProvider` spawns container tasks safely without compromising host security:
- **No Socket Exposure**: Docker daemon sockets (`/var/run/docker.sock`, `npipe://`) are **never** mounted or exposed to user workloads.
- **No Privileged Containers**: `--privileged` flag is strictly forbidden.
- **Cgroup Limits**: CPU (`--cpus=1.0`) and Memory (`--memory=2g`) restrictions enforced on every task.
- **Network Isolation**: Tasks bind to isolated Docker bridge networks.

---

## 5. Drift Detection & State Reconciliation

The `ResourceReconciliationService` periodically compares the control-plane state with provider execution state:
- **IN_SYNC**: Control-plane state matches provider execution state 100%.
- **DRIFTED**: Mismatch detected (e.g. container stopped externally).
- **MISSING**: Provider resource missing.
- **UNKNOWN**: Provider unreachable.

---

## 6. REST API Endpoints

- `POST /api/v1/provisioning/plan`: Generate execution plan preview.
- `POST /api/v1/provisioning/apply`: Execute provisioning request pipeline.
- `GET /api/v1/provisioning/requests`: List active & historical provisioning requests.
- `GET /api/v1/provisioning/requests/:id`: Get detailed execution plan and step progress.
- `GET /api/v1/providers`: List registered infrastructure providers & capabilities.
- `GET /api/v1/resources/:id/drift`: Inspect resource drift status.
- `POST /api/v1/resources/:id/reconcile`: Trigger state reconciliation check.
