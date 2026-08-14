package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/reconciliation"
	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/service"
)

func TestStorage_BucketAndObjectLifecycle(t *testing.T) {
	ctx := context.Background()
	tempDir := filepath.Join(os.TempDir(), "anarva-storage-test")
	defer os.RemoveAll(tempDir)

	prov := provider.NewLocalStorageProvider(tempDir)
	repo := repository.NewStorageRepository()
	signedSvc := service.NewSignedURLService(prov)
	mpSvc := service.NewMultipartService()

	svc := service.NewStorageService(repo, prov, signedSvc, mpSvc, nil)

	// 1. Create Bucket
	b, err := svc.CreateBucket(ctx, "org-test", "proj-test", "media-assets", "ap-hyderabad-1")
	if err != nil || b.Status != "ACTIVE" {
		t.Fatalf("failed to create bucket: %v", err)
	}

	// 2. Put Object (Upload)
	content := []byte("Hello Anarva Object Storage Platform!")
	obj, err := svc.PutObject(ctx, b.ID, "docs/hello.txt", bytes.NewReader(content), int64(len(content)), "text/plain")
	if err != nil || obj.Size != int64(len(content)) {
		t.Fatalf("failed to put object: %v", err)
	}

	// 3. Get Object (Download)
	rc, readObj, err := svc.GetObject(ctx, b.ID, "docs/hello.txt")
	if err != nil || readObj.Key != "docs/hello.txt" {
		t.Fatalf("failed to get object: %v", err)
	}

	readBytes, err := io.ReadAll(rc)
	rc.Close()
	if err != nil || string(readBytes) != string(content) {
		t.Errorf("content mismatch: expected '%s', got '%s'", string(content), string(readBytes))
	}

	// 4. Presigned Signed URL Generation
	pURL, err := svc.GenerateSignedURL(ctx, b.ID, "docs/hello.txt", "GET", 1800)
	if err != nil || pURL.Bucket != b.ID {
		t.Errorf("failed to generate presigned URL: %v", err)
	}

	// 5. Delete Bucket
	if err := svc.DeleteBucket(ctx, b.ID); err != nil {
		t.Errorf("failed to delete bucket: %v", err)
	}
}

func TestStorage_MultipartAndReconciliation(t *testing.T) {
	ctx := context.Background()
	mpSvc := service.NewMultipartService()

	// 1. Create Multipart Upload
	mp, err := mpSvc.CreateMultipartUpload(ctx, "bkt-01", "large-video.mp4")
	if err != nil || mp.Status != "IN_PROGRESS" {
		t.Fatalf("failed to start multipart upload: %v", err)
	}

	// 2. Upload Part
	part, err := mpSvc.UploadPart(ctx, mp.UploadID, 1, 5242880, "\"part-etag-1\"")
	if err != nil || part.PartNumber != 1 {
		t.Errorf("failed to record multipart part: %v", err)
	}

	// 3. Reconciliation Check
	tempDir := filepath.Join(os.TempDir(), "anarva-storage-rec-test")
	defer os.RemoveAll(tempDir)
	prov := provider.NewLocalStorageProvider(tempDir)
	recSvc := reconciliation.NewReconciliationService(prov)

	res, err := recSvc.Reconcile(ctx, &domain.Bucket{ID: "ghost-bucket"})
	if err != nil || !res.DriftDetected {
		t.Errorf("expected drift detected for missing bucket, got: %v", res)
	}
}
