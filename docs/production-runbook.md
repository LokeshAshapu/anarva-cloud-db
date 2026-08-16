# Anarva Cloud Platform — Production Runbook

This runbook provides step-by-step procedures for managing the Anarva Control Plane in production.

---

## 1. Starting Anarva
```bash
export ENVIRONMENT=production
export SERVER_PORT=8080
export DATABASE_URL="postgres://anarva_admin:password@prod-db.internal:5432/anarva_cloud_db?sslmode=require"
export JWT_SECRET="your-secure-production-jwt-secret-key"

./bin/gateway
```

## 2. Stopping Anarva (Graceful Shutdown)
Send a SIGINT or SIGTERM signal to the process:
```bash
kill -SIGINT <pid>
```
The gateway will:
1. Stop accepting new HTTP requests.
2. Complete in-flight requests within a 15-second grace period.
3. Stop background operation recovery worker daemons cleanly.
4. Close PostgreSQL connection pool connections.
5. Exit cleanly with status 0.

## 3. Restarting Anarva
Execute a graceful stop followed by startup. Upon restart, the `RecoveryWorker` daemon automatically scans the control-plane database for any interrupted or stale `RUNNING` operations and reconciles them.

## 4. Checking Health
```bash
curl -i http://localhost:8080/health
```
Expected: `HTTP 200 OK` with `{"status":"UP", "service":"anarva-control-plane"}`.

## 5. Checking Readiness
```bash
curl -i http://localhost:8080/readiness
```
Expected: `HTTP 200 OK` with `{"status":"READY", "checks":{...}}`.

## 6. Checking System Status
```bash
curl -i -H "Authorization: Bearer <admin-jwt-token>" http://localhost:8080/api/v1/system/status
```

## 7. Checking Platform Version
```bash
curl -i http://localhost:8080/api/v1/version
```

## 8. Investigating Failed Operations
Query the operations endpoint or view the Anarva Operations Center Console at `/console/operations`:
```bash
curl -i -H "Authorization: Bearer <admin-jwt-token>" "http://localhost:8080/api/v1/operations?status=FAILED"
```
To view the step-by-step lifecycle timeline of a specific operation:
```bash
curl -i -H "Authorization: Bearer <admin-jwt-token>" "http://localhost:8080/api/v1/operations/<operation-id>/timeline"
```

## 9. Recovering Operations After Restart
Interrupted operations are automatically reconciled by the `RecoveryWorker` daemon on startup. To manually check stale operations:
```bash
curl -i -H "Authorization: Bearer <admin-jwt-token>" "http://localhost:8080/api/v1/operations/summary"
```

## 10. Handling Configuration Failures
If the Gateway exits immediately on startup with `CONFIG_VALIDATION_FAILURE`:
1. Verify `SERVER_PORT` is set.
2. Verify `DATABASE_URL` is set.
3. Ensure secrets are not hardcoded in config files.

## 11. Handling Degraded Providers
View provider health status via the Provider Registry API or `/console/providers`:
```bash
curl -i -H "Authorization: Bearer <admin-jwt-token>" http://localhost:8080/api/v1/providers
```

## 12. Security Diagnostics API
```bash
curl -i -H "Authorization: Bearer <admin-jwt-token>" http://localhost:8080/api/v1/security/status
```

## 13. Verifying Successful Deployment
1. `/health` returns `200 UP`.
2. `/readiness` returns `200 READY`.
3. Operations Center `/console/operations` loads with 0 unhandled errors.
4. Security Center `/console/security` displays status `SECURE`.
