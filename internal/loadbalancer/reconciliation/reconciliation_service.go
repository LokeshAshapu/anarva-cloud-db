package reconciliation

import (
	"context"
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/provider"
)

type ReconciliationResult struct {
	LoadBalancerID string   `json:"loadBalancerId"`
	DriftDetected  bool     `json:"driftDetected"`
	Mismatches     []string `json:"mismatches"`
	Status         string   `json:"status"`
}

type ReconciliationService struct {
	provider provider.LoadBalancerProvider
}

func NewReconciliationService(prov provider.LoadBalancerProvider) *ReconciliationService {
	return &ReconciliationService{provider: prov}
}

func (s *ReconciliationService) Reconcile(ctx context.Context, desired *domain.LoadBalancer) (*ReconciliationResult, error) {
	actual, err := s.provider.GetLoadBalancer(ctx, desired.ID)
	if err != nil {
		return &ReconciliationResult{
			LoadBalancerID: desired.ID,
			DriftDetected:  true,
			Mismatches:     []string{fmt.Sprintf("Load Balancer '%s' missing from provider", desired.ID)},
			Status:         "DRIFTED",
		}, nil
	}

	var mismatches []string
	if desired.Scheme != actual.Scheme {
		mismatches = append(mismatches, fmt.Sprintf("Scheme mismatch: desired '%s' vs actual '%s'", desired.Scheme, actual.Scheme))
	}

	if len(mismatches) > 0 {
		return &ReconciliationResult{
			LoadBalancerID: desired.ID,
			DriftDetected:  true,
			Mismatches:     mismatches,
			Status:         "DRIFTED",
		}, nil
	}

	return &ReconciliationResult{
		LoadBalancerID: desired.ID,
		DriftDetected:  false,
		Mismatches:     nil,
		Status:         "IN_SYNC",
	}, nil
}
