package providers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/aws"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/drift"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/gcp"
	importengine "github.com/anarva-cloud/anarva-cloud-db/internal/providers/import"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/security"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/service"
)

func TestProviders_RegistryAndVerification(t *testing.T) {
	ctx := context.Background()
	reg := registry.NewProviderRegistry()
	mappingRepo := mapping.NewMappingRepository()
	driftEng := drift.NewDriftEngine(mappingRepo)
	importEng := importengine.NewImportEngine(mappingRepo)
	ssrfEng := security.NewSSRFProtectionEngine()

	svc := service.NewProviderService(reg, mappingRepo, driftEng, importEng, ssrfEng, nil)

	// 1. List Providers
	providers, err := svc.ListProviders(ctx)
	if err != nil || len(providers) < 3 {
		t.Fatalf("expected at least 3 providers registered, got: %v", len(providers))
	}

	// 2. Verify AWS Provider Credentials
	awsP, err := svc.VerifyProvider(ctx, "provider-aws", "arn:aws:iam::123456789012:role/AnarvaExecutionRole")
	if err != nil || awsP.Status != registry.StatusConnected {
		t.Errorf("expected AWS status CONNECTED after credential verification, got: %v", awsP)
	}
}

func TestProviders_ErrorMapping(t *testing.T) {
	// AWS AccessDenied -> PROVIDER_PERMISSION_DENIED
	awsErr := aws.MapAWSError(fmt.Errorf("AccessDenied: User is not authorized to perform ec2:RunInstances"))
	if awsErr == nil || !testingContains(awsErr.Error(), "PROVIDER_PERMISSION_DENIED") {
		t.Errorf("expected PROVIDER_PERMISSION_DENIED mapping for AWS AccessDenied, got: %v", awsErr)
	}

	// GCP PERMISSION_DENIED -> PROVIDER_PERMISSION_DENIED
	gcpErr := gcp.MapGCPError(fmt.Errorf("PERMISSION_DENIED: Compute Engine API has not been used in project"))
	if gcpErr == nil || !testingContains(gcpErr.Error(), "PROVIDER_PERMISSION_DENIED") {
		t.Errorf("expected PROVIDER_PERMISSION_DENIED mapping for GCP PERMISSION_DENIED, got: %v", gcpErr)
	}
}

func TestProviders_ResourceImportAdoptionAndRelease(t *testing.T) {
	ctx := context.Background()
	mappingRepo := mapping.NewMappingRepository()
	importEng := importengine.NewImportEngine(mappingRepo)

	// 1. Import Cloud Resource (Defaults to MANAGED = false)
	m, err := importEng.ImportResource(ctx, "AWS", "i-0a1b2c3d4e5f67890", "ec2-instance", "us-east-1")
	if err != nil || m.Managed != false {
		t.Fatalf("expected imported resource to default to MANAGED = false, got: %v", m)
	}

	// 2. Adopt Resource (Sets MANAGED = true)
	adopted, err := importEng.AdoptResource(ctx, m.AnarvaResourceID)
	if err != nil || adopted.Managed != true {
		t.Errorf("expected adopted resource to have MANAGED = true, got: %v", adopted)
	}

	// 3. Release Resource (Removes Anarva management; sets MANAGED = false)
	if err := importEng.ReleaseResource(ctx, m.AnarvaResourceID); err != nil {
		t.Errorf("failed to release resource management: %v", err)
	}

	released, _ := mappingRepo.GetMapping(m.AnarvaResourceID)
	if released.Managed != false || released.Status != "RELEASED" {
		t.Errorf("expected released resource to have MANAGED = false and status RELEASED, got: %v", released)
	}
}

func TestProviders_MetadataSSRFProtection(t *testing.T) {
	ssrfEng := security.NewSSRFProtectionEngine()

	// 1. AWS Metadata Endpoint (Blocked)
	err := ssrfEng.ValidateURL("http://169.254.169.254/latest/meta-data/")
	if err == nil || !testingContains(err.Error(), "SSRF SECURITY RISK") {
		t.Errorf("expected SSRF SECURITY RISK error for 169.254.169.254, got: %v", err)
	}

	// 2. GCP Metadata Endpoint (Blocked)
	err = ssrfEng.ValidateURL("http://metadata.google.internal/computeMetadata/v1/")
	if err == nil || !testingContains(err.Error(), "SSRF SECURITY RISK") {
		t.Errorf("expected SSRF SECURITY RISK error for metadata.google.internal, got: %v", err)
	}

	// 3. Safe External Domain (Allowed)
	err = ssrfEng.ValidateURL("https://api.aws.amazon.com/health")
	if err != nil {
		t.Errorf("expected valid domain URL to pass SSRF check, got: %v", err)
	}
}

func testingContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && stringSearch(s, substr)))
}

func stringSearch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
