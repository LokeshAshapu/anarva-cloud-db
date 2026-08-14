package aws

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	storageDomain "github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
)

type AWSS3Provider struct {
	mu       sync.RWMutex
	s3Client S3Client
	enabled  bool
	region   string
	buckets  map[string]*storageDomain.Bucket
	objects  map[string]map[string]*storageDomain.Object
}

func NewAWSS3Provider(s3Client S3Client) *AWSS3Provider {
	awsEnabled := os.Getenv("AWS_ENABLED") == "true"
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	return &AWSS3Provider{
		s3Client: s3Client,
		enabled:  awsEnabled,
		region:   awsRegion,
		buckets:  make(map[string]*storageDomain.Bucket),
		objects:  make(map[string]map[string]*storageDomain.Object),
	}
}

func (p *AWSS3Provider) GetProviderType() string {
	return "AWS_S3"
}

func (p *AWSS3Provider) CreateBucket(ctx context.Context, bucket *storageDomain.Bucket) (*storageDomain.Bucket, error) {
	if !p.enabled {
		return nil, fmt.Errorf("PROVIDER_DISABLED: AWS Provider is currently disabled (AWS_ENABLED=false)")
	}
	if p.s3Client == nil {
		return nil, fmt.Errorf("PROVIDER_NOT_CONFIGURED: AWS S3 client uninitialized")
	}

	s3BucketName := fmt.Sprintf("anarva-s3-%s", bucket.ID)
	if len(s3BucketName) > 63 {
		s3BucketName = s3BucketName[:63]
	}
	s3BucketName = strings.ToLower(s3BucketName)

	reqID := fmt.Sprintf("req-s3-%d", time.Now().UnixNano())
	tags := map[string]string{
		"AnarvaManaged":              "true",
		"AnarvaOrganizationId":       bucket.OrganizationID,
		"AnarvaProjectId":            bucket.ProjectID,
		"AnarvaResourceId":           bucket.ID,
		"AnarvaProvisioningRequestId": reqID,
		"Environment":                "production",
	}

	params := S3CreateParams{
		BucketName: s3BucketName,
		Region:     p.region,
		Tags:       tags,
	}

	info, err := p.s3Client.CreateBucket(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("AWS_S3_CREATE_BUCKET_FAILED: %w", err)
	}

	// Enforce Block Public Access immediately
	if err := p.s3Client.PutPublicAccessBlock(ctx, s3BucketName); err != nil {
		return nil, fmt.Errorf("AWS_S3_SECURITY_BLOCK_FAILED: Failed to enforce Block Public Access on %s: %w", s3BucketName, err)
	}

	bucket.Provider = "AWS_S3"
	bucket.Region = info.Region
	bucket.Status = "ACTIVE"
	bucket.RealityLabel = "AWS S3 (REAL_CLOUD)"
	bucket.CreatedAt = time.Now()
	bucket.UpdatedAt = time.Now()

	p.mu.Lock()
	p.buckets[bucket.ID] = bucket
	if _, exists := p.objects[bucket.ID]; !exists {
		p.objects[bucket.ID] = make(map[string]*storageDomain.Object)
	}
	p.mu.Unlock()

	return bucket, nil
}

func (p *AWSS3Provider) DeleteBucket(ctx context.Context, bucketID string) error {
	if !p.enabled {
		return fmt.Errorf("PROVIDER_DISABLED: AWS Provider is currently disabled")
	}

	p.mu.Lock()
	b, exists := p.buckets[bucketID]
	objs := p.objects[bucketID]
	p.mu.Unlock()

	if !exists {
		return fmt.Errorf("NoSuchBucket: Bucket %s not found", bucketID)
	}

	if len(objs) > 0 {
		return fmt.Errorf("BUCKET_NOT_EMPTY: Cannot delete S3 bucket %s with %d objects remaining", bucketID, len(objs))
	}

	s3Name := fmt.Sprintf("anarva-s3-%s", bucketID)
	if len(s3Name) > 63 {
		s3Name = s3Name[:63]
	}

	if err := p.s3Client.DeleteBucket(ctx, s3Name); err != nil {
		return fmt.Errorf("AWS_S3_DELETE_BUCKET_FAILED: %w", err)
	}

	p.mu.Lock()
	b.Status = "DELETED"
	delete(p.buckets, bucketID)
	delete(p.objects, bucketID)
	p.mu.Unlock()

	return nil
}

func (p *AWSS3Provider) ListBuckets(ctx context.Context, orgID, projectID string) ([]*storageDomain.Bucket, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var res []*storageDomain.Bucket
	for _, b := range p.buckets {
		if b.Status != "DELETED" && (orgID == "" || b.OrganizationID == orgID) {
			res = append(res, b)
		}
	}
	return res, nil
}

