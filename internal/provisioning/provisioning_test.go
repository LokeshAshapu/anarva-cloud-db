package provisioning_test

import (
	"context"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/usecase"
)

func TestProvisioningEngine_PlannerAndExecution(t *testing.T) {
	registry := provider.NewProviderRegistry()
	dockerProv := provider.NewDockerInfrastructureProvider()
	registry.RegisterProvider(dockerProv)

	uc := usecase.NewProvisioningUseCase(nil, nil, nil, registry)
	ctx := context.Background()

	req := &domain.ProvisioningRequest{
		OrganizationID: "org-test",
		ProjectID:      "proj-test",
		ResourceType:   domain.TypeCompute,
		ResourceID:     "unit-test-node-01",
		Provider:       "LOCAL_DOCKER",
		RegionID:       "us-east-1",
		RequestedBy:    "testuser@anarva.io",
		IdempotencyKey: "idem-key-101",
	}

	planReq, err := uc.CreatePlan(ctx, req)
	if err != nil {
		t.Fatalf("CreatePlan failed: %v", err)
	}

	if planReq.Status != domain.StatusPlanning {
		t.Errorf("Expected status PLANNING, got %s", planReq.Status)
	}

	if planReq.ExecutionPlan == nil || len(planReq.ExecutionPlan.Steps) == 0 {
		t.Fatalf("Expected non-empty execution plan")
	}

	// Test Idempotency
	idemReq, err := uc.CreatePlan(ctx, req)
	if err != nil {
		t.Fatalf("Idempotent CreatePlan failed: %v", err)
	}
	if idemReq.ID != planReq.ID {
		t.Errorf("Idempotency failed: expected request ID %s, got %s", planReq.ID, idemReq.ID)
	}

	// Apply Request
	appliedReq, err := uc.ApplyRequest(ctx, planReq.ID)
	if err != nil {
		t.Fatalf("ApplyRequest failed: %v", err)
	}

	if appliedReq.Status != domain.StatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s", appliedReq.Status)
	}
}

func TestProvisioningEngine_DriftAndReconciliation(t *testing.T) {
	registry := provider.NewProviderRegistry()
	dockerProv := provider.NewDockerInfrastructureProvider()
	registry.RegisterProvider(dockerProv)

	uc := usecase.NewProvisioningUseCase(nil, nil, nil, registry)
	ctx := context.Background()

	drift, err := uc.GetDrift("unit-test-node-01")
	if err != nil {
		t.Fatalf("GetDrift failed: %v", err)
	}

	if drift.Status != domain.DriftInSync {
		t.Errorf("Expected DriftInSync, got %s", drift.Status)
	}

	reconciled, err := uc.ReconcileResource(ctx, "unit-test-node-01")
	if err != nil {
		t.Fatalf("ReconcileResource failed: %v", err)
	}

	if reconciled.Status != domain.DriftInSync {
		t.Errorf("Expected DriftInSync after reconciliation, got %s", reconciled.Status)
	}
}
