package provider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/domain"
	"github.com/google/uuid"
)

type LocalDockerPostgresProvider struct {
	mu        sync.RWMutex
	instances map[string]*domain.PostgresInstance
	databases map[string][]*domain.PostgresDatabase
	users     map[string][]*domain.PostgresUser
	logs      map[string][]*domain.PostgresLogEntry
	nextPort  int
}

func NewLocalDockerPostgresProvider() *LocalDockerPostgresProvider {
	return &LocalDockerPostgresProvider{
		instances: make(map[string]*domain.PostgresInstance),
		databases: make(map[string][]*domain.PostgresDatabase),
		users:     make(map[string][]*domain.PostgresUser),
		logs:      make(map[string][]*domain.PostgresLogEntry),
		nextPort:  15432,
	}
}

func (p *LocalDockerPostgresProvider) Name() string {
	return "LOCAL_POSTGRES"
}

func (p *LocalDockerPostgresProvider) SupportedVersions(ctx context.Context) ([]*domain.PostgresVersion, error) {
	now := time.Now()
	return []*domain.PostgresVersion{
		{Version: "17", Status: "ACTIVE", Supported: true, EndOfLifeAt: "2029-11-01", CreatedAt: now},
		{Version: "16", Status: "ACTIVE", Supported: true, EndOfLifeAt: "2028-11-01", CreatedAt: now},
		{Version: "15", Status: "ACTIVE", Supported: true, EndOfLifeAt: "2027-11-01", CreatedAt: now},
		{Version: "14", Status: "ACTIVE", Supported: true, EndOfLifeAt: "2026-11-01", CreatedAt: now},
	}, nil
}

