package aws

import (
	"context"
	"fmt"
	"strings"
)

type EnvironmentMode string

const (
	EnvDevelopment EnvironmentMode = "DEVELOPMENT"
	EnvStaging     EnvironmentMode = "STAGING"
	EnvProduction  EnvironmentMode = "PRODUCTION"
)

type ExecutionMode string

const (
	ModeLocal ExecutionMode = "LOCAL"
	ModeReal  ExecutionMode = "REAL"
)

type ProviderModeValidator struct{}

func NewProviderModeValidator() *ProviderModeValidator {
	return &ProviderModeValidator{}
}

func (v *ProviderModeValidator) ValidateExecutionMode(ctx context.Context, env string, mode string, accessKey string, secretKey string, isMock bool) error {
	normalizedEnv := strings.ToUpper(strings.TrimSpace(env))
	normalizedMode := strings.ToUpper(strings.TrimSpace(mode))

	if normalizedEnv == "PRODUCTION" {
		if isMock {
			return fmt.Errorf("PROVIDER_INVALID_CONFIGURATION: Production environment MUST NOT execute mock infrastructure clients")
		}
		if normalizedMode == "REAL" {
			if accessKey == "" || secretKey == "" {
				return fmt.Errorf("PROVIDER_AUTHENTICATION_FAILED: Production real provider mode requires active AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY credentials")
			}
		}
	}

	if normalizedMode == "REAL" && accessKey == "" && normalizedEnv != "DEVELOPMENT" {
		return fmt.Errorf("PROVIDER_AUTHENTICATION_FAILED: Real provider mode requires valid AWS credentials")
	}

	return nil
}
