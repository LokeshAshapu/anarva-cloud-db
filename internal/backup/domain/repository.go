package domain

import "context"

type BackupRepository interface {
	Create(ctx context.Context, snapshot *BackupSnapshot) error
	GetByID(ctx context.Context, id string) (*BackupSnapshot, error)
	ListByDatabaseID(ctx context.Context, databaseID string) ([]*BackupSnapshot, error)
	ListByProjectID(ctx context.Context, projectID string) ([]*BackupSnapshot, error)
	Update(ctx context.Context, snapshot *BackupSnapshot) error
	Delete(ctx context.Context, id string) error
}
