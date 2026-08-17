package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	databaseDomain "github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	databaseDriver "github.com/anarva-cloud/anarva-cloud-db/internal/database/driver"
	databaseRepo "github.com/anarva-cloud/anarva-cloud-db/internal/database/repository"
	databaseUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/database/usecase"
	postgresService "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
)

func TestWorkflow1_DatabaseConnectionStringGeneration(t *testing.T) {
	ctx := context.Background()
	repo := databaseRepo.NewMemoryInstanceRepository()
	driver := databaseDriver.NewMockProvisionerDriver()
	encKey := "super-secret-encryption-key-32b"

	uc := databaseUsecase.NewDatabaseUseCase(repo, driver, encKey)

	// 1. Create database
	inst, connStr, err := uc.CreateDatabase(ctx, "proj-test-101", "production-db", databaseDomain.EnginePostgreSQL, 20)
	require.NoError(t, err)
	assert.NotEmpty(t, inst.ID)
	assert.Contains(t, connStr, "postgres://")
	assert.Contains(t, connStr, "sslmode=disable")

	// 2. Fetch connection string independently and verify matching state
	fetchedConnStr, errFetch := uc.GetConnectionString(ctx, inst.ID)
	require.NoError(t, errFetch)
	assert.Equal(t, connStr, fetchedConnStr)
}

func TestWorkflow2_PostgreSQLSQLConsoleStatefulExecution(t *testing.T) {
	ctx := context.Background()
	sqlSvc := postgresService.NewSQLService()

	instID := "pg-inst-wf2-101"

	// 1. Initial SELECT query returns stateful initial table rows
	res1, err1 := sqlSvc.ExecuteQuery(ctx, instID, "SELECT * FROM users LIMIT 10;")
	require.NoError(t, err1)
	assert.GreaterOrEqual(t, res1.RowCount, 1)
	assert.Contains(t, res1.Columns, "username")

	// 2. State Mutation via INSERT statement
	res2, err2 := sqlSvc.ExecuteQuery(ctx, instID, "INSERT INTO users (username, status) VALUES ('new_dev', 'ACTIVE');")
	require.NoError(t, err2)
	assert.Greater(t, res2.RowCount, res1.RowCount)

	// 3. Dangerous query rejection
	_, errDangerous := sqlSvc.ExecuteQuery(ctx, instID, "DROP DATABASE production;")
	assert.Error(t, errDangerous)
	assert.Contains(t, errDangerous.Error(), "dangerous")
}
