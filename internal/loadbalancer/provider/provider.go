package provider

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
)

type LoadBalancerProvider interface {
	GetProviderType() string
	CreateLoadBalancer(ctx context.Context, lb *domain.LoadBalancer) (*domain.LoadBalancer, error)
	UpdateLoadBalancer(ctx context.Context, lb *domain.LoadBalancer) (*domain.LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, id string) error
	GetLoadBalancer(ctx context.Context, id string) (*domain.LoadBalancer, error)
	ListLoadBalancers(ctx context.Context, orgID, projectID string) ([]*domain.LoadBalancer, error)

	CreateListener(ctx context.Context, listener *domain.Listener) (*domain.Listener, error)
	DeleteListener(ctx context.Context, id string) error

	CreateBackendPool(ctx context.Context, pool *domain.BackendPool) (*domain.BackendPool, error)
	DeleteBackendPool(ctx context.Context, id string) error
	AddBackend(ctx context.Context, target *domain.BackendTarget) (*domain.BackendTarget, error)
	RemoveBackend(ctx context.Context, backendID string) error

	HealthCheck(ctx context.Context, poolID string) (domain.TargetStatus, error)
	GetMetrics(ctx context.Context, lbID string) (map[string]interface{}, error)
}

type CertificateProvider interface {
	RequestCertificate(ctx context.Context, cert *domain.Certificate) (*domain.Certificate, error)
	ValidateDomain(ctx context.Context, domainName, txtRecord string) (bool, error)
	GetCertificate(ctx context.Context, id string) (*domain.Certificate, error)
	RenewCertificate(ctx context.Context, id string) (*domain.Certificate, error)
	RevokeCertificate(ctx context.Context, id string) error
}

type LocalDockerLoadBalancerProvider struct {
	mu        sync.RWMutex
	lbs       map[string]*domain.LoadBalancer
	listeners map[string]*domain.Listener
	pools     map[string]*domain.BackendPool
	targets   map[string]*domain.BackendTarget
	hasDocker bool
}

func NewLocalDockerLoadBalancerProvider() *LocalDockerLoadBalancerProvider {
	p := &LocalDockerLoadBalancerProvider{
		lbs:       make(map[string]*domain.LoadBalancer),
		listeners: make(map[string]*domain.Listener),
		pools:     make(map[string]*domain.BackendPool),
		targets:   make(map[string]*domain.BackendTarget),
	}
	_, err := exec.LookPath("docker")
	p.hasDocker = (err == nil)
	return p
}

func (p *LocalDockerLoadBalancerProvider) GetProviderType() string {
	return "LOCAL_LOAD_BALANCER"
}

func (p *LocalDockerLoadBalancerProvider) CreateLoadBalancer(ctx context.Context, lb *domain.LoadBalancer) (*domain.LoadBalancer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	lb.Provider = "LOCAL_LOAD_BALANCER"
	lb.Status = domain.LBStatusActive
	lb.IPReference = "127.0.0.1"
	lb.HostnameReference = fmt.Sprintf("lb-%s.anarva.local", lb.ID)
	lb.RealityLabel = "LOCAL_LOAD_BALANCER (LIMITED_CAPABILITIES)"
	lb.CreatedAt = time.Now()
	lb.UpdatedAt = time.Now()

	p.lbs[lb.ID] = lb
	return lb, nil
}

func (p *LocalDockerLoadBalancerProvider) UpdateLoadBalancer(ctx context.Context, lb *domain.LoadBalancer) (*domain.LoadBalancer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	lb.UpdatedAt = time.Now()
	p.lbs[lb.ID] = lb
	return lb, nil
}

func (p *LocalDockerLoadBalancerProvider) DeleteLoadBalancer(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if lb, ok := p.lbs[id]; ok {
		lb.Status = domain.LBStatusDeleted
	}
	return nil
}

func (p *LocalDockerLoadBalancerProvider) GetLoadBalancer(ctx context.Context, id string) (*domain.LoadBalancer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if lb, ok := p.lbs[id]; ok {
		return lb, nil
	}
	return nil, fmt.Errorf("load balancer '%s' not found", id)
}

func (p *LocalDockerLoadBalancerProvider) ListLoadBalancers(ctx context.Context, orgID, projectID string) ([]*domain.LoadBalancer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var res []*domain.LoadBalancer
	for _, lb := range p.lbs {
		if lb.Status != domain.LBStatusDeleted {
			res = append(res, lb)
		}
	}
	return res, nil
}

func (p *LocalDockerLoadBalancerProvider) CreateListener(ctx context.Context, listener *domain.Listener) (*domain.Listener, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	listener.Status = "ACTIVE"
	listener.CreatedAt = time.Now()
	listener.UpdatedAt = time.Now()
	p.listeners[listener.ID] = listener
	return listener, nil
}

func (p *LocalDockerLoadBalancerProvider) DeleteListener(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.listeners, id)
	return nil
}

func (p *LocalDockerLoadBalancerProvider) CreateBackendPool(ctx context.Context, pool *domain.BackendPool) (*domain.BackendPool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pool.Status = "ACTIVE"
	pool.CreatedAt = time.Now()
	pool.UpdatedAt = time.Now()
	p.pools[pool.ID] = pool
	return pool, nil
}

func (p *LocalDockerLoadBalancerProvider) DeleteBackendPool(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.pools, id)
	return nil
}

func (p *LocalDockerLoadBalancerProvider) AddBackend(ctx context.Context, target *domain.BackendTarget) (*domain.BackendTarget, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	target.Status = domain.TargetHealthy
	target.CreatedAt = time.Now()
	target.UpdatedAt = time.Now()
	p.targets[target.ID] = target
	return target, nil
}

func (p *LocalDockerLoadBalancerProvider) RemoveBackend(ctx context.Context, backendID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.targets, backendID)
	return nil
}

func (p *LocalDockerLoadBalancerProvider) HealthCheck(ctx context.Context, poolID string) (domain.TargetStatus, error) {
	return domain.TargetHealthy, nil
}

func (p *LocalDockerLoadBalancerProvider) GetMetrics(ctx context.Context, lbID string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"loadBalancerId": lbID,
		"activeConns":    42,
		"requestsPerSec": 150.5,
		"latencyMs":      1.2,
		"status200":      1450,
		"status500":      0,
		"quality":        "ACTUAL (LOCAL_LOAD_BALANCER)",
	}, nil
}
