package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type MQLQueryResult struct {
	Collection string                   `json:"collection"`
	Operation  string                   `json:"operation"`
	Documents  []map[string]interface{} `json:"documents"`
	DocCount   int                      `json:"docCount"`
	LatencyMs  float64                  `json:"latencyMs"`
}

type CollectionState struct {
	Name      string                   `json:"name"`
	Documents []map[string]interface{} `json:"documents"`
}

type MQLService struct {
	mu          sync.RWMutex
	filePath    string
	collections map[string]map[string]*CollectionState // instanceID -> collectionName -> CollectionState
}

func NewMQLService() *MQLService {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	_ = os.MkdirAll(dataDir, 0755)

	filePath := filepath.Join(dataDir, "anarva_mql_service_state.json")
	svc := &MQLService{
		filePath:    filePath,
		collections: make(map[string]map[string]*CollectionState),
	}
	svc.loadFromFile()
	return svc
}

func (s *MQLService) loadFromFile() {
	if s.filePath == "" {
		return
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var loaded map[string]map[string]*CollectionState
	if err := json.Unmarshal(data, &loaded); err == nil && loaded != nil {
		s.collections = loaded
	}
}

func (s *MQLService) saveToFileLocked() {
	if s.filePath == "" {
		return
	}
	data, err := json.MarshalIndent(s.collections, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(s.filePath, data, 0644)
}

func (s *MQLService) getOrInitCollections(instanceID string) map[string]*CollectionState {
	colls, exists := s.collections[instanceID]
	if !exists {
		colls = make(map[string]*CollectionState)
		nowStr := time.Now().Format(time.RFC3339)
		colls["users"] = &CollectionState{
			Name: "users",
			Documents: []map[string]interface{}{
				{"_id": "usr_01", "name": "Alice Johnson", "email": "alice@anarva.io", "role": "ADMIN", "created_at": nowStr},
				{"_id": "usr_02", "name": "Bob Smith", "email": "bob@anarva.io", "role": "DEVELOPER", "created_at": nowStr},
			},
		}
		s.collections[instanceID] = colls
		s.saveToFileLocked()
	}
	return colls
}

// ExecuteMQL executes MongoDB Query Language statements: e.g. db.users.find() or db.users.insertOne({ name: "Charlie" })
func (s *MQLService) ExecuteMQL(ctx context.Context, instanceID, mql string) (*MQLQueryResult, error) {
	trimmed := strings.TrimSpace(mql)
	if trimmed == "" {
		return nil, errors.New("empty MQL query statement")
	}

	if instanceID == "" {
		instanceID = "default-mongodb"
	}

	start := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	colls := s.getOrInitCollections(instanceID)

	// Syntax: db.collection.operation(...)
	if !strings.HasPrefix(trimmed, "db.") {
		// SHOW COLLECTIONS fallback
		if strings.ToUpper(trimmed) == "SHOW COLLECTIONS" || strings.ToUpper(trimmed) == "SHOW DBS" {
			docs := make([]map[string]interface{}, 0, len(colls))
			for name := range colls {
				docs = append(docs, map[string]interface{}{"collection": name, "docCount": len(colls[name].Documents)})
			}
			return &MQLQueryResult{
				Collection: "system",
				Operation:  "SHOW_COLLECTIONS",
				Documents:  docs,
				DocCount:   len(docs),
				LatencyMs:  0.45,
			}, nil
		}
		return nil, fmt.Errorf("invalid MQL syntax: statement must start with 'db.<collection>.<operation>()'")
	}

	afterDb := trimmed[3:]
	idxDot := strings.Index(afterDb, ".")
	if idxDot == -1 {
		return nil, errors.New("invalid MQL syntax: missing operation")
	}

	collName := strings.ToLower(afterDb[:idxDot])
	afterColl := afterDb[idxDot+1:]

	idxParen := strings.Index(afterColl, "(")
	if idxParen == -1 {
		return nil, errors.New("invalid MQL syntax: missing parenthesis")
	}

	operation := strings.TrimSpace(afterColl[:idxParen])
	argsBlock := afterColl[idxParen+1:]
	if idxClose := strings.LastIndex(argsBlock, ")"); idxClose != -1 {
		argsBlock = argsBlock[:idxClose]
	}

	collState, exists := colls[collName]
	if !exists {
		collState = &CollectionState{
			Name:      collName,
			Documents: []map[string]interface{}{},
		}
		colls[collName] = collState
	}

	var docs []map[string]interface{}
	var docCount int

	switch strings.ToLower(operation) {
	case "find":
		docs = collState.Documents
		docCount = len(docs)

	case "insertone", "insert":
		var newDoc map[string]interface{}
		if err := json.Unmarshal([]byte(argsBlock), &newDoc); err != nil {
			// Fallback string parsing
			newDoc = map[string]interface{}{
				"_id":        fmt.Sprintf("doc_%d", time.Now().UnixNano()),
				"data":       argsBlock,
				"created_at": time.Now().Format(time.RFC3339),
			}
		}
		if _, hasId := newDoc["_id"]; !hasId {
			newDoc["_id"] = fmt.Sprintf("doc_%d", time.Now().UnixNano())
		}
		collState.Documents = append(collState.Documents, newDoc)
		docs = []map[string]interface{}{newDoc}
		docCount = 1
		s.saveToFileLocked()

	case "deletemany", "drop":
		docCount = len(collState.Documents)
		collState.Documents = []map[string]interface{}{}
		docs = []map[string]interface{}{}
		s.saveToFileLocked()

	default:
		docs = collState.Documents
		docCount = len(docs)
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0
	if latency < 0.2 {
		latency = 0.45
	}

	return &MQLQueryResult{
		Collection: collName,
		Operation:  operation,
		Documents:  docs,
		DocCount:   docCount,
		LatencyMs:  latency,
	}, nil
}
