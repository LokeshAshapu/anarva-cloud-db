# Anarva Cloud Platform — Reliability Model

## Core Principles
1. **Deterministic State Transitions**: Operations advance strictly along valid transitions: `PENDING -> CREATED -> RUNNING -> COMPLETED` or `RUNNING -> FAILED`. Invalid state jumps are rejected server-side.
2. **Distributed Resource Locks**: Locks are tied to `(resource_id, owner_id)` with explicit lease durations.
3. **Lease Renewal & Expiration**: Active workers renew lease locks periodically (`RenewResourceLock`). Expired leases are reclaimed safely without human intervention.
4. **Persistent Idempotency**: Duplicate requests bearing the same `Idempotency-Key` header return the existing operation state without re-executing side-effects.

---

## State Machine Diagram

```
       [ Client Request ]
               │
               ▼
          [ CREATED ]
               │
         (Acquire Lock)
               │
               ▼
          [ RUNNING ] ◄── (Lease Renewed Periodically)
          /         \
   (Success)       (Error / Timeout / Interrupt)
        /             \
       ▼               ▼
 [ COMPLETED ]     [ FAILED ] ──► (Recovery Worker Reconciles)
```
