package provider

import (
	"context"
	"sync"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
)

type InfrastructureProvider interface {
	GetProviderType() string
	ListRegions(ctx context.Context) ([]*domain.Region, error)
	ListZones(ctx context.Context, regionID string) ([]*domain.AvailabilityZone, error)
	GetCapacity(ctx context.Context, regionID string) (*domain.RegionCapacity, error)
	GetHealth(ctx context.Context, regionID string) (*domain.RegionHealth, error)
}

type LocalSimulationProvider struct {
	mu      sync.RWMutex
	regions map[string]*domain.Region
	zones   map[string][]*domain.AvailabilityZone
}

func NewLocalSimulationProvider() *LocalSimulationProvider {
	p := &LocalSimulationProvider{
		regions: make(map[string]*domain.Region),
		zones:   make(map[string][]*domain.AvailabilityZone),
	}

	// Register Local Simulation Regions & Zones
	apHyderabad := &domain.Region{
		ID:                 "ap-hyderabad-1",
		Name:               "Asia Pacific (Hyderabad)",
		Code:               "ap-hyderabad-1",
		Provider:           "LOCAL_SIMULATION",
		Status:             domain.RegionStatusActive,
		LatitudeReference:  17.3850,
		LongitudeReference: 78.4867,
		CountryCode:        "IN",
		CapacityStatus:     "OPTIMAL",
		RealityLabel:       "LOCAL_SIMULATION (LIMITED_CAPABILITIES)",
	}

	usEast := &domain.Region{
		ID:                 "us-east-1",
		Name:               "US East (N. Virginia)",
		Code:               "us-east-1",
		Provider:           "LOCAL_SIMULATION",
		Status:             domain.RegionStatusActive,
		LatitudeReference:  38.9072,
		LongitudeReference: -77.0369,
		CountryCode:        "US",
		CapacityStatus:     "OPTIMAL",
		RealityLabel:       "LOCAL_SIMULATION (LIMITED_CAPABILITIES)",
	}

	p.regions[apHyderabad.ID] = apHyderabad
	p.regions[usEast.ID] = usEast

	p.zones[apHyderabad.ID] = []*domain.AvailabilityZone{
		{ID: "ap-hyderabad-1a", RegionID: apHyderabad.ID, Name: "ap-hyderabad-1a", Status: domain.ZoneStatusActive, CapacityStatus: "AVAILABLE"},
		{ID: "ap-hyderabad-1b", RegionID: apHyderabad.ID, Name: "ap-hyderabad-1b", Status: domain.ZoneStatusActive, CapacityStatus: "AVAILABLE"},
	}

	p.zones[usEast.ID] = []*domain.AvailabilityZone{
		{ID: "us-east-1a", RegionID: usEast.ID, Name: "us-east-1a", Status: domain.ZoneStatusActive, CapacityStatus: "AVAILABLE"},
		{ID: "us-east-1b", RegionID: usEast.ID, Name: "us-east-1b", Status: domain.ZoneStatusActive, CapacityStatus: "AVAILABLE"},
	}

	return p
}

func (p *LocalSimulationProvider) GetProviderType() string {
	return "LOCAL_SIMULATION"
}

func (p *LocalSimulationProvider) ListRegions(ctx context.Context) ([]*domain.Region, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var res []*domain.Region
	for _, r := range p.regions {
		res = append(res, r)
	}
	return res, nil
}

func (p *LocalSimulationProvider) ListZones(ctx context.Context, regionID string) ([]*domain.AvailabilityZone, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if z, ok := p.zones[regionID]; ok {
		return z, nil
	}
	return nil, nil
}

func (p *LocalSimulationProvider) GetCapacity(ctx context.Context, regionID string) (*domain.RegionCapacity, error) {
	return &domain.RegionCapacity{
		RegionID:        regionID,
		TotalACU:        1000,
		AvailableACU:    850,
		AllocatedACU:    150,
		CPUCapacity:     2000,
		MemoryCapacity:  1024 * 1024 * 8,
		StorageCapacity: 1024 * 1024 * 100,
		Status:          "OPTIMAL",
	}, nil
}

func (p *LocalSimulationProvider) GetHealth(ctx context.Context, regionID string) (*domain.RegionHealth, error) {
	return &domain.RegionHealth{
		RegionID:       regionID,
		Status:         "HEALTHY",
		LatencyMs:      12.4,
		ResourceHealth: "HEALTHY",
		NetworkHealth:  "HEALTHY",
		DatabaseHealth: "HEALTHY",
		Quality:        "ACTUAL (LOCAL_SIMULATION)",
	}, nil
}
