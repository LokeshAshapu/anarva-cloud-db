package job

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type JobStatus string

const (
	StatusQueued    JobStatus = "QUEUED"
	StatusRunning   JobStatus = "RUNNING"
	StatusSucceeded JobStatus = "SUCCEEDED"
	StatusFailed    JobStatus = "FAILED"
	StatusCancelled JobStatus = "CANCELLED"
	StatusRetrying  JobStatus = "RETRYING"
)

type JobType string

const (
	JobCreateDatabase    JobType = "CREATE_DATABASE"
	JobBackupDatabase    JobType = "BACKUP_DATABASE"
	JobRestoreDatabase   JobType = "RESTORE_DATABASE"
	JobCreateCompute     JobType = "CREATE_COMPUTE"
	JobCreateNetwork     JobType = "CREATE_NETWORK"
	JobCreateStorage     JobType = "CREATE_STORAGE"
	JobCreateLoadBalancer JobType = "CREATE_LOAD_BALANCER"
)

type Job struct {
	ID             string     `json:"id"`
	Type           JobType    `json:"type"`
	ResourceID     string     `json:"resourceId"`
	OrganizationID string     `json:"organizationId"`
	ProjectID      string     `json:"projectId"`
	Status         JobStatus  `json:"status"`
	Progress       string     `json:"progress"` // e.g. "RUNNING" (no fake percentage unless genuinely known)
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	Error          string     `json:"error,omitempty"`
	RetryCount     int        `json:"retryCount"`
	MaxRetries     int        `json:"maxRetries"`
	IdempotencyKey string     `json:"idempotencyKey,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type JobManager struct {
	mu           sync.RWMutex
	jobs         map[string]*Job
	idempotency map[string]string // key -> jobId
}

func NewJobManager() *JobManager {
	jm := &JobManager{
		jobs:         make(map[string]*Job),
		idempotency: make(map[string]string),
	}
	jm.seedSampleJobs()
	return jm
}

func (jm *JobManager) seedSampleJobs() {
	now := time.Now()
	sampleJobs := []*Job{
		{
			ID:             "job-101",
			Type:           JobCreateDatabase,
			ResourceID:     "res-db-prod-1",
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Status:         StatusSucceeded,
			Progress:       "COMPLETED",
			StartedAt:      &now,
			CompletedAt:    &now,
			MaxRetries:     3,
			CreatedAt:      now.Add(-2 * time.Hour),
			UpdatedAt:      now,
		},
		{
			ID:             "job-102",
			Type:           JobCreateCompute,
			ResourceID:     "res-ace-worker-1",
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			Status:         StatusSucceeded,
			Progress:       "COMPLETED",
			StartedAt:      &now,
			CompletedAt:    &now,
			MaxRetries:     3,
			CreatedAt:      now.Add(-1 * time.Hour),
			UpdatedAt:      now,
		},
	}
	for _, j := range sampleJobs {
		jm.jobs[j.ID] = j
	}
}

func (jm *JobManager) Enqueue(ctx context.Context, jType JobType, resourceID, orgID, projectID, idempotencyKey string) (*Job, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	// Idempotency check: if request has key and key exists, return original job
	if idempotencyKey != "" {
		if existingID, ok := jm.idempotency[idempotencyKey]; ok {
			if existingJob, exists := jm.jobs[existingID]; exists {
				return existingJob, nil
			}
		}
	}

	now := time.Now()
	jobID := fmt.Sprintf("job-%d", now.UnixNano()/1e6)
	j := &Job{
		ID:             jobID,
		Type:           jType,
		ResourceID:     resourceID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Status:         StatusQueued,
		Progress:       "QUEUED",
		MaxRetries:     3,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	jm.jobs[jobID] = j
	if idempotencyKey != "" {
		jm.idempotency[idempotencyKey] = jobID
	}
	return j, nil
}

func (jm *JobManager) GetJob(id string) (*Job, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	j, ok := jm.jobs[id]
	if !ok {
		return nil, fmt.Errorf("job not found")
	}
	return j, nil
}

func (jm *JobManager) ListJobs(orgID, projectID string) []*Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	var result []*Job
	for _, j := range jm.jobs {
		if orgID != "" && j.OrganizationID != orgID {
			continue
		}
		if projectID != "" && j.ProjectID != projectID {
			continue
		}
		result = append(result, j)
	}
	return result
}
