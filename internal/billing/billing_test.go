package billing_test

import (
	"context"
	"sync"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/billing/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/billing/usecase"
)

func TestBillingUseCase_QuotaRaceConditionsAndCostEstimator(t *testing.T) {
	uc := usecase.NewBillingUseCase()
	ctx := context.Background()

	// Test Concurrent Quota Reservations (Race Condition Safety)
	var wg sync.WaitGroup
	errCount := 0
	var mu sync.Mutex

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := uc.ReserveQuota(ctx, "org-default", "proj-default", "COMPUTE", "compute.acu", 2.0)
			if err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Limit is 32.0 ACU. Starting current usage is 4.0 ACU.
	// 20 requests of 2.0 ACU = 40 ACU total requested.
	// Maximum allowed = 28 ACU = 14 requests succeed, 6 requests fail due to quota limit!
	if errCount == 0 {
		t.Errorf("Expected quota limit errors during concurrent requests, got 0 failures")
	}

	// Test Usage Recording
	err := uc.RecordUsage(ctx, &domain.UsageRecord{
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		ResourceID:     "ace-worker-node-01",
		ResourceType:   "COMPUTE",
		Provider:       "LOCAL_DOCKER",
		Metric:         "compute.runtime",
		Quantity:       24.0,
		Unit:           "ACU-hour",
		Source:         domain.SourceLocalProvider,
		Quality:        domain.QualitySimulated,
		RealityLabel:   "LOCAL_DEVELOPMENT_USAGE (NON_BILLABLE)",
	})
	if err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	// Test Cost Estimator
	est, err := uc.CalculateCostEstimate(ctx, "COMPUTE", "LOCAL_DOCKER", 2.0, 720.0)
	if err != nil {
		t.Fatalf("CalculateCostEstimate failed: %v", err)
	}
	if est.EstimatedCost <= 0 {
		t.Errorf("Expected positive cost estimate, got %.2f", est.EstimatedCost)
	}
	if est.RealityLabel != "NOT_BILLABLE (ESTIMATE)" {
		t.Errorf("Expected NOT_BILLABLE reality label, got %s", est.RealityLabel)
	}
}
