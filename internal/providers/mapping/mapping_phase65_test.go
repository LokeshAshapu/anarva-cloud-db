package mapping_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
)

func TestInMemoryMappingRepository_Comprehensive(t *testing.T) {
	ctx := context.Background()
	repo := mapping.NewInMemoryMappingRepository()

	// 1. Create mapping
	m1 := &mapping.ProviderResourceMapping{
		AnarvaResourceID:     "arnv:vm:us-east-1:proj-101:compute/server-01",
		OrganizationID:       "org-101",
		ProjectID:            "proj-101",
		Provider:             "AWS",
		ProviderResourceID:   "i-0a1b2c3d4e5f67890",
		ProviderResourceType: "ec2_instance",
		Region:               "us-east-1",
		Zone:                 "us-east-1a",
		Status:               "RUNNING",
		Managed:              true,
	}

	err := repo.SaveMapping(m1)
	require.NoError(t, err)

	// 2. Get mapping
	got, err := repo.GetMapping(m1.AnarvaResourceID)
	require.NoError(t, err)
	assert.Equal(t, m1.AnarvaResourceID, got.AnarvaResourceID)
	assert.Equal(t, m1.OrganizationID, got.OrganizationID)
	assert.Equal(t, m1.ProviderResourceID, got.ProviderResourceID)

	// 3. Find by provider resource ID
	found, err := repo.FindByProviderResourceID(ctx, "AWS", "i-0a1b2c3d4e5f67890")
	require.NoError(t, err)
	assert.Equal(t, m1.AnarvaResourceID, found.AnarvaResourceID)

	// 4. Tenant isolation verification
	t.Run("Tenant Scoped Match Passes", func(t *testing.T) {
		scoped, err := repo.GetTenantScopedMapping(ctx, "org-101", "proj-101", m1.AnarvaResourceID)
		require.NoError(t, err)
		assert.Equal(t, m1.AnarvaResourceID, scoped.AnarvaResourceID)
	})

	t.Run("Cross-Organization Access Rejected", func(t *testing.T) {
		_, err := repo.GetTenantScopedMapping(ctx, "org-MALICIOUS", "proj-101", m1.AnarvaResourceID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TENANT_ISOLATION_VIOLATION")
	})

	t.Run("Cross-Project Access Rejected", func(t *testing.T) {
		_, err := repo.GetTenantScopedMapping(ctx, "org-101", "proj-MALICIOUS", m1.AnarvaResourceID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TENANT_ISOLATION_VIOLATION")
	})

	// 5. Delete mapping
	err = repo.DeleteMapping(ctx, m1.AnarvaResourceID)
	require.NoError(t, err)

	_, err = repo.GetMapping(m1.AnarvaResourceID)
	require.Error(t, err)
}

func TestPostgresMappingRepository_LiveDBOrSkip(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("Skipping live PostgresMappingRepository tests (no TEST_DATABASE_URL / DATABASE_URL configured)")
		return
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping live PostgresMappingRepository tests (failed to connect to PostgreSQL: %v)", err)
		return
	}

	// AutoMigrate table
	err = db.AutoMigrate(&mapping.ProviderResourceMapping{})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("1. CRUD Operations", func(t *testing.T) {
		repo := mapping.NewPostgresMappingRepository(db)

		m := &mapping.ProviderResourceMapping{
			AnarvaResourceID:     "arnv:db:us-west-2:proj-999:postgres/db-prod-01",
			OrganizationID:       "org-999",
			ProjectID:            "proj-999",
			Provider:             "AWS",
			ProviderResourceID:   "rds-prod-db-999",
			ProviderResourceType: "rds_postgres",
			Region:               "us-west-2",
			Zone:                 "us-west-2b",
			Status:               "AVAILABLE",
			Managed:              true,
		}

		// Clean up prior test runs
		_ = repo.DeleteMapping(ctx, m.AnarvaResourceID)

		// Create
		err := repo.SaveMapping(m)
		require.NoError(t, err)

		// Read
		got, err := repo.GetMapping(m.AnarvaResourceID)
		require.NoError(t, err)
		assert.Equal(t, m.AnarvaResourceID, got.AnarvaResourceID)
		assert.Equal(t, m.ProviderResourceID, got.ProviderResourceID)

		// Find by Provider Resource ID
		found, err := repo.FindByProviderResourceID(ctx, "AWS", "rds-prod-db-999")
		require.NoError(t, err)
		assert.Equal(t, m.AnarvaResourceID, found.AnarvaResourceID)

		// Tenant Isolation Validation
		_, err = repo.GetTenantScopedMapping(ctx, "org-ATTACKER", "proj-999", m.AnarvaResourceID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TENANT_ISOLATION_VIOLATION")

		// Soft Delete
		err = repo.DeleteMapping(ctx, m.AnarvaResourceID)
		require.NoError(t, err)

		_, err = repo.GetMapping(m.AnarvaResourceID)
		require.Error(t, err)
	})

	t.Run("2. STEP 9 Restart Persistence Boundary Verification", func(t *testing.T) {
		// A. Instantiate first repository instance and save mapping
		repo1 := mapping.NewPostgresMappingRepository(db)
		testMapping := &mapping.ProviderResourceMapping{
			AnarvaResourceID:     "arnv:vpc:eu-central-1:proj-restart:vpc/vpc-main",
			OrganizationID:       "org-restart",
			ProjectID:            "proj-restart",
			Provider:             "AWS",
			ProviderResourceID:   "vpc-0987654321fedcba0",
			ProviderResourceType: "vpc",
			Region:               "eu-central-1",
			Status:               "ACTIVE",
			Managed:              true,
		}
		_ = repo1.DeleteMapping(ctx, testMapping.AnarvaResourceID)

		err := repo1.SaveMapping(testMapping)
		require.NoError(t, err)

		// B. Simulate process restart by discarding repo1 instance reference
		repo1 = nil

		// C. Instantiate new repository instance connected to PostgreSQL database
		repo2 := mapping.NewPostgresMappingRepository(db)

		// D. Retrieve mapping from repo2 and verify survival across restart boundary
		retrieved, err := repo2.GetMapping(testMapping.AnarvaResourceID)
		require.NoError(t, err)
		assert.Equal(t, testMapping.AnarvaResourceID, retrieved.AnarvaResourceID)
		assert.Equal(t, testMapping.OrganizationID, retrieved.OrganizationID)
		assert.Equal(t, testMapping.ProviderResourceID, retrieved.ProviderResourceID)

		// Clean up
		_ = repo2.DeleteMapping(ctx, testMapping.AnarvaResourceID)
	})
}
