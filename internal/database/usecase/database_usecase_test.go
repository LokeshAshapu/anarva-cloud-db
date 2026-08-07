package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/driver"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type mockInstanceRepo struct {
	instances map[string]*domain.DatabaseInstance
}

func newMockInstanceRepo() domain.InstanceRepository {
	return &mockInstanceRepo{instances: make(map[string]*domain.DatabaseInstance)}
}

func (m *mockInstanceRepo) Create(ctx context.Context, instance *domain.DatabaseInstance) error {
	m.instances[instance.ID] = instance
	return nil
}

func (m *mockInstanceRepo) GetByID(ctx context.Context, id string) (*domain.DatabaseInstance, error) {
	if inst, ok := m.instances[id]; ok {
		return inst, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "instance not found")
}

func (m *mockInstanceRepo) ListByProjectID(ctx context.Context, projectID string) ([]*domain.DatabaseInstance, error) {
	var list []*domain.DatabaseInstance
	for _, inst := range m.instances {
		if inst.ProjectID == projectID {
			list = append(list, inst)
		}
	}
	return list, nil
}

func (m *mockInstanceRepo) Update(ctx context.Context, instance *domain.DatabaseInstance) error {
	m.instances[instance.ID] = instance
	return nil
}

func (m *mockInstanceRepo) Delete(ctx context.Context, id string) error {
	delete(m.instances, id)
	return nil
}

func (m *mockInstanceRepo) CountByProjectID(ctx context.Context, projectID string) (int64, error) {
	var count int64
	for _, inst := range m.instances {
		if inst.ProjectID == projectID && inst.Status != domain.StatusTerminated {
			count++
		}
	}
	return count, nil
}

func setupDatabaseUseCase() DatabaseUseCase {
	repo := newMockInstanceRepo()
	drv := driver.NewMockProvisionerDriver()
	secretKey := "32-bytes-long-super-secret-key-x"
	return NewDatabaseUseCase(repo, drv, secretKey)
}

func TestDatabaseInstanceLifecycle(t *testing.T) {
	uc := setupDatabaseUseCase()
	ctx := context.Background()

	projectID := "project-uuid-101"
	instance, connStr, err := uc.CreateDatabase(ctx, projectID, "Primary DB", domain.EnginePostgreSQL, 20)
	require.NoError(t, err)
	assert.Equal(t, "Primary DB", instance.Name)
	assert.Equal(t, domain.StatusRunning, instance.Status)
	assert.NotEmpty(t, connStr)

	// Get Connection String
	retrievedConnStr, err := uc.GetConnectionString(ctx, instance.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, retrievedConnStr)

	// Stop & Start
	err = uc.StopDatabase(ctx, instance.ID)
	require.NoError(t, err)

	stoppedInst, _ := uc.GetDatabase(ctx, instance.ID)
	assert.Equal(t, domain.StatusStopped, stoppedInst.Status)

	err = uc.StartDatabase(ctx, instance.ID)
	require.NoError(t, err)

	// Delete
	err = uc.DeleteDatabase(ctx, instance.ID)
	require.NoError(t, err)

	_, err = uc.GetDatabase(ctx, instance.ID)
	assert.Error(t, err)
}

func TestDatabaseQuotaEnforcement(t *testing.T) {
	uc := setupDatabaseUseCase()
	ctx := context.Background()

	projectID := "quota-project-uuid"
	for i := 0; i < 5; i++ {
		_, _, err := uc.CreateDatabase(ctx, projectID, "DB Instance", domain.EnginePostgreSQL, 10)
		require.NoError(t, err)
	}

	// 6th database creation must fail due to quota
	_, _, err := uc.CreateDatabase(ctx, projectID, "Exceeded DB", domain.EnginePostgreSQL, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quota reached")
}
