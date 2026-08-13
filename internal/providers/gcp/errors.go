package gcp

import (
	"fmt"
	"strings"
)

func MapGCPError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	if strings.Contains(msg, "PERMISSION_DENIED") || strings.Contains(msg, "Forbidden") {
		return fmt.Errorf("PROVIDER_PERMISSION_DENIED: GCP IAM role lacks required permissions (%v)", err)
	}
	if strings.Contains(msg, "UNAUTHENTICATED") || strings.Contains(msg, "invalid_grant") {
		return fmt.Errorf("PROVIDER_AUTH_FAILED: Invalid GCP Service Account JSON key (%v)", err)
	}
	if strings.Contains(msg, "RESOURCE_EXHAUSTED") || strings.Contains(msg, "QuotaExceeded") {
		return fmt.Errorf("PROVIDER_QUOTA_EXCEEDED: GCP project quota limit reached (%v)", err)
	}
	if strings.Contains(msg, "NOT_FOUND") {
		return fmt.Errorf("RESOURCE_NOT_FOUND: Target GCP resource does not exist (%v)", err)
	}

	return err
}
