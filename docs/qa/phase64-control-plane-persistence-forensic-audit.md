# ANARVA Cloud Phase 64 — Control-Plane Persistence Forensic Audit

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 20, 2026  
**Auditor**: Principal Cloud Architect, Distributed Systems Engineer, Go Backend Engineer & SRE  
**Baseline Commit**: `1c7c96a` (Phase 63 Complete)  
**Audit Mode**: READ-ONLY FORENSIC AUDIT (No production code modified)  

---

## 1. Executive Summary & Control-Plane Architecture

The ANARVA Cloud V1 architecture strictly distinguishes between:
- **Control Plane**: Tenants, users, organizations, projects, authentication, security policies, operations, database metadata, network topology, compute instance specifications, load balancer configurations, webhook subscriptions, billing records, provider mappings, and audit logs. All control-plane metadata MUST be durable across process restarts and container redeployments by persisting to the control-plane PostgreSQL database.
- **Data Plane / Cloud Execution**: Physical cloud resources (e.g. AWS EC2, Cloudflare R2, AWS RDS, Docker containers) managed via provider interfaces (`ProviderRegistry`, `InfrastructureProvider`).

### Audit Summary Finding
While Phase 56–63 successfully migrated core authentication, project management, PostgreSQL database metadata, networking topology, operational reliability, and backup metadata to GORM-backed PostgreSQL tables (and S3 for large binary objects), **12 control-plane subsystems remain stored strictly in Go process RAM (in-memory maps and slices)**.

When the gateway process restarts or the production container redeploys on platforms like Render:
- **18 Control-Plane Models** remain fully **DURABLE** in PostgreSQL.
- **12 Control-Plane Subsystems** suffer **TOTAL STATE LOSS** (**EPHEMERAL**).

---

## 2. Complete Inventory of Durable Control-Plane Resources

The following control-plane resources are backed by PostgreSQL GORM models and auto-migrated during gateway startup when `dbPool != nil` (`cmd/gateway/main.go:330-355`):

| Resource / Model | GORM Model File | PostgreSQL Table | Repository Implementation | Restart Safe? |
|:---|:---|:---|:---|:---:|
| **Users** | `internal/auth/domain/auth.go` | `users` | `authRepo.NewUserRepository(dbPool.DB)` | 🟢 YES |
| **Sessions** | `internal/auth/domain/auth.go` | `sessions` | `authRepo.NewSessionRepository(dbPool.DB)` | 🟢 YES |
| **API Keys (Auth)** | `internal/auth/domain/auth.go` | `api_keys` | `authRepo.NewAPIKeyRepository(dbPool.DB)` | 🟢 YES |
| **Verification Tokens**| `internal/auth/domain/auth.go` | `verification_tokens` | `authRepo.NewVerificationTokenRepository(dbPool.DB)` | 🟢 YES |
| **Audit Logs (Auth)** | `internal/auth/domain/auth.go` | `audit_logs` | `authRepo.NewAuditLogRepository(dbPool.DB)` | 🟢 YES |
| **Organizations** | `internal/project/domain/project.go` | `organizations` | `projectRepo.NewOrganizationRepository(dbPool.DB)` | 🟢 YES |
| **Projects** | `internal/project/domain/project.go` | `projects` | `projectRepo.NewProjectRepository(dbPool.DB)` | 🟢 YES |
| **Project Members** | `internal/project/domain/project.go` | `project_members` | `projectRepo.NewMemberRepository(dbPool.DB)` | 🟢 YES |
| **Project Invitations** | `internal/project/domain/project.go` | `project_invitations` | `projectRepo.NewInvitationRepository(dbPool.DB)` | 🟢 YES |
| **PostgreSQL Instances**| `internal/database/domain/database.go` | `database_instances` | `databaseRepo.NewInstanceRepository(dbPool.DB)` | 🟢 YES |
| **Backup Metadata** | `internal/backup/domain/backup.go` | `backup_records` | `backupRepo.NewBackupRepository(dbPool.DB)` | 🟢 YES |
| **Virtual Networks** | `internal/networking/domain/networking.go`| `virtual_networks` | `networkingRepo.NewPostgresNetworkingRepository(dbPool.DB)` | 🟢 YES |
| **Subnets** | `internal/networking/domain/networking.go`| `subnets` | `networkingRepo.NewPostgresNetworkingRepository(dbPool.DB)` | 🟢 YES |
| **Security Groups** | `internal/networking/domain/networking.go`| `security_groups` | `networkingRepo.NewPostgresNetworkingRepository(dbPool.DB)` | 🟢 YES |
| **Route Tables** | `internal/networking/domain/networking.go`| `route_tables` | `networkingRepo.NewPostgresNetworkingRepository(dbPool.DB)` | 🟢 YES |
| **Network Interfaces** | `internal/networking/domain/networking.go`| `network_interfaces` | `networkingRepo.NewPostgresNetworkingRepository(dbPool.DB)` | 🟢 YES |
| **IP Allocations** | `internal/networking/domain/networking.go`| `ip_allocations` | `networkingRepo.NewPostgresNetworkingRepository(dbPool.DB)` | 🟢 YES |
| **Control Operations** | `internal/reliability/domain/reliability.go`| `anarva_operations` | `reliabilityRepo.NewReliabilityRepository(dbPool.DB)` | 🟢 YES |
| **Resource Locks** | `internal/reliability/domain/reliability.go`| `resource_lock_leases` | `reliabilityRepo.NewReliabilityRepository(dbPool.DB)` | 🟢 YES |
| **Idempotency Keys** | `internal/reliability/domain/reliability.go`| `idempotency_records` | `reliabilityRepo.NewReliabilityRepository(dbPool.DB)` | 🟢 YES |
| **Tenant Quotas** | `internal/reliability/domain/reliability.go`| `tenant_quotas` | `reliabilityRepo.NewReliabilityRepository(dbPool.DB)` | 🟢 YES |
| **Anarva Audit Events**| `internal/reliability/domain/reliability.go`| `anarva_audit_events` | `reliabilityRepo.NewReliabilityRepository(dbPool.DB)` | 🟢 YES |

