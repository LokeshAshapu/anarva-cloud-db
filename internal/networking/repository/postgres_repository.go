package repository

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type PostgresNetworkingRepository struct {
	db       *gorm.DB
	mu       sync.RWMutex
	networks map[string]*domain.VirtualNetwork
	subnets  map[string]*domain.Subnet
	routes   map[string]*domain.RouteTable
	sgs      map[string]*domain.SecurityGroup
	nics     map[string]*domain.NetworkInterface
}

func NewPostgresNetworkingRepository(db *gorm.DB) *PostgresNetworkingRepository {
	return &PostgresNetworkingRepository{
		db:       db,
		networks: make(map[string]*domain.VirtualNetwork),
		subnets:  make(map[string]*domain.Subnet),
		routes:   make(map[string]*domain.RouteTable),
		sgs:      make(map[string]*domain.SecurityGroup),
		nics:     make(map[string]*domain.NetworkInterface),
	}
}

// 1. Virtual Networks (VPCs)
func (r *PostgresNetworkingRepository) SaveNetwork(net *domain.VirtualNetwork) error {
	if r.db != nil {
		var existing domain.VirtualNetwork
		err := r.db.Where("id = ?", net.ID).First(&existing).Error
		if err == nil {
			net.UpdatedAt = time.Now()
			return r.db.Save(net).Error
		}
		if net.CreatedAt.IsZero() {
			net.CreatedAt = time.Now()
		}
		net.UpdatedAt = time.Now()
		return r.db.Create(net).Error
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.networks[net.ID] = net
	return nil
}

func (r *PostgresNetworkingRepository) GetNetwork(id string) (*domain.VirtualNetwork, error) {
	if r.db != nil {
		var net domain.VirtualNetwork
		if err := r.db.Where("id = ?", id).First(&net).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("VPC '%s' not found", id))
			}
			return nil, err
		}
		return &net, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if net, ok := r.networks[id]; ok {
		return net, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("VPC '%s' not found", id))
}

func (r *PostgresNetworkingRepository) ListNetworks(orgID, projectID string) ([]*domain.VirtualNetwork, error) {
	if r.db != nil {
		var nets []*domain.VirtualNetwork
		query := r.db.Where("status != ?", domain.StatusDeleted)
		if orgID != "" {
			query = query.Where("organization_id = ?", orgID)
		}
		if projectID != "" {
			query = query.Where("project_id = ?", projectID)
		}
		if err := query.Find(&nets).Error; err != nil {
			return nil, err
		}
		return nets, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.VirtualNetwork
	for _, net := range r.networks {
		if net.Status != domain.StatusDeleted {
			if (orgID == "" || net.OrganizationID == orgID) && (projectID == "" || net.ProjectID == projectID) {
				res = append(res, net)
			}
		}
	}
	return res, nil
}

func (r *PostgresNetworkingRepository) DeleteNetwork(id string) error {
	// Cascading Check: VPC must not have active subnets
	subnets, err := r.ListSubnets(id)
	if err == nil && len(subnets) > 0 {
		return appErrors.New(appErrors.CodeConflict, fmt.Sprintf("VPC_NOT_EMPTY: Cannot delete VPC '%s' while %d active subnets exist", id, len(subnets)))
	}

	net, err := r.GetNetwork(id)
	if err != nil {
		return err
	}
	net.Status = domain.StatusDeleted
	return r.SaveNetwork(net)
}

// 2. Subnets
func (r *PostgresNetworkingRepository) SaveSubnet(sn *domain.Subnet) error {
	if r.db != nil {
		var existing domain.Subnet
		err := r.db.Where("id = ?", sn.ID).First(&existing).Error
		if err == nil {
			sn.UpdatedAt = time.Now()
			return r.db.Save(sn).Error
		}
		if sn.CreatedAt.IsZero() {
			sn.CreatedAt = time.Now()
		}
		sn.UpdatedAt = time.Now()
		return r.db.Create(sn).Error
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.subnets[sn.ID] = sn
	return nil
}

func (r *PostgresNetworkingRepository) GetSubnet(id string) (*domain.Subnet, error) {
	if r.db != nil {
		var sn domain.Subnet
		if err := r.db.Where("id = ?", id).First(&sn).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("Subnet '%s' not found", id))
			}
			return nil, err
		}
		return &sn, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if sn, ok := r.subnets[id]; ok {
		return sn, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("Subnet '%s' not found", id))
}

