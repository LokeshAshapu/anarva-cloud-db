package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/project/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
)

type ProjectUseCase interface {
	CreateOrganization(ctx context.Context, ownerID, name, slug string) (*domain.Organization, error)
	GetOrganization(ctx context.Context, id string) (*domain.Organization, error)
	ListOrganizations(ctx context.Context, ownerID string) ([]*domain.Organization, error)

	CreateProject(ctx context.Context, orgID, name, slug, region string) (*domain.Project, error)
	GetProject(ctx context.Context, id string) (*domain.Project, error)
	ListProjects(ctx context.Context, orgID string) ([]*domain.Project, error)
	DeleteProject(ctx context.Context, userID, projectID string) error

	InviteMember(ctx context.Context, orgID, email, role, invitedBy string) (string, error)
	AcceptInvitation(ctx context.Context, token, userID string) error
	ListMembers(ctx context.Context, orgID string) ([]*domain.OrganizationMember, error)
	RemoveMember(ctx context.Context, orgID, requesterID, targetUserID string) error
}

type projectUseCase struct {
	orgRepo        domain.OrganizationRepository
	projectRepo    domain.ProjectRepository
	memberRepo     domain.MemberRepository
	invitationRepo domain.InvitationRepository
}

func NewProjectUseCase(
	orgRepo domain.OrganizationRepository,
	projectRepo domain.ProjectRepository,
	memberRepo domain.MemberRepository,
	invitationRepo domain.InvitationRepository,
) ProjectUseCase {
	return &projectUseCase{
		orgRepo:        orgRepo,
		projectRepo:    projectRepo,
		memberRepo:     memberRepo,
		invitationRepo: invitationRepo,
	}
}

func (u *projectUseCase) CreateOrganization(ctx context.Context, ownerID, name, slug string) (*domain.Organization, error) {
	if ownerID == "" || name == "" || slug == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "ownerID, name, and slug are required")
	}

	org := domain.NewOrganization(ownerID, name, slug)
	if err := u.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	// Add owner as Organization OWNER member
	ownerMember := domain.NewOrganizationMember(org.ID, ownerID, "OWNER")
	if err := u.memberRepo.Create(ctx, ownerMember); err != nil {
		return nil, err
	}

	logger.Context(ctx).Info(fmt.Sprintf("Created organization '%s' (%s)", org.Name, org.ID))
	return org, nil
}

func (u *projectUseCase) GetOrganization(ctx context.Context, id string) (*domain.Organization, error) {
	return u.orgRepo.GetByID(ctx, id)
}

func (u *projectUseCase) ListOrganizations(ctx context.Context, ownerID string) ([]*domain.Organization, error) {
	return u.orgRepo.ListByOwnerID(ctx, ownerID)
}

func (u *projectUseCase) CreateProject(ctx context.Context, orgID, name, slug, region string) (*domain.Project, error) {
	if orgID == "" || name == "" || slug == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "orgID, name, and slug are required")
	}

	_, err := u.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	project := domain.NewProject(orgID, name, slug, region)
	if err := u.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	logger.Context(ctx).Info(fmt.Sprintf("Created project '%s' (%s) in region %s", project.Name, project.ID, project.Region))
	return project, nil
}

func (u *projectUseCase) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	return u.projectRepo.GetByID(ctx, id)
}

func (u *projectUseCase) ListProjects(ctx context.Context, orgID string) ([]*domain.Project, error) {
	return u.projectRepo.ListByOrgID(ctx, orgID)
}

func (u *projectUseCase) DeleteProject(ctx context.Context, userID, projectID string) error {
	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	member, err := u.memberRepo.GetByOrgAndUser(ctx, project.OrgID, userID)
	if err != nil || (member.Role != "OWNER" && member.Role != "ADMIN") {
		return appErrors.New(appErrors.CodeForbidden, "only org OWNER or ADMIN can delete projects")
	}

	return u.projectRepo.Delete(ctx, projectID)
}

func (u *projectUseCase) InviteMember(ctx context.Context, orgID, email, role, invitedBy string) (string, error) {
	if orgID == "" || email == "" {
		return "", appErrors.New(appErrors.CodeInvalidInput, "orgID and email are required")
	}

	invitation := domain.NewInvitation(orgID, email, role, invitedBy, 7*24*time.Hour)
	if err := u.invitationRepo.Create(ctx, invitation); err != nil {
		return "", err
	}

	return invitation.Token, nil
}

func (u *projectUseCase) AcceptInvitation(ctx context.Context, token, userID string) error {
	invitation, err := u.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		return appErrors.New(appErrors.CodeInvalidInput, "invalid or expired invitation token")
	}

	if invitation.IsExpired() {
		_ = u.invitationRepo.Delete(ctx, token)
		return appErrors.New(appErrors.CodeInvalidInput, "invitation token has expired")
	}

	member := domain.NewOrganizationMember(invitation.OrgID, userID, invitation.Role)
	if err := u.memberRepo.Create(ctx, member); err != nil {
		return err
	}

	_ = u.invitationRepo.Delete(ctx, token)
	return nil
}

func (u *projectUseCase) ListMembers(ctx context.Context, orgID string) ([]*domain.OrganizationMember, error) {
	return u.memberRepo.ListByOrgID(ctx, orgID)
}

func (u *projectUseCase) RemoveMember(ctx context.Context, orgID, requesterID, targetUserID string) error {
	member, err := u.memberRepo.GetByOrgAndUser(ctx, orgID, requesterID)
	if err != nil || (member.Role != "OWNER" && member.Role != "ADMIN") {
		return appErrors.New(appErrors.CodeForbidden, "only org OWNER or ADMIN can remove team members")
	}

	return u.memberRepo.Delete(ctx, orgID, targetUserID)
}
