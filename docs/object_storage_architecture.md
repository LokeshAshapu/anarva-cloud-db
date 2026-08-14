# Anarva Cloud Object Storage Platform Architecture (Phase 23)

## Overview
Anarva Cloud Phase 23 introduces the **S3-Compatible Object Storage Platform** under `internal/storage/...`. It provides provider-neutral storage accounts, buckets, objects, presigned upload & download URLs, multipart uploads, lifecycle rules, bucket policies, storage access keys, person-centric media links, and S3 compatibility endpoints (`/s3/*`).

---

## Core Storage Subpackages (`internal/storage/...`)
- **`domain`**: `StorageAccount`, `Bucket`, `Object`, `ObjectVersion`, `MultipartUpload`, `MultipartPart`, `ObjectMetadata`, `LifecycleRule`, `RetentionPolicy`, `ObjectLockPolicy`, `BucketReplicationPolicy`, `EncryptionPolicy`, `BucketPolicy`, `StorageAccessKey`, `BucketCORS`, `StorageUsage`, `StorageHealth`.
- **`provider`**: `ObjectStorageProvider` capability contract interface and adapters for `LOCAL_STORAGE` (`REAL_LOCAL`), `AWS_S3`, and `GCP_GCS`.
- **`service`**:
  - `StorageService`: Central orchestrator managing storage accounts, buckets, objects, and activity streams.
  - `SignedURLService`: Presigned upload & download URL generator with HMAC signature & tenant isolation.
  - `MultipartService`: Large file multipart upload orchestrator & automated stale upload cleanup worker.
- **`handler`**: REST API handler (`/api/v1/storage/*`) and S3 compatibility layer (`/s3/*`).

---

## Security & Presigned URL Architecture (`internal/storage/service/signed_url_service.go`)
- **HMAC SHA-256 Signatures**: Presigned URLs encode HTTP method, bucket name, key, and expiration timestamp with HMAC-SHA256 signatures.
- **Tenant Isolation**: Direct object access validates tenant authorization headers (`organizationId`, `projectId`).

---

## Provider Reality Labels
- **Local Storage**: `LOCAL_STORAGE (REAL_LOCAL)`
- **AWS S3**: `AWS_S3 (CONNECTED)` or `AWS_S3 (NOT_CONFIGURED)`
- **GCP GCS**: `GCP_GCS (CONNECTED)` or `GCP_GCS (NOT_CONFIGURED)`
- **S3 Compatibility**: `ANARVA_COMPATIBLE`
