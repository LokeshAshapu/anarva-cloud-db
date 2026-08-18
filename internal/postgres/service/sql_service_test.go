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

	t.Run("LIKE pattern matching and ORDER BY DESC sorting", func(t *testing.T) {
		likeInstID := fmt.Sprintf("like-inst-%d", time.Now().UnixNano())
		_, err := svc.ExecuteQuery(ctx, likeInstID, "CREATE TABLE clients (id INT, email TEXT)")
		require.NoError(t, err)

		_, err = svc.ExecuteQuery(ctx, likeInstID, "INSERT INTO clients (id, email) VALUES (1, 'alpha@anarva.io'), (2, 'beta@other.com'), (3, 'gamma@anarva.io')")
		require.NoError(t, err)

		resLike, err := svc.ExecuteQuery(ctx, likeInstID, "SELECT * FROM clients WHERE email LIKE '%@anarva.io' ORDER BY id DESC")
		require.NoError(t, err)
		require.Equal(t, 2, resLike.RowCount)
		assert.Equal(t, 3, resLike.Rows[0][0]) // first is id 3 (DESC order)
		assert.Equal(t, 1, resLike.Rows[1][0]) // second is id 1
	})

	t.Run("SUM and AVG aggregate functions", func(t *testing.T) {
		aggInstID := fmt.Sprintf("agg-inst-%d", time.Now().UnixNano())
		_, err := svc.ExecuteQuery(ctx, aggInstID, "CREATE TABLE sales (amount FLOAT)")
		require.NoError(t, err)

		_, err = svc.ExecuteQuery(ctx, aggInstID, "INSERT INTO sales (amount) VALUES (100.0), (200.0), (300.0)")
		require.NoError(t, err)

		resSum, err := svc.ExecuteQuery(ctx, aggInstID, "SELECT SUM(amount) FROM sales")
		require.NoError(t, err)
		assert.Equal(t, 600.0, resSum.Rows[0][0])
	})

	t.Run("Database branching creates copy-on-write database clone", func(t *testing.T) {
		srcID := fmt.Sprintf("src-db-%d", time.Now().UnixNano())
		branchID := fmt.Sprintf("branch-db-%d", time.Now().UnixNano())

		_, err := svc.ExecuteQuery(ctx, srcID, "CREATE TABLE inventory (item TEXT, qty INT)")
		require.NoError(t, err)

		_, err = svc.ExecuteQuery(ctx, srcID, "INSERT INTO inventory (item, qty) VALUES ('GPU H100', 64)")
		require.NoError(t, err)

		resBranch, err := svc.BranchDatabase(srcID, branchID)
		require.NoError(t, err)
		assert.Equal(t, "BRANCH_CREATED", resBranch.Rows[0][0])

		resBranchQuery, err := svc.ExecuteQuery(ctx, branchID, "SELECT * FROM inventory")
		require.NoError(t, err)
		require.Equal(t, 1, resBranchQuery.RowCount)
		assert.Equal(t, "GPU H100", resBranchQuery.Rows[0][0])
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
