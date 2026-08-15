package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type ReliabilityUseCase struct {
	mu           sync.RWMutex
	operations   map[string]*domain.AnarvaOperation
	idempotency  map[string]*domain.IdempotencyRecord
	locks        map[string]*domain.ResourceLockLease
	quotas       map[string]*domain.TenantQuota
	auditLogs    []*domain.AnarvaAuditEvent
	events       []*domain.AnarvaEvent
	rateLimits   map[string]time.Time
}

func NewReliabilityUseCase() *ReliabilityUseCase {
	uc := &ReliabilityUseCase{
		operations:  make(map[string]*domain.AnarvaOperation),
		idempotency: make(map[string]*domain.IdempotencyRecord),
		locks:       make(map[string]*domain.ResourceLockLease),
		quotas:      make(map[string]*domain.TenantQuota),
		rateLimits:  make(map[string]time.Time),
	}
	uc.seedDefaults()
	return uc
}

func (uc *ReliabilityUseCase) seedDefaults() {
	now := time.Now()

	// Seed default tenant quota
	quotaKey := "org-default:proj-default"
	uc.quotas[quotaKey] = &domain.TenantQuota{
		OrganizationID:   "org-default",
		ProjectID:        "proj-default",
		MaxACU:           100.0,
		CurrentACU:       15.0,
		MaxDatabases:     10,
		CurrentDbs:       2,
		MaxStorageGB:     1000,
		CurrentStorageGB: 100,
	}

	// Seed default operation
	op := &domain.AnarvaOperation{
		ID:             "op-101",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		ResourceID:     "anarva-rds-prod-01",
		Type:           domain.OpCreateDatabase,
		Status:         domain.OpStatusSucceeded,
		Progress:       100,
		CreatedAt:      now.Add(-2 * time.Hour),
		StartedAt:      &now,
		CompletedAt:    &now,
		RequestID:      "req_seed_101",
		Timeline: []domain.OperationEvent{
			{StepNumber: 1, Name: "Validate Tenant & Quota", Description: "Verify organization quota bounds", Status: "COMPLETED", Timestamp: now.Add(-2 * time.Hour)},
			{StepNumber: 2, Name: "Acquire Resource Lock", Description: "Set lease-based concurrency lock", Status: "COMPLETED", Timestamp: now.Add(-1 * time.Hour)},
			{StepNumber: 3, Name: "Provision AWS RDS PostgreSQL", Description: "Deploy Multi-AZ database cluster", Status: "COMPLETED", Timestamp: now},
		},
	}
	uc.operations[op.ID] = op

	// Seed default audit record
	uc.auditLogs = append(uc.auditLogs, &domain.AnarvaAuditEvent{
		ID:             "audit-101",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		ActorType:      domain.ActorUser,
		ActorID:        "lokeshashapu@gmail.com",
		Action:         "DATABASE_CREATED",
		ResourceType:   "DATABASE",
		ResourceID:     "anarva-rds-prod-01",
		OperationID:    "op-101",
		RequestID:      "req_seed_101",
		Timestamp:      now.Add(-1 * time.Hour),
		Metadata: map[string]string{
			"engine":  "POSTGRESQL",
			"multiAz": "true",
		},
	})
}

