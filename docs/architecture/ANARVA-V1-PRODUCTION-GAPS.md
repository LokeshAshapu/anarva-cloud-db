# ANARVA Cloud V1 — Production Gaps & Remediation Specification

**Audit Date**: August 19, 2026  
**Auditor**: Principal Cloud Architect & SRE Lead  
**Scope**: ANARVA Cloud Control Plane (`LokeshAshapu/anarva-cloud-db`)  

---

## 1. Production Gaps & Blockers Analysis

| Gap Category | Component / File | Current Limitation | Production Impact | Required Remediation | Priority |
|:---|:---|:---|:---|:---|:---:|
| **Storage Persistence** | `internal/storage/provider/local_storage_provider.go` | Object uploads write to `./data/storage` local server disk | Ephemeral disk wipe on Render restart | Upgrade to S3 / Cloudflare R2 SDK driver | **P0** |
| **Backup Storage** | `internal/backup/repository/backup_repository.go` | Dump archives write to `./data/backups/` local server disk | Ephemeral disk wipe on Render restart | Stream dump archives to S3 Bucket | **P0** |
| **Engine Provisioning**| `internal/postgres/provider/docker_provider.go` | Uses host Docker CLI daemon (`docker run`) | Fails in containerized cloud without Docker socket | Implement AWS RDS / Cloud SQL SDK provider | **P1** |
| **Compute State** | `cmd/gateway/main.go:398` | Compute instances stored in memory map (`newMemComputeRepo`) | State lost on gateway restart | Migrate repository to PostgreSQL GORM | **P1** |
| **Load Balancers** | `internal/loadbalancer/repository/lb_repository.go` | Load balancer rules stored in memory map | State lost on gateway restart | Migrate repository to PostgreSQL GORM | **P1** |
| **Webhooks Engine** | `internal/webhook/usecase/webhook_usecase.go` | Webhook subscriptions stored in memory map | State lost on gateway restart | Migrate repository to PostgreSQL GORM | **P2** |

---

## 2. Render Deployment Readiness Checklist

- [x] Fail-closed assertion on missing `DATABASE_URL` in production mode.
- [x] Fail-closed assertion on `DATABASE_URL` resolving to `localhost` / `127.0.0.1`.
- [x] Public safe persistence diagnostic endpoint (`GET /api/v1/health/persistence`).
- [x] Build & commit version tracking (`gitCommit: "phase-60-cd8ca2a"`).
- [ ] Managed S3 Object Storage Provider implementation (Pending Phase 62).
- [ ] AWS RDS / Cloud Database Provisioner driver implementation (Pending Phase 63).
