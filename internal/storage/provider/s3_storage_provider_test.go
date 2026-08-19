package provider_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
	storageProvider "github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
)

// MockS3Client implements storageProvider.S3APIClient for in-memory testing.
type MockS3Client struct {
	buckets map[string]*s3Types.Bucket
	objects map[string][]byte
}

func NewMockS3Client() *MockS3Client {
	return &MockS3Client{
		buckets: make(map[string]*s3Types.Bucket),
		objects: make(map[string][]byte),
	}
}

func (m *MockS3Client) CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	name := aws.ToString(params.Bucket)
	m.buckets[name] = &s3Types.Bucket{
		Name:         aws.String(name),
		CreationDate: aws.Time(time.Now()),
	}
	return &s3.CreateBucketOutput{}, nil
}

func (m *MockS3Client) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	name := aws.ToString(params.Bucket)
	delete(m.buckets, name)
	return &s3.DeleteBucketOutput{}, nil
}

func (m *MockS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	var list []s3Types.Bucket
	for _, b := range m.buckets {
		list = append(list, *b)
	}
	return &s3.ListBucketsOutput{Buckets: list}, nil
}

func (m *MockS3Client) HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	name := aws.ToString(params.Bucket)
	if _, ok := m.buckets[name]; !ok {
		return nil, fmt.Errorf("NotFound: bucket does not exist")
	}
	return &s3.HeadBucketOutput{}, nil
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	bkt := aws.ToString(params.Bucket)
	key := aws.ToString(params.Key)
	full := fmt.Sprintf("%s/%s", bkt, key)

	if params.Body != nil {
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, params.Body)
		if err != nil {
			return nil, err
		}
		m.objects[full] = buf.Bytes()
	} else {
		m.objects[full] = []byte{}
	}

	return &s3.PutObjectOutput{ETag: aws.String("\"test-etag\"")}, nil
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	bkt := aws.ToString(params.Bucket)
	key := aws.ToString(params.Key)
	full := fmt.Sprintf("%s/%s", bkt, key)

	data, ok := m.objects[full]
	if !ok {
		return nil, fmt.Errorf("NoSuchKey: key does not exist")
	}

	return &s3.GetObjectOutput{
		Body:          io.NopCloser(bytes.NewReader(data)),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   aws.String("application/octet-stream"),
		ETag:          aws.String("\"test-etag\""),
	}, nil
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	bkt := aws.ToString(params.Bucket)
	key := aws.ToString(params.Key)
	full := fmt.Sprintf("%s/%s", bkt, key)
	delete(m.objects, full)
	return &s3.DeleteObjectOutput{}, nil
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	bkt := aws.ToString(params.Bucket)
	prefix := aws.ToString(params.Prefix)
	var items []s3Types.Object

	for full, data := range m.objects {
		if strings.HasPrefix(full, bkt+"/") {
			key := strings.TrimPrefix(full, bkt+"/")
			if prefix == "" || strings.HasPrefix(key, prefix) {
				items = append(items, s3Types.Object{
					Key:          aws.String(key),
					Size:         aws.Int64(int64(len(data))),
					LastModified: aws.Time(time.Now()),
				})
			}
		}
	}

	return &s3.ListObjectsV2Output{Contents: items}, nil
}

type MockPresigner struct{}

func (p *MockPresigner) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	url := fmt.Sprintf("https://s3.amazonaws.com/%s/%s?presigned=true", aws.ToString(params.Bucket), aws.ToString(params.Key))
	return &v4.PresignedHTTPRequest{URL: url, Method: "GET"}, nil
}

func (p *MockPresigner) PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	url := fmt.Sprintf("https://s3.amazonaws.com/%s/%s?presigned=true", aws.ToString(params.Bucket), aws.ToString(params.Key))
	return &v4.PresignedHTTPRequest{URL: url, Method: "PUT"}, nil
}

