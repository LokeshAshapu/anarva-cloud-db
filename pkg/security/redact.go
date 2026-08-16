package security

import (
	"regexp"
	"strings"
)

var (
	// Matches anarva_live_... and anarva_test_... API keys
	apiKeyRegex = regexp.MustCompile(`(anarva_(live|test)_[a-zA-Z0-9_\-]+)`)
	// Matches whsec_live_... webhook secrets
	webhookSecretRegex = regexp.MustCompile(`(whsec_(live|test)_[a-zA-Z0-9_\-]+)`)
	// Matches Bearer tokens
	bearerRegex = regexp.MustCompile(`(?i)(Bearer\s+)[A-Za-z0-9\-\._~\+\/]+=*`)
	// Matches JWT tokens (header.payload.signature)
	jwtRegex = regexp.MustCompile(`ey[A-Za-z0-9_-]{10,}\.ey[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`)
	// Matches Database connection strings with passwords
	dsnPasswordRegex = regexp.MustCompile(`(postgres://|postgresql://|mysql://)([^:]+):([^@]+)@`)
	// Matches AWS Access Key ID and Secret Key patterns
	awsKeyRegex = regexp.MustCompile(`(AKIA[0-9A-Z]{16})`)
	// Matches JSON password / secret fields
	jsonPasswordRegex = regexp.MustCompile(`(?i)"(password|secret|token|access_token|refresh_token|api_key)"\s*:\s*"([^"]+)"`)
	// Matches plain text password occurrences in logs
	plainPasswordRegex = regexp.MustCompile(`(?i)(password[:= \t]+)([^\s"\}]+)`)
)

// RedactSecrets scans an arbitrary string and redacts passwords, API keys, tokens, and credentials.
func RedactSecrets(input string) string {
	if input == "" {
		return ""
	}

	result := input

	// Redact API Keys
	result = apiKeyRegex.ReplaceAllString(result, "[REDACTED_API_KEY]")

	// Redact Webhook Secrets
	result = webhookSecretRegex.ReplaceAllString(result, "[REDACTED_WEBHOOK_SECRET]")

	// Redact Bearer Headers
	result = bearerRegex.ReplaceAllString(result, "${1}[REDACTED]")

	// Redact JWT Tokens
	result = jwtRegex.ReplaceAllString(result, "[REDACTED_JWT_TOKEN]")

	// Redact DSN Passwords
	result = dsnPasswordRegex.ReplaceAllString(result, "${1}${2}:[REDACTED]@")

	// Redact AWS Keys
	result = awsKeyRegex.ReplaceAllString(result, "[REDACTED_AWS_KEY]")

	// Redact JSON Secret Fields
	result = jsonPasswordRegex.ReplaceAllString(result, `"${1}":"[REDACTED]"`)

	// Redact Plaintext Passwords
	result = plainPasswordRegex.ReplaceAllString(result, "${1}[REDACTED]")

	return result
}

// RedactAPIKey redacts an API key string, e.g. "anarva_live_abcdef1234567890" -> "[REDACTED_API_KEY]".
func RedactAPIKey(rawKey string) string {
	if rawKey == "" {
		return ""
	}
	if strings.HasPrefix(rawKey, "anarva_live_") || strings.HasPrefix(rawKey, "anarva_test_") {
		return "[REDACTED_API_KEY]"
	}
	return RedactSecrets(rawKey)
}