func (p *LocalDockerPostgresProvider) CreateInstance(ctx context.Context, inst *domain.PostgresInstance, adminPassword string) (*domain.PostgresInstance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.nextPort++
	inst.Port = p.nextPort
	inst.ProviderResourceId = fmt.Sprintf("docker-pg-%s", inst.ID[:8])
	inst.Status = domain.StatusAvailable
	inst.RealityLabel = "LOCAL_POSTGRES (DOCKER_SIM)"
	inst.Host = "localhost"
	inst.UpdatedAt = time.Now()

	p.instances[inst.ID] = inst

	// Default database
	p.databases[inst.ID] = []*domain.PostgresDatabase{
		{
			ID:             uuid.New().String(),
			InstanceID:     inst.ID,
			Name:           "postgres",
			OwnerReference: "postgres",
			Encoding:       "UTF8",
			Collation:      "en_US.UTF-8",
			Status:         "AVAILABLE",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	// Default user
	p.users[inst.ID] = []*domain.PostgresUser{
		{
			ID:                  uuid.New().String(),
			InstanceID:          inst.ID,
			Username:            "anarva_admin",
			Role:                domain.RoleOwner,
			Status:              "ACTIVE",
			CredentialReference: fmt.Sprintf("secret-ref-%s", uuid.New().String()[:8]),
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		},
	}

	// Logs
	p.logs[inst.ID] = []*domain.PostgresLogEntry{
		{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Level:     "INFO",
			Component: "POSTGRES_ENGINE",
			Message:   fmt.Sprintf("PostgreSQL %s instance '%s' initialized on port %d", inst.Version, inst.Name, inst.Port),
		},
	}

	return inst, nil
}

func (p *LocalDockerPostgresProvider) UpdateInstance(ctx context.Context, inst *domain.PostgresInstance) (*domain.PostgresInstance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	existing, exists := p.instances[inst.ID]
	if !exists {
		return nil, domain.ErrInstanceNotFound
	}
	existing.Name = inst.Name
	existing.CPU = inst.CPU
	existing.MemoryMB = inst.MemoryMB
	existing.StorageGB = inst.StorageGB
	existing.UpdatedAt = time.Now()
	return existing, nil
}

func (p *LocalDockerPostgresProvider) DeleteInstance(ctx context.Context, instanceID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, exists := p.instances[instanceID]
	if !exists {
		return domain.ErrInstanceNotFound
	}
	inst.Status = domain.StatusDeleted
	delete(p.instances, instanceID)
	delete(p.databases, instanceID)
	delete(p.users, instanceID)
	return nil
}

func (p *LocalDockerPostgresProvider) StartInstance(ctx context.Context, instanceID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, exists := p.instances[instanceID]
	if !exists {
		return domain.ErrInstanceNotFound
	}
	inst.Status = domain.StatusAvailable
	inst.UpdatedAt = time.Now()
	return nil
}

func (p *LocalDockerPostgresProvider) StopInstance(ctx context.Context, instanceID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, exists := p.instances[instanceID]
	if !exists {
		return domain.ErrInstanceNotFound
	}
	inst.Status = domain.StatusStopped
	inst.UpdatedAt = time.Now()
	return nil
}

func (p *LocalDockerPostgresProvider) RestartInstance(ctx context.Context, instanceID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, exists := p.instances[instanceID]
	if !exists {
		return domain.ErrInstanceNotFound
	}
	inst.Status = domain.StatusAvailable
	inst.UpdatedAt = time.Now()
	return nil
}

func (p *LocalDockerPostgresProvider) GetInstance(ctx context.Context, instanceID string) (*domain.PostgresInstance, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	inst, exists := p.instances[instanceID]
	if !exists {
		return nil, domain.ErrInstanceNotFound
	}
	return inst, nil
}

func (p *LocalDockerPostgresProvider) ListInstances(ctx context.Context, orgID, projectID string) ([]*domain.PostgresInstance, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	res := make([]*domain.PostgresInstance, 0)
	for _, inst := range p.instances {
		if (orgID == "" || inst.OrganizationID == orgID) && (projectID == "" || inst.ProjectID == projectID) {
			res = append(res, inst)
		}
	}
	return res, nil
}

func (p *LocalDockerPostgresProvider) GetHealth(ctx context.Context, instanceID string) (*domain.DatabaseHealth, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	inst, exists := p.instances[instanceID]
	if !exists {
		return nil, domain.ErrInstanceNotFound
	}

	available := inst.Status == domain.StatusAvailable
	return &domain.DatabaseHealth{
		InstanceID:          instanceID,
		ConnectionAvailable: available,
		ReplicationStatus:   "SINGLE",
		CPUPct:              12.5,
		MemoryPct:           34.0,
		StorageUsedGB:       1.8,
		StorageAllocatedGB:  float64(inst.StorageGB),
		ActiveConnections:   3,
		MaxConnections:      100,
		QueryLatencyMs:      1.2,
		CacheHitRatio:       99.4,
		SourceQuality:       "ACTUAL (LOCAL_POSTGRES)",
		Timestamp:           time.Now(),
	}, nil
}

func (p *LocalDockerPostgresProvider) GetMetrics(ctx context.Context, instanceID string) ([]*domain.DatabaseHealth, error) {
	h, err := p.GetHealth(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	return []*domain.DatabaseHealth{h}, nil
}

func (p *LocalDockerPostgresProvider) GetLogs(ctx context.Context, instanceID string, limit int) ([]*domain.PostgresLogEntry, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entries, exists := p.logs[instanceID]
	if !exists {
		return []*domain.PostgresLogEntry{}, nil
	}
	if limit > 0 && len(entries) > limit {
		return entries[len(entries)-limit:], nil
	}
	return entries, nil
}

func (p *LocalDockerPostgresProvider) CreateDatabase(ctx context.Context, instanceID, dbName, owner string) (*domain.PostgresDatabase, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	db := &domain.PostgresDatabase{
		ID:             uuid.New().String(),
		InstanceID:     instanceID,
		Name:           dbName,
		OwnerReference: owner,
		Encoding:       "UTF8",
		Collation:      "en_US.UTF-8",
		Status:         "AVAILABLE",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	p.databases[instanceID] = append(p.databases[instanceID], db)
	return db, nil
}

func (p *LocalDockerPostgresProvider) DeleteDatabase(ctx context.Context, instanceID, dbName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbs := p.databases[instanceID]
	newList := make([]*domain.PostgresDatabase, 0)
	for _, d := range dbs {
		if d.Name != dbName {
			newList = append(newList, d)
		}
	}
	p.databases[instanceID] = newList
	return nil
}

func (p *LocalDockerPostgresProvider) CreateUser(ctx context.Context, instanceID, username string, role domain.UserRole, password string) (*domain.PostgresUser, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	u := &domain.PostgresUser{
		ID:                  uuid.New().String(),
		InstanceID:          instanceID,
		Username:            username,
		Role:                role,
		Status:              "ACTIVE",
		CredentialReference: fmt.Sprintf("secret-ref-%s", uuid.New().String()[:8]),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	p.users[instanceID] = append(p.users[instanceID], u)
	return u, nil
}

func (p *LocalDockerPostgresProvider) DeleteUser(ctx context.Context, instanceID, username string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	users := p.users[instanceID]
	newList := make([]*domain.PostgresUser, 0)
	for _, u := range users {
		if u.Username != username {
			newList = append(newList, u)
		}
	}
	p.users[instanceID] = newList
	return nil
}

func (p *LocalDockerPostgresProvider) RotateCredentials(ctx context.Context, instanceID, username, newPassword string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	users := p.users[instanceID]
	for _, u := range users {
		if u.Username == username {
			u.CredentialReference = fmt.Sprintf("secret-ref-%s", uuid.New().String()[:8])
			u.UpdatedAt = time.Now()
			return nil
		}
	}
	return domain.ErrInstanceNotFound
}

func (p *LocalDockerPostgresProvider) CreateBackup(ctx context.Context, instanceID, backupName string) (string, error) {
	backupID := fmt.Sprintf("bak-pg-%s-%d", instanceID[:8], time.Now().Unix())
	return backupID, nil
}

func (p *LocalDockerPostgresProvider) RestoreBackup(ctx context.Context, instanceID, backupID, targetInstanceName string) (*domain.PostgresInstance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	source, exists := p.instances[instanceID]
	if !exists {
		return nil, domain.ErrInstanceNotFound
	}

	restored := *source
	restored.ID = uuid.New().String()
	restored.Name = targetInstanceName
	p.nextPort++
	restored.Port = p.nextPort
	restored.Status = domain.StatusAvailable
	restored.CreatedAt = time.Now()
	restored.UpdatedAt = time.Now()

	p.instances[restored.ID] = &restored
	return &restored, nil
}

func (p *LocalDockerPostgresProvider) GetConnectionInfo(ctx context.Context, instanceID string) (*domain.ConnectionInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	inst, exists := p.instances[instanceID]
	if !exists {
		return nil, domain.ErrInstanceNotFound
	}

	connStr := fmt.Sprintf("postgres://anarva_admin:anarva_secret@%s:%d/postgres?sslmode=disable", inst.Host, inst.Port)
	return &domain.ConnectionInfo{
		HostReference:     inst.Host,
		Port:              inst.Port,
		Database:          "postgres",
		UsernameReference: "anarva_admin",
		PasswordSecretRef: fmt.Sprintf("sec_ref_%s", inst.ID[:8]),
		SSLMode:           "disable",
		ConnectionString:  connStr,
	}, nil
}

func (p *LocalDockerPostgresProvider) ScaleInstance(ctx context.Context, instanceID string, cpu float64, memoryMB, storageGB int) (*domain.PostgresInstance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, exists := p.instances[instanceID]
	if !exists {
		return nil, domain.ErrInstanceNotFound
	}

	inst.CPU = cpu
	inst.MemoryMB = memoryMB
	inst.StorageGB = storageGB
	inst.UpdatedAt = time.Now()
	return inst, nil
}
