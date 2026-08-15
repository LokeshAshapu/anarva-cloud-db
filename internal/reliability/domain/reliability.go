package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type OperationStatus string

const (
	OpStatusQueued    OperationStatus = "QUEUED"
	OpStatusRunning   OperationStatus = "RUNNING"
	OpStatusSucceeded OperationStatus = "SUCCEEDED"
	OpStatusFailed    OperationStatus = "FAILED"
	OpStatusCancelled OperationStatus = "CANCELLED"
	OpStatusTimedOut  OperationStatus = "TIMED_OUT"
)

type OperationType string

const (
	OpCreateCompute  OperationType = "CREATE_COMPUTE"
	OpDeleteCompute  OperationType = "DELETE_COMPUTE"
	OpCreateDatabase OperationType = "CREATE_DATABASE"
	OpUpdateDatabase OperationType = "UPDATE_DATABASE"
	OpDeleteDatabase OperationType = "DELETE_DATABASE"
	OpCreateStorage  OperationType = "CREATE_STORAGE"
	OpDeleteStorage  OperationType = "DELETE_STORAGE"
	OpBackupDatabase OperationType = "BACKUP_DATABASE"
	OpRestoreDatabase OperationType = "RESTORE_DATABASE"
	OpFailoverDatabase OperationType = "FAILOVER_DATABASE"
)

type OperationEvent struct {
	StepNumber  int       `json:"stepNumber"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // COMPLETED, RUNNING, PENDING, FAILED
	Timestamp   time.Time `json:"timestamp"`
}

type AnarvaOperation struct {
	ID             string           `json:"id"`
	OrganizationID string           `json:"organizationId"`
	ProjectID      string           `json:"projectId"`
	ResourceID     string           `json:"resourceId"`
	Type           OperationType    `json:"type"`
	Status         OperationStatus  `json:"status"`
	Progress       int              `json:"progress"` // 0 to 100
	CreatedAt      time.Time        `json:"createdAt"`
	StartedAt      *time.Time       `json:"startedAt,omitempty"`
	UpdatedAt      time.Time        `json:"updatedAt"`
	CompletedAt    *time.Time       `json:"completedAt,omitempty"`
	ErrorCode      string           `json:"errorCode,omitempty"`
	ErrorMessage   string           `json:"errorMessage,omitempty"`
	RequestID      string           `json:"requestId"`
	IdempotencyKey string           `json:"idempotencyKey,omitempty"`
	Timeline       []OperationEvent `json:"timeline"`
}

type IdempotencyRecord struct {
	Key            string    `json:"key"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	RequestHash    string    `json:"requestHash"`
	OperationID    string    `json:"operationId"`
	ResourceID     string    `json:"resourceId"`
	CreatedAt      time.Time `json:"createdAt"`
	ExpiresAt      time.Time `json:"expiresAt"`
}

type ResourceLockLease struct {
	ResourceID  string    `json:"resourceId"`
	OperationID string    `json:"operationId"`
	Owner       string    `json:"owner"`
	LockedAt    time.Time `json:"lockedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type TenantQuota struct {
	OrganizationID string  `json:"organizationId"`
	ProjectID      string  `json:"projectId"`
	MaxACU         float64 `json:"maxAcu"`
	CurrentACU     float64 `json:"currentAcu"`
	MaxDatabases   int     `json:"maxDatabases"`
	CurrentDbs     int     `json:"currentDbs"`
	MaxStorageGB   int     `json:"maxStorageGb"`
	CurrentStorageGB int   `json:"currentStorageGb"`
}

type AuditActorType string

const (
	ActorUser   AuditActorType = "USER"
	ActorAPIKey AuditActorType = "API_KEY"
	ActorSystem AuditActorType = "SYSTEM"
)

type AnarvaAuditEvent struct {
	ID             string         `json:"id"`
	OrganizationID string         `json:"organizationId"`
	ProjectID      string         `json:"projectId"`
	ActorType      AuditActorType `json:"actorType"`
	ActorID        string         `json:"actorId"`
	Action         string         `json:"action"`
	ResourceType   string         `json:"resourceType"`
	ResourceID     string         `json:"resourceId"`
	OperationID    string         `json:"operationId,omitempty"`
	RequestID      string         `json:"requestId"`
	Timestamp      time.Time      `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type AnarvaEvent struct {
	EventID        string    `json:"eventId"`
	Timestamp      time.Time `json:"timestamp"`
	EventType      string    `json:"eventType"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	ResourceID     string    `json:"resourceId"`
	OperationID    string    `json:"operationId,omitempty"`
}

func HashRequestPayload(payload string) string {
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func FormatOperationID() string {
	return fmt.Sprintf("op-%d", time.Now().UnixNano()/1e6)
}
