package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/repository"

	activityStream "github.com/anarva-cloud/anarva-cloud-db/internal/activity"
)

type StorageService struct {
	repo      *repository.StorageRepository
	prov      provider.ObjectStorageProvider
	signedUrl *SignedURLService
	mpSvc     *MultipartService
	actStream *activityStream.Stream
}

func NewStorageService(
	repo *repository.StorageRepository,
	prov provider.ObjectStorageProvider,
	signedUrl *SignedURLService,
	mpSvc *MultipartService,
	actStream *activityStream.Stream,
) *StorageService {
	return &StorageService{
		repo:      repo,
		prov:      prov,
		signedUrl: signedUrl,
		mpSvc:     mpSvc,
		actStream: actStream,
	}
}

func (s *StorageService) CreateBucket(ctx context.Context, orgID, projectID, name, region string) (*domain.Bucket, error) {
	bucketID := fmt.Sprintf("bkt-%d", time.Now().UnixNano())
	b := &domain.Bucket{
		ID:               bucketID,
		StorageAccountID: "acc-main",
		OrganizationID:   orgID,
		ProjectID:        projectID,
		Name:             name,
		Region:           region,
		StorageClass:     domain.StorageStandard,
		Versioning:       true,
		PublicAccess:     domain.AccessPrivate,
		EncryptionMode:   domain.EncryptionProviderManaged,
		ObjectLock:       false,
		Status:           "ACTIVE",
		RealityLabel:     "LOCAL_STORAGE (REAL_LOCAL)",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	created, err := s.prov.CreateBucket(ctx, b)
	if err != nil {
		return nil, err
	}

	_ = s.repo.SaveBucket(created)

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: orgID,
			ProjectID:      projectID,
			ResourceID:     name,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.EventAction("BUCKET_CREATED"),
			Timestamp:      time.Now(),
		})
	}

	return created, nil
}

func (s *StorageService) ListBuckets(ctx context.Context, orgID, projectID string) ([]*domain.Bucket, error) {
	return s.prov.ListBuckets(ctx, orgID, projectID)
}

func (s *StorageService) DeleteBucket(ctx context.Context, id string) error {
	b, err := s.repo.GetBucket(id)
	if err != nil {
		return err
	}

	if err := s.prov.DeleteBucket(ctx, id); err != nil {
		return err
	}

	b.Status = "DELETED"
	_ = s.repo.SaveBucket(b)

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: b.OrganizationID,
			ProjectID:      b.ProjectID,
			ResourceID:     b.Name,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.EventAction("BUCKET_DELETED"),
			Timestamp:      time.Now(),
		})
	}
	return nil
}

func (s *StorageService) PutObject(ctx context.Context, bucketID, key string, data io.Reader, size int64, contentType string) (*domain.Object, error) {
	obj, err := s.prov.PutObject(ctx, bucketID, key, data, size, contentType)
	if err != nil {
		return nil, err
	}
	_ = s.repo.SaveObject(obj)
	return obj, nil
}

func (s *StorageService) GetObject(ctx context.Context, bucketID, key string) (io.ReadCloser, *domain.Object, error) {
	return s.prov.GetObject(ctx, bucketID, key)
}

func (s *StorageService) ListObjects(ctx context.Context, bucketID, prefix string) ([]*domain.Object, error) {
	return s.prov.ListObjects(ctx, bucketID, prefix)
}

func (s *StorageService) GenerateSignedURL(ctx context.Context, bucketID, key, method string, expiresSec int) (*domain.PresignedURL, error) {
	return s.signedUrl.GenerateSignedURL(ctx, bucketID, key, method, expiresSec)
}

func (s *StorageService) CreateMultipartUpload(ctx context.Context, bucketID, key string) (*domain.MultipartUpload, error) {
	return s.mpSvc.CreateMultipartUpload(ctx, bucketID, key)
}
