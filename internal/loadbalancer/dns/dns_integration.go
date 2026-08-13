package dns

import (
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
)

type DNSIntegrationService struct{}

func NewDNSIntegrationService() *DNSIntegrationService {
	return &DNSIntegrationService{}
}

func (s *DNSIntegrationService) BindDomainToLoadBalancer(dom *domain.Domain, lb *domain.LoadBalancer) (string, error) {
	if dom.VerificationStatus != domain.DomainVerified {
		return "", fmt.Errorf("cannot route traffic for unverified domain '%s'", dom.Name)
	}

	recordType := "A"
	targetValue := lb.IPReference
	if lb.HostnameReference != "" {
		recordType = "CNAME"
		targetValue = lb.HostnameReference
	}

	return fmt.Sprintf("DNS Record Created: %s %s -> %s", dom.Name, recordType, targetValue), nil
}
