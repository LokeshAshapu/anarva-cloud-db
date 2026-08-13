package simulator

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
)

type OutageSimulator struct{}

func NewOutageSimulator() *OutageSimulator {
	return &OutageSimulator{}
}

func (s *OutageSimulator) SimulateRegionOutage(regionID string) *domain.InfrastructureIncident {
	now := time.Now()
	return &domain.InfrastructureIncident{
		ID:        fmt.Sprintf("inc-sim-%d", now.UnixNano()),
		Severity:  domain.SeverityCritical,
		RegionID:  regionID,
		Type:      "SIMULATED_REGION_OUTAGE",
		Status:    "ACTIVE",
		Summary:   fmt.Sprintf("[SIMULATED] Dev Region '%s' network partition and compute outage simulation active", regionID),
		StartedAt: now,
	}
}
