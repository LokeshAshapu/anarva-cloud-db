package reliability

import (
	"context"
	"testing"

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
