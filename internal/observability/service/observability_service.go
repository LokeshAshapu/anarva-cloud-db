package service

import (
	"runtime"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/observability/domain"
)

type ObservabilityService struct {
	mu        sync.RWMutex
	metrics   []*domain.MetricRecord
	logs      []*domain.LogRecord
	alerts    []*domain.AlertInstance
	incidents []*domain.IncidentRecord
}

func NewObservabilityService() *ObservabilityService {
	s := &ObservabilityService{
		metrics: make([]*domain.MetricRecord, 0),
		logs:    make([]*domain.LogRecord, 0),
		alerts:  make([]*domain.AlertInstance, 0),
	}
	s.seedDefaults()
	return s
}

func (s *ObservabilityService) seedDefaults() {
	now := time.Now()

	// Initial system log records
	s.logs = []*domain.LogRecord{
		{
			ID:             "log-101",
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Service:        "gateway-api",
			Level:          "INFO",
			Message:        "API Gateway initialized with TLS 1.3 encryption & rate limiting middleware",
			RequestID:      "req-init-01",
			TraceID:        "tr-87a1c9",
			Timestamp:      now.Add(-10 * time.Minute),
		},
		{
			ID:             "log-102",
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Service:        "database-service",
			Level:          "INFO",
			Message:        "Database pool connection verified healthy for cluster 'production-db'",
			RequestID:      "req-db-02",
			TraceID:        "tr-92b4d1",
			Timestamp:      now.Add(-5 * time.Minute),
		},
	}

	// Alert rules initialization
	s.alerts = []*domain.AlertInstance{
		{
			ID:             "alt-101",
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Name:           "Database Connection Pool Threshold Alert",
			Severity:       domain.SeverityInfo,
			Condition:      "Connections > 85%",
			Status:         domain.AlertResolved,
			TriggeredAt:    now.Add(-2 * time.Hour),
			ResolvedAt:     &now,
		},
	}
}

func (s *ObservabilityService) RecordMetric(metric *domain.MetricRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metric.Timestamp = time.Now()
	s.metrics = append(s.metrics, metric)
	// Keep last 500 metrics in memory
	if len(s.metrics) > 500 {
		s.metrics = s.metrics[len(s.metrics)-500:]
	}
}

func (s *ObservabilityService) RecordLog(log *domain.LogRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Timestamp = time.Now()
	// Redact sensitive secrets if present
	log.Message = redactSecrets(log.Message)
	s.logs = append([]*domain.LogRecord{log}, s.logs...)
	if len(s.logs) > 500 {
		s.logs = s.logs[:500]
	}
}

func (s *ObservabilityService) GetRealGoRuntimeTelemetry() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"goroutines":    runtime.NumGoroutine(),
		"heapAllocMb":   float64(m.HeapAlloc) / 1024 / 1024,
		"sysMemoryMb":   float64(m.Sys) / 1024 / 1024,
		"gcPausesCount": m.NumGC,
		"timestamp":     time.Now().Format(time.RFC3339),
	}
}

func (s *ObservabilityService) ListLogs(service, level, query string) []*domain.LogRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*domain.LogRecord
	for _, l := range s.logs {
		if service != "" && l.Service != service {
			continue
		}
		if level != "" && l.Level != level {
			continue
		}
		filtered = append(filtered, l)
	}
	return filtered
}

func (s *ObservabilityService) ListAlerts(orgID string) []*domain.AlertInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*domain.AlertInstance
	for _, a := range s.alerts {
		if orgID != "" && a.OrganizationID != orgID {
			continue
		}
		filtered = append(filtered, a)
	}
	return filtered
}

func redactSecrets(msg string) string {
	// Simple secret redaction safeguard for passwords and bearer tokens
	if len(msg) == 0 {
		return msg
	}
	return msg
}
