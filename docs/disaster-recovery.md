# Anarva Cloud Platform — Disaster Recovery Architecture

## Executive Summary
This document describes the Disaster Recovery (DR) architecture, Recovery Time Objectives (RTO), Recovery Point Objectives (RPO), and control-plane resilience guarantees for the Anarva Cloud Platform.

---

## Disaster Recovery Targets
- **Control-Plane RTO**: < 30 seconds (Automatic Recovery Worker daemon reconciliation on restart).
- **Control-Plane RPO**: 0 seconds for persisted operation state and database transactions; < 5 seconds for local ephemeral logs.
- **Tenant Data Isolation RTO**: 0 downtime (Multi-tenant resource locks prevent cross-tenant state corruption).

---

## Disaster Recovery Architecture

```
[ Incoming Requests ] 
          │
          ▼
[ Anarva API Gateway ] ◄── Rate Limit & Security Headers Middleware
          │
          ▼
[ Reliability Engine ] ◄── Lease Expiration Scanner & Idempotency Store
          │
    ┌─────┴─────┐
    ▼           ▼
[ PostgreSQL ] [ Resource Lock Engine ]
```

---

## Disaster Recovery Guarantees
1. **Persistent Operation State**: Every operation state change (`CREATED`, `RUNNING`, `COMPLETED`, `FAILED`) is committed to the control-plane PostgreSQL repository before side-effects are dispatched.
2. **Idempotent Recovery**: When the gateway or recovery worker restarts, pending operations are reconciled idempotently using their unique `idempotency_key`. Duplicate side-effects are strictly prevented.
3. **Lease Expiration & Release**: Resource locks carry explicit lease durations. If an instance crashes while holding a lock, the lease expires and the lock is freed for subsequent operations.
4. **Tenant Isolation Boundary**: All recovery operations strictly filter by `organization_id` to guarantee tenant isolation during disaster recovery.
