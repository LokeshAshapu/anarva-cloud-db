package service

import (
	"fmt"
	"sync"

	"github.com/anarva-cloud/anarva-cloud-db/internal/iam/domain"
)

type AuthorizationService struct {
	mu       sync.RWMutex
	members  map[string]*domain.OrgMember
	apiKeys  map[string]*domain.APIKey
	svcAccts map[string]*domain.ServiceAccount
}

func NewAuthorizationService() *AuthorizationService {
	s := &AuthorizationService{
		members:  make(map[string]*domain.OrgMember),
		apiKeys:  make(map[string]*domain.APIKey),
		svcAccts: make(map[string]*domain.ServiceAccount),
	}
	s.seedDefaults()
	return s
}

func (s *AuthorizationService) seedDefaults() {
	m1 := &domain.OrgMember{
		ID:             "mem-101",
		OrganizationID: "org-default",
		UserID:         "usr-lokesh",
		UserEmail:      "lokeshashapu@gmail.com",
		UserName:       "Lokesh Ashapu",
		Role:           domain.RoleOwner,
		Status:         "ACTIVE",
	}
	s.members[m1.ID] = m1

	_, _, hash := domain.GenerateRawAPIKey("org-default", "Primary CLI Key")
	k1 := &domain.APIKey{
		ID:             "ak-101",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		Name:           "Primary CLI Key",
		KeyPrefix:      "anarva_l",
		HashedSecret:   hash,
		Permissions:    []string{"*"},
	}
	s.apiKeys[k1.ID] = k1

	sa1 := &domain.ServiceAccount{
		ID:             "sa-101",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		Name:           "GitHub Actions CI/CD Deployer",
		Description:    "Automated deployment service account",
		Status:         "ACTIVE",
		Role:           domain.RoleAdmin,
	}
	s.svcAccts[sa1.ID] = sa1
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

func (s *AuthorizationService) HasPermission(userEmail, orgID, perm string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, m := range s.members {
		if m.UserEmail == userEmail && m.OrganizationID == orgID && m.Status == "ACTIVE" {
			// Owner has all permissions
			if m.Role == domain.RoleOwner || m.Role == domain.RoleAdmin {
				return true
			}
		}
	}
	return false
}

func (s *AuthorizationService) ListAPIKeys(orgID string) []*domain.APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.APIKey
	for _, k := range s.apiKeys {
		if k.OrganizationID == orgID && k.RevokedAt == nil {
			result = append(result, k)
		}
	}
	return result
}

func (s *AuthorizationService) CreateAPIKey(orgID, projectID, name string) (*domain.APIKey, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix, secret, hash := domain.GenerateRawAPIKey(orgID, name)
	key := &domain.APIKey{
		ID:             fmt.Sprintf("ak-%d", len(s.apiKeys)+101),
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           name,
		KeyPrefix:      prefix,
		HashedSecret:   hash,
		Permissions:    []string{"database:read", "storage:read", "compute:read"},
	}

	s.apiKeys[key.ID] = key
	return key, secret, nil
}

func (s *AuthorizationService) RevokeAPIKey(id, orgID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	k, ok := s.apiKeys[id]
	if !ok || k.OrganizationID != orgID {
		return fmt.Errorf("api key not found")
	}
	now := k.CreatedAt
	k.RevokedAt = &now
	return nil
}

func (s *AuthorizationService) ListServiceAccounts(orgID string) []*domain.ServiceAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.ServiceAccount
	for _, sa := range s.svcAccts {
		if sa.OrganizationID == orgID {
			result = append(result, sa)
		}
	}
	return result
}

func (s *AuthorizationService) ListMembers(orgID string) []*domain.OrgMember {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.OrgMember
	for _, m := range s.members {
		if m.OrganizationID == orgID {
			result = append(result, m)
		}
	}
	return result
}
