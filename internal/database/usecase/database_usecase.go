package usecase

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

type DatabaseUseCase interface {
	CreateDatabase(ctx context.Context, projectID, name string, engine domain.EngineType, storageGB int) (*domain.DatabaseInstance, string, error)
	GetDatabase(ctx context.Context, id string) (*domain.DatabaseInstance, error)
	ListDatabases(ctx context.Context, projectID string) ([]*domain.DatabaseInstance, error)
	StartDatabase(ctx context.Context, id string) error
	StopDatabase(ctx context.Context, id string) error
	DeleteDatabase(ctx context.Context, id string) error
	GetConnectionString(ctx context.Context, id string) (string, error)
}

type databaseUseCase struct {
	repo          domain.InstanceRepository
	driver        domain.ProvisionerDriver
	encryptionKey []byte
}

func NewDatabaseUseCase(
	repo domain.InstanceRepository,
	driver domain.ProvisionerDriver,
	encryptionKey string,
) DatabaseUseCase {
	keyBytes := []byte(encryptionKey)
	if len(keyBytes) < 32 {
		// Pad to 32 bytes
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	return &databaseUseCase{
		repo:          repo,
		driver:        driver,
		encryptionKey: keyBytes,
	}
}

func (u *databaseUseCase) CreateDatabase(ctx context.Context, projectID, name string, engine domain.EngineType, storageGB int) (*domain.DatabaseInstance, string, error) {
	if projectID == "" || name == "" {
		return nil, "", appErrors.New(appErrors.CodeInvalidInput, "projectID and name are required")
	}

	// Enforce 5 instances limit per project
	count, err := u.repo.CountByProjectID(ctx, projectID)
	if err != nil {
		return nil, "", err
	}
	if count >= 5 {
		return nil, "", appErrors.New(appErrors.CodeForbidden, "project database quota reached (max 5 databases)")
	}

	rawPassword := generateSecurePassword()
	encryptedPassword, err := security.Encrypt([]byte(rawPassword), u.encryptionKey)
	if err != nil {
		return nil, "", appErrors.Wrap(err, appErrors.CodeInternal, "failed to encrypt credentials")
	}

	dbName := fmt.Sprintf("db_%s", uuid.New().String()[:8])
	username := fmt.Sprintf("user_%s", uuid.New().String()[:8])
	port := randPort()

	instance := domain.NewDatabaseInstance(projectID, name, engine, "localhost", port, dbName, username, encryptedPassword, storageGB)
	if err := u.repo.Create(ctx, instance); err != nil {
		return nil, "", err
	}

	// Provision container
	containerID, err := u.driver.ProvisionInstance(ctx, domain.ProvisionParams{
		InstanceID: instance.ID,
		Engine:     instance.Engine,
		DBName:     instance.DBName,
		Username:   instance.Username,
		Password:   rawPassword,
		Port:       instance.Port,
	})
	if err != nil {
		instance.Status = domain.StatusFailed
		_ = u.repo.Update(ctx, instance)
		return nil, "", appErrors.Wrap(err, appErrors.CodeInternal, "failed to provision database container")
	}

	instance.ContainerID = containerID
	instance.Status = domain.StatusRunning
	if err := u.repo.Update(ctx, instance); err != nil {
		return nil, "", err
	}

	connStr := instance.FormatConnectionString(rawPassword)
	logger.Context(ctx).Info(fmt.Sprintf("Provisioned DB instance '%s' (%s) on port %d", instance.Name, instance.ID, instance.Port))
	return instance, connStr, nil
}

func (u *databaseUseCase) GetDatabase(ctx context.Context, id string) (*domain.DatabaseInstance, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *databaseUseCase) ListDatabases(ctx context.Context, projectID string) ([]*domain.DatabaseInstance, error) {
	return u.repo.ListByProjectID(ctx, projectID)
}

func (u *databaseUseCase) StartDatabase(ctx context.Context, id string) error {
	instance, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.driver.StartInstance(ctx, instance.ContainerID); err != nil {
		return appErrors.Wrap(err, appErrors.CodeInternal, "failed to start database container")
	}

	instance.Status = domain.StatusRunning
	return u.repo.Update(ctx, instance)
}

func (u *databaseUseCase) StopDatabase(ctx context.Context, id string) error {
	instance, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.driver.StopInstance(ctx, instance.ContainerID); err != nil {
		return appErrors.Wrap(err, appErrors.CodeInternal, "failed to stop database container")
	}

	instance.Status = domain.StatusStopped
	return u.repo.Update(ctx, instance)
}

func (u *databaseUseCase) DeleteDatabase(ctx context.Context, id string) error {
	instance, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if instance.ContainerID != "" {
		_ = u.driver.TerminateInstance(ctx, instance.ContainerID)
	}

	instance.Status = domain.StatusTerminated
	_ = u.repo.Update(ctx, instance)
	return u.repo.Delete(ctx, id)
}

func (u *databaseUseCase) GetConnectionString(ctx context.Context, id string) (string, error) {
	instance, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := security.Decrypt(instance.PasswordEncrypted, u.encryptionKey)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeInternal, "failed to decrypt database password")
	}

	return instance.FormatConnectionString(string(decryptedBytes)), nil
}

func generateSecurePassword() string {
	return fmt.Sprintf("Pass_%s!", uuid.New().String()[:16])
}

func randPort() int {
	return 15000 + rand.Intn(10000)
}
