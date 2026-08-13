package security

import (
	"fmt"
	"net"
	"strings"
)

type SSRFProtectionEngine struct{}

func NewSSRFProtectionEngine() *SSRFProtectionEngine {
	return &SSRFProtectionEngine{}
}

func (e *SSRFProtectionEngine) ValidateURL(targetURL string) error {
	clean := strings.ToLower(strings.TrimSpace(targetURL))

	// Cloud Metadata Endpoint Blocklist
	if strings.Contains(clean, "169.254.169.254") ||
		strings.Contains(clean, "169.254.169.253") ||
		strings.Contains(clean, "metadata.google.internal") ||
		strings.Contains(clean, "169.254.") {
		return fmt.Errorf("SSRF SECURITY RISK: Access to cloud metadata endpoint '%s' is strictly blocked by policy", targetURL)
	}

	// Loopback Blocklist
	if strings.Contains(clean, "127.0.0.1") || strings.Contains(clean, "localhost") || strings.Contains(clean, "::1") {
		return fmt.Errorf("SSRF SECURITY RISK: Access to local loopback address '%s' is strictly blocked by policy", targetURL)
	}

	// Parse IP directly if host URL
	host := clean
	if idx := strings.Index(clean, "://"); idx != -1 {
		host = clean[idx+3:]
	}
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("SSRF SECURITY RISK: Direct IP '%s' is in restricted metadata or link-local range", host)
		}
	}

	return nil
}
