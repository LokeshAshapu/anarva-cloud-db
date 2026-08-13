# Anarva Cloud — Phase 15: Billing Architecture

## 1. Overview

The **Anarva Billing Service** provides a modular usage metering, quota enforcement, pricing catalog, cost estimation, and draft invoice engine for Anarva Cloud.

---

## 2. Reality Labels & Non-Billable Mandate

All usage metering, cost estimates, and draft invoices explicitly carry reality labels:

- `LOCAL_DEVELOPMENT_USAGE`: Local Docker container runtime and local storage.
- `NON_BILLABLE`: Commercial payment gateways (Stripe/Razorpay) are not connected.
- `SIMULATED_BILLING`: Projected accrued usage calculation.
- `NOT_BILLABLE (ESTIMATE)`: Pre-provisioning cost estimate calculation.

---

## 3. Atomic Quota Engine & Concurrency Locking

Before any provisioning operation succeeds, the **Atomic Quota Engine** checks organization and project resource limits under a `sync.Mutex` lock to prevent race conditions during concurrent provisioning requests:

- `compute.acu`: Maximum Anarva Compute Units allocated.
- `storage.capacity`: Maximum storage in GB.
- `database.count`: Maximum database instances.
- `network.vpc`: Maximum VPC networks.

---

## 4. Versioned Pricing Catalog (v1.0.0)

- **Compute**: `$0.025 / ACU-hour`
- **Database**: `$0.045 / Instance-hour`
- **Storage**: `$0.15 / GB-month`
- **Network Egress**: `$0.09 / GB` (10 GB Free Tier included)

Historical invoices remain immutable and preserve the exact pricing version (`v1.0.0`) active during that billing cycle.
