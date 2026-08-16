# Anarva Cloud Platform — RBAC Access Control Matrix

## Overview
This document defines the Role-Based Access Control (RBAC) authorization matrix for the Anarva Control Plane. Every API request is verified server-side against the caller's role, organization context (`organization_id`), project context (`project_id`), and specific resource instance (`id`).

---

## Roles Overview
- **`OWNER`**: Full organization ownership, billing management, role delegation, API key management, and deletion capabilities.
- **`ADMIN`**: Full administrative access across compute, database, storage, and project resources. Cannot transfer organization ownership or override billing profiles.
- **`DEVELOPER`**: Read/write access to resources within authorized projects. Cannot alter IAM roles, billing, or global organization settings.
- **`VIEWER`**: Read-only access across assigned projects and resources. Cannot create, modify, or delete resources.
- **`BILLING_ADMIN`**: Specialized access to view and manage billing profiles, invoices, payment methods, and usage quotas.
- **`AUDITOR`**: Read-only compliance access to audit log streams, security status checks, security event logs, and system metrics.

---

## Authorization Matrix

| Action Category | Specific Action | OWNER | ADMIN | DEVELOPER | VIEWER | BILLING_ADMIN | AUDITOR |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Organization** | Transfer Ownership | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ |
| | Delete Organization | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ |
| | Update Org Settings | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| | View Org Details | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Projects** | Create / Delete Project | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| | Update Project Settings | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | View Projects | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **IAM / Users** | Invite / Remove Members | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| | Change Roles | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| | View Members | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ |
| **API Keys** | Create / Rotate API Key | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | Revoke API Key | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | View API Keys | ✓ | ✓ | ✓ | ✗ | ✗ | ✓ |
| **Compute** | Launch Instance | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | Start / Stop / Reboot | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | Terminate Instance | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| | View Instances | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ |
| **Managed DB** | Provision Database | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | Trigger Failover | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | Delete Database | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| | View Database Status | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ |
| **Storage** | Create Bucket | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | Put / Delete Object | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | Get Object | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| | Delete Bucket | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| **Backups & PITR**| Create Backup | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | Restore Backup | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| | View Backups | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ |
| **Networking** | Create VPC / Subnet | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | Update Firewall Rules | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| | View Network Config | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ |
| **Billing** | Update Payment Method | ✓ | ✗ | ✗ | ✗ | ✓ | ✗ |
| | View Invoices / Usage | ✓ | ✓ | ✗ | ✗ | ✓ | ✓ |
| **Security & Status**| View Security Status | ✓ | ✓ | ✗ | ✗ | ✗ | ✓ |
| | View Security Events | ✓ | ✓ | ✗ | ✗ | ✗ | ✓ |
| **Audit Log** | View Audit Stream | ✓ | ✓ | ✗ | ✗ | ✗ | ✓ |

---

## Server-Side Enforcement Rules
1. **No Frontend-Only Guards**: Every endpoint enforces role checks on the HTTP request handler before calling business logic.
2. **Strict Tenant Query Scoping**: Every database lookup includes `WHERE organization_id = ?` and `project_id = ?`.
3. **Safe Error Rejection**: Unauthorized attempts return `HTTP 403 FORBIDDEN` with error code `INSUFFICIENT_PERMISSIONS`. Cross-tenant lookup attempts return `HTTP 404 NOT_FOUND` to prevent resource existence enumeration.
