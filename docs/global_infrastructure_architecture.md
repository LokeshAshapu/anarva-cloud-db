# Anarva Cloud Multi-Region & HA Control Plane Architecture (Phase 21)

## Overview
Anarva Cloud Phase 21 introduces the **Global Infrastructure Orchestration Layer** under `internal/infrastructure/...`. It provides provider-neutral multi-region orchestration, Availability Zones, Data Residency enforcement, Infrastructure Health aggregation, Failover Engine with split-brain protection, and a safe Region Outage Simulator (`LOCAL_SIMULATION`).

---

## Core Infrastructure Subpackages (`internal/infrastructure/...`)
- **`region` & `zone`**: Region (`ACTIVE`, `DEGRADED`, `MAINTENANCE`, `UNAVAILABLE`) and Availability Zone models (`ap-hyderabad-1a`, `us-east-1a`).
- **`placement`**: `PlacementEngine` validating region, zone, quota, network compatibility, and data residency policies.
- **`health`**: `InfrastructureHealthEngine` evaluating overall global health (`HEALTHY`, `DEGRADED`, `PARTIAL_OUTAGE`, `MAJOR_OUTAGE`).
- **`failover`**: `FailoverEngine` executing automated or manual failover with distributed generation locks to eliminate split-brain conditions.
- **`evacuation`**: `EvacuationService` orchestrating regional workload migrations.
- **`simulator`**: Safe development-only region outage simulator (`LOCAL_SIMULATION`).

---

## Split-Brain Protection Architecture (`internal/infrastructure/failover/failover_engine.go`)
- **Generation Lock Locking**: Each failover policy maintains a strictly increasing `generationLock` counter.
- **Atomic Rejection**: Stale attempts with `generationLock <= currentLock` are rejected instantly with `SPLIT-BRAIN BLOCKED`.

---

## Provider Reality Labels
- **Local Simulation**: `LOCAL_SIMULATION (LIMITED_CAPABILITIES)`
- **Connected Cloud Region**: `CONNECTED`
- **Unconfigured**: `NOT_CONFIGURED`
