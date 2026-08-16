package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/anarva-cloud/anarva-cloud-db/internal/reliability/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type ReliabilityRepository struct {
	db *gorm.DB
}

func NewReliabilityRepository(db *gorm.DB) *ReliabilityRepository {
	return &ReliabilityRepository{db: db}
}

// 1. Persistent Operations
func (r *ReliabilityRepository) SaveOperation(ctx context.Context, op *domain.AnarvaOperation) error {
	if op == nil || op.ID == "" {
		return fmt.Errorf("INVALID_OPERATION: Cannot save empty operation")
	}

	if len(op.Timeline) > 0 {
		b, _ := json.Marshal(op.Timeline)
		op.TimelineJSON = string(b)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing domain.AnarvaOperation
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", op.ID).First(&existing).Error
		if err == nil {
			// Validate state transition
			if !domain.IsValidStateTransition(existing.Status, op.Status) {
				return appErrors.New(appErrors.CodeInvalidInput, fmt.Sprintf("INVALID_STATE_TRANSITION: Cannot transition operation %s from %s to %s", op.ID, existing.Status, op.Status))
			}
			return tx.Model(&existing).Updates(map[string]interface{}{
				"status":             op.Status,
				"progress":           op.Progress,
				"updated_at":         time.Now(),
				"completed_at":       op.CompletedAt,
				"heartbeat_at":       time.Now(),
				"lease_expires_at":   op.LeaseExpiresAt,
				"retry_count":        op.RetryCount,
				"error_code":         op.ErrorCode,
				"error_message":      op.ErrorMessage,
				"actor_id":           op.ActorID,
				"recovery_attempted": op.RecoveryAttempted,
				"recovery_attempt":   op.RecoveryAttempt,
				"recovery_status":    op.RecoveryStatus,
				"recovery_reason":    op.RecoveryReason,
				"timeline_json":      op.TimelineJSON,
			}).Error
		} else if err == gorm.ErrRecordNotFound {
			if op.CreatedAt.IsZero() {
				op.CreatedAt = time.Now()
			}
			op.UpdatedAt = time.Now()
			op.HeartbeatAt = time.Now()
			if op.LeaseExpiresAt.IsZero() {
				op.LeaseExpiresAt = time.Now().Add(5 * time.Minute)
			}
			return tx.Create(op).Error
		}
		return err
	})
}

func (r *ReliabilityRepository) GetOperation(ctx context.Context, orgID, opID string) (*domain.AnarvaOperation, error) {
	var op domain.AnarvaOperation
	query := r.db.WithContext(ctx).Where("id = ?", opID)
	if orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	if err := query.First(&op).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, appErrors.New(appErrors.CodeNotFound, "Operation not found")
		}
		return nil, err
	}

	if op.TimelineJSON != "" {
		_ = json.Unmarshal([]byte(op.TimelineJSON), &op.Timeline)
	}
	if op.RecoveryAttempted {
		op.Recovery = &domain.RecoveryInfo{
			Attempted: op.RecoveryAttempted,
			Attempt:   op.RecoveryAttempt,
			Status:    op.RecoveryStatus,
			Reason:    op.RecoveryReason,
		}
	}
	return &op, nil
}

type OperationQueryFilters struct {
	Status        string
	ResourceID    string
	OperationType string
	ProjectID     string
	CreatedAfter  time.Time
	CreatedBefore time.Time
	Page          int
	PageSize      int
}

