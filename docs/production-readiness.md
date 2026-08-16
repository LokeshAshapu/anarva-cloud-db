# Anarva Control Plane — Production Readiness Guide

## Overview
This document outlines the operational requirements, configuration validation rules, database pooling strategies, secret protection mechanisms, and health/readiness probe semantics for deploying the Anarva Cloud Platform in production environments.

---

## 1. Production Configuration Validation
When `ENVIRONMENT=production`, Anarva performs fail-closed startup configuration checks:

- **`SERVER.PORT`**: Must be a positive integer (e.g. `8080`).
- **`JWT.SECRET`**: Must be configured and non-default. If not provided via environment variable `JWT_SECRET`, Anarva auto-generates a 256-bit cryptographically secure random key (`anarva_prod_sec_<hex>`) for the process lifetime.
- **`DATABASE.HOST`** / **`DATABASE_URL`**: Must be configured with a valid PostgreSQL connection pool string in production mode.
- **`PROVIDER.MODE`**: When set to `real`, `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` must be active and validated.

Development fallback conveniences (e.g. in-memory repositories) strictly require `ENVIRONMENT=development` and `ENABLE_DEV_AUTH=true`.

---

## 2. Database Operational Readiness
Anarva uses GORM with PostgreSQL connection pooling:

- **`MAX_OPEN_CONNS`**: 25 (default)
- **`MAX_IDLE_CONNS`**: 5 (default)
- **`CONN_MAX_LIFETIME`**: 15 minutes (default)
- **Startup ping**: The control plane validates DB connectivity on startup. If the database is unreachable in production, startup halts with exit status 1.

---

## 3. Secret Protection & Log Sanitization
Anarva uses `pkg/security.RedactSecrets()` across all log entries, API responses, SDK errors, CLI `--debug` output, and Terraform diagnostics.

Redacted values include:
- `anarva_live_...` / `anarva_test_...` -> `[REDACTED_API_KEY]`
- `whsec_live_...` -> `[REDACTED_WEBHOOK_SECRET]`
- `Bearer <token>` -> `Bearer [REDACTED]`
- JWT tokens (`eyJ...`) -> `[REDACTED_JWT_TOKEN]`
- Connection string passwords -> `postgres://user:[REDACTED]@host:port/db`
- AWS access keys -> `[REDACTED_AWS_KEY]`

---

## 4. Health & Readiness Semantics
- **`/health`**: Process liveness probe. Returns HTTP 200 `{"status":"UP"}`.
- **`/readiness`**: Dependency readiness probe. Checks database pool ping, config validation, provider registry status, and reliability engine state.
  - Returns **HTTP 200 READY** when all dependencies are ready.
  - Returns **HTTP 503 NOT_READY** if PostgreSQL or mandatory configuration fails.
- **`/api/v1/system/status`**: Detailed status breakdown across all 8 platform subsystems.
- **`/api/v1/version`**: Centralized version metadata endpoint returning version, git commit, build time, and go version.