func TestPhase62_StorageProviderSuite(t *testing.T) {
	t.Run("1. TestLocalProviderDevelopmentMode", func(t *testing.T) {
		cfg := config.StorageConfig{Driver: "local", LocalPath: t.TempDir()}
		prov, err := storageProvider.NewStorageProvider(cfg, "development")
		require.NoError(t, err)
		assert.Equal(t, "LOCAL_STORAGE", prov.GetProviderType())
	})

	t.Run("2. TestS3ProviderConfiguration", func(t *testing.T) {
		cfg := config.StorageConfig{
			Driver:      "s3",
			S3Bucket:    "test-bucket",
			S3Region:    "us-west-2",
			S3AccessKey: "AKIAIOSFODNN7EXAMPLE",
			S3SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			S3Endpoint:  "https://12345.r2.cloudflarestorage.com",
		}
		prov, err := storageProvider.NewStorageProvider(cfg, "production")
		require.NoError(t, err)
		assert.Equal(t, "S3", prov.GetProviderType())
	})

	t.Run("3. TestMissingS3Configuration", func(t *testing.T) {
		cfg := config.StorageConfig{
			Driver:   "s3",
			S3Bucket: "",
		}
		_, err := storageProvider.NewStorageProvider(cfg, "development")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "STORAGE_CONFIG_MISSING")
	})

	t.Run("4. TestProductionFailsWithoutS3", func(t *testing.T) {
		cfg := config.StorageConfig{
			Driver: "local",
		}
		_, err := storageProvider.NewStorageProvider(cfg, "production")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CONFIG_VALIDATION_FAILURE")
	})

	t.Run("5. TestProviderFactory", func(t *testing.T) {
		devCfg := config.StorageConfig{Driver: "local", LocalPath: t.TempDir()}
		devProv, err := storageProvider.NewStorageProvider(devCfg, "development")
		require.NoError(t, err)
		assert.Equal(t, "LOCAL_STORAGE", devProv.GetProviderType())

		prodCfg := config.StorageConfig{
			Driver:      "s3",
			S3Bucket:    "prod-bkt",
			S3AccessKey: "key",
			S3SecretKey: "sec",
		}
		prodProv, err := storageProvider.NewStorageProvider(prodCfg, "production")
		require.NoError(t, err)
		assert.Equal(t, "S3", prodProv.GetProviderType())
	})

	t.Run("6. TestTenantObjectKeyIsolation & 7. PathTraversalProtection", func(t *testing.T) {
		mockClient := NewMockS3Client()
		mockPresigner := &MockPresigner{}
		prov := storageProvider.NewS3StorageProviderWithClient(mockClient, mockPresigner, "anarva-test-bucket", "us-east-1", "")

		badKeys := []string{
			"../secret.txt",
			"/etc/passwd",
			"..\\windows\\system32",
			"%2e%2e%2fetc/passwd",
			"bucket/../../root",
		}

		ctx := context.Background()
		for _, bk := range badKeys {
			_, err := prov.PutObject(ctx, "bkt-1", bk, strings.NewReader("data"), 4, "text/plain")
			assert.Error(t, err, "Expected error for key: %s", bk)
			assert.Contains(t, err.Error(), "STORAGE_SECURITY_RISK")

			_, _, err = prov.GetObject(ctx, "bkt-1", bk)
			assert.Error(t, err)

			err = prov.DeleteObject(ctx, "bkt-1", bk)
			assert.Error(t, err)
		}
	})

	t.Run("8. TestObjectNotFound & 9. TestUploadDownloadStreaming", func(t *testing.T) {
		mockClient := NewMockS3Client()
		mockPresigner := &MockPresigner{}
		prov := storageProvider.NewS3StorageProviderWithClient(mockClient, mockPresigner, "anarva-bkt", "us-east-1", "")

		ctx := context.Background()

		// 1. Create Bucket
		bkt, err := prov.CreateBucket(ctx, &domain.Bucket{Name: "anarva-bkt"})
		require.NoError(t, err)
		assert.Equal(t, "S3", bkt.Provider)

		// 2. Upload Object (Streaming)
		payload := "Hello, ANARVA Cloud S3 Storage!"
		key := "tenants/t-100/projects/p-200/objects/file.txt"
		obj, err := prov.PutObject(ctx, "anarva-bkt", key, strings.NewReader(payload), int64(len(payload)), "text/plain")
		require.NoError(t, err)
		assert.Equal(t, key, obj.Key)

		// 3. Download Object (Streaming)
		rc, getObj, err := prov.GetObject(ctx, "anarva-bkt", key)
		require.NoError(t, err)
		defer rc.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, rc)
		require.NoError(t, err)

		assert.Equal(t, payload, buf.String())
		assert.Equal(t, int64(len(payload)), getObj.Size)

		// 4. Presigned URL Generation
		purl, err := prov.GenerateSignedURL(ctx, "anarva-bkt", key, "GET", 900)
		require.NoError(t, err)
		assert.Contains(t, purl.URL, "presigned=true")

		// 5. Delete Object
		err = prov.DeleteObject(ctx, "anarva-bkt", key)
		require.NoError(t, err)

		// 6. Object Not Found after deletion
		_, _, err = prov.GetObject(ctx, "anarva-bkt", key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "STORAGE_OBJECT_NOT_FOUND")
	})
}
