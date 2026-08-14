package developer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/developer/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/developer/usecase"
)

func TestDeveloper_CreateAndAuthenticateAPIKey(t *testing.T) {
	uc := usecase.NewDeveloperUseCase()

	key, rawSecret, err := uc.CreateAPIKey(context.Background(), "Prod Backend Key", "org-default", "proj-default", "lokeshashapu@gmail.com", []string{"compute.read", "database.read"}, true)
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.NotEmpty(t, rawSecret)
	assert.Contains(t, rawSecret, "anarva_live_")
	assert.Equal(t, domain.KeyStatusActive, key.Status)

	// Validate authentication with plaintext rawSecret
	validatedKey, err := uc.ValidateAPIKey(context.Background(), rawSecret)
	require.NoError(t, err)
	assert.Equal(t, key.ID, validatedKey.ID)
	assert.NotNil(t, validatedKey.LastUsedAt)
}

func TestDeveloper_RevokeAPIKey(t *testing.T) {
	uc := usecase.NewDeveloperUseCase()

	key, rawSecret, err := uc.CreateAPIKey(context.Background(), "Temporary Key", "org-default", "proj-default", "lokeshashapu@gmail.com", []string{"compute.read"}, false)
	require.NoError(t, err)

	// Revoke Key
	err = uc.RevokeAPIKey(context.Background(), key.ID)
	require.NoError(t, err)

	// Authenticating with revoked secret fails
	_, err = uc.ValidateAPIKey(context.Background(), rawSecret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "revoked")
}

func TestDeveloper_RotateAPIKey(t *testing.T) {
	uc := usecase.NewDeveloperUseCase()

	key, oldSecret, err := uc.CreateAPIKey(context.Background(), "Rotation Key", "org-default", "proj-default", "lokeshashapu@gmail.com", []string{"storage.read"}, true)
	require.NoError(t, err)

	// Rotate Key
	rotatedKey, newSecret, err := uc.RotateAPIKey(context.Background(), key.ID, "lokeshashapu@gmail.com")
	require.NoError(t, err)
	assert.NotNil(t, rotatedKey)
	assert.NotEmpty(t, newSecret)
	assert.NotEqual(t, oldSecret, newSecret)

	// Old secret fails authentication
	_, err = uc.ValidateAPIKey(context.Background(), oldSecret)
	assert.Error(t, err)

	// New secret succeeds authentication
	validatedKey, err := uc.ValidateAPIKey(context.Background(), newSecret)
	require.NoError(t, err)
	assert.Equal(t, key.ID, validatedKey.ID)
}

func TestDeveloper_TenantIsolation(t *testing.T) {
	uc := usecase.NewDeveloperUseCase()

	keyOrgA, _, err := uc.CreateAPIKey(context.Background(), "Org A Key", "org-a", "proj-a", "user-a@anarva.io", []string{"compute.read"}, true)
	require.NoError(t, err)

	// Org B querying API keys for proj-b does not return Org A key
	keysOrgB := uc.ListAPIKeys(context.Background(), "proj-b")
	for _, k := range keysOrgB {
		assert.NotEqual(t, keyOrgA.ID, k.ID)
	}
}
