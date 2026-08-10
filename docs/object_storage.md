# Anarva Object Storage (AOS) System Architecture Documentation

## 1. Overview & Core Abstractions

**Anarva Object Storage (AOS)** provides scalable S3-compatible bucket and object management across all Anarva deployment regions.

### Core Hierarchy:
```
Organization (Anarva Systems)
└── Project (Anarva Cloud Platform)
    └── Bucket (anarva-media-assets)
        └── Object Key (avatars/lokesh/profile.png)
```

---

## 2. Person-Centric Application Layer Mapping

To preserve compatibility with person-centric entity models while maintaining enterprise object storage standards:

```
Person (John Doe)
└── Application References
    └── AOS Bucket (anarva-media-assets)
        └── Object Key (avatars/user-001/profile.png)
```

This allows Anarva Cloud to serve both **Developer S3 Object Storage** and **Person-Centric Media Management** seamlessly.

---

## 3. Storage Provider Abstraction (`internal/storage/provider/provider.go`)

All storage operations route through the `ObjectStorageProvider` interface:

```go
type ObjectStorageProvider interface {
    CreateBucket(ctx context.Context, bucket *Bucket) (*Bucket, error)
    GetBucket(ctx context.Context, id, orgID string) (*Bucket, error)
    ListBuckets(ctx context.Context, orgID, projectID string) ([]*Bucket, error)
    DeleteBucket(ctx context.Context, id, orgID string) error
    PutObject(ctx context.Context, obj *ObjectItem) (*ObjectItem, error)
    ListObjects(ctx context.Context, bucketID, prefix string) ([]*ObjectItem, error)
    DeleteObject(ctx context.Context, objectID string) error
    GenerateSignedURL(ctx context.Context, req SignedURLRequest) (*SignedURLResponse, error)
}
```

The initial `LocalStorageProvider` implements local development object management with strict tenant isolation.

---

## 4. Signed URL Architecture & Security

Private object access uses cryptographically signed temporary authorization tokens:

```http
POST /api/v1/storage/objects/:id/signed-url
{
  "bucketName": "anarva-media-assets",
  "objectKey": "documents/contracts/2026/service-agreement.pdf",
  "expirationSeconds": 3600,
  "method": "GET"
}
```

Returns:
```json
{
  "signedUrl": "https://aos.anarva.cloud/anarva-media-assets/documents/contracts/2026/service-agreement.pdf?token=sig-1770742395-documents&expires=1770742395",
  "expiresAt": "2026-08-10T23:53:15Z"
}
```

---

## 5. REST API Specifications

- `GET /api/v1/storage/buckets` — List buckets for organization/project.
- `POST /api/v1/storage/buckets` — Provision a new AOS storage bucket.
- `GET /api/v1/storage/buckets/:id` — Retrieve bucket details.
- `DELETE /api/v1/storage/buckets/:id` — Safe deletion of bucket.
- `GET /api/v1/storage/buckets/:id/objects` — List objects by bucket ID and optional key prefix.
- `POST /api/v1/storage/buckets/:id/objects` — Register metadata for uploaded object.
- `DELETE /api/v1/storage/objects/:id` — Delete object.
- `POST /api/v1/storage/objects/:id/signed-url` — Generate secure signed URL for object download.
