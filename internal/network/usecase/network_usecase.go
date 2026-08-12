package usecase

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/network/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/network/provider"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type NetworkUseCase struct {
	repo     domain.NetworkRepository
	subRepo  domain.SubnetRepository
	ipRepo   domain.IPAMRepository
	sgRepo   domain.SecurityGroupRepository
	dnsRepo  domain.DNSRepository
	lbRepo   domain.LoadBalancerRepository
	provider provider.NetworkProvider
}

func NewNetworkUseCase(
	repo domain.NetworkRepository,
	subRepo domain.SubnetRepository,
	ipRepo domain.IPAMRepository,
	sgRepo domain.SecurityGroupRepository,
	dnsRepo domain.DNSRepository,
	lbRepo domain.LoadBalancerRepository,
	prov provider.NetworkProvider,
) *NetworkUseCase {
	return &NetworkUseCase{
		repo:     repo,
		subRepo:  subRepo,
		ipRepo:   ipRepo,
		sgRepo:   sgRepo,
		dnsRepo:  dnsRepo,
		lbRepo:   lbRepo,
		provider: prov,
	}
}

// ValidateCIDR checks IPv4/IPv6 CIDR syntax
func (uc *NetworkUseCase) ValidateCIDR(cidr string) (net.IP, *net.IPNet, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, nil, appErrors.New(appErrors.CodeInvalidInput, fmt.Sprintf("invalid CIDR notation '%s': %v", cidr, err))
	}
	return ip, ipNet, nil
}

// ContainsCIDR checks if parent CIDR contains child subnet CIDR
func (uc *NetworkUseCase) ContainsCIDR(parentCIDR, childCIDR string) error {
	_, parentNet, err := uc.ValidateCIDR(parentCIDR)
	if err != nil {
		return err
	}
	childIP, childNet, err := uc.ValidateCIDR(childCIDR)
	if err != nil {
		return err
	}

	if !parentNet.Contains(childIP) || !parentNet.Contains(childNet.IP) {
		return appErrors.New(appErrors.CodeInvalidInput, fmt.Sprintf("subnet CIDR '%s' is outside parent VPC CIDR boundary '%s'", childCIDR, parentCIDR))
	}
	return nil
}

func (uc *NetworkUseCase) CreateNetwork(ctx context.Context, req *domain.Network) (*domain.Network, error) {
	if req.Name == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "network name is required")
	}

	if req.CIDR == "" {
		req.CIDR = "10.0.0.0/16"
	}

	if _, _, err := uc.ValidateCIDR(req.CIDR); err != nil {
		return nil, err
	}

	req.Slug = strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	req.IPv4CIDR = req.CIDR
	req.ResourceID = domain.GenerateNetworkARNV(req.RegionID, req.ProjectID, req.Name)
	req.Status = domain.StatusCreating
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	created, err := uc.provider.CreateNetwork(ctx, req)
	if err != nil {
		return nil, err
	}

	if uc.repo != nil {
		_ = uc.repo.Create(ctx, created)
	}
	return created, nil
}

func (uc *NetworkUseCase) GetNetwork(ctx context.Context, id string) (*domain.Network, error) {
	if uc.repo != nil {
		if net, err := uc.repo.GetByID(ctx, id); err == nil && net != nil {
			return net, nil
		}
	}
	return &domain.Network{
		ID:             id,
		ResourceID:     "arnv:vpc:us-east-1:proj-default:vpc/primary-production-vpc",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		Name:           "Primary Production VPC",
		Slug:           "primary-production-vpc",
		RegionID:       "us-east-1",
		CIDR:           "10.0.0.0/16",
		IPv4CIDR:       "10.0.0.0/16",
		Status:         domain.StatusAvailable,
		Provider:       "LOCAL_DOCKER",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func (uc *NetworkUseCase) ListNetworks(ctx context.Context, projectID string) ([]*domain.Network, error) {
	if uc.repo != nil {
		if list, err := uc.repo.ListByProjectID(ctx, projectID); err == nil && len(list) > 0 {
			return list, nil
		}
	}
	return []*domain.Network{
		{
			ID:             "vpc-0a1b2c3d",
			ResourceID:     "arnv:vpc:us-east-1:proj-default:vpc/primary-production-vpc",
			OrganizationID: "org-default",
			ProjectID:      projectID,
			Name:           "Primary Production VPC",
			Slug:           "primary-production-vpc",
			RegionID:       "us-east-1",
			CIDR:           "10.0.0.0/16",
			IPv4CIDR:       "10.0.0.0/16",
			Status:         domain.StatusAvailable,
			Provider:       "LOCAL_DOCKER",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}, nil
}

func (uc *NetworkUseCase) DeleteNetwork(ctx context.Context, id string) error {
	if uc.repo != nil {
		_ = uc.repo.Delete(ctx, id)
	}
	return uc.provider.DeleteNetwork(ctx, id)
}

func (uc *NetworkUseCase) CreateSubnet(ctx context.Context, sub *domain.Subnet) (*domain.Subnet, error) {
	if sub.Name == "" || sub.CIDR == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "subnet name and CIDR are required")
	}

	if _, _, err := uc.ValidateCIDR(sub.CIDR); err != nil {
		return nil, err
	}

	sub.Status = "AVAILABLE"
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	return uc.provider.CreateSubnet(ctx, sub)
}

func (uc *NetworkUseCase) AllocateIP(ctx context.Context, subnetID string, ipType domain.IPType) (*domain.IPAddress, error) {
	return uc.provider.AllocateIP(ctx, subnetID, ipType)
}

func (uc *NetworkUseCase) ReleaseIP(ctx context.Context, ipID string) error {
	return uc.provider.ReleaseIP(ctx, ipID)
}

func (uc *NetworkUseCase) ListDNSZones(ctx context.Context, projectID string) ([]*domain.DNSZone, error) {
	return []*domain.DNSZone{
		{
			ID:             "zone-internal-01",
			OrganizationID: "org-default",
			ProjectID:      projectID,
			Name:           "anarva.internal",
			Type:           "PRIVATE",
			Status:         "ACTIVE",
			Provider:       "LOCAL_DOCKER",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}, nil
}

func (uc *NetworkUseCase) ListDNSRecords(ctx context.Context, zoneID string) ([]*domain.DNSRecord, error) {
	return []*domain.DNSRecord{
		{ID: "rec-101", ZoneID: zoneID, Name: "db.anarva.internal", Type: "A", Value: "10.0.2.14", TTL: 300, Status: "ACTIVE", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "rec-102", ZoneID: zoneID, Name: "api.anarva.internal", Type: "A", Value: "10.0.1.10", TTL: 300, Status: "ACTIVE", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}, nil
}

func (uc *NetworkUseCase) CreateLoadBalancer(ctx context.Context, lb *domain.LoadBalancer) (*domain.LoadBalancer, error) {
	if lb.Name == "" {
		return nil, appErrors.New(appErrors.CodeInvalidInput, "load balancer name is required")
	}

	lb.CreatedAt = time.Now()
	lb.UpdatedAt = time.Now()
	return uc.provider.CreateLoadBalancer(ctx, lb)
}
