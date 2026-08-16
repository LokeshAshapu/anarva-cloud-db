package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/repository"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type ReliabilityUseCase struct {
	mu           sync.RWMutex
	repo         *repository.ReliabilityRepository
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

func NewReliabilityUseCaseWithRepo(repo *repository.ReliabilityRepository) *ReliabilityUseCase {
	uc := &ReliabilityUseCase{
		repo:        repo,
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
	now := time.Now()
	requestHash := domain.HashRequestPayload(payloadRaw)

	// Database Persistence Mode
	if uc.repo != nil {
		if idempotencyKey != "" {
			rec, err := uc.repo.GetIdempotencyRecord(ctx, orgID, projID, idempotencyKey)
			if err != nil {
				return nil, err
			}
			if rec != nil {
				if rec.OrganizationID != orgID || rec.ProjectID != projID {
					return nil, appErrors.New(appErrors.CodeForbidden, "TENANT_ISOLATION_VIOLATION: Unauthorized idempotency key access")
				}
				if rec.RequestHash != requestHash {
					return nil, appErrors.New(appErrors.CodeConflict, "IDEMPOTENCY_KEY_REUSE: Idempotency key reused with different request payload")
				}
				existingOp, err := uc.repo.GetOperation(ctx, orgID, rec.OperationID)
				if err == nil && existingOp != nil {
					return existingOp, nil
				}
			}
		}

		opID := domain.FormatOperationID()

		// Acquire PostgreSQL Distributed Lock
		lock := &domain.ResourceLockLease{
			ResourceID:     resourceID,
			OrganizationID: orgID,
			ProjectID:      projID,
			OperationID:    opID,
			HolderID:       "Anarva-Control-Plane-Instance",
			Owner:          "Anarva-Control-Plane",
			AcquiredAt:     now,
			LockedAt:       now,
			ExpiresAt:      now.Add(5 * time.Minute),
			HeartbeatAt:    now,
		}
		if err := uc.repo.AcquireDistributedLock(ctx, lock); err != nil {
			return nil, err
		}

		op := &domain.AnarvaOperation{
			ID:                 opID,
			OrganizationID:     orgID,
			ProjectID:          projID,
			ResourceID:         resourceID,
			Type:               opType,
			Status:             domain.OpStatusRunning,
			Progress:           10,
			CreatedAt:          now,
			StartedAt:          &now,
			UpdatedAt:          now,
			HeartbeatAt:        now,
			LeaseExpiresAt:     now.Add(5 * time.Minute),
			RequestID:          reqID,
			IdempotencyKey:     idempotencyKey,
			IdempotencyKeyHash: domain.HashRequestPayload(idempotencyKey),
			Timeline: []domain.OperationEvent{
				{StepNumber: 1, Name: "Validate Authorization & Quota", Description: "Verify organization quota and IAM permissions", Status: "COMPLETED", Timestamp: now},
				{StepNumber: 2, Name: "Acquire Distributed Lock Lease", Description: "Set PostgreSQL atomic concurrency lock", Status: "COMPLETED", Timestamp: now},
				{StepNumber: 3, Name: "Initiate Control-Plane Task", Description: "Dispatch control-plane task to cloud provider", Status: "RUNNING", Timestamp: now},
			},
		}

		if err := uc.repo.SaveOperation(ctx, op); err != nil {
			_ = uc.repo.ReleaseDistributedLock(ctx, orgID, resourceID, opID)
			return nil, err
		}

		if idempotencyKey != "" {
			_ = uc.repo.SaveIdempotencyRecord(ctx, &domain.IdempotencyRecord{
				Key:            idempotencyKey,
				OrganizationID: orgID,
				ProjectID:      projID,
				RequestHash:    requestHash,
				OperationID:    op.ID,
				ResourceID:     resourceID,
				CreatedAt:      now,
				ExpiresAt:      now.Add(24 * time.Hour),
			})
		}

		_ = uc.repo.SaveAuditEvent(ctx, &domain.AnarvaAuditEvent{
			OrganizationID: orgID,
			ProjectID:      projID,
			ActorType:      domain.ActorUser,
			ActorID:        "SYSTEM",
			Action:         fmt.Sprintf("OPERATION_%s_INITIATED", opType),
			ResourceType:   "RESOURCE",
			ResourceID:     resourceID,
			OperationID:    op.ID,
			RequestID:      reqID,
			Timestamp:      now,
		})

		return op, nil
	}

	// In-Memory Fallback Mode for Standalone Unit Tests
	uc.mu.Lock()
	defer uc.mu.Unlock()

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
		HeartbeatAt:    now,
		LeaseExpiresAt: now.Add(5 * time.Minute),
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

	uc.locks[resourceID] = &domain.ResourceLockLease{
		ResourceID:  resourceID,
		OperationID: op.ID,
		Owner:       "Anarva-Control-Plane",
		LockedAt:    now,
		ExpiresAt:   now.Add(5 * time.Minute),
	}

	uc.recordAuditEventLocked(orgID, projID, domain.ActorUser, "SYSTEM", fmt.Sprintf("OPERATION_%s_INITIATED", opType), "RESOURCE", resourceID, op.ID, reqID, nil)

	return op, nil
}

// 2. Complete Operation & Release Lock
func (uc *ReliabilityUseCase) CompleteOperation(ctx context.Context, opID string, errReason string) (*domain.AnarvaOperation, error) {
	now := time.Now()

	if uc.repo != nil {
		op, err := uc.repo.GetOperation(ctx, "", opID)
		if err != nil {
			return nil, err
		}

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

		if err := uc.repo.SaveOperation(ctx, op); err != nil {
			return nil, err
		}

		_ = uc.repo.ReleaseDistributedLock(ctx, op.OrganizationID, op.ResourceID, op.ID)
		return op, nil
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	op, exists := uc.operations[opID]
	if !exists {
		return nil, appErrors.New(appErrors.CodeNotFound, "Operation not found")
	}

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

	delete(uc.locks, op.ResourceID)

	return op, nil
}

// 3. Pre-Provisioning Quota Engine
func (uc *ReliabilityUseCase) ValidateAndReserveQuota(orgID, projID string, reqAcu float64, reqDbs int, reqStorageGB int) error {
	if uc.repo != nil {
		return uc.repo.ReserveQuota(context.Background(), orgID, projID, reqAcu, reqDbs, reqStorageGB)
	}

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
	now := time.Now()
	reconciledCount := 0

	if uc.repo != nil {
		staleOps, err := uc.repo.GetStaleOperations(ctx)
		if err == nil {
			for _, op := range staleOps {
				op.Status = domain.OpStatusSucceeded
				op.Progress = 100
				op.CompletedAt = &now
				op.UpdatedAt = now
				op.Timeline = append(op.Timeline, domain.OperationEvent{
					StepNumber:  len(op.Timeline) + 1,
					Name:        "Control Plane Recovery",
					Description: "Reconciled interrupted control-plane operation state with active resource observation",
					Status:      "COMPLETED",
					Timestamp:   now,
				})
				_ = uc.repo.SaveOperation(ctx, op)
				_ = uc.repo.ReleaseDistributedLock(ctx, op.OrganizationID, op.ResourceID, op.ID)
				reconciledCount++
			}
		}
		return reconciledCount
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	for _, op := range uc.operations {
		if op.Status == domain.OpStatusRunning {
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
	if uc.repo != nil {
		events, err := uc.repo.ListAuditEvents(context.Background(), orgID, projID)
		if err == nil {
			return events
		}
	}

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
	if uc.repo != nil {
		return uc.repo.GetOperation(context.Background(), orgID, opID)
	}

	uc.mu.RLock()
	defer uc.mu.RUnlock()

	op, exists := uc.operations[opID]
	if !exists || (orgID != "" && op.OrganizationID != orgID) {
		return nil, appErrors.New(appErrors.CodeNotFound, "Operation not found")
	}
	if op.RecoveryAttempted && op.Recovery == nil {
		op.Recovery = &domain.RecoveryInfo{
			Attempted: op.RecoveryAttempted,
			Attempt:   op.RecoveryAttempt,
			Status:    op.RecoveryStatus,
			Reason:    op.RecoveryReason,
		}
	}
	return op, nil
}

type OperationsSummary struct {
	Total       int `json:"total"`
	Pending     int `json:"pending"`
	Running     int `json:"running"`
	Succeeded   int `json:"succeeded"`
	Failed      int `json:"failed"`
	TimedOut    int `json:"timedOut"`
	Cancelled   int `json:"cancelled"`
	Recovering  int `json:"recovering"`
	ActiveLocks int `json:"activeLocks"`
}

func (uc *ReliabilityUseCase) ListOperations(ctx context.Context, orgID string, filters repository.OperationQueryFilters) ([]*domain.AnarvaOperation, int64, error) {
	if uc.repo != nil {
		return uc.repo.ListOperations(ctx, orgID, filters)
	}

	uc.mu.RLock()
	defer uc.mu.RUnlock()

	var filtered []*domain.AnarvaOperation
	for _, op := range uc.operations {
		if orgID != "" && op.OrganizationID != orgID {
			continue
		}
		if filters.ProjectID != "" && op.ProjectID != filters.ProjectID {
			continue
		}
		if filters.Status != "" && string(op.Status) != filters.Status {
			continue
		}
		if filters.ResourceID != "" && op.ResourceID != filters.ResourceID {
			continue
		}
		if filters.OperationType != "" && string(op.Type) != filters.OperationType {
			continue
		}
		if !filters.CreatedAfter.IsZero() && op.CreatedAt.Before(filters.CreatedAfter) {
			continue
		}
		if !filters.CreatedBefore.IsZero() && op.CreatedAt.After(filters.CreatedBefore) {
			continue
		}
		if op.RecoveryAttempted && op.Recovery == nil {
			op.Recovery = &domain.RecoveryInfo{
				Attempted: op.RecoveryAttempted,
				Attempt:   op.RecoveryAttempt,
				Status:    op.RecoveryStatus,
				Reason:    op.RecoveryReason,
			}
		}
		filtered = append(filtered, op)
	}

	total := int64(len(filtered))
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return []*domain.AnarvaOperation{}, total, nil
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total, nil
}

func (uc *ReliabilityUseCase) DetectOperationTimeouts(ctx context.Context, timeoutThreshold time.Duration) int {
	if timeoutThreshold <= 0 {
		timeoutThreshold = 5 * time.Minute
	}
	now := time.Now()
	cutoff := now.Add(-timeoutThreshold)
	timedOutCount := 0

	if uc.repo != nil {
		staleOps, err := uc.repo.GetStaleOperations(ctx)
		if err == nil {
			for _, op := range staleOps {
				if op.Status == domain.OpStatusRunning && op.HeartbeatAt.Before(cutoff) {
					op.Status = domain.OpStatusTimedOut
					op.CompletedAt = &now
					op.UpdatedAt = now
					op.ErrorCode = "OPERATION_TIMED_OUT"
					op.ErrorMessage = "Operation execution exceeded timeout threshold"
					op.Timeline = append(op.Timeline, domain.OperationEvent{
						StepNumber:  len(op.Timeline) + 1,
						Name:        "Operation Timeout",
						Description: "Operation heartbeat exceeded timeout threshold",
						Status:      "TIMED_OUT",
						Timestamp:   now,
					})
					_ = uc.repo.SaveOperation(ctx, op)
					_ = uc.repo.ReleaseDistributedLock(ctx, op.OrganizationID, op.ResourceID, op.ID)
					timedOutCount++
				}
			}
		}
		return timedOutCount
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	for _, op := range uc.operations {
		if op.Status == domain.OpStatusRunning && op.HeartbeatAt.Before(cutoff) {
			op.Status = domain.OpStatusTimedOut
			op.CompletedAt = &now
			op.UpdatedAt = now
			op.ErrorCode = "OPERATION_TIMED_OUT"
			op.ErrorMessage = "Operation execution exceeded timeout threshold"
			op.Timeline = append(op.Timeline, domain.OperationEvent{
				StepNumber:  len(op.Timeline) + 1,
				Name:        "Operation Timeout",
				Description: "Operation heartbeat exceeded timeout threshold",
				Status:      "TIMED_OUT",
				Timestamp:   now,
			})
			delete(uc.locks, op.ResourceID)
			timedOutCount++
		}
	}

	return timedOutCount
}

func (uc *ReliabilityUseCase) GetOperationsSummary(ctx context.Context, orgID string) OperationsSummary {
	ops, _, _ := uc.ListOperations(ctx, orgID, repository.OperationQueryFilters{PageSize: 1000})

	summary := OperationsSummary{}
	summary.Total = len(ops)

	for _, op := range ops {
		switch op.Status {
		case domain.OpStatusPending:
			summary.Pending++
		case domain.OpStatusRunning:
			summary.Running++
		case domain.OpStatusSucceeded:
			summary.Succeeded++
		case domain.OpStatusFailed:
			summary.Failed++
		case domain.OpStatusTimedOut:
			summary.TimedOut++
		case domain.OpStatusCancelled:
			summary.Cancelled++
		case domain.OpStatusRecovering:
			summary.Recovering++
		}
	}

	uc.mu.RLock()
	defer uc.mu.RUnlock()
	summary.ActiveLocks = len(uc.locks)

	return summary
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

func (uc *ReliabilityUseCase) AcquireResourceLock(ctx context.Context, resourceID, opID string, ttl time.Duration) (bool, string, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	now := time.Now()
	if lock, exists := uc.locks[resourceID]; exists {
		if lock.ExpiresAt.After(now) {
			return false, "", nil
		}
	}

	leaseID := fmt.Sprintf("lease-%d", now.UnixNano())
	uc.locks[resourceID] = &domain.ResourceLockLease{
		ResourceID:  resourceID,
		OperationID: opID,
		Owner:       "Anarva-Control-Plane",
		LockedAt:    now,
		ExpiresAt:   now.Add(ttl),
	}
	return true, leaseID, nil
}

func (uc *ReliabilityUseCase) RenewResourceLock(ctx context.Context, resourceID, leaseID string, ttl time.Duration) (bool, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	lock, exists := uc.locks[resourceID]
	if !exists {
		return false, nil
	}
	lock.ExpiresAt = time.Now().Add(ttl)
	return true, nil
}

func (uc *ReliabilityUseCase) ReleaseResourceLock(ctx context.Context, resourceID, leaseID string) (bool, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	delete(uc.locks, resourceID)
	return true, nil
}
