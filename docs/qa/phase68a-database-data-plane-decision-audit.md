# ANARVA Cloud — Phase 68A Database Data-Plane Decision Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 23, 2026  
**Auditor**: Principal Cloud Architect & Security Engineer  
**Scope**: `/console/databases`, PostgreSQL Data Plane & SQL Execution Architecture  
**Audit Mode**: READ-ONLY Decision Audit  

---

## 1. Executive Summary & Audit Purpose

Phase 68A evaluates the current database data-plane architecture (`internal/postgres/`) to determine whether `SQLService` is a development simulation that must be replaced by a real database engine execution pipeline.

```
FINDING SUMMARY:
SQLService is a custom Go in-memory regex parser and JSON serializer.
It DOES NOT connect to PostgreSQL, database/sql, pgx, or Docker containers.
Customer SQL tables and rows are stored in a local JSON file (./data/anarva_sql_service_state.json).
```

---

## 2. Audit of `SQLService`

### 1. Does `SQLService` connect to a real PostgreSQL server?
**NO.** `SQLService` (`internal/postgres/service/sql_service.go`) has zero network connection code to any PostgreSQL daemon or server instance.

### 2. Does `SQLService` use `database/sql`, `pgx`, `GORM`, or another client?
**NO.** `SQLService` imports only standard Go library packages (`context`, `encoding/json`, `fmt`, `os`, `path/filepath`, `sort`, `strconv`, `strings`, `sync`, `time`). It imports no SQL drivers or database clients whatsoever.

### 3. Does `SQLService` parse SQL itself and maintain structures in Go?
**YES.** `SQLService` parses incoming SQL strings using custom Go string manipulation (`strings.HasPrefix`, `strings.ToUpper`, regex matching) and maintains state in a nested Go map: `instanceTables map[string]map[string]*TableState`.

### 4. SQL Statements Actually Supported
- `CREATE TABLE` (basic column name extraction and default value parsing)
- `ALTER TABLE` (`ADD COLUMN`, `DROP COLUMN`, `RENAME TO`)
- `TRUNCATE` (clears in-memory row slice)
- `INSERT INTO ... VALUES (...)` (parses literal tuple values)
- `UPDATE ... SET` (updates in-memory row values)
- `DELETE FROM ... WHERE` (filters in-memory row slice)
- `DROP TABLE` (removes table key from map)
- `SELECT` (handles basic `WHERE`, `LIMIT`, `VERSION()`, `SHOW DATABASES`, `EXPLAIN`)
- `BEGIN`, `COMMIT`, `ROLLBACK` (returns dummy success status)

### 5. In-Memory SQL State Structures
- `instanceTables`: `map[string]map[string]*TableState` where key is `instanceID` mapping to `tableName` -> `TableState`.
- `TableState`:
  ```go
  type TableState struct {
      Name           string                 `json:"name"`
      Columns        []string               `json:"columns"`
      ColumnDefaults map[string]interface{} `json:"columnDefaults"`
      Rows           [][]interface{}        `json:"rows"`
  }
  ```

### 6. JSON / File Persistence
- State file: `./data/anarva_sql_service_state.json` (or `DATA_DIR/anarva_sql_service_state.json`)
- Backup mirror file: `os.TempDir()/anarva_sql_service_state.json`
- `saveToFileLocked()` serializes `instanceTables` to formatted JSON on disk after mutating SQL queries.

### 7. SQL Feature Compatibility Matrix

| Feature | Supported by `SQLService`? | Details |
|:---|:---:|:---|
| `CREATE TABLE` | SIMULATED | Basic column names and simple defaults |
| `INSERT` | SIMULATED | Literal value tuples parsed in Go |
| `SELECT` | SIMULATED | Basic equality `WHERE` and `LIMIT` |
| `UPDATE` | SIMULATED | Simple `SET col = val` matching |
| `DELETE` | SIMULATED | Simple `WHERE col = val` deletion |
| `DROP TABLE` | SIMULATED | Deletes map entry in Go |
| `ALTER TABLE` | SIMULATED | Add/Drop column in Go slices |
| `CREATE INDEX` | UNSUPPORTED | Ignored or syntax error |
| Constraints (`FOREIGN KEY`, `CHECK`) | UNSUPPORTED | Stripped or ignored |
| Transactions | UNSUPPORTED | `BEGIN`/`COMMIT` return dummy strings without locking or MVCC |
| `JOIN` | UNSUPPORTED | Multi-table joins return errors or empty sets |
| `GROUP BY` | UNSUPPORTED | Aggregations not computed |
| `ORDER BY` | UNSUPPORTED | Sorting not implemented |
| `LIMIT` | SIMULATED | Slice slicing `[:limit]` |
| PostgreSQL Functions | UNSUPPORTED | `NOW()`, `UUID_GENERATE_V4()`, etc. unsupported except static `VERSION()` |
| PostgreSQL Extensions | UNSUPPORTED | `pgvector`, `postgis`, `pg_trgm` completely unsupported |

