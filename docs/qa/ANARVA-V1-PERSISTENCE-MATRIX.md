# ANARVA Cloud Control Plane — V1 Persistence Matrix

**Audit Date**: August 19, 2026  
**Auditor**: Senior Database Architect & Persistence Reliability Lead  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  
**Production API**: `https://anarva-cloud-db-api.onrender.com`  

---

## 1. Executive Summary & Persistence Boundaries

This matrix documents the actual persistence behavior of every data model and resource type across all persistence boundaries in the ANARVA platform.

### Boundary Definitions:
1. **Process Restart**: Restarting the Go Gateway process on the same machine.
2. **Machine Restart**: Rebooting the underlying host server or VM.
3. **Container Restart**: Redeploying or restarting the container on an ephemeral host (e.g. Render Web Services).
4. **Render Redeployment**: Triggering a new build or deploy on Render, which creates a fresh container with a pristine filesystem.
5. **Cross-Client Access**: Fetching data from a different browser, device, or geographic location.

---

## 2. Model & Resource Persistence Classification Matrix

| Entity / Model | Storage Engine (Prod) | Storage Engine (Dev) | Process Restart | Machine Restart | Container Restart | Render Deploy | Cross-Client Access | Production Classification |
| :--- | :--- | :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| **Users** (`users`) | PostgreSQL | In-Memory + `./data/users.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Sessions** (`sessions`) | PostgreSQL | In-Memory + `./data/sessions.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **API Keys** (`api_keys`) | PostgreSQL | In-Memory + `./data/api_keys.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Verification Tokens** (`verification_tokens`) | PostgreSQL | In-Memory + `./data/tokens.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Audit Logs** (`audit_logs`) | PostgreSQL | In-Memory + `./data/audit_logs.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Organizations** (`organizations`) | PostgreSQL | In-Memory + `./data/organizations.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Projects** (`projects`) | PostgreSQL | In-Memory + `./data/projects.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Org Members** (`organization_members`) | PostgreSQL | In-Memory + `./data/members.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Invitations** (`invitations`) | PostgreSQL | In-Memory + `./data/invitations.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Database Instances** (`database_instances`) | PostgreSQL | In-Memory + `./data/instances.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Backup Metadata** (`backup_records`) | PostgreSQL | In-Memory + `./data/backups.json` | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Reliability Ops** (`anarva_operations`) | PostgreSQL | In-Memory | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Lock Leases** (`resource_lock_leases`) | PostgreSQL | In-Memory | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Idempotency Records** (`idempotency_records`) | PostgreSQL | In-Memory | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Tenant Quotas** (`tenant_quotas`) | PostgreSQL | In-Memory | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Virtual Networks** (`virtual_networks`) | PostgreSQL | In-Memory | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Subnets** (`subnets`) | PostgreSQL | In-Memory | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Security Groups** (`security_groups`) | PostgreSQL | In-Memory | ✅ YES | ✅ YES | ✅ YES | ✅ YES | ✅ YES | `PRODUCTION_DATABASE` |
| **Object Storage Files** | Local Disk (`./data/storage`) | Local Disk (`./data/storage`) | ✅ YES | ✅ YES | ❌ NO | ❌ NO | ⚠️ LOCAL | `LOCAL_FILE_ONLY` |
| **Backup Dump Archives** | Local Disk (`./data/backups/`) | Local Disk (`./data/backups/`) | ✅ YES | ✅ YES | ❌ NO | ❌ NO | ⚠️ LOCAL | `LOCAL_FILE_ONLY` |
| **Compute Metadata** | In-Memory Map | In-Memory Map | ❌ NO | ❌ NO | ❌ NO | ❌ NO | ❌ NO | `IN_MEMORY_ONLY` |
| **Load Balancer Rules** | In-Memory Map | In-Memory Map | ❌ NO | ❌ NO | ❌ NO | ❌ NO | ❌ NO | `IN_MEMORY_ONLY` |
| **MySQL Instance Metadata** | In-Memory Map | In-Memory Map | ❌ NO | ❌ NO | ❌ NO | ❌ NO | ❌ NO | `IN_MEMORY_ONLY` |
| **Webhook Subscriptions** | Gateway Memory | Gateway Memory | ❌ NO | ❌ NO | ❌ NO | ❌ NO | ❌ NO | `IN_MEMORY_ONLY` |
| **Billing & Quota State** | Gateway Memory | Gateway Memory | ❌ NO | ❌ NO | ❌ NO | ❌ NO | ❌ NO | `IN_MEMORY_ONLY` |

---

## 3. Forensic Analysis & Critical Findings

### 1. Control-Plane Data Safety:
When `DATABASE_URL` is set in production, all primary entities (**Users, Organizations, Projects, Database Instances, Virtual Networks, Reliability Operations, Audit Logs**) are stored directly in PostgreSQL via GORM. They survive process restarts, container restarts, Render redeployments, and cross-client access.

### 2. Development Fallback Risk:
When running locally or when `DATABASE_URL` is unconfigured in development, the gateway uses in-memory repositories synced to `./data/*.json` files. These files survive local process restarts on host machines, but on Render web services (which have an ephemeral filesystem), any data stored in `./data/*.json` is **wiped on container restart or redeploy**.

### 3. Object Storage & Backup File Vulnerability:
Object storage bucket objects (`internal/storage/provider/local_storage_provider.go`) and backup dump archives are saved directly to local disk (`./data/storage` and `./data/backups`). On Render, file uploads will **disappear** when the container restarts or redeploys.

---

## 4. Remediation Requirements for Production Cloud Persistence

1. **Object Storage**: Upgrade `LocalStorageProvider` to stream object uploads to AWS S3, Cloudflare R2, or GCP Cloud Storage when in production mode.
2. **Backup Archives**: Stream database dumps directly to S3/R2 compatible object storage rather than writing temporary files to local disk.
3. **Compute & Load Balancer Repositories**: Migrate in-memory compute and load balancer repositories to GORM-backed PostgreSQL tables.
