package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
)

func TestSQLService_StatefulExecution(t *testing.T) {
	svc := service.NewSQLService()
	ctx := context.Background()
	instID := "test-db-inst-01"

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

	t.Run("INSERT INTO users with explicit column mapping (name, email)", func(t *testing.T) {
		res, err := svc.ExecuteQuery(ctx, instID, "INSERT INTO users (name, email) VALUES ('Alice Johnson', 'alice@example.com'), ('Bob Smith', 'bob@example.com')")
		require.NoError(t, err)
		assert.Equal(t, 2, res.RowCount)

		// Verify SELECT * FROM users LIMIT 10
		resSelect, err := svc.ExecuteQuery(ctx, instID, "SELECT * FROM users LIMIT 10")
		require.NoError(t, err)
		assert.Equal(t, 4, resSelect.RowCount) // 2 default + 2 inserted
		assert.Equal(t, 3, resSelect.Rows[2][0]) // id = 3
		assert.Equal(t, "Alice Johnson", resSelect.Rows[2][1]) // username = Alice Johnson
		assert.Equal(t, "alice@example.com", resSelect.Rows[2][2]) // email = alice@example.com
		assert.Equal(t, "ACTIVE", resSelect.Rows[2][3]) // status = ACTIVE
	})

	t.Run("CREATE TABLE orders creates new table schema", func(t *testing.T) {
		res, err := svc.ExecuteQuery(ctx, instID, "CREATE TABLE orders (id INT, item TEXT, price FLOAT)")
		require.NoError(t, err)
		assert.Equal(t, []string{"id", "item", "price"}, res.Columns)
		assert.Equal(t, 0, res.RowCount)
	})

	t.Run("INSERT INTO orders inserts actual parsed values", func(t *testing.T) {
		res, err := svc.ExecuteQuery(ctx, instID, "INSERT INTO orders (id, item, price) VALUES (101, 'MacBook Pro', 1999.99)")
		require.NoError(t, err)
		assert.Equal(t, 1, res.RowCount)
		require.Len(t, res.Rows, 1)
		assert.Equal(t, 101, res.Rows[0][0])
		assert.Equal(t, "MacBook Pro", res.Rows[0][1])
		assert.Equal(t, 1999.99, res.Rows[0][2])
	})

	t.Run("SELECT * FROM orders retrieves inserted data", func(t *testing.T) {
		res, err := svc.ExecuteQuery(ctx, instID, "SELECT * FROM orders")
		require.NoError(t, err)
		assert.Equal(t, []string{"id", "item", "price"}, res.Columns)
		assert.Equal(t, 1, res.RowCount)
		assert.Equal(t, 101, res.Rows[0][0])
		assert.Equal(t, "MacBook Pro", res.Rows[0][1])
	})

	t.Run("SELECT * FROM nonexistent table returns relation does not exist error", func(t *testing.T) {
		res, err := svc.ExecuteQuery(ctx, instID, "SELECT * FROM non_existent_table")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), `relation "non_existent_table" does not exist`)
	})

	t.Run("DROP TABLE orders removes table", func(t *testing.T) {
		res, err := svc.ExecuteQuery(ctx, instID, "DROP TABLE orders")
		require.NoError(t, err)
		assert.NotNil(t, res)

		res2, err2 := svc.ExecuteQuery(ctx, instID, "SELECT * FROM orders")
		assert.Error(t, err2)
		assert.Nil(t, res2)
	})
}
