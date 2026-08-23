package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/domain"
)

type PostgresInstanceRepository interface {
	Create(ctx context.Context, inst *domain.PostgresInstance) error
	GetByID(ctx context.Context, id string) (*domain.PostgresInstance, error)
	GetByIDForTenant(ctx context.Context, orgID, projID, id string) (*domain.PostgresInstance, error)
	ListByProject(ctx context.Context, orgID, projectID string) ([]*domain.PostgresInstance, error)
	ListByOrganization(ctx context.Context, orgID string) ([]*domain.PostgresInstance, error)
	Update(ctx context.Context, inst *domain.PostgresInstance) error
	Delete(ctx context.Context, id string) error
}

type GormPostgresInstanceRepository struct {
	db *gorm.DB
}

func NewGormPostgresInstanceRepository(db *gorm.DB) PostgresInstanceRepository {
	return &GormPostgresInstanceRepository{db: db}
}

func (r *GormPostgresInstanceRepository) Create(ctx context.Context, inst *domain.PostgresInstance) error {
	if inst == nil || inst.ID == "" {
		return errors.New("invalid postgres instance: ID is required")
	}
	if inst.OrganizationID == "" {
		inst.OrganizationID = "org-default"
	}
	if inst.ProjectID == "" {
		inst.ProjectID = "proj-default"
	}
	inst.UpdatedAt = time.Now()
	if inst.CreatedAt.IsZero() {
		inst.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Save(inst).Error
}

func (r *GormPostgresInstanceRepository) GetByID(ctx context.Context, id string) (*domain.PostgresInstance, error) {
	var inst domain.PostgresInstance
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&inst).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("database instance '%s' not found", id)
		}
		return nil, err
	}
	return &inst, nil
}

func (r *GormPostgresInstanceRepository) GetByIDForTenant(ctx context.Context, orgID, projID, id string) (*domain.PostgresInstance, error) {
	inst, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if orgID != "" && inst.OrganizationID != "" && inst.OrganizationID != orgID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Organization '%s' is prohibited from accessing database instance '%s'", orgID, id)
	}
	if projID != "" && inst.ProjectID != "" && inst.ProjectID != projID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Project '%s' is prohibited from accessing database instance '%s'", projID, id)
	}
	return inst, nil
}

func (r *GormPostgresInstanceRepository) ListByProject(ctx context.Context, orgID, projectID string) ([]*domain.PostgresInstance, error) {
	var list []*domain.PostgresInstance
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	err := query.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GormPostgresInstanceRepository) ListByOrganization(ctx context.Context, orgID string) ([]*domain.PostgresInstance, error) {
	var list []*domain.PostgresInstance
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	err := query.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GormPostgresInstanceRepository) Update(ctx context.Context, inst *domain.PostgresInstance) error {
	if inst == nil || inst.ID == "" {
		return errors.New("invalid postgres instance: ID is required")
	}
	inst.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(inst).Error
}

func (r *GormPostgresInstanceRepository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&domain.PostgresInstance{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("database instance '%s' not found", id)
	}
	return nil
}
