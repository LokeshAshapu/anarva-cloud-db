package usecase

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/storage"
)

type mockBackupRepo struct {
	snapshots map[string]*domain.BackupSnapshot
}

func newMockBackupRepo() domain.BackupRepository {
	return &mockBackupRepo{snapshots: make(map[string]*domain.BackupSnapshot)}
}

func (m *mockBackupRepo) Create(ctx context.Context, snapshot *domain.BackupSnapshot) error {
	m.snapshots[snapshot.ID] = snapshot
	return nil
}

func (m *mockBackupRepo) GetByID(ctx context.Context, id string) (*domain.BackupSnapshot, error) {
	if s, ok := m.snapshots[id]; ok {
		return s, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "snapshot not found")
}

func (m *mockBackupRepo) ListByDatabaseID(ctx context.Context, databaseID string) ([]*domain.BackupSnapshot, error) {
	var list []*domain.BackupSnapshot
	for _, s := range m.snapshots {
		if s.DatabaseID == databaseID {
			list = append(list, s)
		}
	}
	return list, nil
}

func (m *mockBackupRepo) ListByProjectID(ctx context.Context, projectID string) ([]*domain.BackupSnapshot, error) {
	var list []*domain.BackupSnapshot
	for _, s := range m.snapshots {
		if s.ProjectID == projectID {
			list = append(list, s)
		}
	}
	return list, nil
}

func (m *mockBackupRepo) Update(ctx context.Context, snapshot *domain.BackupSnapshot) error {
	m.snapshots[snapshot.ID] = snapshot
	return nil
}

func (m *mockBackupRepo) Delete(ctx context.Context, id string) error {
	delete(m.snapshots, id)
	return nil
}

func setupBackupUseCase(t *testing.T) (BackupUseCase, string) {
	tempDir, err := os.MkdirTemp("", "anarva_backup_uc_test_*")
	require.NoError(t, err)

	provider, err := storage.NewLocalStorageProvider(tempDir)
	require.NoError(t, err)

	repo := newMockBackupRepo()
	return NewBackupUseCase(repo, provider), tempDir
}

func TestBackupAndRestoreLifecycle(t *testing.T) {
	uc, tempDir := setupBackupUseCase(t)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()

	databaseID := "db-uuid-555"
	projectID := "proj-uuid-777"

	// Create Backup
	snapshot, err := uc.CreateBackup(ctx, databaseID, projectID, "Nightly Production Backup", domain.BackupTypeSnapshot)
	require.NoError(t, err)
	assert.Equal(t, domain.BackupStatusCompleted, snapshot.Status)
	assert.True(t, snapshot.SizeBytes > 0)

	// List Backups
	list, err := uc.ListBackups(ctx, databaseID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Restore Backup to Target DB
	targetDatabaseID := "db-uuid-888-restored"
	err = uc.RestoreBackup(ctx, snapshot.ID, targetDatabaseID)
	require.NoError(t, err)

	// Delete Backup
	err = uc.DeleteBackup(ctx, snapshot.ID)
	require.NoError(t, err)

	_, err = uc.GetBackup(ctx, snapshot.ID)
	assert.Error(t, err)
}
