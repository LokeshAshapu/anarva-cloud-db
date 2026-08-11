package domain

import (
	"time"
)

type HealthState string

const (
	HealthHealthy     HealthState = "HEALTHY"
	HealthDegraded    HealthState = "DEGRADED"
	HealthUnavailable HealthState = "UNAVAILABLE"
	HealthMaintenance HealthState = "MAINTENANCE"
	HealthUnknown     HealthState = "UNKNOWN"
)

type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "INFO"
	SeverityWarning  AlertSeverity = "WARNING"
	SeverityCritical AlertSeverity = "CRITICAL"
)

type AlertStatus string

const (
	AlertActive       AlertStatus = "ACTIVE"
	AlertAcknowledged AlertStatus = "ACKNOWLEDGED"
	AlertResolved     AlertStatus = "RESOLVED"
)

type MetricRecord struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organizationId"`
	ProjectID      string            `json:"projectId"`
	ResourceID     string            `json:"resourceId,omitempty"`
	MetricName     string            `json:"metricName"`
	Value          float64           `json:"value"`
	Unit           string            `json:"unit"`
	Timestamp      time.Time         `json:"timestamp"`
	Source         string            `json:"source"`
	Labels         map[string]string `json:"labels,omitempty"`
}

type LogRecord struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organizationId"`
	ProjectID      string            `json:"projectId"`
	ResourceID     string            `json:"resourceId,omitempty"`
	Service        string            `json:"service"`
	Level          string            `json:"level"` // DEBUG, INFO, WARN, ERROR, FATAL
	Message        string            `json:"message"`
	RequestID      string            `json:"requestId,omitempty"`
	TraceID        string            `json:"traceId,omitempty"`
	Timestamp      time.Time         `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type AlertInstance struct {
	ID             string        `json:"id"`
	OrganizationID string        `json:"organizationId"`
	ProjectID      string        `json:"projectId"`
	ResourceID     string        `json:"resourceId,omitempty"`
	Name           string        `json:"name"`
	Severity       AlertSeverity `json:"severity"`
	Condition      string        `json:"condition"`
	Status         AlertStatus   `json:"status"`
	TriggeredAt    time.Time     `json:"triggeredAt"`
	ResolvedAt     *time.Time    `json:"resolvedAt,omitempty"`
}

type IncidentRecord struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organizationId"`
	Title          string     `json:"title"`
	Severity       string     `json:"severity"`
	Status         string     `json:"status"` // INVESTIGATING, IDENTIFIED, MONITORING, RESOLVED
	StartedAt      time.Time  `json:"startedAt"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	Summary        string     `json:"summary"`
}
