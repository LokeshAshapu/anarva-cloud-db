# ANARVA Cloud Phase 63 — Durable Backup & Recovery Forensic Architecture Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 20, 2026  
**Auditor**: Principal Cloud Architect, Go Backend Engineer & SRE  
**Current Baseline**: Phase 62 Complete (Commit `b44a8ce`)  

---

## 1. Objective & Scope

The objective of Phase 63 is to eliminate any production dependency on `./data/backups/` and local filesystem writes for database backup archives. 

Per ANARVA architecture principles:
- **Control-Plane PostgreSQL**: Primary source of truth for backup metadata (`backup_records` table).
- **S3 / Cloudflare R2 (`ObjectStorageProvider`)**: Primary store for database dump archive binary payloads.
- **Data-Plane Execution**: Streaming database dump generation and streaming restoration without memory accumulation or local disk storage.

---

## 2. Current Architecture Forensic Inventory

### 2.1 Metadata Model (`internal/backup/domain/backup.go`)
`BackupRecord` struct currently tracks:
- `ID`: Unique backup identifier (`bak-<timestamp>`)
- `OrganizationID`: Tenant organization context
- `ProjectID`: Tenant project context
- `DatabaseID`: Target database identifier
- `DatabaseName`: Target database name
- `Name`: Human-readable snapshot name
- `Type`: `SNAPSHOT`, `AUTOMATED`, `MANUAL`, `WAL_ARCHIVE`
- `Status`: `REQUESTED`, `QUEUED`, `RUNNING`, `UPLOADING`, `COMPLETED`, `FAILED`, `VERIFIED`, `DELETING`, `DELETED`
- `SizeBytes`: Size of the archive payload in bytes
- `StorageBucket`: Configured S3 bucket name
- `StoragePath`: S3 object key (`backups/organizations/{orgID}/projects/{projectID}/databases/{dbID}/backups/{backupID}/backup.archive`)
- `Checksum`: SHA256 payload checksum
- `StartedAt`, `CompletedAt`, `ExpiresAt`, `CreatedAt`, `UpdatedAt`

### 2.2 Repository Layer (`internal/backup/repository/postgres_repository.go`)
- Uses GORM with PostgreSQL to persist `BackupRecord` entities into `backup_records`.
- Implements `Create`, `GetByID`, `ListByDatabaseID`, `ListByProjectID`, `Update`, `Delete`.

### 2.3 Object Storage Provider Abstraction (`internal/storage/provider/`)
- `ObjectStorageProvider` (`internal/storage/provider/provider.go`) defines standard object storage operations:
  - `PutObject(ctx, bucketID, key, reader, size, contentType)`
  - `GetObject(ctx, bucketID, key)`
  - `DeleteObject(ctx, bucketID, key)`
  - `ListObjects(ctx, bucketID, prefix)`
  - `GenerateSignedURL(ctx, bucketID, key, method, expiresSec)`
- In production (`ANARVA_ENV=production`), `NewStorageProvider` instantiates `S3StorageProvider` (built on AWS SDK v2 supporting AWS S3, Cloudflare R2, MinIO, custom endpoints). In development, `LocalStorageProvider` may be used.

### 2.4 Reliability & Idempotency Framework (`internal/reliability/`)
- `ReliabilityUseCase` tracks operational status transitions for `OpBackupDatabase` and `OpRestoreDatabase`.
- `RecoveryWorker` detects stale/interrupted operations (`IN_PROGRESS` or `RUNNING` past timeout window) and transitions them safely to `FAILED` or triggers automated recovery.

---

## 3. Local-Disk Dependency Forensic Mapping

| File / Component | Current Implementation | Action Required in Phase 63 |
|:---|:---|:---|
| `internal/backup/usecase/backup_usecase.go` | Used legacy `pkg/storage.StorageProvider` and mock dump strings | Upgrade to `internal/storage/provider.ObjectStorageProvider` and stream database dump archives directly to S3/R2 |
| `internal/backup/provider/provider.go` | In-memory `ControlPlaneBackupProvider` mock map | Wire to PostgreSQL repository (`BackupRepository`) and `ObjectStorageProvider` |
| `cmd/gateway/main.go` | Instantiated `ControlPlaneBackupProvider` with mock fallback storage | Wire `sProvider` (`ObjectStorageProvider`) directly into backup service & handlers |
| `docs/` & historical references | Referenced `./data/backups/` | Update documentation; enforce fail-closed check for local filesystem in production |

---

## 4. Key Object Namespace Architecture

Deterministic tenant & project isolated object-key path structure:

```
backups/organizations/{organizationID}/projects/{projectID}/databases/{databaseID}/backups/{backupID}/backup.dump
```

- **Validation**: Enforced via `ValidateObjectKey()` in `internal/storage/provider/provider.go`.
- **Isolation**: Tenant ownership validated against PostgreSQL control-plane metadata before any S3 object key is resolved or accessed.

---

## 5. Security & Isolation Invariants

1. **Authorization First**: Every GET/RESTORE/DELETE endpoint verifies PostgreSQL metadata ownership (`OrganizationID` / `ProjectID` match) prior to issuing S3 calls.
2. **Path Traversal Protection**: Client cannot pass arbitrary object paths or `../` references. Object keys are computed strictly by the server.
3. **No Secret Leakage**: No credentials, secret keys, or signed URLs are emitted in HTTP responses or gateway logs.
4. **Fail-Closed Configuration**: In production (`ANARVA_ENV=production`), backup creation fails immediately if `S3StorageProvider` is not active.

---

## 6. Implementation Strategy & File Plan

- **Files to Modify**:
  - `internal/backup/domain/backup.go` (extend `BackupStatus` with `StatusUploading` if needed)
  - `internal/backup/usecase/backup_usecase.go` (implement streaming backup generation & restore via `ObjectStorageProvider`)
  - `internal/backup/provider/provider.go` (unify `BackupProvider` with `BackupUseCase` and PostgreSQL repository)
  - `internal/backup/delivery/http/backup_handler.go` (enforce tenant context authorization and HTTP responses)
  - `cmd/gateway/main.go` (wire production storage provider and recovery worker)
- **New Files**:
  - `internal/backup/backup_phase63_test.go` (comprehensive Phase 63 unit, security, tenant isolation, streaming, and failure handling test suite)
  - `docs/qa/phase63-backup-recovery-certification.md` (final Phase 63 certification report)

---

## 7. Operational & Risk Assessment

- **Risk**: Ephemeral container restarts during backup upload.
- **Mitigation**: Status transitions (`QUEUED` -> `RUNNING` -> `UPLOADING` -> `COMPLETED`/`FAILED`). `RecoveryWorker` detects stuck `RUNNING`/`UPLOADING` backups beyond timeout window and transitions them to `FAILED`.
