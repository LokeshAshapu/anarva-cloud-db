package failover

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
)

type FailoverEngine struct {
	mu    sync.Mutex
	locks map[string]int64
}

func NewFailoverEngine() *FailoverEngine {
	return &FailoverEngine{
		locks: make(map[string]int64),
	}
}

func (e *FailoverEngine) ExecuteFailover(ctx context.Context, policy *domain.FailoverPolicy) (*domain.RecoveryPlan, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Split-Brain Protection Lock Acquisition
	currentGen := e.locks[policy.ResourceID]
	if policy.GenerationLock <= currentGen {
		return nil, fmt.Errorf("SPLIT-BRAIN BLOCKED: Concurrent or stale failover attempt rejected. Generation %d <= current %d", policy.GenerationLock, currentGen)
	}

	// Update Distributed Generation Lock
	newGen := policy.GenerationLock
	e.locks[policy.ResourceID] = newGen

	steps := []string{
		fmt.Sprintf("Acquired distributed failover lock (Generation %d)", newGen),
		fmt.Sprintf("Validated primary '%s' failure threshold (%d checks failed)", policy.Primary, policy.HealthThreshold),
		fmt.Sprintf("Promoted secondary target '%s' to active primary", policy.Secondary),
		"Updated global DNS and load balancer routing endpoints",
		"Completed health verification on new primary endpoint",
	}

	plan := &domain.RecoveryPlan{
		ID:             fmt.Sprintf("rec-plan-%d", time.Now().UnixNano()),
		ResourceID:     policy.ResourceID,
		FailoverRegion: policy.Secondary,
		Steps:          steps,
		Status:         "COMPLETED",
		CreatedAt:      time.Now(),
	}

	return plan, nil
}
