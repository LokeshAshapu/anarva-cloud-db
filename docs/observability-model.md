# Anarva Cloud Platform — Production Observability Model

## Executive Summary
This document defines the production observability model for the Anarva Control Plane. Anarva unifies Request Tracing (`X-Request-ID`), Operation Timelines, Append-Only Audit Streams, Unified Resource Health, Prometheus Instrumentation (`anarva_*` namespace), and Error Intelligence into a zero-trust, tenant-isolated architecture.

---

## Observability Architecture

```
[ Request Ingestion ] ──► [ X-Request-ID Correlation Middleware ]
                                     │
       ┌─────────────────────────────┼─────────────────────────────┐
       ▼                             ▼                             ▼
[ Structured Logs ]       [ Prometheus Metrics ]       [ Operation Timeline ]
(Redacted Secrets)          (anarva_* namespace)       (Step-by-step Events)
       │                             │                             │
       └─────────────────────────────┼─────────────────────────────┘
                                     ▼
                      [ Anarva Operations Center ]
```

---

## Key Pillars

### 1. Request Tracing & Correlation
- Every incoming HTTP request is annotated with a unique `X-Request-ID` header.
- The request ID propagates across gateway middleware, authentication, authorization, usecases, provider execution logs, audit events, and API error responses.

### 2. Operation Timelines
- Asynchronous control-plane operations maintain a structured timeline log tracking state transitions:
  `OPERATION_CREATED -> VALIDATION_STARTED -> AUTHORIZATION_PASSED -> LOCK_ACQUIRED -> EXECUTION_STARTED -> PROVIDER_EXECUTION -> RESOURCE_OBSERVED -> HEALTH_VERIFIED -> OPERATION_COMPLETED`.

### 3. Append-Only Audit Stream
- Immutable audit records record every resource mutation (`action`, `actor_id`, `organization_id`, `project_id`, `resource_id`, `request_id`).
- All audit queries strictly enforce `organization_id` tenant scoping.

### 4. Secret Redaction Protection
- Secret redaction utility (`pkg/security.RedactSecrets`) automatically redacts API keys, webhook secrets, passwords, Bearer headers, JWT tokens, DSN passwords, and AWS credentials across all loggers, audit events, CLI output, SDK errors, and Terraform diagnostics.
