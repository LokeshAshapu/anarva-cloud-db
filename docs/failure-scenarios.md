# Anarva Cloud Platform — Failure Scenarios & Recovery Matrix

## Matrix

| Failure Mode | Detection Mechanism | State Transition | Recovery Action | Verification | Final State | User Result |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Backend Process Crash** | Liveness probe failure / OS SIGKILL | `RUNNING` -> Stale Lease | `RecoveryWorker` daemon scans DB on restart | `GetOperation(id)` | `COMPLETED` or `FAILED` | Safe retry or completed status |
| **DB Connection Loss** | Readiness probe (`/readiness` -> `UNAVAILABLE`) | Database `StatusUnavailable` | Automatic DB pool reconnect ping | `/readiness` returns `READY` | `HEALTHY` | HTTP 503 during outage; normal operations upon reconnect |
| **Operation Timeout** | `DetectOperationTimeouts` daemon scanner | `RUNNING` -> `FAILED` | Operation marked failed, lease released | Timeline audit log entry | `FAILED` | Operation status `FAILED` with timeout reason |
| **Duplicate Concurrent Requests** | Idempotency Store / Lock Engine | Rejected with `HTTP 409` | Active operation executes once | Operation count = 1 | `COMPLETED` | HTTP 409 or cached operation result |
| **Degraded Cloud Provider** | Provider Registry health check | Provider `DEGRADED` | Circuit-breaker error mapping | Provider status endpoint | `DEGRADED` | Safe error response `PROVIDER_UNAVAILABLE` |