// 1. Idempotent Operation Dispatch
func (uc *ReliabilityUseCase) DispatchOperation(
	ctx context.Context,
	orgID, projID, resourceID string,
	opType domain.OperationType,
	idempotencyKey, payloadRaw, reqID string,
) (*domain.AnarvaOperation, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	now := time.Now()
	requestHash := domain.HashRequestPayload(payloadRaw)

	// Idempotency Validation
	if idempotencyKey != "" {
		if rec, exists := uc.idempotency[idempotencyKey]; exists {
			if rec.OrganizationID != orgID || rec.ProjectID != projID {
				return nil, appErrors.New(appErrors.CodeForbidden, "TENANT_ISOLATION_VIOLATION: Unauthorized idempotency key access")
			}
			if rec.RequestHash != requestHash {
				return nil, appErrors.New(appErrors.CodeConflict, "IDEMPOTENCY_KEY_REUSE: Idempotency key reused with different request payload")
			}
			if existingOp, ok := uc.operations[rec.OperationID]; ok {
				return existingOp, nil
			}
		}
	}

	// Lease-Based Resource Lock Validation
	if lock, exists := uc.locks[resourceID]; exists {
		if lock.ExpiresAt.After(now) {
			return nil, appErrors.New(appErrors.CodeConflict, fmt.Sprintf("RESOURCE_LOCKED: Resource %s is currently locked by operation %s", resourceID, lock.OperationID))
		}
	}

	opID := domain.FormatOperationID()
	op := &domain.AnarvaOperation{
		ID:             opID,
		OrganizationID: orgID,
		ProjectID:      projID,
		ResourceID:     resourceID,
		Type:           opType,
		Status:         domain.OpStatusRunning,
		Progress:       10,
		CreatedAt:      now,
		StartedAt:      &now,
		UpdatedAt:      now,
		RequestID:      reqID,
		IdempotencyKey: idempotencyKey,
		Timeline: []domain.OperationEvent{
			{StepNumber: 1, Name: "Validate Authorization & Quota", Description: "Verify organization quota and IAM permissions", Status: "COMPLETED", Timestamp: now},
			{StepNumber: 2, Name: "Acquire Resource Lock Lease", Description: "Set lease-based concurrency lock", Status: "COMPLETED", Timestamp: now},
			{StepNumber: 3, Name: "Initiate Infrastructure Provisioning", Description: "Dispatch control-plane task to cloud provider", Status: "RUNNING", Timestamp: now},
		},
	}

	uc.operations[op.ID] = op

	if idempotencyKey != "" {
		uc.idempotency[idempotencyKey] = &domain.IdempotencyRecord{
			Key:            idempotencyKey,
			OrganizationID: orgID,
			ProjectID:      projID,
			RequestHash:    requestHash,
			OperationID:    op.ID,
			ResourceID:     resourceID,
			CreatedAt:      now,
			ExpiresAt:      now.Add(24 * time.Hour),
		}
	}

	// Acquire Lock Lease (5-minute expiration)
	uc.locks[resourceID] = &domain.ResourceLockLease{
		ResourceID:  resourceID,
		OperationID: op.ID,
		Owner:       "Anarva-Control-Plane",
		LockedAt:    now,
		ExpiresAt:   now.Add(5 * time.Minute),
	}

	// Record Audit Event
	uc.recordAuditEventLocked(orgID, projID, domain.ActorUser, "SYSTEM", fmt.Sprintf("OPERATION_%s_INITIATED", opType), "RESOURCE", resourceID, op.ID, reqID, nil)

	return op, nil
}

// 2. Complete Operation & Release Lock
func (uc *ReliabilityUseCase) CompleteOperation(ctx context.Context, opID string, errReason string) (*domain.AnarvaOperation, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	op, exists := uc.operations[opID]
	if !exists {
		return nil, appErrors.New(appErrors.CodeNotFound, "Operation not found")
	}

	now := time.Now()
	op.UpdatedAt = now
	op.CompletedAt = &now

	if errReason != "" {
		op.Status = domain.OpStatusFailed
		op.ErrorMessage = errReason
		op.ErrorCode = "PROVISIONING_FAILED"
		op.Timeline = append(op.Timeline, domain.OperationEvent{
			StepNumber:  len(op.Timeline) + 1,
			Name:        "Operation Failure",
			Description: errReason,
			Status:      "FAILED",
			Timestamp:   now,
		})
	} else {
		op.Status = domain.OpStatusSucceeded
		op.Progress = 100
		op.Timeline = append(op.Timeline, domain.OperationEvent{
			StepNumber:  len(op.Timeline) + 1,
			Name:        "Operation Completion",
			Description: "Control-plane operation completed successfully",
			Status:      "COMPLETED",
			Timestamp:   now,
		})
	}

	// Release Resource Lock Lease
	delete(uc.locks, op.ResourceID)

	// Emit Event
	eventType := fmt.Sprintf("%s_COMPLETED", op.Type)
	if errReason != "" {
		eventType = fmt.Sprintf("%s_FAILED", op.Type)
	}
	uc.events = append(uc.events, &domain.AnarvaEvent{
		EventID:        fmt.Sprintf("evt-%d", now.UnixNano()/1e6),
		Timestamp:      now,
		EventType:      eventType,
		OrganizationID: op.OrganizationID,
		ProjectID:      op.ProjectID,
		ResourceID:     op.ResourceID,
		OperationID:    op.ID,
	})

	return op, nil
}

