package mapping

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ProviderResourceMapping struct {
	AnarvaResourceID     string    `json:"anarvaResourceId"`
	OrganizationID       string    `json:"organizationId"`
	ProjectID            string    `json:"projectId,omitempty"`
	Provider             string    `json:"provider"`
	ProviderResourceID   string    `json:"providerResourceId"`
	ProviderResourceType string    `json:"providerResourceType"`
	Region               string    `json:"region"`
	Zone                 string    `json:"zone,omitempty"`
	Status               string    `json:"status"`
	Managed              bool      `json:"managed"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type MappingRepository struct {
	mu       sync.RWMutex
	mappings map[string]*ProviderResourceMapping
}

func NewMappingRepository() *MappingRepository {
	return &MappingRepository{
		mappings: make(map[string]*ProviderResourceMapping),
	}
}

func (r *MappingRepository) SaveMapping(m *ProviderResourceMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m.OrganizationID == "" {
		m.OrganizationID = "org-default"
	}
	m.UpdatedAt = time.Now()
	r.mappings[m.AnarvaResourceID] = m
	return nil
}

func (r *MappingRepository) GetMapping(anarvaResourceID string) (*ProviderResourceMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if m, ok := r.mappings[anarvaResourceID]; ok {
		return m, nil
	}
	return nil, fmt.Errorf("mapping for resource '%s' not found", anarvaResourceID)
}

func (r *MappingRepository) GetTenantScopedMapping(ctx context.Context, orgID, projID, anarvaResourceID string) (*ProviderResourceMapping, error) {
	m, err := r.GetMapping(anarvaResourceID)
	if err != nil {
		return nil, err
	}
	if orgID != "" && m.OrganizationID != "" && m.OrganizationID != orgID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Organization '%s' is prohibited from accessing resource mapping for '%s'", orgID, anarvaResourceID)
	}
	return m, nil
}

func (r *MappingRepository) ListMappings() ([]*ProviderResourceMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*ProviderResourceMapping
	for _, m := range r.mappings {
		res = append(res, m)
	}
	return res, nil
}

func (r *MappingRepository) ListTenantScopedMappings(ctx context.Context, orgID string) ([]*ProviderResourceMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*ProviderResourceMapping
	for _, m := range r.mappings {
		if orgID == "" || m.OrganizationID == orgID {
			res = append(res, m)
		}
	}
	return res, nil
}
