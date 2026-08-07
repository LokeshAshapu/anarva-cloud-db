# Backup, Restore & Object Storage Abstraction Service (`internal/backup` & `pkg/storage`)

## Architectural Overview
The Backup, Restore & Storage service manages database snapshot dumps, point-in-time recovery archives, and object storage persistence using an extensible `StorageProvider` abstraction.

```
                    +------------------------------------+
                    |     Backup & Restore API           |
                    +------------------+-----------------+
                                       |
                                       v
                    +------------------------------------+
                    |       BackupUseCase Engine         |
                    +------------------+-----------------+
                                       |
                   +-------------------+-------------------+
                   |                                       |
                   v                                       v
     +---------------------------+           +---------------------------+
     | StorageProvider           |           | BackupRepository          |
     | (Local Disk / S3 / MinIO) |           | (PostgreSQL Metadata DB)  |
     +---------------------------+           +---------------------------+
```

## Storage Providers (`pkg/storage`)
- **Local Storage Provider**: High performance local filesystem driver with directory isolation.
- **S3 / MinIO Storage Provider**: MinIO / AWS S3 object storage driver for production cloud deployments.

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/backups` | Creates a new snapshot backup dump |
| `GET` | `/api/v1/backups/{id}` | Retrieves backup snapshot metadata |
| `GET` | `/api/v1/databases/{database_id}/backups` | Lists backups for a database instance |
| `POST` | `/api/v1/backups/{id}/restore` | Restores snapshot dump into target database |
| `DELETE` | `/api/v1/backups/{id}` | Deletes snapshot dump archive & metadata |
