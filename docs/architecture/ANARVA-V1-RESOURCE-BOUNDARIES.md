# ANARVA Cloud V1 — Resource Boundaries Specification

**Audit Date**: August 19, 2026  
**Auditor**: Principal Cloud Architect & Multi-Tenant SaaS Specialist  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  

---

## 1. Complete Resource Boundary Classification Matrix

| Resource | Control Plane Owner | Data Plane Owner | Persistence Engine | Provider Abstraction | Tenant Scope | Current Status | Target Status |
|:---|:---|:---|:---|:---|:---|:---|:---|
| **Users** | `authUsecase` | N/A | PostgreSQL GORM | `user_repository.go` | User ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **Sessions** | `authUsecase` | N/A | PostgreSQL GORM | `session_repository.go` | Session ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **API Keys** | `authUsecase` | N/A | PostgreSQL GORM | `api_key_repository.go` | User ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **Organizations** | `projectUsecase` | N/A | PostgreSQL GORM | `organization_repository.go` | Org ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **Projects** | `projectUsecase` | N/A | PostgreSQL GORM | `project_repository.go` | Org ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **Database Instances**| `databaseUsecase` | Managed Engine | PostgreSQL GORM | `instance_repository.go` | Project ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **PostgreSQL Engine** | `postgresService` | Docker Host | Docker Host Volume | `docker_provider.go` | Container Port | `REAL_LOCAL` | `UPGRADE_TO_RDS_K8S` |
| **MySQL Engine** | `mysqlService` | Docker Host | Docker Host Volume | `docker_provider.go` | Container Port | `REAL_LOCAL` | `UPGRADE_TO_RDS_K8S` |
| **Virtual Networks** | `networkingService` | AWS VPC / K8s | PostgreSQL GORM | `postgres_repository.go` | Project ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **Subnets** | `networkingService` | AWS Subnet | PostgreSQL GORM | `postgres_repository.go` | Project ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **Security Groups** | `networkingService` | AWS SG / FW | PostgreSQL GORM | `postgres_repository.go` | Project ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **Object Buckets** | `storageService` | AWS S3 / R2 | PostgreSQL GORM | `local_storage_provider.go` | Bucket Prefix | `DEVELOPMENT_ONLY` | `UPGRADE_TO_S3_R2` |
| **Object Files** | N/A (Data Plane) | S3 / R2 Store | Remote S3 Bucket | `s3_storage_provider.go` | Bucket Path | `DEVELOPMENT_ONLY` | `UPGRADE_TO_S3_R2` |
| **Compute Instances** | `computeUsecase` | AWS EC2 / Docker | In-Memory Map | `newMemComputeRepo` | Project ID | `SIMULATED` | `MIGRATE_TO_POSTGRES` |
| **Load Balancers** | `lbService` | AWS ALB / NGINX | In-Memory Map | `lb_repository.go` | Project ID | `CONTROL_PLANE_ONLY` | `MIGRATE_TO_POSTGRES` |
| **Backup Metadata** | `backupUsecase` | S3 Storage | PostgreSQL GORM | `backup_repository.go` | Project ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **Backup Archives** | N/A (Data Plane) | S3 Storage | Remote S3 Bucket | `s3_storage_provider.go` | Project ID | `DEVELOPMENT_ONLY` | `UPGRADE_TO_S3_R2` |
| **Reliability Ops** | `reliabilityUsecase`| N/A | PostgreSQL GORM | `reliability_repository.go` | Tenant ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |
| **Audit Logs** | `observabilityService` | N/A | PostgreSQL GORM | `audit_log_repository.go` | Tenant ID | `CONTROL_PLANE_ONLY` | `PRODUCTION_READY` |

---

## 2. Boundary Isolation Policy

1. **Control-Plane Exclusivity**: The control plane stores resource identifiers, status metadata, configuration parameters, and provider resource IDs.
2. **Data-Plane Isolation**: Raw customer data payloads (object file bytes, database table rows, backup archives) MUST reside in data-plane infrastructure (external S3 buckets, dedicated database engines) and never on the control-plane API gateway disk.
