package service

import (
	"context"
	"fmt"
	"time"

	activityStream "github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/connectivity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/dns"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/firewall"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/ipam"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/repository"
	reliabilityUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type NetworkingService struct {
	repo         *repository.PostgresNetworkingRepository
	provider     provider.NetworkProvider
	ipam         *ipam.IPAMService
	firewall     *firewall.FirewallService
	dnsProvider  dns.DNSProvider
	connectivity *connectivity.ConnectivityService
	actStream    *activityStream.Stream
	relUC        *reliabilityUsecase.ReliabilityUseCase
}

func NewNetworkingService(
	repo *repository.PostgresNetworkingRepository,
	prov provider.NetworkProvider,
	ipamSvc *ipam.IPAMService,
	fwSvc *firewall.FirewallService,
	dnsProv dns.DNSProvider,
	connSvc *connectivity.ConnectivityService,
	actStream *activityStream.Stream,
) *NetworkingService {
	return &NetworkingService{
		repo:         repo,
		provider:     prov,
		ipam:         ipamSvc,
		firewall:     fwSvc,
		dnsProvider:  dnsProv,
		connectivity: connSvc,
		actStream:    actStream,
	}
}

func (s *NetworkingService) SetReliabilityUseCase(relUC *reliabilityUsecase.ReliabilityUseCase) {
	s.relUC = relUC
}

// 1. VPC Management
func (s *NetworkingService) CreateNetwork(ctx context.Context, orgID, projectID, name, regionID, cidr string) (*domain.VirtualNetwork, error) {
	if err := domain.ValidateCIDR(cidr); err != nil {
		return nil, appErrors.New(appErrors.CodeInvalidInput, err.Error())
	}

	netID := fmt.Sprintf("vpc-%d", time.Now().UnixNano()/1e6)

	// Phase 41 Operation Engine Integration
	if s.relUC != nil {
		_, err := s.relUC.DispatchOperation(ctx, orgID, projectID, netID, "CREATE_VPC", "", fmt.Sprintf("%s:%s", name, cidr), fmt.Sprintf("req_vpc_%s", netID))
		if err != nil {
			return nil, err
		}
	}

	vNet := &domain.VirtualNetwork{
		ID:             netID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           name,
		RegionID:       regionID,
		CIDR:           cidr,
		Status:         domain.StatusAvailable,
		DNSEnabled:     true,
		RealityLabel:   "ANARVA_VPC (REAL)",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	created, err := s.provider.CreateNetwork(ctx, vNet)
	if err != nil {
		if s.relUC != nil {
			_, _ = s.relUC.CompleteOperation(ctx, fmt.Sprintf("op-%s", netID), err.Error())
		}
		return nil, err
	}

	if err := s.repo.SaveNetwork(created); err != nil {
		return nil, err
	}

	// Create Default Security Group for VPC
	defaultSG := &domain.SecurityGroup{
		ID:             fmt.Sprintf("sg-default-%s", created.ID),
		NetworkID:      created.ID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           "default",
		Description:    "Default security group for Anarva VPC",
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, _ = s.provider.CreateSecurityGroup(ctx, defaultSG)
	_ = s.repo.SaveSecurityGroup(defaultSG)

	if s.relUC != nil {
		_, _ = s.relUC.CompleteOperation(ctx, fmt.Sprintf("op-%s", netID), "")
	}

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: orgID,
			ProjectID:      projectID,
			ResourceID:     created.ID,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.ActionNetworkCreated,
			Timestamp:      time.Now(),
		})
	}

	return created, nil
}

func (s *NetworkingService) GetNetwork(ctx context.Context, id string) (*domain.VirtualNetwork, error) {
	return s.repo.GetNetwork(id)
}

func (s *NetworkingService) ListNetworks(ctx context.Context, orgID, projectID string) ([]*domain.VirtualNetwork, error) {
	return s.repo.ListNetworks(orgID, projectID)
}

