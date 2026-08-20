package provider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/usecase"
	storageProvider "github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
)

type BackupProvider interface {
	CreateBackup(ctx context.Context, req *domain.BackupRecord) (*domain.BackupRecord, error)
	GetBackup(ctx context.Context, id, orgID string) (*domain.BackupRecord, error)
	ListBackups(ctx context.Context, orgID, dbID string) ([]*domain.BackupRecord, error)
	DeleteBackup(ctx context.Context, id, orgID string) error
	CreateRestoreJob(ctx context.Context, job *domain.RestoreJob) (*domain.RestoreJob, error)
	GetRestoreJob(ctx context.Context, id, orgID string) (*domain.RestoreJob, error)
}

type ControlPlaneBackupProvider struct {
	mu          sync.RWMutex
	uc          usecase.BackupUseCase
	backups     map[string]*domain.BackupRecord
	restores    map[string]*domain.RestoreJob
	storageProv storageProvider.ObjectStorageProvider
}

func NewControlPlaneBackupProvider(sProv storageProvider.ObjectStorageProvider) *ControlPlaneBackupProvider {
	p := &ControlPlaneBackupProvider{
		backups:     make(map[string]*domain.BackupRecord),
		restores:    make(map[string]*domain.RestoreJob),
		storageProv: sProv,
	}
	p.seedDefaults()
	return p
}

func NewControlPlaneBackupProviderWithUseCase(uc usecase.BackupUseCase, sProv storageProvider.ObjectStorageProvider) *ControlPlaneBackupProvider {
	p := &ControlPlaneBackupProvider{
		uc:          uc,
		backups:     make(map[string]*domain.BackupRecord),
		restores:    make(map[string]*domain.RestoreJob),
		storageProv: sProv,
	}
	p.seedDefaults()
	return p
}

func (p *ControlPlaneBackupProvider) seedDefaults() {
	now := time.Now()
	b1 := &domain.BackupRecord{
		ID:                  "bak-prod-101",
		ResourceID:          domain.GenerateBackupARNV("ap-hyderabad-1", "proj-default", "production-db", "daily-snapshot-20260810"),
		OrganizationID:      "org-default",
		ProjectID:           "proj-default",
		DatabaseID:          "res-db-prod-1",
		DatabaseName:        "production-db",
		Name:                "daily-snapshot-20260810",
		Type:                domain.BackupAutomated,
		Status:              domain.StatusVerified,
		Integrity:           domain.IntegrityValid,
		SizeBytes:           14589000,
		RetentionDays:       7,
		StorageBucket:       "anarva-production-backups",
		StoragePath:         "backups/organizations/org-default/projects/proj-default/databases/res-db-prod-1/backups/bak-prod-101/backup.dump",
		Checksum:            "sha256-e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		StartedAt:           now.Add(-24 * time.Hour),
		CompletedAt:         timePtr(now.Add(-23*time.Hour + 45*time.Minute)),
		ExpiresAt:           now.Add(6 * 24 * time.Hour),
		CreatedAt:           now.Add(-24 * time.Hour),
		UpdatedAt:           now,
	}
	p.backups[b1.ID] = b1
}

func (p *ControlPlaneBackupProvider) CreateBackup(ctx context.Context, req *domain.BackupRecord) (*domain.BackupRecord, error) {
	if p.uc != nil {
		return p.uc.CreateBackup(ctx, req.OrganizationID, req.ProjectID, req.DatabaseID, req.DatabaseName, req.Name, req.Type, nil, req.SizeBytes)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if req.ID == "" {
		req.ID = fmt.Sprintf("bak-%d", time.Now().UnixNano())
	}
	if req.ResourceID == "" {
		req.ResourceID = domain.GenerateBackupARNV("ap-hyderabad-1", req.ProjectID, req.DatabaseName, req.Name)
	}
	now := time.Now()
	req.Status = domain.StatusCompleted
	req.Integrity = domain.IntegrityValid
	req.StartedAt = now
	req.CompletedAt = &now
	req.CreatedAt = now
	req.UpdatedAt = now
	if req.RetentionDays == 0 {
		req.RetentionDays = 7
	}
	req.ExpiresAt = now.Add(time.Duration(req.RetentionDays*24) * time.Hour)
	req.StorageBucket = "anarva-production-backups"
	req.StoragePath = domain.GenerateBackupStoragePath(req.OrganizationID, req.ProjectID, req.DatabaseID, req.ID)
	req.Checksum = fmt.Sprintf("sha256-%d", time.Now().UnixNano())

	p.backups[req.ID] = req
	return req, nil
}

func (p *ControlPlaneBackupProvider) GetBackup(ctx context.Context, id, orgID string) (*domain.BackupRecord, error) {
	if p.uc != nil {
		return p.uc.GetBackup(ctx, orgID, id)
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	b, ok := p.backups[id]
	if !ok || b.Status == domain.StatusDeleted {
		return nil, fmt.Errorf("backup record not found")
	}
	if orgID != "" && b.OrganizationID != orgID {
		return nil, fmt.Errorf("authorization violation: cross-tenant access denied")
	}
	return b, nil
}

func (p *ControlPlaneBackupProvider) ListBackups(ctx context.Context, orgID, dbID string) ([]*domain.BackupRecord, error) {
	if p.uc != nil {
		return p.uc.ListBackups(ctx, orgID, dbID)
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*domain.BackupRecord
	for _, b := range p.backups {
		if b.Status == domain.StatusDeleted {
			continue
		}
		if orgID != "" && b.OrganizationID != orgID {
			continue
		}
		if dbID != "" && b.DatabaseID != dbID {
			continue
		}
		result = append(result, b)
	}
	return result, nil
}

func (p *ControlPlaneBackupProvider) DeleteBackup(ctx context.Context, id, orgID string) error {
	if p.uc != nil {
		return p.uc.DeleteBackup(ctx, orgID, id)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	b, ok := p.backups[id]
	if !ok {
		return fmt.Errorf("backup not found")
	}
	if orgID != "" && b.OrganizationID != orgID {
		return fmt.Errorf("authorization violation: cross-tenant access denied")
	}
	b.Status = domain.StatusDeleted
	b.UpdatedAt = time.Now()
	return nil
}

func (p *ControlPlaneBackupProvider) CreateRestoreJob(ctx context.Context, job *domain.RestoreJob) (*domain.RestoreJob, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if job.ID == "" {
		job.ID = fmt.Sprintf("rst-%d", time.Now().UnixNano())
	}

	// Verify backup existence & tenant isolation if usecase is connected
	if p.uc != nil {
		rc, rec, err := p.uc.RestoreBackup(ctx, job.OrganizationID, job.BackupID, job.SourceDatabaseID)
		if err != nil {
			return nil, fmt.Errorf("restore failed: %w", err)
		}
		_ = rc.Close()
		job.SourceDatabaseID = rec.DatabaseID
	}

	now := time.Now()
	job.Status = "COMPLETED"
	job.StartedAt = now
	job.CompletedAt = &now

	p.restores[job.ID] = job
	return job, nil
}

func (p *ControlPlaneBackupProvider) GetRestoreJob(ctx context.Context, id, orgID string) (*domain.RestoreJob, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	j, ok := p.restores[id]
	if !ok {
		return nil, fmt.Errorf("restore job not found")
	}
	if orgID != "" && j.OrganizationID != orgID {
		return nil, fmt.Errorf("authorization violation: cross-tenant access denied")
	}
	return j, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
