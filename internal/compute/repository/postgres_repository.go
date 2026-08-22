package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
)

type PostgresComputeRepository struct {
	db *gorm.DB
}

func NewPostgresComputeRepository(db *gorm.DB) domain.ComputeRepository {
	return &PostgresComputeRepository{db: db}
}

func (r *PostgresComputeRepository) Create(ctx context.Context, inst *domain.ComputeInstance) error {
	if inst == nil || inst.ID == "" {
		return errors.New("invalid compute instance: ID is required")
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

func (r *PostgresComputeRepository) GetByID(ctx context.Context, id string) (*domain.ComputeInstance, error) {
	var inst domain.ComputeInstance
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&inst).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("compute instance '%s' not found", id)
		}
		return nil, err
	}
	return &inst, nil
}

func (r *PostgresComputeRepository) GetTenantScopedByID(ctx context.Context, orgID, projID, id string) (*domain.ComputeInstance, error) {
	inst, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if orgID != "" && inst.OrganizationID != "" && inst.OrganizationID != orgID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Organization '%s' is prohibited from accessing compute instance '%s'", orgID, id)
	}
	if projID != "" && inst.ProjectID != "" && inst.ProjectID != projID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Project '%s' is prohibited from accessing compute instance '%s'", projID, id)
	}
	return inst, nil
}

func (r *PostgresComputeRepository) ListByProjectID(ctx context.Context, projectID string) ([]*domain.ComputeInstance, error) {
	var list []*domain.ComputeInstance
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	err := query.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PostgresComputeRepository) Update(ctx context.Context, inst *domain.ComputeInstance) error {
	if inst == nil || inst.ID == "" {
		return errors.New("invalid compute instance: ID is required")
	}
	inst.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(inst).Error
}

func (r *PostgresComputeRepository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&domain.ComputeInstance{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("compute instance '%s' not found", id)
	}
	return nil
}

type PostgresVolumeRepository struct {
	db *gorm.DB
}

func NewPostgresVolumeRepository(db *gorm.DB) domain.VolumeRepository {
	return &PostgresVolumeRepository{db: db}
}

func (r *PostgresVolumeRepository) CreateVolume(ctx context.Context, vol *domain.Volume) error {
	if vol == nil || vol.ID == "" {
		return errors.New("invalid volume: ID is required")
	}
	if vol.OrganizationID == "" {
		vol.OrganizationID = "org-default"
	}
	if vol.ProjectID == "" {
		vol.ProjectID = "proj-default"
	}
	vol.UpdatedAt = time.Now()
	if vol.CreatedAt.IsZero() {
		vol.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Save(vol).Error
}

func (r *PostgresVolumeRepository) GetVolumeByID(ctx context.Context, id string) (*domain.Volume, error) {
	var vol domain.Volume
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&vol).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("volume '%s' not found", id)
		}
		return nil, err
	}
	return &vol, nil
}

func (r *PostgresVolumeRepository) GetTenantScopedVolumeByID(ctx context.Context, orgID, projID, id string) (*domain.Volume, error) {
	vol, err := r.GetVolumeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if orgID != "" && vol.OrganizationID != "" && vol.OrganizationID != orgID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Organization '%s' is prohibited from accessing volume '%s'", orgID, id)
	}
	if projID != "" && vol.ProjectID != "" && vol.ProjectID != projID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Project '%s' is prohibited from accessing volume '%s'", projID, id)
	}
	return vol, nil
}

func (r *PostgresVolumeRepository) ListVolumesByProjectID(ctx context.Context, projectID string) ([]*domain.Volume, error) {
	var list []*domain.Volume
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	err := query.Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PostgresVolumeRepository) UpdateVolume(ctx context.Context, vol *domain.Volume) error {
	if vol == nil || vol.ID == "" {
		return errors.New("invalid volume: ID is required")
	}
	vol.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(vol).Error
}

func (r *PostgresVolumeRepository) DeleteVolume(ctx context.Context, id string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&domain.Volume{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("volume '%s' not found", id)
	}
	return nil
}
