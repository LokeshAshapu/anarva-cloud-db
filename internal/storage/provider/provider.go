package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type StorageClass string

const (
	ClassStandard         StorageClass = "STANDARD"
	ClassInfrequentAccess StorageClass = "INFREQUENT_ACCESS"
	ClassArchive          StorageClass = "ARCHIVE"
)

type Bucket struct {
	ID                  string       `json:"id"`
	ResourceID          string       `json:"resourceId"`
	OrganizationID      string       `json:"organizationId"`
	ProjectID           string       `json:"projectId"`
	Name                string       `json:"name"`
	RegionID            string       `json:"regionId"`
	Status              string       `json:"status"` // AVAILABLE, CREATING, DELETED
	StorageClass        StorageClass `json:"storageClass"`
	VersioningEnabled   bool         `json:"versioningEnabled"`
	PublicAccessBlocked bool         `json:"publicAccessBlocked"`
	ObjectCount         int          `json:"objectCount"`
	SizeBytes           int64        `json:"sizeBytes"`
	CreatedAt           time.Time    `json:"createdAt"`
	UpdatedAt           time.Time    `json:"updatedAt"`
}

type ObjectItem struct {
	ID              string            `json:"id"`
	BucketID        string            `json:"bucketId"`
	ObjectKey       string            `json:"objectKey"`
	ContentType     string            `json:"contentType"`
	SizeBytes       int64             `json:"sizeBytes"`
	ETag            string            `json:"etag"`
	VersionID       string            `json:"versionId"`
	StorageProvider string            `json:"storageProvider"`
	StoragePath     string            `json:"storagePath"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Checksum        string            `json:"checksum"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

type SignedURLRequest struct {
	BucketName string `json:"bucketName"`
	ObjectKey  string `json:"objectKey"`
	Expiration int    `json:"expirationSeconds"` // e.g. 3600
	Method     string `json:"method"`            // GET, PUT
}

type SignedURLResponse struct {
	SignedURL string    `json:"signedUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

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

type LocalStorageProvider struct {
	buckets map[string]*Bucket
	objects map[string][]*ObjectItem
}

func NewLocalStorageProvider() *LocalStorageProvider {
	p := &LocalStorageProvider{
		buckets: make(map[string]*Bucket),
		objects: make(map[string][]*ObjectItem),
	}
	p.seedDefaults()
	return p
}

func (p *LocalStorageProvider) seedDefaults() {
	now := time.Now()
	b1 := &Bucket{
		ID:                  "res-s3-assets-1",
		ResourceID:          arnv.GenerateARNV("STORAGE_BUCKET", "ap-hyderabad-1", "proj-default", "anarva-media-assets"),
		OrganizationID:      "org-default",
		ProjectID:           "proj-default",
		Name:                "anarva-media-assets",
		RegionID:            "ap-hyderabad-1",
		Status:              "AVAILABLE",
		StorageClass:        ClassStandard,
		VersioningEnabled:   true,
		PublicAccessBlocked: true,
		ObjectCount:         4,
		SizeBytes:           14589000,
		CreatedAt:           now.Add(-48 * time.Hour),
		UpdatedAt:           now,
	}

	p.buckets[b1.ID] = b1
	p.objects[b1.ID] = []*ObjectItem{
		{
			ID:              "obj-101",
			BucketID:        b1.ID,
			ObjectKey:       "avatars/lokesh/profile.png",
			ContentType:     "image/png",
			SizeBytes:       512000,
			ETag:            "\"a3b2c1d4e5\"",
			VersionID:       "v1.0",
			StorageProvider: "LOCAL_AOS",
			StoragePath:     "storage/anarva-media-assets/avatars/lokesh/profile.png",
			Checksum:        "sha256-e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			CreatedAt:       now.Add(-24 * time.Hour),
			UpdatedAt:       now,
		},
		{
			ID:              "obj-102",
			BucketID:        b1.ID,
			ObjectKey:       "documents/contracts/2026/service-agreement.pdf",
			ContentType:     "application/pdf",
			SizeBytes:       2400000,
			ETag:            "\"f9e8d7c6b5\"",
			VersionID:       "v1.0",
			StorageProvider: "LOCAL_AOS",
			StoragePath:     "storage/anarva-media-assets/documents/contracts/2026/service-agreement.pdf",
			Checksum:        "sha256-1c88846c8793b8e788c2ad16a695b28d7085c9eb92f44005b4b1a45a1c726ee8",
			CreatedAt:       now.Add(-12 * time.Hour),
			UpdatedAt:       now,
		},
	}
}

func (p *LocalStorageProvider) CreateBucket(ctx context.Context, b *Bucket) (*Bucket, error) {
	if b.ID == "" {
		b.ID = fmt.Sprintf("bkt-%d", time.Now().UnixNano())
	}
	if b.ResourceID == "" {
		b.ResourceID = arnv.GenerateARNV("STORAGE_BUCKET", b.RegionID, b.ProjectID, b.Name)
	}
	if b.Status == "" {
		b.Status = "AVAILABLE"
	}
	if b.StorageClass == "" {
		b.StorageClass = ClassStandard
	}
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now

	p.buckets[b.ID] = b
	return b, nil
}

func (p *LocalStorageProvider) GetBucket(ctx context.Context, id, orgID string) (*Bucket, error) {
	b, ok := p.buckets[id]
	if !ok || b.Status == "DELETED" {
		return nil, fmt.Errorf("bucket not found")
	}
	if orgID != "" && b.OrganizationID != orgID {
		return nil, fmt.Errorf("authorization violation: cross-tenant access denied")
	}
	return b, nil
}

func (p *LocalStorageProvider) ListBuckets(ctx context.Context, orgID, projectID string) ([]*Bucket, error) {
	var result []*Bucket
	for _, b := range p.buckets {
		if b.Status == "DELETED" {
			continue
		}
		if orgID != "" && b.OrganizationID != orgID {
			continue
		}
		if projectID != "" && b.ProjectID != projectID {
			continue
		}
		result = append(result, b)
	}
	return result, nil
}

func (p *LocalStorageProvider) DeleteBucket(ctx context.Context, id, orgID string) error {
	b, err := p.GetBucket(ctx, id, orgID)
	if err != nil {
		return err
	}
	b.Status = "DELETED"
	b.UpdatedAt = time.Now()
	return nil
}

func (p *LocalStorageProvider) PutObject(ctx context.Context, obj *ObjectItem) (*ObjectItem, error) {
	if obj.ID == "" {
		obj.ID = fmt.Sprintf("obj-%d", time.Now().UnixNano())
	}
	now := time.Now()
	obj.CreatedAt = now
	obj.UpdatedAt = now

	p.objects[obj.BucketID] = append(p.objects[obj.BucketID], obj)
	if b, ok := p.buckets[obj.BucketID]; ok {
		b.ObjectCount++
		b.SizeBytes += obj.SizeBytes
		b.UpdatedAt = now
	}
	return obj, nil
}

func (p *LocalStorageProvider) ListObjects(ctx context.Context, bucketID, prefix string) ([]*ObjectItem, error) {
	list, ok := p.objects[bucketID]
	if !ok {
		return []*ObjectItem{}, nil
	}
	if prefix == "" {
		return list, nil
	}

	var filtered []*ObjectItem
	for _, obj := range list {
		if strings.HasPrefix(obj.ObjectKey, prefix) {
			filtered = append(filtered, obj)
		}
	}
	return filtered, nil
}

func (p *LocalStorageProvider) DeleteObject(ctx context.Context, objectID string) error {
	for bID, list := range p.objects {
		for idx, obj := range list {
			if obj.ID == objectID {
				p.objects[bID] = append(list[:idx], list[idx+1:]...)
				if b, ok := p.buckets[bID]; ok {
					b.ObjectCount--
					b.SizeBytes -= obj.SizeBytes
				}
				return nil
			}
		}
	}
	return fmt.Errorf("object not found")
}

func (p *LocalStorageProvider) GenerateSignedURL(ctx context.Context, req SignedURLRequest) (*SignedURLResponse, error) {
	expSeconds := req.Expiration
	if expSeconds <= 0 {
		expSeconds = 3600
	}
	expiresAt := time.Now().Add(time.Duration(expSeconds) * time.Second)
	token := fmt.Sprintf("sig-%d-%s", expiresAt.Unix(), req.ObjectKey)
	signedURL := fmt.Sprintf("https://aos.anarva.cloud/%s/%s?token=%s&expires=%d", req.BucketName, req.ObjectKey, token, expiresAt.Unix())

	return &SignedURLResponse{
		SignedURL: signedURL,
		ExpiresAt: expiresAt,
	}, nil
}