func (r *PostgresNetworkingRepository) ListSubnets(networkID string) ([]*domain.Subnet, error) {
	if r.db != nil {
		var sns []*domain.Subnet
		query := r.db.Where("status != ?", "DELETED")
		if networkID != "" {
			query = query.Where("network_id = ?", networkID)
		}
		if err := query.Find(&sns).Error; err != nil {
			return nil, err
		}
		return sns, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.Subnet
	for _, sn := range r.subnets {
		if sn.Status != "DELETED" {
			if networkID == "" || sn.NetworkID == networkID {
				res = append(res, sn)
			}
		}
	}
	return res, nil
}

// 3. Security Groups
func (r *PostgresNetworkingRepository) SaveSecurityGroup(sg *domain.SecurityGroup) error {
	if len(sg.Rules) > 0 {
		b, _ := json.Marshal(sg.Rules)
		sg.RulesJSON = string(b)
	}

	if r.db != nil {
		var existing domain.SecurityGroup
		err := r.db.Where("id = ?", sg.ID).First(&existing).Error
		if err == nil {
			sg.UpdatedAt = time.Now()
			return r.db.Save(sg).Error
		}
		if sg.CreatedAt.IsZero() {
			sg.CreatedAt = time.Now()
		}
		sg.UpdatedAt = time.Now()
		return r.db.Create(sg).Error
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.sgs[sg.ID] = sg
	return nil
}

func (r *PostgresNetworkingRepository) GetSecurityGroup(id string) (*domain.SecurityGroup, error) {
	if r.db != nil {
		var sg domain.SecurityGroup
		if err := r.db.Where("id = ?", id).First(&sg).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("Security group '%s' not found", id))
			}
			return nil, err
		}
		if sg.RulesJSON != "" {
			_ = json.Unmarshal([]byte(sg.RulesJSON), &sg.Rules)
		}
		return &sg, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if sg, ok := r.sgs[id]; ok {
		return sg, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("Security group '%s' not found", id))
}

func (r *PostgresNetworkingRepository) ListSecurityGroups(networkID string) ([]*domain.SecurityGroup, error) {
	if r.db != nil {
		var sgs []*domain.SecurityGroup
		query := r.db.Where("status != ?", "DELETED")
		if networkID != "" {
			query = query.Where("network_id = ?", networkID)
		}
		if err := query.Find(&sgs).Error; err != nil {
			return nil, err
		}
		for _, sg := range sgs {
			if sg.RulesJSON != "" {
				_ = json.Unmarshal([]byte(sg.RulesJSON), &sg.Rules)
			}
		}
		return sgs, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.SecurityGroup
	for _, sg := range r.sgs {
		if sg.Status != "DELETED" {
			if networkID == "" || sg.NetworkID == networkID {
				res = append(res, sg)
			}
		}
	}
	return res, nil
}

// 4. Route Tables
func (r *PostgresNetworkingRepository) SaveRouteTable(rt *domain.RouteTable) error {
	if len(rt.Routes) > 0 {
		b, _ := json.Marshal(rt.Routes)
		rt.RoutesJSON = string(b)
	}

	if r.db != nil {
		var existing domain.RouteTable
		err := r.db.Where("id = ?", rt.ID).First(&existing).Error
		if err == nil {
			rt.UpdatedAt = time.Now()
			return r.db.Save(rt).Error
		}
		if rt.CreatedAt.IsZero() {
			rt.CreatedAt = time.Now()
		}
		rt.UpdatedAt = time.Now()
		return r.db.Create(rt).Error
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[rt.ID] = rt
	return nil
}

func (r *PostgresNetworkingRepository) GetRouteTable(id string) (*domain.RouteTable, error) {
	if r.db != nil {
		var rt domain.RouteTable
		if err := r.db.Where("id = ?", id).First(&rt).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("Route table '%s' not found", id))
			}
			return nil, err
		}
		if rt.RoutesJSON != "" {
			_ = json.Unmarshal([]byte(rt.RoutesJSON), &rt.Routes)
		}
		return &rt, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if rt, ok := r.routes[id]; ok {
		return rt, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("Route table '%s' not found", id))
}

func (r *PostgresNetworkingRepository) ListRouteTables(networkID string) ([]*domain.RouteTable, error) {
	if r.db != nil {
		var rts []*domain.RouteTable
		query := r.db.Where("status != ?", "DELETED")
		if networkID != "" {
			query = query.Where("network_id = ?", networkID)
		}
		if err := query.Find(&rts).Error; err != nil {
			return nil, err
		}
		for _, rt := range rts {
			if rt.RoutesJSON != "" {
				_ = json.Unmarshal([]byte(rt.RoutesJSON), &rt.Routes)
			}
		}
		return rts, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.RouteTable
	for _, rt := range r.routes {
		if rt.Status != "DELETED" {
			if networkID == "" || rt.NetworkID == networkID {
				res = append(res, rt)
			}
		}
	}
	return res, nil
}
