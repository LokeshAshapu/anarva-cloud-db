package service

import (
	"context"
	"errors"
	"fmt"
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
	Name           string
	Columns        []string
	ColumnDefaults map[string]interface{}
	Rows           [][]interface{}
}

type SQLService struct {
	mu             sync.RWMutex
	instanceTables map[string]map[string]*TableState // instanceID -> tableName -> TableState
}

func NewSQLService() *SQLService {
	return &SQLService{
		instanceTables: make(map[string]map[string]*TableState),
	}
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
		s.instanceTables[instanceID] = tables
	}
	return tables
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
	case strings.HasPrefix(upper, "CREATE TABLE"):
		var err error
		columns, rows, rowCount, err = s.handleCreateTable(tables, trimmed, upper)
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

	default: // SELECT, SHOW, or other queries
		var err error
		columns, rows, rowCount, err = s.handleSelect(tables, trimmed, upper)
		if err != nil {
			return nil, err
		}
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

					// Check for DEFAULT value clause
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

		// Auto-populate unmapped columns from ColumnDefaults or dynamic rules
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
	if upper == "SHOW TABLES" || strings.Contains(upper, "INFORMATION_SCHEMA.TABLES") {
		cols := []string{"table_name"}
		rows := make([][]interface{}, 0, len(tables))
		for name := range tables {
			rows = append(rows, []interface{}{name})
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
		return nil, nil, 0, fmt.Errorf("relation %q does not exist", tableName)
	}

	whereIdx := strings.Index(upper, "WHERE")
	limitIdx := strings.Index(upper, "LIMIT")

	var filteredRows [][]interface{}
	if whereIdx != -1 {
		var whereClause string
		if limitIdx != -1 && limitIdx > whereIdx {
			whereClause = trimmed[whereIdx+len("WHERE") : limitIdx]
		} else {
			whereClause = trimmed[whereIdx+len("WHERE"):]
		}
		filteredRows = s.filterRows(tbl, whereClause)
	} else {
		filteredRows = tbl.Rows
	}

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
	return []string{}, [][]interface{}{}, 0, nil
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
