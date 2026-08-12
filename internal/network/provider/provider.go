package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/network/domain"
)

type NetworkProvider interface {
	GetProviderType() string
	CreateNetwork(ctx context.Context, net *domain.Network) (*domain.Network, error)
	DeleteNetwork(ctx context.Context, id string) error
	CreateSubnet(ctx context.Context, sub *domain.Subnet) (*domain.Subnet, error)
	DeleteSubnet(ctx context.Context, id string) error
	AllocateIP(ctx context.Context, subnetID string, ipType domain.IPType) (*domain.IPAddress, error)
	ReleaseIP(ctx context.Context, ipID string) error
	CreateLoadBalancer(ctx context.Context, lb *domain.LoadBalancer) (*domain.LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, id string) error
	GetNetworkHealth(ctx context.Context, id string) (string, error)
}

type LocalDockerNetworkProvider struct {
	mu        sync.RWMutex
	networks  map[string]*domain.Network
	subnets   map[string]*domain.Subnet
	ips       map[string]*domain.IPAddress
	lbs       map[string]*domain.LoadBalancer
	hasDocker bool
}

func NewLocalDockerNetworkProvider() *LocalDockerNetworkProvider {
	p := &LocalDockerNetworkProvider{
		networks: make(map[string]*domain.Network),
		subnets:  make(map[string]*domain.Subnet),
		ips:      make(map[string]*domain.IPAddress),
		lbs:      make(map[string]*domain.LoadBalancer),
	}
	_, err := exec.LookPath("docker")
	p.hasDocker = (err == nil)
	return p
}

func (p *LocalDockerNetworkProvider) GetProviderType() string {
	return "LOCAL_DOCKER"
}

func (p *LocalDockerNetworkProvider) CreateNetwork(ctx context.Context, net *domain.Network) (*domain.Network, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	net.Provider = "LOCAL_DOCKER"
	net.Status = domain.StatusAvailable

	if p.hasDocker && net.CIDR != "" {
		netName := fmt.Sprintf("anarva-net-%s", net.Slug)
		cmd := exec.CommandContext(ctx, "docker", "network", "create", "--driver", "bridge", "--subnet", net.CIDR, netName)
		out, err := cmd.CombinedOutput()
		if err == nil {
			net.ProviderNetworkID = strings.TrimSpace(string(out))
		} else {
			net.ProviderNetworkID = fmt.Sprintf("docker-net-sim-%s", net.ID)
		}
	} else {
		net.ProviderNetworkID = fmt.Sprintf("local-net-sim-%s", net.ID)
	}

	p.networks[net.ID] = net
	return net, nil
}

func (p *LocalDockerNetworkProvider) DeleteNetwork(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	net, ok := p.networks[id]
	if !ok {
		return fmt.Errorf("network not found")
	}

	if p.hasDocker && net.ProviderNetworkID != "" && !strings.HasPrefix(net.ProviderNetworkID, "local-net-sim-") {
		_ = exec.CommandContext(ctx, "docker", "network", "rm", net.ProviderNetworkID).Run()
	}

	now := time.Now()
	net.Status = domain.StatusDeleted
	net.DeletedAt = &now
	return nil
}

func (p *LocalDockerNetworkProvider) CreateSubnet(ctx context.Context, sub *domain.Subnet) (*domain.Subnet, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	sub.Status = "AVAILABLE"
	sub.ProviderSubnetID = fmt.Sprintf("sub-sim-%s", sub.ID)
	p.subnets[sub.ID] = sub
	return sub, nil
}

func (p *LocalDockerNetworkProvider) DeleteSubnet(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	sub, ok := p.subnets[id]
	if !ok {
		return fmt.Errorf("subnet not found")
	}
	now := time.Now()
	sub.Status = "DELETED"
	sub.DeletedAt = &now
	return nil
}

func (p *LocalDockerNetworkProvider) AllocateIP(ctx context.Context, subnetID string, ipType domain.IPType) (*domain.IPAddress, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	addr := fmt.Sprintf("10.0.1.%d", len(p.ips)+10)
	if ipType == domain.IPTypePublic {
		addr = fmt.Sprintf("20.198.42.%d", len(p.ips)+10)
	}

	ip := &domain.IPAddress{
		ID:        fmt.Sprintf("ip-%d", len(p.ips)+101),
		SubnetID:  subnetID,
		Address:   addr,
		Version:   domain.IPv4,
		Type:      ipType,
		Status:    domain.IPStatusAllocated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	p.ips[ip.ID] = ip
	return ip, nil
}

func (p *LocalDockerNetworkProvider) ReleaseIP(ctx context.Context, ipID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ip, ok := p.ips[ipID]; ok {
		ip.Status = domain.IPStatusReleased
		ip.UpdatedAt = time.Now()
	}
	return nil
}

func (p *LocalDockerNetworkProvider) CreateLoadBalancer(ctx context.Context, lb *domain.LoadBalancer) (*domain.LoadBalancer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	lb.Status = "AVAILABLE"
	lb.Provider = "LOCAL_DOCKER"
	lb.DNSName = fmt.Sprintf("lb-%s.anarva.cloud", lb.ID)
	p.lbs[lb.ID] = lb
	return lb, nil
}

func (p *LocalDockerNetworkProvider) DeleteLoadBalancer(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if lb, ok := p.lbs[id]; ok {
		lb.Status = "DELETED"
	}
	return nil
}

func (p *LocalDockerNetworkProvider) GetNetworkHealth(ctx context.Context, id string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, ok := p.networks[id]; ok {
		return "HEALTHY", nil
	}
	return "UNKNOWN", nil
}
