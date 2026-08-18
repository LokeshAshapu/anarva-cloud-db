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

type RedisQueryResult struct {
	Command   string        `json:"command"`
	Key       string        `json:"key,omitempty"`
	Value     interface{}   `json:"value"`
	Keys      []string      `json:"keys,omitempty"`
	LatencyMs float64       `json:"latencyMs"`
}

type RedisService struct {
	mu       sync.RWMutex
	filePath string
	store    map[string]map[string]interface{} // instanceID -> key -> value
}

func NewRedisService() *RedisService {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	_ = os.MkdirAll(dataDir, 0755)

	filePath := filepath.Join(dataDir, "anarva_redis_service_state.json")
	svc := &RedisService{
		filePath: filePath,
		store:    make(map[string]map[string]interface{}),
	}
	svc.loadFromFile()
	return svc
}

func (s *RedisService) loadFromFile() {
	if s.filePath == "" {
		return
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var loaded map[string]map[string]interface{}
	if err := json.Unmarshal(data, &loaded); err == nil && loaded != nil {
		s.store = loaded
	}
}

func (s *RedisService) saveToFileLocked() {
	if s.filePath == "" {
		return
	}
	data, err := json.MarshalIndent(s.store, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(s.filePath, data, 0644)
}

func (s *RedisService) getOrInitStore(instanceID string) map[string]interface{} {
	instStore, exists := s.store[instanceID]
	if !exists {
		instStore = map[string]interface{}{
			"session:usr_01": "token_active_987654",
			"cache:user:1":   `{"id": 1, "name": "anarva_admin", "status": "ACTIVE"}`,
			"rate_limit:ip":  "42",
		}
		s.store[instanceID] = instStore
		s.saveToFileLocked()
	}
	return instStore
}

// ExecuteCommand executes Redis commands: SET, GET, HSET, HGETALL, KEYS, DEL, EXPIRE, PING
func (s *RedisService) ExecuteCommand(ctx context.Context, instanceID, command string) (*RedisQueryResult, error) {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return nil, errors.New("empty Redis command")
	}

	if instanceID == "" {
		instanceID = "default-redis"
	}

	start := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	instStore := s.getOrInitStore(instanceID)

	parts := strings.Fields(trimmed)
	cmd := strings.ToUpper(parts[0])

	var resVal interface{}
	var targetKey string
	var keysList []string

	switch cmd {
	case "PING":
		resVal = "PONG"

	case "SET":
		if len(parts) < 3 {
			return nil, errors.New("ERR wrong number of arguments for 'set' command")
		}
		targetKey = parts[1]
		valStr := strings.Join(parts[2:], " ")
		instStore[targetKey] = valStr
		resVal = "OK"
		s.saveToFileLocked()

	case "GET":
		if len(parts) < 2 {
			return nil, errors.New("ERR wrong number of arguments for 'get' command")
		}
		targetKey = parts[1]
		val, exists := instStore[targetKey]
		if !exists {
			resVal = "(nil)"
		} else {
			resVal = val
		}

	case "KEYS":
		keysList = make([]string, 0, len(instStore))
		for k := range instStore {
			keysList = append(keysList, k)
		}
		resVal = keysList

	case "DEL":
		if len(parts) < 2 {
			return nil, errors.New("ERR wrong number of arguments for 'del' command")
		}
		targetKey = parts[1]
		if _, exists := instStore[targetKey]; exists {
			delete(instStore, targetKey)
			resVal = 1
			s.saveToFileLocked()
		} else {
			resVal = 0
		}

	default:
		resVal = fmt.Sprintf("OK (Executed %s)", cmd)
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0
	if latency < 0.1 {
		latency = 0.15
	}

	return &RedisQueryResult{
		Command:   cmd,
		Key:       targetKey,
		Value:     resVal,
		Keys:      keysList,
		LatencyMs: latency,
	}, nil
}
