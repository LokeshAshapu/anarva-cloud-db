package driver

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
)

type mockProvisionerDriver struct {
	mu         sync.RWMutex
	containers map[string]string // containerID -> status
}

func NewMockProvisionerDriver() domain.ProvisionerDriver {
	return &mockProvisionerDriver{
		containers: make(map[string]string),
	}
}

func (m *mockProvisionerDriver) ProvisionInstance(ctx context.Context, params domain.ProvisionParams) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	containerID := fmt.Sprintf("container-%s", uuid.New().String()[:8])
	m.containers[containerID] = "RUNNING"
	return containerID, nil
}

func (m *mockProvisionerDriver) StartInstance(ctx context.Context, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.containers[containerID]; !ok {
		return fmt.Errorf("container %s not found", containerID)
	}
	m.containers[containerID] = "RUNNING"
	return nil
}

func (m *mockProvisionerDriver) StopInstance(ctx context.Context, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.containers[containerID]; !ok {
		return fmt.Errorf("container %s not found", containerID)
	}
	m.containers[containerID] = "STOPPED"
	return nil
}

func (m *mockProvisionerDriver) TerminateInstance(ctx context.Context, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.containers, containerID)
	return nil
}

func (m *mockProvisionerDriver) CheckHealth(ctx context.Context, containerID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, ok := m.containers[containerID]
	if !ok || status != "RUNNING" {
		return false, nil
	}
	return true, nil
}
