package networking_test

import (
	"context"
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/connectivity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/firewall"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/ipam"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/reconciliation"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/repository"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/service"
)

func TestNetworking_VPCNetworkLifecycleAndIPAM(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewNetworkingRepository()
	prov := provider.NewLocalDockerNetworkProvider()
	ipamSvc := ipam.NewIPAMService()
	fwSvc := firewall.NewFirewallService()
	connSvc := connectivity.NewConnectivityService()

	svc := service.NewNetworkingService(repo, prov, ipamSvc, fwSvc, nil, connSvc, nil)

	// 1. Create VPC Network
	net, err := svc.CreateNetwork(ctx, "org-test", "proj-test", "primary-vpc", "ap-hyderabad-1", "10.0.0.0/16")
	if err != nil {
		t.Fatalf("failed to create VPC network: %v", err)
	}
	if net.Status != domain.StatusAvailable {
		t.Errorf("expected StatusAvailable, got %s", net.Status)
	}

	// 2. IPAM CIDR Validation
	if err := ipamSvc.ValidateCIDR("10.0.0.0/16"); err != nil {
		t.Errorf("expected valid CIDR, got %v", err)
	}
	if err := ipamSvc.ValidateCIDR("invalid-cidr-string"); err == nil {
		t.Errorf("expected error for invalid CIDR, got nil")
	}

	// 3. CIDR Overlap Detection
	err = ipamSvc.CheckCIDROverlap([]string{"10.0.0.0/16"}, "10.0.1.0/24")
	if err == nil {
		t.Errorf("expected CIDR overlap error, got nil")
	}

	// 4. IP Allocation
	alloc, err := ipamSvc.Allocate(net.ID, "sub-01", "ace-worker-1", domain.IPv4)
	if err != nil || alloc.IP == "" {
		t.Fatalf("IP allocation failed: %v", err)
	}

	// 5. Delete Network
	if err := svc.DeleteNetwork(ctx, net.ID); err != nil {
		t.Errorf("failed to delete network: %v", err)
	}
}

func TestNetworking_FirewallDatabasePortSecurity(t *testing.T) {
	fwSvc := firewall.NewFirewallService()

	// Safe Inbound Rule
	safeRule := &domain.SecurityRule{
		Direction:   domain.DirectionIngress,
		Protocol:    "TCP",
		FromPort:    5432,
		ToPort:      5432,
		Source:      "10.0.0.0/16",
		Action:      domain.ActionAllow,
		Priority:    100,
		Description: "Allow internal VPC DB traffic",
	}
	if err := fwSvc.ValidateSecurityRule(safeRule); err != nil {
		t.Errorf("expected safe rule to pass, got err: %v", err)
	}

	// Unrestricted Public Access Database Rule (0.0.0.0/0 on 5432)
	publicDbRule := &domain.SecurityRule{
		Direction:   domain.DirectionIngress,
		Protocol:    "TCP",
		FromPort:    5432,
		ToPort:      5432,
		Source:      "0.0.0.0/0",
		Action:      domain.ActionAllow,
		Priority:    100,
		Description: "Public PostgreSQL Access",
	}
	err := fwSvc.ValidateSecurityRule(publicDbRule)
	if err == nil {
		t.Errorf("expected security risk error for 0.0.0.0/0 on port 5432, got nil")
	}
}

func TestNetworking_SSRFProtection(t *testing.T) {
	ctx := context.Background()
	connSvc := connectivity.NewConnectivityService()

	// 1. Cloud Metadata Endpoint (Blocked)
	_, err := connSvc.TestConnectivity(ctx, "vm-01", "169.254.169.254", 80)
	if err == nil || !testingContains(err.Error(), "SSRF BLOCKED") {
		t.Errorf("expected SSRF BLOCKED error for metadata endpoint 169.254.169.254, got: %v", err)
	}

	// 2. Loopback Control Plane Endpoint (Blocked)
	_, err = connSvc.TestConnectivity(ctx, "vm-01", "127.0.0.1", 8080)
	if err == nil || !testingContains(err.Error(), "SSRF BLOCKED") {
		t.Errorf("expected SSRF BLOCKED error for loopback 127.0.0.1, got: %v", err)
	}

	// 3. Valid Internal Workload (Allowed)
	res, err := connSvc.TestConnectivity(ctx, "vm-01", "10.0.1.15", 5432)
	if err != nil || !res.Reachable {
		t.Errorf("expected valid connectivity test to pass, got err: %v", err)
	}
}

func TestNetworking_ReconciliationDrift(t *testing.T) {
	ctx := context.Background()
	prov := provider.NewLocalDockerNetworkProvider()
	recSvc := reconciliation.NewReconciliationService(prov)

	desiredNet := &domain.VirtualNetwork{
		ID:    "vpc-non-existent",
		CIDR:  "10.0.0.0/16",
		Name:  "ghost-vpc",
		Status: domain.StatusAvailable,
	}

	res, err := recSvc.Reconcile(ctx, desiredNet)
	if err != nil {
		t.Fatalf("unexpected reconciliation error: %v", err)
	}

	if !res.DriftDetected {
		t.Errorf("expected drift to be detected for missing provider network")
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
