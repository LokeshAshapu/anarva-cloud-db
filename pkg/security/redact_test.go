package security_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

func TestRedactSecrets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "API Key live",
			input:    "Failed request using key anarva_live_abc1234567890xyz",
			expected: "Failed request using key [REDACTED_API_KEY]",
		},
		{
			name:     "API Key test",
			input:    "Test key anarva_test_9876543210qwerty",
			expected: "Test key [REDACTED_API_KEY]",
		},
		{
			name:     "Webhook secret",
			input:    "Secret whsec_live_9f82a1bc3d4e5f67",
			expected: "Secret [REDACTED_WEBHOOK_SECRET]",
		},
		{
			name:     "Authorization Bearer",
			input:    "Authorization: Bearer my_secret_token_12345",
			expected: "Authorization: Bearer [REDACTED]",
		},
		{
			name:     "DSN password",
			input:    "Connecting to postgres://admin:super_secret_db_password@localhost:5432/anarva_db",
			expected: "Connecting to postgres://admin:[REDACTED]@localhost:5432/anarva_db",
		},
		{
			name:     "JSON password field",
			input:    `{"user":"admin","password":"mySecretPassword123"}`,
			expected: `{"user":"admin","password":"[REDACTED]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := security.RedactSecrets(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactAPIKey(t *testing.T) {
	assert.Equal(t, "[REDACTED_API_KEY]", security.RedactAPIKey("anarva_live_1234567890"))
	assert.Equal(t, "[REDACTED_API_KEY]", security.RedactAPIKey("anarva_test_0987654321"))
}
