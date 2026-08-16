package reliability_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	healthService "github.com/anarva-cloud/anarva-cloud-db/internal/health"
	reliabilityDomain "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/domain"
	reliabilityRepo "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/repository"
	reliabilityUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
)

// 1. Health & Readiness Engine Tests
func TestPhase43_HealthAndReadinessEngine(t *testing.T) {
	uc := reliabilityUsecase.NewReliabilityUseCase()
	cfg := &config.Config{
		Environment: "development",
		Server:      config.ServerConfig{Port: 8080},
		JWT:         config.JWTConfig{Secret: "test_secret"},
	}

	hSvc := healthService.NewHealthService(nil, cfg, nil, uc, "0.1.0")

	// Liveness /health test
	reqHealth := httptest.NewRequest("GET", "/health", nil)
	wHealth := httptest.NewRecorder()
	hSvc.HandleHealth(wHealth, reqHealth)
	assert.Equal(t, http.StatusOK, wHealth.Code)
	assert.Contains(t, wHealth.Body.String(), `"status":"UP"`)
	assert.Contains(t, wHealth.Body.String(), `"service":"anarva-control-plane"`)

	// Readiness /readiness test (dev mode without DB pool returns DEGRADED status with HTTP 200)
	reqReady := httptest.NewRequest("GET", "/readiness", nil)
	wReady := httptest.NewRecorder()
	hSvc.HandleReadiness(wReady, reqReady)
	assert.Equal(t, http.StatusOK, wReady.Code)
	assert.Contains(t, wReady.Body.String(), `"status":"DEGRADED"`)

	// Production readiness failure when config is unready
	prodCfg := &config.Config{
		Environment: "production",
		Server:      config.ServerConfig{Port: 8080},
		JWT:         config.JWTConfig{Secret: "anarva_cloud_db_super_secret_jwt_key_2026"}, // Default secret in prod triggers failure
	}
	hSvcProd := healthService.NewHealthService(nil, prodCfg, nil, uc, "0.1.0")

	wProd := httptest.NewRecorder()
	hSvcProd.HandleReadiness(wProd, reqReady)
	assert.Equal(t, http.StatusServiceUnavailable, wProd.Code)
	assert.Contains(t, wProd.Body.String(), `"status":"NOT_READY"`)
	assert.Contains(t, wProd.Body.String(), `"configuration":"UNAVAILABLE"`)
}

// 2. Operation Lifecycle & Valid State Transitions
func TestPhase43_OperationLifecycleAndTransitions(t *testing.T) {
	ctx := context.Background()
	uc := reliabilityUsecase.NewReliabilityUseCase()

	// Dispatch Operation
	op, err := uc.DispatchOperation(ctx, "org-alpha", "proj-alpha", "res-db-01", reliabilityDomain.OpCreateDatabase, "idemp-key-101", `{"name":"testdb"}`, "req_101")
	require.NoError(t, err)
	assert.Equal(t, reliabilityDomain.OpStatusRunning, op.Status)
	assert.Equal(t, "org-alpha", op.OrganizationID)
	assert.Equal(t, "req_101", op.RequestID)

	// Valid Transition: RUNNING -> SUCCEEDED
	completedOp, err := uc.CompleteOperation(ctx, op.ID, "")
	require.NoError(t, err)
	assert.Equal(t, reliabilityDomain.OpStatusSucceeded, completedOp.Status)
	assert.Equal(t, 100, completedOp.Progress)

	// Invalid Transition: SUCCEEDED -> RUNNING must be rejected
	assert.False(t, reliabilityDomain.IsValidStateTransition(reliabilityDomain.OpStatusSucceeded, reliabilityDomain.OpStatusRunning))
}

// 3. Tenant Isolation & Query Filtering
func TestPhase43_TenantIsolationAndFiltering(t *testing.T) {
	ctx := context.Background()
	uc := reliabilityUsecase.NewReliabilityUseCase()

	// Create ops for Org Alpha and Org Beta
	opAlpha, err := uc.DispatchOperation(ctx, "org-alpha", "proj-alpha", "res-alpha-01", reliabilityDomain.OpCreateCompute, "idemp-alpha", "payload-a", "req_a")
	require.NoError(t, err)

	_, err = uc.DispatchOperation(ctx, "org-beta", "proj-beta", "res-beta-01", reliabilityDomain.OpCreateDatabase, "idemp-beta", "payload-b", "req_b")
	require.NoError(t, err)

	// Org Alpha query must NOT return Org Beta ops
	opsAlpha, totalAlpha, err := uc.ListOperations(ctx, "org-alpha", reliabilityRepo.OperationQueryFilters{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalAlpha)
	assert.Equal(t, opAlpha.ID, opsAlpha[0].ID)

	// Single op lookup cross-tenant MUST fail with Not Found / Access Denied
	_, err = uc.GetOperation("org-beta", opAlpha.ID)
	assert.Error(t, err)
}

// 4. Operation Timeout Detection
func TestPhase43_OperationTimeoutDetection(t *testing.T) {
	ctx := context.Background()
	uc := reliabilityUsecase.NewReliabilityUseCase()

	// Dispatch running operation
	op, err := uc.DispatchOperation(ctx, "org-timeout", "proj-timeout", "res-stale-01", reliabilityDomain.OpBackupDatabase, "", "payload-stale", "req_stale")
	require.NoError(t, err)
	assert.Equal(t, reliabilityDomain.OpStatusRunning, op.Status)

	// Backdate HeartbeatAt past timeout threshold (6 minutes ago)
	op.HeartbeatAt = time.Now().Add(-6 * time.Minute)

	// Trigger Timeout Detector
	timedOutCount := uc.DetectOperationTimeouts(ctx, 5*time.Minute)
	assert.Equal(t, 1, timedOutCount)

	// Verify operation status changed to TIMED_OUT
	fetched, err := uc.GetOperation("org-timeout", op.ID)
	require.NoError(t, err)
	assert.Equal(t, reliabilityDomain.OpStatusTimedOut, fetched.Status)
	assert.Equal(t, "OPERATION_TIMED_OUT", fetched.ErrorCode)
}

// 5. Fail-Closed Production Configuration Validation
func TestPhase43_ProductionConfigurationValidation(t *testing.T) {
	// Dev config allows defaults
	devCfg := &config.Config{
		Environment: "development",
		Server:      config.ServerConfig{Port: 8080},
		JWT:         config.JWTConfig{Secret: "anarva_cloud_db_super_secret_jwt_key_2026"},
	}
	require.NoError(t, config.ValidateProductionConfig(devCfg))

	// Prod config fails closed with default JWT secret
	prodDefaultJwtCfg := &config.Config{
		Environment: "production",
		Server:      config.ServerConfig{Port: 8080},
		JWT:         config.JWTConfig{Secret: "anarva_cloud_db_super_secret_jwt_key_2026"},
		Database:    config.DatabaseConfig{Host: "prod-db.internal"},
	}
	err := config.ValidateProductionConfig(prodDefaultJwtCfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CONFIG_VALIDATION_FAILURE")

	// Prod config with strong parameters passes
	prodValidCfg := &config.Config{
		Environment: "production",
		Server:      config.ServerConfig{Port: 8080},
		JWT:         config.JWTConfig{Secret: "anarva_prod_ultra_secure_jwt_token_key_987654321"},
		Database:    config.DatabaseConfig{Host: "prod-db.internal"},
	}
	require.NoError(t, config.ValidateProductionConfig(prodValidCfg))
}
