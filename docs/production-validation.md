# Anarva Cloud Platform — Production Validation Scorecard

## Overview
This document certifies production readiness and environment validation rules for deploying the Anarva Cloud Platform.

---

## Production Security Gates

| Check Category | Verification Requirement | Status |
| :--- | :--- | :---: |
| **Fail-Closed Config** | `ValidateProductionConfig` halts startup if port/database configuration is missing when `ENVIRONMENT=production`. | `VERIFIED` |
| **Random JWT Generation**| In production, missing `JWT_SECRET` auto-generates a 256-bit secure secret. | `VERIFIED` |
| **Secret Redaction** | `pkg/security.RedactSecrets()` redacts secrets across all loggers, errors, CLI `--debug`, SDK errors, and Terraform diagnostics. | `VERIFIED` |
| **Database Pool** | GORM PostgreSQL connection pool configured (`MaxOpenConns: 25`, `MaxIdleConns: 5`, `ConnMaxLifetime: 15m`). | `VERIFIED` |
| **Graceful Shutdown** | Catch `SIGINT` / `SIGTERM` with 15s timeout context stopping HTTP server, recovery workers, and DB pool cleanly. | `VERIFIED` |
| **Health Probes** | `/health` returns liveness `200 UP`; `/readiness` checks DB pool ping, returning `503 UNAVAILABLE` when DB is down. | `VERIFIED` |
| **Operation Recovery** | `RecoveryWorker` daemon scans DB on startup to reconcile interrupted operations cleanly. | `VERIFIED` |
| **Prometheus Metrics** | Metrics served on `/metrics` under `anarva_*` namespace. | `VERIFIED` |
