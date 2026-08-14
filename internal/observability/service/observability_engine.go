package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/observability/domain"
)

type ResourceObservabilityEngine struct {
	mu                sync.RWMutex
	observations      map[string]*domain.ResourceObservation
	staleThreshold    time.Duration
	workerCancel      context.CancelFunc
	workerRunning     bool
	observationEvents []*domain.LogRecord
}

func NewResourceObservabilityEngine() *ResourceObservabilityEngine {
	e := &ResourceObservabilityEngine{
		observations:      make(map[string]*domain.ResourceObservation),
		staleThreshold:    90 * time.Second,
		observationEvents: make([]*domain.LogRecord, 0),
	}
	e.seedDefaultObservations()
	return e
}

func (e *ResourceObservabilityEngine) seedDefaultObservations() {
	now := time.Now()

	// Seed control-plane resources across EC2, RDS, and S3
	defaults := []*domain.ResourceObservation{
		{
			ResourceID:            "res-ec2-worker-01",
			OrganizationID:        "org-default",
			ProjectID:             "proj-default",
			ResourceName:          "ace-worker-node-01",
			Provider:              "AWS",
			ResourceType:          "EC2",
			ProviderResourceID:    "i-0a8f9c1b2d3e4f5a6",
			Region:                "us-east-1",
			DesiredState:          "RUNNING",
			ObservedState:         "RUNNING",
			HealthState:           domain.HealthHealthy,
			DriftStatus:           domain.DriftInSync,
			LastObservedAt:        now.Add(-12 * time.Second),
			ObservationDurationMs: 45,
			ObservedAttributes: map[string]string{
				"InstanceType":     "t3.medium",
				"AvailabilityZone": "us-east-1a",
				"AnarvaManaged":    "true",
			},
		},
		{
			ResourceID:            "res-rds-postgres-01",
			OrganizationID:        "org-default",
			ProjectID:             "proj-default",
			ResourceName:          "anarva-postgres-production",
			Provider:              "AWS",
			ResourceType:          "RDS_POSTGRESQL",
			ProviderResourceID:    "anarva-rds-res-rds-postgres-01",
			Region:                "us-east-1",
			DesiredState:          "AVAILABLE",
			ObservedState:         "AVAILABLE",
			HealthState:           domain.HealthHealthy,
			DriftStatus:           domain.DriftInSync,
			LastObservedAt:        now.Add(-8 * time.Second),
			ObservationDurationMs: 62,
			ObservedAttributes: map[string]string{
				"Engine":             "postgres",
				"EngineVersion":      "15.4",
				"InstanceClass":      "db.t3.micro",
				"PubliclyAccessible": "false",
				"StorageEncrypted":   "true",
			},
		},
		{
			ResourceID:            "res-s3-assets-01",
			OrganizationID:        "org-default",
			ProjectID:             "proj-default",
			ResourceName:          "anarva-production-media-assets",
			Provider:              "AWS",
			ResourceType:          "S3_BUCKET",
			ProviderResourceID:    "anarva-s3-res-s3-assets-01",
			Region:                "us-east-1",
			DesiredState:          "ACTIVE",
			ObservedState:         "ACTIVE",
			HealthState:           domain.HealthHealthy,
			DriftStatus:           domain.DriftInSync,
			LastObservedAt:        now.Add(-14 * time.Second),
			ObservationDurationMs: 38,
			ObservedAttributes: map[string]string{
				"PublicAccessBlock": "true",
				"EncryptionMode":    "SSE-S3",
			},
		},
	}

	for _, obs := range defaults {
		e.observations[obs.ResourceID] = obs
	}
}

func (e *ResourceObservabilityEngine) RecordObservation(ctx context.Context, obs *domain.ResourceObservation) *domain.ResourceObservation {
	e.mu.Lock()
	defer e.mu.Unlock()

	startTime := time.Now()
	obs.LastObservedAt = startTime

	// Drift Detection & Health Evaluation Logic
	evaluateResourceHealthAndDrift(obs)

	if obs.ObservationDurationMs == 0 {
		obs.ObservationDurationMs = time.Since(startTime).Milliseconds() + 25
	}

	e.observations[obs.ResourceID] = obs
	return obs
}

func (e *ResourceObservabilityEngine) GetObservation(resourceID string) (*domain.ResourceObservation, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	obs, exists := e.observations[resourceID]
	if !exists {
		return nil, false
	}

	// Copy and evaluate staleness
	copyObs := *obs
	if time.Since(copyObs.LastObservedAt) > e.staleThreshold {
		copyObs.IsStale = true
	}
	return &copyObs, true
}

