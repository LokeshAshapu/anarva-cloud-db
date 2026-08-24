# ANARVA Cloud Phase 70G — Compute Secret Encryption Certification Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Certification Date**: August 23, 2026  
**Auditor**: Principal Security Architect & Cloud Infrastructure Lead  
**Scope**: Compute Secret Encryption Implementation & API Secret Redaction  
**Baseline Commit**: Phase 70F Architecture Certified (`2a9b571` + 70A–70F)  

---

## 1. Executive Summary

Phase 70G implements authenticated **AES-256-GCM Envelope Encryption** for ANARVA Compute environment variables (`ComputeInstance.EnvVars`). All secret payloads stored at rest in PostgreSQL (`compute_instances.env_vars_json`) are encrypted using a fresh 96-bit (12-byte) `crypto/rand` nonce per encryption and prefixed with standard version header `anarva:v1:<hex(nonce)>:<hex(ciphertext+tag)>`.

Public HTTP API responses (`GET /api/v1/compute/instances`, `GET /api/v1/compute/instances/{id}`, `POST /api/v1/compute/instances`) now automatically redact secret environment variables via `RedactSecrets()`, preventing accidental secret leaks over the API while preserving internal provider access for container execution.

Production mode (`ANARVA_ENV=production`) enforces **FAIL-CLOSED** key validation, blocking gateway startup if `COMPUTE_SECRET_ENCRYPTION_KEY` is missing or invalid.

---

## 2. Implementation Summary

1. **AES-256-GCM Cipher Package** (`pkg/crypto/secret_cipher.go`):
   - Implements `SecretCipher` interface with `AESGCMCipher`.
   - Generates 12-byte nonces via `crypto/rand.Read()`.
   - Formats ciphertext as `anarva:v1:<hex(nonce)>:<hex(ciphertext+tag)>`.
   - Validates GCM authentication tags on decryption; rejects tampered or wrong-key payloads with `ErrDecryptionFailed`.
2. **GORM Persistence Hooks & Dual-Read Migration Path** (`internal/compute/domain/compute.go`):
   - `BeforeSave`: Encrypts `EnvVars` map to JSON bytes, then seals via `crypto.GetGlobalCipher()`.
   - `AfterFind`: Dual-read path inspects `env_vars_json`.
     - Starts with `anarva:v1:` -> Decrypts and unmarshals JSON into `EnvVars`.
     - Starts with `{` or `[` -> Legacy unencrypted JSON; unmarshals directly and auto-encrypts on next save.
3. **API Secret Redaction** (`internal/compute/delivery/http/compute_handler.go`):
   - Invokes `inst.RedactSecrets()` before returning `ComputeInstance` payloads in HTTP responses.
4. **Production Fail-Closed Key Management** (`cmd/gateway/main.go` & `pkg/config/config.go`):
   - Added `COMPUTE_SECRET_ENCRYPTION_KEY` configuration.
   - Enforces `log.Fatal` if key is missing or invalid in `appEnv == "production"`.

---

## 3. Certification Checklist

```
ENCRYPTION:
    PASS (AES-256-GCM authenticated encryption verified with fresh 96-bit nonces)

LEGACY MIGRATION:
    PASS (Dual-read path reads legacy plaintext JSON and auto-encrypts to anarva:v1: on save)

API REDACTION:
    PASS (ComputeInstance.EnvVars redacted in GET and POST HTTP API responses)

KEY MANAGEMENT:
    PASS (Configured via COMPUTE_SECRET_ENCRYPTION_KEY / SecretEncryptionKey)

PRODUCTION FAIL-CLOSED:
    PASS (Production mode aborts gateway startup if key is missing or invalid)

70A REGRESSION:
    PASS (Compute Tenant Isolation tests 100% PASS)

70B REGRESSION:
    PASS (Gateway Restart Recovery tests 100% PASS)

70C REGRESSION:
    PASS (Real Docker Compute Lifecycle integration test 100% PASS)

70D REGRESSION:
    PASS (Real Compute Web Terminal Execution test 100% PASS)

70G TEST SUITE:
    PASS (All 11 subtests of TestPhase70G_ComputeSecretEncryption 100% PASS)

BUILD RESULTS:
    PASS (gateway.exe and anarva.exe binaries built cleanly)

BLOCKING FINDINGS:
    NO BLOCKING FINDINGS.
```

---

## 4. Empirical Test Results

```bash
go test -v ./internal/compute/...
# === RUN   TestPhase70G_ComputeSecretEncryption
# === RUN   TestPhase70G_ComputeSecretEncryption/1._AES-256-GCM_Encryption_&_Decryption_Round_Trip -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/2._Fresh_Nonce_Per_Encryption_(No_Nonce_Reuse) -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/3._Ciphertext_Tampering_Rejection -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/4._Wrong-Key_Decryption_Rejection -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/5._Malformed_Ciphertext_Rejection -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/6._Legacy_Plaintext_JSON_Read_&_Dual-Read_Fallback -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/7._Legacy_Plaintext_Auto_Re-Encryption_on_Save -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/8._API_EnvVars_Secret_Redaction_in_Public_Responses -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/9._Missing_/_Invalid_Production_Key_Validation -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/10._Gateway_Restart_&_Provider_Re-Hydration_with_Encryption -> PASS
# === RUN   TestPhase70G_ComputeSecretEncryption/11._Tenant_Isolation_Regression_Verification -> PASS
# --- PASS: TestPhase70G_ComputeSecretEncryption (0.05s)
# PASS
```
