package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
)

type MultipartService struct {
	mu      sync.RWMutex
	uploads map[string]*domain.MultipartUpload
	parts   map[string][]*domain.MultipartPart
}

func NewMultipartService() *MultipartService {
	return &MultipartService{
		uploads: make(map[string]*domain.MultipartUpload),
		parts:   make(map[string][]*domain.MultipartPart),
	}
}

func (s *MultipartService) CreateMultipartUpload(ctx context.Context, bucketID, key string) (*domain.MultipartUpload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	uploadID := fmt.Sprintf("mp-%d", time.Now().UnixNano())
	mp := &domain.MultipartUpload{
		ID:        uploadID,
		UploadID:  uploadID,
		BucketID:  bucketID,
		Key:       key,
		Status:    "IN_PROGRESS",
		CreatedAt: time.Now(),
	}
	s.uploads[uploadID] = mp
	return mp, nil
}

func (s *MultipartService) UploadPart(ctx context.Context, uploadID string, partNumber int, size int64, etag string) (*domain.MultipartPart, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.uploads[uploadID]; !ok {
		return nil, fmt.Errorf("multipart upload '%s' not found", uploadID)
	}

	part := &domain.MultipartPart{
		UploadID:   uploadID,
		PartNumber: partNumber,
		Size:       size,
		ETag:       etag,
	}

	s.parts[uploadID] = append(s.parts[uploadID], part)
	return part, nil
}

func (s *MultipartService) CompleteMultipartUpload(ctx context.Context, uploadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mp, ok := s.uploads[uploadID]; ok {
		mp.Status = "COMPLETED"
	}
	return nil
}
