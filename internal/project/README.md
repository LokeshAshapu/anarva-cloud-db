# Project & Tenant Management Service (`internal/project`)

## Architectural Overview
The Project & Tenant Management service handles multi-tenancy hierarchies (**Organizations -> Projects -> Databases**), team invitations, RBAC member management, and project quota enforcement across regions.

```
                  +-----------------------------------+
                  |      Organizations (Tenants)      |
                  +-----------------+-----------------+
                                    |
            +-----------------------+-----------------------+
            |                                               |
            v                                               v
  +-------------------+                           +-------------------+
  | Organization Member|                          |     Projects      |
  |  (RBAC Roles)     |                           |  (Quota Limits)   |
  +-------------------+                           +-------------------+
```

## Multi-Tenancy Rules & Quotas
- Every user belongs to one or more **Organizations**.
- An **Organization** owns multiple **Projects** (isolated environments).
- Default Quotas per Project:
  - **Max Managed Databases**: 5
  - **Max Storage Allocation**: 10 GB
  - **Regions**: `us-east-1`, `us-west-2`, `eu-central-1`, `ap-southeast-1`

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/organizations` | Creates a new organization and assigns owner |
| `GET` | `/api/v1/organizations/{id}` | Retrieves organization details |
| `POST` | `/api/v1/projects` | Creates a project under an organization |
| `GET` | `/api/v1/projects/{id}` | Retrieves project details & quota status |
| `GET` | `/api/v1/organizations/{org_id}/projects` | Lists all projects for an organization |
| `DELETE` | `/api/v1/projects/{id}` | Deletes a project (Requires OWNER or ADMIN role) |
| `POST` | `/api/v1/organizations/{org_id}/invitations` | Invites a member to an organization |
| `POST` | `/api/v1/invitations/accept?token=...` | Accepts a team invitation |
| `GET` | `/api/v1/organizations/{org_id}/members` | Lists organization team members |
