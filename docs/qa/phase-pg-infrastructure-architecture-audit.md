# ANARVA Cloud — PostgreSQL Infrastructure Phase 1 Architecture & Environment Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 23, 2026  
**Auditor**: Principal Infrastructure Architect & Database Lead  
**Scope**: ANARVA PostgreSQL Infrastructure & Data-Plane Hosting Architecture  
**Audit Type**: READ-ONLY Architectural Audit  

---

## 1. Current ANARVA Architecture

ANARVA Cloud is a distributed cloud database platform built in Go (Control Plane Gateway) and Next.js (Management Console).

```
                      +----------------------------------+
                      |     ANARVA Management Console    |
                      |   Next.js 14 Frontend Web App    |
                      +----------------------------------+
                                       │
                             (HTTP / REST JSON API)
                                       │
                                       v
                      +----------------------------------+
                      |       ANARVA Gateway Server      |
                      |     Go Control-Plane Process     |
                      +----------------------------------+
                                  │          │
            ┌─────────────────────┘          └─────────────────────┐
            │                                                      │
            v                                                      v
+-----------------------+                              +-----------------------+
| ANARVA CONTROL PLANE  |                              | ANARVA DATA PLANE     |
| PostgreSQL DB (5432)  |                              | Customer PG (5433)    |
| (ANARVA_CONTROL_PLANE)|                              | (ANARVA_CUSTOMER_DP)  |
+-----------------------+                              +-----------------------+
```

---

## 2. Current PostgreSQL Architecture

Currently, ANARVA operates two distinct database logical structures:

1. **Control-Plane Database (`ANARVA_CONTROL_PLANE_DB`)**:
   - Connection URL: `DATABASE_URL=postgres://anarva_admin:anarva_password@localhost:5432/anarva_cloud_db?sslmode=disable`
   - Docker Container: `anarva-postgres` (`postgres:16-alpine`, port 5432).
   - Storage Volume: `postgres_data` (Docker volume).
   - Responsibility: Stores GORM system metadata (`users`, `organizations`, `projects`, `postgres_instances`, `provider_resource_mappings`, `audit_logs`).
2. **Customer Data-Plane Provider (`PostgresDataPlaneProvider`)**:
   - Admin Connection DSN: `CUSTOMER_DATABASE_ADMIN_URL`
   - Provider Engine: `RealPostgresDataPlaneProvider` (`internal/postgres/provider/dataplane_provider.go`).
   - SQL Execution Engine: `PostgresSQLExecutor` (`internal/postgres/service/sql_executor.go`).
   - Responsibility: Dynamically creates and manages customer PostgreSQL databases (`db_<instance_id>`) and dedicated RBAC roles (`usr_<instance_id>`).

---

## 3. Control-Plane / Data-Plane Boundary

ANARVA Cloud strictly enforces a hard physical and logical boundary between Control Plane and Data Plane:

| Attribute | Control Plane | Data Plane |
|:---|:---|:---|
| **Database Identifier** | `anarva_cloud_db` | `db_<sanitized_instance_id>` |
| **Port** | 5432 | 5433 (Local) / Dedicated Cluster (Cloud) |
| **Access Rights** | ANARVA Gateway System (`anarva_admin`) | Dedicated Customer Role (`usr_<instance_id>`) |
| **Content** | User accounts, API keys, orgs, projects, database instance metadata, billing | Customer tables, indexes, rows, schemas, views, transactions |
| **Query Routing** | GORM ORM (`dbPool.DB`) | `PostgresSQLExecutor` (`database/sql` + `pgx`) |

*Absolute Security Rule: Customer SQL queries executed via `/console/databases` MUST NEVER execute against `ANARVA_CONTROL_PLANE_DB`.*

---

## 4. Existing PostgreSQL Provider Capabilities

The provider architecture in `internal/postgres/` supports:

