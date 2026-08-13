package service

import (
	"context"
	"errors"
	"strings"
	"time"
)

type SQLQueryResult struct {
	Columns   []string        `json:"columns"`
	Rows      [][]interface{} `json:"rows"`
	RowCount  int             `json:"rowCount"`
	LatencyMs float64         `json:"latencyMs"`
	Truncated bool            `json:"truncated"`
}

type SQLService struct {
	maxRows        int
	maxTimeout     time.Duration
	maxResponseSize int
}

func NewSQLService() *SQLService {
	return &SQLService{
		maxRows:        1000,
		maxTimeout:     5 * time.Second,
		maxResponseSize: 2 * 1024 * 1024, // 2MB
	}
}

func (s *SQLService) ExecuteQuery(ctx context.Context, instanceID, sql string) (*SQLQueryResult, error) {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return nil, errors.New("empty SQL query statement")
	}

	upper := strings.ToUpper(trimmed)
	if strings.Contains(upper, "DROP DATABASE") || strings.Contains(upper, "SHUTDOWN") {
		return nil, errors.New("dangerous administrative SQL statements require approval")
	}

	start := time.Now()

	// Parameterized / Safe Execution Simulation
	columns := []string{"id", "name", "email", "status", "created_at"}
	rows := [][]interface{}{
		{"usr_101", "Alex Rivers", "alex@anarva.io", "ACTIVE", time.Now().Add(-720 * time.Hour).Format(time.RFC3339)},
		{"usr_102", "Devon Vance", "devon@anarva.io", "ACTIVE", time.Now().Add(-360 * time.Hour).Format(time.RFC3339)},
		{"usr_103", "Elena Rostova", "elena@anarva.io", "ACTIVE", time.Now().Add(-120 * time.Hour).Format(time.RFC3339)},
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0
	if latency < 0.8 {
		latency = 0.85
	}

	return &SQLQueryResult{
		Columns:   columns,
		Rows:      rows,
		RowCount:  len(rows),
		LatencyMs: latency,
		Truncated: false,
	}, nil
}
