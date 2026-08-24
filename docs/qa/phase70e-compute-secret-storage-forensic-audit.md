# ANARVA Cloud Phase 70E — Compute Secret Storage Forensic Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 23, 2026  
**Auditor**: Principal Security Architect & Cloud Infrastructure Lead  
**Scope**: Compute Secret / Environment Variable Storage & Exposure Forensic Audit  
**Baseline Commit**: Phase 70D Certified (`2a9b571` + 70A–70D)  

---

## 1. Executive Summary

A forensic security audit was conducted on the Compute domain, PostgreSQL repository, HTTP handlers, gateway initialization, and provider execution layers to evaluate the lifecycle, storage, and exposure of environment variables (`ComputeInstance.EnvVars`) and sensitive configuration data.

### Primary Audit Finding
Currently, `ComputeInstance.EnvVars` is serialized via `json.Marshal(c.EnvVars)` in GORM's `BeforeSave` hook and stored directly as an **UNENCRYPTED PLAINTEXT JSON STRING** in the `env_vars_json` column of the `compute_instances` table in PostgreSQL.

Furthermore, `ComputeInstance.EnvVars` carries the tag `json:"envVars,omitempty"`, causing plain-text environment variables to be returned in HTTP API JSON responses (`POST /api/v1/compute/instances`, `GET /api/v1/compute/instances`, `GET /api/v1/compute/instances/{id}`).

---

## 2. Forensic Audit Matrix

```
COMPUTE SECRET STORAGE:
    UNSAFE (Plaintext storage in PostgreSQL env_vars_json text column & API JSON responses)

ENV_VARS_AT_REST:
    PLAINTEXT (Stored as raw unencrypted JSON string in compute_instances.env_vars_json)

API_SECRET_EXPOSURE:
    EXPOSED (EnvVars map included in json:"envVars,omitempty" tag on ComputeInstance)

LOG_SECRET_EXPOSURE:
    SAFE (Activity events log instance name only; raw env vars not explicitly dumped in activity log)

DOCKER_SECRET_EXPOSURE:
    NEEDS REVIEW (EnvVars not yet injected via -e in docker run; docker inspect does not contain them currently, but would if passed via CLI)

TENANT_ISOLATION:
    PASS (GetTenantScopedByID enforces org/project isolation; Tenant B cannot read Tenant A secrets)

RESTART_RECOVERY:
    PASS (AfterFind hook successfully re-hydrates EnvVars map from env_vars_json column)

BACKUP_EXPOSURE:
    UNSAFE (pg_dump and database backups contain raw plaintext secret strings from env_vars_json)

PRODUCTION_KEY_MANAGEMENT:
    NEEDS REVIEW (No COMPUTE_SECRET_ENCRYPTION_KEY or KMS key configured in cmd/gateway/main.go)

BLOCKING FINDINGS:
    1. PLAINTEXT DB STORAGE: compute_instances.env_vars_json column stores raw unencrypted JSON string containing environment variables and secrets at rest.
    2. API SECRET EXPOSURE: ComputeInstance.EnvVars field is serialized in HTTP responses for GET/POST /api/v1/compute/instances without masking or redaction.
    3. BACKUP LEAK RISK: Database backups (pg_dump) contain plaintext secret material from env_vars_json.
    4. MISSING ENCRYPTION KEY: No encryption key (AES-256-GCM / KMS) or key management logic is currently configured in control plane.
```

---

## 3. Detailed Forensic Lifecycle Analysis

```
User API Request (POST /api/v1/compute/instances)
       │ (Contains envVars: {"DB_PASSWORD": "supersecretpassword123"})
       v
ComputeHandler.handleInstances -> ComputeUseCase.CreateInstance
       │
       v
PostgresComputeRepository.Create -> GORM BeforeSave Hook
       │ (json.Marshal(c.EnvVars) -> c.EnvVarsJSON)
       v
PostgreSQL Database (`compute_instances.env_vars_json` column)
       │ [UNENCRYPTED PLAINTEXT IN DATABASE]
       v
PostgresComputeRepository.GetByID -> GORM AfterFind Hook
       │ (json.Unmarshal([]byte(c.EnvVarsJSON), &c.EnvVars))
       v
HTTP Response (GET /api/v1/compute/instances/{id})
       │ [UNMASKED PLAINTEXT IN API RESPONSE]
       v
Database Backup (pg_dump / snapshot)
       │ [PLAINTEXT SECRETS IN DUMP FILES]
```

### 3.1 EnvVars & Security Field Storage
- `internal/compute/domain/compute.go`:
  - `EnvVars` map[string]string `json:"envVars,omitempty" gorm:"-"`
  - `EnvVarsJSON` string `json:"-" gorm:"column:env_vars_json;type:text"`
  - `Security` InstanceSecurityPolicy `json:"security" gorm:"-"`
  - `SecurityJSON` string `json:"-" gorm:"column:security_json;type:text"`

`BeforeSave` serializes `c.EnvVars` into `c.EnvVarsJSON` using standard `json.Marshal()`. No encryption algorithm (AES-GCM, NaCl, KMS) is applied prior to database insertion.

### 3.2 Database At-Rest Exposure & Backups
- Column `env_vars_json` in PostgreSQL `compute_instances` table holds raw JSON strings (e.g. `{"PORT":"8080","SECRET_KEY":"sk_live_12345"}`).
- Any database backup (`pg_dump`), WAL log, or disk snapshot captures these secrets in unencrypted plaintext.

### 3.3 API Response Exposure
- Because `EnvVars` is tagged with `json:"envVars,omitempty"`, all `GET` and `POST` responses for compute instances output `envVars` in plaintext JSON to the client.
- Redaction or masking (e.g. `"DB_PASSWORD": "********"`) is not currently implemented at the API boundary.

### 4.4 Host Docker Engine Exposure
- `internal/compute/provider/provider.go`: `LocalDockerComputeProvider.CreateInstance` currently executes `docker run -d --name <name> --cpus <cpus> --memory <mem> <image>`.
- Environment variables are kept in provider memory (`p.instances`) and PostgreSQL, but are not currently appended as `-e KEY=VAL` flags to `docker run`.
- If environment variables are appended as `-e` in the future, `docker inspect` and host process listings (`ps aux`) will expose them in plaintext unless injected via secure secret files or environment files.

---

## 4. Remediation Plan (Target for Future Phase 70F)

1. **AES-256-GCM Envelope Encryption**:
   - Implement symmetric AES-256-GCM encryption for `EnvVarsJSON` in `BeforeSave` / `AfterFind` hooks or repository data-mapper layer.
   - Use key derivation (HKDF / PBKDF2) or control-plane secret key (`COMPUTE_SECRET_ENCRYPTION_KEY`).
2. **API Secret Masking / Redaction**:
   - Omit or mask sensitive keys in public `GET` API responses unless explicitly requested via authenticated management endpoint.
3. **Database Migration Support**:
   - Provide backward-compatible migration strategy for legacy plaintext `env_vars_json` records.
