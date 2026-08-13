package reconciliation

import (
	"context"
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/provider"
)

type ReconciliationResult struct {
	RegionID      string   `json:"regionId"`
	DriftDetected bool     `json:"driftDetected"`
	Mismatches    []string `json:"mismatches"`
	Status        string   `json:"status"`
}

type ReconciliationService struct {
	prov provider.InfrastructureProvider
}

func NewReconciliationService(prov provider.InfrastructureProvider) *ReconciliationService {
	return &ReconciliationService{prov: prov}
}

func (s *ReconciliationService) Reconcile(ctx context.Context, desired *domain.Region) (*ReconciliationResult, error) {
	regions, err := s.prov.ListRegions(ctx)
	if err != nil {
		return nil, err
	}

	var actual *domain.Region
	for _, r := range regions {
		if r.ID == desired.ID {
			actual = r
			break
		}
	}

	if actual == nil {
		return &ReconciliationResult{
			RegionID:      desired.ID,
			DriftDetected: true,
			Mismatches:    []string{fmt.Sprintf("Region '%s' not discovered by provider", desired.ID)},
			Status:        "DRIFTED",
		}, nil
	}

	return &ReconciliationResult{
		RegionID:      desired.ID,
		DriftDetected: false,
		Mismatches:    nil,
		Status:        "IN_SYNC",
	}, nil
}
