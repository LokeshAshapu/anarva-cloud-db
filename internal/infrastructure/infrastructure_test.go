package infrastructure_test

import (
	"context"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/evacuation"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/failover"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/health"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/placement"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/reconciliation"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/service"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/simulator"
)

func TestInfrastructure_RegionDiscoveryAndPlacement(t *testing.T) {
	ctx := context.Background()
	prov := provider.NewLocalSimulationProvider()
	placementEng := placement.NewPlacementEngine(prov)

	// 1. Valid Placement
	regionID, zoneID, err := placementEng.SelectRegionAndZone(ctx, "ap-hyderabad-1", nil, nil)
	if err != nil || regionID != "ap-hyderabad-1" {
		t.Fatalf("expected region ap-hyderabad-1, got region: %s, zone: %s, err: %v", regionID, zoneID, err)
	}

	// 2. Data Residency Violation
	residency := &domain.DataResidencyPolicy{
		AllowedRegions: []string{"ap-hyderabad-1"},
		Enforcement:    "STRICT",
	}
	_, _, err = placementEng.SelectRegionAndZone(ctx, "us-east-1", nil, residency)
	if err == nil || !testingContains(err.Error(), "DATA RESIDENCY VIOLATION") {
		t.Errorf("expected DATA RESIDENCY VIOLATION error for us-east-1, got: %v", err)
	}
}

func TestInfrastructure_FailoverEngineSplitBrainProtection(t *testing.T) {
	ctx := context.Background()
	failoverEng := failover.NewFailoverEngine()

	policyGen1 := &domain.FailoverPolicy{
		ID:              "pol-01",
		ResourceID:      "db-prod-cluster",
		Primary:         "ap-hyderabad-1",
		Secondary:       "us-east-1",
		HealthThreshold: 3,
		Mode:            domain.FailoverAutomatic,
		GenerationLock:  1,
	}

	// First Failover Attempt (Generation 1 - Allowed)
	plan, err := failoverEng.ExecuteFailover(ctx, policyGen1)
	if err != nil || plan.Status != "COMPLETED" {
		t.Fatalf("expected initial failover to succeed, got err: %v", err)
	}

	// Stale / Concurrent Failover Attempt (Generation 1 - Blocked by Split-Brain Protection)
	_, err = failoverEng.ExecuteFailover(ctx, policyGen1)
	if err == nil || !testingContains(err.Error(), "SPLIT-BRAIN BLOCKED") {
		t.Errorf("expected SPLIT-BRAIN BLOCKED error for stale generation 1, got: %v", err)
	}

	// Valid Next Generation Failover Attempt (Generation 2 - Allowed)
	policyGen2 := *policyGen1
	policyGen2.GenerationLock = 2
	planGen2, err := failoverEng.ExecuteFailover(ctx, &policyGen2)
	if err != nil || planGen2.Status != "COMPLETED" {
		t.Errorf("expected generation 2 failover to succeed, got err: %v", err)
	}
}

func TestInfrastructure_OutageSimulationAndReconciliation(t *testing.T) {
	ctx := context.Background()
	prov := provider.NewLocalSimulationProvider()
	repo := repository.NewInfrastructureRepository()
	sim := simulator.NewOutageSimulator()
	healthEng := health.NewInfrastructureHealthEngine(prov)
	placementEng := placement.NewPlacementEngine(prov)
	failoverEng := failover.NewFailoverEngine()
	evacSvc := evacuation.NewEvacuationService()

	svc := service.NewInfrastructureService(repo, prov, placementEng, healthEng, failoverEng, evacSvc, sim, nil)

	// Simulate Region Outage
	inc, err := svc.SimulateOutage("ap-hyderabad-1")
	if err != nil || inc.Severity != domain.SeverityCritical {
		t.Errorf("expected critical incident from simulation, got: %v", inc)
	}

	// Reconciliation Drift Check
	recSvc := reconciliation.NewReconciliationService(prov)
	res, err := recSvc.Reconcile(ctx, &domain.Region{ID: "non-existent-region"})
	if err != nil || !res.DriftDetected {
		t.Errorf("expected drift detected for missing region, got: %v, err: %v", res, err)
	}
}

func testingContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && stringSearch(s, substr)))
}

func stringSearch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
