# ANARVA CLOUD — RELIABILITY & OPERATION ENGINE

This document details the architecture, models, and failure recovery mechanisms of the **Anarva Reliability & Operation Engine**.

---

## 🏛️ Centralized Control-Plane Architecture

```
                    ANARVA
                       │
                Anarva API v1
                       │
                Authentication & IAM
                       │
                Tenant Rate Limit
                       │
             Anarva Control Plane
                       │
           Anarva Operation Engine
                       │
   ┌───────────────────┼───────────────────┐
   │                   │                   │
Idempotency         Lock Lease        Pre-Quota Check
   │                   │                   │
   └───────────────────┼───────────────────┘
                       │
              Provisioning Engine
                       │
              Observation Engine
                       │
                 Drift Detection
```

---

## ⚡ Key Capabilities

### 1. Persistent Operation Model & State Machine
- **States**: `QUEUED`, `RUNNING`, `SUCCEEDED`, `FAILED`, `CANCELLED`, `TIMED_OUT`.
- Every operation includes `id`, `organization_id`, `project_id`, `resource_id`, `operation_type`, `progress`, `request_id`, `idempotency_key`, and an immutable `timeline`.

### 2. Idempotency (`Idempotency-Key`)
- API endpoints support `Idempotency-Key` headers.
- Identical requests with the same key within a 24-hour window return the original operation result without duplicating cloud infrastructure.
- Reusing an idempotency key with a different request payload returns `IDEMPOTENCY_KEY_REUSE` conflict error.

### 3. Lease-Based Concurrency Lock
- Resources acquire a 5-minute lease lock during operation execution.
- If a process crashes, lock leases automatically expire to prevent permanent resource blocking.

### 4. Restart-Safe Operation Recovery
- Upon backend restart, `ReconcileInterruptedOperations` loads unfinished `RUNNING` operations, compares intent against actual observed cloud state, and safely finalizes operation state.

### 5. Pre-Provisioning Quota Reservation
- Quotas for ACUs, Database count, and Storage capacity are validated and reserved BEFORE cloud provisioning begins to prevent oversubscription.

### 6. Append-Only Audit Logging
- Security-sensitive actions append immutable audit events with actor details (`USER`, `API_KEY`, `SYSTEM`) while automatically redacting credential secrets (`[REDACTED_API_KEY]`).