func (r *ReliabilityRepository) ListOperations(ctx context.Context, orgID string, filters OperationQueryFilters) ([]*domain.AnarvaOperation, int64, error) {
	var ops []*domain.AnarvaOperation
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AnarvaOperation{})

	if orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	if filters.ProjectID != "" {
		query = query.Where("project_id = ?", filters.ProjectID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.ResourceID != "" {
		query = query.Where("resource_id = ?", filters.ResourceID)
	}
	if filters.OperationType != "" {
		query = query.Where("type = ?", filters.OperationType)
	}
	if !filters.CreatedAfter.IsZero() {
		query = query.Where("created_at >= ?", filters.CreatedAfter)
	}
	if !filters.CreatedBefore.IsZero() {
		query = query.Where("created_at <= ?", filters.CreatedBefore)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filters.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	err := query.Order("created_at desc").Limit(pageSize).Offset(offset).Find(&ops).Error
	if err != nil {
		return nil, 0, err
	}

	for _, op := range ops {
		if op.TimelineJSON != "" {
			_ = json.Unmarshal([]byte(op.TimelineJSON), &op.Timeline)
		}
		if op.RecoveryAttempted {
			op.Recovery = &domain.RecoveryInfo{
				Attempted: op.RecoveryAttempted,
				Attempt:   op.RecoveryAttempt,
				Status:    op.RecoveryStatus,
				Reason:    op.RecoveryReason,
			}
		}
	}

	return ops, total, nil
}

func (r *ReliabilityRepository) GetStaleOperations(ctx context.Context) ([]*domain.AnarvaOperation, error) {
	var ops []*domain.AnarvaOperation
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("status IN ? AND lease_expires_at < ?", []domain.OperationStatus{domain.OpStatusRunning, domain.OpStatusRecovering}, now).
		Find(&ops).Error
	if err != nil {
		return nil, err
	}
	for _, op := range ops {
		if op.TimelineJSON != "" {
			_ = json.Unmarshal([]byte(op.TimelineJSON), &op.Timeline)
		}
	}
	return ops, nil
}

// 2. Persistent Distributed Resource Locks
func (r *ReliabilityRepository) AcquireDistributedLock(ctx context.Context, lock *domain.ResourceLockLease) error {
	if lock.ResourceID == "" || lock.OperationID == "" {
		return appErrors.New(appErrors.CodeInvalidInput, "ResourceID and OperationID required for lock acquisition")
	}

	now := time.Now()
	if lock.ExpiresAt.IsZero() {
		lock.ExpiresAt = now.Add(5 * time.Minute)
	}
	lock.AcquiredAt = now
	lock.LockedAt = now
	lock.HeartbeatAt = now
	lock.CreatedAt = now
	lock.UpdatedAt = now

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing domain.ResourceLockLease
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("resource_id = ?", lock.ResourceID).First(&existing).Error
		if err == nil {
			// Lock exists. Check if active and owned by another operation
			if existing.ExpiresAt.After(now) && existing.OperationID != lock.OperationID {
				return appErrors.New(appErrors.CodeConflict, fmt.Sprintf("RESOURCE_LOCKED: Resource %s is currently locked by operation %s", lock.ResourceID, existing.OperationID))
			}
			// Expired or same operation: update lock lease
			return tx.Model(&existing).Updates(map[string]interface{}{
				"organization_id": lock.OrganizationID,
				"project_id":      lock.ProjectID,
				"operation_id":    lock.OperationID,
				"holder_id":       lock.HolderID,
				"owner":           lock.Owner,
				"expires_at":      lock.ExpiresAt,
				"heartbeat_at":    now,
				"updated_at":       now,
			}).Error
		} else if err == gorm.ErrRecordNotFound {
			return tx.Create(lock).Error
		}
		return err
	})
}

func (r *ReliabilityRepository) RenewLockLease(ctx context.Context, resourceID, opID string, extension time.Duration) error {
	now := time.Now()
	newExpiry := now.Add(extension)
	result := r.db.WithContext(ctx).Model(&domain.ResourceLockLease{}).
		Where("resource_id = ? AND operation_id = ?", resourceID, opID).
		Updates(map[string]interface{}{
			"expires_at":   newExpiry,
			"heartbeat_at": now,
			"updated_at":   now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "Active lock lease not found or already released")
	}
	return nil
}

func (r *ReliabilityRepository) ReleaseDistributedLock(ctx context.Context, orgID, resourceID, opID string) error {
	query := r.db.WithContext(ctx).Where("resource_id = ?", resourceID)
	if opID != "" {
		query = query.Where("operation_id = ?", opID)
	}
	if orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	return query.Delete(&domain.ResourceLockLease{}).Error
}

