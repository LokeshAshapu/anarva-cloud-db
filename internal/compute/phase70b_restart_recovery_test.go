package compute_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
	computeRepo "github.com/anarva-cloud/anarva-cloud-db/internal/compute/repository"
	computeUseCase "github.com/anarva-cloud/anarva-cloud-db/internal/compute/usecase"
	mapping "github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
)

type memComputeRepo struct {
	instances map[string]*domain.ComputeInstance
}

func newMemComputeRepo() *memComputeRepo {
	return &memComputeRepo{instances: make(map[string]*domain.ComputeInstance)}
}

func (r *memComputeRepo) Create(ctx context.Context, inst *domain.ComputeInstance) error {
	r.instances[inst.ID] = inst
	return nil
}

func (r *memComputeRepo) GetByID(ctx context.Context, id string) (*domain.ComputeInstance, error) {
	if inst, ok := r.instances[id]; ok && inst.DeletedAt == nil {
		return inst, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *memComputeRepo) GetTenantScopedByID(ctx context.Context, orgID, projID, id string) (*domain.ComputeInstance, error) {
	inst, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if orgID != "" && inst.OrganizationID != "" && inst.OrganizationID != orgID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Organization '%s' is prohibited from accessing compute instance '%s'", orgID, id)
	}
	if projID != "" && inst.ProjectID != "" && inst.ProjectID != projID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Project '%s' is prohibited from accessing compute instance '%s'", projID, id)
	}
	return inst, nil
}

func (r *memComputeRepo) ListByProjectID(ctx context.Context, projectID string) ([]*domain.ComputeInstance, error) {
	var list []*domain.ComputeInstance
	for _, inst := range r.instances {
		if inst.ProjectID == projectID && inst.DeletedAt == nil {
			list = append(list, inst)
		}
	}
	return list, nil
}

func (r *memComputeRepo) Update(ctx context.Context, inst *domain.ComputeInstance) error {
	r.instances[inst.ID] = inst
	return nil
}

func (r *memComputeRepo) Delete(ctx context.Context, id string) error {
	delete(r.instances, id)
	return nil
}

