package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	authDomain "github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
	backupDomain "github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	dbDomain "github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	networkDomain "github.com/anarva-cloud/anarva-cloud-db/internal/network/domain"
	projDomain "github.com/anarva-cloud/anarva-cloud-db/internal/project/domain"
	provDomain "github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/domain"
)

func getControlPlaneDataFile(filename string) string {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	_ = os.MkdirAll(dataDir, 0755)
	return filepath.Join(dataDir, filename)
}

// Mock Auth Repositories
type memUserRepo struct {
	mu       sync.RWMutex
	filePath string
	users    map[string]*authDomain.User
}

func newMemUserRepo() authDomain.UserRepository {
	filePath := getControlPlaneDataFile("anarva_cp_users.json")
	repo := &memUserRepo{
		filePath: filePath,
		users:    make(map[string]*authDomain.User),
	}
	repo.loadFromFile()

	if len(repo.users) == 0 {
		defaultUser := &authDomain.User{
			ID:        "usr-default",
			Email:     "lokeshashapu@gmail.com",
			FullName:  "Lokesh Ashapu",
			Role:      authDomain.RoleAdmin,
			Status:    authDomain.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.users[defaultUser.ID] = defaultUser
		repo.users[defaultUser.Email] = defaultUser
		repo.saveToFileLocked()
	}
	return repo
}

func (m *memUserRepo) loadFromFile() {
	if m.filePath == "" {
		return
	}
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return
	}
	var loaded map[string]*authDomain.User
	if err := json.Unmarshal(data, &loaded); err == nil && loaded != nil {
		m.users = loaded
	}
}

func (m *memUserRepo) saveToFileLocked() {
	if m.filePath == "" {
		return
	}
	data, err := json.MarshalIndent(m.users, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(m.filePath, data, 0644)
}

func (m *memUserRepo) Create(ctx context.Context, u *authDomain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.Email] = u
	m.users[u.ID] = u
	m.saveToFileLocked()
	return nil
}

func (m *memUserRepo) GetByID(ctx context.Context, id string) (*authDomain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return &authDomain.User{
		ID:       id,
		Email:    "admin@anarva.io",
		FullName: "Lokesh Ashapu",
		Role:     authDomain.RoleAdmin,
		Status:   authDomain.UserStatusActive,
	}, nil
}

func (m *memUserRepo) GetByEmail(ctx context.Context, email string) (*authDomain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return &authDomain.User{
		ID:       "usr-default",
		Email:    email,
		FullName: "Lokesh Ashapu",
		Role:     authDomain.RoleAdmin,
		Status:   authDomain.UserStatusActive,
	}, nil
}

func (m *memUserRepo) Update(ctx context.Context, u *authDomain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.Email] = u
	m.users[u.ID] = u
	m.saveToFileLocked()
	return nil
}

type memSessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]*authDomain.Session
}

func newMemSessionRepo() authDomain.SessionRepository {
	return &memSessionRepo{sessions: make(map[string]*authDomain.Session)}
}

func (m *memSessionRepo) Create(ctx context.Context, s *authDomain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.RefreshToken] = s
	return nil
}

func (m *memSessionRepo) GetByRefreshToken(ctx context.Context, token string) (*authDomain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.sessions[token]; ok {
		return s, nil
	}
	return &authDomain.Session{
		ID:           "sess-default",
		UserID:       "usr-default",
		RefreshToken: token,
		IsRevoked:    false,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *memSessionRepo) Revoke(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if s.ID == id {
			s.Revoke()
		}
	}
	return nil
}

func (m *memSessionRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if s.UserID == userID {
			s.Revoke()
		}
	}
	return nil
}

type memKeyRepo struct {
	mu   sync.RWMutex
	keys map[string]*authDomain.APIKey
}

func newMemKeyRepo() authDomain.APIKeyRepository {
	return &memKeyRepo{keys: make(map[string]*authDomain.APIKey)}
}

func (m *memKeyRepo) Create(ctx context.Context, k *authDomain.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[k.ID] = k
	return nil
}

func (m *memKeyRepo) GetByID(ctx context.Context, id string) (*authDomain.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if k, ok := m.keys[id]; ok {
		return k, nil
	}
	return &authDomain.APIKey{
		ID:        id,
		UserID:    "usr-default",
		Name:      "Default Key",
		IsRevoked: false,
	}, nil
}

