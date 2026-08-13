package repository

import (
	"fmt"
	"sync"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
)

type LoadBalancerRepository struct {
	mu        sync.RWMutex
	lbs       map[string]*domain.LoadBalancer
	listeners map[string]*domain.Listener
	pools     map[string]*domain.BackendPool
	targets   map[string]*domain.BackendTarget
	certs     map[string]*domain.Certificate
	domains   map[string]*domain.Domain
	apps      map[string]*domain.Application
}

func NewLoadBalancerRepository() *LoadBalancerRepository {
	return &LoadBalancerRepository{
		lbs:       make(map[string]*domain.LoadBalancer),
		listeners: make(map[string]*domain.Listener),
		pools:     make(map[string]*domain.BackendPool),
		targets:   make(map[string]*domain.BackendTarget),
		certs:     make(map[string]*domain.Certificate),
		domains:   make(map[string]*domain.Domain),
		apps:      make(map[string]*domain.Application),
	}
}

func (r *LoadBalancerRepository) SaveLoadBalancer(lb *domain.LoadBalancer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lbs[lb.ID] = lb
	return nil
}

func (r *LoadBalancerRepository) GetLoadBalancer(id string) (*domain.LoadBalancer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if lb, ok := r.lbs[id]; ok {
		return lb, nil
	}
	return nil, fmt.Errorf("load balancer '%s' not found", id)
}

func (r *LoadBalancerRepository) ListLoadBalancers(orgID, projectID string) ([]*domain.LoadBalancer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.LoadBalancer
	for _, lb := range r.lbs {
		if lb.Status != domain.LBStatusDeleted {
			res = append(res, lb)
		}
	}
	return res, nil
}

func (r *LoadBalancerRepository) SaveApplication(app *domain.Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.apps[app.ID] = app
	return nil
}

func (r *LoadBalancerRepository) GetApplication(id string) (*domain.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if app, ok := r.apps[id]; ok {
		return app, nil
	}
	return nil, fmt.Errorf("application '%s' not found", id)
}

func (r *LoadBalancerRepository) ListApplications(orgID, projectID string) ([]*domain.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.Application
	for _, app := range r.apps {
		if app.Status != domain.AppDeleted {
			res = append(res, app)
		}
	}
	return res, nil
}

func (r *LoadBalancerRepository) SaveCertificate(cert *domain.Certificate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.certs[cert.ID] = cert
	return nil
}

func (r *LoadBalancerRepository) SaveDomain(dom *domain.Domain) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.domains[dom.ID] = dom
	return nil
}