func TestPhase70B_ComputeGatewayRestartRecovery(t *testing.T) {
	ctx := context.Background()

	// Setup initial gateway state
	cRepo := newMemComputeRepo()
	mapRepo := mapping.NewInMemoryMappingRepository()
	prov1 := computeProvider.NewLocalDockerComputeProvider()
	uc1 := computeUseCase.NewComputeUseCase(cRepo, nil, prov1)
	uc1.SetMappingRepository(mapRepo)

	// 1. Create Compute Instance for Tenant A
	instA, err := uc1.CreateInstance(ctx, &domain.ComputeInstance{
		ID:             "acu-inst-restart-70b",
		OrganizationID: "org-tenant-a",
		ProjectID:      "proj-tenant-a",
		Name:           "Worker 70B",
		ACU:            1.0,
		RegionID:       "us-east-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "org-tenant-a", instA.OrganizationID)
	assert.Equal(t, "proj-tenant-a", instA.ProjectID)

	// 2. Verify instance metadata exists in control plane repo
	storedInst, err := cRepo.GetByID(ctx, instA.ID)
	require.NoError(t, err)
	assert.Equal(t, instA.ID, storedInst.ID)

	// 3. Verify provider resource mapping exists
	mapEntry, err := mapRepo.GetMapping(instA.ID)
	require.NoError(t, err)
	assert.Equal(t, instA.ID, mapEntry.AnarvaResourceID)
	assert.Equal(t, instA.ProviderInstanceID, mapEntry.ProviderResourceID)

	// 4. SIMULATE GATEWAY RESTART:
	// Instantiate brand new provider and usecase sharing the SAME durable repo & mappingRepo
	prov2 := computeProvider.NewLocalDockerComputeProvider()
	uc2 := computeUseCase.NewComputeUseCase(cRepo, nil, prov2)
	uc2.SetMappingRepository(mapRepo)

	// 5. Tenant A GET recovered instance succeeds
	gotInst, err := uc2.GetInstanceForTenant(ctx, "org-tenant-a", "proj-tenant-a", instA.ID)
	require.NoError(t, err)
	assert.Equal(t, instA.ID, gotInst.ID)
	assert.Equal(t, "Worker 70B", gotInst.Name)

	// 6. Tenant A STOP recovered instance succeeds
	err = uc2.StopInstance(ctx, instA.ID)
	require.NoError(t, err)

	stoppedInst, err := uc2.GetInstanceForTenant(ctx, "org-tenant-a", "proj-tenant-a", instA.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusStopped, stoppedInst.Status)

	// 7. Tenant A START recovered instance succeeds
	err = uc2.StartInstance(ctx, instA.ID)
	require.NoError(t, err)

	startedInst, err := uc2.GetInstanceForTenant(ctx, "org-tenant-a", "proj-tenant-a", instA.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusRunning, startedInst.Status)

	// 8. Tenant A RESTART recovered instance succeeds
	err = uc2.RestartInstance(ctx, instA.ID)
	require.NoError(t, err)

	restartedInst, err := uc2.GetInstanceForTenant(ctx, "org-tenant-a", "proj-tenant-a", instA.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusRunning, restartedInst.Status)

	// 9. Tenant A EXECUTE command post-restart succeeds
	execRes, err := uc2.ExecuteCommand(ctx, instA.ID, &domain.CommandExecutionRequest{
		Command: "uname -a",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, execRes.ExitCode)

	// 10. Tenant B CANNOT access/recover Tenant A instance (Tenant Isolation Preserved)
	_, err = uc2.GetInstanceForTenant(ctx, "org-tenant-b", "proj-tenant-b", instA.ID)
	require.Error(t, err)

	// 11. Missing provider resource handling
	fakeMissingInst := &domain.ComputeInstance{
		ID:                 "acu-inst-missing",
		OrganizationID:     "org-tenant-a",
		ProjectID:          "proj-tenant-a",
		Name:               "Missing Worker",
		Provider:           domain.ProviderLocalDocker,
		ProviderInstanceID: "non-existent-container-xyz-9999",
	}
	err = prov2.RehydrateInstance(ctx, fakeMissingInst)
	if _, lookErr := os.Stat("/var/run/docker.sock"); lookErr == nil {
		// Real Docker host - non-existent container inspect returns error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no longer exists")
	}

	// 12. Durable metadata remains intact
	finalInst, err := cRepo.GetByID(ctx, instA.ID)
	require.NoError(t, err)
	assert.Equal(t, instA.ID, finalInst.ID)
	assert.Equal(t, instA.OrganizationID, finalInst.OrganizationID)
	assert.Equal(t, instA.ProjectID, finalInst.ProjectID)
}

func TestPhase70B_PostgresComputeRepository_LiveDBRestartRecovery(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("Skipping live PostgresComputeRepository restart recovery test (no TEST_DATABASE_URL / DATABASE_URL configured)")
		return
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping live test (failed to connect to PostgreSQL: %v)", err)
		return
	}

	err = db.AutoMigrate(&domain.ComputeInstance{}, &domain.Volume{}, &mapping.ProviderResourceMapping{})
	require.NoError(t, err)

	ctx := context.Background()
	cRepo := computeRepo.NewPostgresComputeRepository(db)
	mapRepo := mapping.NewPostgresMappingRepository(db)

	prov1 := computeProvider.NewLocalDockerComputeProvider()
	uc1 := computeUseCase.NewComputeUseCase(cRepo, nil, prov1)
	uc1.SetMappingRepository(mapRepo)

	instID := "acu-inst-livedb-70b"
	_ = uc1.DeleteInstance(ctx, instID)

	// Create
	created, err := uc1.CreateInstance(ctx, &domain.ComputeInstance{
		ID:             instID,
		OrganizationID: "org-live-70b",
		ProjectID:      "proj-live-70b",
		Name:           "Live DB Worker 70B",
		ACU:            1.0,
		RegionID:       "us-east-1",
	})
	require.NoError(t, err)
	assert.Equal(t, instID, created.ID)

	// Simulate restart
	prov2 := computeProvider.NewLocalDockerComputeProvider()
	uc2 := computeUseCase.NewComputeUseCase(cRepo, nil, prov2)
	uc2.SetMappingRepository(mapRepo)

	// Recover via GetInstanceForTenant
	recovered, err := uc2.GetInstanceForTenant(ctx, "org-live-70b", "proj-live-70b", instID)
	require.NoError(t, err)
	assert.Equal(t, instID, recovered.ID)
	assert.Equal(t, "Live DB Worker 70B", recovered.Name)

	// Cleanup
	_ = uc2.DeleteInstance(ctx, instID)
}
