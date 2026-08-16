# Anarva Cloud Platform — Operations Intelligence Guide

## Overview
Operations Intelligence allows authorized operators to investigate failures, track operation timelines, and inspect recovery events.

---

## Failure Investigation Path

Starting from a failed operation ID, an operator follows this 6-step path:

1. **Operation Summary**: Query `/api/v1/operations/<id>` to retrieve status (`FAILED`), error code, error message, and timestamps.
2. **Timeline Analysis**: Query `/api/v1/operations/<id>/timeline` to inspect the exact step where execution failed (`VALIDATION_STARTED -> LOCK_ACQUIRED -> PROVIDER_EXECUTION`).
3. **Request ID Correlation**: Extract `requestId` to inspect associated HTTP access logs.
4. **Tenant & Resource Boundary**: Verify `organizationId`, `projectId`, and `resourceId`.
5. **Recovery Verification**: Check `op.Recovery` block to verify whether `RecoveryWorker` attempted automated reconciliation.
6. **Audit Event Trace**: Query `/api/v1/audit` filtering by `operationId` to review actor details and side-effects.
