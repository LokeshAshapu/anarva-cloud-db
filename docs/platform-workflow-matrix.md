# ANARVA Cloud Control Plane — Platform Workflow Matrix

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Document**: Workflow Execution Trace & Provider Chain Matrix  

---

## Complete End-to-End Workflow Execution Chains

### 1. PostgreSQL Database Provisioning & Stateful SQL Execution
```
UI Submit (Console /databases)
  └─► POST /api/v1/databases
       └─► Auth Middleware (JWT / dev-token / supa-session)
            └─► RBAC & Tenant Scoping (OrgID / ProjectID)
                 └─► PostgresHandler.CreateInstance
                      └─► PostgresService.CreateInstance
                           └─► LocalDockerPostgresProvider.CreateInstance
                                └─► PostgresRepository.Save
                                     └─► POST /api/v1/databases/{id}/query (SQL Execution)
                                          └─► SQLService.ExecuteQuery
                                               └─► Stateful AST Parser (DDL/DML/DCL)
                                                    └─► Disk File Sync (anarva_sql_service_state.json)
                                                         └─► GET /api/v1/databases/{id} (Read-Back & UI Persistence)
```
- **Reality Classification**: **REAL LOCAL**

---

### 2. MongoDB Document Querying (MQL)
```
UI Submit (Console /databases MQL Tab)
  └─► POST /api/v1/databases/{id}/mql
       └─► Auth & Tenant Validation
            └─► MQLService.ExecuteMQL
                 └─► Parse db.<collection>.<op>() AST
                      └─► Update CollectionState (JSON BSON Store)
                           └─► Save to ./data/anarva_mql_service_state.json
                                └─► Return Documents & UI Render
```
- **Reality Classification**: **REAL LOCAL**

---

### 3. Redis In-Memory Command Execution
```
UI Submit (Console /databases Redis Tab)
  └─► POST /api/v1/databases/{id}/redis
       └─► Auth & Tenant Validation
            └─► RedisService.ExecuteCommand
                 └─► Parse Command (SET, GET, KEYS, DEL, PING)
                      └─► Update Store Map
                           └─► Save to ./data/anarva_redis_service_state.json
                                └─► Return Command Result & UI Render
```
- **Reality Classification**: **REAL LOCAL**

---

### 4. S3 Object Storage Lifecycle
```
UI Upload Object (Console /storage)
  └─► PUT /api/v1/storage/objects
       └─► Auth & Tenant Scoping
            └─► StorageHandler.PutObject
                 └─► StorageUseCase.PutObject
                      └─► ValidateObjectKey (Path Traversal Security Check)
                           └─► LocalStorageProvider.PutObject
                                └─► Write Bytes to Disk (anarva-local-storage/<bucket>/<key>)
                                     └─► GET /api/v1/storage/objects (Read-Back Bytes & Download)
```
- **Reality Classification**: **REAL LOCAL**

---

### 5. Compute Instance Provisioning & Command Execution
```
UI Submit (Console /compute)
  └─► POST /api/v1/compute/instances
       └─► Auth & Tenant Scoping
            └─► ComputeHandler.CreateInstance
                 └─► ComputeUseCase.CreateInstance
                      └─► LocalDockerComputeProvider.CreateInstance
                           └─► exec.Command("docker", "run", ...) [if Docker present]
                                └─► ComputeRepository.Save
                                     └─► GET /api/v1/compute/instances/{id}
```
- **Reality Classification**: **REAL LOCAL / SIMULATED**

---

### 6. Control-Plane Operation Recovery
```
Background Operation Recovery Daemon (Every 50ms)
  └─► RecoveryWorker.Run
       └─► ReliabilityUseCase.ReconcileStaleOperations
            └─► LockLease & Idempotency Key Validation
                 └─► AuditLog.Append
                      └─► Update Operation State to COMPLETED
```
- **Reality Classification**: **REAL LOCAL**
