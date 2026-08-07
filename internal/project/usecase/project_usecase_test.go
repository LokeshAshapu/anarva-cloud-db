package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/project/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type mockOrgRepo struct {
	orgs map[string]*domain.Organization
}

func newMockOrgRepo() domain.OrganizationRepository {
	return &mockOrgRepo{orgs: make(map[string]*domain.Organization)}
}

func (m *mockOrgRepo) Create(ctx context.Context, org *domain.Organization) error {
	m.orgs[org.ID] = org
	m.orgs[org.Slug] = org
	return nil
}

func (m *mockOrgRepo) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	if o, ok := m.orgs[id]; ok {
		return o, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "org not found")
}

func (m *mockOrgRepo) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	if o, ok := m.orgs[slug]; ok {
		return o, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "org not found")
}

func (m *mockOrgRepo) ListByOwnerID(ctx context.Context, ownerID string) ([]*domain.Organization, error) {
	var list []*domain.Organization
	for _, o := range m.orgs {
		if o.OwnerID == ownerID {
			list = append(list, o)
		}
	}
	return list, nil
}

type mockProjectRepo struct {
	projects map[string]*domain.Project
}

func newMockProjectRepo() domain.ProjectRepository {
	return &mockProjectRepo{projects: make(map[string]*domain.Project)}
}

func (m *mockProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	m.projects[p.ID] = p
	return nil
}

func (m *mockProjectRepo) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "project not found")
}

func (m *mockProjectRepo) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.Slug == slug {
			return p, nil
		}
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "project not found")
}

func (m *mockProjectRepo) ListByOrgID(ctx context.Context, orgID string) ([]*domain.Project, error) {
	var list []*domain.Project
	for _, p := range m.projects {
		if p.OrgID == orgID {
			list = append(list, p)
		}
	}
	return list, nil
}

func (m *mockProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	m.projects[p.ID] = p
	return nil
}

func (m *mockProjectRepo) Delete(ctx context.Context, id string) error {
	delete(m.projects, id)
	return nil
}

type mockMemberRepo struct {
	members map[string]*domain.OrganizationMember
}

func newMockMemberRepo() domain.MemberRepository {
	return &mockMemberRepo{members: make(map[string]*domain.OrganizationMember)}
}

func (m *mockMemberRepo) Create(ctx context.Context, member *domain.OrganizationMember) error {
	key := member.OrgID + ":" + member.UserID
	m.members[key] = member
	return nil
}

func (m *mockMemberRepo) GetByOrgAndUser(ctx context.Context, orgID, userID string) (*domain.OrganizationMember, error) {
	key := orgID + ":" + userID
	if mem, ok := m.members[key]; ok {
		return mem, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "member not found")
}

func (m *mockMemberRepo) ListByOrgID(ctx context.Context, orgID string) ([]*domain.OrganizationMember, error) {
	var list []*domain.OrganizationMember
	for _, mem := range m.members {
		if mem.OrgID == orgID {
			list = append(list, mem)
		}
	}
	return list, nil
}

func (m *mockMemberRepo) Delete(ctx context.Context, orgID, userID string) error {
	key := orgID + ":" + userID
	delete(m.members, key)
	return nil
}

type mockInvRepo struct {
	invs map[string]*domain.Invitation
}

func newMockInvRepo() domain.InvitationRepository {
	return &mockInvRepo{invs: make(map[string]*domain.Invitation)}
}

func (m *mockInvRepo) Create(ctx context.Context, inv *domain.Invitation) error {
	m.invs[inv.Token] = inv
	return nil
}

func (m *mockInvRepo) GetByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	if i, ok := m.invs[token]; ok {
		return i, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "invitation not found")
}

func (m *mockInvRepo) Delete(ctx context.Context, token string) error {
	delete(m.invs, token)
	return nil
}

func setupProjectUseCase() ProjectUseCase {
	return NewProjectUseCase(
		newMockOrgRepo(),
		newMockProjectRepo(),
		newMockMemberRepo(),
		newMockInvRepo(),
	)
}

func TestOrganizationAndProjectLifecycle(t *testing.T) {
	uc := setupProjectUseCase()
	ctx := context.Background()

	ownerID := "owner-uuid-1"
	org, err := uc.CreateOrganization(ctx, ownerID, "Acme Corp", "acme-corp")
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", org.Name)

	project, err := uc.CreateProject(ctx, org.ID, "Production Database", "acme-prod-db", "us-east-1")
	require.NoError(t, err)
	assert.Equal(t, "Production Database", project.Name)
	assert.Equal(t, 5, project.MaxDatabases)

	projects, err := uc.ListProjects(ctx, org.ID)
	require.NoError(t, err)
	assert.Len(t, projects, 1)

	// Delete Project by Owner
	err = uc.DeleteProject(ctx, ownerID, project.ID)
	require.NoError(t, err)

	projectsAfter, _ := uc.ListProjects(ctx, org.ID)
	assert.Len(t, projectsAfter, 0)
}

func TestInvitationAndMemberLifecycle(t *testing.T) {
	uc := setupProjectUseCase()
	ctx := context.Background()

	ownerID := "owner-uuid-1"
	org, err := uc.CreateOrganization(ctx, ownerID, "Beta Systems", "beta-systems")
	require.NoError(t, err)

	token, err := uc.InviteMember(ctx, org.ID, "engineer@beta.io", "DEVELOPER", ownerID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	newUserID := "user-uuid-99"
	err = uc.AcceptInvitation(ctx, token, newUserID)
	require.NoError(t, err)

	members, err := uc.ListMembers(ctx, org.ID)
	require.NoError(t, err)
	assert.Len(t, members, 2) // Owner + New User
}