func (s *NetworkingService) DeleteNetwork(ctx context.Context, id string) error {
	net, err := s.repo.GetNetwork(id)
	if err != nil {
		return err
	}

	// Phase 41 Operation Integration
	if s.relUC != nil {
		_, err := s.relUC.DispatchOperation(ctx, net.OrganizationID, net.ProjectID, id, "DELETE_VPC", "", id, fmt.Sprintf("req_del_vpc_%s", id))
		if err != nil {
			return err
		}
	}

	if err := s.repo.DeleteNetwork(id); err != nil {
		if s.relUC != nil {
			_, _ = s.relUC.CompleteOperation(ctx, fmt.Sprintf("op-%s", id), err.Error())
		}
		return err
	}

	_ = s.provider.DeleteNetwork(ctx, id)

	if s.relUC != nil {
		_, _ = s.relUC.CompleteOperation(ctx, fmt.Sprintf("op-%s", id), "")
	}

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: net.OrganizationID,
			ProjectID:      net.ProjectID,
			ResourceID:     id,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.ActionNetworkDeleted,
			Timestamp:      time.Now(),
		})
	}
	return nil
}

// 2. Subnet Management
func (s *NetworkingService) CreateSubnet(ctx context.Context, orgID, projectID, vpcID, name, cidr, zone string, subnetType domain.SubnetType) (*domain.Subnet, error) {
	vpc, err := s.repo.GetNetwork(vpcID)
	if err != nil {
		return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("Target VPC '%s' not found", vpcID))
	}

	if vpc.OrganizationID != orgID || vpc.ProjectID != projectID {
		return nil, appErrors.New(appErrors.CodeForbidden, "TENANT_ISOLATION_VIOLATION: Organization cannot create subnets inside another tenant's VPC")
	}

	// 1. CIDR Format Validation
	if err := domain.ValidateCIDR(cidr); err != nil {
		return nil, appErrors.New(appErrors.CodeInvalidInput, err.Error())
	}

	// 2. CIDR Containment Check (Subnet CIDR must fit strictly inside VPC CIDR)
	if err := domain.ValidateCIDRContainment(vpc.CIDR, cidr); err != nil {
		return nil, appErrors.New(appErrors.CodeInvalidInput, err.Error())
	}

	// 3. Subnet Overlap Check inside same VPC
	existingSubnets, _ := s.repo.ListSubnets(vpcID)
	var existingCIDRs []string
	for _, sn := range existingSubnets {
		existingCIDRs = append(existingCIDRs, sn.CIDR)
	}
	if err := s.ipam.CheckCIDROverlap(existingCIDRs, cidr); err != nil {
		return nil, appErrors.New(appErrors.CodeConflict, err.Error())
	}

	subnetID := fmt.Sprintf("sub-%d", time.Now().UnixNano()/1e6)

	// Phase 41 Operation Integration
	if s.relUC != nil {
		_, err := s.relUC.DispatchOperation(ctx, orgID, projectID, subnetID, "CREATE_SUBNET", "", cidr, fmt.Sprintf("req_sub_%s", subnetID))
		if err != nil {
			return nil, err
		}
	}

	sn := &domain.Subnet{
		ID:               subnetID,
		NetworkID:        vpcID,
		OrganizationID:   orgID,
		ProjectID:        projectID,
		Name:             name,
		CIDR:             cidr,
		AvailabilityZone: zone,
		Type:             subnetType,
		Status:           "AVAILABLE",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.SaveSubnet(sn); err != nil {
		return nil, err
	}

	if s.relUC != nil {
		_, _ = s.relUC.CompleteOperation(ctx, fmt.Sprintf("op-%s", subnetID), "")
	}

	return sn, nil
}

func (s *NetworkingService) ListSubnets(ctx context.Context, vpcID string) ([]*domain.Subnet, error) {
	return s.repo.ListSubnets(vpcID)
}

func (s *NetworkingService) GetSubnet(ctx context.Context, id string) (*domain.Subnet, error) {
	return s.repo.GetSubnet(id)
}

