# Anarva Cloud Platform — Prometheus Metrics Reference

## Overview
Anarva exports Prometheus metrics under the `anarva_*` metric namespace via the `/metrics` HTTP endpoint.

---

## Metric Inventory

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `anarva_http_requests_total` | Counter | `status`, `method`, `path` | Total HTTP requests processed by Gateway. |
| `anarva_http_request_duration_seconds` | Histogram | `method`, `path` | Latency distribution of HTTP requests. |
| `anarva_db_queries_total` | Counter | `operation`, `status` | Database query counts. |
| `anarva_db_query_duration_seconds` | Histogram | `operation` | Database query latencies. |
| `anarva_active_connections` | Gauge | None | Number of active client connections. |
| `anarva_operations_total` | Counter | `type`, `status` | Control-plane operations created and completed. |
| `anarva_active_operations` | Gauge | `status` | Current count of in-flight operations. |
| `anarva_active_resource_locks` | Gauge | None | Active lease-based resource locks. |
| `anarva_operation_timeouts_total` | Counter | None | Total operation timeouts detected by scanner. |
| `anarva_operation_recoveries_total` | Counter | `result` | Operation recovery worker reconciliations (`SUCCEEDED` / `FAILED`). |
| `anarva_idempotency_conflicts_total` | Counter | None | Total idempotency key reuse or lock conflict events. |
