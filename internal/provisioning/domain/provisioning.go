package domain

import (
	"time"
)

type ProvisioningStatus string

const (
	StatusRequested   ProvisioningStatus = "REQUESTED"
	StatusValidating  ProvisioningStatus = "VALIDATING"
	StatusPlanning    ProvisioningStatus = "PLANNING"
	StatusQueued      ProvisioningStatus = "QUEUED"
	StatusProvisioning ProvisioningStatus = "PROVISIONING"
	StatusConfiguring ProvisioningStatus = "CONFIGURING"
	StatusVerifying   ProvisioningStatus = "VERIFYING"
	StatusCompleted   ProvisioningStatus = "COMPLETED"
	StatusFailed      ProvisioningStatus = "FAILED"
	StatusRollingBack ProvisioningStatus = "ROLLING_BACK"
	StatusRolledBack  ProvisioningStatus = "ROLLED_BACK"
	StatusCancelled   ProvisioningStatus = "CANCELLED"
	StatusUnknown     ProvisioningStatus = "UNKNOWN"
)

type ResourceType string

const (
	TypeCompute      ResourceType = "COMPUTE"
	TypeDatabase     ResourceType = "DATABASE"
	TypeStorage      ResourceType = "STORAGE"
	TypeNetwork      ResourceType = "NETWORK"
	TypeSubnet       ResourceType = "SUBNET"
	TypeVolume       ResourceType = "VOLUME"
	TypeLoadBalancer ResourceType = "LOAD_BALANCER"
	TypeDNS          ResourceType = "DNS"
	TypeBackup       ResourceType = "BACKUP"
)

type CapabilityOperation string

const (
	OpCreate  CapabilityOperation = "CREATE"
	OpUpdate  CapabilityOperation = "UPDATE"
	OpDelete  CapabilityOperation = "DELETE"
	OpStart   CapabilityOperation = "START"
	OpStop    CapabilityOperation = "STOP"
	OpRestart CapabilityOperation = "RESTART"
	OpResize  CapabilityOperation = "RESIZE"
	OpAttach  CapabilityOperation = "ATTACH"
	OpDetach  CapabilityOperation = "DETACH"
	OpRestore CapabilityOperation = "RESTORE"
)

type LockState string

const (
	LockStateAvailable LockState = "AVAILABLE"
	LockStateBusy      LockState = "BUSY"
	LockStateLocked    LockState = "LOCKED"
)

type DriftStatus string

const (
	DriftInSync  DriftStatus = "IN_SYNC"
	DriftDrifted DriftStatus = "DRIFTED"
	DriftMissing DriftStatus = "MISSING"
	DriftUnknown DriftStatus = "UNKNOWN"
)

type ProviderCapability struct {
	Provider     string              `json:"provider"`
	ResourceType ResourceType        `json:"resourceType"`
	Operation    CapabilityOperation `json:"operation"`
	Status       string              `json:"status"` // SUPPORTED, UNSUPPORTED, DEPRECATED
	Version      string              `json:"version"`
	Metadata     map[string]string   `json:"metadata,omitempty"`
}

type ExecutionStep struct {
	StepNumber  int    `json:"stepNumber"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"` // PENDING, RUNNING, COMPLETED, FAILED, SKIPPED
	Error       string `json:"error,omitempty"`
}

type ExecutionPlan struct {
	ID               string          `json:"id"`
	RequestID        string          `json:"requestId"`
	Steps            []ExecutionStep `json:"steps"`
	TotalActions     int             `json:"totalActions"`
	EstimatedTimeSec int             `json:"estimatedTimeSec"`
}

type ProvisioningRequest struct {
	ID             string             `json:"id"`
	OrganizationID string             `json:"organizationId"`
	ProjectID      string             `json:"projectId"`
	ResourceType   ResourceType       `json:"resourceType"`
	ResourceID     string             `json:"resourceId"`
	Provider       string             `json:"provider"`
	RegionID       string             `json:"regionId"`
	ZoneID         string             `json:"zoneId,omitempty"`
	Status         ProvisioningStatus `json:"status"`
	RequestedBy    string             `json:"requestedBy"`
	IdempotencyKey string             `json:"idempotencyKey,omitempty"`
	Plan           string             `json:"plan,omitempty"`
	ExecutionPlan  *ExecutionPlan     `json:"executionPlan,omitempty"`
	ErrorCode      string             `json:"errorCode,omitempty"`
	ErrorMessage   string             `json:"errorMessage,omitempty"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	StartedAt      *time.Time         `json:"startedAt,omitempty"`
	CompletedAt    *time.Time         `json:"completedAt,omitempty"`
}

type ResourceLock struct {
	ResourceID string    `json:"resourceId"`
	Status     LockState `json:"status"`
	LockedBy   string    `json:"lockedBy"`
	LockedAt   time.Time `json:"lockedAt"`
}

type ResourceDrift struct {
	ResourceID        string      `json:"resourceId"`
	ControlPlaneState string      `json:"controlPlaneState"`
	ProviderState     string      `json:"providerState"`
	Status            DriftStatus `json:"status"`
	Details           string      `json:"details"`
	DetectedAt        time.Time   `json:"detectedAt"`
}
