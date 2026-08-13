package health

import (
	"context"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
)

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) EvaluateTargetHealth(ctx context.Context, target *domain.BackendTarget, check *domain.LoadBalancerHealthCheck) domain.TargetStatus {
	if target.AddressReference == "" || target.Port == 0 {
		return domain.TargetUnhealthy
	}
	return domain.TargetHealthy
}

func (s *HealthService) AggregateApplicationHealth(targets []domain.BackendTarget) string {
	if len(targets) == 0 {
		return "UNKNOWN"
	}

	healthyCount := 0
	for _, t := range targets {
		if t.Status == domain.TargetHealthy {
			healthyCount++
		}
	}

	if healthyCount == len(targets) {
		return "HEALTHY"
	} else if healthyCount > 0 {
		return "DEGRADED"
	}
	return "UNHEALTHY"
}
