# ANARVA CLOUD — DEVELOPER EXPERIENCE & FEEDBACK INTELLIGENCE

This document describes the **Feedback Intelligence Model, REST API Contracts, Console Management UI, and SDK Integration** built in Phase 40.

---

## 🏛️ Architecture & Automatic Context Resolution

```
                  Anarva Console / SDK
                            │
               POST /api/v1/feedback
                            │
              Automatic Context Resolver
       (user_id, organization_id, project_id)
                            │
              Feedback Validation Engine
                 (Rating 1-5, Max 5000 chars)
                            │
               Feedback Persistence Store
                            │
    ┌───────────────────────┼───────────────────────┐
    │                       │                       │
Audit System           Email Dispatcher       Analytics Engine
(FEEDBACK_SUBMITTED) (23w61a0506@gmail.com)  (Aggregates/KPIs)
```

---

## ⚡ Feedback Data Model

| Field | Type | Description |
| :--- | :--- | :--- |
| `feedback_id` | `string` | Unique feedback identifier (`fb-1786801999123-1`) |
| `user_id` | `string` | Authenticated user ID (resolved server-side) |
| `user_email` | `string` | Submitter email address |
| `organization_id` | `string` | Authenticated tenant organization ID (resolved server-side) |
| `project_id` | `string` | Tenant project ID context |
| `rating` | `integer` | User experience rating (1 to 5 Stars) |
| `category` | `string` | `GENERAL`, `BUG_REPORT`, `FEATURE_REQUEST`, `PERFORMANCE`, `USABILITY` |
| `subject` | `string` | Feedback summary title (max 250 chars) |
| `message` | `string` | Detailed message text (max 5000 chars) |
| `status` | `string` | `NEW`, `REVIEWING`, `PLANNED`, `IN_PROGRESS`, `RESOLVED`, `CLOSED` |
| `target_email` | `string` | `23w61a0506@gmail.com` |
| `created_at` | `timestamp` | Submission timestamp |
| `updated_at` | `timestamp` | Status modification timestamp |
| `request_id` | `string` | Tracing Request ID |

---

## 🔒 Security & Tenant Isolation Directives

1. **Automatic Context Resolution**: Users cannot spoof `organization_id` or `user_id` in request payloads; these are extracted from auth session context.
2. **Tenant Isolation**: Organization A can NEVER list or view Organization B feedback submissions.
3. **Audit Integration**: Submissions generate `FEEDBACK_SUBMITTED` events; status changes generate `FEEDBACK_STATUS_UPDATED` audit records.
4. **Secret Protection**: Credentials, bearer tokens, and API keys are NEVER present in feedback records or emails.
