package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type WorkloadMetadata struct {
	WorkloadID     string    `json:"workloadId"`
	ResourceID     string    `json:"resourceId"`
	ContainerID    string    `json:"containerId"`
	Name           string    `json:"name"`
	Image          string    `json:"image"`
	Command        []string  `json:"command,omitempty"`
	VCPU           float64   `json:"vcpu"`
	MemoryMB       int       `json:"memoryMb"`
	OrgID          string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	Region         string    `json:"region"`
	Status         string    `json:"status"` // RUNNING, STOPPED, DELETED, FAILED
	Health         string    `json:"health"` // HEALTHY, DEGRADED, UNAVAILABLE
	PrivateIP      string    `json:"privateIp,omitempty"`
	PublicIP       string    `json:"publicIp,omitempty"`
	IdempotencyKey string    `json:"idempotencyKey,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type MetadataStore interface {
	Save(m *WorkloadMetadata) error
	Get(workloadID string) (*WorkloadMetadata, error)
	GetByIdempotencyKey(key string) (*WorkloadMetadata, error)
	List() ([]*WorkloadMetadata, error)
	Delete(workloadID string) error
}

type FileMetadataStore struct {
	mu       sync.RWMutex
	filePath string
	records  map[string]*WorkloadMetadata
}

func NewFileMetadataStore(filePath string) (*FileMetadataStore, error) {
	if filePath == "" {
		filePath = "worker_metadata.json"
	}

	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	store := &FileMetadataStore{
		filePath: filePath,
		records:  make(map[string]*WorkloadMetadata),
	}

	// Load existing state if file exists
	if _, err := os.Stat(filePath); err == nil {
		data, readErr := os.ReadFile(filePath)
		if readErr == nil && len(data) > 0 {
			var loaded map[string]*WorkloadMetadata
			if jsonErr := json.Unmarshal(data, &loaded); jsonErr == nil {
				store.records = loaded
			}
		}
	}

	return store, nil
}

func (s *FileMetadataStore) saveToFileLocked() error {
	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal worker metadata: %w", err)
	}

	tmpPath := s.filePath + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temporary metadata file: %w", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write temporary metadata file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to sync temporary metadata file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close temporary metadata file: %w", err)
	}

	if err := os.Rename(tmpPath, s.filePath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to atomically rename metadata file: %w", err)
	}

	_ = os.Chmod(s.filePath, 0600)
	return nil
}

func (s *FileMetadataStore) Save(m *WorkloadMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if m.WorkloadID == "" {
		return errors.New("workloadId is required")
	}

	m.UpdatedAt = time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	s.records[m.WorkloadID] = m
	return s.saveToFileLocked()
}

func (s *FileMetadataStore) Get(workloadID string) (*WorkloadMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.records[workloadID]
	if !ok || m.Status == "DELETED" {
		return nil, errors.New("workload not found")
	}
	return m, nil
}

func (s *FileMetadataStore) GetByIdempotencyKey(key string) (*WorkloadMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if key == "" {
		return nil, nil
	}

	for _, m := range s.records {
		if m.IdempotencyKey == key && m.Status != "DELETED" {
			return m, nil
		}
	}
	return nil, nil
}

func (s *FileMetadataStore) List() ([]*WorkloadMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*WorkloadMetadata
	for _, m := range s.records {
		if m.Status != "DELETED" {
			list = append(list, m)
		}
	}
	return list, nil
}

func (s *FileMetadataStore) Delete(workloadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, ok := s.records[workloadID]
	if !ok {
		return errors.New("workload not found")
	}

	m.Status = "DELETED"
	m.Health = "UNAVAILABLE"
	m.UpdatedAt = time.Now()

	s.records[workloadID] = m
	return s.saveToFileLocked()
}
