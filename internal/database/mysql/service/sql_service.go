package service

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type QueryResult struct {
	Columns      []string                 `json:"columns"`
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int                      `json:"rowCount"`
	ExecutionMs  float64                  `json:"executionMs"`
	Warning      string                   `json:"warning,omitempty"`
}

type SQLService struct{}

func NewSQLService() *SQLService {
	return &SQLService{}
}

func (s *SQLService) ExecuteQuery(ctx context.Context, query string) (*QueryResult, error) {
	qClean := strings.TrimSpace(strings.ToUpper(query))

	// Dangerous Query Filter
	if strings.Contains(qClean, "DROP DATABASE") || strings.Contains(qClean, "SHUTDOWN") || strings.Contains(qClean, "RESET MASTER") {
		return nil, fmt.Errorf("SECURITY RISK: Dangerous query execution blocked by policy")
	}

	start := time.Now()

	// Mock SQL execution response for console
	cols := []string{"id", "username", "status", "created_at"}
	rows := []map[string]interface{}{
		{"id": 1, "username": "admin", "status": "ACTIVE", "created_at": time.Now().Format(time.RFC3339)},
		{"id": 2, "username": "app_user", "status": "ACTIVE", "created_at": time.Now().Format(time.RFC3339)},
	}

	execMs := float64(time.Since(start).Microseconds()) / 1000.0

	return &QueryResult{
		Columns:     cols,
		Rows:        rows,
		RowCount:    len(rows),
		ExecutionMs: execMs,
	}, nil
}
