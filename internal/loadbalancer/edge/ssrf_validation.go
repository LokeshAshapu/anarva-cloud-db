package edge

import (
	"fmt"
	"net"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
)

type SSRFValidationService struct{}

func NewSSRFValidationService() *SSRFValidationService {
	return &SSRFValidationService{}
}

func (s *SSRFValidationService) ValidateOrigin(origin *domain.Origin) error {
	host := strings.TrimSpace(strings.ToLower(origin.HostnameReference))

	if host == "169.254.169.254" || host == "169.254.169.253" || strings.HasPrefix(host, "169.254.") {
		return fmt.Errorf("SSRF BLOCKED: Origin hostname '%s' points to cloud provider metadata endpoint", host)
	}

	if host == "127.0.0.1" || host == "localhost" || host == "::1" {
		return fmt.Errorf("SSRF BLOCKED: Origin hostname '%s' points to local loopback control plane", host)
	}

	ip := net.ParseIP(host)
	if ip != nil && (ip.IsLoopback() || ip.IsLinkLocalUnicast()) {
		return fmt.Errorf("SSRF BLOCKED: Origin IP '%s' is a restricted address", host)
	}

	return nil
}
