# Anarva Cloud Resource Hierarchy & Resource Model Documentation

## 1. Resource Hierarchy

Anarva Cloud structures all resources using a strict 5-tier logical hierarchy:

```
Organization (Anarva Systems)
└── Project (Anarva Cloud Platform)
    └── Environment (Production / Staging / Development)
        └── Region (ap-hyderabad-1)
            └── Resource (production-db)
```

---

## 2. Anarva Resource Identifier (ARNV)

Every cloud resource is identified by a unique, stable, non-sensitive identifier:

```
arnv:<type>:<regionId>:<projectId>:<type-slug>/<name>
```

### Examples:
- Database: `arnv:db:ap-hyderabad-1:proj-default:database/production-db`
- Storage Bucket: `arnv:s3:ap-hyderabad-1:proj-default:storage/anarva-media-assets`
- Compute Node: `arnv:vm:ap-hyderabad-1:proj-default:compute/ace-worker-node-01`
- Network VPC: `arnv:vpc:ap-hyderabad-1:proj-default:network/anarva-primary-vpc`

---

## 3. Data Models (`web/types/resource.ts` & `internal/resource/resource.go`)

### Common Resource Attributes:
- `id`: Internal UUID or resource key.
- `resourceId`: ARNV string.
- `name`: Human-readable resource identifier.
- `type`: `DATABASE`, `STORAGE_BUCKET`, `COMPUTE`, `NETWORK`, `BACKUP`, `REPLICA`.
- `status`: `CREATING`, `AVAILABLE`, `UPDATING`, `DELETING`, `DELETED`, `FAILED`, `STOPPED`, `COMING_SOON`, `MAINTENANCE`.
- `organizationId`: Multi-tenant organization scope key.
- `projectId`: Logical project scope key.
- `environment`: `Production`, `Staging`, `Development`.
- `regionId`: Target deployment region.
- `ownerId`: User identity key.
- `tags`: Key-value annotation pairs.

---

## 4. REST API Endpoints

- `GET /api/v1/organizations` — List organizations.
- `GET /api/v1/projects` — List projects for active organization.
- `POST /api/v1/projects` — Create project scope.
- `GET /api/v1/resources` — Filter and list resources (`organizationId`, `projectId`, `regionId`, `type`, `status`, `query`).
- `POST /api/v1/resources` — Register application-level resource record.
- `GET /api/v1/resources/:id` — Get resource by ID with tenant isolation check.
- `PATCH /api/v1/resources/:id` — Update resource metadata, status, or tags.
- `DELETE /api/v1/resources/:id` — Safe deletion of resource.
- `GET /api/v1/regions` — List available and coming-soon deployment regions.
- `GET /api/v1/activities` — Organization audit event stream.

---

## 5. Security & Isolation Safeguards

1. **Tenant Isolation**: Backend registry verifies `OrganizationID` matching on all `GetByID`, `Update`, and `SafeDelete` operations. Cross-tenant requests return `400 Bad Request` or `404 Not Found`.
2. **IDOR & Enumeration Prevention**: Internal IDs and ARNV strings are validated server-side against authenticated session scopes.
3. **Safe Deletion Framework**: Destructive deletion requires explicit input matching the resource name in the UI dialog (`CloudResourceList.tsx`).
