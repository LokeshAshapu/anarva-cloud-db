package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/iam/domain"
)

type AuthorizationService struct {
	mu           sync.RWMutex
	orgs         map[string]*domain.Organization
	projects     map[string]*domain.Project
	members      map[string]*domain.OrgMember
	invitations  map[string]*domain.Invitation
	projectRes   map[string]int // projectID -> resource count
	rolePolicies map[domain.RoleType][]domain.Permission
}

func NewAuthorizationService() *AuthorizationService {
	s := &AuthorizationService{
		orgs:         make(map[string]*domain.Organization),
		projects:     make(map[string]*domain.Project),
		members:      make(map[string]*domain.OrgMember),
		invitations:  make(map[string]*domain.Invitation),
		projectRes:   make(map[string]int),
		rolePolicies: make(map[domain.RoleType][]domain.Permission),
	}
	s.setupRolePolicies()
	s.seedDefaults()
	return s
}

func (s *AuthorizationService) setupRolePolicies() {
	allPerms := []domain.Permission{
		domain.PermOrgRead, domain.PermOrgUpdate, domain.PermOrgDelete,
		domain.PermOrgMembersRead, domain.PermOrgMembersInvite, domain.PermOrgMembersRemove,
		domain.PermProjCreate, domain.PermProjRead, domain.PermProjUpdate, domain.PermProjDelete, domain.PermProjMembersManage,
		domain.PermComputeRead, domain.PermComputeCreate, domain.PermComputeUpdate, domain.PermComputeDelete,
		domain.PermDatabaseRead, domain.PermDatabaseCreate, domain.PermDatabaseUpdate, domain.PermDatabaseDelete,
		domain.PermStorageRead, domain.PermStorageCreate, domain.PermStorageUpload, domain.PermStorageDelete,
		domain.PermMetricsRead, domain.PermBillingRead, domain.PermBillingManage, domain.PermBillingPricingManage,
		domain.PermAuditRead, domain.PermIAMManage,
	}

	s.rolePolicies[domain.RoleOwner] = allPerms

	s.rolePolicies[domain.RoleAdmin] = []domain.Permission{
		domain.PermOrgRead, domain.PermOrgUpdate,
		domain.PermOrgMembersRead, domain.PermOrgMembersInvite, domain.PermOrgMembersRemove,
		domain.PermProjCreate, domain.PermProjRead, domain.PermProjUpdate, domain.PermProjDelete, domain.PermProjMembersManage,
		domain.PermComputeRead, domain.PermComputeCreate, domain.PermComputeUpdate, domain.PermComputeDelete,
		domain.PermDatabaseRead, domain.PermDatabaseCreate, domain.PermDatabaseUpdate, domain.PermDatabaseDelete,
		domain.PermStorageRead, domain.PermStorageCreate, domain.PermStorageUpload, domain.PermStorageDelete,
		domain.PermMetricsRead, domain.PermBillingRead, domain.PermAuditRead, domain.PermIAMManage,
	}

	s.rolePolicies[domain.RoleDeveloper] = []domain.Permission{
		domain.PermOrgRead, domain.PermProjRead,
		domain.PermComputeRead, domain.PermComputeCreate, domain.PermComputeUpdate, domain.PermComputeDelete,
		domain.PermDatabaseRead, domain.PermDatabaseCreate, domain.PermDatabaseUpdate, domain.PermDatabaseDelete,
		domain.PermStorageRead, domain.PermStorageCreate, domain.PermStorageUpload, domain.PermStorageDelete,
		domain.PermMetricsRead,
	}

	s.rolePolicies[domain.RoleViewer] = []domain.Permission{
		domain.PermOrgRead, domain.PermProjRead,
		domain.PermComputeRead, domain.PermDatabaseRead, domain.PermStorageRead, domain.PermMetricsRead,
	}

	s.rolePolicies[domain.RoleBillingAdmin] = []domain.Permission{
		domain.PermOrgRead, domain.PermBillingRead, domain.PermBillingManage, domain.PermBillingPricingManage, domain.PermMetricsRead,
	}

	s.rolePolicies[domain.RoleAuditor] = []domain.Permission{
		domain.PermOrgRead, domain.PermProjRead, domain.PermAuditRead, domain.PermMetricsRead,
		domain.PermComputeRead, domain.PermDatabaseRead, domain.PermStorageRead, domain.PermBillingRead,
	}
}

