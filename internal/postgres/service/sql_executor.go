package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/domain"
)

var dsnSanitizerRegex = regexp.MustCompile(`postgres://[^@]+@`)

// SanitizeSQLError removes sensitive DSNs, passwords, or connection strings from error messages.
func SanitizeSQLError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	msg = dsnSanitizerRegex.ReplaceAllString(msg, "postgres://***:***@")
	if strings.Contains(msg, "password=") {
		msg = regexp.MustCompile(`password=[^\s]+`).ReplaceAllString(msg, "password=***")
	}
	return fmt.Errorf("SQL execution error: %s", msg)
}

type PostgresSQLExecutor struct {
	fallbackSQLService *SQLService
}

func NewPostgresSQLExecutor(fallback *SQLService) *PostgresSQLExecutor {
	return &PostgresSQLExecutor{
		fallbackSQLService: fallback,
	}
}

func (e *PostgresSQLExecutor) Execute(ctx context.Context, inst *domain.PostgresInstance, customerDSN string, sqlText string) (*SQLQueryResult, error) {
	sqlText = strings.TrimSpace(sqlText)
	if sqlText == "" {
		return nil, fmt.Errorf("SQL query cannot be empty")
	}

	// 1. Fallback to SQLService simulation if no data-plane admin DSN is configured
	if customerDSN == "" {
		if e.fallbackSQLService != nil {
			return e.fallbackSQLService.ExecuteQuery(ctx, inst.ID, sqlText)
		}
		return nil, fmt.Errorf("missing CUSTOMER_DATABASE_ADMIN_URL and fallback SQLService is unconfigured")
	}

	// 4. Open Real Database Connection to Customer Database
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", customerDSN)
	if err != nil {
		return nil, SanitizeSQLError(err)
	}
	defer db.Close()

	start := time.Now()

	// Enforce 30s statement timeout in PostgreSQL
	_, _ = db.ExecContext(execCtx, "SET statement_timeout = 30000;")

	upperSQL := strings.ToUpper(sqlText)
	isSelect := strings.HasPrefix(upperSQL, "SELECT") || strings.HasPrefix(upperSQL, "EXPLAIN") || strings.HasPrefix(upperSQL, "SHOW") || strings.HasPrefix(upperSQL, "WITH")

	if isSelect {
		rows, err := db.QueryContext(execCtx, sqlText)
		if err != nil {
			return nil, SanitizeSQLError(err)
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return nil, SanitizeSQLError(err)
		}

		var resultRows [][]interface{}
		for rows.Next() {
			rowValues := make([]interface{}, len(cols))
			scanArgs := make([]interface{}, len(cols))
			for i := range rowValues {
				scanArgs[i] = &rowValues[i]
			}

			if err := rows.Scan(scanArgs...); err != nil {
				return nil, SanitizeSQLError(err)
			}

			// Format scanned byte slices and null values cleanly
			for i, val := range rowValues {
				if b, ok := val.([]byte); ok {
					rowValues[i] = string(b)
				}
			}
			resultRows = append(resultRows, rowValues)
		}

		if resultRows == nil {
			resultRows = [][]interface{}{}
		}

		latency := float64(time.Since(start).Microseconds()) / 1000.0

		return &SQLQueryResult{
			Columns:   cols,
			Rows:      resultRows,
			RowCount:  len(resultRows),
			LatencyMs: latency,
			Truncated: false,
		}, nil
	}

	// DDL / DML Query Execution (CREATE TABLE, INSERT, UPDATE, DELETE, DROP TABLE, TRUNCATE)
	res, err := db.ExecContext(execCtx, sqlText)
	if err != nil {
		return nil, SanitizeSQLError(err)
	}

	rowsAffected, _ := res.RowsAffected()
	latency := float64(time.Since(start).Microseconds()) / 1000.0

	var statusMsg string
	if rowsAffected > 0 {
		statusMsg = fmt.Sprintf("Query executed successfully. Affected rows: %d", rowsAffected)
	} else {
		statusMsg = "Query executed successfully."
	}

	return &SQLQueryResult{
		Columns:   []string{"status"},
		Rows:      [][]interface{}{{statusMsg}},
		RowCount:  1,
		LatencyMs: latency,
		Truncated: false,
	}, nil
}
