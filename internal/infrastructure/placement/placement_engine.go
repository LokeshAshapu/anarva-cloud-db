package placement

import (
	"context"
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/provider"
)

type PlacementEngine struct {
	prov provider.InfrastructureProvider
}

func NewPlacementEngine(prov provider.InfrastructureProvider) *PlacementEngine {
	return &PlacementEngine{prov: prov}
}

func (e *PlacementEngine) SelectRegionAndZone(ctx context.Context, preferredRegion string, policy *domain.PlacementPolicy, residency *domain.DataResidencyPolicy) (string, string, error) {
	// Data Residency Enforcement
	if residency != nil && len(residency.AllowedRegions) > 0 {
		allowed := false
		for _, r := range residency.AllowedRegions {
			if r == preferredRegion {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", "", fmt.Errorf("DATA RESIDENCY VIOLATION: Preferred region '%s' is prohibited by policy. Allowed: %v", preferredRegion, residency.AllowedRegions)
		}
	}

	regions, err := e.prov.ListRegions(ctx)
	if err != nil {
		return "", "", err
	}

	var targetRegion *domain.Region
	for _, r := range regions {
		if preferredRegion == "" || r.ID == preferredRegion {
			if r.Status == domain.RegionStatusActive {
				targetRegion = r
				break
			}
		}
	}

	if targetRegion == nil {
		return "", "", fmt.Errorf("no available active region matching criteria '%s'", preferredRegion)
	}

	zones, err := e.prov.ListZones(ctx, targetRegion.ID)
	if err != nil || len(zones) == 0 {
		return targetRegion.ID, fmt.Sprintf("%sa", targetRegion.ID), nil
	}

	return targetRegion.ID, zones[0].ID, nil
}