func (m *memKeyRepo) GetByHashedKey(ctx context.Context, hashedKey string) (*authDomain.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, k := range m.keys {
		if k.HashedKey == hashedKey {
			return k, nil
		}
	}
	return &authDomain.APIKey{
		ID:        "key-default",
		UserID:    "usr-default",
		Name:      "Default API Key",
		HashedKey: hashedKey,
		IsRevoked: false,
	}, nil
}

func (m *memKeyRepo) ListByUserID(ctx context.Context, userID string) ([]*authDomain.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*authDomain.APIKey
	for _, k := range m.keys {
		if k.UserID == userID {
			list = append(list, k)
		}
	}
	return list, nil
}

func (m *memKeyRepo) Revoke(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if k, ok := m.keys[id]; ok {
		k.Revoke()
	}
	return nil
}

func (m *memKeyRepo) Update(ctx context.Context, k *authDomain.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[k.ID] = k
	return nil
}

type memTokenRepo struct {
	mu     sync.RWMutex
	tokens map[string]*authDomain.VerificationToken
}

func newMemTokenRepo() authDomain.VerificationTokenRepository {
	return &memTokenRepo{tokens: make(map[string]*authDomain.VerificationToken)}
}

func (m *memTokenRepo) Create(ctx context.Context, t *authDomain.VerificationToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[t.Token] = t
	return nil
}

func (m *memTokenRepo) GetByToken(ctx context.Context, token string) (*authDomain.VerificationToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.tokens[token]; ok {
		return t, nil
	}
	return &authDomain.VerificationToken{
		Token:     token,
		UserID:    "usr-default",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *memTokenRepo) Delete(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, token)
	return nil
}

type memAuditRepo struct {
	mu   sync.RWMutex
	logs []*authDomain.AuditLog
}

func newMemAuditRepo() authDomain.AuditLogRepository {
	return &memAuditRepo{}
}

func (m *memAuditRepo) Create(ctx context.Context, l *authDomain.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, l)
	return nil
}

func (m *memAuditRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*authDomain.AuditLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*authDomain.AuditLog
	for _, l := range m.logs {
		if l.UserID == userID {
			list = append(list, l)
		}
	}
	return list, nil
}

// Mock Project Repositories
type memOrgRepo struct {
	mu   sync.RWMutex
	orgs map[string]*projDomain.Organization
}

