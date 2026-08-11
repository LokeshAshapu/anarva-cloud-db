# Anarva Cloud Backup, Snapshots & Disaster Recovery Documentation

## 1. Overview & Core Architecture

The **Anarva Cloud Backup & Disaster Recovery Engine** manages database snapshots, object storage retention, and asynchronous restore job state machines.

### Integration Pipeline:
```
DatabaseCluster (Phase 5)
└── ControlPlaneBackupProvider (Phase 9)
    └── LocalStorageProvider / AOS Object Storage (Phase 6)
        └── Backup Object (backups/org-default/proj-default/production-db/snapshots/daily-20260810.snap)
```

---

## 2. Feature Status & Provider Integration Matrix

| Capability | Control Plane Status | Infrastructure Provider Requirement | UI Display Status |
| :--- | :--- | :--- | :--- |
| **Control-Plane Snapshots** | **REAL & ACTIVE** | Integrated with AOS Object Storage (`anarva-media-assets`) | **VERIFIED** |
| **Restore to New Database** | **REAL & ACTIVE** | Asynchronous restore job state machine | **COMPLETED** |
| **Backup Retention Policy** | **CONFIGURED** | Control-plane lifecycle rules (7-90 days) | **CONFIGURED** |
| **WAL Archival & PITR** | **PROVIDER NOT CONNECTED** | Requires physical PostgreSQL WAL streaming daemon | **PROVIDER NOT CONNECTED** |
| **Cross-Region Failover** | **COMING SOON** | Bare-metal replication driver | **COMING SOON** |

---

## 3. Restore Job State Machine

Restore jobs transition through explicit lifecycle states:

```
REQUESTED -> VALIDATING -> QUEUED -> RESTORING -> VERIFYING -> COMPLETED / FAILED
```

---

## 4. REST API Specifications

- `GET /api/v1/databases/:id/backups`: Fetch backups & snapshots for specified database cluster.
- `POST /api/v1/databases/:id/backups`: Provision a new manual snapshot stored in AOS Object Storage.
- `GET /api/v1/backups/:id`: Retrieve backup metadata and SHA-256 integrity checksum.
- `DELETE /api/v1/backups/:id`: Safe deletion of snapshot.
- `POST /api/v1/backups/:id/restore`: Submit an asynchronous restore job into a target database cluster.
- `GET /api/v1/databases/:id/recovery-points`: Check continuous WAL recovery point availability.
- `GET /api/v1/databases/:id/backup-config`: Fetch automated backup retention window configuration.
