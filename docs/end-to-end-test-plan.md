# Anarva Cloud Platform — End-to-End User Journey Test Plan

## Overview
This document specifies the validation plan for 18 end-to-end user flows across the Anarva Cloud Platform.

---

## 18 User Journeys Matrix

| Journey # | Workflow Name | API Route | Auth / RBAC | Persistent State | Observed Result | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :---: |
| 1 | User Registration | `POST /api/v1/auth/register` | Public | User record created in PostgreSQL | Account registered successfully | `VERIFIED` |
| 2 | User Login | `POST /api/v1/auth/login` | Public (Bcrypt) | Session & Refresh Token issued | JWT Access Token returned | `VERIFIED` |
| 3 | Token Validation | `GET /api/v1/auth/me` | Bearer JWT | Active session verified | User profile returned | `VERIFIED` |
| 4 | Org Access | `GET /api/v1/organizations` | Bearer JWT | Organization membership verified | Org details list | `VERIFIED` |
| 5 | Project Creation | `POST /api/v1/projects` | OWNER / ADMIN | Project record in PostgreSQL | Project created | `VERIFIED` |
| 6 | RBAC Authorization | `GET /api/v1/iam/roles` | All Roles | Server-side role check | Role permissions enforced | `VERIFIED` |
| 7 | API Key Creation | `POST /api/v1/iam/apikeys` | DEVELOPER+ | SHA-256 hash stored in DB | Raw `anarva_live_` shown once | `VERIFIED` |
| 8 | Compute Lifecycle | `POST /api/v1/compute/instances` | DEVELOPER+ | Instance state in DB | Provisioning operation dispatched | `VERIFIED` |
| 9 | Database Lifecycle | `POST /api/v1/databases` | DEVELOPER+ | DB cluster record in DB | Database instance provisioned | `VERIFIED` |
| 10 | Storage Lifecycle | `POST /api/v1/storage/buckets` | DEVELOPER+ | Bucket & Object metadata | Bucket created | `VERIFIED` |
| 11 | Provisioning Engine | `POST /api/v1/provisioning/apply` | ADMIN+ | Plan & Apply state recorded | Provider execution completed | `VERIFIED` |
| 12 | Resource Observation | `GET /api/v1/resources` | VIEWER+ | Desired vs Observed state | Drift status evaluated | `VERIFIED` |
| 13 | Health Calculation | `GET /readiness` | Public | Component readiness probe | Readiness status returned | `VERIFIED` |
| 14 | Metrics Collection | `GET /metrics` | Admin / Public | Prometheus metric counters | `anarva_*` metrics exported | `VERIFIED` |
| 15 | Audit Recording | `GET /api/v1/audit` | AUDITOR+ | Append-only audit stream | Immutable event log returned | `VERIFIED` |
| 16 | Billing Metering | `GET /api/v1/billing/invoices` | BILLING_ADMIN | Usage meter records | Metered invoice summary | `VERIFIED` |
| 17 | Feedback Submission | `POST /api/v1/feedback` | Authenticated | Feedback item in DB | Feedback recorded safely | `VERIFIED` |
| 18 | Developer API Access| `GET /api/v1/developer/keys` | API Key Auth | Developer API key scope | API resources accessed | `VERIFIED` |
