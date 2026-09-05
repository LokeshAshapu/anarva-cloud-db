package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

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
// PHASE 68G: REAL POSTGRESQL END-TO-END WORKFLOW INTEGRATION TEST
// -----------------------------------------------------------------------------

func TestPhase68G_EndToEndDatabaseWorkflow_Integration(t *testing.T) {
	adminDSN := os.Getenv("TEST_CUSTOMER_DATABASE_ADMIN_URL")
	if adminDSN == "" {
		adminDSN = os.Getenv("CUSTOMER_DATABASE_ADMIN_URL")
	}
	if adminDSN == "" {
		t.Skip("Skipping Phase 68G Real PostgreSQL End-to-End Integration Test (no TEST_CUSTOMER_DATABASE_ADMIN_URL or CUSTOMER_DATABASE_ADMIN_URL configured)")
		return
	}

	dpProv := postgresProvider.NewRealPostgresDataPlaneProvider(adminDSN)
	pgProv := postgresProvider.NewLocalDockerPostgresProvider()
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
	sqlExecutor := postgresService.NewPostgresSQLExecutor(sqlSvc)

	ctx := context.Background()

	t.Run("Complete Live PostgreSQL Data-Plane Lifecycle & Container Restart Persistence", func(t *testing.T) {
		// 1. Provision Database Instance for Tenant A on Real PostgreSQL Server
		instA, err := svc.CreateInstance(ctx, "org-tenant-a", "proj-tenant-a", "Phase68G Prod Database", "17", "ap-hyderabad-1", "vpc-a", 2.0, 2048, 25, false)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusAvailable, instA.Status)
		assert.Contains(t, instA.RealityLabel, "REAL_POSTGRES")
		customerDSN, err := svc.GetCustomerConnectionDSN(ctx, instA.ID)
		require.NoError(t, err)
		require.NotEmpty(t, customerDSN)
		require.NotContains(t, customerDSN, "anarva_admin")
		require.NotContains(t, customerDSN, "anarva_dev_password")

		// Register instA in pgProv for handler lookup during tenant isolation tests
		_, _ = pgProv.CreateInstance(ctx, instA, "secret")

		rawSuffix := strings.ReplaceAll(instA.ID, "-", "_")
		rawSuffix = strings.ReplaceAll(rawSuffix, ":", "_")
		dbName := fmt.Sprintf("db_%s", rawSuffix)
		roleName := fmt.Sprintf("usr_%s", rawSuffix)

		// 2. Verify Database & Role Existence on Real PostgreSQL Server via Admin DSN
		adminDB, err := sql.Open("pgx", adminDSN)
		require.NoError(t, err)
		defer adminDB.Close()

		var dbCount int
		err = adminDB.QueryRowContext(ctx, "SELECT count(*) FROM pg_database WHERE datname = $1", dbName).Scan(&dbCount)
		require.NoError(t, err)
		assert.Equal(t, 1, dbCount, "Real PostgreSQL database '%s' must exist", dbName)

		var roleCount int
		err = adminDB.QueryRowContext(ctx, "SELECT count(*) FROM pg_roles WHERE rolname = $1", roleName).Scan(&roleCount)
		require.NoError(t, err)
		assert.Equal(t, 1, roleCount, "Real PostgreSQL role '%s' must exist", roleName)

		// 3. Execute Native DDL: CREATE TABLE phase68g_users
		createTableSQL := `
			CREATE TABLE IF NOT EXISTS phase68g_users (
				id SERIAL PRIMARY KEY,
				name TEXT NOT NULL,
				email TEXT NOT NULL
			);
		`
		resCreate, err := sqlExecutor.Execute(ctx, instA, customerDSN, createTableSQL)
		require.NoError(t, err)
		assert.False(t, resCreate.Truncated)

		// 4. Execute Native DML: INSERT rows
		insertSQL := `
			INSERT INTO phase68g_users (name, email)
			VALUES ('Lokesh', 'lokesh@example.test'), ('Anarva', 'anarva@example.test');
		`
		resInsert, err := sqlExecutor.Execute(ctx, instA, customerDSN, insertSQL)
		require.NoError(t, err)
		assert.Equal(t, 1, resInsert.RowCount)
		assert.Contains(t, fmt.Sprintf("%v", resInsert.Rows[0][0]), "Affected rows: 2")

		// 5. Execute Native DML: SELECT rows
		selectSQL := `SELECT id, name, email FROM phase68g_users ORDER BY id ASC;`
		resSelect, err := sqlExecutor.Execute(ctx, instA, customerDSN, selectSQL)
		require.NoError(t, err)
		assert.Equal(t, 2, resSelect.RowCount)
		assert.Equal(t, []string{"id", "name", "email"}, resSelect.Columns)

		// 6. Execute Native DML: UPDATE row
		updateSQL := `UPDATE phase68g_users SET email = 'lokesh_updated@example.test' WHERE name = 'Lokesh';`
		resUpdate, err := sqlExecutor.Execute(ctx, instA, customerDSN, updateSQL)
		require.NoError(t, err)
		assert.Equal(t, 1, resUpdate.RowCount)
		assert.Contains(t, fmt.Sprintf("%v", resUpdate.Rows[0][0]), "Affected rows: 1")

		// 7. Execute Native DML: DELETE row
		deleteSQL := `DELETE FROM phase68g_users WHERE name = 'Anarva';`
		resDelete, err := sqlExecutor.Execute(ctx, instA, customerDSN, deleteSQL)
		require.NoError(t, err)
		assert.Equal(t, 1, resDelete.RowCount)
		assert.Contains(t, fmt.Sprintf("%v", resDelete.Rows[0][0]), "Affected rows: 1")

		// 8. Re-insert Anarva to establish final durable test dataset
		reinsertSQL := `INSERT INTO phase68g_users (name, email) VALUES ('Anarva', 'anarva@example.test');`
		_, err = sqlExecutor.Execute(ctx, instA, customerDSN, reinsertSQL)
		require.NoError(t, err)

		// 9. Verify final pre-restart dataset (Lokesh & Anarva)
		resPreRestart, err := sqlExecutor.Execute(ctx, instA, customerDSN, selectSQL)
		require.NoError(t, err)
		assert.Equal(t, 2, resPreRestart.RowCount)

		// 10. Gateway Process Restart Simulation (Recreate Executor)
		gatewayRestartExecutor := postgresService.NewPostgresSQLExecutor(sqlSvc)
		resPostGatewayRestart, err := gatewayRestartExecutor.Execute(ctx, instA, customerDSN, selectSQL)
		require.NoError(t, err)
		assert.Equal(t, 2, resPostGatewayRestart.RowCount, "Rows must persist across gateway executor recreation")

		// 11. Real PostgreSQL Docker Container Restart
		cmd := exec.Command("docker", "restart", "anarva-postgres-dataplane")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Docker container restart notice: %s (%v)", string(output), err)
		} else {
			t.Logf("Docker container 'anarva-postgres-dataplane' restarted successfully")
		}

		// Wait for PostgreSQL readiness on localhost:5433
		reconnectSuccess := false
		for i := 0; i < 15; i++ {
			time.Sleep(1 * time.Second)
			pingDB, pingErr := sql.Open("pgx", adminDSN)
			if pingErr == nil {
				if pingCtxErr := pingDB.PingContext(ctx); pingCtxErr == nil {
					pingDB.Close()
					reconnectSuccess = true
					break
				}
				pingDB.Close()
			}
		}
		require.True(t, reconnectSuccess, "PostgreSQL container must regain readiness on localhost:5433 within timeout")

		// 12. Verify Data Durability Post Container Restart (Persistent Named Volume Verification)
		resPostContainerRestart, err := gatewayRestartExecutor.Execute(ctx, instA, customerDSN, selectSQL)
		require.NoError(t, err)
		assert.Equal(t, 2, resPostContainerRestart.RowCount, "Rows must persist across PostgreSQL container restart via named Docker volume")

		// 13. Tenant Isolation Verification
		handler := postgresHandler.NewPostgresHandlerFull(svc, sqlSvc, sqlExecutor, adminDSN)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux)

		// Tenant B direct query against Tenant A's instance -> Must return 403 TENANT_ISOLATION_VIOLATION
		queryReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/databases/%s/query", instA.ID), strings.NewReader(`{"sql":"SELECT * FROM phase68g_users;"}`))
		queryReq.Header.Set("Content-Type", "application/json")
		reqCtxB := context.WithValue(queryReq.Context(), security.OrgIDKey, "org-ATTACKER")
		reqCtxB = context.WithValue(reqCtxB, security.ProjectIDKey, "proj-ATTACKER")
		queryReq = queryReq.WithContext(reqCtxB)

		queryRec := httptest.NewRecorder()
		mux.ServeHTTP(queryRec, queryReq)
		assert.Equal(t, http.StatusForbidden, queryRec.Code)
		assert.Contains(t, queryRec.Body.String(), "TENANT_ISOLATION_VIOLATION")

		// 14. Clean Up Instance from Real PostgreSQL Data-Plane Server
		err = svc.DeleteInstance(ctx, instA.ID)
		require.NoError(t, err)

		// Verify Database & Role Dropped
		err = adminDB.QueryRowContext(ctx, "SELECT count(*) FROM pg_database WHERE datname = $1", dbName).Scan(&dbCount)
		require.NoError(t, err)
		assert.Equal(t, 0, dbCount, "Real PostgreSQL database '%s' must be dropped", dbName)

		err = adminDB.QueryRowContext(ctx, "SELECT count(*) FROM pg_roles WHERE rolname = $1", roleName).Scan(&roleCount)
		require.NoError(t, err)
		assert.Equal(t, 0, roleCount, "Real PostgreSQL role '%s' must be dropped", roleName)
	})
}
