package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
)

type Executor interface {
	Execute(ctx context.Context, connectionString string, parsedQuery *ParsedQuery, params []interface{}) (*QueryResult, error)
}

type postgresExecutor struct{}

func NewPostgresExecutor() Executor {
	return &postgresExecutor{}
}

func (e *postgresExecutor) Execute(ctx context.Context, connectionString string, parsedQuery *ParsedQuery, params []interface{}) (*QueryResult, error) {
	start := time.Now()

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to connect to target database instance")
	}
	defer db.Close()

	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if parsedQuery.IsReadOnly {
		rows, err := db.QueryContext(ctxTimeout, parsedQuery.RawSQL, params...)
		if err != nil {
			metrics.RecordDatabaseQuery(string(parsedQuery.StatementType), "error", time.Since(start).Seconds())
			return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to execute SELECT query")
		}
		defer rows.Close()

		cols, err := rows.ColumnTypes()
		if err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch column metadata")
		}

		var columns []ColumnInfo
		for _, col := range cols {
			columns = append(columns, ColumnInfo{
				Name: col.Name(),
				Type: col.DatabaseTypeName(),
			})
		}

		var resultRows []map[string]interface{}
		colNames, _ := rows.Columns()

		for rows.Next() {
			columnPointers := make([]interface{}, len(colNames))
			columnData := make([]interface{}, len(colNames))
			for i := range columnData {
				columnPointers[i] = &columnData[i]
			}

			if err := rows.Scan(columnPointers...); err != nil {
				return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to scan query row")
			}

			rowMap := make(map[string]interface{})
			for i, colName := range colNames {
				val := columnData[i]
				if b, ok := val.([]byte); ok {
					rowMap[colName] = string(b)
				} else {
					rowMap[colName] = val
				}
			}
			resultRows = append(resultRows, rowMap)
		}

		elapsed := time.Since(start).Seconds() * 1000
		metrics.RecordDatabaseQuery(string(parsedQuery.StatementType), "success", time.Since(start).Seconds())

		return &QueryResult{
			Columns:         columns,
			Rows:            resultRows,
			RowsAffected:    int64(len(resultRows)),
			ExecutionTimeMs: elapsed,
		}, nil
	}

	// Non-select mutation query
	res, err := db.ExecContext(ctxTimeout, parsedQuery.RawSQL, params...)
	if err != nil {
		metrics.RecordDatabaseQuery(string(parsedQuery.StatementType), "error", time.Since(start).Seconds())
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, fmt.Sprintf("failed to execute %s query", parsedQuery.StatementType))
	}

	rowsAffected, _ := res.RowsAffected()
	elapsed := time.Since(start).Seconds() * 1000
	metrics.RecordDatabaseQuery(string(parsedQuery.StatementType), "success", time.Since(start).Seconds())

	return &QueryResult{
		Columns:         []ColumnInfo{},
		Rows:            []map[string]interface{}{},
		RowsAffected:    rowsAffected,
		ExecutionTimeMs: elapsed,
	}, nil
}

type mockExecutor struct{}

func NewMockExecutor() Executor {
	return &mockExecutor{}
}

