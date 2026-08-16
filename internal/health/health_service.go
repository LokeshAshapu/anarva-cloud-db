package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/database"
)

type ComponentStatus string

const (
	StatusReady         ComponentStatus = "READY"
	StatusDegraded      ComponentStatus = "DEGRADED"
	StatusUnavailable   ComponentStatus = "UNAVAILABLE"
	StatusNotConfigured ComponentStatus = "NOT_CONFIGURED"
)

type HealthCheckDetails struct {
	Database      ComponentStatus `json:"database"`
	Configuration ComponentStatus `json:"configuration"`
	Providers     ComponentStatus `json:"providers"`
	Operations    ComponentStatus `json:"operations"`
}

type HealthResponse struct {
	Status    string             `json:"status"`
	Service   string             `json:"service"`
	Version   string             `json:"version"`
	Checks    HealthCheckDetails `json:"checks"`
	RequestID string             `json:"requestId"`
}

type SystemComponentStatus struct {
	Name        string          `json:"name"`
	Key         string          `json:"key"`
	Status      ComponentStatus `json:"status"`
	Description string          `json:"description"`
}

type SystemStatusResponse struct {
	Status     string                  `json:"status"`
	Components []SystemComponentStatus `json:"components"`
	RequestID  string                  `json:"requestId"`
}

type HealthService struct {
	mu            sync.RWMutex
	dbPool        *database.DB
	cfg           *config.Config
	providerReg   *registry.ProviderRegistry
	reliabilityUC *usecase.ReliabilityUseCase
	version       string
}

func NewHealthService(
	dbPool *database.DB,
	cfg *config.Config,
	providerReg *registry.ProviderRegistry,
	reliabilityUC *usecase.ReliabilityUseCase,
	version string,
) *HealthService {
	if version == "" {
		version = "0.1.0"
	}
	return &HealthService{
		dbPool:        dbPool,
		cfg:           cfg,
		providerReg:   providerReg,
		reliabilityUC: reliabilityUC,
		version:       version,
	}
}

func (s *HealthService) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := getRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	resp := map[string]interface{}{
		"status":    "UP",
		"service":   "anarva-control-plane",
		"version":   s.version,
		"requestId": reqID,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *HealthService) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := getRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	checks := s.CheckReadiness(r.Context())

	isReady := checks.Database != StatusUnavailable && checks.Configuration != StatusUnavailable && checks.Operations != StatusUnavailable

	overallStatus := "READY"
	statusCode := http.StatusOK
	if !isReady {
		overallStatus = "NOT_READY"
		statusCode = http.StatusServiceUnavailable
	} else if checks.Database == StatusDegraded || checks.Providers == StatusDegraded || checks.Operations == StatusDegraded {
		overallStatus = "DEGRADED"
	}

	resp := HealthResponse{
		Status:    overallStatus,
		Service:   "anarva-control-plane",
		Version:   s.version,
		Checks:    checks,
		RequestID: reqID,
	}

	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *HealthService) CheckReadiness(ctx context.Context) HealthCheckDetails {
	s.mu.RLock()
	defer s.mu.RUnlock()

	checks := HealthCheckDetails{
		Database:      StatusReady,
		Configuration: StatusReady,
		Providers:     StatusReady,
		Operations:    StatusReady,
	}

	// 1. Database Check
	if s.dbPool == nil {
		if s.cfg != nil && s.cfg.Environment == "production" {
			checks.Database = StatusUnavailable
		} else {
			checks.Database = StatusDegraded // Dev in-memory mode
		}
	} else {
		timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err := s.dbPool.HealthCheck(timeoutCtx); err != nil {
			checks.Database = StatusUnavailable
		}
	}

	// 2. Configuration Check
	if s.cfg == nil {
		checks.Configuration = StatusUnavailable
	} else if err := config.ValidateProductionConfig(s.cfg); err != nil {
		checks.Configuration = StatusUnavailable
	}

	// 3. Provider Check
	if s.providerReg != nil {
		providers, err := s.providerReg.ListProviders(ctx)
		if err == nil {
			hasConnected := false
			for _, p := range providers {
				if string(p.Status) == string(registry.StatusConnected) {
					hasConnected = true
					break
				}
			}
			if !hasConnected {
				checks.Providers = StatusDegraded
			}
		}
	} else {
		checks.Providers = StatusNotConfigured
	}

	// 4. Operations Engine Check
	if s.reliabilityUC == nil {
		checks.Operations = StatusUnavailable
	}

	return checks
}

func (s *HealthService) GetSystemStatus(ctx context.Context, reqID string) SystemStatusResponse {
	checks := s.CheckReadiness(ctx)

	overall := "READY"
	if checks.Database == StatusUnavailable || checks.Configuration == StatusUnavailable || checks.Operations == StatusUnavailable {
		overall = "UNAVAILABLE"
	} else if checks.Database == StatusDegraded || checks.Providers == StatusDegraded || checks.Operations == StatusDegraded {
		overall = "DEGRADED"
	}

	components := []SystemComponentStatus{
		{Name: "Control Plane", Key: "controlPlane", Status: StatusReady, Description: "Anarva Core Gateway & API Engine"},
		{Name: "Database", Key: "database", Status: checks.Database, Description: "Control Plane PostgreSQL Persistence"},
		{Name: "Operations Engine", Key: "operationsEngine", Status: checks.Operations, Description: "Operation Dispatch & State Machine"},
		{Name: "Reliability Engine", Key: "reliabilityEngine", Status: checks.Operations, Description: "Distributed Leases, Idempotency & Recovery"},
		{Name: "Provider Registry", Key: "providerRegistry", Status: checks.Providers, Description: "Infrastructure Provider Integration Matrix"},
		{Name: "Observability", Key: "observability", Status: StatusReady, Description: "Metrics, Traces & Audit Pipeline"},
		{Name: "Billing", Key: "billing", Status: StatusReady, Description: "Usage Metering & Tenant Quota Service"},
		{Name: "Feedback", Key: "feedback", Status: StatusReady, Description: "Platform Diagnostics & Developer Telemetry"},
	}

	return SystemStatusResponse{
		Status:     overall,
		Components: components,
		RequestID:  reqID,
	}
}

func getRequestID(r *http.Request) string {
	if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
		return reqID
	}
	if ctxReqID, ok := r.Context().Value("requestID").(string); ok && ctxReqID != "" {
		return ctxReqID
	}
	return "req-" + time.Now().Format("20060102150405")
}
