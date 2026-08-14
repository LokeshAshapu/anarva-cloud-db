package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/observability/domain"
)

func TestObservabilityEngine_EC2HealthAndDrift(t *testing.T) {
	engine := NewResourceObservabilityEngine()

	// Healthy EC2
	healthyEC2 := &domain.ResourceObservation{
		ResourceID:     "res-ec2-test-01",
		OrganizationID: "org-alpha",
		ProjectID:      "proj-1",
		ResourceName:   "test-ec2",
		Provider:       "AWS",
		ResourceType:   "EC2",
		DesiredState:   "RUNNING",
		ObservedState:  "RUNNING",
	}
	recorded := engine.RecordObservation(context.Background(), healthyEC2)
	assert.Equal(t, domain.HealthHealthy, recorded.HealthState)
	assert.Equal(t, domain.DriftInSync, recorded.DriftStatus)

	// Stopped EC2 State Drift
	stoppedEC2 := &domain.ResourceObservation{
		ResourceID:     "res-ec2-test-02",
		OrganizationID: "org-alpha",
		ProjectID:      "proj-1",
		ResourceName:   "drifted-ec2",
		Provider:       "AWS",
		ResourceType:   "EC2",
		DesiredState:   "RUNNING",
		ObservedState:  "STOPPED",
	}
	drifted := engine.RecordObservation(context.Background(), stoppedEC2)
	assert.Equal(t, domain.HealthStopped, drifted.HealthState)
	assert.Equal(t, domain.DriftStateDrift, drifted.DriftStatus)
}

func TestObservabilityEngine_RDSHealthAndSecurityDrift(t *testing.T) {
	engine := NewResourceObservabilityEngine()

	// Publicly Accessible RDS Security Drift
	publicRDS := &domain.ResourceObservation{
		ResourceID:     "res-rds-public",
		OrganizationID: "org-alpha",
		ProjectID:      "proj-1",
		ResourceName:   "insecure-rds",
		Provider:       "AWS",
		ResourceType:   "RDS_POSTGRESQL",
		DesiredState:   "AVAILABLE",
		ObservedState:  "AVAILABLE",
		ObservedAttributes: map[string]string{
			"PubliclyAccessible": "true",
		},
	}

	recorded := engine.RecordObservation(context.Background(), publicRDS)
	assert.Equal(t, domain.HealthDrifted, recorded.HealthState)
	assert.Equal(t, domain.DriftSecurityDrift, recorded.DriftStatus)
	assert.Contains(t, recorded.DriftDetails, "SECURITY DRIFT")
}

func TestObservabilityEngine_S3HealthAndEncryptionDrift(t *testing.T) {
	engine := NewResourceObservabilityEngine()

	// Unencrypted / Public S3 Security Drift
	publicS3 := &domain.ResourceObservation{
		ResourceID:     "res-s3-public",
		OrganizationID: "org-alpha",
		ProjectID:      "proj-1",
		ResourceName:   "insecure-s3",
		Provider:       "AWS",
		ResourceType:   "S3_BUCKET",
		DesiredState:   "ACTIVE",
		ObservedState:  "ACTIVE",
		ObservedAttributes: map[string]string{
			"PublicAccessBlock": "false",
			"EncryptionMode":    "NONE",
		},
	}

	recorded := engine.RecordObservation(context.Background(), publicS3)
	assert.Equal(t, domain.HealthDrifted, recorded.HealthState)
	assert.Equal(t, domain.DriftSecurityDrift, recorded.DriftStatus)
}

func TestObservabilityEngine_ExternallyDeletedResource(t *testing.T) {
	engine := NewResourceObservabilityEngine()

	missingRes := &domain.ResourceObservation{
		ResourceID:     "res-deleted-externally",
		OrganizationID: "org-alpha",
		ProjectID:      "proj-1",
		ResourceName:   "gone-instance",
		Provider:       "AWS",
		ResourceType:   "EC2",
		DesiredState:   "RUNNING",
		ObservedState:  "EXTERNALLY_DELETED",
	}

	recorded := engine.RecordObservation(context.Background(), missingRes)
	assert.Equal(t, domain.HealthExternallyDeleted, recorded.HealthState)
	assert.Equal(t, domain.DriftMissingResource, recorded.DriftStatus)
}

func TestObservabilityEngine_StaleObservation(t *testing.T) {
	engine := NewResourceObservabilityEngine()

	oldObs := &domain.ResourceObservation{
		ResourceID:     "res-stale-01",
		OrganizationID: "org-alpha",
		LastObservedAt: time.Now().Add(-5 * time.Minute),
	}
	engine.RecordObservation(context.Background(), oldObs)

	// Override timestamp to simulate stale observation
	engine.mu.Lock()
	engine.observations["res-stale-01"].LastObservedAt = time.Now().Add(-5 * time.Minute)
	engine.mu.Unlock()

	retrieved, exists := engine.GetObservation("res-stale-01")
	require.True(t, exists)
	assert.True(t, retrieved.IsStale)
}

func TestObservabilityEngine_TenantIsolation(t *testing.T) {
	engine := NewResourceObservabilityEngine()

	obs1 := &domain.ResourceObservation{
		ResourceID:     "res-org-a",
		OrganizationID: "org-a",
		ResourceType:   "EC2",
	}
	obs2 := &domain.ResourceObservation{
		ResourceID:     "res-org-b",
		OrganizationID: "org-b",
		ResourceType:   "RDS_POSTGRESQL",
	}

	engine.RecordObservation(context.Background(), obs1)
	engine.RecordObservation(context.Background(), obs2)

	listOrgA := engine.ListObservations("org-a", "", "", "", "")
	assert.Len(t, listOrgA, 1)
	assert.Equal(t, "res-org-a", listOrgA[0].ResourceID)
}