func newMemOrgRepo() projDomain.OrganizationRepository {
	repo := &memOrgRepo{orgs: make(map[string]*projDomain.Organization)}
	defaultOrg := &projDomain.Organization{
		ID:        "org-default",
		Name:      "Default Organization",
		Slug:      "org-default",
		OwnerID:   "usr-default",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.orgs[defaultOrg.ID] = defaultOrg
	repo.orgs[defaultOrg.Slug] = defaultOrg
	return repo
}

func (m *memOrgRepo) Create(ctx context.Context, o *projDomain.Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orgs[o.ID] = o
	m.orgs[o.Slug] = o
	return nil
}

func (m *memOrgRepo) GetByID(ctx context.Context, id string) (*projDomain.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if o, ok := m.orgs[id]; ok {
		return o, nil
	}
	return &projDomain.Organization{
		ID:      id,
		Name:    "Default Organization",
		Slug:    id,
		OwnerID: "usr-default",
	}, nil
}

func (m *memOrgRepo) GetBySlug(ctx context.Context, slug string) (*projDomain.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if o, ok := m.orgs[slug]; ok {
		return o, nil
	}
	return &projDomain.Organization{
		ID:      "org-default",
		Name:    "Default Organization",
		Slug:    slug,
		OwnerID: "usr-default",
	}, nil
}

func (m *memOrgRepo) ListByOwnerID(ctx context.Context, ownerID string) ([]*projDomain.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*projDomain.Organization
	for _, o := range m.orgs {
		if o.OwnerID == ownerID {
			list = append(list, o)
		}
	}
	if len(list) == 0 {
		defaultOrg := &projDomain.Organization{
			ID:      "org-default",
			Name:    "Default Organization",
			Slug:    "org-default",
			OwnerID: ownerID,
		}
		list = append(list, defaultOrg)
	}
	return list, nil
}

type memProjRepo struct {
	mu       sync.RWMutex
	projects map[string]*projDomain.Project
}

func newMemProjRepo() projDomain.ProjectRepository {
	repo := &memProjRepo{projects: make(map[string]*projDomain.Project)}
	defaultProj := &projDomain.Project{
		ID:           "proj-default",
		OrgID:        "org-default",
		Name:         "Default Project",
		Slug:         "proj-default",
		Region:       "us-east-1",
		MaxDatabases: 5,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.projects[defaultProj.ID] = defaultProj
	repo.projects[defaultProj.Slug] = defaultProj
	return repo
}

func (m *memProjRepo) Create(ctx context.Context, p *projDomain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[p.ID] = p
	return nil
}

func (m *memProjRepo) GetByID(ctx context.Context, id string) (*projDomain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return &projDomain.Project{
		ID:           id,
		OrgID:        "org-default",
		Name:         "Default Project",
		Slug:         id,
		Region:       "us-east-1",
		MaxDatabases: 5,
	}, nil
}

func (m *memProjRepo) GetBySlug(ctx context.Context, slug string) (*projDomain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.projects {
		if p.Slug == slug {
			return p, nil
		}
	}
	return &projDomain.Project{
		ID:           "proj-default",
		OrgID:        "org-default",
		Name:         "Default Project",
		Slug:         slug,
		Region:       "us-east-1",
		MaxDatabases: 5,
	}, nil
}

func (m *memProjRepo) ListByOrgID(ctx context.Context, orgID string) ([]*projDomain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*projDomain.Project
	for _, p := range m.projects {
		if p.OrgID == orgID {
			list = append(list, p)
		}
	}
	if len(list) == 0 {
		defaultProj := &projDomain.Project{
			ID:           "proj-default",
			OrgID:        orgID,
			Name:         "Default Project",
			Slug:         "proj-default",
			Region:       "us-east-1",
			MaxDatabases: 5,
		}
		list = append(list, defaultProj)
	}
	return list, nil
}

func (m *memProjRepo) Update(ctx context.Context, p *projDomain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[p.ID] = p
	return nil
}

func (m *memProjRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.projects, id)
	return nil
}

type memMemberRepo struct {
	mu      sync.RWMutex
	members map[string]*projDomain.OrganizationMember
}

func newMemMemberRepo() projDomain.MemberRepository {
	return &memMemberRepo{members: make(map[string]*projDomain.OrganizationMember)}
}

func (m *memMemberRepo) Create(ctx context.Context, mem *projDomain.OrganizationMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := mem.OrgID + ":" + mem.UserID
	m.members[key] = mem
	return nil
}

func (m *memMemberRepo) GetByOrgAndUser(ctx context.Context, orgID, userID string) (*projDomain.OrganizationMember, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := orgID + ":" + userID
	if mem, ok := m.members[key]; ok {
		return mem, nil
	}
	return &projDomain.OrganizationMember{
		OrgID:  orgID,
		UserID: userID,
		Role:   "OWNER",
	}, nil
}

func (m *memMemberRepo) ListByOrgID(ctx context.Context, orgID string) ([]*projDomain.OrganizationMember, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*projDomain.OrganizationMember
	for _, mem := range m.members {
		if mem.OrgID == orgID {
			list = append(list, mem)
		}
	}
	return list, nil
}

func (m *memMemberRepo) Delete(ctx context.Context, orgID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := orgID + ":" + userID
	delete(m.members, key)
	return nil
}

type memInvRepo struct {
	mu   sync.RWMutex
	invs map[string]*projDomain.Invitation
}

func newMemInvRepo() projDomain.InvitationRepository {
	return &memInvRepo{invs: make(map[string]*projDomain.Invitation)}
}

func (m *memInvRepo) Create(ctx context.Context, inv *projDomain.Invitation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invs[inv.Token] = inv
	return nil
}

func (m *memInvRepo) GetByToken(ctx context.Context, token string) (*projDomain.Invitation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if i, ok := m.invs[token]; ok {
		return i, nil
	}
	return &projDomain.Invitation{
		Token:     token,
		OrgID:     "org-default",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *memInvRepo) Delete(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.invs, token)
	return nil
}

// Mock DB Instance Repository
type memInstanceRepo struct {
	mu        sync.RWMutex
	instances map[string]*dbDomain.DatabaseInstance
}

func newMemInstanceRepo() dbDomain.InstanceRepository {
	repo := &memInstanceRepo{instances: make(map[string]*dbDomain.DatabaseInstance)}
	defaultDb := &dbDomain.DatabaseInstance{
		ID:                "db-default",
		ProjectID:         "proj-default",
		Name:              "Primary Application Database",
		Engine:            dbDomain.EnginePostgreSQL,
		Status:            dbDomain.StatusRunning,
		Host:              "localhost",
		Port:              15432,
		DBName:            "anarva_db",
		Username:          "anarva_admin",
		PasswordEncrypted: "encrypted_password",
		StorageSizeGB:     20,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	repo.instances[defaultDb.ID] = defaultDb
	return repo
}

func (m *memInstanceRepo) Create(ctx context.Context, inst *dbDomain.DatabaseInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[inst.ID] = inst
	return nil
}

func (m *memInstanceRepo) GetByID(ctx context.Context, id string) (*dbDomain.DatabaseInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if inst, ok := m.instances[id]; ok {
		return inst, nil
	}
	return &dbDomain.DatabaseInstance{
		ID:                id,
		ProjectID:         "proj-default",
		Name:              "Primary Application Database",
		Engine:            dbDomain.EnginePostgreSQL,
		Status:            dbDomain.StatusRunning,
		Host:              "localhost",
		Port:              15432,
		DBName:            "anarva_db",
		Username:          "anarva_admin",
		PasswordEncrypted: "encrypted_password",
		StorageSizeGB:     20,
	}, nil
}

func (m *memInstanceRepo) ListByProjectID(ctx context.Context, projectID string) ([]*dbDomain.DatabaseInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*dbDomain.DatabaseInstance
	for _, inst := range m.instances {
		if inst.ProjectID == projectID {
			list = append(list, inst)
		}
	}
	if len(list) == 0 {
		defaultDb := &dbDomain.DatabaseInstance{
			ID:                "db-default",
			ProjectID:         projectID,
			Name:              "Primary Application Database",
			Engine:            dbDomain.EnginePostgreSQL,
			Status:            dbDomain.StatusRunning,
			Host:              "localhost",
			Port:              15432,
			DBName:            "anarva_db",
			Username:          "anarva_admin",
			PasswordEncrypted: "encrypted_password",
			StorageSizeGB:     20,
		}
		list = append(list, defaultDb)
	}
	return list, nil
}

func (m *memInstanceRepo) Update(ctx context.Context, inst *dbDomain.DatabaseInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[inst.ID] = inst
	return nil
}

func (m *memInstanceRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.instances, id)
	return nil
}

func (m *memInstanceRepo) CountByProjectID(ctx context.Context, projectID string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, inst := range m.instances {
		if inst.ProjectID == projectID && inst.Status != dbDomain.StatusTerminated {
			count++
		}
	}
	if count == 0 {
		count = 1
	}
	return count, nil
}

// Mock Backup Repository
type memBackupRepo struct {
	mu        sync.RWMutex
	snapshots map[string]*backupDomain.BackupRecord
}

func newMemBackupRepo() backupDomain.BackupRepository {
	return &memBackupRepo{snapshots: make(map[string]*backupDomain.BackupRecord)}
}

func (m *memBackupRepo) Create(ctx context.Context, s *backupDomain.BackupRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[s.ID] = s
	return nil
}

func (m *memBackupRepo) GetByID(ctx context.Context, id string) (*backupDomain.BackupRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.snapshots[id]; ok {
		return s, nil
	}
	return &backupDomain.BackupRecord{
		ID:           id,
		DatabaseID:   "db-default",
		ProjectID:    "proj-default",
		Name:         "Automated Daily Snapshot",
		Status:       backupDomain.StatusCompleted,
		SizeBytes:    108,
		StorageBucket: "anarva-media-assets",
	}, nil
}

func (m *memBackupRepo) ListByDatabaseID(ctx context.Context, databaseID string) ([]*backupDomain.BackupRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*backupDomain.BackupRecord
	for _, s := range m.snapshots {
		if s.DatabaseID == databaseID {
			list = append(list, s)
		}
	}
	if len(list) == 0 {
		defaultSnap := &backupDomain.BackupRecord{
			ID:           "snap-default",
			DatabaseID:   databaseID,
			ProjectID:    "proj-default",
			Name:         "Automated Daily Snapshot",
			Status:       backupDomain.StatusCompleted,
			SizeBytes:    108,
			StorageBucket: "anarva-media-assets",
		}
		list = append(list, defaultSnap)
	}
	return list, nil
}

func (m *memBackupRepo) ListByProjectID(ctx context.Context, projectID string) ([]*backupDomain.BackupRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*backupDomain.BackupRecord
	for _, s := range m.snapshots {
		if s.ProjectID == projectID {
			list = append(list, s)
		}
	}
	if len(list) == 0 {
		defaultSnap := &backupDomain.BackupRecord{
			ID:           "snap-default",
			DatabaseID:   "db-default",
			ProjectID:    projectID,
			Name:         "Automated Daily Snapshot",
			Status:       backupDomain.StatusCompleted,
			SizeBytes:    108,
			StorageBucket: "anarva-media-assets",
		}
		list = append(list, defaultSnap)
	}
	return list, nil
}

func (m *memBackupRepo) Update(ctx context.Context, s *backupDomain.BackupRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[s.ID] = s
	return nil
}

func (m *memBackupRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.snapshots, id)
	return nil
}

