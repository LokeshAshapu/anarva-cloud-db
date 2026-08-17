package networking_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	activityStream "github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	networkingConn "github.com/anarva-cloud/anarva-cloud-db/internal/networking/connectivity"
	networkingDns "github.com/anarva-cloud/anarva-cloud-db/internal/networking/dns"
	networkingFw "github.com/anarva-cloud/anarva-cloud-db/internal/networking/firewall"
	networkingIpam "github.com/anarva-cloud/anarva-cloud-db/internal/networking/ipam"
	networkingProv "github.com/anarva-cloud/anarva-cloud-db/internal/networking/provider"
	networkingRepo "github.com/anarva-cloud/anarva-cloud-db/internal/networking/repository"
	networkingSvc "github.com/anarva-cloud/anarva-cloud-db/internal/networking/service"
	reliabilityUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
)

func TestWorkflow4_VPCProvisioningAndOperationCompletion(t *testing.T) {
	ctx := context.Background()

	repo := networkingRepo.NewPostgresNetworkingRepository(nil)
	prov := networkingProv.NewLocalDockerNetworkProvider()
	ipamSvc := networkingIpam.NewIPAMService()
	fwSvc := networkingFw.NewFirewallService()
	dnsProv := networkingDns.NewLocalDNSProvider()
	connSvc := networkingConn.NewConnectivityService()
	actStream := activityStream.NewStream()
	relUC := reliabilityUsecase.NewReliabilityUseCase()

	svc := networkingSvc.NewNetworkingService(repo, prov, ipamSvc, fwSvc, dnsProv, connSvc, actStream)
	svc.SetReliabilityUseCase(relUC)

	orgID := "org-test-wf4"
	projID := "proj-test-wf4"
	vpcName := "anarva-prod-vpc"
	cidr := "10.0.0.0/16"

	// 1. Create VPC Network
	vNet, err := svc.CreateNetwork(ctx, orgID, projID, vpcName, "us-east-1", cidr)
	require.NoError(t, err)
	assert.NotEmpty(t, vNet.ID)
	assert.Equal(t, vpcName, vNet.Name)
	assert.Equal(t, cidr, vNet.CIDR)
	assert.Equal(t, "AVAILABLE", string(vNet.Status))

	// 2. Verify state persistence via ListNetworks
	nets, errList := svc.ListNetworks(ctx, orgID, projID)
	require.NoError(t, errList)
	assert.Len(t, nets, 1)
	assert.Equal(t, vNet.ID, nets[0].ID)

	// 3. Verify Default Security Group Creation
	sgs, errSG := repo.ListSecurityGroups(vNet.ID)
	require.NoError(t, errSG)
	assert.Len(t, sgs, 1)
	assert.Equal(t, "default", sgs[0].Name)

	// 4. Invalid CIDR format rejection
	_, errInvalidCIDR := svc.CreateNetwork(ctx, orgID, projID, "bad-vpc", "us-east-1", "invalid_cidr")
	assert.Error(t, errInvalidCIDR)
}
