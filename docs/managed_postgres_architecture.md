# Managed PostgreSQL Database Platform Architecture (Phase 17)

## Overview
Anarva Cloud Phase 17 introduces a production-ready **Managed PostgreSQL Database Platform Abstraction** under `internal/postgres/...`. It decouples PostgreSQL service orchestration from underlying infrastructure providers, driving local Docker PostgreSQL containers (`LOCAL_POSTGRES`) and exposing interfaces for managed cloud providers.

---

## Key Architectural Components

### 1. Domain & Provider Decoupling (`internal/postgres/...`)
- **Domain Models (`internal/postgres/domain/postgres.go`)**:
  - `PostgresInstance`: Core database instance metadata including version, CPU/Memory ACUs, storage size, availability mode (`SINGLE`, `PRIMARY_STANDBY`, `MULTI_ZONE`), and provider status.
  - `CredentialReference`: Zero-trust credential wrapper. Passwords and connection secrets are never stored in plaintext columns.
  - `ConnectionInfo`: Parameterized host, port, dbname, and secret reference.
  - `DatabaseHealth`: Real-time connection availability, CPU/Memory/Storage metrics, latency, and cache hit ratios with quality labels (`ACTUAL`, `ESTIMATED`, `UNKNOWN`).

- **Provider Interface (`internal/postgres/provider/provider.go`)**:
  - Exposes standardized capability contract: `CreateInstance`, `UpdateInstance`, `DeleteInstance`, `StartInstance`, `StopInstance`, `RestartInstance`, `GetHealth`, `GetMetrics`, `GetLogs`, `CreateDatabase`, `CreateUser`, `RotateCredentials`, `CreateBackup`, `RestoreBackup`, `ScaleInstance`.

- **Local Docker Driver (`internal/postgres/provider/docker_provider.go`)**:
  - Real implementation leveraging Docker container runtime (`postgres:17-alpine`), container port binding, `pg_isready` health probing, and database/user initialization.

---

## Security & Credential Rules
1. **Zero Plaintext Passwords**:
   - Connection strings and credentials use `CredentialReference` and are revealed only via explicit, authorized, one-time reveal endpoint calls.
2. **Backend SQL Proxy (`internal/postgres/service/sql_service.go`)**:
   - Direct browser-to-PostgreSQL connections are prohibited.
   - Enforces 5s statement timeouts, 1000-row response caps, 2MB payload limits, parameterized metadata queries, and dangerous statement protection.
3. **Public Access Security**:
   - PostgreSQL instances default to `PRIVATE` network attachment. Enabling public access requires explicit firewall and IAM confirmation.

---

## Provisioning Pipeline
```
12-Step Console Wizard → REST API (/api/v1/databases) → IAM Check → Quota Reservation → Provisioning Engine → PostgresProvider Driver → Health Probe (pg_isready) → Resource Registration
```

---

## Provider Reality Labels
- **Local Development**: `LOCAL_POSTGRES (DOCKER_SIM)`
- **Cloud Provider**: `MANAGED_POSTGRES_CONNECTED`
- **Unconfigured**: `MANAGED_POSTGRES_NOT_CONFIGURED`
