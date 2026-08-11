# Anarva Cloud Observability, Monitoring & Telemetry Documentation

## 1. Overview & Core Telemetry Abstraction

The **Anarva Cloud Observability System** provides time-series metrics, structured logging streams, alert rule evaluation, and health checking across all control plane services and resource hierarchies.

---

## 2. Real Telemetry Sources Currently Connected

| Telemetry Target | Telemetry Signal | Source Type | Status |
| :--- | :--- | :--- | :--- |
| **Go API Gateway Router** | HTTP Request Count, Response Latency (ms), Status Codes | Real-Time Go Meter | **CONNECTED_REALTIME** |
| **Go Runtime Engine** | Heap Memory Alloc (`HeapAlloc`), Sys Memory, Goroutine Count | Real-Time `runtime.MemStats` | **CONNECTED_REALTIME** |
| **System Health Engine** | Database Pool Connection (`/health`), Liveness (`/livez`), Readiness (`/ready`) | Health Checker | **CONNECTED_REALTIME** |
| **AOS Object Storage** | Local Bucket Objects & Size Bytes | Storage Provider | **CONNECTED_REALTIME** |

---

## 3. Disconnected / Pending Telemetry Sources

| Telemetry Target | Pending Telemetry Signal | Status Indicator in UI |
| :--- | :--- | :--- |
| **Bare-Metal Compute Node** | Physical vCPU Utilization & Node Memory Egress | `Telemetry unavailable - Agent pending` |
| **OpenTelemetry Collector** | Distributed Span Tracing | `Provider pending connection` |

---

## 4. Structured Logging & Secret Redaction

Log records route through `ObservabilityService.RecordLog()`. Sensitive fields matching passwords, secrets, or bearer tokens are automatically redacted prior to storage or UI delivery.

```json
{
  "id": "log-101",
  "service": "gateway-api",
  "level": "INFO",
  "message": "API Gateway initialized with TLS 1.3 encryption & rate limiting middleware",
  "requestId": "req-init-01",
  "traceId": "tr-87a1c9",
  "timestamp": "2026-08-11T07:12:00Z"
}
```

---

## 5. REST API Specifications

- `GET /api/v1/monitoring/overview`: Aggregated telemetry summary and source status flags.
- `GET /api/v1/monitoring/metrics`: Real-time Go runtime and memory allocation metrics.
- `GET /api/v1/monitoring/logs`: Structured log stream with service, level, and query filters.
- `GET /api/v1/monitoring/alerts`: Alert rules and evaluation status.
- `GET /api/v1/monitoring/health`: Detailed subsystem health state.
- `GET /livez`: Gateway liveness endpoint (200 OK).
- `GET /ready`: Gateway readiness endpoint (200 OK).
