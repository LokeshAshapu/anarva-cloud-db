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
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresHandler "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/handler"
	postgresProvider "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/provider"
	postgresService "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

// -----------------------------------------------------------------------------
// UNIT TESTS: Query Authorization, Tenant Isolation & Error Sanitization
// -----------------------------------------------------------------------------

func TestPhase68F_SQLErrorSanitizationAndCredentialProtection(t *testing.T) {
	t.Run("SanitizeSQLError Strips Sensitive DSN and Passwords", func(t *testing.T) {
		rawErr := errors.New("failed to connect: postgres://admin:supersecretpassword123@db.host.com:5432/mydb?password=supersecretpassword123")
		sanitized := postgresService.SanitizeSQLError(rawErr)
		require.NotNil(t, sanitized)

		msg := sanitized.Error()
		assert.NotContains(t, msg, "supersecretpassword123")
		assert.Contains(t, msg, "postgres://***:***@")
		assert.Contains(t, msg, "password=***")
	})

	t.Run("Query Result Serialization Contains No Secret Data", func(t *testing.T) {
		res := &postgresService.SQLQueryResult{
			Columns:   []string{"id", "name"},
			Rows:      [][]interface{}{{1, "ANARVA_DATA"}},
			RowCount:  1,
			LatencyMs: 1.25,
			Truncated: false,
		}

		data, err := json.Marshal(res)
		require.NoError(t, err)
		assert.NotContains(t, string(data), "password")
		assert.NotContains(t, string(data), "secret")
		assert.NotContains(t, string(data), "dsn")
	})
}

func TestPhase68F_TenantAuthorizationOrderAndIsolation(t *testing.T) {
	adminDSN := os.Getenv("CUSTOMER_DATABASE_ADMIN_URL")
	require.NotEmpty(t, adminDSN, "CUSTOMER_DATABASE_ADMIN_URL must be configured for this integration test")

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
	exec := postgresService.NewPostgresSQLExecutor(sqlSvc)

	ctx := context.Background()

	// Provision Database for Tenant A (org-tenant-a / proj-tenant-a)
	instA, err := svc.CreateInstance(ctx, "org-tenant-a", "proj-tenant-a", "Tenant A DB", "17", "ap-hyderabad-1", "vpc-a", 1.0, 1024, 10, false)
	require.NoError(t, err)

	handler := postgresHandler.NewPostgresHandlerFull(svc, sqlSvc, exec, "")
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Tenant A Querying Tenant A Database -> Authorized (200 OK)", func(t *testing.T) {
		reqBody := []byte(`{"sql": "CREATE TABLE test_users (id INT, name TEXT);"}`)
		req := httptest.NewRequest("POST", "/api/v1/databases/"+instA.ID+"/query", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-a")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-a")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "data")
	})

	t.Run("Tenant B Querying Tenant A Database -> Rejected (403 Forbidden)", func(t *testing.T) {
		reqBody := []byte(`{"sql": "SELECT * FROM test_users;"}`)
		req := httptest.NewRequest("POST", "/api/v1/databases/"+instA.ID+"/query", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// Attacker Tenant B credentials
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-ATTACKER")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-ATTACKER")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")

		// Verify zero database data or credentials leaked in forbidden error body
		assert.NotContains(t, rec.Body.String(), "test_users")
		assert.NotContains(t, rec.Body.String(), "password")
	})
}

func TestPhase68F_ExecutorRestartPersistenceBoundary(t *testing.T) {
	dpProv := postgresProvider.NewSimulatedDataPlaneProvider()
	pgProv := postgresProvider.NewLocalDockerPostgresProvider()
	svc := postgresService.NewPostgresServiceFull(nil, nil, pgProv, dpProv, nil)
	sqlSvc := postgresService.NewSQLService()

	ctx := context.Background()

	inst, err := svc.CreateInstance(ctx, "org-dev", "proj-dev", "Persistence Test DB", "17", "ap-hyderabad-1", "vpc-dev", 1.0, 1024, 10, false)
	require.NoError(t, err)

	// Step 1: Create executor 1 and execute DDL + DML
	exec1 := postgresService.NewPostgresSQLExecutor(sqlSvc)
	_, err = exec1.Execute(ctx, inst, "", "CREATE TABLE items (id INT, title TEXT);")
	require.NoError(t, err)

	_, err = exec1.Execute(ctx, inst, "", "INSERT INTO items VALUES (101, 'Durable Item');")
	require.NoError(t, err)

	// Step 2: Destroy executor 1 (Simulate Gateway Restart / Executor Recreation)
	exec1 = nil

	// Step 3: Instantiate fresh Executor 2
	exec2 := postgresService.NewPostgresSQLExecutor(sqlSvc)
	res, err := exec2.Execute(ctx, inst, "", "SELECT * FROM items;")
	require.NoError(t, err)
	assert.Equal(t, 1, res.RowCount)
}