// 3. Security Group Management
func (s *NetworkingService) CreateSecurityGroup(ctx context.Context, orgID, projectID, vpcID, name, description string) (*domain.SecurityGroup, error) {
	_, err := s.repo.GetNetwork(vpcID)
	if err != nil {
		return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("Target VPC '%s' not found", vpcID))
	}

	sgID := fmt.Sprintf("sg-%d", time.Now().UnixNano()/1e6)

	if s.relUC != nil {
		_, err := s.relUC.DispatchOperation(ctx, orgID, projectID, sgID, "CREATE_SECURITY_GROUP", "", name, fmt.Sprintf("req_sg_%s", sgID))
		if err != nil {
			return nil, err
		}
	}

	sg := &domain.SecurityGroup{
		ID:             sgID,
		NetworkID:      vpcID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           name,
		Description:    description,
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	created, err := s.provider.CreateSecurityGroup(ctx, sg)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveSecurityGroup(created); err != nil {
		return nil, err
	}

	if s.relUC != nil {
		_, _ = s.relUC.CompleteOperation(ctx, fmt.Sprintf("op-%s", sgID), "")
	}

	return created, nil
}

func (s *NetworkingService) ListSecurityGroups(ctx context.Context, vpcID string) ([]*domain.SecurityGroup, error) {
	return s.repo.ListSecurityGroups(vpcID)
}

func (s *NetworkingService) AddSecurityRule(ctx context.Context, sgID string, rule domain.SecurityRule) (*domain.SecurityGroup, error) {
	if err := s.firewall.ValidateSecurityRule(&rule); err != nil {
		return nil, appErrors.New(appErrors.CodeInvalidInput, err.Error())
	}

	sg, err := s.repo.GetSecurityGroup(sgID)
	if err != nil {
		return nil, err
	}

	rule.ID = fmt.Sprintf("rule-%d", time.Now().UnixNano()/1e6)
	rule.SecurityGroupID = sgID
	sg.Rules = append(sg.Rules, rule)

	updated, err := s.provider.UpdateSecurityGroup(ctx, sg)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveSecurityGroup(updated); err != nil {
		return nil, err
	}

	return updated, nil
}

// 4. Route Table Management
func (s *NetworkingService) CreateRouteTable(ctx context.Context, orgID, projectID, vpcID, name string) (*domain.RouteTable, error) {
	vpc, err := s.repo.GetNetwork(vpcID)
	if err != nil {
		return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("Target VPC '%s' not found", vpcID))
	}

	rtID := fmt.Sprintf("rt-%d", time.Now().UnixNano()/1e6)

	rt := &domain.RouteTable{
		ID:             rtID,
		NetworkID:      vpcID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           name,
		Status:         "ACTIVE",
		Routes: []domain.Route{
			{
				ID:          fmt.Sprintf("route-local-%s", rtID),
				Destination: vpc.CIDR,
				Target:      "local",
				TargetType:  domain.TargetLocal,
				Status:      "ACTIVE",
				CreatedAt:   time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.SaveRouteTable(rt); err != nil {
		return nil, err
	}

	return rt, nil
}

func (s *NetworkingService) ListRouteTables(ctx context.Context, vpcID string) ([]*domain.RouteTable, error) {
	return s.repo.ListRouteTables(vpcID)
}

func (s *NetworkingService) AddRoute(ctx context.Context, routeTableID, destination, target string, targetType domain.TargetType) (*domain.RouteTable, error) {
	if err := domain.ValidateCIDR(destination); err != nil {
		return nil, appErrors.New(appErrors.CodeInvalidInput, err.Error())
	}

	rt, err := s.repo.GetRouteTable(routeTableID)
	if err != nil {
		return nil, err
	}

	newRoute := domain.Route{
		ID:           fmt.Sprintf("route-%d", time.Now().UnixNano()/1e6),
		RouteTableID: routeTableID,
		Destination:  destination,
		Target:       target,
		TargetType:   targetType,
		Status:       "ACTIVE",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	rt.Routes = append(rt.Routes, newRoute)
	if err := s.repo.SaveRouteTable(rt); err != nil {
		return nil, err
	}

	return rt, nil
}

func (s *NetworkingService) TestConnectivity(ctx context.Context, source, destination string, port int) (*domain.ConnectivityTest, error) {
	return s.connectivity.TestConnectivity(ctx, source, destination, port)
}
