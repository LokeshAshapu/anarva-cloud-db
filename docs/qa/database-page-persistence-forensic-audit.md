# ANARVA Cloud — Database Page & Customer Data Plane Forensic Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 22, 2026  
**Auditor**: Principal Cloud Architect & Security Engineer  
**Scope**: Database Control-Plane & Customer Data Plane (`/console/databases`, `internal/postgres/`, `internal/database/`)  
**Audit Type**: READ-ONLY Forensic Investigation  

---

## 1. Executive Summary

This forensic audit investigates the persistence, execution, tenant isolation, and cloud reality of the `/console/databases` workflow and customer database data plane in ANARVA Cloud V1.

```
USER WORKFLOW:
Login -> /console/databases -> Create Database -> SQL Console -> CREATE TABLE -> INSERT -> SELECT
                                                                                            │
                                                                                    [ RESTART GATEWAY ]
                                                                                    [ REDEPLOY RENDER ]
                                                                                            │
                                                                                            v
                                                                                  DATA & METADATA LOST!
```

---

## 2. Database Page Architecture & Request Flow

### Frontend (`web/app/console/databases/page.tsx`)
- **Instance Loading**: Loaded directly from browser `localStorage` key `anarva_user_managed_dbs_v2`. If empty, hardcodes a mock default instance (`postgresql-prod-01`). Does **NOT** fetch database list from backend API on page load.
- **Instance Creation**: User submits creation wizard -> frontend sends `POST /api/v1/databases` via `fetch()` (ignoring errors with `.catch(() => null)`) -> saves instance object into browser `localStorage`.
- **SQL Execution**: Sends `POST /api/v1/databases/{id}/query` with `{ "sql": "..." }`.
- **Query Caching**: Query results are cached in browser `localStorage` under `anarva_sql_query_cache_{id}` and query text under `anarva_sql_query_text_{id}`.
- **Page Refresh Behavior**: `useEffect` restores instance list, active selection, tab, and query results directly from `localStorage` without hitting the backend API.

### Backend Request Trace (`cmd/gateway/main.go` & `internal/postgres/`)
```
Frontend: fetchAPI('/api/v1/databases/{id}/query')
    ↓
Gateway Mux: postgresHandler.NewPostgresHandler(pgService, sqlService).RegisterRoutes(mux)
    ↓
Handler: internal/postgres/handler/postgres_handler.go (handleDatabaseSubroutes -> case "sql", "query":)
    ↓
Service: internal/postgres/service/sql_service.go (ExecuteQuery)
    ↓
Storage: Local JSON File ./data/anarva_sql_service_state.json
```

---

## 3. Control Plane Metadata Classification

| Metadata | Storage Location | Classification |
|:---|:---|:---:|
| Database Instance ID | Go `map[string]*PostgresInstance` in `LocalDockerPostgresProvider` & browser `localStorage` | `IN_MEMORY` / `REAL_LOCAL` |
| Instance Name & Engine | Go `map` & browser `localStorage` | `IN_MEMORY` / `REAL_LOCAL` |
| Region, CPU, Memory, Storage | Go `map` & browser `localStorage` | `IN_MEMORY` / `REAL_LOCAL` |
| Connection Host & Port | Simulated (`localhost`, port 15433+) | `SIMULATED` |
| Admin Credential Reference | Generated reference `secret-ref-xxxxxxxx` in Go map | `IN_MEMORY` |

*Note: Database control-plane instances (`PostgresInstance`) are currently **NOT** registered in GORM `AutoMigrate` for the PostgreSQL control-plane database in `cmd/gateway/main.go`.*

---

## 4. Customer Data Plane Investigation

### Storage Location of Customer Tables and Rows
1. `CREATE TABLE`: Parsed by custom Go regex/string logic in `SQLService` (`internal/postgres/service/sql_service.go`). Creates `TableState` in Go map `s.instanceTables[instanceID][tableName]`.
2. `INSERT INTO`: Appends row slice `[]interface{}` to `TableState.Rows`.
3. **Physical Persistence**: Serialized to local JSON file: `./data/anarva_sql_service_state.json` (and `os.TempDir()/anarva_sql_service_state.json`).
4. **PostgreSQL Engine Presence**: **NONE**. Customer SQL queries are NOT executed inside a PostgreSQL database container or PostgreSQL daemon. They are evaluated by custom Go code and written to a local JSON file.

---

## 5. PostgreSQL Provider Analysis & Docker Reality

- **Provider Implementation**: `LocalDockerPostgresProvider` in `internal/postgres/provider/docker_provider.go`.
- **Docker Container Creation**: **NONE**. The provider does NOT invoke Docker CLI, Docker API, or spawn Docker containers.
- **Docker Volume**: **NONE**. No Docker volumes are created or mounted.
- **Reality Label**: `LOCAL_POSTGRES (DOCKER_SIM)`.

---

## 6. Render Production Reality

- **Render Sandbox Limitation**: Render standard web services run inside unprivileged Linux application containers without access to `/var/run/docker.sock` or host Docker daemon. Spawning Docker containers on Render is impossible.
- **Ephemeral Filesystem**: Render container filesystems are wiped completely on every git deployment, service restart, or host machine migration.
- **Impact on Data Plane**: Any file stored at `./data/anarva_sql_service_state.json` is completely destroyed during a Render deployment.

---

## 7. Restart Boundary Matrix