// -----------------------------------------------------------------------------
// REAL INTEGRATION TEST: Executes native SQL on Real PostgreSQL Data-Plane Server
// -----------------------------------------------------------------------------

func TestPhase68F_RealPostgresSQLExecution_Integration(t *testing.T) {
	adminDSN := os.Getenv("TEST_CUSTOMER_DATABASE_ADMIN_URL")
	if adminDSN == "" {
		adminDSN = os.Getenv("CUSTOMER_DATABASE_ADMIN_URL")
	}
	if adminDSN == "" {
		t.Skip("REAL POSTGRESQL INTEGRATION: NOT VERIFIED (no TEST_CUSTOMER_DATABASE_ADMIN_URL or CUSTOMER_DATABASE_ADMIN_URL configured)")
		return
	}

	dpProv := postgresProvider.NewRealPostgresDataPlaneProvider(adminDSN)
	pgProv := postgresProvider.NewLocalDockerPostgresProvider()
	svc := postgresService.NewPostgresServiceFull(nil, nil, pgProv, dpProv, nil)
	exec := postgresService.NewPostgresSQLExecutor(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Provision Real PostgreSQL DatabaseInstance
	inst, err := svc.CreateInstance(ctx, "org-real-sql", "proj-real-sql", "Real SQL Test DB", "17", "ap-hyderabad-1", "vpc-real", 1.0, 1024, 10, false)
	require.NoError(t, err)
	defer func() {
		_ = svc.DeleteInstance(ctx, inst.ID)
	}()

	t.Run("Real Native DDL & CRUD Execution Chain", func(t *testing.T) {
		// CREATE TABLE
		_, err := exec.Execute(ctx, inst, adminDSN, "CREATE TABLE test_items (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		require.NoError(t, err)

		// INSERT
		_, err = exec.Execute(ctx, inst, adminDSN, "INSERT INTO test_items (name) VALUES ('ANARVA_INITIAL');")
		require.NoError(t, err)

		// SELECT
		selectRes, err := exec.Execute(ctx, inst, adminDSN, "SELECT id, name FROM test_items;")
		require.NoError(t, err)
		assert.Equal(t, 1, selectRes.RowCount)
		assert.Equal(t, []string{"id", "name"}, selectRes.Columns)
		assert.Equal(t, "ANARVA_INITIAL", selectRes.Rows[0][1])

		// UPDATE
		_, err = exec.Execute(ctx, inst, adminDSN, "UPDATE test_items SET name = 'ANARVA_UPDATED' WHERE id = 1;")
		require.NoError(t, err)

		// SELECT AFTER UPDATE
		selectRes2, err := exec.Execute(ctx, inst, adminDSN, "SELECT name FROM test_items WHERE id = 1;")
		require.NoError(t, err)
		assert.Equal(t, "ANARVA_UPDATED", selectRes2.Rows[0][0])

		// DELETE
		_, err = exec.Execute(ctx, inst, adminDSN, "DELETE FROM test_items WHERE id = 1;")
		require.NoError(t, err)

		// SELECT AFTER DELETE
		selectRes3, err := exec.Execute(ctx, inst, adminDSN, "SELECT * FROM test_items;")
		require.NoError(t, err)
		assert.Equal(t, 0, selectRes3.RowCount)

		// DROP TABLE
		_, err = exec.Execute(ctx, inst, adminDSN, "DROP TABLE test_items;")
		require.NoError(t, err)
	})
}