---

## 3. Complete Inventory of Ephemeral Control-Plane Resources (Gaps)

The following 12 control-plane subsystems currently initialize with in-memory Go maps or mock functions in `cmd/gateway/main.go`, causing complete metadata wipe upon gateway restart or Render container redeployment:

### 3.1 Compute Instance Metadata
- **Source Evidence**: `cmd/gateway/main.go:403` -> `compUC := computeUsecase.NewComputeUseCase(newMemComputeRepo(), nil, compProv)`
- **In-Memory Store**: `memComputeRepo.instances map[string]*computeDomain.ComputeInstance` in `cmd/gateway/mock_repos.go:800-904`.
- **API Routes**: `GET/POST/DELETE /api/v1/compute/instances`, `/api/v1/compute/instances/{id}`
- **Restart Impact**: All created compute instances (virtual machine specifications, IP assignments, status, provider IDs) vanish on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.2 Load Balancers & Edge Delivery Metadata
- **Source Evidence**: `cmd/gateway/main.go:444` -> `lbRepository := lbRepo.NewLoadBalancerRepository()`
- **In-Memory Store**: `LoadBalancerRepository` (`lbs`, `listeners`, `pools`, `targets`, `certs`, `domains`, `apps` maps) in `internal/loadbalancer/repository/loadbalancer_repository.go:10-31`.
- **API Routes**: `GET/POST/DELETE /api/v1/loadbalancers`, `/api/v1/certificates`, `/api/v1/domains`, `/api/v1/applications`
- **Restart Impact**: All load balancing rules, listener target groups, SSL/TLS certificates, and application routing configs vanish on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.3 Provisioning Engine Jobs & Step Logs
- **Source Evidence**: `cmd/gateway/main.go:414` -> `provUC := provUsecase.NewProvisioningUseCase(newMemProvisioningRepo(), nil, nil, provRegistry)`
- **In-Memory Store**: `memProvisioningRepo.requests map[string]*provDomain.ProvisioningRequest` in `cmd/gateway/mock_repos.go:990-1048`.
- **API Routes**: `GET/POST /api/v1/provisioning/plans`, `/api/v1/provisioning/requests`
- **Restart Impact**: Active & historical infrastructure provisioning execution step logs and idempotency keys vanish on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.4 Webhooks & Delivery History
- **Source Evidence**: `cmd/gateway/main.go:418` -> `whUC := whUsecase.NewWebhookUseCase()`
- **In-Memory Store**: `WebhookUseCase.endpoints map[string]*domain.WebhookEndpoint` & `deliveries []*domain.WebhookDelivery` in `internal/webhook/usecase/webhook_usecase.go:15-28`.
- **API Routes**: `GET/POST/DELETE /api/v1/webhooks`, `/api/v1/webhooks/endpoints`
- **Restart Impact**: All registered customer webhook endpoint subscriptions, HMAC secrets, and delivery audit histories vanish on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.5 Developer Platform (Service Accounts & API Keys)
- **Source Evidence**: `cmd/gateway/main.go:417` -> `devUC := devUsecase.NewDeveloperUseCase()`
- **In-Memory Store**: `DeveloperUseCase.keys`, `serviceAccounts`, `usageRecords` maps in `internal/developer/usecase/developer_usecase.go:13-29`.
- **API Routes**: `GET/POST/DELETE /api/v1/developer/keys`, `/api/v1/developer/service-accounts`
- **Restart Impact**: All developer service accounts, custom developer API keys, and API rate usage metrics vanish on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.6 Billing Accounts, Usage Records & Invoices
- **Source Evidence**: `cmd/gateway/main.go:421` -> `billUC := billUsecase.NewBillingUseCase()`
- **In-Memory Store**: `BillingUseCase.account`, `profile`, `quotas`, `records`, `invoices` maps in `internal/billing/usecase/billing_usecase.go:13-34`.
- **API Routes**: `GET/POST /api/v1/billing`, `/api/v1/billing/invoices`, `/api/v1/billing/usage`
- **Restart Impact**: Customer billing subscriptions, accumulated usage metrics, pricing components, and generated invoices vanish on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.7 Managed MySQL Database Metadata
- **Source Evidence**: `cmd/gateway/main.go:455` -> `myRepository := mysqlRepo.NewMySQLRepository()`
- **In-Memory Store**: `MySQLRepository.instances`, `databases`, `users`, `backups` maps in `internal/database/mysql/repository/mysql_repository.go:10-25`.
- **API Routes**: `GET/POST/DELETE /api/v1/mysql/instances`
- **Restart Impact**: All managed MySQL database instance specifications, databases, database user grants, and backup entries vanish on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.8 Infrastructure AZ Placement, Evacuation & Incident State
- **Source Evidence**: `cmd/gateway/main.go:462` -> `infRepository := infraRepo.NewInfrastructureRepository()`
- **In-Memory Store**: `InfrastructureRepository.placements`, `haPolicies`, `failovers`, `incidents` maps in `internal/infrastructure/repository/infrastructure_repository.go:9-24`.
- **API Routes**: `GET/POST /api/v1/infrastructure/placement`, `/api/v1/infrastructure/incidents`
- **Restart Impact**: Availability zone placement policies, host node evacuations, outage simulation state, and infrastructure incidents vanish on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.9 Cloud Provider Resource Mappings & Drift Detection Logs
- **Source Evidence**: `cmd/gateway/main.go:474` -> `prvMapRepo := prvMapping.NewMappingRepository()`
- **In-Memory Store**: `MappingRepository.mappings map[string]*ProviderResourceMapping` in `internal/providers/mapping/mapping.go:25-34`.
- **API Routes**: `GET/POST /api/v1/providers/mappings`, `/api/v1/providers/drift`
- **Restart Impact**: **CRITICAL DATA SEVERANCE**: Mappings linking ANARVA control-plane resource UUIDs to actual AWS/Cloud provider resource IDs (e.g. `arn:aws:ec2:i-123456`) vanish on restart, producing orphaned cloud infrastructure!
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.10 Managed Storage Subsystem Bucket & Object Metadata
- **Source Evidence**: `cmd/gateway/main.go:482` -> `stgRepository := stgRepo.NewStorageRepository()`
- **In-Memory Store**: `StorageRepository.accounts`, `buckets`, `objects`, `keys` maps in `internal/storage/repository/storage_repository.go:10-25`.
- **API Routes**: `GET/POST/DELETE /api/v1/storage/buckets`, `/api/v1/storage/objects`
- **Restart Impact**: While Phase 62/63 handles actual S3/R2 binary object streams, the legacy `/api/v1/storage/` metadata repository maintains bucket/object catalog records in RAM maps that vanish on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.11 Centralized Activity Stream & Event Log
- **Source Evidence**: `cmd/gateway/main.go:398` -> `actStream := activity.NewStream()`
- **In-Memory Store**: `Stream.events []*ActivityEvent` in `internal/activity/activity.go:55-65`.
- **API Routes**: `GET /api/v1/activity`, `/api/v1/audit`
- **Restart Impact**: Real-time activity events recorded across all control-plane operations reset on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

