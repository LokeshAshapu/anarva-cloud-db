package usecase

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	storageProvider "github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type mockBackupRepo struct {
	snapshots map[string]*domain.BackupRecord
}

func newMockBackupRepo() domain.BackupRepository {
	return &mockBackupRepo{snapshots: make(map[string]*domain.BackupRecord)}
}

func (m *mockBackupRepo) Create(ctx context.Context, snapshot *domain.BackupRecord) error {
	m.snapshots[snapshot.ID] = snapshot
	return nil
}

func (m *mockBackupRepo) GetByID(ctx context.Context, id string) (*domain.BackupRecord, error) {
	if s, ok := m.snapshots[id]; ok {
		return s, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "snapshot not found")
}

func (m *mockBackupRepo) ListByDatabaseID(ctx context.Context, databaseID string) ([]*domain.BackupRecord, error) {
	var list []*domain.BackupRecord
	for _, s := range m.snapshots {
		if s.DatabaseID == databaseID {
			list = append(list, s)
		}
	}
	return list, nil
}

func (m *mockBackupRepo) ListByProjectID(ctx context.Context, projectID string) ([]*domain.BackupRecord, error) {
	var list []*domain.BackupRecord
	for _, s := range m.snapshots {
		if s.ProjectID == projectID {
			list = append(list, s)
		}
	}
	return list, nil
}

func (m *mockBackupRepo) Update(ctx context.Context, snapshot *domain.BackupRecord) error {
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

	provider := storageProvider.NewLocalStorageProvider(tempDir)
	repo := newMockBackupRepo()
	return NewBackupUseCase(repo, provider, "anarva-test-bucket"), tempDir
}

func TestBackupAndRestoreLifecycle(t *testing.T) {
	uc, tempDir := setupBackupUseCase(t)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()

	orgID := "org-uuid-111"
	databaseID := "db-uuid-555"
	projectID := "proj-uuid-777"

	// Create Backup
	snapshot, err := uc.CreateBackup(ctx, orgID, projectID, databaseID, "production-db", "Nightly Production Backup", domain.BackupSnapshot, nil, 0)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusCompleted, snapshot.Status)
	assert.True(t, snapshot.SizeBytes > 0)

	// List Backups
	list, err := uc.ListBackups(ctx, orgID, databaseID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Restore Backup Stream to Target DB
	targetDatabaseID := "db-uuid-888-restored"
	rc, restoredSnapshot, err := uc.RestoreBackup(ctx, orgID, snapshot.ID, targetDatabaseID)
	require.NoError(t, err)
	defer rc.Close()

	restoredBytes, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.True(t, len(restoredBytes) > 0)
	assert.Equal(t, snapshot.ID, restoredSnapshot.ID)

	// Delete Backup
	err = uc.DeleteBackup(ctx, orgID, snapshot.ID)
	require.NoError(t, err)

	_, err = uc.GetBackup(ctx, orgID, snapshot.ID)
	assert.Error(t, err)
}
