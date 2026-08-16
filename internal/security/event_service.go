package security

import (
	"fmt"
	"sync"
	"time"

	pkgSecurity "github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

type EventType string

const (
	EventAuthFailure             EventType = "AUTH_FAILURE"
	EventAuthzDenial             EventType = "AUTHORIZATION_DENIAL"
	EventInvalidAPIKey           EventType = "INVALID_API_KEY"
	EventRevokedAPIKeyUsage      EventType = "REVOKED_API_KEY_USAGE"
	EventRateLimitViolation      EventType = "RATE_LIMIT_VIOLATION"
	EventTenantIsolationBreach   EventType = "TENANT_ISOLATION_VIOLATION"
	EventSSRFRejection           EventType = "SSRF_REJECTION"
	EventInvalidSignature        EventType = "INVALID_SIGNATURE"
	EventSecurityConfigFailure   EventType = "SECURITY_CONFIG_FAILURE"
	EventStorageTraversalBlocked EventType = "STORAGE_TRAVERSAL_BLOCKED"
)

type EventSeverity string

const (
	SeverityLow      EventSeverity = "LOW"
	SeverityMedium   EventSeverity = "MEDIUM"
	SeverityHigh     EventSeverity = "HIGH"
	SeverityCritical EventSeverity = "CRITICAL"
)

type SecurityEvent struct {
	ID             string        `json:"id"`
	Timestamp      time.Time     `json:"timestamp"`
	EventType      EventType     `json:"event"`
	Severity       EventSeverity `json:"severity"`
	Result         string        `json:"result"` // REJECTED, BLOCKED, DENIED
	ActorID        string        `json:"actor"`
	OrganizationID string        `json:"organizationId,omitempty"`
	ProjectID      string        `json:"projectId,omitempty"`
	ResourceID     string        `json:"resourceId,omitempty"`
	RequestID      string        `json:"requestId"`
	Details        string        `json:"details"`
}

type SecurityEventService struct {
	mu     sync.RWMutex
	events []*SecurityEvent
}

func NewSecurityEventService() *SecurityEventService {
	svc := &SecurityEventService{
		events: make([]*SecurityEvent, 0),
	}
	svc.seedSampleEvents()
	return svc
}

func (s *SecurityEventService) RecordEvent(eventType EventType, severity EventSeverity, result, actorID, orgID, projID, resID, reqID, details string) *SecurityEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	redactedDetails := pkgSecurity.RedactSecrets(details)

	now := time.Now()
	evt := &SecurityEvent{
		ID:             fmt.Sprintf("secev-%d", now.UnixNano()/1e6),
		Timestamp:      now,
		EventType:      eventType,
		Severity:       severity,
		Result:         result,
		ActorID:        actorID,
		OrganizationID: orgID,
		ProjectID:      projID,
		ResourceID:     resID,
		RequestID:      reqID,
		Details:        redactedDetails,
	}

	// Maintain sliding window of last 200 security events
	s.events = append([]*SecurityEvent{evt}, s.events...)
	if len(s.events) > 200 {
		s.events = s.events[:200]
	}

	return evt
}

func (s *SecurityEventService) ListEvents(orgID string) []*SecurityEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*SecurityEvent
	for _, e := range s.events {
		if orgID != "" && e.OrganizationID != "" && e.OrganizationID != orgID {
			continue
		}
		result = append(result, e)
	}
	return result
}

func (s *SecurityEventService) seedSampleEvents() {
	now := time.Now()
	s.events = append(s.events, &SecurityEvent{
		ID:             "secev-101",
		Timestamp:      now.Add(-15 * time.Minute),
		EventType:      EventSSRFRejection,
		Severity:       SeverityHigh,
		Result:         "BLOCKED",
		ActorID:        "usr-admin",
		OrganizationID: "org-default",
		RequestID:      "req_ssrf_block_01",
		Details:        "SSRF Protection Engine blocked outbound request targeting restricted metadata IP 169.254.169.254",
	}, &SecurityEvent{
		ID:             "secev-102",
		Timestamp:      now.Add(-45 * time.Minute),
		EventType:      EventRateLimitViolation,
		Severity:       SeverityMedium,
		Result:         "DENIED",
		ActorID:        "anonymous-ip-198.51.100.42",
		OrganizationID: "org-default",
		RequestID:      "req_ratelimit_02",
		Details:        "Rate limit threshold exceeded on /api/v1/auth/login from IP 198.51.100.42. Returned HTTP 429.",
	}, &SecurityEvent{
		ID:             "secev-103",
		Timestamp:      now.Add(-2 * time.Hour),
		EventType:      EventStorageTraversalBlocked,
		Severity:       SeverityHigh,
		Result:         "REJECTED",
		ActorID:        "usr-dev",
		OrganizationID: "org-default",
		RequestID:      "req_storage_trav_03",
		Details:        "Storage Path Security blocked object key attempting directory traversal '../etc/passwd'",
	})
}
