package repository

import (
	"fmt"
	"sync"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type NetworkingRepository struct {
	mu       sync.RWMutex
	networks map[string]*domain.VirtualNetwork
	subnets  map[string]*domain.Subnet
	routes   map[string]*domain.RouteTable
	sgs      map[string]*domain.SecurityGroup
	nics     map[string]*domain.NetworkInterface
	zones    map[string]*domain.DNSZone
	records  map[string]*domain.DNSRecord
}

func NewNetworkingRepository() *NetworkingRepository {
	return &NetworkingRepository{
		networks: make(map[string]*domain.VirtualNetwork),
		subnets:  make(map[string]*domain.Subnet),
		routes:   make(map[string]*domain.RouteTable),
		sgs:      make(map[string]*domain.SecurityGroup),
		nics:     make(map[string]*domain.NetworkInterface),
		zones:    make(map[string]*domain.DNSZone),
		records:  make(map[string]*domain.DNSRecord),
	}
}

func (r *NetworkingRepository) SaveNetwork(net *domain.VirtualNetwork) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.networks[net.ID] = net
	return nil
}

func (r *NetworkingRepository) GetNetwork(id string) (*domain.VirtualNetwork, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if net, ok := r.networks[id]; ok {
		return net, nil
	}
	return nil, fmt.Errorf("network not found")
}

func (r *NetworkingRepository) ListNetworks(orgID, projectID string) ([]*domain.VirtualNetwork, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.VirtualNetwork
	for _, net := range r.networks {
		if net.Status != domain.StatusDeleted {
			res = append(res, net)
		}
	}
	return res, nil
}

func (r *NetworkingRepository) SaveSecurityGroup(sg *domain.SecurityGroup) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sgs[sg.ID] = sg
	return nil
}

func (r *NetworkingRepository) GetSecurityGroup(id string) (*domain.SecurityGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if sg, ok := r.sgs[id]; ok {
		return sg, nil
	}
	return nil, fmt.Errorf("security group not found")
}

func (r *NetworkingRepository) ListSecurityGroups(networkID string) ([]*domain.SecurityGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.SecurityGroup
	for _, sg := range r.sgs {
		if networkID == "" || sg.NetworkID == networkID {
			res = append(res, sg)
		}
	}
	return res, nil
}
