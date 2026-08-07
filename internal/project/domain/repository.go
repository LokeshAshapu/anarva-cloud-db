package domain

import "context"

type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id string) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	ListByOwnerID(ctx context.Context, ownerID string) ([]*Organization, error)
}

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	GetBySlug(ctx context.Context, slug string) (*Project, error)
	ListByOrgID(ctx context.Context, orgID string) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
}

type MemberRepository interface {
	Create(ctx context.Context, member *OrganizationMember) error
	GetByOrgAndUser(ctx context.Context, orgID, userID string) (*OrganizationMember, error)
	ListByOrgID(ctx context.Context, orgID string) ([]*OrganizationMember, error)
	Delete(ctx context.Context, orgID, userID string) error
}

type InvitationRepository interface {
	Create(ctx context.Context, invitation *Invitation) error
	GetByToken(ctx context.Context, token string) (*Invitation, error)
	Delete(ctx context.Context, token string) error
}
