package repository

import (
	"context"
	"errors"
	"sync"

	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type instanceRepository struct {
	db *gorm.DB
}

func NewInstanceRepository(db *gorm.DB) domain.InstanceRepository {
	return &instanceRepository{db: db}
}

func (r *instanceRepository) Create(ctx context.Context, instance *domain.DatabaseInstance) error {
	if err := r.db.WithContext(ctx).Create(instance).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to record database instance")
	}
	return nil
}

func (r *instanceRepository) GetByID(ctx context.Context, id string) (*domain.DatabaseInstance, error) {
	var instance domain.DatabaseInstance
	if err := r.db.WithContext(ctx).First(&instance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "database instance not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch database instance")
	}
	return &instance, nil
}

func (r *instanceRepository) ListByProjectID(ctx context.Context, projectID string) ([]*domain.DatabaseInstance, error) {
	var instances []*domain.DatabaseInstance
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&instances).Error; err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list database instances")
	}
	return instances, nil
}

func (r *instanceRepository) Update(ctx context.Context, instance *domain.DatabaseInstance) error {
	if err := r.db.WithContext(ctx).Save(instance).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update database instance")
	}
	return nil
}

func (r *instanceRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&domain.DatabaseInstance{}, "id = ?", id).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete database instance")
	}
	return nil
}

func (r *instanceRepository) CountByProjectID(ctx context.Context, projectID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.DatabaseInstance{}).Where("project_id = ? AND status != ?", projectID, domain.StatusTerminated).Count(&count).Error; err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to count database instances")
	}
	return count, nil
}

type memoryInstanceRepo struct {
	mu        sync.RWMutex
	instances map[string]*domain.DatabaseInstance
}

func NewMemoryInstanceRepository() domain.InstanceRepository {
	return &memoryInstanceRepo{instances: make(map[string]*domain.DatabaseInstance)}
}

func (m *memoryInstanceRepo) Create(ctx context.Context, instance *domain.DatabaseInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[instance.ID] = instance
	return nil
}

func (m *memoryInstanceRepo) GetByID(ctx context.Context, id string) (*domain.DatabaseInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if inst, ok := m.instances[id]; ok {
		return inst, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "database instance not found")
}

func (m *memoryInstanceRepo) ListByProjectID(ctx context.Context, projectID string) ([]*domain.DatabaseInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.DatabaseInstance
	for _, inst := range m.instances {
		if inst.ProjectID == projectID {
			list = append(list, inst)
		}
	}
	return list, nil
}

func (m *memoryInstanceRepo) Update(ctx context.Context, instance *domain.DatabaseInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[instance.ID] = instance
	return nil
}

func (m *memoryInstanceRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.instances, id)
	return nil
}

func (m *memoryInstanceRepo) CountByProjectID(ctx context.Context, projectID string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, inst := range m.instances {
		if inst.ProjectID == projectID && inst.Status != domain.StatusTerminated {
			count++
		}
	}
	return count, nil
}
