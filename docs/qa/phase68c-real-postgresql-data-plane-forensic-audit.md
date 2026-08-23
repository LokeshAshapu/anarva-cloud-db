# ANARVA Cloud — Phase 68C Real PostgreSQL Data-Plane Forensic Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 23, 2026  
**Auditor**: Principal Cloud Architect, Database Engineer & Security Engineer  
**Scope**: `/console/databases`, PostgreSQL Customer Data Plane & SQL Execution Engine  
**Audit Type**: READ-ONLY Forensic Audit & Architecture Decision  

---

## 1. Executive Summary

Phase 68C conducts a read-only forensic audit of the customer PostgreSQL data plane in ANARVA Cloud V1. 

While Phase 68B successfully made **database control-plane instance metadata** durable in the PostgreSQL `postgres_instances` table and enforced tenant authorization on API endpoints, the **customer SQL data plane** remains a simulated in-memory evaluator (`SQLService`) persisting to a local JSON file (`./data/anarva_sql_service_state.json`).

```
CURRENT ARCHITECTURE (SIMULATED DATA PLANE):
User Query -> PostgresHandler -> SQLService (Go String Parser) -> ./data/anarva_sql_service_state.json
                                                                           │
                                                                   [ RENDER REDEPLOY ]
                                                                           │
                                                                           v
                                                                 CUSTOMER ROWS WIPED!

TARGET ARCHITECTURE (REAL POSTGRESQL DATA PLANE):
User Query -> PostgresHandler -> Tenant Check -> PostgresSQLExecutor (pgx/database/sql) -> Data-Plane PostgreSQL Server
                                                                                                    │
                                                                                            [ RENDER REDEPLOY ]
                                                                                                    │
                                                                                                    v
                                                                                          CUSTOMER ROWS DURABLE!
```

---

## 2. Current Architecture & Forensic Audit Findings

### Question 1 & 2: Real PostgreSQL Container Spawning
- **Does ANARVA currently start a real PostgreSQL server/container?**  
  **NO.** `LocalDockerPostgresProvider` (`internal/postgres/provider/docker_provider.go`) allocates a port number (`15433+`) and creates Go RAM map records (`p.instances[inst.ID]`). It does **not** invoke `docker run`, execute Docker CLI commands, or create container volumes.

### Question 3 & 4: SQL Simulation Location & Connection Status
- **Does `SQLService` connect to a real PostgreSQL server?**  
  **NO.** `SQLService` (`internal/postgres/service/sql_service.go`) imports only standard Go library packages (`context`, `encoding/json`, `fmt`, `os`, `path/filepath`, `sort`, `strconv`, `strings`, `sync`, `time`). It has zero network socket, DSN connection, or SQL driver code (`database/sql`, `pgx`, `gorm`).

### Question 5 & 6: SQL Handler & Data Storage Location
- **What code handles DDL & DML statements?**  
  Custom Go regex and string parsers in `SQLService` methods: `handleCreateTable`, `handleInsert`, `handleUpdate`, `handleDelete`, `handleSelect`, `handleAlterTable`, `handleTruncate`, `handleDropTable`.
- **Where do customer rows physically exist?**  
  Customer tables and rows exist as JSON arrays inside `./data/anarva_sql_service_state.json` on the gateway container disk, as well as cached in browser `localStorage`.

### Question 7: Customer Row Survival Matrix

| Event | Metadata Survives? | Customer Rows Survive? | Physical Reason |
|:---|:---:|:---:|:---|
| **Browser Refresh** | YES | YES | Frontend loads cached query results from `localStorage` (`anarva_sql_query_cache_*`) |
| **Frontend Restart** | YES | YES | Restored from browser `localStorage` |
| **Gateway Restart (Local Dev)** | YES | YES | `SQLService.loadFromFile()` reloads `./data/anarva_sql_service_state.json` from local disk |
| **PostgreSQL Engine Restart** | N/A | N/A | No PostgreSQL engine process running |
| **Render Redeploy** | YES (Control Plane) | **NO (DATA LOST)** | Render ephemeral container filesystem wipes `./data/anarva_sql_service_state.json` |
| **Render Service Restart** | YES (Control Plane) | **NO (DATA LOST)** | Ephemeral container environment reset |

