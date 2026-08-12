package developer_test

import (
	"context"
	"strings"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/developer/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/developer/usecase"
)

func TestDeveloperPlatform_APIKeysAndServiceAccounts(t *testing.T) {
	uc := usecase.NewDeveloperUseCase()
	ctx := context.Background()

	// Test Key Creation
	key, secretKey, err := uc.CreateAPIKey(ctx, "Test Key", "org-test", "proj-test", "test@anarva.io", []string{"compute.read"}, true)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	if key.Status != domain.KeyStatusActive {
		t.Errorf("Expected status KeyStatusActive, got %s", key.Status)
	}

	if !strings.HasPrefix(secretKey, "ank_live_") {
		t.Errorf("Expected prefix ank_live_, got secret: %s", secretKey)
	}

	// Test Key Validation
	validated, err := uc.ValidateAPIKey(ctx, secretKey)
	if err != nil {
		t.Fatalf("ValidateAPIKey failed: %v", err)
	}
	if validated.ID != key.ID {
		t.Errorf("Expected key ID %s, got %s", key.ID, validated.ID)
	}

	// Test Revocation
	if err := uc.RevokeAPIKey(ctx, key.ID); err != nil {
		t.Fatalf("RevokeAPIKey failed: %v", err)
	}

	_, err = uc.ValidateAPIKey(ctx, secretKey)
	if err == nil {
		t.Errorf("Expected error validating revoked key, got nil")
	}

	// Test Service Account Creation
	sa, err := uc.CreateServiceAccount(ctx, "CI/CD Bot", "Deployment Account", "ADMIN", "org-test", "proj-test", "test@anarva.io")
	if err != nil {
		t.Fatalf("CreateServiceAccount failed: %v", err)
	}
	if sa.Role != "ADMIN" {
		t.Errorf("Expected ADMIN role, got %s", sa.Role)
	}
}
