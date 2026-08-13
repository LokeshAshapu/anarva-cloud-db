package service

import (
	"context"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/connectivity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/dns"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/firewall"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/ipam"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/repository"

	activityStream "github.com/anarva-cloud/anarva-cloud-db/internal/activity"
)

type NetworkingService struct {
	repo         *repository.NetworkingRepository
	provider     provider.NetworkProvider
	ipam         *ipam.IPAMService
	firewall     *firewall.FirewallService
	dnsProvider  dns.DNSProvider
	connectivity *connectivity.ConnectivityService
	actStream    *activityStream.Stream
}

func NewNetworkingService(
	repo *repository.NetworkingRepository,
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

func (s *NetworkingService) CreateNetwork(ctx context.Context, orgID, projectID, name, regionID, cidr string) (*domain.VirtualNetwork, error) {
	if err := domain.ValidateCIDR(cidr); err != nil {
		return nil, err
	}

	netID := fmt.Sprintf("vpc-%d", time.Now().UnixNano())
	vNet := &domain.VirtualNetwork{
		ID:             netID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           name,
		RegionID:       regionID,
		CIDR:           cidr,
		Status:         domain.StatusCreating,
		DNSEnabled:     true,
		RealityLabel:   "LOCAL_NETWORK (LIMITED_CAPABILITIES)",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	created, err := s.provider.CreateNetwork(ctx, vNet)
	if err != nil {
		return nil, err
	}

	_ = s.repo.SaveNetwork(created)

	// Create Default Security Group
	defaultSG := s.firewall.CreateDefaultSecurityGroup(created.ID)
	_, _ = s.provider.CreateSecurityGroup(ctx, defaultSG)
	_ = s.repo.SaveSecurityGroup(defaultSG)

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: orgID,
			ProjectID:      projectID,
			ResourceID:     name,
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

	if err := s.provider.DeleteNetwork(ctx, id); err != nil {
		return err
	}

	net.Status = domain.StatusDeleted
	_ = s.repo.SaveNetwork(net)

	if s.actStream != nil {
		s.actStream.Record(&activityStream.ActivityEvent{
			ID:             fmt.Sprintf("act-%d", time.Now().UnixNano()),
			OrganizationID: net.OrganizationID,
			ProjectID:      net.ProjectID,
			ResourceID:     net.Name,
			ActorID:        "system@anarva.cloud",
			Action:         activityStream.ActionNetworkDeleted,
			Timestamp:      time.Now(),
		})
	}
	return nil
}

func (s *NetworkingService) CreateSecurityGroup(ctx context.Context, networkID, name, description string) (*domain.SecurityGroup, error) {
	sg := &domain.SecurityGroup{
		ID:          fmt.Sprintf("sg-%d", time.Now().UnixNano()),
		NetworkID:   networkID,
		Name:        name,
		Description: description,
		Status:      "ACTIVE",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	created, err := s.provider.CreateSecurityGroup(ctx, sg)
	if err != nil {
		return nil, err
	}

	_ = s.repo.SaveSecurityGroup(created)
	return created, nil
}

func (s *NetworkingService) AddSecurityRule(ctx context.Context, sgID string, rule domain.SecurityRule) (*domain.SecurityGroup, error) {
	if err := s.firewall.ValidateSecurityRule(&rule); err != nil {
		return nil, err
	}

	sg, err := s.repo.GetSecurityGroup(sgID)
	if err != nil {
		return nil, err
	}

	rule.ID = fmt.Sprintf("rule-%d", time.Now().UnixNano())
	rule.SecurityGroupID = sgID
	sg.Rules = append(sg.Rules, rule)

	updated, err := s.provider.UpdateSecurityGroup(ctx, sg)
	if err != nil {
		return nil, err
	}

	_ = s.repo.SaveSecurityGroup(updated)
	return updated, nil
}

func (s *NetworkingService) TestConnectivity(ctx context.Context, source, destination string, port int) (*domain.ConnectivityTest, error) {
	return s.connectivity.TestConnectivity(ctx, source, destination, port)
}
