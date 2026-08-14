package ha

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/aws"
)

func TestHAEngine_ToggleMultiAZ(t *testing.T) {
	mockRDS := aws.NewMockRDSClient(true)
	engine := NewHAOrchestrationEngine(mockRDS)

	cfg, err := engine.ToggleMultiAZ(context.Background(), "org-default", "proj-default", "anarva-rds-prod-01", true)
	require.NoError(t, err)
	assert.True(t, cfg.MultiAZ)
	assert.Equal(t, "HA_ENABLED", cfg.HAStatus)
	assert.NotEmpty(t, cfg.SecondaryAvailabilityZone)
}

func TestHAEngine_ControlledFailover_Success(t *testing.T) {
	mockRDS := aws.NewMockRDSClient(true)
	engine := NewHAOrchestrationEngine(mockRDS)

	// Primary starts in ap-south-1a, Secondary in ap-south-1b
	job, err := engine.TriggerFailover(context.Background(), "org-default", "proj-default", "anarva-rds-prod-01", "lokeshashapu@gmail.com")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "COMPLETED", job.Status)
	assert.Equal(t, "ap-south-1a", job.PreviousPrimaryAZ)
	assert.Equal(t, "ap-south-1b", job.NewPrimaryAZ)

	// Verify HAConfig updated primary AZ to ap-south-1b
	cfg, err := engine.GetHAConfig("org-default", "anarva-rds-prod-01")
	require.NoError(t, err)
	assert.Equal(t, "ap-south-1b", cfg.PrimaryAvailabilityZone)
	assert.Equal(t, "ap-south-1a", cfg.SecondaryAvailabilityZone)
}

func TestHAEngine_FailoverOnSingleAZ_Blocked(t *testing.T) {
	mockRDS := aws.NewMockRDSClient(true)
	engine := NewHAOrchestrationEngine(mockRDS)

	// Disable Multi-AZ
	_, err := engine.ToggleMultiAZ(context.Background(), "org-default", "proj-default", "anarva-rds-prod-01", false)
	require.NoError(t, err)

	// Triggering failover on Single-AZ DB must fail
	_, err = engine.TriggerFailover(context.Background(), "org-default", "proj-default", "anarva-rds-prod-01", "lokeshashapu@gmail.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_HA_STATE")
}

func TestHAEngine_TenantIsolation(t *testing.T) {
	mockRDS := aws.NewMockRDSClient(true)
	engine := NewHAOrchestrationEngine(mockRDS)

	// Org B attempting failover on Org A resource must fail
	_, err := engine.TriggerFailover(context.Background(), "org-b", "proj-b", "anarva-rds-prod-01", "attacker@anarva.io")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RESOURCE_NOT_FOUND")
}
