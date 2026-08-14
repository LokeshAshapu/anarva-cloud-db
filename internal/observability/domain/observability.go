package domain

import (
	"time"
)

type HealthState string

const (
	HealthHealthy           HealthState = "HEALTHY"
	HealthDegraded          HealthState = "DEGRADED"
	HealthUnavailable       HealthState = "UNAVAILABLE"
	HealthProvisioning      HealthState = "PROVISIONING"
	HealthUpdating          HealthState = "UPDATING"
	HealthDeleting          HealthState = "DELETING"
	HealthStopped           HealthState = "STOPPED"
	HealthDrifted           HealthState = "DRIFTED"
	HealthUnknown           HealthState = "UNKNOWN"
	HealthExternallyDeleted HealthState = "EXTERNALLY_DELETED"
)

type DriftStatus string

const (
	DriftInSync            DriftStatus = "IN_SYNC"
	DriftStateDrift        DriftStatus = "STATE_DRIFT"
	DriftConfigDrift       DriftStatus = "CONFIGURATION_DRIFT"
	DriftSecurityDrift     DriftStatus = "SECURITY_DRIFT"
	DriftMissingResource   DriftStatus = "MISSING_RESOURCE"
	DriftExternalChange    DriftStatus = "EXTERNAL_CHANGE"
	DriftUnknown           DriftStatus = "UNKNOWN"
)

type ResourceObservation struct {
	ResourceID             string            `json:"resourceId"`
	OrganizationID         string            `json:"organizationId"`
	ProjectID              string            `json:"projectId"`
	ResourceName           string            `json:"resourceName"`
	Provider               string            `json:"provider"`
	ResourceType           string            `json:"resourceType"` // EC2, RDS_POSTGRESQL, S3_BUCKET
	ProviderResourceID     string            `json:"providerResourceId"`
	Region                 string            `json:"region"`
	DesiredState           string            `json:"desiredState"`
	ObservedState          string            `json:"observedState"`
	HealthState            HealthState       `json:"healthState"`
	DriftStatus            DriftStatus       `json:"driftStatus"`
	DriftDetails           string            `json:"driftDetails,omitempty"`
	LastObservedAt         time.Time         `json:"lastObservedAt"`
	ObservationDurationMs  int64             `json:"observationDurationMs"`
	ObservationError       string            `json:"observationError,omitempty"`
	IsStale                bool              `json:"isStale"`
	ObservedAttributes     map[string]string `json:"observedAttributes,omitempty"`
}

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