// 3. Persistent Idempotency
func (r *ReliabilityRepository) SaveIdempotencyRecord(ctx context.Context, rec *domain.IdempotencyRecord) error {
	if rec.Key == "" || rec.OrganizationID == "" || rec.ProjectID == "" {
		return appErrors.New(appErrors.CodeInvalidInput, "Idempotency Key, OrganizationID, and ProjectID required")
	}

	if rec.KeyHash == "" {
		rec.KeyHash = domain.HashRequestPayload(rec.Key)
	}
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = time.Now()
	}
	if rec.ExpiresAt.IsZero() {
		rec.ExpiresAt = rec.CreatedAt.Add(24 * time.Hour)
	}
	if rec.ID == "" {
		rec.ID = fmt.Sprintf("idemp-%d", time.Now().UnixNano()/1e6)
	}

	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *ReliabilityRepository) GetIdempotencyRecord(ctx context.Context, orgID, projID, key string) (*domain.IdempotencyRecord, error) {
	keyHash := domain.HashRequestPayload(key)
	var rec domain.IdempotencyRecord
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND project_id = ? AND key_hash = ?", orgID, projID, keyHash).
		First(&rec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

// 4. Persistent Tenant Quota
func (r *ReliabilityRepository) GetOrCreateQuota(ctx context.Context, orgID, projID string) (*domain.TenantQuota, error) {
	var q domain.TenantQuota
	err := r.db.WithContext(ctx).Where("organization_id = ? AND project_id = ?", orgID, projID).First(&q).Error
	if err == nil {
		return &q, nil
	}
	if err == gorm.ErrRecordNotFound {
		now := time.Now()
		q = domain.TenantQuota{
			ID:               fmt.Sprintf("quota-%s-%s", orgID, projID),
			OrganizationID:   orgID,
			ProjectID:        projID,
			MaxACU:           100.0,
			CurrentACU:       0.0,
			MaxDatabases:     10,
			CurrentDbs:       0,
			MaxStorageGB:     1000,
			CurrentStorageGB: 0,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if createErr := r.db.WithContext(ctx).Create(&q).Error; createErr != nil {
			return nil, createErr
		}
		return &q, nil
	}
	return nil, err
}

func (r *ReliabilityRepository) ReserveQuota(ctx context.Context, orgID, projID string, reqAcu float64, reqDbs int, reqStorageGB int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var q domain.TenantQuota
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("organization_id = ? AND project_id = ?", orgID, projID).First(&q).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				now := time.Now()
				q = domain.TenantQuota{
					ID:               fmt.Sprintf("quota-%s-%s", orgID, projID),
					OrganizationID:   orgID,
					ProjectID:        projID,
					MaxACU:           100.0,
					CurrentACU:       0.0,
					MaxDatabases:     10,
					CurrentDbs:       0,
					MaxStorageGB:     1000,
					CurrentStorageGB: 0,
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				if createErr := tx.Create(&q).Error; createErr != nil {
					return createErr
				}
			} else {
				return err
			}
		}

		if q.CurrentACU+reqAcu > q.MaxACU {
			return appErrors.New(appErrors.CodeQuotaExceeded, fmt.Sprintf("QUOTA_EXCEEDED: Requested ACU (%.1f) exceeds project quota limit (%.1f / %.1f)", reqAcu, q.CurrentACU+reqAcu, q.MaxACU))
		}
		if q.CurrentDbs+reqDbs > q.MaxDatabases {
			return appErrors.New(appErrors.CodeQuotaExceeded, fmt.Sprintf("QUOTA_EXCEEDED: Requested database count (%d) exceeds project quota limit (%d / %d)", reqDbs, q.CurrentDbs+reqDbs, q.MaxDatabases))
		}

		return tx.Model(&q).Updates(map[string]interface{}{
			"current_acu":        q.CurrentACU + reqAcu,
			"current_dbs":        q.CurrentDbs + reqDbs,
			"current_storage_gb": q.CurrentStorageGB + reqStorageGB,
			"updated_at":         time.Now(),
		}).Error
	})
}

// 5. Persistent Audit Logging
func (r *ReliabilityRepository) SaveAuditEvent(ctx context.Context, evt *domain.AnarvaAuditEvent) error {
	if evt.ID == "" {
		evt.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano()/1e6)
	}
	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now()
	}
	if len(evt.Metadata) > 0 {
		b, _ := json.Marshal(evt.Metadata)
		evt.MetadataJSON = string(b)
	}
	return r.db.WithContext(ctx).Create(evt).Error
}

func (r *ReliabilityRepository) ListAuditEvents(ctx context.Context, orgID, projID string) ([]*domain.AnarvaAuditEvent, error) {
	var events []*domain.AnarvaAuditEvent
	query := r.db.WithContext(ctx).Order("timestamp desc")
	if orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	if projID != "" {
		query = query.Where("project_id = ?", projID)
	}
	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	for _, evt := range events {
		if evt.MetadataJSON != "" {
			_ = json.Unmarshal([]byte(evt.MetadataJSON), &evt.Metadata)
		}
	}
	return events, nil
}
