package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
)

func TestSQLService_StatefulExecution(t *testing.T) {
	svc := service.NewSQLService()
	ctx := context.Background()
	instID := fmt.Sprintf("test-db-inst-%d", time.Now().UnixNano())

	t.Run("Duplicate CREATE TABLE returns relation already exists error", func(t *testing.T) {
		res, err := svc.ExecuteQuery(ctx, instID, "CREATE TABLE users (id INT, username TEXT)")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), `relation "users" already exists`)
	})

	t.Run("CREATE TABLE IF NOT EXISTS succeeds idempotently", func(t *testing.T) {
		res, err := svc.ExecuteQuery(ctx, instID, "CREATE TABLE IF NOT EXISTS users (id INT, username TEXT)")
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("CREATE TABLE custom_users with DEFAULT values and INSERT", func(t *testing.T) {
		createSQL := `CREATE TABLE custom_users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(150) UNIQUE NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
		resCreate, err := svc.ExecuteQuery(ctx, instID, createSQL)
		require.NoError(t, err)
		assert.Equal(t, []string{"id", "name", "email", "is_active", "created_at"}, resCreate.Columns)

		insertSQL := `INSERT INTO custom_users (name, email) VALUES
			('Alice Johnson', 'alice@example.com'),
			('Bob Smith', 'bob@example.com'),
			('Charlie Brown', 'charlie@example.com')`
		resInsert, err := svc.ExecuteQuery(ctx, instID, insertSQL)
		require.NoError(t, err)
		assert.Equal(t, 3, resInsert.RowCount)

		// Verify is_active is boolean true instead of null
		assert.Equal(t, true, resInsert.Rows[0][3])
		assert.Equal(t, "Alice Johnson", resInsert.Rows[0][1])
		assert.Equal(t, "alice@example.com", resInsert.Rows[0][2])
	})

	t.Run("ALTER TABLE ADD COLUMN and DROP COLUMN", func(t *testing.T) {
		alterInstID := fmt.Sprintf("alter-inst-%d", time.Now().UnixNano())
		_, err := svc.ExecuteQuery(ctx, alterInstID, "CREATE TABLE employees (id INT, name TEXT)")
		require.NoError(t, err)

		_, err = svc.ExecuteQuery(ctx, alterInstID, "INSERT INTO employees (id, name) VALUES (1, 'Eve')")
		require.NoError(t, err)

		resAdd, err := svc.ExecuteQuery(ctx, alterInstID, "ALTER TABLE employees ADD COLUMN department TEXT DEFAULT 'Engineering'")
		require.NoError(t, err)
		assert.Equal(t, []string{"id", "name", "department"}, resAdd.Columns)
		assert.Equal(t, "Engineering", resAdd.Rows[0][2])

		resDrop, err := svc.ExecuteQuery(ctx, alterInstID, "ALTER TABLE employees DROP COLUMN department")
		require.NoError(t, err)
		assert.Equal(t, []string{"id", "name"}, resDrop.Columns)
	})

	t.Run("TRUNCATE TABLE clears rows without dropping schema", func(t *testing.T) {
		truncInstID := fmt.Sprintf("trunc-inst-%d", time.Now().UnixNano())
		_, err := svc.ExecuteQuery(ctx, truncInstID, "CREATE TABLE logs (id INT, message TEXT)")
		require.NoError(t, err)

		_, err = svc.ExecuteQuery(ctx, truncInstID, "INSERT INTO logs (id, message) VALUES (1, 'Error 500')")
		require.NoError(t, err)

		resTrunc, err := svc.ExecuteQuery(ctx, truncInstID, "TRUNCATE TABLE logs")
		require.NoError(t, err)
		assert.Equal(t, 1, resTrunc.RowCount)

		resSel, err := svc.ExecuteQuery(ctx, truncInstID, "SELECT * FROM logs")
		require.NoError(t, err)
		assert.Equal(t, 0, resSel.RowCount)
		assert.Equal(t, []string{"id", "message"}, resSel.Columns)
	})

	t.Run("BEGIN, COMMIT, ROLLBACK statements", func(t *testing.T) {
		resBegin, err := svc.ExecuteQuery(ctx, instID, "BEGIN")
		require.NoError(t, err)
		assert.Equal(t, "BEGIN", resBegin.Rows[0][0])

		resCommit, err := svc.ExecuteQuery(ctx, instID, "COMMIT")
		require.NoError(t, err)
		assert.Equal(t, "COMMIT", resCommit.Rows[0][0])
	})

	t.Run("SELECT VERSION and DESCRIBE statements", func(t *testing.T) {
		resVer, err := svc.ExecuteQuery(ctx, instID, "SELECT VERSION()")
		require.NoError(t, err)
		assert.Contains(t, resVer.Rows[0][0].(string), "PostgreSQL 17.2")

		resDesc, err := svc.ExecuteQuery(ctx, instID, "DESCRIBE users")
		require.NoError(t, err)
		assert.Equal(t, []string{"column_name", "data_type", "default_value"}, resDesc.Columns)
	})

	t.Run("SELECT * FROM nonexistent table returns relation does not exist error", func(t *testing.T) {
		res, err := svc.ExecuteQuery(ctx, instID, "SELECT * FROM non_existent_table")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), `relation "non_existent_table" does not exist`)
	})

	t.Run("Server restart simulation retains created tables and rows permanently", func(t *testing.T) {
		restartInstID := fmt.Sprintf("persistent-restart-db-%d", time.Now().UnixNano())

		svc1 := service.NewSQLService()
		_, err := svc1.ExecuteQuery(ctx, restartInstID, "CREATE TABLE persistent_products (id INT, product_name TEXT, is_in_stock BOOLEAN DEFAULT true)")
		require.NoError(t, err)

		_, err = svc1.ExecuteQuery(ctx, restartInstID, "INSERT INTO persistent_products (id, product_name) VALUES (501, 'MacBook Pro M4')")
		require.NoError(t, err)

		svc2 := service.NewSQLService()
		resSelect, err := svc2.ExecuteQuery(ctx, restartInstID, "SELECT * FROM persistent_products")
		require.NoError(t, err)
		require.Equal(t, 1, resSelect.RowCount)
		assert.Equal(t, []string{"id", "product_name", "is_in_stock"}, resSelect.Columns)
		assert.Equal(t, 501, resSelect.Rows[0][0])
		assert.Equal(t, "MacBook Pro M4", resSelect.Rows[0][1])
		assert.Equal(t, true, resSelect.Rows[0][2])
	})
}
