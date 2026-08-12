package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/developer/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type DeveloperUseCase struct {
	mu              sync.RWMutex
	keys            map[string]*domain.APIKey
	keysByHash      map[string]*domain.APIKey
	serviceAccounts map[string]*domain.ServiceAccount
	usageRecords    []*domain.APIUsageRecord
}

func NewDeveloperUseCase() *DeveloperUseCase {
	uc := &DeveloperUseCase{
		keys:            make(map[string]*domain.APIKey),
		keysByHash:      make(map[string]*domain.APIKey),
		serviceAccounts: make(map[string]*domain.ServiceAccount),
	}
	uc.seedDefaults()
	return uc
}

func (uc *DeveloperUseCase) seedDefaults() {
	now := time.Now()
	// Seed default API Key
	secret, hash, prefix := domain.GenerateAPIKey(true)
	_ = secret // Only stored as hash

	key := &domain.APIKey{
		ID:             "ank-101",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		Name:           "Primary CLI Key",
		KeyPrefix:      prefix,
		KeyHash:        hash,
		Status:         domain.KeyStatusActive,
		Permissions:    []string{"compute.read", "compute.create", "database.read", "storage.read", "network.read", "provisioning.read"},
		CreatedBy:      "lokeshashapu@gmail.com",
		CreatedAt:      now.Add(-24 * time.Hour),
	}
	uc.keys[key.ID] = key
	uc.keysByHash[key.KeyHash] = key

	// Seed default Service Account
	sa := &domain.ServiceAccount{
		ID:             "sa-101",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		Name:           "GitHub Actions CI/CD Deployer",
		Description:    "Automated infrastructure deployment service account",
		Status:         "ACTIVE",
		Role:           "ADMIN",
		CreatedBy:      "lokeshashapu@gmail.com",
		CreatedAt:      now.Add(-48 * time.Hour),
		UpdatedAt:      now,
	}
	uc.serviceAccounts[sa.ID] = sa
}

func (uc *DeveloperUseCase) CreateAPIKey(ctx context.Context, name, orgID, projectID, createdBy string, permissions []string, isLive bool) (*domain.APIKey, string, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if name == "" {
		return nil, "", appErrors.New(appErrors.CodeInvalidInput, "API key name is required")
	}

	secretKey, keyHash, keyPrefix := domain.GenerateAPIKey(isLive)
	now := time.Now()

	if len(permissions) == 0 {
		permissions = []string{"compute.read", "database.read", "storage.read", "network.read", "provisioning.read"}
	}

	key := &domain.APIKey{
		ID:             fmt.Sprintf("ank-%d", now.UnixNano()/1e6),
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           name,
		KeyPrefix:      keyPrefix,
		KeyHash:        keyHash,
		Status:         domain.KeyStatusActive,
		Permissions:    permissions,
		CreatedBy:      createdBy,
		CreatedAt:      now,
	}

	uc.keys[key.ID] = key
	uc.keysByHash[keyHash] = key

	return key, secretKey, nil
}

func (uc *DeveloperUseCase) ValidateAPIKey(ctx context.Context, secretKey string) (*domain.APIKey, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	hash := domain.HashSecret(secretKey)
	key, ok := uc.keysByHash[hash]
	if !ok || key.Status != domain.KeyStatusActive {
		return nil, appErrors.New(appErrors.CodeUnauthorized, "invalid or revoked API key")
	}

	now := time.Now()
	key.LastUsedAt = &now
	return key, nil
}

func (uc *DeveloperUseCase) ListAPIKeys(ctx context.Context, projectID string) []*domain.APIKey {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	var list []*domain.APIKey
	for _, k := range uc.keys {
		if projectID == "" || k.ProjectID == projectID {
			list = append(list, k)
		}
	}
	return list
}

func (uc *DeveloperUseCase) RevokeAPIKey(ctx context.Context, id string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	key, ok := uc.keys[id]
	if !ok {
		return appErrors.New(appErrors.CodeNotFound, "API key not found")
	}

	now := time.Now()
	key.Status = domain.KeyStatusRevoked
	key.RevokedAt = &now
	return nil
}

func (uc *DeveloperUseCase) CreateServiceAccount(ctx context.Context, name, description, role, orgID, projectID, createdBy string) (*domain.ServiceAccount, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if name == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "Service account name is required")
	}

	now := time.Now()
	sa := &domain.ServiceAccount{
		ID:             fmt.Sprintf("sa-%d", now.UnixNano()/1e6),
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           name,
		Description:    description,
		Status:         "ACTIVE",
		Role:           role,
		CreatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	uc.serviceAccounts[sa.ID] = sa
	return sa, nil
}

func (uc *DeveloperUseCase) ListServiceAccounts(ctx context.Context, projectID string) []*domain.ServiceAccount {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	var list []*domain.ServiceAccount
	for _, sa := range uc.serviceAccounts {
		if projectID == "" || sa.ProjectID == projectID {
			list = append(list, sa)
		}
	}
	return list
}

func (uc *DeveloperUseCase) RecordUsage(rec *domain.APIUsageRecord) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	rec.Timestamp = time.Now()
	uc.usageRecords = append([]*domain.APIUsageRecord{rec}, uc.usageRecords...)
	if len(uc.usageRecords) > 100 {
		uc.usageRecords = uc.usageRecords[:100]
	}
}

func (uc *DeveloperUseCase) ListUsage() []*domain.APIUsageRecord {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.usageRecords
}
