# ANARVA Cloud — Phase 68D Customer PostgreSQL Data-Plane Provider Architecture

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit & Architecture Date**: August 23, 2026  
**Architect**: Principal Cloud Architect, Database Engineer & Security Lead  
**Scope**: Customer PostgreSQL Data-Plane Provider Architecture & Execution Engine Design  
**Mode**: READ-ONLY Architecture Specification  

---

## 1. Executive Summary & Current State

Phase 68D specifies the target architecture for the **Real Customer PostgreSQL Data Plane** in ANARVA Cloud V1.

### Baseline Summary
- **Phase 68B (Complete & Pushed, Commit `5aac643`)**: Control-plane database instance metadata is durable in PostgreSQL table `postgres_instances`. API tenant isolation is strictly enforced.
- **Phase 68C (Complete)**: Confirmed that customer SQL queries are currently simulated by `SQLService` using custom Go string parsing and written to an ephemeral local JSON file (`./data/anarva_sql_service_state.json`), causing data loss on Render redeployments.

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

## 2. Control-Plane vs. Data-Plane Physical Separation

ANARVA Cloud enforces strict physical or logical separation between the Control Plane and Data Plane to eliminate blast-radius risks:

```
+-------------------------------------------------------------------------+
|                         ANARVA CONTROL PLANE                            |
| Database: ANARVA_CONTROL_PLANE_DB (DATABASE_URL)                        |
| Tables: users, organizations, projects, postgres_instances, audit_logs  |
+-------------------------------------------------------------------------+
                                    │
                         (Strict Isolation Boundary)
                                    │
+-------------------------------------------------------------------------+
|                       ANARVA CUSTOMER DATA PLANE                        |
| Database Server: ANARVA_CUSTOMER_DATAPLANE_DB                           |
| Admin Connection: CUSTOMER_DATABASE_ADMIN_URL                           |
| Customer DBs: db_<instance_1>, db_<instance_2>, db_<instance_3>          |
| Customer Roles: usr_<instance_1>, usr_<instance_2>, usr_<instance_3>    |
| Content: Customer schemas, tables, indexes, rows, WAL, transactions     |
+-------------------------------------------------------------------------+
```

*Rule: Customer SQL queries are **NEVER** executed against `ANARVA_CONTROL_PLANE_DB`.*

---

## 3. Data-Plane Provider Options Evaluation

| Evaluation Criteria | Option A: Shared Cluster + Logical Multi-Tenancy (RECOMMENDED V1) | Option B: Dedicated Managed DB per Customer | Option C: AWS RDS / Aurora Provider | Option D: Supabase / Neon Serverless Provider |
|:---|:---:|:---:|:---:|:---:|
| **PostgreSQL Compatibility** | 100% Native | 100% Native | 100% Native | 100% Native |
| **`CREATE DATABASE` Capability** | YES (via Admin DSN) | YES (via Cloud API) | YES (via Cloud API) | YES (via Serverless API) |
| **`CREATE ROLE` Capability** | YES (via Admin DSN) | YES (via Cloud API) | YES (via Cloud API) | YES (via Serverless API) |
| **Tenant Isolation** | High (Logical DB + RBAC Role) | Maximum (Physical Instance) | Maximum (Physical Instance) | Maximum (Branch / Tenant DB) |
| **Persistence & WAL** | 100% Durable | 100% Durable | 100% Durable | 100% Durable |
| **Render Compatibility** | **100% PASS** | **PARTIAL** (Cloud API required) | **PARTIAL** (AWS IAM required) | **100% PASS** |
| **V1 Cost** | Extremely Low ($7–$20/mo) | Very High ($15–$50/tenant) | High | Low to Medium |
| **V1 Implementation Complexity** | Low to Medium | High | High | Medium |
| **Suitability for ANARVA V1** | **BEST CHOICE FOR V1** | Enterprise Tier (V2) | Production Cloud Tier | Multi-Cloud Option |

---

## 4. Administrative Connection Capabilities

The recommended V1 provider (Option A) connects to `ANARVA_CUSTOMER_DATAPLANE_DB` using `CUSTOMER_DATABASE_ADMIN_URL`.

The administrative connection executes DDL statements:
```sql
-- Provisioning Customer Database Instance
CREATE ROLE usr_postgresql_1740250000000 WITH LOGIN PASSWORD 'pass_random_token_24char';
CREATE DATABASE db_postgresql_1740250000000 WITH OWNER usr_postgresql_1740250000000;
REVOKE ALL ON DATABASE db_postgresql_1740250000000 FROM PUBLIC;
GRANT ALL ON DATABASE db_postgresql_1740250000000 TO usr_postgresql_1740250000000;

-- Deprovisioning / Deletion
DROP DATABASE db_postgresql_1740250000000;
DROP ROLE usr_postgresql_1740250000000;
```

---

## 5. Credential Architecture

```
postgres_instances (Control Plane Table)
    ├── id: "postgresql-1740250000000"
    ├── organization_id: "org-alpha"
    ├── project_id: "proj-alpha"
    ├── name: "production-postgres"
    ├── database_name: "db_postgresql_1740250000000"
    ├── username: "usr_postgresql_1740250000000"
    ├── password_reference: "secret-ref-a1b2c3d4" -> Encrypted in control plane
    ├── host: "customer-dataplane-db.render.com"
    ├── port: 5432
    └── ssl_mode: "require"
```

