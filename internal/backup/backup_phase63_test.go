package backup_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	backupRepo "github.com/anarva-cloud/anarva-cloud-db/internal/backup/repository"
	backupUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/backup/usecase"
	storageDomain "github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
	storageProvider "github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
)

func TestPhase63_BackupStorageArchitecture(t *testing.T) {
	t.Run("1. Object Key Generation & Deterministic Tenancy", func(t *testing.T) {
		path := domain.GenerateBackupStoragePath("org-101", "proj-202", "db-303", "bak-404")
		expected := "backups/organizations/org-101/projects/proj-202/databases/db-303/backups/bak-404/backup.dump"
		assert.Equal(t, expected, path)
		assert.NoError(t, storageProvider.ValidateObjectKey(path))
	})

	t.Run("2. Path Traversal Rejection in Object Keys", func(t *testing.T) {
		maliciousKeys := []string{
			"../../etc/passwd",
			"backups/org-1/../secret.txt",
			"%2e%2e%2fmalicious",
		}
		for _, mk := range maliciousKeys {
			err := storageProvider.ValidateObjectKey(mk)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "STORAGE_SECURITY_RISK")
		}
	})

	t.Run("3. Full Lifecycle Streaming Backup & Restore to Object Storage", func(t *testing.T) {
		ctx := context.Background()
		memStorage := storageProvider.NewLocalStorageProvider(t.TempDir())

		// Setup in-memory mock repository using postgres_repository interface or mock
		memRepo := &MockBackupRepo{records: make(map[string]*domain.BackupRecord)}
		uc := backupUsecase.NewBackupUseCase(memRepo, memStorage, "anarva-test-bucket")

		// Create S3 bucket in storage provider
		_, err := memStorage.CreateBucket(ctx, &storageDomain.Bucket{Name: "anarva-test-bucket"})
		require.NoError(t, err)

		// A. Create Backup (Stream Upload)
		dumpPayload := "CREATE TABLE test_data (id INT); INSERT INTO test_data VALUES (42);"
		record, err := uc.CreateBackup(ctx, "org-alpha", "proj-alpha", "db-alpha", "prod-alpha-db", "daily-backup", domain.BackupManual, strings.NewReader(dumpPayload), int64(len(dumpPayload)))
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, record.Status)
		assert.Equal(t, domain.IntegrityValid, record.Integrity)
		assert.True(t, strings.HasPrefix(record.StoragePath, "backups/organizations/org-alpha/"))

		// B. Retrieve Metadata
		fetched, err := uc.GetBackup(ctx, "org-alpha", record.ID)
		require.NoError(t, err)
		assert.Equal(t, record.ID, fetched.ID)

		// C. Restore Stream Download
		rc, restoredRec, err := uc.RestoreBackup(ctx, "org-alpha", record.ID, "target-db-clone")
		require.NoError(t, err)
		defer rc.Close()

		restoredData, err := io.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, dumpPayload, string(restoredData))
		assert.Equal(t, record.DatabaseID, restoredRec.DatabaseID)

		// D. Delete Backup (Object + Metadata)
		err = uc.DeleteBackup(ctx, "org-alpha", record.ID)
		require.NoError(t, err)

		// E. Verify deletion
		_, err = uc.GetBackup(ctx, "org-alpha", record.ID)
		assert.Error(t, err)
	})

	t.Run("4. Security Test: Tenant Isolation Enforced", func(t *testing.T) {
		ctx := context.Background()
		memStorage := storageProvider.NewLocalStorageProvider(t.TempDir())
		memRepo := &MockBackupRepo{records: make(map[string]*domain.BackupRecord)}
		uc := backupUsecase.NewBackupUseCase(memRepo, memStorage, "anarva-test-bucket")

		_, err := memStorage.CreateBucket(ctx, &storageDomain.Bucket{Name: "anarva-test-bucket"})
		require.NoError(t, err)

		// Tenant A creates backup
		recordA, err := uc.CreateBackup(ctx, "tenant-A", "proj-A", "db-A", "db-A-name", "backup-A", domain.BackupManual, strings.NewReader("tenant-A-secret-dump"), 18)
		require.NoError(t, err)

		// Tenant B attempts to read Tenant A backup -> DENIED
		_, err = uc.GetBackup(ctx, "tenant-B", recordA.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authorization violation")

		// Tenant B attempts to restore Tenant A backup -> DENIED
		_, _, err = uc.RestoreBackup(ctx, "tenant-B", recordA.ID, "db-B-target")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authorization violation")

		// Tenant B attempts to delete Tenant A backup -> DENIED
		err = uc.DeleteBackup(ctx, "tenant-B", recordA.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authorization violation")
	})

	t.Run("5. Production Fail-Closed Safety", func(t *testing.T) {
		cfg := config.StorageConfig{Driver: "local"}
		_, err := storageProvider.NewStorageProvider(cfg, "production")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CONFIG_VALIDATION_FAILURE")
	})
}

type MockBackupRepo struct {
	records map[string]*domain.BackupRecord
}

func (m *MockBackupRepo) Create(ctx context.Context, snapshot *domain.BackupRecord) error {
	m.records[snapshot.ID] = snapshot
	return nil
}

func (m *MockBackupRepo) GetByID(ctx context.Context, id string) (*domain.BackupRecord, error) {
	rec, ok := m.records[id]
	if !ok || rec.Status == domain.StatusDeleted {
		return nil, assert.AnError
	}
	return rec, nil
}

func (m *MockBackupRepo) ListByDatabaseID(ctx context.Context, databaseID string) ([]*domain.BackupRecord, error) {
	var list []*domain.BackupRecord
	for _, rec := range m.records {
		if rec.DatabaseID == databaseID && rec.Status != domain.StatusDeleted {
			list = append(list, rec)
		}
	}
	return list, nil
}

func (m *MockBackupRepo) ListByProjectID(ctx context.Context, projectID string) ([]*domain.BackupRecord, error) {
	var list []*domain.BackupRecord
	for _, rec := range m.records {
		if rec.ProjectID == projectID && rec.Status != domain.StatusDeleted {
			list = append(list, rec)
		}
	}
	return list, nil
}

func (m *MockBackupRepo) Update(ctx context.Context, snapshot *domain.BackupRecord) error {
	m.records[snapshot.ID] = snapshot
	return nil
}

func (m *MockBackupRepo) Delete(ctx context.Context, id string) error {
	delete(m.records, id)
	return nil
}

var _ = backupRepo.NewBackupRepository
var _ = bytes.Buffer{}
var _ = time.Second