- **Identifier Sanitization**: `SanitizeIdentifier` enforces strict `^[a-zA-Z0-9_]+$` validation on database and role names.
- **Cryptographic Passwords**: `GenerateStrongPassword()` generates 24-character random hex tokens via `crypto/rand`.
- **Administrative Provisioning**: `CreateDatabase` connects via `CUSTOMER_DATABASE_ADMIN_URL` to execute:
  ```sql
  CREATE ROLE usr_<instance_id> WITH LOGIN PASSWORD 'pass_<random_24char>';
  CREATE DATABASE db_<instance_id> WITH OWNER usr_<instance_id>;
  REVOKE ALL ON DATABASE db_<instance_id> FROM PUBLIC;
  GRANT ALL ON DATABASE db_<instance_id> TO usr_<instance_id>;
  ```
- **Error Sanitization**: `SanitizeSQLError` strips sensitive passwords and DSN strings from error output before returning responses to clients.
- **Query Timeout Enforcement**: Enforces `SET statement_timeout = 30000;` (30 seconds) and a 30-second Go context timeout.
- **Deprovisioning & Cleanup**: `DeleteDatabase` executes `DROP DATABASE db_<id> WITH (FORCE); DROP ROLE usr_<id>;`.

---

## 5. Infrastructure Options Evaluation

### OPTION A: Dedicated Customer PostgreSQL Container in Docker with Persistent Volume (RECOMMENDED V1)
- **Description**: Run a second PostgreSQL container (`anarva-customer-postgres-dataplane`) using `postgres:17-alpine` on host port `5433` with a named Docker volume (`customer_postgres_dataplane_data`).
- **Development Suitability**: Maximum. Zero external cloud account required; runs instantly via `docker-compose`.
- **Persistence**: High. Customer databases (`db_<id>`) survive gateway restarts, container restarts, and machine reboots.
- **Networking**: Host port 5433 bound to localhost; isolated from control-plane port 5432.
- **Security**: Dedicated admin DSN (`CUSTOMER_DATABASE_ADMIN_URL=postgres://anarva_dataplane_admin:secret@localhost:5433/postgres?sslmode=disable`).
- **Cost**: $0.
- **Migration Path**: 100% compatible. The exact same `PostgresDataPlaneProvider` code connects to a cloud-hosted PostgreSQL cluster in production by simply changing `CUSTOMER_DATABASE_ADMIN_URL`.

### OPTION B: Host-Native PostgreSQL Installation
- **Description**: Install PostgreSQL directly on the host OS daemon.
- **Suitability**: Moderate. Dependent on OS platform (Windows vs Linux vs macOS installation scripts).
- **Cost**: $0.
- **Migration Path**: Low portability compared to containerization.

### OPTION C: Dedicated Cloud VPS (AWS EC2 / Hetzner / DigitalOcean)
- **Description**: Spin up a dedicated Linux VPS running PostgreSQL 17 with NVMe storage.
- **Suitability**: Production target. High performance and control.
- **Cost**: $10–$40/month.
- **Migration Path**: Target for Phase 3 Production Cloud Deployment.

### OPTION D: Cloud Native PostgreSQL Orchestrator (Kubernetes CloudNativePG / Patroni Cluster)
- **Description**: Deploy multi-node HA PostgreSQL cluster with automated failover and WAL archiving.
- **Suitability**: Enterprise Tier (V3). High operational complexity.

---

## 6. Recommended V1 Infrastructure Architecture

**RECOMMENDED V1 ARCHITECTURE: OPTION A (Dedicated Local Dockerized Customer Data-Plane PostgreSQL Server)**

```
+------------------------------------------------------------------------------------+
|                                LOCAL DOCKER HOST                                   |
|                                                                                    |
|  +-------------------------------------+   +------------------------------------+  |
|  | Container: anarva-postgres          |   | Container: anarva-customer-postgres|  |
|  | Image: postgres:16-alpine           |   | Image: postgres:17-alpine          |  |
|  | Host Port: 5432                     |   | Host Port: 5433                    |  |
|  | Database: anarva_cloud_db           |   | Database: postgres (Admin)         |  |
|  | Volume: postgres_data               |   | Volume: customer_postgres_data     |  |
|  | (ANARVA CONTROL PLANE METADATA)     |   | (ANARVA CUSTOMER DATA PLANE)       |  |
|  +-------------------------------------+   +------------------------------------+  |
|                   ^                                         ^                      |
+-------------------|-----------------------------------------|----------------------+
                    │                                         │
        (DATABASE_URL)                       (CUSTOMER_DATABASE_ADMIN_URL)
                    │                                         │
                    +--------------------+--------------------+
                                         │
                       +-----------------------------------+
                       |       ANARVA Gateway Server       |
                       |       Go Control-Plane API        |
                       +-----------------------------------+
```

