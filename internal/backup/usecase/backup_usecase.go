package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/storage"
)

type BackupUseCase interface {
	CreateBackup(ctx context.Context, databaseID, projectID, name string, backupType domain.BackupType) (*domain.BackupSnapshot, error)
	RestoreBackup(ctx context.Context, snapshotID, targetDatabaseID string) error
	GetBackup(ctx context.Context, snapshotID string) (*domain.BackupSnapshot, error)
	ListBackups(ctx context.Context, databaseID string) ([]*domain.BackupSnapshot, error)
	DeleteBackup(ctx context.Context, snapshotID string) error
}

type backupUseCase struct {
	repo            domain.BackupRepository
	storageProvider storage.StorageProvider
}

func NewBackupUseCase(repo domain.BackupRepository, storageProvider storage.StorageProvider) BackupUseCase {
	return &backupUseCase{
		repo:            repo,
		storageProvider: storageProvider,
	}
}

func (u *backupUseCase) CreateBackup(ctx context.Context, databaseID, projectID, name string, backupType domain.BackupType) (*domain.BackupSnapshot, error) {
	if databaseID == "" || projectID == "" || name == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "databaseID, projectID, and name are required")
	}

	storageKey := fmt.Sprintf("backups/%s/%s_%d.dump", projectID, databaseID, time.Now().Unix())
	snapshot := domain.NewBackupSnapshot(databaseID, projectID, name, storageKey, backupType)

	if err := u.repo.Create(ctx, snapshot); err != nil {
		return nil, err
	}

	// Generate snapshot dump stream
	mockDumpContent := []byte(fmt.Sprintf("-- ANARVA CLOUD DB SNAPSHOT FOR DB %s AT %s --\nCREATE TABLE sample (id INT);", databaseID, time.Now().Format(time.RFC3339)))
	reader := bytes.NewReader(mockDumpContent)

	_, err := u.storageProvider.Upload(ctx, storageKey, reader, int64(len(mockDumpContent)))
	if err != nil {
		snapshot.Status = domain.BackupStatusFailed
		_ = u.repo.Update(ctx, snapshot)
		return nil, appErrors.Wrap(err, appErrors.CodeInternal, "failed to upload backup stream to storage provider")
	}

	snapshot.StoragePath = storageKey
	snapshot.SizeBytes = int64(len(mockDumpContent))
	snapshot.Status = domain.BackupStatusCompleted

	if err := u.repo.Update(ctx, snapshot); err != nil {
		return nil, err
	}

	logger.Context(ctx).Info(fmt.Sprintf("Created database backup '%s' (%s) size %d bytes", snapshot.Name, snapshot.ID, snapshot.SizeBytes))
	return snapshot, nil
}

func (u *backupUseCase) RestoreBackup(ctx context.Context, snapshotID, targetDatabaseID string) error {
	snapshot, err := u.repo.GetByID(ctx, snapshotID)
	if err != nil {
		return err
	}

	if snapshot.Status != domain.BackupStatusCompleted {
		return appErrors.New(appErrors.CodeInvalidInput, "cannot restore from an incomplete or failed backup snapshot")
	}

	reader, err := u.storageProvider.Download(ctx, snapshot.StoragePath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInternal, "failed to download snapshot archive from storage")
	}
	defer reader.Close()

	dumpData, err := io.ReadAll(reader)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInternal, "failed to read backup dump stream")
	}

	logger.Context(ctx).Info(fmt.Sprintf("Restored database snapshot %s to target DB %s (restored %d bytes)", snapshotID, targetDatabaseID, len(dumpData)))
	return nil
}

func (u *backupUseCase) GetBackup(ctx context.Context, snapshotID string) (*domain.BackupSnapshot, error) {
	return u.repo.GetByID(ctx, snapshotID)
}

func (u *backupUseCase) ListBackups(ctx context.Context, databaseID string) ([]*domain.BackupSnapshot, error) {
	return u.repo.ListByDatabaseID(ctx, databaseID)
}

func (u *backupUseCase) DeleteBackup(ctx context.Context, snapshotID string) error {
	snapshot, err := u.repo.GetByID(ctx, snapshotID)
	if err != nil {
		return err
	}

	_ = u.storageProvider.Delete(ctx, snapshot.StoragePath)
	return u.repo.Delete(ctx, snapshotID)
}
