# Anarva Cloud Database Platform 2.0 Documentation

## 1. Overview & Engine Specifications

The **Anarva Cloud Database Platform (DaaS)** manages containerized and clustered PostgreSQL and MySQL instances with automated point-in-time recovery, connection pooling, and multi-region replication specs.

### Supported Engines:
- **PostgreSQL**: Versions `17.2`, `16.4`, `15.8`
- **MySQL**: Versions `8.4.0`, `8.0.36`

---

## 2. Provider Abstraction (`internal/database/provider/provider.go`)

All database operations execute against the `DatabaseProvider` interface:

```go
type DatabaseProvider interface {
    CreateDatabase(ctx context.Context, cluster *DatabaseCluster) (*DatabaseCluster, error)
    GetDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error)
    ListDatabases(ctx context.Context, orgID, projectID string) ([]*DatabaseCluster, error)
    UpdateDatabase(ctx context.Context, id, orgID string, updater func(*DatabaseCluster)) (*DatabaseCluster, error)
    DeleteDatabase(ctx context.Context, id, orgID string) error
    StartDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error)
    StopDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error)
    RestartDatabase(ctx context.Context, id, orgID string) (*DatabaseCluster, error)
}
```

The initial `ControlPlaneDatabaseProvider` manages database cluster resource state machines with strict tenant isolation safeguards.

---

## 3. Provisioning State Machine

Database creation transitions through explicit lifecycle states:

```
REQUESTED -> VALIDATING -> PROVISIONING -> AVAILABLE / FAILED
```

Updates and safe deletions transition:
```
AVAILABLE -> UPDATING -> AVAILABLE
AVAILABLE -> DELETING -> DELETED
```

---

## 4. SQL Workspace Security & Auditing

- **Authorized Context Only**: SQL queries execute in an isolated sandbox context. Control-plane metadata databases are completely isolated.
- **Audit Stream**: Every DDL/DML operation logs a `DATABASE_QUERY_EXECUTED` activity event (with secrets, passwords, and tokens stripped).

---

## 5. REST API Endpoints

- `GET /api/v1/projects/:projectId/databases` — List databases in project.
- `POST /api/v1/databases` — Provision a new managed database cluster.
- `GET /api/v1/databases/:id` — Fetch cluster metadata and status.
- `PATCH /api/v1/databases/:id` — Update compute ACUs, storage GB, or backup parameters.
- `DELETE /api/v1/databases/:id` — Safe deletion of cluster.
- `POST /api/v1/databases/:id/start` — Start stopped database.
- `POST /api/v1/databases/:id/stop` — Graceful stop.
- `POST /api/v1/databases/:id/restart` — Cluster reboot.
