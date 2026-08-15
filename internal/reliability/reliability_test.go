package reliability

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
)

func TestReliability_IdempotentOperationDispatch(t *testing.T) {
	uc := usecase.NewReliabilityUseCase()
	ctx := context.Background()

	key := "idem-key-101"
	payload := `{"name":"prod-db","acu":2.0}`

	// First Request
	op1, err := uc.DispatchOperation(ctx, "org-default", "proj-default", "res-rds-01", domain.OpCreateDatabase, key, payload, "req_01")
	require.NoError(t, err)
	assert.NotNil(t, op1)

	// Second Identical Request with Same Idempotency Key
	op2, err := uc.DispatchOperation(ctx, "org-default", "proj-default", "res-rds-01", domain.OpCreateDatabase, key, payload, "req_02")
	require.NoError(t, err)
	assert.Equal(t, op1.ID, op2.ID) // Returns same operation instance
}

func TestReliability_IdempotencyKeyReuseConflict(t *testing.T) {
	uc := usecase.NewReliabilityUseCase()
	ctx := context.Background()

	key := "idem-key-reuse-test"
	payloadOriginal := `{"name":"prod-db","acu":2.0}`
	payloadConflicting := `{"name":"different-db","acu":50.0}`

	// First Request
	_, err := uc.DispatchOperation(ctx, "org-default", "proj-default", "res-rds-01", domain.OpCreateDatabase, key, payloadOriginal, "req_01")
	require.NoError(t, err)

	// Second Request using Same Key but Different Payload MUST Fail
	_, err = uc.DispatchOperation(ctx, "org-default", "proj-default", "res-rds-01", domain.OpCreateDatabase, key, payloadConflicting, "req_02")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "IDEMPOTENCY_KEY_REUSE")
}

func TestReliability_QuotaEnforcement(t *testing.T) {
	uc := usecase.NewReliabilityUseCase()

	// Under Quota Request
	err := uc.ValidateAndReserveQuota("org-default", "proj-default", 10.0, 1, 100)
	require.NoError(t, err)

	// Over Quota Request (ACU limit is 100.0)
	err = uc.ValidateAndReserveQuota("org-default", "proj-default", 200.0, 1, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "QUOTA_EXCEEDED")
}

func TestReliability_ResourceLockLease(t *testing.T) {
	uc := usecase.NewReliabilityUseCase()
	ctx := context.Background()

	// Dispatch Operation A (Acquires Lock on res-locked-01)
	opA, err := uc.DispatchOperation(ctx, "org-default", "proj-default", "res-locked-01", domain.OpUpdateDatabase, "", `{"action":"update"}`, "req_lock_01")
	require.NoError(t, err)

	// Dispatch Operation B on Same Resource (MUST fail with RESOURCE_LOCKED)
	_, err = uc.DispatchOperation(ctx, "org-default", "proj-default", "res-locked-01", domain.OpDeleteDatabase, "", `{"action":"delete"}`, "req_lock_02")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RESOURCE_LOCKED")

	// Complete Operation A (Releases Lock)
	_, err = uc.CompleteOperation(ctx, opA.ID, "")
	require.NoError(t, err)

	// Dispatch Operation B now succeeds
	opB, err := uc.DispatchOperation(ctx, "org-default", "proj-default", "res-locked-01", domain.OpDeleteDatabase, "", `{"action":"delete"}`, "req_lock_03")
	require.NoError(t, err)
	assert.NotNil(t, opB)
}

func TestReliability_BackendRestartOperationRecovery(t *testing.T) {
	uc := usecase.NewReliabilityUseCase()
	ctx := context.Background()

	// Create RUNNING operation
	op, err := uc.DispatchOperation(ctx, "org-default", "proj-default", "res-restart-test", domain.OpCreateCompute, "", `{"name":"ace-01"}`, "req_restart_01")
	require.NoError(t, err)
	assert.Equal(t, domain.OpStatusRunning, op.Status)

	// Simulate Backend Process Crash & Restart Recovery
	reconciled := uc.ReconcileInterruptedOperations(ctx)
	assert.Equal(t, 1, reconciled)

	// Assert operation transitioned to SUCCEEDED with timeline reconciliation event
	recoveredOp, err := uc.GetOperation("org-default", op.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.OpStatusSucceeded, recoveredOp.Status)
	assert.NotEmpty(t, recoveredOp.Timeline)
}

func TestReliability_AppendOnlyAuditLog(t *testing.T) {
	uc := usecase.NewReliabilityUseCase()

	audits := uc.ListAuditEvents("org-default", "proj-default")
	assert.NotEmpty(t, audits)

	// Verify no raw secrets in metadata
	for _, a := range audits {
		for _, v := range a.Metadata {
			assert.NotContains(t, v, "anarva_live_")
			assert.NotContains(t, v, "anarva_test_")
		}
	}
}

func TestStateTransitions_ValidAndInvalid(t *testing.T) {
	// Valid Transitions
	assert.True(t, domain.IsValidStateTransition(domain.OpStatusPending, domain.OpStatusRunning))
	assert.True(t, domain.IsValidStateTransition(domain.OpStatusRunning, domain.OpStatusSucceeded))
	assert.True(t, domain.IsValidStateTransition(domain.OpStatusRunning, domain.OpStatusFailed))
	assert.True(t, domain.IsValidStateTransition(domain.OpStatusRunning, domain.OpStatusRecovering))
	assert.True(t, domain.IsValidStateTransition(domain.OpStatusRecovering, domain.OpStatusRunning))

	// Invalid Transitions from Terminal States
	assert.False(t, domain.IsValidStateTransition(domain.OpStatusSucceeded, domain.OpStatusRunning))
	assert.False(t, domain.IsValidStateTransition(domain.OpStatusFailed, domain.OpStatusPending))
}

func TestDistributedLock_ConcurrentAcquireConflict(t *testing.T) {
	uc := usecase.NewReliabilityUseCase()
	ctx := context.Background()

	resID := "res-concurrent-lock-01"
	var successCount int32
	var conflictCount int32

	var wg sync.WaitGroup
	wg.Add(2)

	// Worker A
	go func() {
		defer wg.Done()
		_, err := uc.DispatchOperation(ctx, "org-default", "proj-default", resID, domain.OpCreateDatabase, "", `{"w":"A"}`, "req_A")
		if err == nil {
			atomic.AddInt32(&successCount, 1)
		} else {
			atomic.AddInt32(&conflictCount, 1)
		}
	}()

	// Worker B
	go func() {
		defer wg.Done()
		_, err := uc.DispatchOperation(ctx, "org-default", "proj-default", resID, domain.OpCreateDatabase, "", `{"w":"B"}`, "req_B")
		if err == nil {
			atomic.AddInt32(&successCount, 1)
		} else {
			atomic.AddInt32(&conflictCount, 1)
		}
	}()

	wg.Wait()

	// Exactly ONE worker succeeds, exactly ONE receives lock conflict
	assert.Equal(t, int32(1), successCount)
	assert.Equal(t, int32(1), conflictCount)
}

func TestRecoveryWorker_LifecycleDaemon(t *testing.T) {
	uc := usecase.NewReliabilityUseCase()
	worker := usecase.NewRecoveryWorker(uc, usecase.RecoveryWorkerConfig{Interval: 50 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	worker.Stop()
}
