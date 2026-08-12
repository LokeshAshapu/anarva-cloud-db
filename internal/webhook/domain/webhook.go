package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type WebhookEndpoint struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	URL            string    `json:"url"`
	Description    string    `json:"description"`
	Status         string    `json:"status"` // ACTIVE, DISABLED
	SecretPrefix   string    `json:"secretPrefix"`
	SecretHash     string    `json:"-"`
	Events         []string  `json:"events"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type WebhookDelivery struct {
	ID           string    `json:"id"`
	EndpointID   string    `json:"endpointId"`
	EventID      string    `json:"eventId"`
	EventType    string    `json:"eventType"`
	Status       string    `json:"status"` // SUCCESS, FAILED, RETRYING
	Attempts     int       `json:"attempts"`
	ResponseCode int       `json:"responseCode"`
	DeliveredAt  time.Time `json:"deliveredAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

type WebhookEvent struct {
	ID        string            `json:"id"`
	EventType string            `json:"eventType"`
	Timestamp time.Time         `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// ValidateWebhookURL protects against SSRF by disallowing private/internal IP ranges
func ValidateWebhookURL(targetURL string) error {
	parsed, err := url.Parse(targetURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid webhook URL protocol: must be http or https")
	}

	hostname := parsed.Hostname()
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return fmt.Errorf("webhooks to localhost or 127.0.0.1 are forbidden (SSRF protection)")
	}

	ips, err := net.LookupIP(hostname)
	if err == nil {
		for _, ip := range ips {
			if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
				return fmt.Errorf("webhooks targeting private/internal network addresses (%s) are forbidden (SSRF protection)", ip.String())
			}
		}
	}

	if strings.Contains(targetURL, "169.254.169.254") || strings.Contains(targetURL, "metadata.google.internal") {
		return fmt.Errorf("webhooks targeting cloud metadata endpoints are forbidden (SSRF protection)")
	}

	return nil
}

// ComputeHMACSignature generates HMAC-SHA256 signature for webhook payload verification
func ComputeHMACSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
