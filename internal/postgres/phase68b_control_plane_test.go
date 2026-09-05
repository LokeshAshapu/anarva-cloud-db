package postgres_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/domain"
	postgresHandler "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/handler"
	postgresProvider "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/provider"
	postgresRepo "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/repository"
	postgresService "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

func TestPostgresInstanceRepository_LiveDBOrSkip(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("Skipping live PostgresInstanceRepository tests (no TEST_DATABASE_URL / DATABASE_URL configured)")
		return
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping live PostgresInstanceRepository tests (failed to connect to PostgreSQL: %v)", err)
		return
	}

	// AutoMigrate database instance tables
	err = db.AutoMigrate(&domain.PostgresInstance{})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("1. Create, Read, Update, Delete & Restart Persistence Boundary", func(t *testing.T) {
		repo1 := postgresRepo.NewGormPostgresInstanceRepository(db)

		inst := &domain.PostgresInstance{
			ID:             "postgresql-prod-alpha-101",
			OrganizationID: "org-alpha",
			ProjectID:      "proj-alpha",
			Name:           "Production DB Alpha",
			Provider:       "LOCAL_POSTGRES",
			Version:        "17",
			Status:         domain.StatusAvailable,
			RegionID:       "ap-hyderabad-1",
			ZoneId:         "ap-hyderabad-1a",
			CPU:            2.0,
			MemoryMB:       2048,
			StorageGB:      25,
			StorageType:    "SSD",
			NetworkID:      "vpc-alpha",
			Host:           "localhost",
			Port:           15433,
			PublicAccess:   false,
			RealityLabel:   "LOCAL_POSTGRES (STATEFUL_STORAGE)",
		}

		_ = repo1.Delete(ctx, inst.ID)

		// Create
		err := repo1.Create(ctx, inst)
		require.NoError(t, err)

		// Restart Persistence Boundary Simulation (nullify repo1 reference)
		repo1 = nil

		// Create fresh repository instance connected to PostgreSQL
		repo2 := postgresRepo.NewGormPostgresInstanceRepository(db)

		// Retrieve instance and verify metadata survival across restart
		retrieved, err := repo2.GetByID(ctx, inst.ID)
		require.NoError(t, err)
		assert.Equal(t, inst.ID, retrieved.ID)
		assert.Equal(t, inst.OrganizationID, retrieved.OrganizationID)
		assert.Equal(t, inst.Name, retrieved.Name)
		assert.Equal(t, 15433, retrieved.Port)

		// Tenant Isolation Verification
		_, err = repo2.GetByIDForTenant(ctx, "org-ATTACKER", "proj-alpha", inst.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TENANT_ISOLATION_VIOLATION")

		// Cross-Project Verification
		_, err = repo2.GetByIDForTenant(ctx, "org-alpha", "proj-OTHER", inst.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TENANT_ISOLATION_VIOLATION")

		// Soft Delete Verification
		err = repo2.Delete(ctx, inst.ID)
		require.NoError(t, err)

		_, err = repo2.GetByID(ctx, inst.ID)
		require.Error(t, err)
	})
}

func TestPostgresQueryEndpoint_TenantAuthorization(t *testing.T) {
	adminDSN := os.Getenv("CUSTOMER_DATABASE_ADMIN_URL")
	require.NotEmpty(t, adminDSN, "CUSTOMER_DATABASE_ADMIN_URL must be configured for this integration test")

	pgProv := postgresProvider.NewLocalDockerPostgresProvider()
	dpProv := postgresProvider.NewRealPostgresDataPlaneProvider(adminDSN)
	instanceRepo := newTestPostgresInstanceRepository()
	userRepo := &testPostgresUserRepository{}
	encryptionKey := []byte("01234567890123456789012345678901")

	svc := postgresService.NewPostgresServiceFull(
		instanceRepo,
		userRepo,
		pgProv,
		dpProv,
		encryptionKey,
	)
	sqlSvc := postgresService.NewSQLService()
	ctx := context.Background()

	// Create database instance for Tenant A (org-tenant-a / proj-tenant-a)
	instA, err := svc.CreateInstance(ctx, "org-tenant-a", "proj-tenant-a", "Tenant A Database", "17", "ap-hyderabad-1", "vpc-a", 1.0, 1024, 10, false)
	require.NoError(t, err)

	handler := postgresHandler.NewPostgresHandler(svc, sqlSvc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Tenant A Querying Tenant A Database -> Authorized (200 OK)", func(t *testing.T) {
		reqBody := []byte(`{"sql": "SELECT 1"}`)
		req := httptest.NewRequest("POST", "/api/v1/databases/"+instA.ID+"/query", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Context with Tenant A credentials
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-a")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-a")
		reqCtx = context.WithValue(reqCtx, security.UserIDKey, "usr-tenant-a")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, "response body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "data")
	})

	t.Run("Tenant B Querying Tenant A Database -> Rejected (403 Forbidden)", func(t *testing.T) {
		reqBody := []byte(`{"sql": "SELECT * FROM users"}`)
		req := httptest.NewRequest("POST", "/api/v1/databases/"+instA.ID+"/query", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Context with Tenant B (ATTACKER) credentials
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-ATTACKER")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-ATTACKER")
		reqCtx = context.WithValue(reqCtx, security.UserIDKey, "usr-attacker")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")

		var res struct {
			Code string `json:"code"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		assert.Equal(t, "TENANT_ISOLATION_VIOLATION", res.Code)
	})
}

type testPostgresUserRepository struct {
	user *domain.PostgresUser
}
type testPostgresInstanceRepository struct {
	instances map[string]*domain.PostgresInstance
}

func newTestPostgresInstanceRepository() *testPostgresInstanceRepository {
	return &testPostgresInstanceRepository{
		instances: make(map[string]*domain.PostgresInstance),
	}
}

func (r *testPostgresInstanceRepository) Create(
	ctx context.Context,
	inst *domain.PostgresInstance,
) error {
	if inst == nil || inst.ID == "" {
		return errors.New("invalid postgres instance")
	}
	r.instances[inst.ID] = inst
	return nil
}

func (r *testPostgresInstanceRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.PostgresInstance, error) {
	inst, ok := r.instances[id]
	if !ok {
		return nil, errors.New("postgres instance not found")
	}
	return inst, nil
}

func (r *testPostgresInstanceRepository) GetByIDForTenant(
	ctx context.Context,
	orgID, projID, id string,
) (*domain.PostgresInstance, error) {
	inst, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if orgID != "" && inst.OrganizationID != orgID {
		return nil, errors.New("TENANT_ISOLATION_VIOLATION")
	}

	if projID != "" && inst.ProjectID != projID {
		return nil, errors.New("TENANT_ISOLATION_VIOLATION")
	}

	return inst, nil
}

func (r *testPostgresInstanceRepository) ListByProject(
	ctx context.Context,
	orgID, projectID string,
) ([]*domain.PostgresInstance, error) {
	var result []*domain.PostgresInstance

	for _, inst := range r.instances {
		if inst.OrganizationID == orgID && inst.ProjectID == projectID {
			result = append(result, inst)
		}
	}

	return result, nil
}

func (r *testPostgresInstanceRepository) ListByOrganization(
	ctx context.Context,
	orgID string,
) ([]*domain.PostgresInstance, error) {
	var result []*domain.PostgresInstance

	for _, inst := range r.instances {
		if inst.OrganizationID == orgID {
			result = append(result, inst)
		}
	}

	return result, nil
}

func (r *testPostgresInstanceRepository) Update(
	ctx context.Context,
	inst *domain.PostgresInstance,
) error {
	if inst == nil || inst.ID == "" {
		return errors.New("invalid postgres instance")
	}

	r.instances[inst.ID] = inst
	return nil
}

func (r *testPostgresInstanceRepository) Delete(
	ctx context.Context,
	id string,
) error {
	delete(r.instances, id)
	return nil
}
func (r *testPostgresUserRepository) Create(ctx context.Context, user *domain.PostgresUser) error {
	r.user = user
	return nil
}

func (r *testPostgresUserRepository) GetByInstanceAndUsername(
	ctx context.Context,
	instanceID, username string,
) (*domain.PostgresUser, error) {
	if r.user == nil {
		return nil, errors.New("postgres user not found")
	}

	if r.user.InstanceID != instanceID || r.user.Username != username {
		return nil, errors.New("postgres user not found")
	}

	return r.user, nil
}

func (r *testPostgresUserRepository) Delete(
	ctx context.Context,
	instanceID, username string,
) error {
	return nil
}

func (r *testPostgresUserRepository) Update(
	ctx context.Context,
	user *domain.PostgresUser,
) error {
	r.user = user
	return nil
}
