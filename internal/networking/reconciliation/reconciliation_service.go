package reconciliation

import (
	"context"
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/provider"
)

type ReconciliationResult struct {
	NetworkID     string   `json:"networkId"`
	DriftDetected bool     `json:"driftDetected"`
	Mismatches    []string `json:"mismatches"`
	Status        string   `json:"status"`
}

type ReconciliationService struct {
	provider provider.NetworkProvider
}

func NewReconciliationService(prov provider.NetworkProvider) *ReconciliationService {
	return &ReconciliationService{provider: prov}
}

func (s *ReconciliationService) Reconcile(ctx context.Context, desired *domain.VirtualNetwork) (*ReconciliationResult, error) {
	actual, err := s.provider.GetNetwork(ctx, desired.ID)
	if err != nil {
		return &ReconciliationResult{
			NetworkID:     desired.ID,
			DriftDetected: true,
			Mismatches:    []string{fmt.Sprintf("Network '%s' missing from provider", desired.ID)},
			Status:        "DRIFTED",
		}, nil
	}

	var mismatches []string
	if desired.CIDR != actual.CIDR {
		mismatches = append(mismatches, fmt.Sprintf("CIDR mismatch: desired '%s' vs actual '%s'", desired.CIDR, actual.CIDR))
	}

	if len(mismatches) > 0 {
		return &ReconciliationResult{
			NetworkID:     desired.ID,
			DriftDetected: true,
			Mismatches:    mismatches,
			Status:        "DRIFTED",
		}, nil
	}

	return &ReconciliationResult{
		NetworkID:     desired.ID,
		DriftDetected: false,
		Mismatches:    nil,
		Status:        "IN_SYNC",
	}, nil
}