### Question 8 & 9: Instance ID & Connection String Connection
- Database instance IDs (`postgresql-1740250000000`) map to `SQLService` internal map keys, not PostgreSQL engine PIDs or sockets.
- Connection strings (`postgres://anarva_admin:pass_xxx@localhost:15433/postgres`) are formatted strings; no PostgreSQL server listens on port 15433.

### Question 10, 11, 12 & 13: Isolation & Control Plane Co-Location Risks
- **Multi-Tenant API Isolation**: Enforced in Phase 68B at the API layer (`GetInstanceForTenant`), blocking Tenant B from executing queries against Tenant A's database ID.
- **Storage Layer Isolation**: In `SQLService`, all tenant tables co-exist inside the same single JSON file (`./data/anarva_sql_service_state.json`).
- **Co-Location Risk**: Co-locating customer SQL data inside the control-plane PostgreSQL database without strict logical database/role boundary isolation would present **catastrophic security risks** (a malicious customer query like `DROP TABLE users;` could destroy control-plane accounts and system metadata).

---

## 3. Render Reality Check

| Feature | Render Capability | Impact on Customer Data Plane |
|:---|:---:|:---|
| **Docker Daemon Access** | ABSENT | Cannot run `docker run` or Docker-in-Docker inside Render application containers. |
| **Docker Socket (`/var/run/docker.sock`)** | ABSENT | `LocalDockerPostgresProvider` cannot spawn container instances on Render. |
| **Persistent Docker Volumes** | ABSENT | Container storage is 100% ephemeral. Files created in container disk are wiped on redeploy. |
| **Arbitrary TCP Ports** | ABSENT | Render web services expose only a single HTTP port (`PORT`). Exposing direct PostgreSQL TCP ports (e.g. 5432, 15433) per customer instance is unsupported on standard web services. |
| **PostgreSQL Container Survival** | IMPOSSIBLE | Docker container spawning is unsupported. |

---

## 4. Evaluation of Data-Plane Options A–E

### OPTION A: PostgreSQL Docker Container per Customer Database
- **Durability**: High on bare-metal; ZERO on Render.
- **Tenant Isolation**: Complete (process & network boundaries).
- **Render Compatibility**: **FAIL (0%)** — Docker daemon unavailable on Render.
- **Verdict**: Unsuitable for Render cloud deployment.

### OPTION B: Dedicated Customer Data-Plane PostgreSQL Server with Logical Multi-Tenancy (RECOMMENDED FOR V1)
- **Architecture**: A dedicated PostgreSQL data-plane instance/cluster (`ANARVA_CUSTOMER_DATAPLANE_DB`). Each customer database is created as an isolated PostgreSQL database with a dedicated database role:
  ```sql
  CREATE ROLE usr_<instance_id> WITH LOGIN PASSWORD 'pass_<random>';
  CREATE DATABASE db_<instance_id> WITH OWNER usr_<instance_id>;
  REVOKE ALL ON DATABASE db_<instance_id> FROM PUBLIC;
  ```
- **Durability**: 100% (backed by real PostgreSQL WAL, tables, indexes, and persistent disk/cloud DB).
- **Tenant Isolation**: Excellent (PostgreSQL role-level RBAC & database-level isolation prevent `usr_A` from accessing `db_B`).
- **Render Compatibility**: **100% PASS** — Connects via standard DSN to a managed PostgreSQL service (e.g., Render Managed PostgreSQL, Supabase, Neon, or AWS RDS).
- **Cost & Scalability**: High efficiency; low overhead per tenant database.
- **Verdict**: **RECOMMENDED ARCHITECTURE FOR ANARVA V1**.

### OPTION C: Managed PostgreSQL Provider per Database (AWS RDS / Neon / Supabase API)
- **Durability**: Maximum (automated cloud snapshots & PITR).
- **Tenant Isolation**: Maximum (separate cloud instance or serverless database branch).
- **Render Compatibility**: **100% PASS** — Provisions via cloud API and connects via DSN.
- **Verdict**: Recommended for production multi-cloud provider integration in Phase 69+.

