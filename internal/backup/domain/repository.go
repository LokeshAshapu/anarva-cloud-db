package domain

import "context"

type BackupRepository interface {
	Create(ctx context.Context, snapshot *BackupRecord) error
	GetByID(ctx context.Context, id string) (*BackupRecord, error)
	ListByDatabaseID(ctx context.Context, databaseID string) ([]*BackupRecord, error)
	ListByProjectID(ctx context.Context, projectID string) ([]*BackupRecord, error)
	Update(ctx context.Context, snapshot *BackupRecord) error
	Delete(ctx context.Context, id string) error
}
