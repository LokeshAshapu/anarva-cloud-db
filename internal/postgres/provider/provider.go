package provider

import (
	"context"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/domain"
)

// PostgresProvider interface defining capability contract for PostgreSQL provisioners
type PostgresProvider interface {
	Name() string
	SupportedVersions(ctx context.Context) ([]*domain.PostgresVersion, error)

	CreateInstance(ctx context.Context, inst *domain.PostgresInstance, adminPassword string) (*domain.PostgresInstance, error)
	UpdateInstance(ctx context.Context, inst *domain.PostgresInstance) (*domain.PostgresInstance, error)
	DeleteInstance(ctx context.Context, instanceID string) error

	StartInstance(ctx context.Context, instanceID string) error
	StopInstance(ctx context.Context, instanceID string) error
	RestartInstance(ctx context.Context, instanceID string) error

	GetInstance(ctx context.Context, instanceID string) (*domain.PostgresInstance, error)
	ListInstances(ctx context.Context, orgID, projectID string) ([]*domain.PostgresInstance, error)

	GetHealth(ctx context.Context, instanceID string) (*domain.DatabaseHealth, error)
	GetMetrics(ctx context.Context, instanceID string) ([]*domain.DatabaseHealth, error)
	GetLogs(ctx context.Context, instanceID string, limit int) ([]*domain.PostgresLogEntry, error)

	CreateDatabase(ctx context.Context, instanceID, dbName, owner string) (*domain.PostgresDatabase, error)
	DeleteDatabase(ctx context.Context, instanceID, dbName string) error

	CreateUser(ctx context.Context, instanceID, username string, role domain.UserRole, password string) (*domain.PostgresUser, error)
	DeleteUser(ctx context.Context, instanceID, username string) error
	RotateCredentials(ctx context.Context, instanceID, username, newPassword string) error

	CreateBackup(ctx context.Context, instanceID, backupName string) (string, error)
	RestoreBackup(ctx context.Context, instanceID, backupID, targetInstanceName string) (*domain.PostgresInstance, error)

	GetConnectionInfo(ctx context.Context, instanceID string) (*domain.ConnectionInfo, error)
	ScaleInstance(ctx context.Context, instanceID string, cpu float64, memoryMB, storageGB int) (*domain.PostgresInstance, error)
}