func (s *AuthorizationService) seedDefaults() {
	now := time.Now()

	org := &domain.Organization{
		ID:          "org-default",
		Name:        "Anarva Cloud Technologies",
		Slug:        "anarva-cloud-technologies",
		OwnerUserID: "usr-lokesh",
		Status:      "ACTIVE",
		CreatedAt:   now.Add(-720 * time.Hour),
		UpdatedAt:   now,
	}
	s.orgs[org.ID] = org

	proj := &domain.Project{
		ID:             "proj-default",
		OrganizationID: "org-default",
		Name:           "Default Production Project",
		Slug:           "default-production-project",
		Description:    "Primary production workload project",
		Status:         "ACTIVE",
		CreatedAt:      now.Add(-720 * time.Hour),
		UpdatedAt:      now,
	}
	s.projects[proj.ID] = proj
	s.projectRes[proj.ID] = 3 // EC2, RDS, S3 active resources

	m1 := &domain.OrgMember{
		ID:             "mem-101",
		OrganizationID: "org-default",
		UserID:         "usr-lokesh",
		UserEmail:      "lokeshashapu@gmail.com",
		UserName:       "Lokesh Ashapu",
		Role:           domain.RoleOwner,
		Status:         "ACTIVE",
		JoinedAt:       now.Add(-720 * time.Hour),
	}
	s.members[m1.ID] = m1
}

// Authorize returns ALLOW (true) or DENY (false) based on user's role policy
func (s *AuthorizationService) Authorize(userEmail, orgID, projectID string, perm domain.Permission) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var member *domain.OrgMember
	for _, m := range s.members {
		if m.UserEmail == userEmail && m.OrganizationID == orgID && m.Status == "ACTIVE" {
			member = m
			break
		}
	}

	if member == nil {
		return false, fmt.Errorf("AUTHORIZATION_DENIED: User %s is not an active member of organization %s", userEmail, orgID)
	}

	allowedPerms, exists := s.rolePolicies[member.Role]
	if !exists {
		return false, fmt.Errorf("AUTHORIZATION_DENIED: Role %s has no permissions policy defined", member.Role)
	}

	for _, p := range allowedPerms {
		if p == perm {
			return true, nil
		}
	}

	return false, fmt.Errorf("AUTHORIZATION_DENIED: Role %s does not have permission %s", member.Role, perm)
}

func (s *AuthorizationService) CanAccessOrganization(userEmail, orgID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, m := range s.members {
		if m.UserEmail == userEmail && m.OrganizationID == orgID && m.Status == "ACTIVE" {
			return true
		}
	}
	return false
}

func (s *AuthorizationService) RemoveMember(orgID, memberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	target, exists := s.members[memberID]
	if !exists || target.OrganizationID != orgID {
		return fmt.Errorf("Member not found")
	}

	// Last Owner Protection
	if target.Role == domain.RoleOwner {
		ownerCount := 0
		for _, m := range s.members {
			if m.OrganizationID == orgID && m.Role == domain.RoleOwner && m.Status == "ACTIVE" {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return fmt.Errorf("LAST_OWNER_PROTECTION: Cannot remove the final owner of the organization")
		}
	}

	target.Status = "REMOVED"
	return nil
}

func (s *AuthorizationService) DeleteProject(orgID, projectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	proj, exists := s.projects[projectID]
	if !exists || proj.OrganizationID != orgID {
		return fmt.Errorf("Project not found")
	}

	// Project Deletion Protection
	if count, ok := s.projectRes[projectID]; ok && count > 0 {
		return fmt.Errorf("PROJECT_NOT_EMPTY: Cannot delete project containing %d active resources", count)
	}

	proj.Status = "ARCHIVED"
	return nil
}

func (s *AuthorizationService) CreateInvitation(orgID, email string, role domain.RoleType) (*domain.Invitation, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	rawToken := fmt.Sprintf("inv_token_%d_%s", now.UnixNano(), email)
	hash := domain.HashInvitationToken(rawToken)

	inv := &domain.Invitation{
		ID:             fmt.Sprintf("inv-%d", now.UnixNano()/1e6),
		OrganizationID: orgID,
		Email:          email,
		Role:           role,
		TokenHash:      hash,
		Status:         "PENDING",
		ExpiresAt:      now.Add(7 * 24 * time.Hour),
		CreatedAt:      now,
	}

	s.invitations[inv.ID] = inv
	return inv, rawToken, nil
}

func (s *AuthorizationService) ListMembers(orgID string) []*domain.OrgMember {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.OrgMember
	for _, m := range s.members {
		if m.OrganizationID == orgID && m.Status != "REMOVED" {
			result = append(result, m)
		}
	}
	return result
}

func (s *AuthorizationService) ListInvitations(orgID string) []*domain.Invitation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.Invitation
	for _, inv := range s.invitations {
		if inv.OrganizationID == orgID {
			result = append(result, inv)
		}
	}
	return result
}

func (s *AuthorizationService) ListAPIKeys(orgID string) []*domain.APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return []*domain.APIKey{}
}

func (s *AuthorizationService) CreateAPIKey(orgID, projectID, name string) (*domain.APIKey, string, error) {
	return nil, "", fmt.Errorf("DEVELOPER_API_KEYS = NOT_IMPLEMENTED (Phase 36 scope)")
}

func (s *AuthorizationService) RevokeAPIKey(id, orgID string) error {
	return fmt.Errorf("DEVELOPER_API_KEYS = NOT_IMPLEMENTED (Phase 36 scope)")
}

func (s *AuthorizationService) ListServiceAccounts(orgID string) []*domain.ServiceAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return []*domain.ServiceAccount{}
}
