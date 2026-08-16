package provider

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
)

type ObjectStorageProvider interface {
	GetProviderType() string
	CreateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error)
	DeleteBucket(ctx context.Context, bucketID string) error
	ListBuckets(ctx context.Context, orgID, projectID string) ([]*domain.Bucket, error)
	GetBucket(ctx context.Context, bucketID string) (*domain.Bucket, error)

	PutObject(ctx context.Context, bucketID, key string, data io.Reader, size int64, contentType string) (*domain.Object, error)
	GetObject(ctx context.Context, bucketID, key string) (io.ReadCloser, *domain.Object, error)
	DeleteObject(ctx context.Context, bucketID, key string) error
	ListObjects(ctx context.Context, bucketID, prefix string) ([]*domain.Object, error)

	GenerateSignedURL(ctx context.Context, bucketID, key, method string, expiresSec int) (*domain.PresignedURL, error)
}

// ValidateObjectKey protects against path traversal attempts in storage keys.
func ValidateObjectKey(key string) error {
	if key == "" {
		return fmt.Errorf("STORAGE_SECURITY_RISK: Object key cannot be empty")
	}

	// 1. Check for null bytes
	if strings.Contains(key, "\x00") {
		return fmt.Errorf("STORAGE_SECURITY_RISK: Object key contains illegal null bytes")
	}

	// 2. Decode URL encoding if present
	decodedKey := key
	if unescaped, err := url.QueryUnescape(key); err == nil {
		decodedKey = unescaped
	}

	// 3. Check for URL-encoded traversal patterns
	lowerKey := strings.ToLower(key)
	if strings.Contains(lowerKey, "%2e%2e%2f") || strings.Contains(lowerKey, "%2e%2e/") ||
		strings.Contains(lowerKey, "%2e%2e%5c") || strings.Contains(lowerKey, "%2e%2e\\") {
		return fmt.Errorf("STORAGE_SECURITY_RISK: Object key contains illegal URL-encoded path traversal")
	}

	// 4. Check for direct traversal patterns
	if strings.Contains(decodedKey, "../") || strings.Contains(decodedKey, "..\\") ||
		strings.HasPrefix(decodedKey, "/") || strings.HasPrefix(decodedKey, "\\") || filepath.IsAbs(decodedKey) {
		return fmt.Errorf("STORAGE_SECURITY_RISK: Object key '%s' attempts directory traversal outside storage root", key)
	}

	// 5. Clean path check
	cleaned := filepath.Clean(decodedKey)
	if strings.HasPrefix(cleaned, "..") || strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "\\") {
		return fmt.Errorf("STORAGE_SECURITY_RISK: Object key '%s' attempts directory traversal outside storage root", key)
	}

	return nil
}

type LocalStorageProvider struct {
	mu      sync.RWMutex
	baseDir string
	buckets map[string]*domain.Bucket
	objects map[string]*domain.Object
}

func NewLocalStorageProvider(baseDir string) *LocalStorageProvider {
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "anarva-local-storage")
	}
	_ = os.MkdirAll(baseDir, 0755)

	return &LocalStorageProvider{
		baseDir: baseDir,
		buckets: make(map[string]*domain.Bucket),
		objects: make(map[string]*domain.Object),
	}
}

func (p *LocalStorageProvider) GetProviderType() string {
	return "LOCAL_STORAGE"
}

func (p *LocalStorageProvider) CreateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	bucket.Provider = "LOCAL_STORAGE"
	bucket.Status = "ACTIVE"
	bucket.RealityLabel = "LOCAL_STORAGE (REAL_LOCAL)"
	bucket.CreatedAt = time.Now()
	bucket.UpdatedAt = time.Now()

	bucketDir := filepath.Join(p.baseDir, bucket.Name)
	_ = os.MkdirAll(bucketDir, 0755)

	p.buckets[bucket.ID] = bucket
	return bucket, nil
}

func (p *LocalStorageProvider) DeleteBucket(ctx context.Context, bucketID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if b, ok := p.buckets[bucketID]; ok {
		b.Status = "DELETED"
		_ = os.RemoveAll(filepath.Join(p.baseDir, b.Name))
	}
	return nil
}

