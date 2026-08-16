# Anarva Cloud Platform — Request Tracing & Correlation

## Overview
Anarva implements request correlation using `X-Request-ID` and `X-Correlation-ID` headers to trace every request from entry at the API Gateway through authentication, authorization, operation state machine dispatch, provider execution, audit logs, and response headers.

---

## Flow Diagram

```
[ Client Request ] (Optional X-Request-ID: req-123)
       │
       ▼
[ CorrelationMiddleware ] ──► (Generates req-sys-<timestamp> if missing)
       │
       ├─► Inject into Context (`r.Context()`)
       ├─► Set Response Header `X-Request-ID`
       ├─► Attach to Operation (`op.RequestID`)
       └─► Attach to Audit Event (`evt.RequestID`)
```

---

## Tracing Error Responses
Every safe error response includes the request correlation ID:
```json
{
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "Invalid API key [REDACTED_API_KEY]",
    "requestId": "req-custom-trace-101"
  }
}
```
An authorized operator can use the `requestId` to filter structured logs and audit events in the Anarva Console.
