package ha

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/aws"
)

type HAStatus string

const (
	HAStatusEnabled   HAStatus = "HA_ENABLED"
	HAStatusDisabled  HAStatus = "HA_DISABLED"
	HAStatusModifying HAStatus = "HA_MODIFYING"
	HAStatusDegraded  HAStatus = "HA_DEGRADED"
	HAStatusUnknown   HAStatus = "HA_UNKNOWN"
)

type MultiAZConfig struct {
	ResourceID                string    `json:"resourceId"`
	OrganizationID            string    `json:"organizationId"`
	ProjectID                 string    `json:"projectId"`
	MultiAZ                   bool      `json:"multiAz"`
	PrimaryAvailabilityZone   string    `json:"primaryAvailabilityZone"`
	SecondaryAvailabilityZone string    `json:"secondaryAvailabilityZone"`
	HAStatus                  string    `json:"haStatus"` // HA_ENABLED, HA_DISABLED, HA_MODIFYING, HA_DEGRADED
	UpdatedAt                 time.Time `json:"updatedAt"`
}

type FailoverJob struct {
	ID                string     `json:"id"`
	OrganizationID    string     `json:"organizationId"`
	ProjectID         string     `json:"projectId"`
	ResourceID        string     `json:"resourceId"`
	Status            string     `json:"status"` // QUEUED, FAILOVER_INITIATED, PRIMARY_CHANGED, COMPLETED, FAILED
	RequestedBy       string     `json:"requestedBy"`
	PreviousPrimaryAZ string     `json:"previousPrimaryAz"`
	NewPrimaryAZ      string     `json:"newPrimaryAz"`
	FailureReason     string     `json:"failureReason,omitempty"`
	RequestedAt       time.Time  `json:"requestedAt"`
	CompletedAt       *time.Time `json:"completedAt,omitempty"`
}

type HAOrchestrationEngine struct {
	mu            sync.RWMutex
	rdsClient     aws.RDSClient
	configs       map[string]*MultiAZConfig
	failoverJobs  map[string]*FailoverJob
	resourceLocks map[string]bool
}

func NewHAOrchestrationEngine(rdsClient aws.RDSClient) *HAOrchestrationEngine {
	e := &HAOrchestrationEngine{
		rdsClient:     rdsClient,
		configs:       make(map[string]*MultiAZConfig),
		failoverJobs:  make(map[string]*FailoverJob),
		resourceLocks: make(map[string]bool),
	}
	e.seedDefaults()
	return e
}

func (e *HAOrchestrationEngine) seedDefaults() {
	now := time.Now()

	cfg := &MultiAZConfig{
		ResourceID:                "anarva-rds-prod-01",
		OrganizationID:            "org-default",
		ProjectID:                 "proj-default",
		MultiAZ:                   true,
		PrimaryAvailabilityZone:   "ap-south-1a",
		SecondaryAvailabilityZone: "ap-south-1b",
		HAStatus:                  "HA_ENABLED",
		UpdatedAt:                 now,
	}
	e.configs[cfg.ResourceID] = cfg
}

func (e *HAOrchestrationEngine) GetHAConfig(orgID, resourceID string) (*MultiAZConfig, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cfg, exists := e.configs[resourceID]
	if !exists || cfg.OrganizationID != orgID {
		return nil, fmt.Errorf("RESOURCE_NOT_FOUND: RDS instance %s not found", resourceID)
	}
	return cfg, nil
}

func (e *HAOrchestrationEngine) ToggleMultiAZ(ctx context.Context, orgID, projectID, resourceID string, enableMultiAZ bool) (*MultiAZConfig, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cfg, exists := e.configs[resourceID]
	if !exists {
		cfg = &MultiAZConfig{
			ResourceID:     resourceID,
			OrganizationID: orgID,
			ProjectID:      projectID,
		}
		e.configs[resourceID] = cfg
	}

	if cfg.OrganizationID != orgID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: Unauthorized HA modification for resource %s", resourceID)
	}

	if e.resourceLocks[resourceID] {
		return nil, fmt.Errorf("RESOURCE_LOCKED: Resource %s is currently performing another operation", resourceID)
	}
	e.resourceLocks[resourceID] = true
	defer delete(e.resourceLocks, resourceID)

	updatedInfo, err := e.rdsClient.ModifyDBInstanceMultiAZ(ctx, resourceID, enableMultiAZ)
	if err != nil {
		return nil, fmt.Errorf("AWS_MULTI_AZ_TOGGLE_FAILED: %w", err)
	}

	cfg.MultiAZ = updatedInfo.MultiAZ
	cfg.PrimaryAvailabilityZone = updatedInfo.AvailabilityZone
	cfg.SecondaryAvailabilityZone = updatedInfo.SecondaryAvailabilityZone
	if updatedInfo.MultiAZ {
		cfg.HAStatus = "HA_ENABLED"
	} else {
		cfg.HAStatus = "HA_DISABLED"
	}
	cfg.UpdatedAt = time.Now()

	return cfg, nil
}

func (e *HAOrchestrationEngine) TriggerFailover(ctx context.Context, orgID, projectID, resourceID, userEmail string) (*FailoverJob, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cfg, exists := e.configs[resourceID]
	if !exists || cfg.OrganizationID != orgID {
		return nil, fmt.Errorf("RESOURCE_NOT_FOUND: Resource %s not found for organization %s", resourceID, orgID)
	}

	if !cfg.MultiAZ {
		return nil, fmt.Errorf("INVALID_HA_STATE: Cannot trigger failover on single-AZ instance %s", resourceID)
	}

	if e.resourceLocks[resourceID] {
		return nil, fmt.Errorf("RESOURCE_LOCKED: Resource %s is currently locked for failover/modification", resourceID)
	}
	e.resourceLocks[resourceID] = true
	defer delete(e.resourceLocks, resourceID)

	now := time.Now()
	jobID := fmt.Sprintf("job-failover-%d", now.UnixNano()/1e6)
	prevPrimary := cfg.PrimaryAvailabilityZone

	job := &FailoverJob{
		ID:                jobID,
		OrganizationID:    orgID,
		ProjectID:         projectID,
		ResourceID:        resourceID,
		Status:            "FAILOVER_INITIATED",
		RequestedBy:       userEmail,
		PreviousPrimaryAZ: prevPrimary,
		RequestedAt:       now,
	}

	// Trigger real AWS RDS RebootDBInstance with ForceFailover = true
	updatedInfo, err := e.rdsClient.RebootDBInstance(ctx, resourceID, true)
	if err != nil {
		job.Status = "FAILED"
		job.FailureReason = err.Error()
		e.failoverJobs[job.ID] = job
		return job, fmt.Errorf("AWS_FAILOVER_FAILED: %w", err)
	}

	// Failover verification & Primary AZ swap detection
	cfg.PrimaryAvailabilityZone = updatedInfo.AvailabilityZone
	cfg.SecondaryAvailabilityZone = updatedInfo.SecondaryAvailabilityZone
	cfg.UpdatedAt = time.Now()

	completedTime := time.Now()
	job.Status = "COMPLETED"
	job.NewPrimaryAZ = updatedInfo.AvailabilityZone
	job.CompletedAt = &completedTime
	e.failoverJobs[job.ID] = job

	return job, nil
}

func (e *HAOrchestrationEngine) ListFailoverJobs(orgID, resourceID string) []*FailoverJob {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var list []*FailoverJob
	for _, j := range e.failoverJobs {
		if j.OrganizationID == orgID && (resourceID == "" || j.ResourceID == resourceID) {
			list = append(list, j)
		}
	}
	return list
}
