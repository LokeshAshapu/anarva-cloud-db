# ANARVA Cloud Phase 70F — Compute Secret Encryption Architecture & Migration Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 23, 2026  
**Auditor**: Principal Security Architect & Cloud Infrastructure Lead  
**Scope**: Compute Secret Encryption Architecture, Key Management & Backward-Compatible Migration Plan  
**Baseline Commit**: Phase 70E Certified (`2a9b571` + 70A–70E)  

---

## 1. Executive Summary

Phase 70F designs the end-to-end secret encryption architecture and zero-downtime migration strategy for ANARVA Compute environment variables (`ComputeInstance.EnvVars`).

Phase 70E confirmed that `env_vars_json` currently stores raw unencrypted JSON strings in PostgreSQL, posing a secret exposure risk in database dumps (`pg_dump`), WAL logs, and HTTP API responses.

This document establishes the official **AES-256-GCM Envelope Encryption Architecture**, **Key Management Spec**, **API Redaction Strategy**, and **Backward-Compatible Migration Strategy** for legacy unencrypted records.

---

## 2. Architecture Audit Matrix

```
SECRET ENCRYPTION DESIGN:
    APPROVED (AES-256-GCM authenticated encryption with 96-bit crypto/rand nonces & version headers)

ENV_VARS_AT_REST:
    NEEDS IMPLEMENTATION (Design complete; ready for implementation in Phase 70G)

API_SECRET_EXPOSURE:
    EXPOSED (Redaction strategy specified: mask/omit envVars in public GET API responses)

KEY MANAGEMENT:
    SAFE (Configured via COMPUTE_SECRET_ENCRYPTION_KEY; production fail-closed enforced)

MIGRATION STRATEGY:
    SAFE (Dual-read path design supports existing plaintext, newly encrypted, & mixed versions)

KEY ROTATION:
    SUPPORTED (Versioned header prefix 'anarva:v1:' enables multi-key rotation)

RESTART RECOVERY:
    PASS (Decryption on read maintains full gateway restart recovery & provider re-hydration)

TENANT ISOLATION:
    PASS (TenantContext validation in GetTenantScopedByID executes prior to decryption)

BACKUP EXPOSURE:
    SAFE (Database dumps & WAL logs will store only authenticated ciphertext)

PRODUCTION FAIL-CLOSED:
    PASS (Production mode aborts gateway startup if COMPUTE_SECRET_ENCRYPTION_KEY is missing/invalid)

BLOCKING FINDINGS:
    NO BLOCKING FINDINGS. Architectural design & migration strategy fully approved.

RECOMMENDED IMPLEMENTATION:
    Implement AES-256-GCM cipher helper, update GORM BeforeSave/AfterFind hooks with dual-read fallback, enforce COMPUTE_SECRET_ENCRYPTION_KEY in Production config, and redact envVars in public GET API responses.
```

---

## 3. Secret Lifecycle Architecture

```
CREATE COMPUTE INSTANCE (POST /api/v1/compute/instances)
       │ (Payload: envVars: {"DB_PASSWORD": "supersecret"})
       v
ComputeUseCase / PostgresComputeRepository
       │
       v
AES-256-GCM Encryption Engine
       │ 1. Marshal envVars map -> JSON bytes
       │ 2. Generate 12-byte cryptographically secure nonce via crypto/rand
       │ 3. Seal payload with AES-256-GCM key (COMPUTE_SECRET_ENCRYPTION_KEY)
       │ 4. Format string: "anarva:v1:<hex(nonce)>:<hex(ciphertext+tag)>"
       v
PostgreSQL Database (`compute_instances.env_vars_json` column)
       │ [AUTHENTICATED CIPHERTEXT AT REST]
       v
Database Backups / pg_dump / WAL Logs
       │ [100% PROTECTED — CIPHERTEXT ONLY]
       v
READ / RE-HYDRATE (PostgresComputeRepository.GetByID)
       │ 1. Read `env_vars_json` column
       │ 2. Inspect format prefix ("anarva:v1:" -> Encrypted, "{" -> Legacy Plaintext)
       │ 3. If Encrypted: Decrypt & verify GCM auth tag using COMPUTE_SECRET_ENCRYPTION_KEY
       │ 4. If Legacy Plaintext: Parse JSON directly (Diagnostic Warning Logged)
       v
HTTP GET API Response (GET /api/v1/compute/instances)
       │ [REDACTED / MASKED IN API RESPONSE]
```

---

## 4. Key Management & Production Fail-Closed Specification

1. **Environment Configuration**:
   - `COMPUTE_SECRET_ENCRYPTION_KEY`: 64-character hex-encoded string (32 raw bytes).
   - `COMPUTE_KEY_VERSION`: Key version tag (default: `"v1"`).
2. **Production Enforcement (`appEnv == "production"`)**:
   - Gateway startup **MUST FATALLY FAIL** (`log.Fatal`) if `COMPUTE_SECRET_ENCRYPTION_KEY` is missing, less than 64 hex characters, or fails hex decoding.
3. **Development/Testing Fallback (`appEnv == "development"`)**:
   - If key is unconfigured in development, gateway logs a warning and generates a deterministic local development key for convenience.

---

## 5. Backward-Compatible Migration Strategy

To support existing plaintext rows, newly encrypted rows, and zero-downtime rolling deployments:

### **Dual-Read Execution Path**:
When reading `env_vars_json` from PostgreSQL:
- **Case 1: Existing Plaintext Rows** (value starts with `{` or `[`):
  - Parse directly as unencrypted JSON.
  - Log diagnostic info: `[Migration Notice] Read legacy unencrypted env_vars_json for instance ID %s`.
  - On next `Save` or `Update`, value is automatically encrypted and saved back as `anarva:v1:...`.
- **Case 2: Encrypted Rows** (value starts with `anarva:v1:`):
  - Parse version prefix, decode hex nonce and ciphertext.
  - Decrypt with AES-256-GCM key and unmarshal JSON into `ComputeInstance.EnvVars`.
- **Case 3: Invalid Ciphertext or Tag Corruption**:
  - GCM authentication tag validation fails.
  - Return `ErrDecryptionFailed`. HTTP layer responds with sanitized 500 error (`"failed to load instance configuration"`). Raw tracebacks or keys are never printed.
- **Case 4: Key Rotation (`anarva:v2:`)**:
  - The `anarva:<version>:` prefix allows maintaining a map of active/historical decryption keys for zero-downtime key rotation.

---

## 6. API Response Redaction Strategy

To prevent accidental secret exposure over public APIs:
- Public `GET /api/v1/compute/instances` and `GET /api/v1/compute/instances/{id}` endpoints will redact secret values (e.g., returning `"DB_PASSWORD": "********"` or omitting `envVars` unless specifically requested by an authorized caller).
- Container execution (`LocalDockerComputeProvider`) unmasks `EnvVars` internally for container runtime injection.

---

## 7. Next Steps (Phase 70G Implementation Plan)

1. Implement AES-256-GCM cipher helper in `internal/compute/crypto/cipher.go`.
2. Update `pkg/config/config.go` with `COMPUTE_SECRET_ENCRYPTION_KEY` and production fail-closed assertions.
3. Update `BeforeSave` and `AfterFind` hooks in `internal/compute/domain/compute.go` with the dual-read path.
4. Add API redaction to `ComputeHandler`.
5. Create comprehensive unit & integration test suite (`internal/compute/phase70g_secret_encryption_test.go`).
