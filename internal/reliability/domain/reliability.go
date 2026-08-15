package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

type OperationStatus string

const (
	OpStatusPending    OperationStatus = "PENDING"
	OpStatusQueued     OperationStatus = "QUEUED"
	OpStatusRunning    OperationStatus = "RUNNING"
	OpStatusSucceeded  OperationStatus = "SUCCEEDED"
	OpStatusFailed     OperationStatus = "FAILED"
	OpStatusCancelled  OperationStatus = "CANCELLED"
	OpStatusTimedOut   OperationStatus = "TIMED_OUT"
	OpStatusRecovering OperationStatus = "RECOVERING"
)

func IsValidStateTransition(from, to OperationStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case OpStatusPending:
		return to == OpStatusQueued || to == OpStatusRunning || to == OpStatusCancelled || to == OpStatusFailed
	case OpStatusQueued:
		return to == OpStatusRunning || to == OpStatusCancelled || to == OpStatusFailed
	case OpStatusRunning:
		return to == OpStatusSucceeded || to == OpStatusFailed || to == OpStatusCancelled || to == OpStatusTimedOut || to == OpStatusRecovering
	case OpStatusRecovering:
		return to == OpStatusRunning || to == OpStatusSucceeded || to == OpStatusFailed || to == OpStatusCancelled
	case OpStatusSucceeded, OpStatusFailed, OpStatusCancelled, OpStatusTimedOut:
		return false // Terminal states
	default:
		return true
	}
}

type OperationType string

const (
	OpCreateCompute    OperationType = "CREATE_COMPUTE"
	OpDeleteCompute    OperationType = "DELETE_COMPUTE"
	OpCreateDatabase   OperationType = "CREATE_DATABASE"
	OpUpdateDatabase   OperationType = "UPDATE_DATABASE"
	OpDeleteDatabase   OperationType = "DELETE_DATABASE"
	OpCreateStorage    OperationType = "CREATE_STORAGE"
	OpDeleteStorage    OperationType = "DELETE_STORAGE"
	OpBackupDatabase   OperationType = "BACKUP_DATABASE"
	OpRestoreDatabase  OperationType = "RESTORE_DATABASE"
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
	ID                 string           `json:"id" gorm:"primaryKey"`
	OrganizationID     string           `json:"organizationId" gorm:"index:idx_op_tenant"`
	ProjectID          string           `json:"projectId" gorm:"index:idx_op_tenant"`
	ResourceID         string           `json:"resourceId" gorm:"index:idx_op_resource"`
	ResourceType       string           `json:"resourceType,omitempty"`
	Type               OperationType    `json:"type"`
	Status             OperationStatus  `json:"status" gorm:"index:idx_op_status"`
	Progress           int              `json:"progress"` // 0 to 100
	CreatedAt          time.Time        `json:"createdAt"`
	StartedAt          *time.Time       `json:"startedAt,omitempty"`
	UpdatedAt          time.Time        `json:"updatedAt"`
	CompletedAt        *time.Time       `json:"completedAt,omitempty"`
	HeartbeatAt        time.Time        `json:"heartbeatAt" gorm:"index:idx_op_heartbeat"`
	LeaseExpiresAt     time.Time        `json:"leaseExpiresAt" gorm:"index:idx_op_lease"`
	RetryCount         int              `json:"retryCount"`
	ErrorCode          string           `json:"errorCode,omitempty"`
	ErrorMessage       string           `json:"errorMessage,omitempty"`
	RequestID          string           `json:"requestId" gorm:"index:idx_op_req"`
	IdempotencyKey     string           `json:"idempotencyKey,omitempty"`
	IdempotencyKeyHash string           `json:"idempotencyKeyHash,omitempty" gorm:"index:idx_op_idemp_hash"`
	Timeline           []OperationEvent `json:"timeline" gorm:"-"`
	TimelineJSON       string           `json:"-" gorm:"type:text"`
}

type IdempotencyRecord struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	Key            string    `json:"key"`
	KeyHash        string    `json:"keyHash" gorm:"uniqueIndex:idx_tenant_idemp,priority:3"`
	OrganizationID string    `json:"organizationId" gorm:"uniqueIndex:idx_tenant_idemp,priority:1"`
	ProjectID      string    `json:"projectId" gorm:"uniqueIndex:idx_tenant_idemp,priority:2"`
	RequestHash    string    `json:"requestHash"`
	OperationID    string    `json:"operationId"`
	ResourceID     string    `json:"resourceId"`
	CreatedAt      time.Time `json:"createdAt"`
	ExpiresAt      time.Time `json:"expiresAt" gorm:"index:idx_idemp_exp"`
}

type ResourceLockLease struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ResourceID     string    `json:"resourceId" gorm:"uniqueIndex:idx_res_lock_res"`
	OrganizationID string    `json:"organizationId" gorm:"index:idx_lock_tenant"`
	ProjectID      string    `json:"projectId" gorm:"index:idx_lock_tenant"`
	ResourceType   string    `json:"resourceType,omitempty"`
	OperationID    string    `json:"operationId" gorm:"index:idx_lock_op"`
	HolderID       string    `json:"holderId"`
	Owner          string    `json:"owner"`
	AcquiredAt     time.Time `json:"acquiredAt"`
	LockedAt       time.Time `json:"lockedAt"`
	ExpiresAt      time.Time `json:"expiresAt" gorm:"index:idx_lock_exp"`
	HeartbeatAt    time.Time `json:"heartbeatAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type TenantQuota struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	OrganizationID   string    `json:"organizationId" gorm:"uniqueIndex:idx_tenant_quota_scope,priority:1"`
	ProjectID        string    `json:"projectId" gorm:"uniqueIndex:idx_tenant_quota_scope,priority:2"`
	MaxACU           float64   `json:"maxAcu"`
	CurrentACU       float64   `json:"currentAcu"`
	MaxDatabases     int       `json:"maxDatabases"`
	CurrentDbs       int       `json:"currentDbs"`
	MaxStorageGB     int       `json:"maxStorageGb"`
	CurrentStorageGB int       `json:"currentStorageGb"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type AuditActorType string

const (
	ActorUser   AuditActorType = "USER"
	ActorAPIKey AuditActorType = "API_KEY"
	ActorSystem AuditActorType = "SYSTEM"
)

type AnarvaAuditEvent struct {
	ID             string            `json:"id" gorm:"primaryKey"`
	OrganizationID string            `json:"organizationId" gorm:"index:idx_audit_tenant"`
	ProjectID      string            `json:"projectId" gorm:"index:idx_audit_tenant"`
	ActorType      AuditActorType    `json:"actorType"`
	ActorID        string            `json:"actorId"`
	Action         string            `json:"action"`
	ResourceType   string            `json:"resourceType"`
	ResourceID     string            `json:"resourceId"`
	OperationID    string            `json:"operationId,omitempty"`
	RequestID      string            `json:"requestId"`
	Timestamp      time.Time         `json:"timestamp" gorm:"index:idx_audit_time"`
	Metadata       map[string]string `json:"metadata,omitempty" gorm:"-"`
	MetadataJSON   string            `json:"-" gorm:"type:text"`
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

var opCounter uint64

func FormatOperationID() string {
	seq := atomic.AddUint64(&opCounter, 1)
	return fmt.Sprintf("op-%d-%d", time.Now().UnixNano()/1e6, seq)
}