func (p *AWSS3Provider) GetBucket(ctx context.Context, bucketID string) (*storageDomain.Bucket, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if b, ok := p.buckets[bucketID]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("NoSuchBucket: Bucket %s not found", bucketID)
}

func (p *AWSS3Provider) GenerateSignedURL(ctx context.Context, bucketID, key, method string, expiresSec int) (*storageDomain.PresignedURL, error) {
	if !p.enabled {
		return nil, fmt.Errorf("PROVIDER_DISABLED: AWS Provider is currently disabled")
	}

	// Step 1: Object Key Path Traversal Security Verification
	cleanKey := path.Clean(key)
	if strings.Contains(key, "../") || strings.Contains(key, "..\\") || strings.HasPrefix(key, "/") || cleanKey == ".." {
		return nil, fmt.Errorf("INVALID_OBJECT_KEY: Unsafe path traversal detected in object key: %s", key)
	}

	s3Name := fmt.Sprintf("anarva-s3-%s", bucketID)
	if len(s3Name) > 63 {
		s3Name = s3Name[:63]
	}

	if expiresSec <= 0 {
		expiresSec = 900 // Default 15 minutes
	}

	var signedURL string
	var err error

	if strings.ToUpper(method) == "PUT" {
		signedURL, err = p.s3Client.GeneratePresignedUploadURL(ctx, s3Name, cleanKey, "application/octet-stream", expiresSec)
	} else {
		signedURL, err = p.s3Client.GeneratePresignedDownloadURL(ctx, s3Name, cleanKey, expiresSec)
	}

	if err != nil {
		return nil, fmt.Errorf("AWS_S3_PRESIGN_FAILED: %w", err)
	}

	return &storageDomain.PresignedURL{
		URL:       signedURL,
		ExpiresAt: time.Now().Add(time.Duration(expiresSec) * time.Second),
		Method:    strings.ToUpper(method),
	}, nil
}

func (p *AWSS3Provider) PutObject(ctx context.Context, bucketID, key string, data io.Reader, size int64, contentType string) (*storageDomain.Object, error) {
	if !p.enabled {
		return nil, fmt.Errorf("PROVIDER_DISABLED: AWS Provider is currently disabled")
	}

	cleanKey := path.Clean(key)
	if strings.Contains(key, "../") || strings.Contains(key, "..\\") || strings.HasPrefix(key, "/") {
		return nil, fmt.Errorf("INVALID_OBJECT_KEY: Path traversal characters forbidden: %s", key)
	}

	obj := &storageDomain.Object{
		ID:          fmt.Sprintf("obj-%d", time.Now().UnixNano()),
		BucketID:    bucketID,
		Key:         cleanKey,
		Size:        size,
		ContentType: contentType,
		Status:      "ACTIVE",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	p.mu.Lock()
	if _, ok := p.objects[bucketID]; !ok {
		p.objects[bucketID] = make(map[string]*storageDomain.Object)
	}
	p.objects[bucketID][cleanKey] = obj
	p.mu.Unlock()

	return obj, nil
}

func (p *AWSS3Provider) GetObject(ctx context.Context, bucketID, key string) (io.ReadCloser, *storageDomain.Object, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if objs, ok := p.objects[bucketID]; ok {
		if obj, found := objs[key]; found {
			return nil, obj, nil
		}
	}
	return nil, nil, fmt.Errorf("NoSuchKey: Object %s not found in bucket %s", key, bucketID)
}

func (p *AWSS3Provider) DeleteObject(ctx context.Context, bucketID, key string) error {
	if !p.enabled {
		return fmt.Errorf("PROVIDER_DISABLED: AWS Provider is currently disabled")
	}

	s3Name := fmt.Sprintf("anarva-s3-%s", bucketID)
	if len(s3Name) > 63 {
		s3Name = s3Name[:63]
	}

	if err := p.s3Client.DeleteObject(ctx, s3Name, key); err != nil {
		return fmt.Errorf("AWS_S3_DELETE_OBJECT_FAILED: %w", err)
	}

	p.mu.Lock()
	if objs, ok := p.objects[bucketID]; ok {
		delete(objs, key)
	}
	p.mu.Unlock()

	return nil
}

func (p *AWSS3Provider) ListObjects(ctx context.Context, bucketID, prefix string) ([]*storageDomain.Object, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var res []*storageDomain.Object
	if objs, ok := p.objects[bucketID]; ok {
		for k, obj := range objs {
			if prefix == "" || strings.HasPrefix(k, prefix) {
				res = append(res, obj)
			}
		}
	}
	return res, nil
}
