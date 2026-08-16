package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

type SSRFProtectionEngine struct{}

func NewSSRFProtectionEngine() *SSRFProtectionEngine {
	return &SSRFProtectionEngine{}
}

func (e *SSRFProtectionEngine) ValidateURL(targetURL string) error {
	clean := strings.ToLower(strings.TrimSpace(targetURL))

	if clean == "" {
		return fmt.Errorf("SSRF SECURITY RISK: Target URL cannot be empty")
	}

	// Cloud Metadata Endpoint Blocklist
	if strings.Contains(clean, "169.254.169.254") ||
		strings.Contains(clean, "169.254.169.253") ||
		strings.Contains(clean, "metadata.google.internal") ||
		strings.Contains(clean, "169.254.") {
		return fmt.Errorf("SSRF SECURITY RISK: Access to cloud metadata endpoint '%s' is strictly blocked by policy", targetURL)
	}

	// Loopback & Internal Hostname Blocklist
	if strings.Contains(clean, "127.0.0.1") ||
		strings.Contains(clean, "localhost") ||
		strings.Contains(clean, "::1") ||
		strings.Contains(clean, ".internal") ||
		strings.Contains(clean, ".local") {
		return fmt.Errorf("SSRF SECURITY RISK: Access to local loopback or internal host '%s' is strictly blocked by policy", targetURL)
	}

	parsed, err := url.Parse(clean)
	if err == nil && parsed.Host != "" {
		host := parsed.Hostname()
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return fmt.Errorf("SSRF SECURITY RISK: Access to localhost '%s' is strictly blocked by policy", host)
		}

		ip := net.ParseIP(host)
		if ip != nil {
			if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
				return fmt.Errorf("SSRF SECURITY RISK: Direct IP '%s' is in restricted private or link-local range", host)
			}
		} else {
			ips, err := net.LookupIP(host)
			if err == nil {
				for _, resolvedIP := range ips {
					if resolvedIP.IsLoopback() || resolvedIP.IsLinkLocalUnicast() || resolvedIP.IsLinkLocalMulticast() || resolvedIP.IsPrivate() {
						return fmt.Errorf("SSRF SECURITY RISK: Resolved IP '%s' for host '%s' is in restricted private range", resolvedIP.String(), host)
					}
				}
			}
		}
	}

	return nil
}
