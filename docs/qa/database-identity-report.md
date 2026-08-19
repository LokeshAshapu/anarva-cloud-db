# ANARVA Database Identity & Connection Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Auditor**: Persistence Reliability Engineer  
**Report Date**: August 19, 2026  

---

## 1. Database Connection & Environment Configuration

| Setting | Production Mode (`APP_ENV=production`) | Development / Standalone Mode |
| :--- | :--- | :--- |
| **DATABASE ENGINE** | PostgreSQL 17.2 | PostgreSQL 17.2 or File-Synced State Engine |
| **DATABASE HOST** | Configured via `DATABASE_URL` (Render / Cloud DSN) | `localhost` or local data directory |
| **DATABASE PORT** | 5432 | 5432 |
| **DATABASE NAME** | Configured via `DATABASE_URL` (e.g. `anarva_cloud_db`) | `anarva_db` / `./data/anarva_cp_*.json` |
| **DATABASE USER** | `[REDACTED]` | `anarva_admin` / `usr-dev` |
| **DATABASE PASSWORD** | `[REDACTED]` | `[REDACTED]` |
| **DATABASE STORAGE MODE** | Persistent Cloud Volume / RDS Multi-AZ | Disk File Synchronization (`./data/`) |
| **FAIL-CLOSED PROTECTION** | **ENABLED** (`cmd/gateway/main.go` terminates if `DATABASE_URL` is missing in production) | Fallback to disk file state synchronization |

---

## 2. DSN Connection String Redaction Rules

All credentials, passwords, and JWT secret tokens are strictly redacted before output:
- DSN Format: `postgres://[USER]:[REDACTED]@[HOST]:[PORT]/[DB_NAME]?sslmode=require`
- JWT Manager: HMAC-SHA256 signature key `[REDACTED]`
