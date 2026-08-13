package repository

import (
	"fmt"
	"sync"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/domain"
)

type MySQLRepository struct {
	mu        sync.RWMutex
	instances map[string]*domain.MySQLInstance
	databases map[string]*domain.MySQLDatabase
	users     map[string]*domain.MySQLUser
	backups   map[string]*domain.MySQLBackup
}

func NewMySQLRepository() *MySQLRepository {
	return &MySQLRepository{
		instances: make(map[string]*domain.MySQLInstance),
		databases: make(map[string]*domain.MySQLDatabase),
		users:     make(map[string]*domain.MySQLUser),
		backups:   make(map[string]*domain.MySQLBackup),
	}
}

func (r *MySQLRepository) SaveInstance(inst *domain.MySQLInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instances[inst.ID] = inst
	return nil
}

func (r *MySQLRepository) GetInstance(id string) (*domain.MySQLInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if inst, ok := r.instances[id]; ok {
		return inst, nil
	}
	return nil, fmt.Errorf("MySQL instance '%s' not found", id)
}

func (r *MySQLRepository) ListInstances(orgID, projectID string) ([]*domain.MySQLInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.MySQLInstance
	for _, inst := range r.instances {
		if inst.Status != domain.StatusDeleted {
			res = append(res, inst)
		}
	}
	return res, nil
}

func (r *MySQLRepository) SaveDatabase(db *domain.MySQLDatabase) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.databases[db.ID] = db
	return nil
}

func (r *MySQLRepository) SaveUser(user *domain.MySQLUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *MySQLRepository) SaveBackup(backup *domain.MySQLBackup) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backups[backup.ID] = backup
	return nil
}
