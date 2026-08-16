# Anarva Cloud Platform — Audit Investigation Guide

## Overview
Anarva maintains an append-only, immutable audit stream. Every resource mutation is recorded with actor details, action type, resource ID, operation ID, request ID, and timestamp under strict multi-tenant isolation boundaries (`organization_id`).

---

## Audit Log Structure
```json
{
  "id": "audit-178620000",
  "organizationId": "org-alpha",
  "projectId": "proj-prod",
  "actorType": "USER",
  "actorId": "usr-owner-101",
  "action": "COMPUTE_INSTANCE_TERMINATED",
  "resourceType": "COMPUTE",
  "resourceId": "inst-101",
  "operationId": "op-del-505",
  "requestId": "req-del-707",
  "timestamp": "2026-08-16T11:40:00Z"
}
```

---

## Compliance Query Rules
1. **Tenant Isolation**: Queries by Organization A callers return ONLY Organization A audit events.
2. **Secret Redaction**: Raw API keys, JWT secrets, and passwords are automatically redacted (`pkg/security.RedactSecrets`).
3. **Audited Security Events**:
   - `AUTHENTICATION_FAILURE`
   - `AUTHORIZATION_DENIAL`
   - `API_KEY_CREATED`
   - `API_KEY_REVOKED`
   - `DATABASE_CREATED`
   - `DATABASE_RESTORED`
   - `COMPUTE_TERMINATED`
   - `SECURITY_CONFIG_CHANGED`
