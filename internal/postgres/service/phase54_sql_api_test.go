package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

func TestPhase54_SQLAPI_FullRegressionSuite(t *testing.T) {
	ctx := context.WithValue(context.Background(), security.UserIDKey, "usr-dev")
	ctx = context.WithValue(ctx, security.OrgIDKey, "org-default")
	ctx = context.WithValue(ctx, security.ProjectIDKey, "proj-default")

	sqlSvc := service.NewSQLService()
	instID := fmt.Sprintf("postgresql-%d", time.Now().UnixNano())

	t.Run("1. Original Production Query: SELECT * FROM databases;", func(t *testing.T) {
		res, err := sqlSvc.ExecuteQuery(ctx, instID, "SELECT * FROM databases;")
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Contains(t, res.Columns, "id")
		assert.Contains(t, res.Columns, "name")
		assert.GreaterOrEqual(t, len(res.Rows), 1)
	})

	t.Run("2. Original Production Query: DROP TABLE databases;", func(t *testing.T) {
		res, err := sqlSvc.ExecuteQuery(ctx, instID, "DROP TABLE databases;")
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, 1, res.RowCount)

		// Verification: subsequent SELECT * FROM databases; returns virtual catalog view cleanly
		res2, err2 := sqlSvc.ExecuteQuery(ctx, instID, "SELECT * FROM databases;")
		require.NoError(t, err2)
		require.NotNil(t, res2)
		assert.Contains(t, res2.Columns, "id")
	})

	t.Run("3. Complete DDL & CRUD Execution Chain", func(t *testing.T) {
		// CREATE TABLE
		resCreate, errCreate := sqlSvc.ExecuteQuery(ctx, instID, "CREATE TABLE test_items (id INT, item_name TEXT, is_active BOOLEAN);")
		require.NoError(t, errCreate)
		assert.Contains(t, resCreate.Columns, "id")

		// INSERT INTO
		resInsert, errInsert := sqlSvc.ExecuteQuery(ctx, instID, "INSERT INTO test_items (id, item_name, is_active) VALUES (101, 'Alpha Unit', true);")
		require.NoError(t, errInsert)
		assert.Equal(t, 1, resInsert.RowCount)

		// SELECT
		resSelect, errSelect := sqlSvc.ExecuteQuery(ctx, instID, "SELECT * FROM test_items WHERE id = 101;")
		require.NoError(t, errSelect)
		assert.Equal(t, 1, len(resSelect.Rows))

		// UPDATE
		resUpdate, errUpdate := sqlSvc.ExecuteQuery(ctx, instID, "UPDATE test_items SET item_name = 'Beta Unit' WHERE id = 101;")
		require.NoError(t, errUpdate)
		assert.Equal(t, 1, resUpdate.RowCount)

		// DELETE FROM
		resDelete, errDelete := sqlSvc.ExecuteQuery(ctx, instID, "DELETE FROM test_items WHERE id = 101;")
		require.NoError(t, errDelete)
		assert.Equal(t, 1, resDelete.RowCount)

		// DROP TABLE
		resDrop, errDrop := sqlSvc.ExecuteQuery(ctx, instID, "DROP TABLE test_items;")
		require.NoError(t, errDrop)
		assert.Equal(t, 1, resDrop.RowCount)
	})

	t.Run("4. Empty SQL Query Validation", func(t *testing.T) {
		_, errEmpty := sqlSvc.ExecuteQuery(ctx, instID, "")
		assert.Error(t, errEmpty)
		assert.Contains(t, errEmpty.Error(), "empty SQL query statement")
	})

	t.Run("5. Dangerous Administrative SQL Statement Rejection", func(t *testing.T) {
		_, errAdmin := sqlSvc.ExecuteQuery(ctx, instID, "DROP DATABASE production_db;")
		assert.Error(t, errAdmin)
		assert.Contains(t, errAdmin.Error(), "dangerous administrative SQL statements require approval")
	})
}