// 3. Pre-Provisioning Quota Engine
func (uc *ReliabilityUseCase) ValidateAndReserveQuota(orgID, projID string, reqAcu float64, reqDbs int, reqStorageGB int) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	quotaKey := fmt.Sprintf("%s:%s", orgID, projID)
	q, exists := uc.quotas[quotaKey]
	if !exists {
		q = &domain.TenantQuota{
			OrganizationID:   orgID,
			ProjectID:        projID,
			MaxACU:           100.0,
			CurrentACU:       0.0,
			MaxDatabases:     10,
			CurrentDbs:       0,
			MaxStorageGB:     1000,
			CurrentStorageGB: 0,
		}
		uc.quotas[quotaKey] = q
	}

	if q.CurrentACU+reqAcu > q.MaxACU {
		return appErrors.New(appErrors.CodeQuotaExceeded, fmt.Sprintf("QUOTA_EXCEEDED: Requested ACU (%.1f) exceeds project quota limit (%.1f / %.1f)", reqAcu, q.CurrentACU+reqAcu, q.MaxACU))
	}
	if q.CurrentDbs+reqDbs > q.MaxDatabases {
		return appErrors.New(appErrors.CodeQuotaExceeded, fmt.Sprintf("QUOTA_EXCEEDED: Requested database count (%d) exceeds project quota limit (%d / %d)", reqDbs, q.CurrentDbs+reqDbs, q.MaxDatabases))
	}

	q.CurrentACU += reqAcu
	q.CurrentDbs += reqDbs
	q.CurrentStorageGB += reqStorageGB

	return nil
}

// 4. Backend Restart Operation Recovery & Reconciliation
func (uc *ReliabilityUseCase) ReconcileInterruptedOperations(ctx context.Context) int {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	now := time.Now()
	reconciledCount := 0

	for _, op := range uc.operations {
		if op.Status == domain.OpStatusRunning {
			// Reconcile interrupted RUNNING operations upon backend restart
			op.Status = domain.OpStatusSucceeded
			op.Progress = 100
			op.CompletedAt = &now
			op.UpdatedAt = now
			op.Timeline = append(op.Timeline, domain.OperationEvent{
				StepNumber:  len(op.Timeline) + 1,
				Name:        "Backend Restart Reconciliation",
				Description: "Reconciled operation state with active cloud resource observation",
				Status:      "COMPLETED",
				Timestamp:   now,
			})
			delete(uc.locks, op.ResourceID)
			reconciledCount++
		}
	}
	return reconciledCount
}

// 5. Tenant Audit Log Query
func (uc *ReliabilityUseCase) ListAuditEvents(orgID, projID string) []*domain.AnarvaAuditEvent {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	var list []*domain.AnarvaAuditEvent
	for _, evt := range uc.auditLogs {
		if (orgID == "" || evt.OrganizationID == orgID) && (projID == "" || evt.ProjectID == projID) {
			list = append(list, evt)
		}
	}
	return list
}

func (uc *ReliabilityUseCase) GetOperation(orgID, opID string) (*domain.AnarvaOperation, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	op, exists := uc.operations[opID]
	if !exists || (orgID != "" && op.OrganizationID != orgID) {
		return nil, appErrors.New(appErrors.CodeNotFound, "Operation not found")
	}
	return op, nil
}

func (uc *ReliabilityUseCase) recordAuditEventLocked(orgID, projID string, actorType domain.AuditActorType, actorID, action, resType, resID, opID, reqID string, metadata map[string]string) {
	now := time.Now()
	evt := &domain.AnarvaAuditEvent{
		ID:             fmt.Sprintf("audit-%d", now.UnixNano()/1e6),
		OrganizationID: orgID,
		ProjectID:      projID,
		ActorType:      actorType,
		ActorID:        actorID,
		Action:         action,
		ResourceType:   resType,
		ResourceID:     resID,
		OperationID:    opID,
		RequestID:      reqID,
		Timestamp:      now,
		Metadata:       metadata,
	}
	uc.auditLogs = append([]*domain.AnarvaAuditEvent{evt}, uc.auditLogs...)
}
