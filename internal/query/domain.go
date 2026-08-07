package query

type StatementType string

const (
	StatementSelect StatementType = "SELECT"
	StatementInsert StatementType = "INSERT"
	StatementUpdate StatementType = "UPDATE"
	StatementDelete StatementType = "DELETE"
	StatementDDL    StatementType = "DDL"
	StatementUnknown StatementType = "UNKNOWN"
)

type QueryRequest struct {
	DatabaseID string        `json:"database_id"`
	SQL        string        `json:"sql"`
	Parameters []interface{} `json:"parameters,omitempty"`
}

type ColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type QueryResult struct {
	Columns         []ColumnInfo             `json:"columns"`
	Rows            []map[string]interface{} `json:"rows"`
	RowsAffected    int64                    `json:"rows_affected"`
	ExecutionTimeMs float64                  `json:"execution_time_ms"`
}

type ParsedQuery struct {
	RawSQL        string
	StatementType StatementType
	IsReadOnly    bool
	TableNames    []string
}
