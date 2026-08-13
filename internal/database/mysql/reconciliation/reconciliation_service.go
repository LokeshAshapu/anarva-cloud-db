package reconciliation

import (
	"context"
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/provider"
)

type ReconciliationResult struct {
	InstanceID    string   `json:"instanceId"`
	DriftDetected bool     `json:"driftDetected"`
	Mismatches    []string `json:"mismatches"`
	Status        string   `json:"status"`
}

type ReconciliationService struct {
	provider provider.MySQLProvider
}

func NewReconciliationService(prov provider.MySQLProvider) *ReconciliationService {
	return &ReconciliationService{provider: prov}
}

func (s *ReconciliationService) Reconcile(ctx context.Context, desired *domain.MySQLInstance) (*ReconciliationResult, error) {
	actual, err := s.provider.GetInstance(ctx, desired.ID)
	if err != nil {
		return &ReconciliationResult{
			InstanceID:    desired.ID,
			DriftDetected: true,
			Mismatches:    []string{fmt.Sprintf("MySQL instance '%s' missing from provider", desired.ID)},
			Status:        "DRIFTED",
		}, nil
	}

	var mismatches []string
	if desired.Version != actual.Version {
		mismatches = append(mismatches, fmt.Sprintf("Version mismatch: desired '%s' vs actual '%s'", desired.Version, actual.Version))
	}

	if len(mismatches) > 0 {
		return &ReconciliationResult{
			InstanceID:    desired.ID,
			DriftDetected: true,
			Mismatches:    mismatches,
			Status:        "DRIFTED",
		}, nil
	}

	return &ReconciliationResult{
		InstanceID:    desired.ID,
		DriftDetected: false,
		Mismatches:    nil,
		Status:        "IN_SYNC",
	}, nil
}
