package backup

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/aws"
)

type BackupPolicy struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	ResourceID     string    `json:"resourceId"`
	Enabled        bool      `json:"enabled"`
	RetentionDays  int       `json:"retentionDays"`
	BackupWindow   string    `json:"backupWindow"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type BackupSnapshot struct {
	ID                   string    `json:"id"`
	OrganizationID       string    `json:"organizationId"`
	ProjectID            string    `json:"projectId"`
	ResourceID           string    `json:"resourceId"`
	AWSSnapshotIdentifier string    `json:"awsSnapshotIdentifier"`
	Status               string    `json:"status"` // creating, available, deleting, failed
	AllocatedStorageGB   int       `json:"allocatedStorageGb"`
	Encrypted            bool      `json:"encrypted"`
	Region               string    `json:"region"`
	SnapshotType         string    `json:"snapshotType"` // manual, automated
	CreatedAt            time.Time `json:"createdAt"`
}

type BackupRestoreJob struct {
	ID                string    `json:"id"`
	OrganizationID    string    `json:"organizationId"`
	ProjectID         string    `json:"projectId"`
	SourceResourceID  string    `json:"sourceResourceId"`
	TargetResourceID  string    `json:"targetResourceId"`
	RestoreType       string    `json:"restoreType"` // PITR, SNAPSHOT
	Status            string    `json:"status"`      // QUEUED, VALIDATING, RESTORING, VERIFYING, COMPLETED, FAILED
	RecoveryTimestamp time.Time `json:"recoveryTimestamp"`
	FailureReason     string    `json:"failureReason,omitempty"`
	RequestedAt       time.Time `json:"requestedAt"`
	CompletedAt       *time.Time `json:"completedAt,omitempty"`
}

type BackupRecoveryWindow struct {
	ResourceID             string    `json:"resourceId"`
	EarliestRestorableTime time.Time `json:"earliestRestorableTime"`
	LatestRestorableTime   time.Time `json:"latestRestorableTime"`
}

type BackupOrchestrationEngine struct {
	mu        sync.RWMutex
	rdsClient aws.RDSClient
	policies  map[string]*BackupPolicy
	snapshots map[string]*BackupSnapshot
	jobs      map[string]*BackupRestoreJob
}

func NewBackupOrchestrationEngine(rdsClient aws.RDSClient) *BackupOrchestrationEngine {
	e := &BackupOrchestrationEngine{
		rdsClient: rdsClient,
		policies:  make(map[string]*BackupPolicy),
		snapshots: make(map[string]*BackupSnapshot),
		jobs:      make(map[string]*BackupRestoreJob),
	}
	e.seedDefaults()
	return e
}

func (e *BackupOrchestrationEngine) seedDefaults() {
	now := time.Now()

	policy := &BackupPolicy{
		ID:             "pol-rds-prod-01",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		ResourceID:     "anarva-rds-prod-01",
		Enabled:        true,
		RetentionDays:  7,
		BackupWindow:   "03:00-04:00",
		CreatedAt:      now.Add(-720 * time.Hour),
		UpdatedAt:      now,
	}
	e.policies[policy.ResourceID] = policy

	snap := &BackupSnapshot{
		ID:                   "snap-rds-manual-01",
		OrganizationID:       "org-default",
		ProjectID:            "proj-default",
		ResourceID:           "anarva-rds-prod-01",
		AWSSnapshotIdentifier: "snap-manual-20260814-01",
		Status:               "available",
		AllocatedStorageGB:   20,
		Encrypted:            true,
		Region:               "ap-south-1",
		SnapshotType:         "manual",
		CreatedAt:            now.Add(-24 * time.Hour),
	}
	e.snapshots[snap.ID] = snap
}

func (e *BackupOrchestrationEngine) CreateManualSnapshot(ctx context.Context, orgID, projectID, resourceID, snapshotName string) (*BackupSnapshot, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	awsSnapID := fmt.Sprintf("snap-%s-%d", resourceID, time.Now().Unix())
	awsSnap, err := e.rdsClient.CreateDBSnapshot(ctx, resourceID, awsSnapID)
	if err != nil {
		return nil, fmt.Errorf("AWS_SNAPSHOT_FAILED: %w", err)
	}

	snap := &BackupSnapshot{
		ID:                   fmt.Sprintf("snap-%d", time.Now().UnixNano()/1e6),
		OrganizationID:       orgID,
		ProjectID:            projectID,
		ResourceID:           resourceID,
		AWSSnapshotIdentifier: awsSnap.SnapshotIdentifier,
		Status:               awsSnap.Status,
		AllocatedStorageGB:   awsSnap.AllocatedStorageGB,
		Encrypted:            awsSnap.StorageEncrypted,
		Region:               "ap-south-1",
		SnapshotType:         "manual",
		CreatedAt:            awsSnap.SnapshotCreateTime,
	}

	e.snapshots[snap.ID] = snap
	return snap, nil
}

func (e *BackupOrchestrationEngine) DeleteSnapshot(ctx context.Context, orgID, snapshotID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	snap, exists := e.snapshots[snapshotID]
	if !exists || snap.OrganizationID != orgID {
		return fmt.Errorf("SNAPSHOT_NOT_FOUND: Snapshot %s does not exist", snapshotID)
	}

	err := e.rdsClient.DeleteDBSnapshot(ctx, snap.AWSSnapshotIdentifier)
	if err != nil {
		return fmt.Errorf("AWS_SNAPSHOT_DELETE_FAILED: %w", err)
	}

	delete(e.snapshots, snapshotID)
	return nil
}

func (e *BackupOrchestrationEngine) RestorePointInTime(ctx context.Context, orgID, projectID, sourceResourceID string, targetDBName string, restoreTime time.Time) (*BackupRestoreJob, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	if targetDBName == "" {
		targetDBName = fmt.Sprintf("%s-restore-%s", sourceResourceID, restoreTime.Format("20060102-1504"))
	}

	jobID := fmt.Sprintf("job-pitr-%d", now.UnixNano()/1e6)
	job := &BackupRestoreJob{
		ID:                jobID,
		OrganizationID:    orgID,
		ProjectID:         projectID,
		SourceResourceID:  sourceResourceID,
		TargetResourceID:  targetDBName,
		RestoreType:       "PITR",
		Status:            "RESTORING",
		RecoveryTimestamp: restoreTime,
		RequestedAt:       now,
	}

	// Trigger real AWS RDS RestoreDBInstanceToPointInTime API
	restoredInfo, err := e.rdsClient.RestoreDBInstanceToPointInTime(ctx, sourceResourceID, targetDBName, restoreTime)
	if err != nil {
		job.Status = "FAILED"
		job.FailureReason = err.Error()
		e.jobs[job.ID] = job
		return job, fmt.Errorf("PITR_RESTORE_FAILED: %w", err)
	}

	// Verify encryption safety - Encryption cannot be downgraded
	if !restoredInfo.StorageEncrypted {
		job.Status = "FAILED"
		job.FailureReason = "ENCRYPTION_DOWNGRADE_BLOCKED: Restored database failed encryption validation"
		e.jobs[job.ID] = job
		return job, fmt.Errorf("ENCRYPTION_DOWNGRADE_BLOCKED")
	}

	// Mark Job Completed
	completedTime := time.Now()
	job.Status = "COMPLETED"
	job.CompletedAt = &completedTime
	e.jobs[job.ID] = job

	return job, nil
}

func (e *BackupOrchestrationEngine) GetRecoveryWindow(ctx context.Context, resourceID string) (*BackupRecoveryWindow, error) {
	windowInfo, err := e.rdsClient.GetRecoveryWindow(ctx, resourceID)
	if err != nil {
		return nil, err
	}

	return &BackupRecoveryWindow{
		ResourceID:             resourceID,
		EarliestRestorableTime: windowInfo.EarliestRestorableTime,
		LatestRestorableTime:   windowInfo.LatestRestorableTime,
	}, nil
}

func (e *BackupOrchestrationEngine) ListSnapshots(orgID, resourceID string) []*BackupSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var list []*BackupSnapshot
	for _, s := range e.snapshots {
		if s.OrganizationID == orgID && (resourceID == "" || s.ResourceID == resourceID) {
			list = append(list, s)
		}
	}
	return list
}

func (e *BackupOrchestrationEngine) ListRestoreJobs(orgID string) []*BackupRestoreJob {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var list []*BackupRestoreJob
	for _, j := range e.jobs {
		if j.OrganizationID == orgID {
			list = append(list, j)
		}
	}
	return list
}

func (e *BackupOrchestrationEngine) GetBackupPolicy(orgID, resourceID string) (*BackupPolicy, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pol, exists := e.policies[resourceID]
	if exists && pol.OrganizationID == orgID {
		return pol, true
	}
	return nil, false
}
