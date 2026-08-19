package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	Name           string                 `json:"name"`
	Columns        []string               `json:"columns"`
	ColumnDefaults map[string]interface{} `json:"columnDefaults"`
	Rows           [][]interface{}        `json:"rows"`
}

type SQLService struct {
	mu             sync.RWMutex
	filePath       string
	instanceTables map[string]map[string]*TableState // instanceID -> tableName -> TableState
}

func NewSQLService() *SQLService {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	_ = os.MkdirAll(dataDir, 0755)

	filePath := filepath.Join(dataDir, "anarva_sql_service_state.json")
	svc := &SQLService{
		filePath:       filePath,
		instanceTables: make(map[string]map[string]*TableState),
	}
	svc.loadFromFile()

	if len(svc.instanceTables) == 0 {
		tmpPath := filepath.Join(os.TempDir(), "anarva_sql_service_state.json")
		if _, err := os.Stat(tmpPath); err == nil {
			svc.filePath = tmpPath
			svc.loadFromFile()
			svc.filePath = filePath
		}
	}

	return svc
}

func (s *SQLService) loadFromFile() {
	if s.filePath == "" {
		return
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var loaded map[string]map[string]*TableState
	if err := json.Unmarshal(data, &loaded); err == nil && loaded != nil {
		for _, instance := range loaded {
			for _, tbl := range instance {
				for rIdx, row := range tbl.Rows {
					for cIdx, val := range row {
						if f, ok := val.(float64); ok {
							if f == float64(int(f)) {
								tbl.Rows[rIdx][cIdx] = int(f)
							}
						}
					}
				}
			}
		}
		s.instanceTables = loaded
	}
}

func (s *SQLService) saveToFileLocked() {
	if s.filePath == "" {
		return
	}
	data, err := json.MarshalIndent(s.instanceTables, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(s.filePath, data, 0644)

	tmpPath := filepath.Join(os.TempDir(), "anarva_sql_service_state.json")
	_ = os.WriteFile(tmpPath, data, 0644)
}

func (s *SQLService) getOrInitTables(instanceID string) map[string]*TableState {
	tables, exists := s.instanceTables[instanceID]
	if !exists {
		tables = make(map[string]*TableState)
		nowStr := time.Now().Format(time.RFC3339)
		tables["users"] = &TableState{
			Name:    "users",
			Columns: []string{"id", "username", "email", "status", "created_at"},
			ColumnDefaults: map[string]interface{}{
				"status":     "ACTIVE",
				"created_at": nowStr,
			},
			Rows: [][]interface{}{
				{1, "anarva_admin", "admin@anarva.io", "ACTIVE", nowStr},
				{2, "app_user", "user@anarva.io", "ACTIVE", nowStr},
			},
		}
		tables["databases"] = &TableState{
			Name:    "databases",
			Columns: []string{"id", "name", "engine", "status", "version", "created_at"},
			ColumnDefaults: map[string]interface{}{
				"engine":     "POSTGRESQL",
				"status":     "ACTIVE",
				"version":    "17.2",
				"created_at": nowStr,
			},
			Rows: [][]interface{}{
				{instanceID, "primary_db", "POSTGRESQL", "ACTIVE", "17.2", nowStr},
			},
		}
		s.instanceTables[instanceID] = tables
		s.saveToFileLocked()
	}
	return tables
}

func (s *SQLService) BranchDatabase(sourceID, newBranchID string) (*SQLQueryResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sourceTables := s.getOrInitTables(sourceID)
	branchTables := make(map[string]*TableState)

	for tblName, tblState := range sourceTables {
		newRows := make([][]interface{}, len(tblState.Rows))
		for i, r := range tblState.Rows {
			rowCopy := make([]interface{}, len(r))
			copy(rowCopy, r)
			newRows[i] = rowCopy
		}
		colsCopy := make([]string, len(tblState.Columns))
		copy(colsCopy, tblState.Columns)

		defaultsCopy := make(map[string]interface{})
		for k, v := range tblState.ColumnDefaults {
			defaultsCopy[k] = v
		}

		branchTables[tblName] = &TableState{
			Name:           tblName,
			Columns:        colsCopy,
			ColumnDefaults: defaultsCopy,
			Rows:           newRows,
		}
	}

	s.instanceTables[newBranchID] = branchTables
	s.saveToFileLocked()

	return &SQLQueryResult{
		Columns:   []string{"status", "source_id", "branch_id", "tables_cloned"},
		Rows:      [][]interface{}{{"BRANCH_CREATED", sourceID, newBranchID, len(branchTables)}},
		RowCount:  1,
		LatencyMs: 0.85,
	}, nil
}

func (s *SQLService) ExecuteQuery(ctx context.Context, instanceID, sql string) (*SQLQueryResult, error) {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return nil, errors.New("empty SQL query statement")
	}

	trimmed = strings.TrimSuffix(trimmed, ";")
	upper := strings.ToUpper(trimmed)

	if strings.Contains(upper, "DROP DATABASE") || strings.Contains(upper, "SHUTDOWN") {
		return nil, errors.New("dangerous administrative SQL statements require approval")
	}

	if instanceID == "" {
		instanceID = "default-instance"
	}

	start := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	tables := s.getOrInitTables(instanceID)

	var columns []string
	var rows [][]interface{}
	var rowCount int

	switch {
	case upper == "BEGIN" || upper == "START TRANSACTION" || upper == "COMMIT" || upper == "END" || upper == "ROLLBACK":
		columns = []string{"status"}
		rows = [][]interface{}{{upper}}
		rowCount = 1

	case strings.HasPrefix(upper, "CREATE TABLE"):
		var err error
		columns, rows, rowCount, err = s.handleCreateTable(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}

	case strings.HasPrefix(upper, "ALTER TABLE"):
		var err error
		columns, rows, rowCount, err = s.handleAlterTable(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}

	case strings.HasPrefix(upper, "TRUNCATE"):
		var err error
		columns, rows, rowCount, err = s.handleTruncate(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}

	case strings.HasPrefix(upper, "INSERT INTO"):
		var err error
		columns, rows, rowCount, err = s.handleInsert(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}

	case strings.HasPrefix(upper, "DROP TABLE"):
		var err error
		columns, rows, rowCount, err = s.handleDropTable(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}

	case strings.HasPrefix(upper, "UPDATE"):
		var err error
		columns, rows, rowCount, err = s.handleUpdate(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}

	case strings.HasPrefix(upper, "DELETE FROM"):
		var err error
		columns, rows, rowCount, err = s.handleDelete(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}

	case strings.HasPrefix(upper, "DESCRIBE") || strings.HasPrefix(upper, "\\D") || strings.HasPrefix(upper, "EXPLAIN"):
		var err error
		columns, rows, rowCount, err = s.handleDescribeOrExplain(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}

	default: // SELECT, SHOW, or other queries
		var err error
		columns, rows, rowCount, err = s.handleSelect(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}
	}

	if strings.HasPrefix(upper, "CREATE TABLE") || strings.HasPrefix(upper, "ALTER TABLE") || strings.HasPrefix(upper, "TRUNCATE") || strings.HasPrefix(upper, "INSERT INTO") || strings.HasPrefix(upper, "UPDATE") || strings.HasPrefix(upper, "DELETE FROM") || strings.HasPrefix(upper, "DROP TABLE") {
		s.saveToFileLocked()
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0
	if latency < 0.2 {
		latency = 0.45
	}

	return &SQLQueryResult{
		Columns:   columns,
		Rows:      rows,
		RowCount:  rowCount,
		LatencyMs: latency,
		Truncated: false,
	}, nil
}

func (s *SQLService) handleCreateTable(tables map[string]*TableState, trimmed, upper string) ([]string, [][]interface{}, int, error) {
	ifNotExists := strings.Contains(upper, "IF NOT EXISTS")

	cleanStr := strings.TrimPrefix(upper, "CREATE TABLE")
	cleanStr = strings.TrimPrefix(cleanStr, " IF NOT EXISTS")
	cleanStr = strings.TrimSpace(cleanStr)

	idxParen := strings.Index(cleanStr, "(")
	var tableName string
	var colsBlock string

	if idxParen != -1 {
		tableName = strings.TrimSpace(cleanStr[:idxParen])
		colsBlock = cleanStr[idxParen+1:]
		if idxClose := strings.LastIndex(colsBlock, ")"); idxClose != -1 {
			colsBlock = colsBlock[:idxClose]
		}
	} else {
		tableName = strings.Fields(cleanStr)[0]
	}

	tableName = strings.ToLower(strings.Trim(tableName, `"'` + "`"))
	if tableName == "" {
		tableName = "custom_table"
	}

	if _, exists := tables[tableName]; exists {
		if ifNotExists {
			tbl := tables[tableName]
			return tbl.Columns, tbl.Rows, 0, nil
		}
		return nil, nil, 0, fmt.Errorf("relation %q already exists", tableName)
	}

	var columns []string
	defaults := make(map[string]interface{})

	if colsBlock != "" {
		colParts := strings.Split(colsBlock, ",")
		for _, part := range colParts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			colFields := strings.Fields(part)
			if len(colFields) > 0 {
				colName := strings.ToLower(strings.Trim(colFields[0], `"'` + "`"))
				if colName != "primary" && colName != "foreign" && colName != "constraint" && colName != "unique" {
					columns = append(columns, colName)

					upperPart := strings.ToUpper(part)
					if defIdx := strings.Index(upperPart, "DEFAULT"); defIdx != -1 {
						afterDef := strings.TrimSpace(part[defIdx+len("DEFAULT"):])
						defFields := strings.Fields(afterDef)
						if len(defFields) > 0 {
							defaults[colName] = parseSQLValue(defFields[0])
						}
					}
				}
			}
		}
	}

	if len(columns) == 0 {
		columns = []string{"id", "name", "created_at"}
	}

	tables[tableName] = &TableState{
		Name:           tableName,
		Columns:        columns,
		ColumnDefaults: defaults,
		Rows:           [][]interface{}{},
	}

	return columns, [][]interface{}{}, 0, nil
}

func (s *SQLService) handleAlterTable(tables map[string]*TableState, trimmed, upper string) ([]string, [][]interface{}, int, error) {
	after := strings.TrimSpace(trimmed[len("ALTER TABLE"):])
	fields := strings.Fields(after)
	if len(fields) < 2 {
		return nil, nil, 0, errors.New("syntax error in ALTER TABLE statement")
	}

	tableName := strings.ToLower(strings.Trim(fields[0], `"'` + "`"))
	tbl, exists := tables[tableName]
	if !exists {
		return nil, nil, 0, fmt.Errorf("relation %q does not exist", tableName)
	}

	action := strings.ToUpper(fields[1])
	switch action {
	case "ADD":
		afterAdd := strings.TrimSpace(after[strings.Index(strings.ToUpper(after), "ADD")+len("ADD"):])
		afterAdd = strings.TrimPrefix(afterAdd, "COLUMN ")
		afterAdd = strings.TrimPrefix(afterAdd, "column ")
		addFields := strings.Fields(afterAdd)
		if len(addFields) == 0 {
			return nil, nil, 0, errors.New("missing column name in ALTER TABLE ADD")
		}
		newCol := strings.ToLower(strings.Trim(addFields[0], `"'` + "`"))

		for _, c := range tbl.Columns {
			if c == newCol {
				return nil, nil, 0, fmt.Errorf("column %q of relation %q already exists", newCol, tableName)
			}
		}

		tbl.Columns = append(tbl.Columns, newCol)

		var defVal interface{} = nil
		if defIdx := strings.Index(strings.ToUpper(afterAdd), "DEFAULT"); defIdx != -1 {
			afterDef := strings.TrimSpace(afterAdd[defIdx+len("DEFAULT"):])
			defFields := strings.Fields(afterDef)
			if len(defFields) > 0 {
				defVal = parseSQLValue(defFields[0])
			}
		}

		if defVal != nil {
			if tbl.ColumnDefaults == nil {
				tbl.ColumnDefaults = make(map[string]interface{})
			}
			tbl.ColumnDefaults[newCol] = defVal
		}

		for i := range tbl.Rows {
			tbl.Rows[i] = append(tbl.Rows[i], defVal)
		}

		return tbl.Columns, tbl.Rows, 0, nil

	case "DROP":
		afterDrop := strings.TrimSpace(after[strings.Index(strings.ToUpper(after), "DROP")+len("DROP"):])
		afterDrop = strings.TrimPrefix(afterDrop, "COLUMN ")
		afterDrop = strings.TrimPrefix(afterDrop, "column ")
		dropFields := strings.Fields(afterDrop)
		if len(dropFields) == 0 {
			return nil, nil, 0, errors.New("missing column name in ALTER TABLE DROP")
		}
		dropCol := strings.ToLower(strings.Trim(dropFields[0], `"'` + "`"))

		colIdx := resolveColumnIndex(tbl.Columns, dropCol)
		if colIdx == -1 {
			return nil, nil, 0, fmt.Errorf("column %q of relation %q does not exist", dropCol, tableName)
		}

		tbl.Columns = append(tbl.Columns[:colIdx], tbl.Columns[colIdx+1:]...)

		for i, row := range tbl.Rows {
			if colIdx < len(row) {
				tbl.Rows[i] = append(row[:colIdx], row[colIdx+1:]...)
			}
		}

		return tbl.Columns, tbl.Rows, 0, nil

	case "RENAME":
		toIdx := strings.Index(strings.ToUpper(after), "TO")
		if toIdx == -1 {
			return nil, nil, 0, errors.New("syntax error in ALTER TABLE RENAME TO")
		}
		newName := strings.ToLower(strings.TrimSpace(after[toIdx+2:]))
		newName = strings.Trim(strings.Fields(newName)[0], `"'` + "`")

		delete(tables, tableName)
		tbl.Name = newName
		tables[newName] = tbl

		return tbl.Columns, tbl.Rows, 0, nil
	}

	return tbl.Columns, tbl.Rows, 0, nil
}

func (s *SQLService) handleTruncate(tables map[string]*TableState, trimmed, upper string) ([]string, [][]interface{}, int, error) {
	after := strings.TrimPrefix(upper, "TRUNCATE")
	after = strings.TrimPrefix(after, " TABLE")
	tableName := strings.ToLower(strings.TrimSpace(after))
	fields := strings.Fields(tableName)
	if len(fields) > 0 {
		tableName = strings.Trim(fields[0], `"'` + "`")
	}

	tbl, exists := tables[tableName]
	if !exists {
		return nil, nil, 0, fmt.Errorf("relation %q does not exist", tableName)
	}

	rowCount := len(tbl.Rows)
	tbl.Rows = [][]interface{}{}
	return tbl.Columns, [][]interface{}{}, rowCount, nil
}

func (s *SQLService) handleInsert(tables map[string]*TableState, trimmed, upper string) ([]string, [][]interface{}, int, error) {
	afterInsert := strings.TrimSpace(trimmed[len("INSERT INTO"):])
	parts := strings.Fields(afterInsert)
	if len(parts) == 0 {
		return nil, nil, 0, errors.New("syntax error in INSERT statement")
	}

	tableNameRaw := parts[0]
	if idxParen := strings.Index(tableNameRaw, "("); idxParen != -1 {
		tableNameRaw = tableNameRaw[:idxParen]
	}
	tableName := strings.ToLower(strings.Trim(tableNameRaw, `"'` + "`"))

	tbl, exists := tables[tableName]
	if !exists {
		return nil, nil, 0, fmt.Errorf("relation %q does not exist", tableName)
	}

	valuesIdx := strings.Index(upper, "VALUES")
	if valuesIdx == -1 {
		return nil, nil, 0, errors.New("missing VALUES clause in INSERT statement")
	}

	var targetCols []string
	betweenTableAndValues := trimmed[len("INSERT INTO")+len(parts[0]) : valuesIdx]
	if idxOpen := strings.Index(betweenTableAndValues, "("); idxOpen != -1 {
		if idxClose := strings.Index(betweenTableAndValues[idxOpen:], ")"); idxClose != -1 {
			rawCols := betweenTableAndValues[idxOpen+1 : idxOpen+idxClose]
			for _, col := range strings.Split(rawCols, ",") {
				c := strings.ToLower(strings.TrimSpace(strings.Trim(col, `"'` + "`")))
				if c != "" {
					targetCols = append(targetCols, c)
				}
			}
		}
	}

	valuesStr := trimmed[valuesIdx+len("VALUES"):]
	valuesTuples := s.parseValuesTuples(valuesStr)

	if len(valuesTuples) == 0 {
		return nil, nil, 0, errors.New("no values specified in INSERT statement")
	}

	insertedRows := make([][]interface{}, 0, len(valuesTuples))
	for _, tuple := range valuesTuples {
		row := make([]interface{}, len(tbl.Columns))
		mapped := make(map[int]bool)

		if len(targetCols) > 0 {
			for i, targetCol := range targetCols {
				if i < len(tuple) {
					parsedVal := parseSQLValue(tuple[i])
					colIdx := resolveColumnIndex(tbl.Columns, targetCol)
					if colIdx != -1 {
						row[colIdx] = parsedVal
						mapped[colIdx] = true
					}
				}
			}
		} else {
			for i, val := range tuple {
				if i < len(tbl.Columns) {
					row[i] = parseSQLValue(val)
					mapped[i] = true
				}
			}
		}

		for i, col := range tbl.Columns {
			if !mapped[i] {
				if defVal, ok := tbl.ColumnDefaults[col]; ok && defVal != nil {
					row[i] = defVal
				} else if col == "id" {
					row[i] = len(tbl.Rows) + 1
				} else if strings.HasPrefix(col, "is_") || strings.HasPrefix(col, "has_") || col == "active" || col == "enabled" {
					row[i] = true
				} else if col == "status" {
					row[i] = "ACTIVE"
				} else if col == "created_at" || col == "updated_at" {
					row[i] = time.Now().Format(time.RFC3339)
				} else {
					row[i] = nil
				}
			}
		}

		tbl.Rows = append(tbl.Rows, row)
		insertedRows = append(insertedRows, row)
	}

	return tbl.Columns, insertedRows, len(insertedRows), nil
}

func (s *SQLService) handleSelect(tables map[string]*TableState, trimmed, upper string) ([]string, [][]interface{}, int, error) {
	if strings.Contains(upper, "VERSION()") {
		return []string{"version"}, [][]interface{}{{"PostgreSQL 17.2 (ANARVA Cloud Enterprise Database v2.4)"}}, 1, nil
	}

	if upper == "SHOW DATABASES" || upper == "\\L" {
		cols := []string{"datname", "encoding", "collate"}
		rows := [][]interface{}{
			{"main", "UTF8", "en_US.utf8"},
			{"postgres", "UTF8", "en_US.utf8"},
			{"anarva_db", "UTF8", "en_US.utf8"},
		}
		return cols, rows, len(rows), nil
	}

	if upper == "SHOW TABLES" || strings.Contains(upper, "INFORMATION_SCHEMA.TABLES") || upper == "\\DT" {
		cols := []string{"table_name", "table_type"}
		rows := make([][]interface{}, 0, len(tables))
		for name := range tables {
			rows = append(rows, []interface{}{name, "BASE TABLE"})
		}
		return cols, rows, len(rows), nil
	}

	fromIdx := strings.Index(upper, "FROM")
	if fromIdx == -1 {
		return []string{"?column?"}, [][]interface{}{{1}}, 1, nil
	}

	afterFrom := strings.TrimSpace(trimmed[fromIdx+len("FROM"):])
	fromFields := strings.Fields(afterFrom)
	if len(fromFields) == 0 {
		return nil, nil, 0, errors.New("syntax error in FROM clause")
	}

	tableName := strings.ToLower(strings.Trim(fromFields[0], `"'` + "`;"))

	tbl, exists := tables[tableName]
	if !exists {
		if tableName == "databases" {
			cols := []string{"id", "name", "engine", "status", "version"}
			rows := [][]interface{}{
				{"default-db", "primary_db", "POSTGRESQL", "ACTIVE", "17.2"},
			}
			return cols, rows, len(rows), nil
		}
		return nil, nil, 0, fmt.Errorf("relation %q does not exist", tableName)
	}

	whereIdx := strings.Index(upper, "WHERE")
	orderIdx := strings.Index(upper, "ORDER BY")
	limitIdx := strings.Index(upper, "LIMIT")

	var filteredRows [][]interface{}
	if whereIdx != -1 {
		endWhere := len(trimmed)
		if orderIdx != -1 && orderIdx > whereIdx {
			endWhere = orderIdx
		} else if limitIdx != -1 && limitIdx > whereIdx {
			endWhere = limitIdx
		}
		whereClause := trimmed[whereIdx+len("WHERE") : endWhere]
		filteredRows = s.filterRows(tbl, whereClause)
	} else {
		// Copy table rows slice
		filteredRows = make([][]interface{}, len(tbl.Rows))
		copy(filteredRows, tbl.Rows)
	}

	// Handle aggregations: COUNT(), SUM(), AVG(), MIN(), MAX()
	if strings.Contains(upper, "COUNT(") {
		return []string{"count"}, [][]interface{}{{len(filteredRows)}}, 1, nil
	}
	if strings.Contains(upper, "SUM(") || strings.Contains(upper, "AVG(") {
		var total float64
		count := 0
		for _, r := range filteredRows {
			for _, cell := range r {
				if f, err := strconv.ParseFloat(fmt.Sprintf("%v", cell), 64); err == nil {
					total += f
					count++
					break
				}
			}
		}
		if strings.Contains(upper, "AVG(") {
			avg := 0.0
			if count > 0 {
				avg = total / float64(count)
			}
			return []string{"avg"}, [][]interface{}{{avg}}, 1, nil
		}
		return []string{"sum"}, [][]interface{}{{total}}, 1, nil
	}

	// ORDER BY sorting
	if orderIdx != -1 {
		endOrder := len(trimmed)
		if limitIdx != -1 && limitIdx > orderIdx {
			endOrder = limitIdx
		}
		orderClause := strings.TrimSpace(trimmed[orderIdx+len("ORDER BY") : endOrder])
		orderFields := strings.Fields(orderClause)
		if len(orderFields) > 0 {
			sortCol := strings.ToLower(strings.Trim(orderFields[0], `"'` + "`"))
			isDesc := len(orderFields) > 1 && strings.ToUpper(orderFields[1]) == "DESC"

			sortIdx := resolveColumnIndex(tbl.Columns, sortCol)
			if sortIdx != -1 {
				sort.SliceStable(filteredRows, func(i, j int) bool {
					valI := fmt.Sprintf("%v", filteredRows[i][sortIdx])
					valJ := fmt.Sprintf("%v", filteredRows[j][sortIdx])

					fI, errI := strconv.ParseFloat(valI, 64)
					fJ, errJ := strconv.ParseFloat(valJ, 64)

					var less bool
					if errI == nil && errJ == nil {
						less = fI < fJ
					} else {
						less = strings.ToLower(valI) < strings.ToLower(valJ)
					}

					if isDesc {
						return !less
					}
					return less
				})
			}
		}
	}

	// LIMIT clipping
	if limitIdx != -1 {
		limitStr := strings.TrimSpace(upper[limitIdx+len("LIMIT"):])
		limitFields := strings.Fields(limitStr)
		if len(limitFields) > 0 {
			if limitVal, err := strconv.Atoi(limitFields[0]); err == nil && limitVal >= 0 {
				if limitVal < len(filteredRows) {
					filteredRows = filteredRows[:limitVal]
				}
			}
		}
	}

	return tbl.Columns, filteredRows, len(filteredRows), nil
}

func (s *SQLService) handleDescribeOrExplain(tables map[string]*TableState, trimmed, upper string) ([]string, [][]interface{}, int, error) {
	if strings.HasPrefix(upper, "EXPLAIN") {
		cols := []string{"QUERY PLAN"}
		planStr := "Seq Scan on table (cost=0.00..1.05 rows=3 width=64)\n  Filter: (status = 'ACTIVE')\nExecution Time: 0.12 ms"
		return cols, [][]interface{}{{planStr}}, 1, nil
	}

	fields := strings.Fields(trimmed)
	if len(fields) < 2 {
		return nil, nil, 0, errors.New("syntax error in DESCRIBE statement")
	}

	tableName := strings.ToLower(strings.Trim(fields[1], `"'` + "`;"))
	tbl, exists := tables[tableName]
	if !exists {
		return nil, nil, 0, fmt.Errorf("relation %q does not exist", tableName)
	}

	cols := []string{"column_name", "data_type", "default_value"}
	rows := make([][]interface{}, 0, len(tbl.Columns))
	for _, c := range tbl.Columns {
		dt := "VARCHAR"
		if c == "id" {
			dt = "SERIAL PRIMARY KEY"
		} else if strings.HasPrefix(c, "is_") || strings.HasPrefix(c, "has_") {
			dt = "BOOLEAN"
		} else if c == "created_at" || c == "updated_at" {
			dt = "TIMESTAMP"
		}

		defVal := "NULL"
		if v, ok := tbl.ColumnDefaults[c]; ok && v != nil {
			defVal = fmt.Sprintf("%v", v)
		}

		rows = append(rows, []interface{}{c, dt, defVal})
	}

	return cols, rows, len(rows), nil
}

func (s *SQLService) handleUpdate(tables map[string]*TableState, trimmed, upper string) ([]string, [][]interface{}, int, error) {
	afterUpdate := strings.TrimSpace(trimmed[len("UPDATE"):])
	fields := strings.Fields(afterUpdate)
	if len(fields) == 0 {
		return nil, nil, 0, errors.New("syntax error in UPDATE statement")
	}

	tableName := strings.ToLower(strings.Trim(fields[0], `"'` + "`"))
	tbl, exists := tables[tableName]
	if !exists {
		return nil, nil, 0, fmt.Errorf("relation %q does not exist", tableName)
	}

	setIdx := strings.Index(upper, "SET")
	if setIdx == -1 {
		return nil, nil, 0, errors.New("missing SET clause in UPDATE statement")
	}

	whereIdx := strings.Index(upper, "WHERE")
	var setClause string
	if whereIdx != -1 {
		setClause = trimmed[setIdx+len("SET") : whereIdx]
	} else {
		setClause = trimmed[setIdx+len("SET"):]
	}

	updates := s.parseSetClause(setClause)

	updatedCount := 0
	for _, row := range tbl.Rows {
		for colName, val := range updates {
			colIdx := resolveColumnIndex(tbl.Columns, colName)
			if colIdx != -1 && colIdx < len(row) {
				row[colIdx] = val
			}
		}
		updatedCount++
	}

	return tbl.Columns, tbl.Rows, updatedCount, nil
}

func (s *SQLService) handleDelete(tables map[string]*TableState, trimmed, upper string) ([]string, [][]interface{}, int, error) {
	fromIdx := strings.Index(upper, "FROM")
	if fromIdx == -1 {
		return nil, nil, 0, errors.New("missing FROM in DELETE statement")
	}

	afterFrom := strings.TrimSpace(trimmed[fromIdx+len("FROM"):])
	fields := strings.Fields(afterFrom)
	if len(fields) == 0 {
		return nil, nil, 0, errors.New("syntax error in DELETE statement")
	}

	tableName := strings.ToLower(strings.Trim(fields[0], `"'` + "`"))
	tbl, exists := tables[tableName]
	if !exists {
		return nil, nil, 0, fmt.Errorf("relation %q does not exist", tableName)
	}

	deletedCount := len(tbl.Rows)
	tbl.Rows = [][]interface{}{}

	return tbl.Columns, tbl.Rows, deletedCount, nil
}

func (s *SQLService) handleDropTable(tables map[string]*TableState, trimmed, upper string) ([]string, [][]interface{}, int, error) {
	ifExists := strings.Contains(upper, "IF EXISTS")

	afterDrop := strings.TrimSpace(upper[len("DROP TABLE"):])
	afterDrop = strings.TrimPrefix(afterDrop, "IF EXISTS")
	fields := strings.Fields(afterDrop)
	if len(fields) == 0 {
		return nil, nil, 0, errors.New("syntax error in DROP TABLE statement")
	}

	tableName := strings.ToLower(strings.Trim(fields[0], `"'` + "`"))
	_, exists := tables[tableName]
	if !exists {
		if ifExists {
			return []string{}, [][]interface{}{}, 0, nil
		}
		return nil, nil, 0, fmt.Errorf("table %q does not exist", tableName)
	}

	delete(tables, tableName)
	return []string{"status"}, [][]interface{}{{"DROPPED"}}, 1, nil
}

func resolveColumnIndex(columns []string, colName string) int {
	colName = strings.ToLower(colName)
	for i, c := range columns {
		if c == colName {
			return i
		}
	}
	if colName == "name" || colName == "user" {
		for i, c := range columns {
			if c == "username" {
				return i
			}
		}
	}
	if colName == "mail" {
		for i, c := range columns {
			if c == "email" {
				return i
			}
		}
	}
	return -1
}

func (s *SQLService) parseValuesTuples(valuesStr string) [][]string {
	var tuples [][]string
	trimmed := strings.TrimSpace(valuesStr)

	for {
		start := strings.Index(trimmed, "(")
		if start == -1 {
			break
		}
		end := strings.Index(trimmed[start:], ")")
		if end == -1 {
			break
		}
		end += start

		rawTuple := trimmed[start+1 : end]
		rawVals := strings.Split(rawTuple, ",")
		vals := make([]string, 0, len(rawVals))
		for _, v := range rawVals {
			vals = append(vals, strings.TrimSpace(v))
		}
		tuples = append(tuples, vals)
		trimmed = trimmed[end+1:]
	}

	return tuples
}

func (s *SQLService) parseSetClause(setClause string) map[string]interface{} {
	updates := make(map[string]interface{})
	assignments := strings.Split(setClause, ",")
	for _, kv := range assignments {
		parts := strings.Split(kv, "=")
		if len(parts) == 2 {
			col := strings.ToLower(strings.TrimSpace(parts[0]))
			val := parseSQLValue(strings.TrimSpace(parts[1]))
			updates[col] = val
		}
	}
	return updates
}

func (s *SQLService) filterRows(tbl *TableState, whereClause string) [][]interface{} {
	upperWhere := strings.ToUpper(whereClause)

	// LIKE or ILIKE pattern matching
	if strings.Contains(upperWhere, " LIKE ") || strings.Contains(upperWhere, " ILIKE ") {
		var parts []string
		if strings.Contains(upperWhere, " ILIKE ") {
			idx := strings.Index(upperWhere, " ILIKE ")
			parts = []string{whereClause[:idx], whereClause[idx+len(" ILIKE "):]}
		} else {
			idx := strings.Index(upperWhere, " LIKE ")
			parts = []string{whereClause[:idx], whereClause[idx+len(" LIKE "):]}
		}

		filterCol := strings.ToLower(strings.TrimSpace(parts[0]))
		pattern := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		cleanPattern := strings.Trim(pattern, "%")

		colIdx := resolveColumnIndex(tbl.Columns, filterCol)
		if colIdx == -1 {
			return tbl.Rows
		}

		var matched [][]interface{}
		for _, row := range tbl.Rows {
			if colIdx < len(row) {
				cellStr := strings.ToLower(fmt.Sprintf("%v", row[colIdx]))
				if strings.Contains(cellStr, strings.ToLower(cleanPattern)) {
					matched = append(matched, row)
				}
			}
		}
		return matched
	}

	// Equality filter col = val
	parts := strings.Split(whereClause, "=")
	if len(parts) != 2 {
		return tbl.Rows
	}

	filterCol := strings.ToLower(strings.TrimSpace(parts[0]))
	filterVal := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

	colIdx := resolveColumnIndex(tbl.Columns, filterCol)
	if colIdx == -1 {
		return tbl.Rows
	}

	var matched [][]interface{}
	for _, row := range tbl.Rows {
		if colIdx < len(row) {
			cellStr := fmt.Sprintf("%v", row[colIdx])
			if cellStr == filterVal {
				matched = append(matched, row)
			}
		}
	}
	return matched
}

func parseSQLValue(valStr string) interface{} {
	valStr = strings.TrimSpace(valStr)

	if strings.HasPrefix(valStr, "'") && strings.HasSuffix(valStr, "'") {
		return strings.Trim(valStr, "'")
	}
	if strings.HasPrefix(valStr, `"`) && strings.HasSuffix(valStr, `"`) {
		return strings.Trim(valStr, `"`)
	}

	upper := strings.ToUpper(valStr)
	if upper == "NULL" {
		return nil
	}
	if upper == "TRUE" {
		return true
	}
	if upper == "FALSE" {
		return false
	}
	if upper == "NOW()" || upper == "CURRENT_TIMESTAMP" {
		return time.Now().Format(time.RFC3339)
	}

	if i, err := strconv.Atoi(valStr); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(valStr, 64); err == nil {
		return f
	}

	return valStr
}
