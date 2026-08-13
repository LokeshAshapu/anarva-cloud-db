package health

import (
	"context"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/provider"
)

type InfrastructureHealthEngine struct {
	prov provider.InfrastructureProvider
}

func NewInfrastructureHealthEngine(prov provider.InfrastructureProvider) *InfrastructureHealthEngine {
	return &InfrastructureHealthEngine{prov: prov}
}

func (e *InfrastructureHealthEngine) EvaluateGlobalHealth(ctx context.Context) (*domain.GlobalHealth, error) {
	regions, err := e.prov.ListRegions(ctx)
	if err != nil {
		return nil, err
	}

	total := len(regions)
	active := 0
	degraded := 0
	outage := 0

	for _, r := range regions {
		switch r.Status {
		case domain.RegionStatusActive:
			active++
		case domain.RegionStatusDegraded:
			degraded++
		case domain.RegionStatusUnavailable:
			outage++
		}
	}

	globalStatus := "HEALTHY"
	if outage > 0 {
		globalStatus = "PARTIAL_OUTAGE"
	} else if degraded > 0 {
		globalStatus = "DEGRADED"
	}

	return &domain.GlobalHealth{
		Status:        globalStatus,
		TotalRegions:  total,
		ActiveRegions: active,
		DegradedCount: degraded,
		OutageCount:   outage,
		Timestamp:     time.Now(),
	}, nil
}