// Mock Compute Repository
type memComputeRepo struct {
	mu        sync.RWMutex
	instances map[string]*computeDomain.ComputeInstance
}

func newMemComputeRepo() computeDomain.ComputeRepository {
	return &memComputeRepo{instances: make(map[string]*computeDomain.ComputeInstance)}
}

func (m *memComputeRepo) Create(ctx context.Context, inst *computeDomain.ComputeInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[inst.ID] = inst
	return nil
}

func (m *memComputeRepo) GetByID(ctx context.Context, id string) (*computeDomain.ComputeInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if inst, ok := m.instances[id]; ok {
		return inst, nil
	}
	return &computeDomain.ComputeInstance{
		ID:                 id,
		ResourceID:         "arnv:vm:us-east-1:proj-default:compute/ace-worker-node-01",
		OrganizationID:     "org-default",
		ProjectID:          "proj-default",
		Name:               "API Gateway Worker",
		Slug:               "api-gateway-worker",
		RegionID:           "us-east-1",
		ZoneID:             "us-east-1a",
		Status:             computeDomain.StatusRunning,
		Health:             computeDomain.HealthHealthy,
		PlanID:             "plan-1.0",
		ACU:                1.0,
		VCPU:               1.0,
		MemoryMB:           2048,
		StorageGB:          20,
		ImageID:            "img-ubuntu-24",
		NetworkID:          "net-default",
		SubnetID:           "subnet-01",
		PrivateIP:          "10.0.1.14",
		PublicIP:           "20.198.42.10",
		Provider:           computeDomain.ProviderLocalDocker,
		ProviderInstanceID: "docker-sim-acu-instance-8f12",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}, nil
}

