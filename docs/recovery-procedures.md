# Anarva Cloud Platform — Recovery Procedures

This runbook outlines operational recovery procedures for Anarva engineers.

---

## 1. Recovering Interrupted Control-Plane Operations
1. Restart the Anarva API Gateway process:
   ```bash
   ./bin/gateway
   ```
2. The `RecoveryWorker` daemon automatically scans the control-plane database for stale `RUNNING` operations and reconciles them.
3. Query the operations status endpoint to confirm 0 stuck operations:
   ```bash
   curl -H "Authorization: Bearer <admin-token>" http://localhost:8080/api/v1/operations/summary
   ```

## 2. Recovering Database Connectivity
1. Verify PostgreSQL connection pool health via `/readiness`.
2. If readiness reports `UNAVAILABLE`, check database network connectivity and connection pool metrics.
3. Once PostgreSQL returns online, Anarva automatically resumes servicing API traffic without process restart.

## 3. Restoring Managed Database / Storage Backups
1. List available backups:
   ```bash
   anarva backup list --project <project-id>
   ```
2. Initiate PITR restore:
   ```bash
   anarva backup restore <backup-id> --target-name "restored-db-cluster"
   ```
3. Monitor operation timeline until state reaches `COMPLETED`.
