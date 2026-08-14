package backup

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/aws"
)

func TestBackupEngine_CreateManualSnapshot(t *testing.T) {
	mockRDS := aws.NewMockRDSClient(true)
	engine := NewBackupOrchestrationEngine(mockRDS)

	snap, err := engine.CreateManualSnapshot(context.Background(), "org-default", "proj-default", "anarva-rds-prod-01", "pre-migration-snap")
	require.NoError(t, err)
	assert.NotNil(t, snap)
	assert.Equal(t, "anarva-rds-prod-01", snap.ResourceID)
	assert.Equal(t, "available", snap.Status)
	assert.True(t, snap.Encrypted)
	assert.Equal(t, "manual", snap.SnapshotType)
}

func TestBackupEngine_DeleteSnapshot(t *testing.T) {
	mockRDS := aws.NewMockRDSClient(true)
	engine := NewBackupOrchestrationEngine(mockRDS)

	snap, err := engine.CreateManualSnapshot(context.Background(), "org-default", "proj-default", "anarva-rds-prod-01", "snap-to-delete")
	require.NoError(t, err)

	err = engine.DeleteSnapshot(context.Background(), "org-default", snap.ID)
	require.NoError(t, err)

	// Verify snapshot is deleted
	snaps := engine.ListSnapshots("org-default", "anarva-rds-prod-01")
	for _, s := range snaps {
		assert.NotEqual(t, snap.ID, s.ID)
	}
}

func TestBackupEngine_PointInTimeRecovery_CreatesNewResource(t *testing.T) {
	mockRDS := aws.NewMockRDSClient(true)
	engine := NewBackupOrchestrationEngine(mockRDS)

	restoreTime := time.Now().Add(-1 * time.Hour)
	targetName := "anarva-rds-prod-01-pitr-restored"

	job, err := engine.RestorePointInTime(context.Background(), "org-default", "proj-default", "anarva-rds-prod-01", targetName, restoreTime)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "COMPLETED", job.Status)
	assert.Equal(t, targetName, job.TargetResourceID)

	// Verify original DB still exists (NEVER overwritten)
	origInfo, err := mockRDS.DescribeDBInstances(context.Background(), "anarva-rds-prod-01")
	require.NoError(t, err)
	assert.Equal(t, "available", origInfo.Status)

	// Verify restored DB exists with PubliclyAccessible = false
	restoredInfo, err := mockRDS.DescribeDBInstances(context.Background(), targetName)
	require.NoError(t, err)
	assert.Equal(t, "available", restoredInfo.Status)
	assert.False(t, restoredInfo.PubliclyAccessible)
	assert.True(t, restoredInfo.StorageEncrypted)
}

func TestBackupEngine_TenantIsolation(t *testing.T) {
	mockRDS := aws.NewMockRDSClient(true)
	engine := NewBackupOrchestrationEngine(mockRDS)

	snap, err := engine.CreateManualSnapshot(context.Background(), "org-a", "proj-a", "anarva-rds-prod-01", "snap-org-a")
	require.NoError(t, err)

	// Org B attempting to delete Org A snapshot fails
	err = engine.DeleteSnapshot(context.Background(), "org-b", snap.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SNAPSHOT_NOT_FOUND")
}