| Event | Metadata survives? | Tables survive? | Rows survive? | Root Cause |
|:---|:---:|:---:|:---:|:---|
| **Browser Refresh** | YES | YES | YES | Frontend loads instance list & cached query results from `localStorage` (`anarva_user_managed_dbs_v2`) |
| **SQL Console Reopen** | YES | YES | YES | Restored from browser `localStorage` (`anarva_sql_query_cache_*`) |
| **Gateway Restart (Local Dev)** | NO (Backend RAM) / YES (Browser) | YES | YES | `SQLService` reloads `./data/anarva_sql_service_state.json` from local disk; backend RAM map lost |
| **PostgreSQL Restart** | N/A | N/A | N/A | No actual PostgreSQL container running |
| **PostgreSQL Container Deletion** | N/A | N/A | N/A | No Docker containers created by provider |
| **Render Redeploy** | NO | NO | NO | Ephemeral filesystem wipes `./data/anarva_sql_service_state.json` and gateway RAM |
| **Render Machine Replacement** | NO | NO | NO | Ephemeral container environment reset |

---

## 8. Tenant Isolation Analysis

- **Endpoint**: `POST /api/v1/databases/{id}/query`
- **Vulnerability**: `handleDatabaseSubroutes` extracts `instanceID` directly from URL path and invokes `h.sqlService.ExecuteQuery(r.Context(), instanceID, sqlText)` without verifying whether `instanceID` belongs to the requesting tenant's `OrganizationID` or `ProjectID`.
- **Risk Level**: **CRITICAL (IDOR)**. Any tenant who knows or guesses another tenant's `instanceID` can execute SQL statements against their database instance tables.

---

## 9. Credential Security Analysis

- **Password Generation**: Generated dynamically as `pass_xxxxxxxx`.
- **Storage**: Credential references (`secret-ref-xxxxxxxx`) stored in Go in-memory maps.
- **Driver Code Leakage**: Frontend generates deterministic fallback passwords in driver connection code strings (`pass_${cleanId.slice(-8)}`).

---

## 10. Exact Data Loss Root Cause

"Why did customer data appear to disappear previously?"

1. **Control-Plane Metadata Loss**: `PostgresInstance` records live in `LocalDockerPostgresProvider` Go RAM maps, not PostgreSQL control-plane database tables. Gateway restarts wipe this map.
2. **Data-Plane Ephemeral Storage**: Customer SQL tables/rows live in `./data/anarva_sql_service_state.json` on local container disk. Render redeployments wipe this file.
3. **Browser Mismatch**: Frontend `localStorage` holds stale database IDs after backend restarts, causing queries to fail or target orphaned IDs.

---

## 11. Current Reality Classification

- **Control Plane Metadata**: `IN_MEMORY` / `REAL_LOCAL`
- **Database Engine Execution**: `SIMULATED` (Custom Go JSON Evaluator)
- **Customer Table / Row Persistence**: `EPHEMERAL` (Local JSON file `./data/anarva_sql_service_state.json`)
- **Tenant Isolation**: `NEEDS_REMEDIATION` (Missing ownership check on `/query`)

---

## 12. Smallest Possible Fix (Specification Only)

1. **Control Plane Metadata Persistence**:
   - Register `PostgresInstance` in GORM `AutoMigrate` for control-plane PostgreSQL database (`postgres_instances` table).
   - Replace `LocalDockerPostgresProvider` in-memory map with `PostgresDatabaseRepository` using GORM.
2. **Data Plane Persistence**:
   - Update `SQLService` to persist table states into control-plane PostgreSQL (or durable object/relational storage) instead of local file `./data/anarva_sql_service_state.json`.
3. **Tenant Isolation**:
   - Enforce `TenantContext` validation in `handleDatabaseSubroutes` prior to executing SQL queries.

---

## 13. Verification Plan & Remaining Limitations

- **Automated Verification**: Integration tests validating database instance creation, query execution, restart survival across DB instances, and tenant isolation checks.
- **Remaining Limitations**: Full multi-tenant isolation of isolated PostgreSQL container workloads requires dedicated cloud database instances or managed PostgreSQL engines in production cloud environments.

---

## DATABASE PAGE FORENSIC CERTIFICATION

Control Plane Persistence:
    FAIL

Customer Database Persistence:
    FAIL

Table Persistence:
    PARTIAL (Local disk JSON file on dev machine; lost on Render)

Row Persistence:
    PARTIAL (Local disk JSON file on dev machine; lost on Render)

Gateway Restart Persistence:
    PARTIAL (Tables reload from JSON file locally; provider RAM map lost)

Database Restart Persistence:
    N/A (No Docker engine container)

Render Production Persistence:
    FAIL (Ephemeral filesystem wipes ./data/anarva_sql_service_state.json)

Tenant Isolation:
    FAIL (Missing tenant ownership validation on /api/v1/databases/{id}/query)

Credential Security:
    PARTIAL (In-memory secret references; deterministic frontend driver strings)

Exact Data Loss Boundary:
    Local container filesystem (./data/anarva_sql_service_state.json) and Provider Go RAM map (LocalDockerPostgresProvider.instances)

Smallest Required Fix:
    Persist PostgresInstance in GORM control-plane PostgreSQL and persist SQLService table states into PostgreSQL control-plane storage with tenant authorization checks.

Production Ready:
    NO

Blocking Findings:
    1. Control-plane database instances stored in Go RAM map instead of PostgreSQL control-plane database.
    2. Customer SQL tables and rows stored in local JSON file ./data/anarva_sql_service_state.json, which is wiped on Render redeployments.
    3. Missing tenant isolation verification on POST /api/v1/databases/{id}/query.
