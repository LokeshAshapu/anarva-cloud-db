package query

import (
	"fmt"
	"strings"

	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

// ParseAndValidate parses a raw SQL query string and enforces safety rules.
func (p *Parser) ParseAndValidate(rawSQL string) (*ParsedQuery, error) {
	trimmed := strings.TrimSpace(rawSQL)
	if trimmed == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "SQL query cannot be empty")
	}

	upper := strings.ToUpper(trimmed)

	// Check forbidden destructive statements
	forbiddenKeywords := []string{"DROP DATABASE", "ALTER SYSTEM", "SHUTDOWN", "GRANT ALL"}
	for _, kw := range forbiddenKeywords {
		if strings.Contains(upper, kw) {
			return nil, appErrors.New(appErrors.CodeForbidden, fmt.Sprintf("forbidden SQL command: %s", kw))
		}
	}

	var stmtType StatementType
	isReadOnly := false

	switch {
	case strings.HasPrefix(upper, "SELECT"), strings.HasPrefix(upper, "WITH"):
		stmtType = StatementSelect
		isReadOnly = true
	case strings.HasPrefix(upper, "INSERT"):
		stmtType = StatementInsert
	case strings.HasPrefix(upper, "UPDATE"):
		stmtType = StatementUpdate
	case strings.HasPrefix(upper, "DELETE"):
		stmtType = StatementDelete
	case strings.HasPrefix(upper, "CREATE"), strings.HasPrefix(upper, "ALTER"), strings.HasPrefix(upper, "DROP TABLE"):
		stmtType = StatementDDL
	default:
		stmtType = StatementUnknown
	}

	return &ParsedQuery{
		RawSQL:        trimmed,
		StatementType: stmtType,
		IsReadOnly:    isReadOnly,
	}, nil
}
