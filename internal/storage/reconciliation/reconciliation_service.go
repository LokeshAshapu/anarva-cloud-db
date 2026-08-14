package reconciliation

import (
	"context"
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
)

type ReconciliationResult struct {
	BucketID      string   `json:"bucketId"`
	DriftDetected bool     `json:"driftDetected"`
	Mismatches    []string `json:"mismatches"`
	Status        string   `json:"status"`
}

type ReconciliationService struct {
	prov provider.ObjectStorageProvider
}

func NewReconciliationService(prov provider.ObjectStorageProvider) *ReconciliationService {
	return &ReconciliationService{prov: prov}
}

func (s *ReconciliationService) Reconcile(ctx context.Context, desired *domain.Bucket) (*ReconciliationResult, error) {
	actual, err := s.prov.GetBucket(ctx, desired.ID)
	if err != nil {
		return &ReconciliationResult{
			BucketID:      desired.ID,
			DriftDetected: true,
			Mismatches:    []string{fmt.Sprintf("Bucket '%s' missing from provider", desired.ID)},
			Status:        "DRIFTED",
		}, nil
	}

	var mismatches []string
	if desired.StorageClass != actual.StorageClass {
		mismatches = append(mismatches, fmt.Sprintf("Storage class mismatch: desired '%s' vs actual '%s'", desired.StorageClass, actual.StorageClass))
	}

	if len(mismatches) > 0 {
		return &ReconciliationResult{
			BucketID:      desired.ID,
			DriftDetected: true,
			Mismatches:    mismatches,
			Status:        "DRIFTED",
		}, nil
	}

	return &ReconciliationResult{
		BucketID:      desired.ID,
		DriftDetected: false,
		Mismatches:    nil,
		Status:        "IN_SYNC",
	}, nil
}
