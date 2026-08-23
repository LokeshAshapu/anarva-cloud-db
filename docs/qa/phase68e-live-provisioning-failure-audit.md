# ANARVA Cloud Phase 68E — Live Provisioning Failure Forensic Audit Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 23, 2026  
**Auditor**: Principal Cloud Architect & Security Lead  
**Scope**: `TestPhase68E_RealPostgresDataPlaneProvisioning_Integration` Failure Forensic Audit  
**Baseline Commit**: `ab11649` (Phase 68F Complete)  

---

## 1. Executive Summary & Root Cause

The real PostgreSQL integration test `TestPhase68E_RealPostgresDataPlaneProvisioning_Integration` fails at Step 3 (`svc.DeleteInstance(ctx, created.ID)`) with error:

```
postgres instance not found
```

### Root Cause
1. In `TestPhase68E_RealPostgresDataPlaneProvisioning_Integration` (`internal/postgres/phase68e_data_plane_provisioning_test.go`), `PostgresService` is instantiated with a `nil` repository:
   ```go
   svc := postgresService.NewPostgresServiceFull(nil, pgProv, dpProv)
   ```
2. During `svc.CreateInstance(...)`:
   - Because `s.dataPlaneProvider` (`RealPostgresDataPlaneProvider`) is set, `CreateInstance` invokes `s.dataPlaneProvider.CreateDatabase(ctx, inst)` to provision `db_<instance_id>` and `usr_<instance_id>` directly on the live PostgreSQL server at `localhost:5433`.
   - `s.provider.CreateInstance(...)` is **NOT** called. Consequently, `inst` is **never added** to `LocalDockerPostgresProvider.instances` in-memory map.
3. During Step 3 (`svc.DeleteInstance(ctx, created.ID)`):
   - `DeleteInstance` checks `if s.repo != nil`. Since `s.repo` is `nil`, it falls back to:
     ```go
     inst, err = s.provider.GetInstance(ctx, instanceID)
     ```
   - `s.provider.GetInstance` searches `LocalDockerPostgresProvider.instances` map for `created.ID`.
   - Because `created.ID` was never inserted into `s.provider.instances` during creation, `s.provider.GetInstance` returns `domain.ErrInstanceNotFound` (`"postgres instance not found"`).
   - `DeleteInstance` aborts before calling `s.dataPlaneProvider.DeleteDatabase(ctx, inst)`!

---

## 2. Call Chain Trace

```
TestPhase68E_RealPostgresDataPlaneProvisioning_Integration
  │
  ├─> 1. CreateInstance(ctx, ...)
  │      ├─> s.repo is nil (Skipped)
  │      ├─> s.dataPlaneProvider.CreateDatabase(...)
  │      │     └─> Connects to localhost:5433 via CUSTOMER_DATABASE_ADMIN_URL
  │      │     └─> Executes CREATE ROLE usr_<id> and CREATE DATABASE db_<id> [SUCCESS]
  │      └─> Returns created PostgresInstance struct [SUCCESS]
  │
  ├─> 2. Admin DSN Query Verification
  │      ├─> SELECT count(*) FROM pg_database WHERE datname = 'db_<id>' => 1 [SUCCESS]
  │      └─> SELECT count(*) FROM pg_roles WHERE rolname = 'usr_<id>'   => 1 [SUCCESS]
  │
  └─> 3. DeleteInstance(ctx, created.ID)
         ├─> s.repo is nil
         ├─> s.provider.GetInstance(ctx, created.ID)
         │     └─> Searches LocalDockerPostgresProvider.instances in-memory map
         │     └─> ID not found in map -> Returns domain.ErrInstanceNotFound
         └─> Fails with: "postgres instance not found" [FAILURE]
```

---

## 3. Findings Matrix

| Diagnostic Question | Result | Details |
|:---|:---:|:---|
| **Exact Failure Location** | `internal/postgres/service/postgres_service.go:138` | Inside `DeleteInstance` when resolving control-plane `inst` struct |
| **Data-Plane PostgreSQL Reached?** | **YES** | Real PostgreSQL server at `localhost:5433` was reached during Step 1 & Step 2; `db_<id>` and `usr_<id>` were successfully created in PostgreSQL |
| **Control-Plane Lookup Status** | **FAIL** | Failed because `s.repo` was `nil` and `s.provider` map was empty |
| **Instance Creation Status** | **PASS** | Instance creation on real PostgreSQL data-plane succeeded 100% |
| **Repository Involved** | `nil` repo / `LocalDockerPostgresProvider` | Service was initialized with `repo = nil` |
| **Origin of Error** | **Control Plane** | Error is 100% a control-plane lookup lookup bug, NOT a PostgreSQL database error |

---

## 4. Proposed Smallest Fix Options

### Fix Option A (Recommended — Service Level Fix)
In `internal/postgres/service/postgres_service.go`, update `DeleteInstance` to use `s.GetInstance(ctx, instanceID)` or construct a minimal `domain.PostgresInstance` struct when `inst` is not found in control plane, so that data-plane cleanup can complete:

```go
func (s *PostgresService) DeleteInstance(ctx context.Context, instanceID string) error {
	var inst *domain.PostgresInstance
	var err error
	if s.repo != nil {
		inst, err = s.repo.GetByID(ctx, instanceID)
	} else {
		inst, err = s.GetInstance(ctx, instanceID)
		if err != nil {
			// Fallback: construct minimal instance with ID so data-plane cleanup can proceed
			inst = &domain.PostgresInstance{ID: instanceID}
			err = nil
		}
	}
	if err != nil {
		return err
	}

	if s.dataPlaneProvider != nil {
		_ = s.dataPlaneProvider.DeleteDatabase(ctx, inst)
	} else {
		_ = s.provider.DeleteInstance(ctx, instanceID)
	}

	if s.repo != nil {
		return s.repo.Delete(ctx, instanceID)
	}
	return nil
}
```

### Fix Option B (Test Suite Fix)
In `internal/postgres/phase68e_data_plane_provisioning_test.go`, populate `pgProv.instances` map or pass a repository instance when constructing `PostgresServiceFull`:
```go
pgProv := postgresProvider.NewLocalDockerPostgresProvider()
dpProv := postgresProvider.NewRealPostgresDataPlaneProvider(adminDSN)
svc := postgresService.NewPostgresServiceFull(nil, pgProv, dpProv)
```

---

## 5. Verification Plan

1. Verify `TestPhase68E_RealPostgresDataPlaneProvisioning_Integration` passes 100% when running `go test -v ./internal/postgres/...` with `CUSTOMER_DATABASE_ADMIN_URL`.
2. Verify `TestPhase68F_RealPostgresSQLExecution_Integration` continues to pass 100%.
3. Run full gateway build and Next.js web application build.

---

```
PHASE 68E DEBUG AUDIT:
    ROOT CAUSE: DeleteInstance in postgres_service.go attempted to look up inst via s.provider.GetInstance(ctx, instanceID) when s.repo was nil, but s.provider (LocalDockerPostgresProvider) was never populated during CreateInstance when s.dataPlaneProvider (RealPostgresDataPlaneProvider) was active.

DATA-PLANE POSTGRESQL:
    REACHED

CONTROL-PLANE LOOKUP:
    FAIL

INSTANCE CREATION:
    PASS

TEST FIX REQUIRED:
    YES

BLOCKING FINDINGS:
    - DeleteInstance in postgres_service.go fails control-plane lookup when s.repo is nil and s.dataPlaneProvider is used.
```