---

## 7. Network Architecture

- **Control-Plane Network**: Port `5432` bound to `127.0.0.1`. Accepts connections strictly from ANARVA Gateway.
- **Data-Plane Network**: Port `5433` bound to `127.0.0.1`. Accepts admin connection `CUSTOMER_DATABASE_ADMIN_URL` from ANARVA Gateway.
- **Direct Access**: Customer client connections (if enabled) route through host port `5433` specifying database `db_<instance_id>` and role `usr_<instance_id>`.

---

## 8. Storage Architecture

- **Volume Name**: `customer_postgres_dataplane_data`
- **Mount Point**: `/var/lib/postgresql/data` inside container `anarva-customer-postgres-dataplane`.
- **Durability**: Docker named volume persists across container `stop`, `restart`, `recreate`, and system reboots.
- **Directory Isolation**: Each customer database instance (`db_<instance_id>`) maintains native PostgreSQL tablespace directories managed by PostgreSQL 17 engine.

---

## 9. Security Architecture

1. **Administrative Credentials**: `CUSTOMER_DATABASE_ADMIN_URL` configured server-side in `.env` / environment variable. Never exposed to frontend or API clients.
2. **Customer Passwords**: Cryptographically random 24-character passwords generated per database instance via `crypto/rand`.
3. **Role-Level RBAC**:
   ```sql
   REVOKE ALL ON DATABASE db_<instance_id> FROM PUBLIC;
   GRANT ALL ON DATABASE db_<instance_id> TO usr_<instance_id>;
   ```
4. **Credential Non-Exposure**: Passwords omitted from JSON API responses (`json:"-"`) and sanitized from error logs via `SanitizeSQLError`.

---

## 10. Customer Database Isolation Model

- **Database-Level Isolation**: Each database instance resides in a separate PostgreSQL database (`CREATE DATABASE db_<id>`).
- **Role-Level Isolation**: Each database instance has a dedicated owner role (`CREATE ROLE usr_<id>`). `usr_A` cannot list tables or query data in `db_B`.
- **Resource Limits**: Configured via `statement_timeout = 30000` (30 seconds) to prevent runaway queries from consuming server resources.

---

## 11. ANARVA Integration Model

- **Gateway Configuration**:
  ```env
  DATABASE_URL=postgres://anarva_admin:anarva_password@localhost:5432/anarva_cloud_db?sslmode=disable
  CUSTOMER_DATABASE_ADMIN_URL=postgres://anarva_dataplane_admin:anarva_dataplane_secret@localhost:5433/postgres?sslmode=disable
  ```
- **Provisioning Flow**:
  1. `POST /api/v1/databases` -> Gateway creates `postgres_instances` record in Control Plane DB (`5432`).
  2. `RealPostgresDataPlaneProvider` connects to Data Plane DB (`5433`) via `CUSTOMER_DATABASE_ADMIN_URL`.
  3. Executes `CREATE ROLE usr_<id>` & `CREATE DATABASE db_<id>`.
  4. Updates control-plane status to `READY`.
- **Query Flow**:
  1. `POST /api/v1/databases/{id}/query` -> Gateway validates tenant ownership (`GetInstanceForTenant`).
  2. `PostgresSQLExecutor` connects to Data Plane DB (`5433`) database `db_<id>`.
  3. Executes native SQL statement and returns formatted result.

---

## 12. Persistence Model

- **Control-Plane Persistence**: GORM ORM -> `postgres_instances` table on port 5432.
- **Data-Plane Persistence**: Native PostgreSQL WAL + `pg_class` / `pg_attribute` tables on port 5433.
- **Restart Survival**:
  - Gateway Restart -> Customer data persists on port 5433.
  - Container Restart -> Named volume `customer_postgres_dataplane_data` retains all customer databases and rows.

---

## 13. Backup Considerations