### 3.12 IAM Roles & Custom Permission Policies
- **Source Evidence**: `cmd/gateway/main.go:399` -> `authSvc := iamService.NewAuthorizationService()`
- **In-Memory Store**: `AuthorizationService.orgs`, `projects`, `members`, `invitations` maps in `internal/iam/service/authorization.go:11-33`.
- **API Routes**: `GET/POST /api/v1/iam/roles`, `/api/v1/iam/policies`
- **Restart Impact**: Custom IAM roles and assigned permission policies reset on process restart.
- **Classification**: `EPHEMERAL` / `SIMULATED`

---

## 4. Required PostgreSQL Models & Schema Migrations

To convert all 12 ephemeral subsystems into durable control-plane state, the following GORM models must be migrated to PostgreSQL:

1. **`ComputeInstance` & `ComputeVolume`** (`internal/compute/domain/compute.go`)
   - Table: `compute_instances`, `compute_volumes`
   - Migrations required: GORM AutoMigrate for `ComputeInstance` and `Volume`.
2. **`LoadBalancer`, `Listener`, `BackendPool`, `BackendTarget`, `Certificate`, `Domain`** (`internal/loadbalancer/domain/`)
   - Table: `load_balancers`, `load_balancer_listeners`, `backend_pools`, `backend_targets`, `certificates`, `domains`
