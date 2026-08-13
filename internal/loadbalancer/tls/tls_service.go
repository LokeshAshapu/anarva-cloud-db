package tls

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
)

type TLSService struct{}

func NewTLSService() *TLSService {
	return &TLSService{}
}

func (s *TLSService) RequestCertificate(orgID, projectID, domainName string) *domain.Certificate {
	return &domain.Certificate{
		ID:                   fmt.Sprintf("cert-%d", time.Now().UnixNano()),
		OrganizationID:       orgID,
		ProjectID:            projectID,
		Domain:               domainName,
		Provider:             "LOCAL_TLS_ISSUER",
		Status:               domain.CertActive,
		CertificateReference: fmt.Sprintf("arnv:cert:ap-hyderabad-1:%s:cert/%s", projectID, domainName),
		Issuer:               "Anarva Cloud Local CA",
		ExpiresAt:            time.Now().AddDate(1, 0, 0),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func (s *TLSService) ValidateHTTPSRedirect(listener *domain.Listener, cert *domain.Certificate) error {
	if listener.Protocol == domain.ProtocolHTTPS && !listener.TLSEnabled {
		return fmt.Errorf("HTTPS listener must have TLS enabled")
	}
	if listener.Protocol == domain.ProtocolHTTPS && (cert == nil || cert.Status != domain.CertActive) {
		return fmt.Errorf("cannot activate HTTPS listener without an ACTIVE TLS certificate")
	}
	return nil
}
