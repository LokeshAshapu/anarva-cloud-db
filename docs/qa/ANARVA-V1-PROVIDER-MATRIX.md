# ANARVA Cloud V1 — Provider Architecture & Abstraction Audit

**Audit Date**: August 19, 2026  
**Auditor**: Lead Infrastructure Engineer & Go Backend Architect  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  

---

## 1. Provider Abstraction Architecture

ANARVA utilizes clean Go interfaces for multi-provider infrastructure operations. The current architecture defines interface abstractions for Database Engine Provisioning, Object Storage, Virtual Networking, Load Balancing, Compute, and Infrastructure Simulation.

---

## 2. Provider Abstraction Audit Table

| Infrastructure Subsystem | Interface Definition File | Local Host Implementation | Cloud Provider Implementation | Simulation / Mock Implementation | Active Selection Mechanism | Production Cloud Readiness | Critical Architectural Gap |
|:---|:---|:---|:---|:---|:---|:---|:---|
| **PostgreSQL Provisioner** | `internal/postgres/provider/docker_provider.go` | `LocalDockerPostgresProvider` | None | `MockProvisionerDriver` | Config / Factory | `LOCAL_ONLY` | Lacks AWS RDS / Cloud SQL driver |
| **MySQL Provisioner** | `internal/mysql/provider/docker_provider.go` | `LocalDockerMySQLProvider` | None | `MockProvisionerDriver` | Config / Factory | `LOCAL_ONLY` | Lacks AWS Aurora / GCP Cloud SQL driver |
| **Object Storage Driver** | `internal/storage/provider/local_storage_provider.go` | `LocalStorageProvider` | None | None | Hardcoded `./data/storage` | `NOT_CLOUD_READY` | Lacks S3 / R2 SDK implementation |
| **Network Provider** | `internal/networking/provider/docker_provider.go` | `LocalDockerNetworkProvider` | None | None | Gateway Initialization | `CONTROL_PLANE_ONLY` | Lacks AWS VPC / Cloud Subnet API calls |
| **Load Balancer Provider**| `internal/loadbalancer/provider/docker_provider.go` | `LocalDockerLoadBalancerProvider` | None | None | Gateway Initialization | `CONTROL_PLANE_ONLY` | Lacks AWS ALB / NGINX Ingress driver |
| **Multi-Cloud Integration**| `internal/providers/aws/aws_provider.go` | `LocalSimulationProvider` | `AWSInfrastructureProvider` | `LocalSimulationProvider` | `ProviderRegistry` | `REQUIRES_AWS_KEYS` | Requires real AWS API keys environment variables |
| **Infrastructure HA Engine**| `internal/infrastructure/provider/simulation_provider.go` | None | None | `LocalSimulationProvider` | Gateway Initialization | `SIMULATION_ONLY` | Operates in-memory for demonstration |

---

## 3. Target Future Provider Architecture Diagram

```
                              [ ANARVA Provider Interface ]
                                            │
         ┌──────────────────┬───────────────┴───────────────┬──────────────────┐
         ▼                  ▼                               ▼                  ▼
[ Local Docker Driver ]  [ AWS Provider Driver ]  [ Kubernetes Driver ]  [ Cloudflare Driver ]
(Local Host Dev)        (AWS RDS, S3, VPC)       (K8s Operator, CRDs)   (R2, Workers, Edge)
```

---

## 4. Provider Audit Findings

1. **Local Host Docker Coupling**: Database engine provisioners (`internal/postgres/provider/docker_provider.go` and `internal/mysql/provider/docker_provider.go`) interact directly with the host Docker daemon socket via standard library `exec.Command("docker", ...)`.
2. **AWS Provider Stubs**: The codebase contains structured AWS SDK provider integration (`internal/providers/aws/aws_provider.go`). However, it requires active AWS environment credentials (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`).
3. **Storage Driver Absence**: No S3/R2 API client exists under `internal/storage/provider/`. The system currently relies exclusively on `LocalStorageProvider`.
