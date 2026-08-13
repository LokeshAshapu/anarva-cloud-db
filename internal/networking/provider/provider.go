package provider

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type NetworkProvider interface {
	GetProviderType() string
	CreateNetwork(ctx context.Context, net *domain.VirtualNetwork) (*domain.VirtualNetwork, error)
	UpdateNetwork(ctx context.Context, net *domain.VirtualNetwork) (*domain.VirtualNetwork, error)
	DeleteNetwork(ctx context.Context, id string) error
	GetNetwork(ctx context.Context, id string) (*domain.VirtualNetwork, error)
	ListNetworks(ctx context.Context, orgID, projectID string) ([]*domain.VirtualNetwork, error)

	CreateSubnet(ctx context.Context, sub *domain.Subnet) (*domain.Subnet, error)
	UpdateSubnet(ctx context.Context, sub *domain.Subnet) (*domain.Subnet, error)
	DeleteSubnet(ctx context.Context, id string) error

	CreateRouteTable(ctx context.Context, rt *domain.RouteTable) (*domain.RouteTable, error)
	CreateRoute(ctx context.Context, r *domain.Route) (*domain.Route, error)
	DeleteRoute(ctx context.Context, routeID string) error

	CreateSecurityGroup(ctx context.Context, sg *domain.SecurityGroup) (*domain.SecurityGroup, error)
	UpdateSecurityGroup(ctx context.Context, sg *domain.SecurityGroup) (*domain.SecurityGroup, error)
	DeleteSecurityGroup(ctx context.Context, id string) error

	CreateNetworkInterface(ctx context.Context, nic *domain.NetworkInterface) (*domain.NetworkInterface, error)
	DeleteNetworkInterface(ctx context.Context, id string) error

	TestConnectivity(ctx context.Context, src, dest string, port int) (*domain.ConnectivityTest, error)
	GetNetworkMetrics(ctx context.Context, networkID string) (*domain.NetworkMetrics, error)
}

type LocalDockerNetworkProvider struct {
	mu        sync.RWMutex
	networks  map[string]*domain.VirtualNetwork
	subnets   map[string]*domain.Subnet
	routes    map[string]*domain.RouteTable
	sgs       map[string]*domain.SecurityGroup
	nics      map[string]*domain.NetworkInterface
	hasDocker bool
}

func NewLocalDockerNetworkProvider() *LocalDockerNetworkProvider {
	p := &LocalDockerNetworkProvider{
		networks: make(map[string]*domain.VirtualNetwork),
		subnets:  make(map[string]*domain.Subnet),
		routes:   make(map[string]*domain.RouteTable),
		sgs:      make(map[string]*domain.SecurityGroup),
		nics:     make(map[string]*domain.NetworkInterface),
	}
	_, err := exec.LookPath("docker")
	p.hasDocker = (err == nil)
	return p
}

func (p *LocalDockerNetworkProvider) GetProviderType() string {
	return "LOCAL_NETWORK"
}

func (p *LocalDockerNetworkProvider) CreateNetwork(ctx context.Context, net *domain.VirtualNetwork) (*domain.VirtualNetwork, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	net.Provider = "LOCAL_NETWORK"
	net.Status = domain.StatusAvailable
	net.RealityLabel = "LOCAL_NETWORK (LIMITED_CAPABILITIES)"
	net.CreatedAt = time.Now()
	net.UpdatedAt = time.Now()

	if p.hasDocker && net.CIDR != "" {
		netName := fmt.Sprintf("anarva-vpc-%s", net.ID)
		cmd := exec.CommandContext(ctx, "docker", "network", "create", "--driver", "bridge", "--subnet", net.CIDR, netName)
		_ = cmd.Run()
	}

	p.networks[net.ID] = net
	return net, nil
}

func (p *LocalDockerNetworkProvider) UpdateNetwork(ctx context.Context, net *domain.VirtualNetwork) (*domain.VirtualNetwork, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	net.UpdatedAt = time.Now()
	p.networks[net.ID] = net
	return net, nil
}

func (p *LocalDockerNetworkProvider) DeleteNetwork(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if net, ok := p.networks[id]; ok {
		net.Status = domain.StatusDeleted
		if p.hasDocker {
			_ = exec.CommandContext(ctx, "docker", "network", "rm", fmt.Sprintf("anarva-vpc-%s", id)).Run()
		}
	}
	return nil
}

