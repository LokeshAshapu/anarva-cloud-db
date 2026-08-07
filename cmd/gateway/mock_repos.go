package main

import (
	"context"
	"sync"

	authDomain "github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
	backupDomain "github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	dbDomain "github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	projDomain "github.com/anarva-cloud/anarva-cloud-db/internal/project/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

// Mock Auth Repositories
type memUserRepo struct {
	mu    sync.RWMutex
	users map[string]*authDomain.User
}

func newMemUserRepo() authDomain.UserRepository {
	return &memUserRepo{users: make(map[string]*authDomain.User)}
}

func (m *memUserRepo) Create(ctx context.Context, u *authDomain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.Email] = u
	m.users[u.ID] = u
	return nil
}

func (m *memUserRepo) GetByID(ctx context.Context, id string) (*authDomain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "user not found")
}

func (m *memUserRepo) GetByEmail(ctx context.Context, email string) (*authDomain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "user not found")
}

func (m *memUserRepo) Update(ctx context.Context, u *authDomain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.Email] = u
	m.users[u.ID] = u
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
	return nil, appErrors.New(appErrors.CodeNotFound, "session not found")
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
	return nil, appErrors.New(appErrors.CodeNotFound, "API key not found")
}

func (m *memKeyRepo) GetByHashedKey(ctx context.Context, hashedKey string) (*authDomain.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, k := range m.keys {
		if k.HashedKey == hashedKey {
			return k, nil
		}
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "API key not found")
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
	return nil, appErrors.New(appErrors.CodeNotFound, "token not found")
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
	return &memOrgRepo{orgs: make(map[string]*projDomain.Organization)}
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
	return nil, appErrors.New(appErrors.CodeNotFound, "org not found")
}

func (m *memOrgRepo) GetBySlug(ctx context.Context, slug string) (*projDomain.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if o, ok := m.orgs[slug]; ok {
		return o, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "org not found")
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
	return list, nil
}

type memProjRepo struct {
	mu       sync.RWMutex
	projects map[string]*projDomain.Project
}

func newMemProjRepo() projDomain.ProjectRepository {
	return &memProjRepo{projects: make(map[string]*projDomain.Project)}
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
	return nil, appErrors.New(appErrors.CodeNotFound, "project not found")
}

func (m *memProjRepo) GetBySlug(ctx context.Context, slug string) (*projDomain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.projects {
		if p.Slug == slug {
			return p, nil
		}
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "project not found")
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
	return nil, appErrors.New(appErrors.CodeNotFound, "member not found")
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
	return nil, appErrors.New(appErrors.CodeNotFound, "invitation not found")
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
	return &memInstanceRepo{instances: make(map[string]*dbDomain.DatabaseInstance)}
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
	return nil, appErrors.New(appErrors.CodeNotFound, "instance not found")
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
	return count, nil
}

// Mock Backup Repository
type memBackupRepo struct {
	mu        sync.RWMutex
	snapshots map[string]*backupDomain.BackupSnapshot
}

func newMemBackupRepo() backupDomain.BackupRepository {
	return &memBackupRepo{snapshots: make(map[string]*backupDomain.BackupSnapshot)}
}

func (m *memBackupRepo) Create(ctx context.Context, s *backupDomain.BackupSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[s.ID] = s
	return nil
}

func (m *memBackupRepo) GetByID(ctx context.Context, id string) (*backupDomain.BackupSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.snapshots[id]; ok {
		return s, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "snapshot not found")
}

func (m *memBackupRepo) ListByDatabaseID(ctx context.Context, databaseID string) ([]*backupDomain.BackupSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*backupDomain.BackupSnapshot
	for _, s := range m.snapshots {
		if s.DatabaseID == databaseID {
			list = append(list, s)
		}
	}
	return list, nil
}

func (m *memBackupRepo) ListByProjectID(ctx context.Context, projectID string) ([]*backupDomain.BackupSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*backupDomain.BackupSnapshot
	for _, s := range m.snapshots {
		if s.ProjectID == projectID {
			list = append(list, s)
		}
	}
	return list, nil
}

func (m *memBackupRepo) Update(ctx context.Context, s *backupDomain.BackupSnapshot) error {
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
