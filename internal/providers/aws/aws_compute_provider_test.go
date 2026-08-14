package aws

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
)

func TestAWSComputeProvider_CreateInstance_Success(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	os.Setenv("AWS_REGION", "us-east-1")
	defer os.Unsetenv("AWS_ENABLED")
	defer os.Unsetenv("AWS_REGION")

	mockEC2 := NewMockEC2Client(true)
	provider := NewAWSComputeProvider(mockEC2)

	inst := &computeDomain.ComputeInstance{
		ID:             "inst-test-01",
		ResourceID:     "arnv:vm:us-east-1:proj-1:compute/test-ec2",
		OrganizationID: "org-alpha-101",
		ProjectID:      "proj-main-202",
		Name:           "test-ec2-node",
		Slug:           "test-ec2-node",
		PlanID:         "plan-small",
		ImageID:        "ami-0c55b159cbfafe1f0",
	}

	created, err := provider.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, computeDomain.ProviderAWS, created.Provider)
	assert.Contains(t, created.ProviderInstanceID, "i-")
	assert.Equal(t, computeDomain.StatusRunning, created.Status)
	assert.Equal(t, computeDomain.HealthHealthy, created.Health)
	assert.NotEmpty(t, created.PrivateIP)
	assert.NotEmpty(t, created.PublicIP)

	// Verify instance retrieval & state mapping
	retrieved, err := provider.GetInstance(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, computeDomain.StatusRunning, retrieved.Status)
}

func TestAWSComputeProvider_CreateInstance_InvalidAMI(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockEC2 := NewMockEC2Client(true)
	provider := NewAWSComputeProvider(mockEC2)

	inst := &computeDomain.ComputeInstance{
		ID:             "inst-test-invalid-ami",
		ResourceID:     "arnv:vm:us-east-1:proj-1:compute/invalid-ami",
		OrganizationID: "org-alpha-101",
		ProjectID:      "proj-main-202",
		ImageID:        "invalid-ami-garbage-xyz",
	}

	created, err := provider.CreateInstance(context.Background(), inst)
	assert.Error(t, err)
	assert.Nil(t, created)
	assert.Contains(t, err.Error(), "PROVISIONING_BLOCKED")
}

func TestAWSComputeProvider_DisabledMode(t *testing.T) {
	os.Setenv("AWS_ENABLED", "false")
	defer os.Unsetenv("AWS_ENABLED")

	mockEC2 := NewMockEC2Client(true)
	provider := NewAWSComputeProvider(mockEC2)

	// Test HealthCheck reports NOT_CONFIGURED
	status, err := provider.HealthCheck(context.Background())
	require.NoError(t, err)
	assert.Equal(t, registry.StatusNotConfigured, status)

	// Test CreateInstance fails safely when AWS is disabled
	inst := &computeDomain.ComputeInstance{
		ID:             "inst-disabled-test",
		OrganizationID: "org-1",
		ProjectID:      "proj-1",
	}
	_, err = provider.CreateInstance(context.Background(), inst)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PROVIDER_DISABLED")
}

func TestAWSComputeProvider_Unauthenticated_AuthFailed(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	// Disconnected mock client simulates missing/invalid AWS credentials
	mockEC2 := NewMockEC2Client(false)
	provider := NewAWSComputeProvider(mockEC2)

	status, err := provider.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Equal(t, registry.StatusAuthFailed, status)
}

func TestAWSComputeProvider_StartStopDelete(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockEC2 := NewMockEC2Client(true)
	provider := NewAWSComputeProvider(mockEC2)

	inst := &computeDomain.ComputeInstance{
		ID:             "inst-lifecycle-01",
		OrganizationID: "org-1",
		ProjectID:      "proj-1",
		ImageID:        "ami-0c55b159cbfafe1f0",
	}

	created, err := provider.CreateInstance(context.Background(), inst)
	require.NoError(t, err)

	// Stop Instance
	err = provider.StopInstance(context.Background(), created.ID)
	require.NoError(t, err)
	instStop, _ := provider.GetInstance(context.Background(), created.ID)
	assert.Equal(t, computeDomain.StatusStopped, instStop.Status)

	// Start Instance
	err = provider.StartInstance(context.Background(), created.ID)
	require.NoError(t, err)
	instStart, _ := provider.GetInstance(context.Background(), created.ID)
	assert.Equal(t, computeDomain.StatusRunning, instStart.Status)

	// Delete Instance
	err = provider.DeleteInstance(context.Background(), created.ID)
	require.NoError(t, err)
	instDel, _ := provider.GetInstance(context.Background(), created.ID)
	assert.Equal(t, computeDomain.StatusDeleted, instDel.Status)
}

func TestAWSComputeProvider_TenantIsolation(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockEC2 := NewMockEC2Client(true)
	provider := NewAWSComputeProvider(mockEC2)

	instOrgA := &computeDomain.ComputeInstance{
		ID:             "inst-org-a",
		OrganizationID: "org-tenant-A",
		ProjectID:      "proj-A",
		ImageID:        "ami-0c55b159cbfafe1f0",
	}
	_, err := provider.CreateInstance(context.Background(), instOrgA)
	require.NoError(t, err)

	instOrgB := &computeDomain.ComputeInstance{
		ID:             "inst-org-b",
		OrganizationID: "org-tenant-B",
		ProjectID:      "proj-B",
		ImageID:        "ami-0c55b159cbfafe1f0",
	}
	_, err = provider.CreateInstance(context.Background(), instOrgB)
	require.NoError(t, err)

	// List instances for Project A should only contain Org A instances
	listProjA, err := provider.ListInstances(context.Background(), "proj-A")
	require.NoError(t, err)
	assert.Len(t, listProjA, 1)
	assert.Equal(t, "org-tenant-A", listProjA[0].OrganizationID)
}

func TestAWSComputeProvider_MetricsAndHealth(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockEC2 := NewMockEC2Client(true)
	provider := NewAWSComputeProvider(mockEC2)

	inst := &computeDomain.ComputeInstance{
		ID:             "inst-metrics-01",
		OrganizationID: "org-1",
		ProjectID:      "proj-1",
		ImageID:        "ami-0c55b159cbfafe1f0",
	}
	created, err := provider.CreateInstance(context.Background(), inst)
	require.NoError(t, err)

	health, err := provider.GetInstanceHealth(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, computeDomain.HealthHealthy, health)

	metrics, err := provider.GetInstanceMetrics(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "NOT_IMPLEMENTED", metrics["cloudwatch_status"])
}
