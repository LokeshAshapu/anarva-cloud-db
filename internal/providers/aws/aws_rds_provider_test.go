package aws

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	databaseDomain "github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
)

func TestAWSRDSProvider_CreateDatabaseInstance_Success(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	os.Setenv("AWS_REGION", "us-east-1")
	defer os.Unsetenv("AWS_ENABLED")
	defer os.Unsetenv("AWS_REGION")

	mockRDS := NewMockRDSClient(true)
	provider := NewAWSRDSProvider(mockRDS)

	inst := databaseDomain.NewDatabaseInstance(
		"proj-101",
		"production-postgres-db",
		databaseDomain.EnginePostgreSQL,
		"",
		5432,
		"prod_db",
		"anarva_admin",
		"",
		20,
	)

	created, err := provider.CreateDatabaseInstance(context.Background(), inst, "org-alpha-101")
	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, databaseDomain.EnginePostgreSQL, created.Engine)
	assert.Equal(t, databaseDomain.StatusRunning, created.Status)
	assert.Contains(t, created.Host, "anarva-rds-")
	assert.Equal(t, 5432, created.Port)

	// Verify secret safety: password is KMS encrypted, never plain text
	assert.NotEqual(t, "plain-password", created.PasswordEncrypted)
	assert.Equal(t, "KMS_ENCRYPTED_VAULT_REF", created.PasswordEncrypted)
}

func TestAWSRDSProvider_InvalidEngine_Blocked(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockRDS := NewMockRDSClient(true)
	provider := NewAWSRDSProvider(mockRDS)

	inst := &databaseDomain.DatabaseInstance{
		ID:        "db-mysql-test",
		ProjectID: "proj-101",
		Name:      "unsupported-mysql",
		Engine:    databaseDomain.EngineType("mysql"),
	}

	created, err := provider.CreateDatabaseInstance(context.Background(), inst, "org-alpha-101")
	assert.Error(t, err)
	assert.Nil(t, created)
	assert.Contains(t, err.Error(), "PROVISIONING_BLOCKED")
}

func TestAWSRDSProvider_DisabledMode(t *testing.T) {
	os.Setenv("AWS_ENABLED", "false")
	defer os.Unsetenv("AWS_ENABLED")

	mockRDS := NewMockRDSClient(true)
	provider := NewAWSRDSProvider(mockRDS)

	status, err := provider.HealthCheck(context.Background())
	require.NoError(t, err)
	assert.Equal(t, registry.StatusNotConfigured, status)

	inst := &databaseDomain.DatabaseInstance{ID: "db-disabled-test"}
	_, err = provider.CreateDatabaseInstance(context.Background(), inst, "org-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PROVIDER_DISABLED")
}

func TestAWSRDSProvider_DeleteDatabaseInstance(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	mockRDS := NewMockRDSClient(true)
	provider := NewAWSRDSProvider(mockRDS)

	inst := databaseDomain.NewDatabaseInstance("proj-1", "db-to-delete", databaseDomain.EnginePostgreSQL, "", 5432, "mydb", "admin", "", 10)
	created, err := provider.CreateDatabaseInstance(context.Background(), inst, "org-1")
	require.NoError(t, err)

	err = provider.DeleteDatabaseInstance(context.Background(), created.ID, "org-1")
	require.NoError(t, err)

	deleted, err := provider.GetDatabaseInstance(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, databaseDomain.StatusTerminated, deleted.Status)
}

func TestAWSRDSProvider_SecretGeneration(t *testing.T) {
	pass1, err1 := generateSecureRandomPassword(24)
	require.NoError(t, err1)
	assert.Len(t, pass1, 24)

	pass2, err2 := generateSecureRandomPassword(24)
	require.NoError(t, err2)
	assert.Len(t, pass2, 24)

	// Ensure generated passwords are unique
	assert.NotEqual(t, pass1, pass2)
}
