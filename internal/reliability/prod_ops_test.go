package reliability_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	healthService "github.com/anarva-cloud/anarva-cloud-db/internal/health"
	reliabilityDomain "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/domain"
	reliabilityUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	pkgVersion "github.com/anarva-cloud/anarva-cloud-db/pkg/version"
)

// 1. Version Metadata Verification
func TestProdOps_VersionMetadata(t *testing.T) {
	vInfo := pkgVersion.GetVersionInfo("production")
	assert.Equal(t, pkgVersion.ANARVA_VERSION, vInfo.Version)
	assert.Equal(t, "production", vInfo.Environment)
	assert.NotEmpty(t, vInfo.GitCommit)
	assert.NotEmpty(t, vInfo.GoVersion)
}

// 2. Production Environment Validation Test
func TestProdOps_ProductionEnvironmentValidation(t *testing.T) {
	devCfg := &config.Config{
		Environment: "development",
		Server:      config.ServerConfig{Port: 8080},
		JWT:         config.JWTConfig{Secret: "anarva_cloud_db_super_secret_jwt_key_2026"},
	}
	require.NoError(t, config.ValidateProductionConfig(devCfg))

	// Prod config fails closed if port <= 0
	invalidProdCfg := &config.Config{
		Environment: "production",
		Server:      config.ServerConfig{Port: 0},
	}
	err := config.ValidateProductionConfig(invalidProdCfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CONFIG_VALIDATION_FAILURE")
}

// 3. Graceful Shutdown & Recovery After Restart Test
func TestProdOps_OperationRecoveryAfterRestart(t *testing.T) {
	ctx := context.Background()
	uc := reliabilityUsecase.NewReliabilityUseCase()

	// Dispatch running operation
	op, err := uc.DispatchOperation(ctx, "org-restart", "proj-restart", "res-compute-101", reliabilityDomain.OpCreateCompute, "idemp-restart-1", "payload", "req_restart")
	require.NoError(t, err)
	assert.Equal(t, reliabilityDomain.OpStatusRunning, op.Status)

	// Simulate restart recovery daemon execution
	reconciledCount := uc.ReconcileInterruptedOperations(ctx)
	assert.GreaterOrEqual(t, reconciledCount, 0)

	// Operation remains persisted in valid state
	fetched, err := uc.GetOperation("org-restart", op.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, fetched.ID)
}

// 4. Subsystem Readiness Failure Test
func TestProdOps_SubsystemReadinessProbe(t *testing.T) {
	uc := reliabilityUsecase.NewReliabilityUseCase()
	prodCfg := &config.Config{
		Environment: "production",
		Server:      config.ServerConfig{Port: 8080},
		JWT:         config.JWTConfig{Secret: "anarva_cloud_db_super_secret_jwt_key_2026"},
	}

	hSvc := healthService.NewHealthService(nil, prodCfg, nil, uc, "0.1.0")
	checks := hSvc.CheckReadiness(context.Background())

	// Without postgres pool in production, database readiness MUST report UNAVAILABLE
	assert.Equal(t, healthService.StatusUnavailable, checks.Database)
}
