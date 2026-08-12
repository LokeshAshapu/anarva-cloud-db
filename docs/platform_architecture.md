# ANARVA CLOUD — UNIFIED PLATFORM ARCHITECTURE SPECIFICATION

## Executive Overview
Phase 12 completes the system validation, correlation tracking, resource relationship hierarchy, unified job system, and reality audit for **Anarva Cloud Platform**.

---

## 1. Platform Resource Hierarchy
All platform resources follow a strict hierarchical ownership model:

```
Organization (org-default)
└── Project (proj-default)
    └── Region (us-east-1 / ap-hyderabad-1)
        └── Network (VPC 10.0.0.0/16)
            └── Subnet (10.0.1.0/24 Public / 10.0.2.0/24 Private)
                ├── Compute Instance (ACE Node 1.0 ACU)
                ├── Database Cluster (PostgreSQL 17.2)
                ├── Storage Bucket (AOS S3)
                └── Load Balancer (Application ALB)
```

---

## 2. Resource Identity & Registry
Central resource registry (`internal/resource/resource.go`) indexes all cloud resources with ARNV identifiers:
- `ID`: Unique internal resource identifier (`res-*`)
- `ResourceID`: ARNV identifier (`arnv:<type>:<region>:<project>:<resource>`)
- `OrganizationID`: Tenant ownership boundary (`org-*`)
- `ProjectID`: Project scope boundary (`proj-*`)
- `Type`: `DATABASE`, `STORAGE_BUCKET`, `COMPUTE`, `NETWORK`, `SUBNET`, `VOLUME`, `LOAD_BALANCER`, `BACKUP`, `SNAPSHOT`, `DNS_ZONE`
- `Dependencies`: Pre-deletion dependency check enforcement

---

## 3. Unified Asynchronous Job System
Asynchronous jobs (`internal/job/job.go`) manage platform operations with standard state machine:
- `QUEUED` → `RUNNING` → `SUCCEEDED` / `FAILED` / `CANCELLED` / `RETRYING`
- Exponential backoff retry policies and `Idempotency-Key` header tracking.

---

## 4. Correlation & Tracing Middleware
All HTTP API Gateway requests (`cmd/gateway/main.go`) pass through `CorrelationMiddleware`:
- `X-Request-ID`: Unique tracing token (`req-*`)
- `X-Correlation-ID`: Cross-service correlation token
- `Idempotency-Key`: Duplicate creation prevention header

---

## 5. Reality Status & Provider Labels
- **`REAL`**: Control plane software-defined APIs, CIDR math, IAM authorization, ARNV generator, audit trail.
- **`LOCAL DEVELOPMENT PROVIDER`**: Local Docker container tasks, local filesystem storage, PostgreSQL driver.
- **`CONFIGURED`**: In-memory and control-plane registries.
- **`PROVIDER NOT CONNECTED`**: External AWS EC2/Route53/S3 cloud API drivers.
