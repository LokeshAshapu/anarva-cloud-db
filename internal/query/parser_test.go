package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLParser_Validation(t *testing.T) {
	parser := NewParser()

	// Test valid SELECT
	parsed, err := parser.ParseAndValidate("SELECT id, email FROM users WHERE status = 'ACTIVE'")
	require.NoError(t, err)
	assert.True(t, parsed.IsReadOnly)
	assert.Equal(t, StatementSelect, parsed.StatementType)

	// Test valid INSERT
	parsedInsert, err := parser.ParseAndValidate("INSERT INTO products (name, price) VALUES ('Laptop', 999.99)")
	require.NoError(t, err)
	assert.False(t, parsedInsert.IsReadOnly)
	assert.Equal(t, StatementInsert, parsedInsert.StatementType)

	// Test forbidden DROP DATABASE
	_, err = parser.ParseAndValidate("DROP DATABASE production_db")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden SQL command")
}
