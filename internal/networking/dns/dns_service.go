package dns

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type DNSProvider interface {
	CreateZone(ctx context.Context, zone *domain.DNSZone) (*domain.DNSZone, error)
	DeleteZone(ctx context.Context, id string) error
	CreateRecord(ctx context.Context, record *domain.DNSRecord) (*domain.DNSRecord, error)
	UpdateRecord(ctx context.Context, record *domain.DNSRecord) (*domain.DNSRecord, error)
	DeleteRecord(ctx context.Context, id string) error
	Resolve(ctx context.Context, hostname string) (string, error)
}

type LocalDNSProvider struct {
	mu      sync.RWMutex
	zones   map[string]*domain.DNSZone
	records map[string]*domain.DNSRecord
}

func NewLocalDNSProvider() *LocalDNSProvider {
	return &LocalDNSProvider{
		zones:   make(map[string]*domain.DNSZone),
		records: make(map[string]*domain.DNSRecord),
	}
}

func (p *LocalDNSProvider) CreateZone(ctx context.Context, zone *domain.DNSZone) (*domain.DNSZone, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	zone.Status = "ACTIVE"
	zone.Provider = "LOCAL_DNS"
	p.zones[zone.ID] = zone
	return zone, nil
}

func (p *LocalDNSProvider) DeleteZone(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.zones, id)
	return nil
}

func (p *LocalDNSProvider) CreateRecord(ctx context.Context, record *domain.DNSRecord) (*domain.DNSRecord, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	record.Status = "ACTIVE"
	p.records[record.ID] = record
	return record, nil
}

func (p *LocalDNSProvider) UpdateRecord(ctx context.Context, record *domain.DNSRecord) (*domain.DNSRecord, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.records[record.ID] = record
	return record, nil
}

func (p *LocalDNSProvider) DeleteRecord(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.records, id)
	return nil
}

func (p *LocalDNSProvider) Resolve(ctx context.Context, hostname string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, rec := range p.records {
		if strings.EqualFold(rec.Name, hostname) {
			return rec.Value, nil
		}
	}
	return "", fmt.Errorf("DNS resolution failed: hostname '%s' not found", hostname)
}

type DNSResolutionService struct {
	provider DNSProvider
}

func NewDNSResolutionService(provider DNSProvider) *DNSResolutionService {
	return &DNSResolutionService{provider: provider}
}

func (s *DNSResolutionService) Resolve(ctx context.Context, hostname string) (string, error) {
	return s.provider.Resolve(ctx, hostname)
}

func (s *DNSResolutionService) HealthCheck(ctx context.Context) (string, error) {
	return "HEALTHY (LOCAL_DNS_RESOLVER)", nil
}
