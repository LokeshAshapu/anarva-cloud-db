package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type S3BucketInfo struct {
	Name                    string
	Region                  string
	Status                  string // ACTIVE, DELETING, DELETED
	PublicAccessBlockActive bool
	EncryptionType          string // SSE-S3
	Tags                    map[string]string
	CreatedAt               time.Time
}

type S3ObjectInfo struct {
	BucketName   string
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	VersionID    string
	LastModified time.Time
}

type S3CreateParams struct {
	BucketName string
	Region     string
	Tags       map[string]string
}

// S3Client abstracts AWS S3 operations for real AWS SDK v2 and mock testing
type S3Client interface {
	VerifyConnectivity(ctx context.Context) error
	CreateBucket(ctx context.Context, params S3CreateParams) (*S3BucketInfo, error)
	DeleteBucket(ctx context.Context, bucketName string) error
	PutPublicAccessBlock(ctx context.Context, bucketName string) error
	GeneratePresignedUploadURL(ctx context.Context, bucketName, key, contentType string, expiresSec int) (string, error)
	GeneratePresignedDownloadURL(ctx context.Context, bucketName, key string, expiresSec int) (string, error)
	DeleteObject(ctx context.Context, bucketName, key string) error
	ListObjects(ctx context.Context, bucketName, prefix string) ([]*S3ObjectInfo, error)
}

// MockS3Client provides in-memory simulated S3 responses for unit testing & development
type MockS3Client struct {
	mu          sync.RWMutex
	buckets     map[string]*S3BucketInfo
	objects     map[string]map[string]*S3ObjectInfo
	isConnected bool
}

func NewMockS3Client(isConnected bool) *MockS3Client {
	return &MockS3Client{
		buckets:     make(map[string]*S3BucketInfo),
		objects:     make(map[string]map[string]*S3ObjectInfo),
		isConnected: isConnected,
	}
}

func (m *MockS3Client) VerifyConnectivity(ctx context.Context) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: Invalid AWS credentials for S3 operation")
	}
	return nil
}

func (m *MockS3Client) CreateBucket(ctx context.Context, params S3CreateParams) (*S3BucketInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API CreateBucket unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	b := &S3BucketInfo{
		Name:                    params.BucketName,
		Region:                  params.Region,
		Status:                  "ACTIVE",
		PublicAccessBlockActive: true,
		EncryptionType:          "SSE-S3",
		Tags:                    params.Tags,
		CreatedAt:               time.Now(),
	}
	m.buckets[params.BucketName] = b
	if _, exists := m.objects[params.BucketName]; !exists {
		m.objects[params.BucketName] = make(map[string]*S3ObjectInfo)
	}
	return b, nil
}

func (m *MockS3Client) DeleteBucket(ctx context.Context, bucketName string) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: AWS API DeleteBucket unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	objs, exists := m.objects[bucketName]
	if exists && len(objs) > 0 {
		return fmt.Errorf("BUCKET_NOT_EMPTY: Cannot delete non-empty S3 bucket %s (%d objects remaining)", bucketName, len(objs))
	}

	if b, ok := m.buckets[bucketName]; ok {
		b.Status = "DELETED"
		delete(m.buckets, bucketName)
		delete(m.objects, bucketName)
		return nil
	}
	return fmt.Errorf("NoSuchBucket: The specified bucket %s does not exist", bucketName)
}

func (m *MockS3Client) PutPublicAccessBlock(ctx context.Context, bucketName string) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: AWS API PutPublicAccessBlock unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if b, ok := m.buckets[bucketName]; ok {
		b.PublicAccessBlockActive = true
		return nil
	}
	return fmt.Errorf("NoSuchBucket: Bucket %s not found", bucketName)
}

func (m *MockS3Client) GeneratePresignedUploadURL(ctx context.Context, bucketName, key, contentType string, expiresSec int) (string, error) {
	if !m.isConnected {
		return "", fmt.Errorf("AUTH_FAILED: AWS API presign upload unauthorized")
	}
	if strings.Contains(key, "../") || strings.Contains(key, "..\\") || strings.HasPrefix(key, "/") {
		return "", fmt.Errorf("INVALID_OBJECT_KEY: Path traversal characters are forbidden in key: %s", key)
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Expires=%d", bucketName, key, expiresSec), nil
}

func (m *MockS3Client) GeneratePresignedDownloadURL(ctx context.Context, bucketName, key string, expiresSec int) (string, error) {
	if !m.isConnected {
		return "", fmt.Errorf("AUTH_FAILED: AWS API presign download unauthorized")
	}
	if strings.Contains(key, "../") || strings.Contains(key, "..\\") || strings.HasPrefix(key, "/") {
		return "", fmt.Errorf("INVALID_OBJECT_KEY: Path traversal characters are forbidden in key: %s", key)
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Expires=%d", bucketName, key, expiresSec), nil
}

func (m *MockS3Client) DeleteObject(ctx context.Context, bucketName, key string) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: AWS API DeleteObject unauthorized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if objs, ok := m.objects[bucketName]; ok {
		delete(objs, key)
		return nil
	}
	return fmt.Errorf("NoSuchBucket: Bucket %s not found", bucketName)
}

func (m *MockS3Client) ListObjects(ctx context.Context, bucketName, prefix string) ([]*S3ObjectInfo, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API ListObjects unauthorized")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*S3ObjectInfo
	if objs, ok := m.objects[bucketName]; ok {
		for k, obj := range objs {
			if prefix == "" || strings.HasPrefix(k, prefix) {
				res = append(res, obj)
			}
		}
	}
	return res, nil
}
