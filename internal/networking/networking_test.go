package networking_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	repo := repository.NewPostgresNetworkingRepository(nil)
	prov := provider.NewLocalDockerNetworkProvider()
	ipamSvc := ipam.NewIPAMService()
	fwSvc := firewall.NewFirewallService()
	connSvc := connectivity.NewConnectivityService()

	svc := service.NewNetworkingService(repo, prov, ipamSvc, fwSvc, nil, connSvc, nil)

	// 1. Create VPC Network
	net, err := svc.CreateNetwork(ctx, "org-test", "proj-test", "primary-vpc", "us-east-1", "10.0.0.0/16")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusAvailable, net.Status)

	// 2. IPAM CIDR Validation
	require.NoError(t, ipamSvc.ValidateCIDR("10.0.0.0/16"))
	assert.Error(t, ipamSvc.ValidateCIDR("invalid-cidr-string"))

	// 3. CIDR Overlap Detection
	err = ipamSvc.CheckCIDROverlap([]string{"10.0.0.0/16"}, "10.0.1.0/24")
	assert.Error(t, err)

	// 4. IP Allocation
	alloc, err := ipamSvc.Allocate(net.ID, "sub-01", "ace-worker-1", domain.IPv4)
	require.NoError(t, err)
	assert.NotEmpty(t, alloc.IP)

	// 5. Delete Network
	require.NoError(t, svc.DeleteNetwork(ctx, net.ID))
}

func TestNetworking_SubnetContainmentAndOverlap(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewPostgresNetworkingRepository(nil)
	prov := provider.NewLocalDockerNetworkProvider()
	ipamSvc := ipam.NewIPAMService()
	fwSvc := firewall.NewFirewallService()
	connSvc := connectivity.NewConnectivityService()

	svc := service.NewNetworkingService(repo, prov, ipamSvc, fwSvc, nil, connSvc, nil)

	// Create Parent VPC (10.0.0.0/16)
	vpc, err := svc.CreateNetwork(ctx, "org-alpha", "proj-alpha", "vpc-alpha", "us-east-1", "10.0.0.0/16")
	require.NoError(t, err)

	// Valid Subnet Creation (10.0.1.0/24 inside 10.0.0.0/16)
	sub1, err := svc.CreateSubnet(ctx, "org-alpha", "proj-alpha", vpc.ID, "subnet-1", "10.0.1.0/24", "us-east-1a", domain.SubnetPrivate)
	require.NoError(t, err)
	assert.Equal(t, "10.0.1.0/24", sub1.CIDR)

	// Out of Bounds Subnet Creation (10.1.0.0/24 outside 10.0.0.0/16) MUST Fail
	_, err = svc.CreateSubnet(ctx, "org-alpha", "proj-alpha", vpc.ID, "subnet-out-of-bounds", "10.1.0.0/24", "us-east-1a", domain.SubnetPrivate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SUBNET_CIDR_OUT_OF_BOUNDS")

	// Overlapping Subnet Creation (10.0.1.0/24 again) MUST Fail
	_, err = svc.CreateSubnet(ctx, "org-alpha", "proj-alpha", vpc.ID, "subnet-overlap", "10.0.1.0/24", "us-east-1b", domain.SubnetPrivate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CIDR overlap detected")
}

func TestNetworking_TenantIsolationAndVpcNotEmpty(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewPostgresNetworkingRepository(nil)
	prov := provider.NewLocalDockerNetworkProvider()
	ipamSvc := ipam.NewIPAMService()
	fwSvc := firewall.NewFirewallService()
	connSvc := connectivity.NewConnectivityService()

	svc := service.NewNetworkingService(repo, prov, ipamSvc, fwSvc, nil, connSvc, nil)

	// Create VPC for Org Alpha
	vpc, err := svc.CreateNetwork(ctx, "org-alpha", "proj-alpha", "vpc-alpha", "us-east-1", "10.0.0.0/16")
	require.NoError(t, err)

	// Org Beta attempt to create subnet inside Org Alpha's VPC MUST fail
	_, err = svc.CreateSubnet(ctx, "org-beta", "proj-beta", vpc.ID, "subnet-illegal", "10.0.2.0/24", "us-east-1a", domain.SubnetPrivate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TENANT_ISOLATION_VIOLATION")

	// Create subnet for Org Alpha
	_, err = svc.CreateSubnet(ctx, "org-alpha", "proj-alpha", vpc.ID, "subnet-alpha", "10.0.2.0/24", "us-east-1a", domain.SubnetPrivate)
	require.NoError(t, err)

	// Delete VPC with active subnets MUST fail with VPC_NOT_EMPTY
	err = svc.DeleteNetwork(ctx, vpc.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VPC_NOT_EMPTY")
}

func TestNetworking_RouteTableAndRoutes(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewPostgresNetworkingRepository(nil)
	prov := provider.NewLocalDockerNetworkProvider()
	ipamSvc := ipam.NewIPAMService()
	fwSvc := firewall.NewFirewallService()
	connSvc := connectivity.NewConnectivityService()

	svc := service.NewNetworkingService(repo, prov, ipamSvc, fwSvc, nil, connSvc, nil)

	vpc, err := svc.CreateNetwork(ctx, "org-test", "proj-test", "vpc-rt-test", "us-east-1", "10.0.0.0/16")
	require.NoError(t, err)

	rt, err := svc.CreateRouteTable(ctx, "org-test", "proj-test", vpc.ID, "public-rt")
	require.NoError(t, err)
	assert.Equal(t, "public-rt", rt.Name)
	assert.NotEmpty(t, rt.Routes)

	// Add Internet Gateway Route (0.0.0.0/0)
	updatedRt, err := svc.AddRoute(ctx, rt.ID, "0.0.0.0/0", "igw-01", domain.TargetInternetGateway)
	require.NoError(t, err)
	assert.Equal(t, 2, len(updatedRt.Routes))
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
	require.NoError(t, fwSvc.ValidateSecurityRule(safeRule))

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
	assert.Error(t, err)
}

func TestNetworking_SSRFProtection(t *testing.T) {
	ctx := context.Background()
	connSvc := connectivity.NewConnectivityService()

	// 1. Cloud Metadata Endpoint (Blocked)
	_, err := connSvc.TestConnectivity(ctx, "vm-01", "169.254.169.254", 80)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSRF BLOCKED")

	// 2. Loopback Control Plane Endpoint (Blocked)
	_, err = connSvc.TestConnectivity(ctx, "vm-01", "127.0.0.1", 8080)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSRF BLOCKED")

	// 3. Valid Internal Workload (Allowed)
	res, err := connSvc.TestConnectivity(ctx, "vm-01", "10.0.1.15", 5432)
	require.NoError(t, err)
	assert.True(t, res.Reachable)
}

func TestNetworking_ReconciliationDrift(t *testing.T) {
	ctx := context.Background()
	prov := provider.NewLocalDockerNetworkProvider()
	recSvc := reconciliation.NewReconciliationService(prov)

	desiredNet := &domain.VirtualNetwork{
		ID:     "vpc-non-existent",
		CIDR:   "10.0.0.0/16",
		Name:   "ghost-vpc",
		Status: domain.StatusAvailable,
	}

	res, err := recSvc.Reconcile(ctx, desiredNet)
	require.NoError(t, err)
	assert.True(t, res.DriftDetected)
}