func (p *LocalStorageProvider) ListBuckets(ctx context.Context, orgID, projectID string) ([]*domain.Bucket, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var res []*domain.Bucket
	for _, b := range p.buckets {
		if b.Status != "DELETED" {
			res = append(res, b)
		}
	}
	return res, nil
}

func (p *LocalStorageProvider) GetBucket(ctx context.Context, bucketID string) (*domain.Bucket, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if b, ok := p.buckets[bucketID]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("bucket '%s' not found", bucketID)
}

func (p *LocalStorageProvider) PutObject(ctx context.Context, bucketID, key string, data io.Reader, size int64, contentType string) (*domain.Object, error) {
	if err := ValidateObjectKey(key); err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	b, ok := p.buckets[bucketID]
	if !ok {
		return nil, fmt.Errorf("bucket '%s' not found", bucketID)
	}

	objPath := filepath.Join(p.baseDir, b.Name, key)
	_ = os.MkdirAll(filepath.Dir(objPath), 0755)

	f, err := os.Create(objPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	n, err := io.Copy(f, data)
	if err != nil {
		return nil, err
	}

	objID := fmt.Sprintf("obj-%d", time.Now().UnixNano())
	cat := domain.CategoryOther
	if contentType != "" {
		if len(contentType) >= 5 && contentType[:5] == "image" {
			cat = domain.CategoryImages
		} else if len(contentType) >= 5 && contentType[:5] == "video" {
			cat = domain.CategoryVideos
		}
	}

	obj := &domain.Object{
		ID:           objID,
		BucketID:     bucketID,
		Key:          key,
		Size:         n,
		ContentType:  contentType,
		Category:     cat,
		ETag:         fmt.Sprintf("\"etag-%d\"", time.Now().UnixNano()),
		StorageClass: domain.StorageStandard,
		VersionID:    "v1",
		Status:       "ACTIVE",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	p.objects[fmt.Sprintf("%s/%s", bucketID, key)] = obj
	return obj, nil
}

func (p *LocalStorageProvider) GetObject(ctx context.Context, bucketID, key string) (io.ReadCloser, *domain.Object, error) {
	if err := ValidateObjectKey(key); err != nil {
		return nil, nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	b, ok := p.buckets[bucketID]
	if !ok {
		return nil, nil, fmt.Errorf("bucket '%s' not found", bucketID)
	}

	objKey := fmt.Sprintf("%s/%s", bucketID, key)
	obj, ok := p.objects[objKey]
	if !ok {
		return nil, nil, fmt.Errorf("object '%s' not found in bucket '%s'", key, bucketID)
	}

	objPath := filepath.Join(p.baseDir, b.Name, key)
	f, err := os.Open(objPath)
	if err != nil {
		return nil, nil, err
	}

	return f, obj, nil
}

func (p *LocalStorageProvider) DeleteObject(ctx context.Context, bucketID, key string) error {
	if err := ValidateObjectKey(key); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if b, ok := p.buckets[bucketID]; ok {
		_ = os.Remove(filepath.Join(p.baseDir, b.Name, key))
		delete(p.objects, fmt.Sprintf("%s/%s", bucketID, key))
	}
	return nil
}

func (p *LocalStorageProvider) ListObjects(ctx context.Context, bucketID, prefix string) ([]*domain.Object, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var res []*domain.Object
	for _, obj := range p.objects {
		if obj.BucketID == bucketID {
			if prefix == "" || len(obj.Key) >= len(prefix) && obj.Key[:len(prefix)] == prefix {
				res = append(res, obj)
			}
		}
	}
	return res, nil
}

func (p *LocalStorageProvider) GenerateSignedURL(ctx context.Context, bucketID, key, method string, expiresSec int) (*domain.PresignedURL, error) {
	if err := ValidateObjectKey(key); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, ok := p.buckets[bucketID]; !ok {
		return nil, fmt.Errorf("bucket '%s' not found", bucketID)
	}

	if expiresSec <= 0 {
		expiresSec = 900
	}
	expiresAt := time.Now().Add(time.Duration(expiresSec) * time.Second)

	return &domain.PresignedURL{
		URL:       fmt.Sprintf("http://localhost:8080/api/v1/storage/buckets/%s/objects/%s?method=%s&expires=%d", bucketID, key, method, expiresAt.Unix()),
		Method:    method,
		ExpiresAt: expiresAt,
	}, nil
}
