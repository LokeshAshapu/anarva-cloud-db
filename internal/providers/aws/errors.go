package aws

import (
	"fmt"
	"regexp"
	"strings"
)

type AWSError struct {
	Code    string
	Message string
}

func (e *AWSError) Error() string {
	return fmt.Sprintf("AWS Error [%s]: %s", e.Code, RedactSecrets(e.Message))
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(anarva_live_[a-zA-Z0-9_]{10,})`),
	regexp.MustCompile(`(?i)(anarva_test_[a-zA-Z0-9_]{10,})`),
	regexp.MustCompile(`(?i)(aws_secret_access_key=[\w/+=\-]+)`),
	regexp.MustCompile(`(?i)(password=[\w/+=\-]+)`),
}

func RedactSecrets(msg string) string {
	redacted := msg
	for _, p := range secretPatterns {
		redacted = p.ReplaceAllString(redacted, "[REDACTED_SECRET]")
	}
	return redacted
}

func MapAWSError(err error) error {
	if err == nil {
		return nil
	}

	msg := RedactSecrets(err.Error())

	if strings.Contains(msg, "PROVIDER_AUTHENTICATION_FAILED") || strings.Contains(msg, "PROVIDER_INVALID_CONFIGURATION") || strings.Contains(msg, "PROVIDER_CAPABILITY_NOT_SUPPORTED") {
		return fmt.Errorf("%s", msg)
	}
	if strings.Contains(msg, "AccessDenied") || strings.Contains(msg, "UnauthorizedOperation") {
		return fmt.Errorf("PROVIDER_PERMISSION_DENIED: IAM policy prohibits requested operation (%s)", msg)
	}
	if strings.Contains(msg, "InvalidCredentials") || strings.Contains(msg, "UnrecognizedClientException") || strings.Contains(msg, "AUTH_FAILED") {
		return fmt.Errorf("PROVIDER_AUTHENTICATION_FAILED: Invalid AWS Access Key or Session Token (%s)", msg)
	}
	if strings.Contains(msg, "Throttling") || strings.Contains(msg, "RequestLimitExceeded") {
		return fmt.Errorf("PROVIDER_RATE_LIMITED: AWS API request throttled (%s)", msg)
	}
	if strings.Contains(msg, "ResourceLimitExceeded") || strings.Contains(msg, "QuotaExceededException") {
		return fmt.Errorf("PROVIDER_QUOTA_EXCEEDED: AWS account service limit reached (%s)", msg)
	}
	if strings.Contains(msg, "NotFound") || strings.Contains(msg, "InvalidInstanceID.NotFound") {
		return fmt.Errorf("PROVIDER_RESOURCE_NOT_FOUND: Target resource does not exist (%s)", msg)
	}
	if strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "timeout") {
		return fmt.Errorf("PROVIDER_TIMEOUT: Provider request timed out (%s)", msg)
	}

	return fmt.Errorf("PROVIDER_OPERATION_FAILED: %s", msg)
}
