package aws

import (
	"fmt"
	"strings"
)

type AWSError struct {
	Code    string
	Message string
}

func (e *AWSError) Error() string {
	return fmt.Sprintf("AWS Error [%s]: %s", e.Code, e.Message)
}

func MapAWSError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	if strings.Contains(msg, "AccessDenied") || strings.Contains(msg, "UnauthorizedOperation") {
		return fmt.Errorf("PROVIDER_PERMISSION_DENIED: IAM policy prohibits requested operation (%v)", err)
	}
	if strings.Contains(msg, "InvalidCredentials") || strings.Contains(msg, "UnrecognizedClientException") {
		return fmt.Errorf("PROVIDER_AUTH_FAILED: Invalid AWS Access Key or Session Token (%v)", err)
	}
	if strings.Contains(msg, "Throttling") || strings.Contains(msg, "RequestLimitExceeded") {
		return fmt.Errorf("PROVIDER_RATE_LIMITED: AWS API request throttled (%v)", err)
	}
	if strings.Contains(msg, "ResourceLimitExceeded") || strings.Contains(msg, "QuotaExceededException") {
		return fmt.Errorf("PROVIDER_QUOTA_EXCEEDED: AWS account service limit reached (%v)", err)
	}
	if strings.Contains(msg, "NotFound") {
		return fmt.Errorf("RESOURCE_NOT_FOUND: Target AWS resource does not exist (%v)", err)
	}

	return err
}