### 8. PostgreSQL Compatibility Assessment
`SQLService` is a **simplified development simulation**. It is not PostgreSQL-compatible and cannot support real application frameworks (ORM models, migrations, complex queries, transactions, or extensions).

---

## 3. Real PostgreSQL Provider Audit & Disconnect Analysis

### Instance Creation & Execution Path
```
User Click: Create PostgreSQL Instance
    ↓
POST /api/v1/databases (PostgresHandler.handleDatabases)
    ↓
PostgresService.CreateInstance
    ↓
LocalDockerPostgresProvider.CreateInstance
    ↓
Allocates port (15433+) & stores instance metadata in p.instances Go RAM map
```

### SQL Execution Path
```
User Click: Execute Query in SQL Console
    ↓
POST /api/v1/databases/{id}/query (PostgresHandler.handleDatabaseSubroutes)
    ↓
SQLService.ExecuteQuery
    ↓
Reads/Writes s.instanceTables Go map & ./data/anarva_sql_service_state.json
```

### The Architectural Disconnect
The SQL Console execution path **completely bypasses** `PostgresService` and `LocalDockerPostgresProvider`. `PostgresHandler` routes `/api/v1/databases/{id}/query` directly into `SQLService`, which operates on isolated local JSON files without inspecting instance host, port, credentials, container status, or connection parameters.

---

## 4. Data Flow Comparison

### CURRENT ARCHITECTURE
```
Frontend (/console/databases)
    ↓ (fetch /api/v1/databases/{id}/query)
PostgresHandler.handleDatabaseSubroutes
    ↓
SQLService.ExecuteQuery (Go Regex & Map Operations)
    ↓
./data/anarva_sql_service_state.json (Local Container Filesystem)
    ↓
Customer Tables + Rows (Wiped on Render Redeploy)
```

### INTENDED ARCHITECTURE
```
Frontend (/console/databases)
    ↓ (HTTP + Authorization: Bearer <JWT>)
Gateway Middleware (JWT Authentication & TenantContext)
    ↓
PostgresHandler (Tenant Ownership Verification)
    ↓
PostgresDatabaseRepository (PostgreSQL Control-Plane DB)
    ↓
PostgresProvider (Local Docker / Cloud Managed DB)
    ↓
Real PostgreSQL Database Connection (database/sql or pgx pool)
    ↓
PostgreSQL Data Directory / Persistent Cloud Volume
    ↓
Durable Customer Tables + Rows
```

---

## 5. Tenant Security & IDOR Vulnerability Audit

- **Endpoint**: `POST /api/v1/databases/{id}/query`
- **Authentication**: JWT validated at gateway middleware layer.
- **TenantContext**: Context populated with `OrganizationID` and `ProjectID`.
- **Handler Verification**: **FAIL**. `handleDatabaseSubroutes` extracts `instanceID` from the URL path (`parts[0]`) and directly calls `SQLService.ExecuteQuery(r.Context(), instanceID, sqlText)`.
- **Database Ownership Validation**: **NONE**. The handler does not check if `instanceID` belongs to the authenticated tenant.
- **Vulnerability**: **CRITICAL (IDOR)**. Any authenticated user who knows or guesses another tenant's `instanceID` can query or mutate tables/rows of that database instance.

---

## 6. Provider Boundary & Interface Reusability

The existing `PostgresProvider` interface (`internal/postgres/provider/provider.go`) defines the required lifecycle operations:
- `CreateInstance(ctx, inst, adminPassword)`
- `GetInstance(ctx, instanceID)`
- `ListInstances(ctx, orgID, projectID)`
- `GetConnectionInfo(ctx, instanceID)`

This contract provides a clean separation for:
1. **Local Development**: `LocalDockerPostgresProvider` (spawning actual Docker containers with named volumes).
2. **Production Cloud**: `CloudPostgresProvider` (provisioning AWS RDS, Cloudflare Hyperdrive, or managed PostgreSQL instances).

---

## PHASE 68A DECISION

SQLService Classification:
    SIMULATION

Customer SQL Data Location:
    Local JSON file (./data/anarva_sql_service_state.json) and browser localStorage

Does SQLService execute against real PostgreSQL:
    NO

Docker PostgreSQL Provider:
    SIMULATED (In-memory map without container spawning)

Persistent Volume:
    ABSENT (Ephemeral filesystem on cloud deployment)

Tenant Isolation:
    FAIL (Missing database ownership check on /api/v1/databases/{id}/query)

Production Render Compatibility:
    NO (Ephemeral filesystem wipes local JSON data; Docker daemon unavailable)

Recommended Architecture:
    Persist database control-plane metadata in PostgreSQL GORM control-plane database and connect database query execution through real database drivers (database/sql / pgx) with strict tenant ownership validation.

Recommended Next Implementation:
    Phase 68B — Implement GORM control-plane persistence for PostgresInstance and enforce tenant ownership validation on query endpoints.
