package compute_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeRepo "github.com/anarva-cloud/anarva-cloud-db/internal/compute/repository"
)

func TestComputeDomain_JSONSerializationHooks(t *testing.T) {
	inst := &domain.ComputeInstance{
		ID:             "acu-test-101",
		OrganizationID: "org-101",
		ProjectID:      "proj-101",
		Name:           "Worker Node Alpha",
		Security: domain.InstanceSecurityPolicy{
			SSHKeyIDs:       []string{"key-1", "key-2"},
			SecurityGroupIDs: []string{"sg-101"},
			IAMRole:          "role-worker",
		},
		EnvVars: map[string]string{
			"ENVIRONMENT": "production",
			"LOG_LEVEL":   "info",
		},
	}

	// Trigger GORM BeforeSave hook
	err := inst.BeforeSave(nil)
	require.NoError(t, err)
	assert.Contains(t, inst.SecurityJSON, "key-1")
	assert.Contains(t, inst.EnvVarsJSON, "production")

	// Create new struct, set JSON strings, trigger AfterFind hook
	restored := &domain.ComputeInstance{
		ID:           inst.ID,
		SecurityJSON: inst.SecurityJSON,
		EnvVarsJSON:  inst.EnvVarsJSON,
	}
	err = restored.AfterFind(nil)
	require.NoError(t, err)
	assert.Equal(t, inst.Security.SSHKeyIDs, restored.Security.SSHKeyIDs)
	assert.Equal(t, inst.EnvVars["ENVIRONMENT"], restored.EnvVars["ENVIRONMENT"])
}

func TestPostgresComputeRepository_LiveDBOrSkip(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("Skipping live PostgresComputeRepository tests (no TEST_DATABASE_URL / DATABASE_URL configured)")
		return
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping live PostgresComputeRepository tests (failed to connect to PostgreSQL: %v)", err)
		return
	}

	// AutoMigrate tables
	err = db.AutoMigrate(&domain.ComputeInstance{}, &domain.Volume{})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("1. ComputeInstance CRUD & Tenant Isolation", func(t *testing.T) {
		repo := computeRepo.NewPostgresComputeRepository(db)

		inst := &domain.ComputeInstance{
			ID:                 "acu-prod-worker-01",
			ResourceID:         "arnv:vm:us-east-1:proj-alpha:compute/worker-01",
			OrganizationID:     "org-alpha",
			ProjectID:          "proj-alpha",
			Name:               "Worker 01",
			Slug:               "worker-01",
			RegionID:           "us-east-1",
			ZoneID:             "us-east-1a",
			Status:             domain.StatusRunning,
			Health:             domain.HealthHealthy,
			PlanID:             "plan-1.0",
			ACU:                1.0,
			VCPU:               1.0,
			MemoryMB:           2048,
			StorageGB:          20,
			ImageID:            "img-ubuntu-24",
			Provider:           domain.ProviderLocalDocker,
			ProviderInstanceID: "docker-container-abc12345",
			Security: domain.InstanceSecurityPolicy{
				SSHKeyIDs: []string{"ssh-key-prod"},
			},
			EnvVars: map[string]string{"WORKER_ID": "01"},
		}

		// Clean prior test runs
		_ = repo.Delete(ctx, inst.ID)

		// Create
		err := repo.Create(ctx, inst)
		require.NoError(t, err)

		// Read
		got, err := repo.GetByID(ctx, inst.ID)
		require.NoError(t, err)
		assert.Equal(t, inst.ID, got.ID)
		assert.Equal(t, inst.Name, got.Name)
		assert.Equal(t, "docker-container-abc12345", got.ProviderInstanceID)
		assert.Equal(t, []string{"ssh-key-prod"}, got.Security.SSHKeyIDs)
		assert.Equal(t, "01", got.EnvVars["WORKER_ID"])

		// List by project
		list, err := repo.ListByProjectID(ctx, "proj-alpha")
		require.NoError(t, err)
		assert.NotEmpty(t, list)

		// Tenant Isolation Verification
		_, err = repo.GetTenantScopedByID(ctx, "org-ATTACKER", "proj-alpha", inst.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TENANT_ISOLATION_VIOLATION")

		// Update
		got.Status = domain.StatusStopped
		err = repo.Update(ctx, got)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, inst.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusStopped, updated.Status)

		// Delete
		err = repo.Delete(ctx, inst.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, inst.ID)
		require.Error(t, err)
	})

	t.Run("2. STEP 9 Restart Persistence Boundary Verification", func(t *testing.T) {
		// A. Instantiate first repository instance and save compute instance
		repo1 := computeRepo.NewPostgresComputeRepository(db)
		testInst := &domain.ComputeInstance{
			ID:                 "acu-restart-node-777",
			ResourceID:         "arnv:vm:us-west-2:proj-restart:compute/node-777",
			OrganizationID:     "org-restart",
			ProjectID:          "proj-restart",
			Name:               "Node 777 Persistence Test",
			Slug:               "node-777",
			RegionID:           "us-west-2",
			ZoneID:             "us-west-2a",
			Status:             domain.StatusRunning,
			Health:             domain.HealthHealthy,
			PlanID:             "plan-2.0",
			ACU:                2.0,
			VCPU:               2.0,
			MemoryMB:           4096,
			StorageGB:          40,
			ImageID:            "img-debian-12",
			Provider:           domain.ProviderAWS,
			ProviderInstanceID: "i-0987654321fedcba0",
		}
		_ = repo1.Delete(ctx, testInst.ID)

		err := repo1.Create(ctx, testInst)
		require.NoError(t, err)

		// B. Simulate gateway process restart by nullifying repo1 instance reference
		repo1 = nil

		// C. Instantiate new repository instance connected to PostgreSQL database
		repo2 := computeRepo.NewPostgresComputeRepository(db)

		// D. Retrieve compute instance from repo2 and verify survival across restart boundary
		retrieved, err := repo2.GetByID(ctx, testInst.ID)
		require.NoError(t, err)
		assert.Equal(t, testInst.ID, retrieved.ID)
		assert.Equal(t, testInst.OrganizationID, retrieved.OrganizationID)
		assert.Equal(t, testInst.ProviderInstanceID, retrieved.ProviderInstanceID)

		// Clean up
		_ = repo2.Delete(ctx, testInst.ID)
	})

	t.Run("3. Volume CRUD Operations", func(t *testing.T) {
		volRepo := computeRepo.NewPostgresVolumeRepository(db)

		vol := &domain.Volume{
			ID:               "vol-prod-ssd-01",
			OrganizationID:   "org-alpha",
			ProjectID:        "proj-alpha",
			InstanceID:       "acu-prod-worker-01",
			Name:             "NVMe Data Volume",
			SizeGB:           100,
			RegionID:         "us-east-1",
			ZoneID:           "us-east-1a",
			Type:             "NVME_SSD",
			ProviderVolumeID: "vol-aws-01234567",
			Status:           "ATTACHED",
		}

		_ = volRepo.DeleteVolume(ctx, vol.ID)

		err := volRepo.CreateVolume(ctx, vol)
		require.NoError(t, err)

		got, err := volRepo.GetVolumeByID(ctx, vol.ID)
		require.NoError(t, err)
		assert.Equal(t, vol.ID, got.ID)

		_, err = volRepo.GetTenantScopedVolumeByID(ctx, "org-ATTACKER", "proj-alpha", vol.ID)
		require.Error(t, err)

		_ = volRepo.DeleteVolume(ctx, vol.ID)
	})
}
