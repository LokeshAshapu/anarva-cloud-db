package security_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

func TestSecretLeakagePrevention(t *testing.T) {
	rawApiKey := "anarva_live_ak_9876543210qwerty"
	rawWebhookSecret := "whsec_live_9f82a1bc3d4e5f67"
	rawPassword := "SuperSecretDBPass123!"
	rawJwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNyLTEwMSJ9.somesignaturehash12345"

	tests := []struct {
		category string
		rawError string
	}{
		{
			category: "1. Normal Request Output",
			rawError: fmt.Sprintf("Processed request for user usr-101 using key %s", rawApiKey),
		},
		{
			category: "2. Failed Request Output",
			rawError: fmt.Sprintf("HTTP 500 error processing key %s on webhook %s", rawApiKey, rawWebhookSecret),
		},
		{
			category: "3. Authentication Failure",
			rawError: fmt.Sprintf("Authentication failed for Authorization: Bearer %s with password %s", rawJwt, rawPassword),
		},
		{
			category: "4. Provider Failure",
			rawError: fmt.Sprintf("AWS provider error with AKIAIOSFODNN7EXAMPLE using key %s", rawApiKey),
		},
		{
			category: "5. Database Failure",
			rawError: fmt.Sprintf("Failed to connect to postgres://admin:%s@localhost:5432/anarva_db", rawPassword),
		},
		{
			category: "6. CLI --debug Output",
			rawError: fmt.Sprintf("[DEBUG] Executing anarva compute list --api-key %s --token %s", rawApiKey, rawJwt),
		},
		{
			category: "7. SDK Error Output",
			rawError: fmt.Sprintf("AnarvaSDKError: request failed with Authorization: Bearer %s", rawJwt),
		},
		{
			category: "8. Terraform Diagnostic Output",
			rawError: fmt.Sprintf("Terraform provider error: failed to authenticate with secret %s and key %s", rawWebhookSecret, rawApiKey),
		},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			redacted := security.RedactSecrets(tt.rawError)

			// Assert no raw secret is exposed
			assert.NotContains(t, redacted, rawApiKey)
			assert.NotContains(t, redacted, rawWebhookSecret)
			assert.NotContains(t, redacted, rawPassword)
			assert.NotContains(t, redacted, rawJwt)

			// Assert safe redactions are present
			assert.True(t,
				assert.ObjectsAreEqual(redacted != tt.rawError, true) ||
					assert.ObjectsAreEqual(redacted, tt.rawError) == false,
			)
		})
	}
}