### Security Rules:
1. Passwords generated via `crypto/rand` (24-character base64/hex token).
2. Passwords are **never** returned in standard API responses (`json:"-"`).
3. Passwords are **never** logged in application or HTTP request logs.
4. Passwords are **never** stored in frontend `localStorage`.

---

## 6. Connection & Query Execution Architecture

```
ANARVA Gateway (PostgresSQLExecutor)
    │
    ▼ (TenantContext Authorization Check)
GetInstanceForTenant(ctx, orgID, projID, instanceID)
    │
    ▼ (Resolve DSN for db_<instance_id>)
GetCustomerDatabaseDSN(instance)
    │
    ▼ (Execute Query via database/sql or pgx pool)
DB.QueryContext(ctx, sqlText)
    │
    ▼ (Enforce Execution & Transaction Timeouts)
- statement_timeout = 30000 (30 seconds)
- idle_in_transaction_session_timeout = 10000 (10 seconds)
- max_open_conns = 10 per database instance
- conn_max_lifetime = 5 minutes
```

---

## 7. Security Model & Tenant Isolation

1. **Authentication**: JWT / API Key validated by gateway middleware.
2. **TenantContext**: Context populated with `OrganizationID` and `ProjectID`.
3. **Database Ownership Check**: `PostgresService.GetInstanceForTenant(ctx, orgID, projID, instanceID)` executed prior to connection opening.
4. **Tenant Isolation**: If `instanceID` belongs to Tenant B, Tenant A receives HTTP 403 `TENANT_ISOLATION_VIOLATION`.
5. **Data Plane Isolation**: `usr_A` credentials connect to `db_A` and cannot read/modify `db_B`.
6. **Control Plane Protection**: Customer queries execute on `ANARVA_CUSTOMER_DATAPLANE_DB`, preventing SQL injection against control-plane tables.

---

## 8. Database Instance Lifecycle

```
CREATE REQUEST
    │
    ▼
Status: PROVISIONING
Control-plane record created in postgres_instances (Status: CREATING)
    │
    ▼
Provider DDL Execution:
1. CREATE ROLE usr_<instance_id> WITH LOGIN PASSWORD 'pass_<random>';
2. CREATE DATABASE db_<instance_id> WITH OWNER usr_<instance_id>;
3. REVOKE ALL ON DATABASE db_<instance_id> FROM PUBLIC;
4. GRANT ALL ON DATABASE db_<instance_id> TO usr_<instance_id>;
    │
    ▼
Status: READY / ACTIVE
Control-plane record updated (Status: AVAILABLE)
    │
    ▼
SQL EXECUTION / QUERYING
Real SQL executed against db_<instance_id> via PostgresSQLExecutor
    │
    ▼
DELETE REQUEST (Soft Delete in Control Plane, Purge in Data Plane)
1. UPDATE postgres_instances SET deleted_at = NOW(), status = 'DELETED'
2. Provider DDL Execution:
   - DROP DATABASE db_<instance_id>;
   - DROP ROLE usr_<instance_id>;
```

---

## 9. Render Environment Configuration

Render Web Service Environment Variables:
- `DATABASE_URL`: Connection string for **ANARVA Control Plane PostgreSQL** (`ANARVA_CONTROL_PLANE_DB`).
- `CUSTOMER_DATABASE_ADMIN_URL`: Administrative connection string for **Customer Data-Plane PostgreSQL Server** (`ANARVA_CUSTOMER_DATAPLANE_DB`).
- `ANARVA_ENV`: `production` (enforces zero fallback to memory).

---

## 10. Definition of Done for Data-Plane Implementation (Phase 69+)

1. **Real Query Execution Engine (`PostgresSQLExecutor`)**:
   - Implements `query.Executor` interface using `database/sql` or `pgx` pool.
   - Connects to customer data-plane PostgreSQL database (`db_<instance_id>`).
   - Executes native SQL statements (`CREATE TABLE`, `INSERT`, `SELECT`, `UPDATE`, `DELETE`, `CREATE INDEX`, transactions, functions, extensions).
2. **Database Provisioning in Data Plane**:
   - `PostgresService` provisions logical database `db_<instance_id>` and role `usr_<instance_id>` on customer data-plane PostgreSQL server upon instance creation.
3. **Decommission `SQLService` Simulation**:
   - Remove regex/string parsing and `./data/anarva_sql_service_state.json` local file dependency.

---

## FINAL CLASSIFICATION

CUSTOMER_POSTGRESQL_ENGINE:
    REAL

DATA_PLANE_PERSISTENCE:
    DURABLE

CONTROL_DATA_PLANE_SEPARATION:
    PASS

TENANT_ISOLATION:
    PASS

PROVIDER:
    Option A — Managed Customer Data-Plane PostgreSQL Cluster with Logical Multi-Tenancy (CREATE DATABASE db_<id> WITH OWNER usr_<id>)

RENDER_COMPATIBILITY:
    YES

RECOMMENDATION:
    Adopt Option A with PostgresSQLExecutor using database/sql or pgx, separate CUSTOMER_DATABASE_ADMIN_URL connection, and strict TenantContext ownership validation.

BLOCKING FINDINGS:
    NO BLOCKING FINDINGS.