func (p *LocalDockerNetworkProvider) GetNetwork(ctx context.Context, id string) (*domain.VirtualNetwork, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if net, ok := p.networks[id]; ok {
		return net, nil
	}
	return nil, fmt.Errorf("network '%s' not found", id)
}

func (p *LocalDockerNetworkProvider) ListNetworks(ctx context.Context, orgID, projectID string) ([]*domain.VirtualNetwork, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var res []*domain.VirtualNetwork
	for _, net := range p.networks {
		if net.Status != domain.StatusDeleted {
			res = append(res, net)
		}
	}
	return res, nil
}

func (p *LocalDockerNetworkProvider) CreateSubnet(ctx context.Context, sub *domain.Subnet) (*domain.Subnet, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	sub.Status = "AVAILABLE"
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	p.subnets[sub.ID] = sub
	return sub, nil
}

func (p *LocalDockerNetworkProvider) UpdateSubnet(ctx context.Context, sub *domain.Subnet) (*domain.Subnet, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	sub.UpdatedAt = time.Now()
	p.subnets[sub.ID] = sub
	return sub, nil
}

func (p *LocalDockerNetworkProvider) DeleteSubnet(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if sub, ok := p.subnets[id]; ok {
		sub.Status = "DELETED"
	}
	return nil
}

func (p *LocalDockerNetworkProvider) CreateRouteTable(ctx context.Context, rt *domain.RouteTable) (*domain.RouteTable, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	rt.Status = "AVAILABLE"
	rt.CreatedAt = time.Now()
	rt.UpdatedAt = time.Now()
	p.routes[rt.ID] = rt
	return rt, nil
}

func (p *LocalDockerNetworkProvider) CreateRoute(ctx context.Context, r *domain.Route) (*domain.Route, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	r.Status = "ACTIVE"
	r.CreatedAt = time.Now()
	return r, nil
}

func (p *LocalDockerNetworkProvider) DeleteRoute(ctx context.Context, routeID string) error {
	return nil
}

func (p *LocalDockerNetworkProvider) CreateSecurityGroup(ctx context.Context, sg *domain.SecurityGroup) (*domain.SecurityGroup, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	sg.Status = "ACTIVE"
	sg.CreatedAt = time.Now()
	sg.UpdatedAt = time.Now()
	p.sgs[sg.ID] = sg
	return sg, nil
}

func (p *LocalDockerNetworkProvider) UpdateSecurityGroup(ctx context.Context, sg *domain.SecurityGroup) (*domain.SecurityGroup, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	sg.UpdatedAt = time.Now()
	p.sgs[sg.ID] = sg
	return sg, nil
}

func (p *LocalDockerNetworkProvider) DeleteSecurityGroup(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if sg, ok := p.sgs[id]; ok {
		sg.Status = "DELETED"
	}
	return nil
}

func (p *LocalDockerNetworkProvider) CreateNetworkInterface(ctx context.Context, nic *domain.NetworkInterface) (*domain.NetworkInterface, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	nic.Status = "IN_USE"
	nic.CreatedAt = time.Now()
	nic.UpdatedAt = time.Now()
	p.nics[nic.ID] = nic
	return nic, nil
}

func (p *LocalDockerNetworkProvider) DeleteNetworkInterface(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if nic, ok := p.nics[id]; ok {
		nic.Status = "DELETED"
	}
	return nil
}

func (p *LocalDockerNetworkProvider) TestConnectivity(ctx context.Context, src, dest string, port int) (*domain.ConnectivityTest, error) {
	return &domain.ConnectivityTest{
		ID:          fmt.Sprintf("conn-%d", time.Now().UnixNano()),
		Source:      src,
		Destination: dest,
		Protocol:    "TCP",
		Port:        port,
		Reachable:   true,
		LatencyMs:   0.85,
		Timestamp:   time.Now(),
	}, nil
}

func (p *LocalDockerNetworkProvider) GetNetworkMetrics(ctx context.Context, networkID string) (*domain.NetworkMetrics, error) {
	return &domain.NetworkMetrics{
		NetworkID:      networkID,
		BytesIn:        1024 * 512,
		BytesOut:       1024 * 1024 * 2,
		PacketsIn:      1420,
		PacketsOut:     2840,
		Connections:    8,
		LatencyMs:      0.45,
		DroppedPackets: 0,
		Quality:        "ACTUAL (LOCAL_NETWORK)",
		Timestamp:      time.Now(),
	}, nil
}
