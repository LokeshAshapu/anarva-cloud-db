package mapping

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type PostgresMappingRepository struct {
	db *gorm.DB
}

func NewPostgresMappingRepository(db *gorm.DB) MappingRepository {
	return &PostgresMappingRepository{db: db}
}

func (r *PostgresMappingRepository) SaveMapping(m *ProviderResourceMapping) error {
	if m == nil || m.AnarvaResourceID == "" {
		return errors.New("invalid provider resource mapping: anarvaResourceID is required")
	}
	if m.OrganizationID == "" {
		m.OrganizationID = "org-default"
	}
	m.UpdatedAt = time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return r.db.Save(m).Error
}

func (r *PostgresMappingRepository) GetMapping(anarvaResourceID string) (*ProviderResourceMapping, error) {
	var m ProviderResourceMapping
	err := r.db.Where("anarva_resource_id = ? AND deleted_at IS NULL", anarvaResourceID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("mapping for resource '%s' not found", anarvaResourceID)
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresMappingRepository) GetTenantScopedMapping(ctx context.Context, orgID, projID, anarvaResourceID string) (*ProviderResourceMapping, error) {
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

func (r *PostgresMappingRepository) ListMappings() ([]*ProviderResourceMapping, error) {
	var list []*ProviderResourceMapping
	err := r.db.Where("deleted_at IS NULL").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PostgresMappingRepository) ListTenantScopedMappings(ctx context.Context, orgID string) ([]*ProviderResourceMapping, error) {
	var list []*ProviderResourceMapping
	query := r.db.Where("deleted_at IS NULL")
	if orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	err := query.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PostgresMappingRepository) FindByProviderResourceID(ctx context.Context, provider, providerResourceID string) (*ProviderResourceMapping, error) {
	var m ProviderResourceMapping
	err := r.db.Where("provider = ? AND provider_resource_id = ? AND deleted_at IS NULL", provider, providerResourceID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("provider resource mapping for provider '%s' and resource ID '%s' not found", provider, providerResourceID)
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresMappingRepository) DeleteMapping(ctx context.Context, anarvaResourceID string) error {
	now := time.Now()
	res := r.db.Model(&ProviderResourceMapping{}).
		Where("anarva_resource_id = ? AND deleted_at IS NULL", anarvaResourceID).
		Update("deleted_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("mapping for resource '%s' not found", anarvaResourceID)
	}
	return nil
}
