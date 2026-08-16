# Anarva Cloud Platform — Security Incident Response Procedures

## Overview
This runbook defines the official 9-step Security Incident Response Procedure for the Anarva Cloud Platform.

---

## 9-Step Incident Response Lifecycle

### Step 1: Detect
- **Sources**: Automated Security Event Engine (`/api/v1/security/events`), rate limit violation spikes (HTTP 429), failed authentication bursts, or user-submitted vulnerability reports.
- **Triage**: Evaluate severity (CRITICAL, HIGH, MEDIUM, LOW).

### Step 2: Contain
- **API Key Compromise**: Revoke key immediately via CLI (`anarva iam apikey revoke <key-id>`) or Console `/console/security`.
- **Account Compromise**: Invalidate active user sessions and disable API key access.
- **Traffic Isolation**: Update Rate Limit Middleware thresholds or temporarily block abusing IP addresses.

### Step 3: Investigate
- Query the Anarva Security Event Stream (`GET /api/v1/security/events`).
- Query audit logs using request correlation IDs (`X-Request-ID`).
- Trace operation timelines via `/api/v1/operations/<op-id>/timeline`.

### Step 4: Revoke Compromised Credentials
- Rotate JWT signing secret if session token compromise is suspected.
- Revoke all API keys issued under affected user or organization profiles.

### Step 5: Isolate Affected Tenant / Resource
- Lock affected compute instances, database clusters, or storage buckets to prevent data exfiltration.

### Step 6: Review Audit Events
- Verify append-only audit stream for unauthorized mutations:
  - Resource deletions or creations
  - Permission escalation attempts
  - Cross-tenant access queries

### Step 7: Recover
- Restore affected database or storage state from clean PITR backups.
- Re-issue fresh credentials and API keys to verified organization owners.

### Step 8: Validate Recovery
- Run Security Health Status API (`GET /api/v1/security/status`). Ensure all 9 subsystem checks report `SECURE`.
- Execute automated security test suite (`go test -v ./internal/security/...`).

### Step 9: Document & Post-Mortem
- Archive incident summary, root cause analysis, timeline, and remediation actions in internal security logs.

---

## Incident Response Scenarios

### Scenario A: API Key Compromise
1. Identify `key_id` and associated `organization_id` from security log.
2. Execute key revocation in control plane repository.
3. Notify organization owner and issue fresh `anarva_live_` secret.

### Scenario B: Suspected Tenant Isolation Violation Attempt
1. Check Security Event Log for `TENANT_ISOLATION_VIOLATION` events.
2. Confirm request returned `HTTP 404 NOT_FOUND` and zero data was exposed.
3. Temporarily suspend caller's API access pending review.