3. **`ProvisioningRequest` & `ProvisioningStep`** (`internal/provisioning/domain/`)
   - Table: `provisioning_requests`, `provisioning_steps`
4. **`WebhookEndpoint` & `WebhookDelivery`** (`internal/webhook/domain/`)
   - Table: `webhook_endpoints`, `webhook_deliveries`
5. **`DeveloperAPIKey`, `ServiceAccount`, `APIUsageRecord`** (`internal/developer/domain/`)
   - Table: `developer_api_keys`, `service_accounts`, `api_usage_records`
6. **`BillingAccount`, `BillingProfile`, `Invoice`, `UsageRecord`** (`internal/billing/domain/`)
   - Table: `billing_accounts`, `billing_profiles`, `invoices`, `usage_records`
7. **`MySQLInstance`, `MySQLDatabase`, `MySQLUser`** (`internal/database/mysql/domain/`)
   - Table: `mysql_instances`, `mysql_databases`, `mysql_users`
8. **`PlacementPolicy`, `FailoverPolicy`, `InfrastructureIncident`** (`internal/infrastructure/domain/`)
   - Table: `placement_policies`, `failover_policies`, `infrastructure_incidents`
9. **`ProviderResourceMapping`** (`internal/providers/mapping/mapping.go`)
   - Table: `provider_resource_mappings` (**CRITICAL**)
10. **`StorageAccount`, `StorageBucket`, `StorageObject`** (`internal/storage/domain/`)
    - Table: `storage_accounts`, `storage_buckets`, `storage_objects`
11. **`ActivityEvent`** (`internal/activity/activity.go`)
    - Table: `activity_events`
12. **`IAMRole`, `IAMPolicy`** (`internal/iam/domain/`)
    - Table: `iam_roles`, `iam_policies`

---

## 5. Tenant Isolation, Idempotency & Recovery Worker Implications

- **Tenant Isolation**: Every new PostgreSQL repository MUST enforce strict `OrganizationID` and `ProjectID` scoping on all `Get`, `List`, `Update`, and `Delete` queries to prevent cross-tenant access.
- **Idempotency**: All creation requests (compute instance creation, provisioning plans, webhook registrations, provider mapping associations) MUST integrate with `internal/reliability` (`idempotency_records`).
- **Recovery Worker**: Stale `PROVISIONING` or `RECONCILING` states for compute instances, load balancers, and MySQL instances will be automatically picked up by `RecoveryWorker` upon Gateway restart to reconcile state with underlying cloud providers.

---

## 6. Recommended Implementation Order for Phase 64 Execution

1. **P0 (CRITICAL - Cloud Resource Association)**: `ProviderResourceMapping` (`provider_resource_mappings` table) — prevents control plane from losing track of physical AWS/Cloud resources on restart.
2. **P0 (CRITICAL - Compute Platform)**: `ComputeInstance` & `ComputeVolume` (`compute_instances` table) — persists VM specifications & metadata.
3. **P1 (HIGH - Database & Storage Metadata)**: `MySQLInstance` & `StorageBucket`/`StorageObject` — persists MySQL instances and storage bucket metadata.
4. **P1 (HIGH - Load Balancing & Networking)**: `LoadBalancer`, `Listener`, `BackendPool`, `Certificate`, `Domain` — persists ingress routing & TLS certificates.
5. **P2 (MEDIUM - Developer & Integration)**: `WebhookEndpoint`, `DeveloperAPIKey`, `ServiceAccount` — persists integrations & developer credentials.
6. **P2 (MEDIUM - Provisioning & Operations)**: `ProvisioningRequest`, `ActivityEvent`, `InfrastructureIncident` — persists execution history & audit streams.
7. **P3 (STANDARD - Business Systems)**: `BillingAccount`, `Invoice`, `UsageRecord` — persists billing & customer subscriptions.
