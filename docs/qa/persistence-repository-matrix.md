# ANARVA Persistence Repository Matrix

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Document**: Storage Engine & Persistence Matrix  

---

| Resource | Repository Implementation | Storage Engine | Durable | Restart Safe | Multi-Instance Safe | Production Safe |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Users** | `userRepository` / `memUserRepo` | PostgreSQL GORM / Disk JSON (`./data/anarva_cp_users.json`) | **YES** | **YES** | YES (PostgreSQL) | YES |
| **Sessions** | `sessionRepository` / `memSessionRepo` | PostgreSQL GORM / Disk JSON (`./data/anarva_cp_sessions.json`) | **YES** | **YES** | YES (PostgreSQL) | YES |
| **API Keys** | `apiKeyRepository` / `memKeyRepo` | PostgreSQL GORM / Disk JSON (`./data/anarva_cp_apikeys.json`) | **YES** | **YES** | YES (PostgreSQL) | YES |
| **Organizations** | `organizationRepository` / `memOrgRepo` | PostgreSQL GORM / Disk JSON (`./data/anarva_cp_orgs.json`) | **YES** | **YES** | YES (PostgreSQL) | YES |
| **Projects** | `projectRepository` / `memProjRepo` | PostgreSQL GORM / Disk JSON (`./data/anarva_cp_projects.json`) | **YES** | **YES** | YES (PostgreSQL) | YES |
| **Databases** | `instanceRepository` / `SQLService` | PostgreSQL GORM / Disk JSON (`./data/anarva_sql_service_state.json`) | **YES** | **YES** | YES (PostgreSQL) | YES |
| **Compute** | `LocalDockerComputeProvider` | Docker Daemon / Container State Map | **YES** | **YES** | YES (Docker CLI) | YES |
| **Storage Buckets** | `LocalStorageProvider` | Disk Directory (`anarva-local-storage/<bucket>/`) | **YES** | **YES** | YES | YES |
| **Storage Objects** | `LocalStorageProvider` | Disk File (`anarva-local-storage/<bucket>/<key>`) | **YES** | **YES** | YES | YES |
| **Networking** | `PostgresNetworkingRepository` | PostgreSQL GORM / Disk JSON (`./data/anarva_cp_networks.json`) | **YES** | **YES** | YES (PostgreSQL) | YES |
| **Load Balancers** | `lbRepository` | PostgreSQL GORM / Disk JSON (`./data/anarva_cp_loadbalancers.json`) | **YES** | **YES** | YES (PostgreSQL) | YES |
| **Backups** | `backupRepository` / `BackupEngine` | PostgreSQL GORM / Disk JSON (`./data/backups/`) | **YES** | **YES** | YES (PostgreSQL) | YES |
| **Audit Logs** | `auditLogRepository` / `memAuditRepo` | PostgreSQL GORM / Disk JSON (`./data/anarva_cp_audits.json`) | **YES** | **YES** | YES (PostgreSQL) | YES |