### OPTION D: External Custom PostgreSQL Cluster Controlled by ANARVA
- **Verdict**: Overkill for V1 initial data-plane release; high operational complexity.

### OPTION E: Embedded Database Engine per Database Instance (e.g., SQLite / DuckDB with S3 Backup)
- **Verdict**: Useful fallback for lightweight offline edge dev, but lacks full PostgreSQL feature set.

---

## 5. Control Plane vs Data Plane Separation

To guarantee security, system stability, and blast-radius isolation, ANARVA Cloud strictly separates Control Plane and Data Plane:

```
+-------------------------------------------------------------------------+
|                         ANARVA CONTROL PLANE                            |
| Database: ANARVA_CONTROL_PLANE_DB                                       |
| Tables: users, organizations, projects, postgres_instances, audit_logs  |
+-------------------------------------------------------------------------+
                                    │
                         (Logical Separation Boundary)
                                    │
+-------------------------------------------------------------------------+
|                       ANARVA CUSTOMER DATA PLANE                        |
| Database Server: ANARVA_CUSTOMER_DATAPLANE_DB                           |
| Customer DBs: db_<instance_1>, db_<instance_2>, db_<instance_3>          |
| Customer Roles: usr_<instance_1>, usr_<instance_2>, usr_<instance_3>    |
| Content: Customer schemas, tables, indexes, rows, WAL, transactions     |
+-------------------------------------------------------------------------+
```

---

## 6. Definition of Done & Recommended Implementation Path (Phase 68D+)

### Phase 68D Definition of Done
1. **Real Query Execution Engine (`PostgresSQLExecutor`)**:
   - Implements `query.Executor` interface using `database/sql` or `pgx` pool.
   - Connects to customer data-plane PostgreSQL database (`db_<instance_id>`).
   - Executes native SQL queries (`CREATE TABLE`, `INSERT`, `SELECT`, `UPDATE`, `DELETE`, `CREATE INDEX`, transactions, functions, extensions).
   - Formats native PostgreSQL `sql.Rows` into `SQLQueryResult`.
2. **Database Provisioning in Data Plane**:
   - When a user creates a `PostgresInstance`, `PostgresService` provisions the logical database `db_<instance_id>` and role `usr_<instance_id>` on the data-plane PostgreSQL server.
3. **Decommission `SQLService` Simulation**:
   - Remove regex/string parsing and `./data/anarva_sql_service_state.json` local file dependency.

---

## FINAL CLASSIFICATION

REAL_POSTGRESQL_ENGINE:
    NO (Currently simulated by SQLService)

CUSTOMER_DATA_DURABILITY:
    EPHEMERAL (Stored in local JSON file ./data/anarva_sql_service_state.json; wiped on Render redeploy)

SQL_EXECUTION:
    SIMULATED (Go string parser in SQLService)

RENDER_COMPATIBILITY:
    NO (Current Docker container simulation cannot run on Render; ephemeral disk wipes JSON state)

TENANT_ISOLATION:
    PASS (Enforced at API layer in Phase 68B; requires logical DB isolation at data plane)

CONTROL_PLANE_DATA_PLANE_SEPARATION:
    PASS (Control-plane instance metadata stored in GORM postgres_instances; data plane kept separate)

RECOMMENDED_ARCHITECTURE:
    OPTION B — Dedicated Customer Data-Plane PostgreSQL Server with Logical Multi-Tenancy (CREATE DATABASE db_<id> WITH OWNER usr_<id>) and Real DSN Query Execution via database/sql or pgx.

BLOCKING FINDINGS:
    1. SQL queries are currently evaluated by custom Go regex string parsing (SQLService) rather than a real PostgreSQL engine.
    2. Customer SQL data is stored in a local JSON file (./data/anarva_sql_service_state.json) on container disk, which is wiped on Render redeployments.
    3. Docker container spawning in LocalDockerPostgresProvider is simulated in memory and cannot execute on Render web services.
