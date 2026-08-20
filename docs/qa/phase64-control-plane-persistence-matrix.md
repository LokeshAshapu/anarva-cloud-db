# ANARVA Cloud Phase 64 — Control-Plane Persistence Matrix

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Audit Date**: August 20, 2026  
**Auditor**: Principal Cloud Architect, Distributed Systems Engineer, Go Backend Engineer & SRE  
**Baseline Commit**: `1c7c96a`  

---

## Control-Plane System Persistence Matrix

| Capability | Current Storage | Durable? | Tenant Scoped? | Restart Safe? | Production Status | Required Action |
|:---|:---|:---:|:---:|:---:|:---|:---|
| **Users & Authentication** | PostgreSQL (`users`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **User Sessions** | PostgreSQL (`sessions`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Auth API Keys** | PostgreSQL (`api_keys`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Verification Tokens** | PostgreSQL (`verification_tokens`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Auth Audit Logs** | PostgreSQL (`audit_logs`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Organizations** | PostgreSQL (`organizations`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Projects** | PostgreSQL (`projects`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Project Members** | PostgreSQL (`project_members`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Project Invitations** | PostgreSQL (`project_invitations`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **PostgreSQL Database Instances** | PostgreSQL (`database_instances`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Backup Metadata** | PostgreSQL (`backup_records`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Backup Binary Archives** | S3 / Cloudflare R2 (`ObjectStorageProvider`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Stream payloads to S3) |
| **Object Storage Payloads** | S3 / Cloudflare R2 (`ObjectStorageProvider`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Stream payloads to S3) |
| **Virtual Networks (VPC)** | PostgreSQL (`virtual_networks`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Subnets** | PostgreSQL (`subnets`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Security Groups** | PostgreSQL (`security_groups`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Route Tables** | PostgreSQL (`route_tables`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Network Interfaces** | PostgreSQL (`network_interfaces`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **IP Allocations** | PostgreSQL (`ip_allocations`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Control-Plane Operations** | PostgreSQL (`anarva_operations`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Resource Lock Leases** | PostgreSQL (`resource_lock_leases`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Idempotency Records** | PostgreSQL (`idempotency_records`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Tenant Quotas** | PostgreSQL (`tenant_quotas`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Anarva Audit Events** | PostgreSQL (`anarva_audit_events`) | 🟢 YES | 🟢 YES | 🟢 YES | `PRODUCTION_READY` | None (Keep PostgreSQL GORM) |
| **Compute Instance Metadata** | Go RAM Map (`newMemComputeRepo`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`compute_instances`) |
| **Compute Volumes** | Go RAM Map (`memComputeRepo`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`compute_volumes`) |
| **Load Balancers** | Go RAM Map (`LoadBalancerRepository`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`load_balancers`) |
| **TLS Certificates** | Go RAM Map (`LoadBalancerRepository`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`certificates`) |
| **Custom Domains** | Go RAM Map (`LoadBalancerRepository`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`domains`) |
| **Provisioning Jobs & Steps** | Go RAM Map (`newMemProvisioningRepo`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`provisioning_requests`) |
| **Webhook Endpoints** | Go RAM Map (`WebhookUseCase`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`webhook_endpoints`) |
| **Webhook Deliveries** | Go RAM Slice (`WebhookUseCase`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`webhook_deliveries`) |
| **Developer API Keys** | Go RAM Map (`DeveloperUseCase`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`developer_api_keys`) |
| **Developer Service Accounts**| Go RAM Map (`DeveloperUseCase`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`service_accounts`) |
| **Billing Accounts & Invoices**| Go RAM Map (`BillingUseCase`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`billing_accounts`) |
| **MySQL Database Instances** | Go RAM Map (`MySQLRepository`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`mysql_instances`) |
| **AZ Placement & Evacuation**| Go RAM Map (`InfrastructureRepository`)| 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`placement_policies`) |
| **Cloud Provider Mappings** | Go RAM Map (`MappingRepository`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`provider_resource_mappings`) |
| **Storage Bucket Metadata** | Go RAM Map (`StorageRepository`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`storage_buckets`) |
| **Central Activity Stream** | Go RAM Slice (`Stream`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`activity_events`) |
| **IAM Custom Roles & Policies**| Go RAM Map (`AuthorizationService`) | 🔴 NO | 🟢 YES | 🔴 NO | `EPHEMERAL` | Migrate to GORM (`iam_roles`, `iam_policies`) |
