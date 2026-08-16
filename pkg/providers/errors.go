package providers

import (
	"errors"
	"fmt"
)

var (
	ErrProviderNotFound           = errors.New("provider not found")
	ErrProviderUnavailable        = errors.New("provider unavailable")
	ErrUnsupportedCapability      = errors.New("unsupported provider capability")
	ErrResourceNotFound           = errors.New("resource not found")
	ErrResourceAlreadyExists      = errors.New("resource already exists")
	ErrProviderExecutionFailed    = errors.New("provider execution failed")
	ErrProviderTimeout            = errors.New("provider execution timeout")
	ErrProviderAuthenticationFailed = errors.New("provider authentication failed")
)

// ProviderError wraps a provider error with standard ANARVA domain code.
type ProviderError struct {
	ProviderID string
	Code       string
	Message    string
	Err        error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] provider '%s' error: %s (%v)", e.Code, e.ProviderID, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] provider '%s' error: %s", e.Code, e.ProviderID, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

func NewProviderError(providerID, code, message string, err error) *ProviderError {
	return &ProviderError{
		ProviderID: providerID,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}
