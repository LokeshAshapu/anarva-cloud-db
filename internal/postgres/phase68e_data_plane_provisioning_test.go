package postgres_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/domain"
	postgresHandler "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/handler"
	postgresProvider "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/provider"
	postgresService "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

// -----------------------------------------------------------------------------
// UNIT TESTS: Data-Plane Provisioning, Sanitization & Tenant Isolation
// -----------------------------------------------------------------------------

func TestPhase68E_IdentifierSanitizationAndSecurity(t *testing.T) {
	t.Run("Valid Alphanumeric Identifiers Succeed", func(t *testing.T) {
		clean, err := postgresProvider.SanitizeIdentifier("db_proj_prod_01")
		require.NoError(t, err)
		assert.Equal(t, "db_proj_prod_01", clean)
	})

	t.Run("Unsafe SQL Characters Rejected", func(t *testing.T) {
		unsafeList := []string{
			"db_name; DROP TABLE users; --",
			"db_name' OR 1=1--",
			"db name with spaces",
			"db-name-with-dashes",
			"db_name/*comment*/",
		}
		for _, unsafeIdent := range unsafeList {
			_, err := postgresProvider.SanitizeIdentifier(unsafeIdent)
			assert.Error(t, err, "Should reject unsafe identifier: %s", unsafeIdent)
		}
	})

	t.Run("Cryptographically Strong Password Generation", func(t *testing.T) {
		pass1, err1 := postgresProvider.GenerateStrongPassword()
		require.NoError(t, err1)
		assert.Len(t, pass1, 24)

		pass2, err2 := postgresProvider.GenerateStrongPassword()
		require.NoError(t, err2)
		assert.NotEqual(t, pass1, pass2, "Generated passwords must be unique and non-deterministic")
	})
}

func TestPhase68E_SimulatedProvisioningAndTenantIsolation(t *testing.T) {
	dpProv := postgresProvider.NewSimulatedDataPlaneProvider()
	pgProv := postgresProvider.NewLocalDockerPostgresProvider()
	svc := postgresService.NewPostgresServiceFull(nil, pgProv, dpProv)
	sqlSvc := postgresService.NewSQLService()

	ctx := context.Background()

	t.Run("Provision Database for Tenant A", func(t *testing.T) {
		inst, err := svc.CreateInstance(ctx, "org-tenant-a", "proj-tenant-a", "Tenant A Database", "17", "ap-hyderabad-1", "vpc-a", 1.0, 1024, 10, false)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusAvailable, inst.Status)
		assert.Contains(t, inst.ProviderResourceId, "db_")
		assert.Equal(t, "LOCAL_POSTGRES (SIMULATED_PROVISIONING)", inst.RealityLabel)

		// Verify API response serialization contains NO secret fields
		data, err := json.Marshal(inst)
		require.NoError(t, err)
		assert.NotContains(t, string(data), "password")
		assert.NotContains(t, string(data), "secret")
	})

	t.Run("Tenant Isolation on Subroute Operations", func(t *testing.T) {
		instA, err := svc.CreateInstance(ctx, "org-tenant-a", "proj-tenant-a", "Database Alpha", "17", "ap-hyderabad-1", "vpc-a", 1.0, 1024, 10, false)
		require.NoError(t, err)

		handler := postgresHandler.NewPostgresHandler(svc, sqlSvc)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux)

		// Tenant B attempts to delete Tenant A database
		req := httptest.NewRequest("DELETE", "/api/v1/databases/"+instA.ID, nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-ATTACKER")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-ATTACKER")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")
	})
}

// -----------------------------------------------------------------------------
// REAL INTEGRATION TEST: Executes only when TEST_CUSTOMER_DATABASE_ADMIN_URL is set
// -----------------------------------------------------------------------------

func TestPhase68E_RealPostgresDataPlaneProvisioning_Integration(t *testing.T) {
	adminDSN := os.Getenv("TEST_CUSTOMER_DATABASE_ADMIN_URL")
	if adminDSN == "" {
		adminDSN = os.Getenv("CUSTOMER_DATABASE_ADMIN_URL")
	}
	if adminDSN == "" {
		t.Skip("Skipping Real PostgreSQL Data-Plane Provisioning Integration Test (no TEST_CUSTOMER_DATABASE_ADMIN_URL or CUSTOMER_DATABASE_ADMIN_URL configured)")
		return
	}

	dpProv := postgresProvider.NewRealPostgresDataPlaneProvider(adminDSN)
	pgProv := postgresProvider.NewLocalDockerPostgresProvider()
	svc := postgresService.NewPostgresServiceFull(nil, pgProv, dpProv)

	ctx := context.Background()

	t.Run("Real Provisioning, Validation & Deletion Lifecycle", func(t *testing.T) {
		inst := domain.NewPostgresInstance("org-integration", "proj-integration", "Real Integration Test DB", "17", "ap-hyderabad-1", "vpc-integration", 1.0, 1024, 10)

		// 1. Create Instance on Real PostgreSQL Data-Plane Server
		created, err := svc.CreateInstance(ctx, inst.OrganizationID, inst.ProjectID, inst.Name, inst.Version, inst.RegionID, inst.NetworkID, inst.CPU, inst.MemoryMB, inst.StorageGB, false)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusAvailable, created.Status)
		assert.Contains(t, created.RealityLabel, "REAL_POSTGRES")

		// 2. Verify Database & Role Existence on Real PostgreSQL Server via Admin DSN
		adminDB, err := sql.Open("pgx", adminDSN)
		require.NoError(t, err)
		defer adminDB.Close()

		rawSuffix := strings.ReplaceAll(created.ID, "-", "_")
		rawSuffix = strings.ReplaceAll(rawSuffix, ":", "_")
		expectedDBName := fmt.Sprintf("db_%s", rawSuffix)
		expectedRoleName := fmt.Sprintf("usr_%s", rawSuffix)

		var dbCount int
		err = adminDB.QueryRowContext(ctx, "SELECT count(*) FROM pg_database WHERE datname = $1", expectedDBName).Scan(&dbCount)
		require.NoError(t, err)
		assert.Equal(t, 1, dbCount, "Real PostgreSQL database 'db_%s' must exist", expectedDBName)

		var roleCount int
		err = adminDB.QueryRowContext(ctx, "SELECT count(*) FROM pg_roles WHERE rolname = $1", expectedRoleName).Scan(&roleCount)
		require.NoError(t, err)
		assert.Equal(t, 1, roleCount, "Real PostgreSQL role 'usr_%s' must exist", expectedRoleName)

		// 3. Delete Instance from Real PostgreSQL Data-Plane Server
		err = svc.DeleteInstance(ctx, created.ID)
		require.NoError(t, err)

		// 4. Verify Database & Role Cleaned Up
		err = adminDB.QueryRowContext(ctx, "SELECT count(*) FROM pg_database WHERE datname = $1", expectedDBName).Scan(&dbCount)
		require.NoError(t, err)
		assert.Equal(t, 0, dbCount, "Real PostgreSQL database 'db_%s' must be dropped after deletion", expectedDBName)

		err = adminDB.QueryRowContext(ctx, "SELECT count(*) FROM pg_roles WHERE rolname = $1", expectedRoleName).Scan(&roleCount)
		require.NoError(t, err)
		assert.Equal(t, 0, roleCount, "Real PostgreSQL role 'usr_%s' must be dropped after deletion", expectedRoleName)
	})
}
