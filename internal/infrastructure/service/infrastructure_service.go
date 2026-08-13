package service

import (
	"context"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/evacuation"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/failover"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/health"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/placement"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/simulator"

	activityStream "github.com/anarva-cloud/anarva-cloud-db/internal/activity"
)

type InfrastructureService struct {
	repo         *repository.InfrastructureRepository
	prov         provider.InfrastructureProvider
	placementEngine *placement.PlacementEngine
	healthEngine *health.InfrastructureHealthEngine
	failoverEngine *failover.FailoverEngine
	evacSvc      *evacuation.EvacuationService
	simulator    *simulator.OutageSimulator
	actStream    *activityStream.Stream
}

func NewInfrastructureService(
	repo *repository.InfrastructureRepository,
	prov provider.InfrastructureProvider,
	placementEng *placement.PlacementEngine,
	healthEng *health.InfrastructureHealthEngine,
	failoverEng *failover.FailoverEngine,
	evacSvc *evacuation.EvacuationService,
	sim *simulator.OutageSimulator,
	actStream *activityStream.Stream,
) *InfrastructureService {
	return &InfrastructureService{
		repo:            repo,
		prov:            prov,
		placementEngine: placementEng,
		healthEngine:    healthEng,
		failoverEngine: failoverEng,
		evacSvc:         evacSvc,
		simulator:       sim,
		actStream:       actStream,
	}
}

func (s *InfrastructureService) ListRegions(ctx context.Context) ([]*domain.Region, error) {
	return s.prov.ListRegions(ctx)
}

func (s *InfrastructureService) ListZones(ctx context.Context, regionID string) ([]*domain.AvailabilityZone, error) {
	return s.prov.ListZones(ctx, regionID)
}

func (s *InfrastructureService) GetGlobalHealth(ctx context.Context) (*domain.GlobalHealth, error) {
	return s.healthEngine.EvaluateGlobalHealth(ctx)
}

func (s *InfrastructureService) ExecuteFailover(ctx context.Context, policy *domain.FailoverPolicy) (*domain.RecoveryPlan, error) {
	return s.failoverEngine.ExecuteFailover(ctx, policy)
}

func (s *InfrastructureService) SimulateOutage(regionID string) (*domain.InfrastructureIncident, error) {
	inc := s.simulator.SimulateRegionOutage(regionID)
	_ = s.repo.SaveIncident(inc)
	return inc, nil
}

func (s *InfrastructureService) ListIncidents(ctx context.Context) ([]*domain.InfrastructureIncident, error) {
	return s.repo.ListIncidents()
}
