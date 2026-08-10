package providers

import (
	"context"
	"io"
	"time"
)

// ObjectMetadata encapsulates object properties in AOS (Anarva Object Storage).
type ObjectMetadata struct {
	Bucket       string            `json:"bucket"`
	Key          string            `json:"key"`
	SizeBytes    int64             `json:"size_bytes"`
	ContentType  string            `json:"content_type"`
	ETag         string            `json:"etag"`
	UserMetadata map[string]string `json:"user_metadata"`
	LastModified time.Time         `json:"last_modified"`
}

// BucketPolicy represents access rules for an object storage bucket.
type BucketPolicy struct {
	BucketName string `json:"bucket_name"`
	IsPublic   bool   `json:"is_public"`
	Versioning bool   `json:"versioning"`
}

// StorageProvider defines provider-independent object storage operations.
type StorageProvider interface {
	CreateBucket(ctx context.Context, bucketName string, policy BucketPolicy) error
	DeleteBucket(ctx context.Context, bucketName string) error
	ListBuckets(ctx context.Context, orgID string) ([]string, error)
	PutObject(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) (*ObjectMetadata, error)
	GetObject(ctx context.Context, bucketName, key string) (io.ReadCloser, *ObjectMetadata, error)
	DeleteObject(ctx context.Context, bucketName, key string) error
	ListObjects(ctx context.Context, bucketName, prefix string) ([]*ObjectMetadata, error)
	GenerateSignedURL(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error)
}
