package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type SQLQueryResult struct {
	Columns   []string        `json:"columns"`
	Rows      [][]interface{} `json:"rows"`
	RowCount  int             `json:"rowCount"`
	LatencyMs float64         `json:"latencyMs"`
	Truncated bool            `json:"truncated"`
}

type TableState struct {
	Columns []string
	Rows    [][]interface{}
}

type SQLService struct {
	mu              sync.RWMutex
	maxRows         int
	maxTimeout      time.Duration
	maxResponseSize int
	instanceTables  map[string]map[string]*TableState // instanceID -> tableName -> TableState
}

func NewSQLService() *SQLService {
	return &SQLService{
		maxRows:         1000,
		maxTimeout:      5 * time.Second,
		maxResponseSize: 2 * 1024 * 1024,
		instanceTables:  make(map[string]map[string]*TableState),
	}
}

func (s *SQLService) getOrInitTables(instanceID string) map[string]*TableState {
	s.mu.Lock()
	defer s.mu.Unlock()

	tables, exists := s.instanceTables[instanceID]
	if !exists {
		tables = make(map[string]*TableState)
		// Default users table
		nowStr := time.Now().Format(time.RFC3339)
		tables["users"] = &TableState{
			Columns: []string{"id", "username", "status", "created_at"},
			Rows: [][]interface{}{
				{1, "anarva_admin", "ACTIVE", nowStr},
				{2, "app_user", "ACTIVE", nowStr},
			},
		}
		s.instanceTables[instanceID] = tables
	}
	return tables
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

	if instanceID == "" {
		instanceID = "default-instance"
	}

	tables := s.getOrInitTables(instanceID)
	start := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	var columns []string
	var rows [][]interface{}

	if strings.HasPrefix(upper, "INSERT INTO") {
		// Basic insert support into users or created tables
		for tName, tbl := range tables {
			if strings.Contains(upper, strings.ToUpper(tName)) {
				newID := len(tbl.Rows) + 1
				newRow := []interface{}{newID, fmt.Sprintf("user_%d", newID), "ACTIVE", time.Now().Format(time.RFC3339)}
				tbl.Rows = append(tbl.Rows, newRow)
				columns = tbl.Columns
				rows = tbl.Rows
				break
			}
		}
		if columns == nil {
			// Insert into default table
			tbl := tables["users"]
			newID := len(tbl.Rows) + 1
			newRow := []interface{}{newID, fmt.Sprintf("user_%d", newID), "ACTIVE", time.Now().Format(time.RFC3339)}
			tbl.Rows = append(tbl.Rows, newRow)
			columns = tbl.Columns
			rows = tbl.Rows
		}
	} else if strings.HasPrefix(upper, "CREATE TABLE") {
		// Extract table name or default to custom_table
		tName := "custom_table"
		parts := strings.Fields(trimmed)
		if len(parts) >= 3 {
			tName = strings.Trim(parts[2], "()")
		}
		tables[tName] = &TableState{
			Columns: []string{"id", "data", "created_at"},
			Rows: [][]interface{}{
				{1, "initial_record", time.Now().Format(time.RFC3339)},
			},
		}
		columns = tables[tName].Columns
		rows = tables[tName].Rows
	} else {
		// SELECT or other read query
		found := false
		for tName, tbl := range tables {
			if strings.Contains(upper, strings.ToUpper(tName)) {
				columns = tbl.Columns
				rows = tbl.Rows
				found = true
				break
			}
		}
		if !found {
			tbl := tables["users"]
			columns = tbl.Columns
			rows = tbl.Rows
		}
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0
	if latency < 0.5 {
		latency = 0.55
	}

	return &SQLQueryResult{
		Columns:   columns,
		Rows:      rows,
		RowCount:  len(rows),
		LatencyMs: latency,
		Truncated: false,
	}, nil
}
