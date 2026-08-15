package aws

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
)

func TestProviderMode_ProductionFailClosed(t *testing.T) {
	v := NewProviderModeValidator()
	ctx := context.Background()

	// PRODUCTION + REAL without credentials MUST fail closed
	err := v.ValidateExecutionMode(ctx, "PRODUCTION", "REAL", "", "", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PROVIDER_AUTHENTICATION_FAILED")
	assert.Contains(t, err.Error(), "Production real provider mode requires active AWS_ACCESS_KEY_ID")

	// PRODUCTION + REAL with active credentials MUST succeed
	err = v.ValidateExecutionMode(ctx, "PRODUCTION", "REAL", "AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", false)
	assert.NoError(t, err)
}

func TestProviderMode_ProductionMockRejection(t *testing.T) {
	v := NewProviderModeValidator()
	ctx := context.Background()

	// PRODUCTION + MOCK client MUST be rejected
	err := v.ValidateExecutionMode(ctx, "PRODUCTION", "LOCAL", "AKIAIOSFODNN7EXAMPLE", "secret", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PROVIDER_INVALID_CONFIGURATION")
	assert.Contains(t, err.Error(), "Production environment MUST NOT execute mock infrastructure clients")
}

func TestProviderMode_DevelopmentLocalAllowed(t *testing.T) {
	v := NewProviderModeValidator()
	ctx := context.Background()

	// DEVELOPMENT + LOCAL mode MUST be allowed
	err := v.ValidateExecutionMode(ctx, "DEVELOPMENT", "LOCAL", "", "", false)
	assert.NoError(t, err)
}

func TestProviderRegistry_CapabilityGuard(t *testing.T) {
	reg := registry.NewProviderRegistry()
	ctx := context.Background()

	// Compute capability check for Local Docker MUST succeed
	err := reg.ValidateCapability(ctx, "provider-local-docker", "compute")
	assert.NoError(t, err)

	// Kubernetes capability check for Local Docker MUST fail (capability = false)
	err = reg.ValidateCapability(ctx, "provider-local-docker", "kubernetes")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PROVIDER_CAPABILITY_NOT_SUPPORTED")
	assert.Contains(t, err.Error(), "does not support capability 'kubernetes'")
}

func TestAWSError_NormalizationAndSecretRedaction(t *testing.T) {
	// Secret Redaction Check
	rawErrWithSecret := fmt.Errorf("AccessDenied for user with key anarva_live_secret123456789 and password=SuperSecretPassword123")
	mappedErr := MapAWSError(rawErrWithSecret)

	assert.Error(t, mappedErr)
	assert.Contains(t, mappedErr.Error(), "PROVIDER_PERMISSION_DENIED")
	assert.False(t, strings.Contains(mappedErr.Error(), "anarva_live_secret123456789"))
	assert.False(t, strings.Contains(mappedErr.Error(), "SuperSecretPassword123"))
	assert.Contains(t, mappedErr.Error(), "[REDACTED_SECRET]")
}

func TestRetry_ExponentialBackoffAndJitter(t *testing.T) {
	ctx := context.Background()

	// Non-retryable error (401 Unauthorized) MUST fail immediately without retries
	attempts := 0
	_, err := ExecuteWithRetry(ctx, DefaultRetryConfig(), "TestNonRetryable", func(c context.Context) (string, error) {
		attempts++
		return "", fmt.Errorf("UnrecognizedClientException: 401 Unauthorized")
	})

	assert.Error(t, err)
	assert.Equal(t, 1, attempts)
	assert.Contains(t, err.Error(), "PROVIDER_AUTHENTICATION_FAILED")

	// Retryable error (429 Throttling) MUST retry and succeed on 2nd attempt
	retryAttempts := 0
	res, err := ExecuteWithRetry(ctx, RetryConfig{MaxRetries: 3, BaseDelay: 1 * time.Millisecond, Timeout: 1 * time.Second}, "TestRetryable", func(c context.Context) (string, error) {
		retryAttempts++
		if retryAttempts < 2 {
			return "", fmt.Errorf("RequestLimitExceeded: 429 Throttling")
		}
		return "PROVISIONED", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "PROVISIONED", res)
	assert.Equal(t, 2, retryAttempts)
}

func TestMapping_TenantIsolation(t *testing.T) {
	repo := mapping.NewMappingRepository()
	ctx := context.Background()

	require.NoError(t, repo.SaveMapping(&mapping.ProviderResourceMapping{
		AnarvaResourceID:     "res-rds-101",
		OrganizationID:       "org-alpha",
		ProjectID:            "proj-alpha",
		Provider:             "AWS",
		ProviderResourceID:   "db-instance-101",
		ProviderResourceType: "RDS_POSTGRESQL",
		Region:               "us-east-1",
		Status:               "ACTIVE",
	}))

	// Org Alpha query MUST succeed
	m, err := repo.GetTenantScopedMapping(ctx, "org-alpha", "proj-alpha", "res-rds-101")
	require.NoError(t, err)
	assert.Equal(t, "db-instance-101", m.ProviderResourceID)

	// Org Beta query MUST fail with TENANT_ISOLATION_VIOLATION
	_, err = repo.GetTenantScopedMapping(ctx, "org-beta", "proj-beta", "res-rds-101")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TENANT_ISOLATION_VIOLATION")
}
