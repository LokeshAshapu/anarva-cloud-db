package webhook_test

import (
	"testing"

	"github.com/anarva-cloud/anarva-cloud-db/internal/webhook/domain"
)

func TestWebhookEngine_SSRFProtectionAndHMAC(t *testing.T) {
	// Test SSRF Protection
	err := domain.ValidateWebhookURL("http://localhost:8080/api/v1")
	if err == nil {
		t.Errorf("Expected SSRF error for localhost, got nil")
	}

	err = domain.ValidateWebhookURL("http://127.0.0.1:8080/webhook")
	if err == nil {
		t.Errorf("Expected SSRF error for 127.0.0.1, got nil")
	}

	err = domain.ValidateWebhookURL("https://api.anarva.io/v1/webhooks/receive")
	if err != nil {
		t.Errorf("Expected valid URL for api.anarva.io, got error: %v", err)
	}

	// Test HMAC Signature Generation
	payload := []byte(`{"event":"resource.created","resourceId":"ace-worker-node-01"}`)
	secret := "whsec_live_9f82a1bc3d4e5f67"

	sig1 := domain.ComputeHMACSignature(payload, secret)
	sig2 := domain.ComputeHMACSignature(payload, secret)

	if sig1 == "" || sig1 != sig2 {
		t.Errorf("HMAC signatures do not match or are empty: %s vs %s", sig1, sig2)
	}
}
