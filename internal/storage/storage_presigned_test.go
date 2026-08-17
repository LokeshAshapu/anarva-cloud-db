package storage_test

import (
	"context"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storageDomain "github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
	storageProvider "github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
	storageService "github.com/anarva-cloud/anarva-cloud-db/internal/storage/service"
)

func TestWorkflow3_StorageObjectPresignedURLGenerationAndVerification(t *testing.T) {
	ctx := context.Background()
	prov := storageProvider.NewLocalStorageProvider("")
	signedSvc := storageService.NewSignedURLService(prov)

	bucketID := "bkt-test-wf3"

	// 1. Create Bucket in Local Provider
	b := &storageDomain.Bucket{
		ID:        bucketID,
		Name:      "test-presigned-bucket",
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
	_, err := prov.CreateBucket(ctx, b)
	require.NoError(t, err)

	// 2. Generate Presigned URL for valid object key
	objectKey := "documents/2026/report.pdf"
	pURL, errGen := signedSvc.GenerateSignedURL(ctx, bucketID, objectKey, "GET", 1800)
	require.NoError(t, errGen)
	assert.NotEmpty(t, pURL.URL)
	assert.Equal(t, objectKey, pURL.Key)

	// Parse URL params
	parsedURL, errParse := url.Parse(pURL.URL)
	require.NoError(t, errParse)

	sig := parsedURL.Query().Get("signature")
	expiresStr := parsedURL.Query().Get("expires")
	expiresUnix, _ := strconv.ParseInt(expiresStr, 10, 64)

	// 3. Verify valid signature
	assert.NoError(t, signedSvc.ValidateSignedURL(bucketID, objectKey, "GET", sig, expiresUnix))

	// 4. Path Traversal Key Validation Rejection
	_, errTraversal := signedSvc.GenerateSignedURL(ctx, bucketID, "../../../etc/passwd", "GET", 1800)
	assert.Error(t, errTraversal)
	assert.Contains(t, errTraversal.Error(), "STORAGE_SECURITY_RISK")

	// 5. Tampered signature rejection
	assert.Error(t, signedSvc.ValidateSignedURL(bucketID, objectKey, "GET", "invalid_signature", expiresUnix))
}
