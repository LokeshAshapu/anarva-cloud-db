package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type backupRepository struct {
	db *gorm.DB
}

func NewBackupRepository(db *gorm.DB) domain.BackupRepository {
	return &backupRepository{db: db}
}

func (r *backupRepository) Create(ctx context.Context, snapshot *domain.BackupSnapshot) error {
	if err := r.db.WithContext(ctx).Create(snapshot).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to record backup snapshot metadata")
	}
	return nil
}

func (r *backupRepository) GetByID(ctx context.Context, id string) (*domain.BackupSnapshot, error) {
	var snapshot domain.BackupSnapshot
	if err := r.db.WithContext(ctx).First(&snapshot, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "backup snapshot not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch backup snapshot")
	}
	return &snapshot, nil
}

func (r *backupRepository) ListByDatabaseID(ctx context.Context, databaseID string) ([]*domain.BackupSnapshot, error) {
	var snapshots []*domain.BackupSnapshot
	if err := r.db.WithContext(ctx).Where("database_id = ?", databaseID).Order("created_at desc").Find(&snapshots).Error; err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list database backups")
	}
	return snapshots, nil
}

func (r *backupRepository) ListByProjectID(ctx context.Context, projectID string) ([]*domain.BackupSnapshot, error) {
	var snapshots []*domain.BackupSnapshot
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at desc").Find(&snapshots).Error; err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list project backups")
	}
	return snapshots, nil
}

func (r *backupRepository) Update(ctx context.Context, snapshot *domain.BackupSnapshot) error {
	if err := r.db.WithContext(ctx).Save(snapshot).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update backup snapshot metadata")
	}
	return nil
}

func (r *backupRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&domain.BackupSnapshot{}, "id = ?", id).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete backup snapshot metadata")
	}
	return nil
}