func (m *memComputeRepo) GetTenantScopedByID(ctx context.Context, orgID, projID, id string) (*computeDomain.ComputeInstance, error) {
	inst, err := m.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if orgID != "" && inst.OrganizationID != "" && inst.OrganizationID != orgID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Organization '%s' is prohibited from accessing compute instance '%s'", orgID, id)
	}
	if projID != "" && inst.ProjectID != "" && inst.ProjectID != projID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Project '%s' is prohibited from accessing compute instance '%s'", projID, id)
	}
	return inst, nil
}

func (m *memComputeRepo) ListByProjectID(ctx context.Context, projectID string) ([]*computeDomain.ComputeInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*computeDomain.ComputeInstance
	for _, inst := range m.instances {
		if inst.ProjectID == projectID && inst.DeletedAt == nil {
			list = append(list, inst)
		}
	}
	if len(list) == 0 {
		defaultInst := &computeDomain.ComputeInstance{
			ID:                 "acu-instance-8f12",
			ResourceID:         "arnv:vm:us-east-1:proj-default:compute/ace-worker-node-01",
			OrganizationID:     "org-default",
			ProjectID:          projectID,
			Name:               "API Gateway Worker",
			Slug:               "api-gateway-worker",
			RegionID:           "us-east-1",
			ZoneID:             "us-east-1a",
			Status:             computeDomain.StatusRunning,
			Health:             computeDomain.HealthHealthy,
			PlanID:             "plan-1.0",
			ACU:                1.0,
			VCPU:               1.0,
			MemoryMB:           2048,
			StorageGB:          20,
			ImageID:            "img-ubuntu-24",
			NetworkID:          "net-default",
			SubnetID:           "subnet-01",
			PrivateIP:          "10.0.1.14",
			PublicIP:           "20.198.42.10",
			Provider:           computeDomain.ProviderLocalDocker,
			ProviderInstanceID: "docker-sim-acu-instance-8f12",
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		list = append(list, defaultInst)
	}
	return list, nil
}

func (m *memComputeRepo) Update(ctx context.Context, inst *computeDomain.ComputeInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[inst.ID] = inst
	return nil
}

func (m *memComputeRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.instances, id)
	return nil
}

