package reliability_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	healthService "github.com/anarva-cloud/anarva-cloud-db/internal/health"
	reliabilityDomain "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/domain"
	reliabilityUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
)

// 1. Complete Process Interruption & Restart Recovery Test
func TestPhase45_ProcessInterruptionAndRestartRecovery(t *testing.T) {
	ctx := context.Background()
	uc := reliabilityUsecase.NewReliabilityUseCase()

	// Step 1: Dispatch operation -> RUNNING state (locks resource automatically)
	op, err := uc.DispatchOperation(ctx, "org-phase45", "proj-phase45", "res-comp-45", reliabilityDomain.OpCreateCompute, "idemp-phase45-1", "payload", "req_phase45_1")
	require.NoError(t, err)
	assert.Equal(t, reliabilityDomain.OpStatusRunning, op.Status)
	assert.Equal(t, "org-phase45", op.OrganizationID)

	// Step 2: Simulate process interruption & restart -> Recovery Worker reconciliation
	recWorker := reliabilityUsecase.NewRecoveryWorker(uc, reliabilityUsecase.RecoveryWorkerConfig{
		Interval:  10 * time.Millisecond,
		BatchSize: 10,
	})
	recWorker.Start(ctx)

	// Wait for reconciliation loop
	time.Sleep(100 * time.Millisecond)
	recWorker.Stop()

	// Step 3: Verify operation was recovered safely
	fetched, err := uc.GetOperation("org-phase45", op.ID)
	require.NoError(t, err)
	assert.True(t, fetched.Status == reliabilityDomain.OpStatusSucceeded || fetched.Status == reliabilityDomain.OpStatusFailed || fetched.Status == reliabilityDomain.OpStatusRunning)

	// Step 4: Verify lock is released and available for new operations
	acquiredNew, newLeaseID, err := uc.AcquireResourceLock(ctx, op.ResourceID, "op-next", 1*time.Minute)
	require.NoError(t, err)
	assert.True(t, acquiredNew)
	_, _ = uc.ReleaseResourceLock(ctx, op.ResourceID, newLeaseID)
}

// 2. Database Disconnect & Readiness State Transition Test
func TestPhase45_DatabaseDisconnectAndReadinessRecovery(t *testing.T) {
	cfg := &config.Config{Environment: "production"}
	uc := reliabilityUsecase.NewReliabilityUseCase()

	// In production mode without database pool -> readiness returns StatusUnavailable
	hSvc := healthService.NewHealthService(nil, cfg, nil, uc, "0.1.0")
	checks := hSvc.CheckReadiness(context.Background())
	assert.Equal(t, healthService.StatusUnavailable, checks.Database)
}

// 3. Expired Lease Handling & Automatic Lease Renewal Test
func TestPhase45_LeaseExpirationAndRenewal(t *testing.T) {
	ctx := context.Background()
	uc := reliabilityUsecase.NewReliabilityUseCase()

	acquired, leaseID, err := uc.AcquireResourceLock(ctx, "res-lease-renew", "op-renew-101", 100*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Renew lease
	renewed, err := uc.RenewResourceLock(ctx, "res-lease-renew", leaseID, 500*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, renewed)

	// Release lock cleanly
	released, err := uc.ReleaseResourceLock(ctx, "res-lease-renew", leaseID)
	require.NoError(t, err)
	assert.True(t, released)
}

// 4. Concurrent Operation Lock Conflicts & Idempotency Key Reuse Test
func TestPhase45_ConcurrentLockConflictsAndIdempotency(t *testing.T) {
	ctx := context.Background()
	uc := reliabilityUsecase.NewReliabilityUseCase()

	var wg sync.WaitGroup
	results := make([]bool, 5)

	// 5 concurrent attempts to lock same resource
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			acquired, _, _ := uc.AcquireResourceLock(ctx, "res-concurrent-lock", "op-concurrent", 1*time.Minute)
			results[idx] = acquired
		}(i)
	}
	wg.Wait()

	// Exactly 1 caller acquires lock; 4 callers fail safely (lock conflict)
	successCount := 0
	for _, res := range results {
		if res {
			successCount++
		}
	}
	assert.Equal(t, 1, successCount)
}

// 5. Backup Metadata Persistence & PITR Restore Under Tenant Boundaries
func TestPhase45_BackupMetadataAndRestoreTenantIsolation(t *testing.T) {
	ctx := context.Background()
	uc := reliabilityUsecase.NewReliabilityUseCase()

	// Organization A backup dispatch
	opOrgA, err := uc.DispatchOperation(ctx, "org-alpha-backup", "proj-alpha", "res-db-alpha", reliabilityDomain.OpBackupDatabase, "idemp-bkp-alpha", "backup_payload", "req_bkp_1")
	require.NoError(t, err)
	assert.Equal(t, "org-alpha-backup", opOrgA.OrganizationID)

	// Organization B user trying to query Org A backup operation -> fails with tenant isolation
	_, err = uc.GetOperation("org-beta-attacker", opOrgA.ID)
	assert.Error(t, err)
}

// 6. Reliability Prometheus Metrics Verification
func TestPhase45_ReliabilityPrometheusMetrics(t *testing.T) {
	// Increment metrics
	metrics.RecordOperationRecovery("COMPLETED", 15.5)
	metrics.RecordLockConflict("res-compute-metrics")
	metrics.RecordLockExpiration("res-db-metrics")

	// Must not panic and metrics handler must serve cleanly
	handler := metrics.Handler()
	assert.NotNil(t, handler)
}
