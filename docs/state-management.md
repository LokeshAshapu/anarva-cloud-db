# ANARVA CLOUD — INFRASTRUCTURE STATE & RECONCILIATION MODEL

This document details the **state management, observation, and reconciliation model** for Anarva Cloud resources managed via the Anarva Terraform Provider (`anarva/anarva`).

---

## 🔄 The 5-Layer State Hierarchy

To prevent state ambiguity and inconsistent drift handling, Anarva Cloud distinguishes between 5 distinct state layers:

1. **Desired State (HCL)**: The user-defined infrastructure intent declared in `.tf` configuration files (e.g. `multi_az = true`).
2. **Observed Cloud State**: The actual infrastructure state observed from AWS by the Anarva Control Plane & Observation Engine (`HA_ENABLED`, `ap-south-1a` primary AZ).
3. **Terraform State**: The local or remote Terraform state file storing the stable Anarva Resource ID (`res-rds-postgres-01`) and non-secret properties.
4. **Provider State**: The internal Go structs (`DatabaseResourceState`, `ComputeResourceState`, `StorageBucketResourceState`) populated during `Read`/`Create`/`Update`.
5. **Operation State**: The status of asynchronous control-plane jobs (`QUEUED`, `IN_PROGRESS`, `COMPLETED`, `FAILED`).

---

## ⚡ Lifecycle & Reconciliation Rules

### 1. Create Lifecycle (Async Verification)
`Terraform Create` ➔ `Provider POST /api/v1/resources` ➔ `Anarva Operation ID` ➔ `Poll /api/v1/operations/:id` ➔ `Operation COMPLETED` ➔ `Provider Read` ➔ `Set Terraform State`.
- A resource is **NEVER** marked as created in Terraform state before control-plane operation completion is verified.

### 2. Read Lifecycle & 404 State Removal
`Terraform Read` ➔ `Provider GET /api/v1/resources/:id` ➔ `Current Observed State` ➔ `Update Terraform State`.
- If the Anarva API returns **HTTP `404 Not Found`**, the resource ID is removed from Terraform state so Terraform can recreate the resource on subsequent `terraform apply`.
- **STALE** or **UNAVAILABLE** observation statuses do NOT remove resources from state to prevent accidental recreation during temporary telemetry outages.

### 3. Update Lifecycle & `ForceNew`
- Mutable attributes (e.g., `multi_az`, `storage_gb`, `acu_units`) are updated in-place via API modifications.
- Immutable attributes (e.g., resource `name`, `engine`, `region_id`) trigger `ForceNew` replacement in Terraform schema.

### 4. Eventual Consistency & Thread-Safety
- The provider client handles HTTP status codes `408`, `429`, `502`, `503`, `504` with exponential backoff and randomized full jitter.
- The provider HTTP client is completely thread-safe for concurrent Terraform resource operations.