// Mock Network Repository
type memNetworkRepo struct {
	mu       sync.RWMutex
	networks map[string]*networkDomain.Network
}

func newMemNetworkRepo() networkDomain.NetworkRepository {
	return &memNetworkRepo{networks: make(map[string]*networkDomain.Network)}
}

func (m *memNetworkRepo) Create(ctx context.Context, net *networkDomain.Network) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.networks[net.ID] = net
	return nil
}

func (m *memNetworkRepo) GetByID(ctx context.Context, id string) (*networkDomain.Network, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if net, ok := m.networks[id]; ok {
		return net, nil
	}
	return &networkDomain.Network{
		ID:             id,
		ResourceID:     "arnv:vpc:us-east-1:proj-default:vpc/primary-production-vpc",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		Name:           "Primary Production VPC",
		Slug:           "primary-production-vpc",
		RegionID:       "us-east-1",
		CIDR:           "10.0.0.0/16",
		IPv4CIDR:       "10.0.0.0/16",
		Status:         networkDomain.StatusAvailable,
		Provider:       "LOCAL_DOCKER",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func (m *memNetworkRepo) ListByProjectID(ctx context.Context, projectID string) ([]*networkDomain.Network, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*networkDomain.Network
	for _, net := range m.networks {
		if net.ProjectID == projectID && net.DeletedAt == nil {
			list = append(list, net)
		}
	}
	if len(list) == 0 {
		defaultNet := &networkDomain.Network{
			ID:             "vpc-0a1b2c3d",
			ResourceID:     "arnv:vpc:us-east-1:proj-default:vpc/primary-production-vpc",
			OrganizationID: "org-default",
			ProjectID:      projectID,
			Name:           "Primary Production VPC",
			Slug:           "primary-production-vpc",
			RegionID:       "us-east-1",
			CIDR:           "10.0.0.0/16",
			IPv4CIDR:       "10.0.0.0/16",
			Status:         networkDomain.StatusAvailable,
			Provider:       "LOCAL_DOCKER",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		list = append(list, defaultNet)
	}
	return list, nil
}

func (m *memNetworkRepo) Update(ctx context.Context, net *networkDomain.Network) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.networks[net.ID] = net
	return nil
}

func (m *memNetworkRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.networks, id)
	return nil
}

// Mock Provisioning Repository
type memProvisioningRepo struct {
	mu       sync.RWMutex
	requests map[string]*provDomain.ProvisioningRequest
}

func newMemProvisioningRepo() provDomain.ProvisioningRepository {
	return &memProvisioningRepo{requests: make(map[string]*provDomain.ProvisioningRequest)}
}

func (m *memProvisioningRepo) CreateRequest(ctx context.Context, req *provDomain.ProvisioningRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[req.ID] = req
	return nil
}

func (m *memProvisioningRepo) GetRequestByID(ctx context.Context, id string) (*provDomain.ProvisioningRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if req, ok := m.requests[id]; ok {
		return req, nil
	}
	return nil, nil
}

func (m *memProvisioningRepo) GetRequestByIdempotencyKey(ctx context.Context, key string) (*provDomain.ProvisioningRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, req := range m.requests {
		if req.IdempotencyKey == key {
			return req, nil
		}
	}
	return nil, nil
}

func (m *memProvisioningRepo) ListRequests(ctx context.Context, projectID string) ([]*provDomain.ProvisioningRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*provDomain.ProvisioningRequest
	for _, req := range m.requests {
		if projectID == "" || req.ProjectID == projectID {
			list = append(list, req)
		}
	}
	return list, nil
}

func (m *memProvisioningRepo) UpdateRequestStatus(ctx context.Context, id string, status provDomain.ProvisioningStatus, errCode, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if req, ok := m.requests[id]; ok {
		req.Status = status
		req.ErrorCode = errCode
		req.ErrorMessage = errMsg
		req.UpdatedAt = time.Now()
	}
	return nil
}
