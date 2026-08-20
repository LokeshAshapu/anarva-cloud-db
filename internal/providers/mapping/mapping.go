package mapping

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ProviderResourceMapping struct {
	AnarvaResourceID     string     `json:"anarvaResourceId" gorm:"primaryKey;column:anarva_resource_id;type:varchar(255)"`
	OrganizationID       string     `json:"organizationId" gorm:"column:organization_id;type:varchar(255);index"`
	ProjectID            string     `json:"projectId,omitempty" gorm:"column:project_id;type:varchar(255);index"`
	Provider             string     `json:"provider" gorm:"column:provider;type:varchar(100);index"`
	ProviderResourceID   string     `json:"providerResourceId" gorm:"column:provider_resource_id;type:varchar(255);index"`
	ProviderResourceType string     `json:"providerResourceType" gorm:"column:provider_resource_type;type:varchar(100)"`
	Region               string     `json:"region" gorm:"column:region;type:varchar(100)"`
	Zone                 string     `json:"zone,omitempty" gorm:"column:zone;type:varchar(100)"`
	Status               string     `json:"status" gorm:"column:status;type:varchar(50)"`
	Managed              bool       `json:"managed" gorm:"column:managed"`
	CreatedAt            time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt            time.Time  `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt            *time.Time `json:"deletedAt,omitempty" gorm:"column:deleted_at;index"`
}

func (ProviderResourceMapping) TableName() string {
	return "provider_resource_mappings"
}

type MappingRepository interface {
	SaveMapping(m *ProviderResourceMapping) error
	GetMapping(anarvaResourceID string) (*ProviderResourceMapping, error)
	GetTenantScopedMapping(ctx context.Context, orgID, projID, anarvaResourceID string) (*ProviderResourceMapping, error)
	ListMappings() ([]*ProviderResourceMapping, error)
	ListTenantScopedMappings(ctx context.Context, orgID string) ([]*ProviderResourceMapping, error)
	FindByProviderResourceID(ctx context.Context, provider, providerResourceID string) (*ProviderResourceMapping, error)
	DeleteMapping(ctx context.Context, anarvaResourceID string) error
}

type InMemoryMappingRepository struct {
	mu       sync.RWMutex
	mappings map[string]*ProviderResourceMapping
}

func NewMappingRepository() MappingRepository {
	return NewInMemoryMappingRepository()
}

func NewInMemoryMappingRepository() *InMemoryMappingRepository {
	return &InMemoryMappingRepository{
		mappings: make(map[string]*ProviderResourceMapping),
	}
}

func (r *InMemoryMappingRepository) SaveMapping(m *ProviderResourceMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m.OrganizationID == "" {
		m.OrganizationID = "org-default"
	}
	m.UpdatedAt = time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	r.mappings[m.AnarvaResourceID] = m
	return nil
}

func (r *InMemoryMappingRepository) GetMapping(anarvaResourceID string) (*ProviderResourceMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if m, ok := r.mappings[anarvaResourceID]; ok && m.DeletedAt == nil {
		return m, nil
	}
	return nil, fmt.Errorf("mapping for resource '%s' not found", anarvaResourceID)
}

func (r *InMemoryMappingRepository) GetTenantScopedMapping(ctx context.Context, orgID, projID, anarvaResourceID string) (*ProviderResourceMapping, error) {
	m, err := r.GetMapping(anarvaResourceID)
	if err != nil {
		return nil, err
	}
	if orgID != "" && m.OrganizationID != "" && m.OrganizationID != orgID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Organization '%s' is prohibited from accessing resource mapping for '%s'", orgID, anarvaResourceID)
	}
	if projID != "" && m.ProjectID != "" && m.ProjectID != projID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Project '%s' is prohibited from accessing resource mapping for '%s'", projID, anarvaResourceID)
	}
	return m, nil
}

func (r *InMemoryMappingRepository) ListMappings() ([]*ProviderResourceMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*ProviderResourceMapping
	for _, m := range r.mappings {
		if m.DeletedAt == nil {
			res = append(res, m)
		}
	}
	return res, nil
}

func (r *InMemoryMappingRepository) ListTenantScopedMappings(ctx context.Context, orgID string) ([]*ProviderResourceMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*ProviderResourceMapping
	for _, m := range r.mappings {
		if m.DeletedAt == nil && (orgID == "" || m.OrganizationID == orgID) {
			res = append(res, m)
		}
	}
	return res, nil
}

func (r *InMemoryMappingRepository) FindByProviderResourceID(ctx context.Context, provider, providerResourceID string) (*ProviderResourceMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.mappings {
		if m.DeletedAt == nil && m.Provider == provider && m.ProviderResourceID == providerResourceID {
			return m, nil
		}
	}
	return nil, fmt.Errorf("provider resource mapping for provider '%s' and resource ID '%s' not found", provider, providerResourceID)
}

func (r *InMemoryMappingRepository) DeleteMapping(ctx context.Context, anarvaResourceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.mappings[anarvaResourceID]; ok {
		now := time.Now()
		m.DeletedAt = &now
		return nil
	}
	return fmt.Errorf("mapping for resource '%s' not found", anarvaResourceID)
}