- **Logical Backups**: `pg_dump -U anarva_dataplane_admin -h localhost -p 5433 db_<instance_id>` produces clean SQL backup files for individual customer instances.
- **Physical Backups**: File-level snapshot of Docker named volume `/var/lib/docker/volumes/customer_postgres_dataplane_data/_data`.
- **Point-in-Time Recovery (PITR)**: Enable WAL archiving (`wal_level = replica`, `archive_mode = on`) for Phase 3 production deployments.

---

## 14. Failure Scenarios & Recovery

| Failure Scenario | Automatic Handler / Recovery |
|:---|:---|
| `CREATE ROLE` succeeds, `CREATE DATABASE` fails | `RealPostgresDataPlaneProvider` catches error and executes rollback `DROP ROLE IF EXISTS usr_<id>`. Control-plane status set to `FAILED`. |
| Gateway crashes during query execution | Connection closed automatically by PostgreSQL timeout (`idle_in_transaction_session_timeout`). No data corruption. |
| Customer query attempts SQL injection | Strict identifier sanitization (`SanitizeIdentifier`) rejects invalid characters before query formatting. |
| Customer query runs indefinitely | Terminated automatically by `statement_timeout = 30000;` (30s) and Go context cancellation. |

---

## 15. Migration Path: Development -> Production Cloud

```
STEP 1: Development Environment (Phase 2)
Docker Compose Service `customer-postgres-dataplane` on localhost:5433
CUSTOMER_DATABASE_ADMIN_URL=postgres://admin:secret@localhost:5433/postgres

                             │
                             ▼
STEP 2: Staging Environment (Phase 3A)
Dedicated Staging VPS / Cloud Managed PostgreSQL Cluster
CUSTOMER_DATABASE_ADMIN_URL=postgres://admin:secret@staging-pg.anarva.internal:5432/postgres?sslmode=verify-full

                             │
                             ▼
STEP 3: Production Environment (Phase 3B)
High-Availability PostgreSQL Cluster with Standby Replicas & Automated WAL Archiving
CUSTOMER_DATABASE_ADMIN_URL=postgres://admin:secret@prod-pg-primary.anarva.internal:5432/postgres?sslmode=verify-full
```

---

## 16. Exact Phase 2 Implementation Plan

1. **Add Customer Data-Plane PostgreSQL Service to `docker-compose.yml`**:
   - Add `customer-postgres-dataplane` service (`postgres:17-alpine`, host port `5433:5432`, volume `customer_postgres_dataplane_data`).
2. **Update Development Environment Configuration**:
   - Add `CUSTOMER_DATABASE_ADMIN_URL=postgres://anarva_dataplane_admin:anarva_dataplane_secret@localhost:5433/postgres?sslmode=disable` to `.env` / Makefile startup script.
3. **Execute Live Integration Suite**:
   - Run `go test -v -run TestPhase68E_RealPostgresDataPlaneProvisioning_Integration ./internal/postgres/...`.
   - Run `go test -v -run TestPhase68F_RealPostgresSQLExecution_Integration ./internal/postgres/...`.
4. **Verify Live End-to-End Workflow**:
   - Provision instance -> execute DDL (`CREATE TABLE`) -> DML (`INSERT`, `SELECT`, `UPDATE`, `DELETE`) -> verify persistence across container & gateway restarts.

---

## 17. Definition of Done for Infrastructure Phase 2

[ ] `docker-compose.yml` updated with `customer-postgres-dataplane` container on port 5433.  
[ ] `CUSTOMER_DATABASE_ADMIN_URL` configured in local development environment.  
[ ] Real customer database (`db_<id>`) and role (`usr_<id>`) provisioned on port 5433.  
[ ] Native SQL DDL (`CREATE TABLE`) and DML (`INSERT`, `SELECT`, `UPDATE`, `DELETE`) executed successfully against port 5433.  
[ ] Rows survive gateway restart and Docker container restart.  
[ ] Real integration tests pass (`TestPhase68E_RealPostgresDataPlaneProvisioning_Integration` & `TestPhase68F_RealPostgresSQLExecution_Integration`).  
[ ] `REAL POSTGRESQL INTEGRATION: VERIFIED` certified.  
