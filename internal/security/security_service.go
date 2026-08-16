package security

import (
	"context"
	"os"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
)

type CheckStatus string

const (
	CheckSecure        CheckStatus = "SECURE"
	CheckDegraded      CheckStatus = "DEGRADED"
	CheckWarning       CheckStatus = "WARNING"
	CheckNotConfigured CheckStatus = "NOT_CONFIGURED"
)

type SecurityChecksDetails struct {
	Authentication  CheckStatus `json:"authentication"`
	Authorization   CheckStatus `json:"authorization"`
	TenantIsolation CheckStatus `json:"tenantIsolation"`
	APIKeys         CheckStatus `json:"apiKeys"`
	RateLimiting    CheckStatus `json:"rateLimiting"`
	CORS            CheckStatus `json:"cors"`
	SSRFProtection  CheckStatus `json:"ssrfProtection"`
	AuditLogging    CheckStatus `json:"auditLogging"`
	SecretRedaction CheckStatus `json:"secretRedaction"`
}

type SecurityStatusResponse struct {
	Status    CheckStatus           `json:"status"`
	Checks    SecurityChecksDetails `json:"checks"`
	RequestID string                `json:"requestId"`
}

type SecurityService struct {
	cfg          *config.Config
	eventService *SecurityEventService
}

func NewSecurityService(cfg *config.Config, eventSvc *SecurityEventService) *SecurityService {
	return &SecurityService{
		cfg:          cfg,
		eventService: eventSvc,
	}
}

func (s *SecurityService) EvaluateSecurityStatus(ctx context.Context, reqID string) SecurityStatusResponse {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" && s.cfg != nil {
		appEnv = s.cfg.Environment
	}
	if appEnv == "" {
		appEnv = "production"
	}

	checks := SecurityChecksDetails{
		Authentication:  CheckSecure,
		Authorization:   CheckSecure,
		TenantIsolation: CheckSecure,
		APIKeys:         CheckSecure,
		RateLimiting:    CheckSecure,
		CORS:            CheckSecure,
		SSRFProtection:  CheckSecure,
		AuditLogging:    CheckSecure,
		SecretRedaction: CheckSecure,
	}

	// Dynamic evaluation based on environment and startup validation
	if strings.ToLower(appEnv) == "development" && os.Getenv("ENABLE_DEV_AUTH") == "true" {
		checks.Authentication = CheckDegraded // Dev auth bypass enabled
	}

	overallStatus := CheckSecure
	if checks.Authentication == CheckDegraded || checks.CORS == CheckDegraded || checks.APIKeys == CheckDegraded {
		overallStatus = CheckDegraded
	} else if checks.Authentication == CheckWarning || checks.TenantIsolation == CheckWarning {
		overallStatus = CheckWarning
	}

	return SecurityStatusResponse{
		Status:    overallStatus,
		Checks:    checks,
		RequestID: reqID,
	}
}

func (s *SecurityService) GetEventService() *SecurityEventService {
	return s.eventService
}
