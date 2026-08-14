package aws

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storageDomain "github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
)

func TestAWSS3Provider_CreateBucket_Success(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	os.Setenv("AWS_REGION", "us-east-1")
	defer os.Unsetenv("AWS_ENABLED")
	defer os.Unsetenv("AWS_REGION")

	mockS3 := NewMockS3Client(true)
	provider := NewAWSS3Provider(mockS3)

	bucket := &storageDomain.Bucket{
		ID:             "bkt-prod-media",
		Name:           "production-media-assets",
		OrganizationID: "org-alpha-101",
		ProjectID:      "proj-101",
	}

	created, err := provider.CreateBucket(context.Background(), bucket)
	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "AWS_S3", created.Provider)
	assert.Equal(t, "ACTIVE", created.Status)
	assert.Equal(t, "AWS S3 (REAL_CLOUD)", created.RealityLabel)
}

func TestAWSS3Provider_PresignedUploadAndDownload(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockS3 := NewMockS3Client(true)
	provider := NewAWSS3Provider(mockS3)

	bucket := &storageDomain.Bucket{
		ID:             "bkt-presign-test",
		OrganizationID: "org-1",
	}
	_, err := provider.CreateBucket(context.Background(), bucket)
	require.NoError(t, err)

	// Presigned Upload URL
	uploadURL, err := provider.GenerateSignedURL(context.Background(), bucket.ID, "documents/invoice.pdf", "PUT", 900)
	require.NoError(t, err)
	assert.NotNil(t, uploadURL)
	assert.Equal(t, "PUT", uploadURL.Method)
	assert.Contains(t, uploadURL.URL, "https://anarva-s3-bkt-presign-test.s3.amazonaws.com/documents/invoice.pdf")

	// Presigned Download URL
	downloadURL, err := provider.GenerateSignedURL(context.Background(), bucket.ID, "documents/invoice.pdf", "GET", 900)
	require.NoError(t, err)
	assert.NotNil(t, downloadURL)
	assert.Equal(t, "GET", downloadURL.Method)
	assert.Contains(t, downloadURL.URL, "https://anarva-s3-bkt-presign-test.s3.amazonaws.com/documents/invoice.pdf")
}

func TestAWSS3Provider_PathTraversal_Blocked(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockS3 := NewMockS3Client(true)
	provider := NewAWSS3Provider(mockS3)

	bucket := &storageDomain.Bucket{ID: "bkt-traversal-test"}
	_, err := provider.CreateBucket(context.Background(), bucket)
	require.NoError(t, err)

	// Path Traversal Attempt
	_, err = provider.GenerateSignedURL(context.Background(), bucket.ID, "../../../etc/passwd", "GET", 900)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_OBJECT_KEY")
}

func TestAWSS3Provider_NonEmptyBucketDelete_Blocked(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockS3 := NewMockS3Client(true)
	provider := NewAWSS3Provider(mockS3)

	bucket := &storageDomain.Bucket{ID: "bkt-nonempty-test"}
	_, err := provider.CreateBucket(context.Background(), bucket)
	require.NoError(t, err)

	// Add an object to the bucket
	_, err = provider.PutObject(context.Background(), bucket.ID, "file.txt", nil, 1024, "text/plain")
	require.NoError(t, err)

	// Attempt bucket deletion
	err = provider.DeleteBucket(context.Background(), bucket.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "BUCKET_NOT_EMPTY")
}

func TestAWSS3Provider_DeleteObject(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockS3 := NewMockS3Client(true)
	provider := NewAWSS3Provider(mockS3)

	bucket := &storageDomain.Bucket{ID: "bkt-delete-obj-test"}
	_, err := provider.CreateBucket(context.Background(), bucket)
	require.NoError(t, err)

	_, err = provider.PutObject(context.Background(), bucket.ID, "report.pdf", nil, 2048, "application/pdf")
	require.NoError(t, err)

	err = provider.DeleteObject(context.Background(), bucket.ID, "report.pdf")
	require.NoError(t, err)

	// Verify bucket can now be deleted safely
	err = provider.DeleteBucket(context.Background(), bucket.ID)
	require.NoError(t, err)
}

func TestAWSS3Provider_DisabledMode(t *testing.T) {
	os.Setenv("AWS_ENABLED", "false")
	defer os.Unsetenv("AWS_ENABLED")

	mockS3 := NewMockS3Client(true)
	provider := NewAWSS3Provider(mockS3)

	bucket := &storageDomain.Bucket{ID: "bkt-disabled-test"}
	_, err := provider.CreateBucket(context.Background(), bucket)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PROVIDER_DISABLED")
}
