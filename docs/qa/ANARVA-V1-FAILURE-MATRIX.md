# ANARVA Cloud V1 — Failure, Idempotency & Recovery Audit

**Audit Date**: August 19, 2026  
**Auditor**: Distributed Systems & Reliability Engineer  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  

---

## 1. Reliability Architecture Overview

ANARVA incorporates a distributed operations framework (`internal/reliability/`) designed to guarantee operation tracking, distributed resource lock leases, idempotency key validation, and automated background recovery for stale operations.

---

## 2. Failure & Resilience Matrix

| Failure Scenario | Engine Component | System Behavior | Idempotency Protection | State Recovery Mechanism | Reliability Rating | Source Code Evidence |
|:---|:---|:---|:---|:---|:---|:---|
| **Database Disconnection** | `pkg/database/postgres.go` | Health check reports `DATABASE_UNAVAILABLE`; fails closed in production | N/A | Automated connection pool retry | 🟢 SAFE | `postgres.go` |
| **Duplicate Operation Submit**| `internal/reliability/usecase/reliability_usecase.go` | Checks `idempotency_records` table by key | Idempotency Key Match | Returns cached initial response | 🟢 SAFE | `reliability_usecase.go` |
| **Concurrent Resource Modify** | `internal/reliability/repository/reliability_repository.go` | Acquires lease in `resource_lock_leases` | Lease Exclusivity | Rejects concurrent lock with error | 🟢 SAFE | `reliability_repository.go` |
| **Gateway Process Crash** | `cmd/gateway/main.go:507` | Gateway restarts; recovery worker scans `anarva_operations` | Operation ID Tracking | Recovery worker marks stale ops as `FAILED` / `RECOVERED` | 🟢 SAFE | `recovery_worker.go` |
| **Docker Daemon Unavailable** | `internal/postgres/provider/docker_provider.go` | Provisioning returns error; operation marked `FAILED` | Operation State Update | Rollback operation status in DB | 🟡 PARTIAL | `docker_provider.go` |
| **Storage Upload Interrupted**| `internal/storage/provider/local_storage_provider.go` | File write fails; partial file may remain on server disk | None | Requires manual retry | 🔴 UNSAFE | `local_storage_provider.go` |
| **Network Timeout During Deploy**| `internal/networking/service/networking_service.go` | State set to `PROVISIONING_FAILED` | Transaction Rollback | Re-trigger provisioning job | 🟢 SAFE | `networking_service.go` |

---

## 3. Distributed Recovery Worker Audit

- **Implementation**: `internal/reliability/usecase/recovery_worker.go`
- **Execution Loop**: Periodically scans `anarva_operations` for operations stuck in `IN_PROGRESS` state beyond the configured threshold (e.g. 5 minutes).
- **Behavior**: Automatically transitions stale operations to `RECOVERY_FAILED` or re-enqueues them for reconciliation.
- **Evidence**: Initialized in `cmd/gateway/main.go` line 507 (`recWorker.Start(context.Background())`).
