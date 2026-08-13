package service

import (
	"context"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/drift"
	importengine "github.com/anarva-cloud/anarva-cloud-db/internal/providers/import"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/security"

	activityStream "github.com/anarva-cloud/anarva-cloud-db/internal/activity"
)

type ProviderService struct {
	reg          *registry.ProviderRegistry
	mappingRepo  *mapping.MappingRepository
	driftEngine  *drift.DriftEngine
	importEngine *importengine.ImportEngine
	ssrfEng      *security.SSRFProtectionEngine
	actStream    *activityStream.Stream
}

func NewProviderService(
	reg *registry.ProviderRegistry,
	mappingRepo *mapping.MappingRepository,
	driftEngine *drift.DriftEngine,
	importEngine *importengine.ImportEngine,
	ssrfEng *security.SSRFProtectionEngine,
	actStream *activityStream.Stream,
) *ProviderService {
	return &ProviderService{
		reg:          reg,
		mappingRepo:  mappingRepo,
		driftEngine:  driftEngine,
		importEngine: importEngine,
		ssrfEng:      ssrfEng,
		actStream:    actStream,
	}
}

func (s *ProviderService) ListProviders(ctx context.Context) ([]*registry.ProviderInfo, error) {
	return s.reg.ListProviders(ctx)
}

func (s *ProviderService) GetProvider(ctx context.Context, id string) (*registry.ProviderInfo, error) {
	return s.reg.GetProvider(ctx, id)
}

func (s *ProviderService) VerifyProvider(ctx context.Context, id, credRef string) (*registry.ProviderInfo, error) {
	return s.reg.VerifyProvider(ctx, id, credRef)
}

func (s *ProviderService) ImportResource(ctx context.Context, provider, providerResourceID, resourceType, region string) (*mapping.ProviderResourceMapping, error) {
	return s.importEngine.ImportResource(ctx, provider, providerResourceID, resourceType, region)
}

func (s *ProviderService) AdoptResource(ctx context.Context, anarvaResourceID string) (*mapping.ProviderResourceMapping, error) {
	return s.importEngine.AdoptResource(ctx, anarvaResourceID)
}

func (s *ProviderService) ReleaseResource(ctx context.Context, anarvaResourceID string) error {
	return s.importEngine.ReleaseResource(ctx, anarvaResourceID)
}

func (s *ProviderService) DetectDrift(ctx context.Context, anarvaResourceID, desiredState, observedState string) (*drift.DriftRecord, error) {
	return s.driftEngine.DetectDrift(ctx, anarvaResourceID, desiredState, observedState)
}
