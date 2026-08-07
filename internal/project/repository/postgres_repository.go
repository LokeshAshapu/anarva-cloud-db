package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/project/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type organizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) domain.OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	if err := r.db.WithContext(ctx).Create(org).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return appErrors.New(appErrors.CodeAlreadyExists, "organization slug already exists")
		}
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to create organization")
	}
	return nil
}

func (r *organizationRepository) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	var org domain.Organization
	if err := r.db.WithContext(ctx).First(&org, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "organization not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch organization")
	}
	return &org, nil
}

func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	var org domain.Organization
	if err := r.db.WithContext(ctx).First(&org, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "organization not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch organization by slug")
	}
	return &org, nil
}

func (r *organizationRepository) ListByOwnerID(ctx context.Context, ownerID string) ([]*domain.Organization, error) {
	var orgs []*domain.Organization
	if err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&orgs).Error; err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list organizations")
	}
	return orgs, nil
}

// Project repository
type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) domain.ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *domain.Project) error {
	if err := r.db.WithContext(ctx).Create(project).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return appErrors.New(appErrors.CodeAlreadyExists, "project slug already exists")
		}
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to create project")
	}
	return nil
}

func (r *projectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	var project domain.Project
	if err := r.db.WithContext(ctx).First(&project, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "project not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch project")
	}
	return &project, nil
}

func (r *projectRepository) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	var project domain.Project
	if err := r.db.WithContext(ctx).First(&project, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "project not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch project by slug")
	}
	return &project, nil
}

func (r *projectRepository) ListByOrgID(ctx context.Context, orgID string) ([]*domain.Project, error) {
	var projects []*domain.Project
	if err := r.db.WithContext(ctx).Where("org_id = ?", orgID).Find(&projects).Error; err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list projects")
	}
	return projects, nil
}

func (r *projectRepository) Update(ctx context.Context, project *domain.Project) error {
	if err := r.db.WithContext(ctx).Save(project).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update project")
	}
	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Project{}, "id = ?", id).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete project")
	}
	return nil
}

// Member Repository
type memberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) domain.MemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Create(ctx context.Context, member *domain.OrganizationMember) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to add member")
	}
	return nil
}

func (r *memberRepository) GetByOrgAndUser(ctx context.Context, orgID, userID string) (*domain.OrganizationMember, error) {
	var member domain.OrganizationMember
	if err := r.db.WithContext(ctx).First(&member, "org_id = ? AND user_id = ?", orgID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "member not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch member")
	}
	return &member, nil
}

func (r *memberRepository) ListByOrgID(ctx context.Context, orgID string) ([]*domain.OrganizationMember, error) {
	var members []*domain.OrganizationMember
	if err := r.db.WithContext(ctx).Where("org_id = ?", orgID).Find(&members).Error; err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list members")
	}
	return members, nil
}

func (r *memberRepository) Delete(ctx context.Context, orgID, userID string) error {
	if err := r.db.WithContext(ctx).Where("org_id = ? AND user_id = ?", orgID, userID).Delete(&domain.OrganizationMember{}).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to remove member")
	}
	return nil
}

// Invitation Repository
type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) domain.InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(ctx context.Context, invitation *domain.Invitation) error {
	if err := r.db.WithContext(ctx).Create(invitation).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to create invitation")
	}
	return nil
}

func (r *invitationRepository) GetByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	var invitation domain.Invitation
	if err := r.db.WithContext(ctx).First(&invitation, "token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "invitation not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch invitation")
	}
	return &invitation, nil
}

func (r *invitationRepository) Delete(ctx context.Context, token string) error {
	if err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&domain.Invitation{}).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete invitation")
	}
	return nil
}
