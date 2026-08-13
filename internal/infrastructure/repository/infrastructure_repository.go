package repository

import (
	"sync"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
)

type InfrastructureRepository struct {
	mu        sync.RWMutex
	placements map[string]*domain.PlacementPolicy
	haPolicies map[string]*domain.HighAvailabilityPolicy
	failovers  map[string]*domain.FailoverPolicy
	incidents  map[string]*domain.InfrastructureIncident
}

func NewInfrastructureRepository() *InfrastructureRepository {
	return &InfrastructureRepository{
		placements: make(map[string]*domain.PlacementPolicy),
		haPolicies: make(map[string]*domain.HighAvailabilityPolicy),
		failovers:  make(map[string]*domain.FailoverPolicy),
		incidents:  make(map[string]*domain.InfrastructureIncident),
	}
}

func (r *InfrastructureRepository) SavePlacementPolicy(policy *domain.PlacementPolicy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.placements[policy.ID] = policy
	return nil
}

func (r *InfrastructureRepository) ListPlacementPolicies(projectID string) ([]*domain.PlacementPolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.PlacementPolicy
	for _, p := range r.placements {
		if projectID == "" || p.ProjectID == projectID {
			res = append(res, p)
		}
	}
	return res, nil
}

func (r *InfrastructureRepository) SaveIncident(inc *domain.InfrastructureIncident) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.incidents[inc.ID] = inc
	return nil
}

func (r *InfrastructureRepository) ListIncidents() ([]*domain.InfrastructureIncident, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.InfrastructureIncident
	for _, inc := range r.incidents {
		res = append(res, inc)
	}
	return res, nil
}