func (e *ResourceObservabilityEngine) ListObservations(orgID, projectID, resourceType, healthState, driftStatus string) []*domain.ResourceObservation {
	e.mu.RLock()
	defer e.mu.RUnlock()

	now := time.Now()
	var filtered []*domain.ResourceObservation

	for _, obs := range e.observations {
		if orgID != "" && obs.OrganizationID != orgID {
			continue
		}
		if projectID != "" && obs.ProjectID != projectID {
			continue
		}
		if resourceType != "" && !strings.EqualFold(obs.ResourceType, resourceType) {
			continue
		}
		if healthState != "" && !strings.EqualFold(string(obs.HealthState), healthState) {
			continue
		}
		if driftStatus != "" && !strings.EqualFold(string(obs.DriftStatus), driftStatus) {
			continue
		}

		copyObs := *obs
		if now.Sub(copyObs.LastObservedAt) > e.staleThreshold {
			copyObs.IsStale = true
		}
		filtered = append(filtered, &copyObs)
	}

	return filtered
}

func (e *ResourceObservabilityEngine) GetControlPlaneSummary(orgID string) map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	now := time.Now()
	total := 0
	healthy := 0
	degraded := 0
	provisioning := 0
	drifted := 0
	failed := 0
	externallyDeleted := 0

	for _, obs := range e.observations {
		if orgID != "" && obs.OrganizationID != orgID {
			continue
		}
		total++
		switch obs.HealthState {
		case domain.HealthHealthy:
			healthy++
		case domain.HealthDegraded, domain.HealthUnavailable:
			degraded++
		case domain.HealthProvisioning, domain.HealthUpdating:
			provisioning++
		case domain.HealthDrifted:
			drifted++
		case domain.HealthStopped, domain.HealthUnknown:
			failed++
		case domain.HealthExternallyDeleted:
			externallyDeleted++
		}
	}

	return map[string]interface{}{
		"totalResources":     total,
		"healthy":            healthy,
		"degraded":           degraded,
		"provisioning":       provisioning,
		"drifted":            drifted,
		"failed":             failed,
		"externallyDeleted":  externallyDeleted,
		"providerStatus":     "CONNECTED",
		"lastObservedTime":   now.Format(time.RFC3339),
	}
}

func evaluateResourceHealthAndDrift(obs *domain.ResourceObservation) {
	// Step 1: Missing Resource Detection
	if obs.ObservedState == "EXTERNALLY_DELETED" || obs.ObservedState == "MISSING" {
		obs.HealthState = domain.HealthExternallyDeleted
		obs.DriftStatus = domain.DriftMissingResource
		obs.DriftDetails = "Resource was deleted externally on cloud provider outside Anarva control plane"
		return
	}

	// Step 2: EC2 Evaluation
	if obs.ResourceType == "EC2" {
		if obs.DesiredState == "RUNNING" && obs.ObservedState == "STOPPED" {
			obs.HealthState = domain.HealthStopped
			obs.DriftStatus = domain.DriftStateDrift
			obs.DriftDetails = "Desired state is RUNNING, but observed AWS EC2 state is STOPPED"
			return
		}
		if obs.ObservedState == "RUNNING" {
			obs.HealthState = domain.HealthHealthy
			obs.DriftStatus = domain.DriftInSync
			return
		}
	}

	// Step 3: RDS PostgreSQL Evaluation
	if obs.ResourceType == "RDS_POSTGRESQL" {
		if obs.ObservedAttributes != nil && obs.ObservedAttributes["PubliclyAccessible"] == "true" {
			obs.HealthState = domain.HealthDrifted
			obs.DriftStatus = domain.DriftSecurityDrift
			obs.DriftDetails = "SECURITY DRIFT: RDS database instance has PubliclyAccessible=true enabled on AWS"
			return
		}
		if obs.ObservedState == "AVAILABLE" {
			obs.HealthState = domain.HealthHealthy
			obs.DriftStatus = domain.DriftInSync
			return
		}
	}

	// Step 4: S3 Bucket Evaluation
	if obs.ResourceType == "S3_BUCKET" {
		if obs.ObservedAttributes != nil && (obs.ObservedAttributes["PublicAccessBlock"] == "false" || obs.ObservedAttributes["EncryptionMode"] == "NONE") {
			obs.HealthState = domain.HealthDrifted
			obs.DriftStatus = domain.DriftSecurityDrift
			obs.DriftDetails = "SECURITY DRIFT: S3 Bucket lacks Block Public Access or server-side encryption"
			return
		}
		if obs.ObservedState == "ACTIVE" {
			obs.HealthState = domain.HealthHealthy
			obs.DriftStatus = domain.DriftInSync
			return
		}
	}

	// Default Fallback
	if obs.HealthState == "" {
		obs.HealthState = domain.HealthHealthy
		obs.DriftStatus = domain.DriftInSync
	}
}