func (e *mockExecutor) Execute(ctx context.Context, connectionString string, parsedQuery *ParsedQuery, params []interface{}) (*QueryResult, error) {
	start := time.Now()

	var columns []ColumnInfo
	var rows []map[string]interface{}

	sqlLower := strings.ToLower(parsedQuery.RawSQL)

	if strings.Contains(sqlLower, "from users") {
		columns = []ColumnInfo{
			{Name: "id", Type: "UUID"},
			{Name: "email", Type: "VARCHAR"},
			{Name: "full_name", Type: "VARCHAR"},
			{Name: "role", Type: "VARCHAR"},
			{Name: "status", Type: "VARCHAR"},
		}
		rows = []map[string]interface{}{
			{"id": "usr-87a1d9b3", "email": "admin@anarva.io", "full_name": "Anarva Admin", "role": "owner", "status": "active"},
			{"id": "usr-92c4b8e1", "email": "lokesh@anarva.io", "full_name": "Lokesh Ashapu", "role": "admin", "status": "active"},
			{"id": "usr-11f8e2d4", "email": "developer@anarva.io", "full_name": "Dev Team", "role": "developer", "status": "active"},
			{"id": "usr-55a7c3e9", "email": "security@anarva.io", "full_name": "Security Lead", "role": "auditor", "status": "active"},
		}
	} else if strings.Contains(sqlLower, "from databases") || strings.Contains(sqlLower, "from database") {
		columns = []ColumnInfo{
			{Name: "id", Type: "UUID"},
			{Name: "name", Type: "VARCHAR"},
			{Name: "engine", Type: "VARCHAR"},
			{Name: "status", Type: "VARCHAR"},
			{Name: "port", Type: "INTEGER"},
		}
		rows = []map[string]interface{}{
			{"id": "db-uuid-1", "name": "Primary Application Database", "engine": "postgres", "status": "RUNNING", "port": 15432},
			{"id": "db-uuid-2", "name": "Analytics Data Warehouse", "engine": "postgres", "status": "RUNNING", "port": 15433},
		}
	} else if strings.Contains(sqlLower, "from metrics") {
		columns = []ColumnInfo{
			{Name: "metric_name", Type: "VARCHAR"},
			{Name: "value", Type: "DOUBLE"},
			{Name: "timestamp", Type: "TIMESTAMP"},
		}
		rows = []map[string]interface{}{
			{"metric_name": "cpu_usage_percent", "value": 12.4, "timestamp": time.Now().Format(time.RFC3339)},
			{"metric_name": "memory_usage_bytes", "value": float64(2576980377), "timestamp": time.Now().Format(time.RFC3339)},
			{"metric_name": "query_latency_ms", "value": 1.42, "timestamp": time.Now().Format(time.RFC3339)},
		}
	} else {
		if parsedQuery.StatementType == StatementSelect {
			if strings.Contains(sqlLower, "customer_orders") || strings.Contains(sqlLower, "orders") {
				columns = []ColumnInfo{
					{Name: "id", Type: "INT4"},
					{Name: "customer_name", Type: "VARCHAR"},
					{Name: "amount", Type: "NUMERIC"},
					{Name: "status", Type: "VARCHAR"},
					{Name: "created_at", Type: "TIMESTAMP"},
				}
				rows = []map[string]interface{}{
					{"id": 1, "customer_name": "Lokesh Ashapu", "amount": "299.99", "status": "COMPLETED", "created_at": time.Now().Format("2006-01-02 15:04:05")},
					{"id": 2, "customer_name": "Enterprise Client", "amount": "1499.00", "status": "PROCESSING", "created_at": time.Now().Format("2006-01-02 15:04:05")},
					{"id": 3, "customer_name": "Acme Corp", "amount": "850.50", "status": "PAID", "created_at": time.Now().Format("2006-01-02 15:04:05")},
				}
			} else {
				columns = []ColumnInfo{
					{Name: "id", Type: "INT4"},
					{Name: "query_status", Type: "VARCHAR"},
					{Name: "executed_sql", Type: "VARCHAR"},
					{Name: "timestamp", Type: "TIMESTAMP"},
				}
				rows = []map[string]interface{}{
					{"id": 1, "query_status": "EXECUTED_SUCCESSFULLY", "executed_sql": parsedQuery.RawSQL, "timestamp": time.Now().Format("2006-01-02 15:04:05")},
				}
			}
		}
	}

	elapsed := time.Since(start).Seconds() * 1000
	metrics.RecordDatabaseQuery(string(parsedQuery.StatementType), "success", time.Since(start).Seconds())

	return &QueryResult{
		Columns:         columns,
		Rows:            rows,
		RowsAffected:    int64(len(rows)),
		ExecutionTimeMs: elapsed,
	}, nil
}

