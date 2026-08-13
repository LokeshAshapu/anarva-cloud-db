package service

import (
	"context"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/repository"

	activityStream "github.com/anarva-cloud/anarva-cloud-db/internal/activity"
)

type MySQLService struct {
	repo      *repository.MySQLRepository
	provider  provider.MySQLProvider
	actStream *activityStream.Stream
}

func NewMySQLService(
	repo *repository.MySQLRepository,
	prov provider.MySQLProvider,
	actStream *activityStream.Stream,
) *MySQLService {
	return &MySQLService{
		repo:      repo,
		provider:  prov,
		actStream: actStream,
	}
}

func (s *MySQLService) CreateInstance(ctx context.Context, orgID, projectID, name, version, regionID, networkID string, acuCount, storageGB int) (*domain.MySQLInstance, error) {
	instID := fmt.Sprintf("mysql-%d", time.Now().UnixNano())
	if version == "" {
		version = "8.0"
	}

	inst := &domain.MySQLInstance{
		ID:                instID,
		OrganizationID:    orgID,
		ProjectID:         projectID,
		Name:              name,
		Version:           version,
		Status:            domain.StatusCreating,
		RegionID:          regionID,
		ZoneID:            fmt.Sprintf("%sa", regionID),
		CPU:               acuCount / 2,
		MemoryMB:          acuCount * 1024,
		StorageGB:         storageGB,
		StorageType:       "GP3_SSD",
		NetworkID:         networkID,
		SubnetID:          "sub-01",
		AvailabilityMode:  domain.AvailabilitySingle,
		BackupMode:        "AUTOMATED_DAILY",
		MaintenanceWindow: "Sun:03:00-Sun:04:00",
		RealityLabel:      "LOCAL_MYSQL (LIMITED_CAPABILITIES)",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	created, err := s.provider.CreateInstance(ctx, inst)
	if err != nil {
		return nil, err
	}

	_ = s.repo.SaveInstance(created)

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: orgID,
			ProjectID:      projectID,
			ResourceID:     name,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.EventAction("MYSQL_CREATED"),
			Timestamp:      time.Now(),
		})
	}

	return created, nil
}

func (s *MySQLService) GetInstance(ctx context.Context, id string) (*domain.MySQLInstance, error) {
	return s.repo.GetInstance(id)
}

func (s *MySQLService) ListInstances(ctx context.Context, orgID, projectID string) ([]*domain.MySQLInstance, error) {
	return s.repo.ListInstances(orgID, projectID)
}

func (s *MySQLService) DeleteInstance(ctx context.Context, id string) error {
	inst, err := s.repo.GetInstance(id)
	if err != nil {
		return err
	}

	if err := s.provider.DeleteInstance(ctx, id); err != nil {
		return err
	}

	inst.Status = domain.StatusDeleted
	_ = s.repo.SaveInstance(inst)

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: inst.OrganizationID,
			ProjectID:      inst.ProjectID,
			ResourceID:     inst.Name,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.EventAction("MYSQL_DELETED"),
			Timestamp:      time.Now(),
		})
	}
	return nil
}

func (s *MySQLService) StartInstance(ctx context.Context, id string) error {
	return s.provider.StartInstance(ctx, id)
}

func (s *MySQLService) StopInstance(ctx context.Context, id string) error {
	return s.provider.StopInstance(ctx, id)
}

func (s *MySQLService) RestartInstance(ctx context.Context, id string) error {
	return s.provider.RestartInstance(ctx, id)
}

func (s *MySQLService) GetHealth(ctx context.Context, id string) (*domain.MySQLHealth, error) {
	return s.provider.GetHealth(ctx, id)
}

func (s *MySQLService) GetConnectionInfo(ctx context.Context, id string) (map[string]interface{}, error) {
	return s.provider.GetConnectionInfo(ctx, id)
}
