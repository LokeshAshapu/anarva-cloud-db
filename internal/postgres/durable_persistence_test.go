package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authDomain "github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
	authRepo "github.com/anarva-cloud/anarva-cloud-db/internal/auth/repository"
	databaseDomain "github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	databaseRepo "github.com/anarva-cloud/anarva-cloud-db/internal/database/repository"
	pgService "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
	projDomain "github.com/anarva-cloud/anarva-cloud-db/internal/project/domain"
	projRepo "github.com/anarva-cloud/anarva-cloud-db/internal/project/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	pkgDatabase "github.com/anarva-cloud/anarva-cloud-db/pkg/database"
)

func TestPhase55_ProductionDurablePersistence_IntegrationChain(t *testing.T) {
	// Skip if no real PostgreSQL DB configured in environment
	dbConfig := config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "anarva_admin",
		Password:        "anarva_secret_pass",
		DBName:          "anarva_db",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	dbPool, err := pkgDatabase.NewPostgresDB(dbConfig)
	if err != nil {
		t.Skipf("Skipping live PostgreSQL integration test (PostgreSQL pool connection error: %v)", err)
		return
	}
	defer dbPool.Close()

	ctx := context.Background()
	ts := time.Now().UnixNano()

	// 1. User Persistence Test across Process / Repo Recreation
	t.Run("1. User Persistence across Repo Recreation", func(t *testing.T) {
		uRepo1 := authRepo.NewUserRepository(dbPool.DB)
		userID := fmt.Sprintf("usr-p55-%d", ts)
		userEmail := fmt.Sprintf("user-p55-%d@anarva.io", ts)

		u1 := &authDomain.User{
			ID:        userID,
			Email:     userEmail,
			FullName:  "Phase 55 Persistence User",
			Role:      authDomain.RoleAdmin,
			Status:    authDomain.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, uRepo1.Create(ctx, u1))

		// Recreate Repository instance (simulating process restart)
		uRepo2 := authRepo.NewUserRepository(dbPool.DB)
		fetchedUser, errFetch := uRepo2.GetByID(ctx, userID)
		require.NoError(t, errFetch)
		assert.Equal(t, userID, fetchedUser.ID)
		assert.Equal(t, userEmail, fetchedUser.Email)
	})

	// 2. Complete Tenant Resource Chain Persistence Test
	t.Run("2. Complete Resource Chain Persistence (Org -> Proj -> DB -> Table -> Rows)", func(t *testing.T) {
		oRepo := projRepo.NewOrganizationRepository(dbPool.DB)
		pRepo := projRepo.NewProjectRepository(dbPool.DB)
		dRepo := databaseRepo.NewInstanceRepository(dbPool.DB)
		sqlSvc := pgService.NewSQLService()

		orgID := fmt.Sprintf("org-p55-%d", ts)
		projID := fmt.Sprintf("proj-p55-%d", ts)
		dbID := fmt.Sprintf("db-p55-%d", ts)

		// Create Org
		org := &projDomain.Organization{
			ID:        orgID,
			Name:      "Phase 55 Org",
			Slug:      orgID,
			OwnerID:   "usr-default",
			CreatedAt: time.Now(),
		}
		require.NoError(t, oRepo.Create(ctx, org))

		// Create Project
		proj := &projDomain.Project{
			ID:        projID,
			OrgID:     orgID,
			Name:      "Phase 55 Project",
			Slug:      projID,
			Region:    "us-east-1",
			CreatedAt: time.Now(),
		}
		require.NoError(t, pRepo.Create(ctx, proj))

		// Create Database Metadata
		inst := &databaseDomain.DatabaseInstance{
			ID:                dbID,
			ProjectID:         projID,
			Name:              "Production Core DB",
			Engine:            databaseDomain.EnginePostgreSQL,
			Status:            databaseDomain.StatusRunning,
			Host:              "localhost",
			Port:              5432,
			DBName:            "anarva_db",
			Username:          "anarva_admin",
			PasswordEncrypted: "enc",
			CreatedAt:         time.Now(),
		}
		require.NoError(t, dRepo.Create(ctx, inst))

		// Create Table & Insert Rows
		ctxTenant := context.WithValue(ctx, security.OrgIDKey, orgID)
		ctxTenant = context.WithValue(ctxTenant, security.ProjectIDKey, projID)

		_, errCreate := sqlSvc.ExecuteQuery(ctxTenant, dbID, "CREATE TABLE production_metrics (id INT, metric_name TEXT, metric_val FLOAT);")
		require.NoError(t, errCreate)

		_, errInsert := sqlSvc.ExecuteQuery(ctxTenant, dbID, "INSERT INTO production_metrics (id, metric_name, metric_val) VALUES (1, 'cpu_utilization', 42.5);")
		require.NoError(t, errInsert)

		// Query back
		res, errQuery := sqlSvc.ExecuteQuery(ctxTenant, dbID, "SELECT * FROM production_metrics WHERE id = 1;")
		require.NoError(t, errQuery)
		assert.Equal(t, 1, len(res.Rows))
	})
}
