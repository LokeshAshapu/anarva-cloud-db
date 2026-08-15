# ANARVA CLOUD — REAL INFRASTRUCTURE PROVIDER ARCHITECTURE

This document describes the **Real Infrastructure Provider Execution Architecture** established in Phase 40.

---

## 🏛️ Provider Execution Flow

```
Anarva API / CLI / SDK / Terraform
                 │
  Anarva Control Plane Gateway
                 │
  Tenant Authorization Engine
                 │
   Anarva Provisioning Engine
                 │
  Operation & Lock Lease Engine
                 │
   Anarva Provider Registry
                 │
    ┌────────────┴────────────┐
    ▼                         ▼
Local Execution Engine   Real Cloud Provider Engine
(Docker / Local Storage)  (AWS EC2, RDS, S3, CloudWatch)
    │                         │
    └────────────┬────────────┘
                 ▼
  Observation & Drift Engine
                 │
   Anarva Resource Mapping
```

---

## ⚙️ Execution Modes & Configuration

Anarva supports explicit execution modes configured via environment variables:

| Environment (`ANARVA_ENV`) | Provider Mode (`ANARVA_PROVIDER_MODE`) | Behavior & Safeguards |
| :--- | :--- | :--- |
| `development` | `local` | Executes via local Docker engine & filesystem storage. |
| `staging` | `real` | Dispatches provider calls to Cloud APIs if credentials valid. |
| `production` | `real` | **Fails Closed**. Requires valid `AWS_ACCESS_KEY_ID` & `AWS_SECRET_ACCESS_KEY`. Rejects mock fallbacks. |

---

## 🔒 Security & Credential Protection

1. **Fail-Closed Gatekeeper**: Production mode terminates startup immediately with `PROVIDER_INVALID_CONFIGURATION` if required cloud credentials are missing.
2. **Secret Redaction**: AWS secret access keys, passwords, and `anarva_live_...` API keys are automatically redacted in debug logs and error tracebacks (`[REDACTED_SECRET]`).
3. **Tenant Isolation**: Resource mappings enforce organization scoping (`Organization A` cannot resolve `Organization B` mappings).
4. **Capability Guarding**: Operations check capability matrices (`ValidateCapability`) and return `PROVIDER_CAPABILITY_NOT_SUPPORTED` for un-supported features.
