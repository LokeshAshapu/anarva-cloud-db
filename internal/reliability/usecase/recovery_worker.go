package usecase

import (
	"context"
	"log"
	"sync"
	"time"
)

type RecoveryWorkerConfig struct {
	Interval time.Duration
	BatchSize int
}

func DefaultRecoveryWorkerConfig() RecoveryWorkerConfig {
	return RecoveryWorkerConfig{
		Interval: 5 * time.Second,
		BatchSize: 50,
	}
}

type RecoveryWorker struct {
	uc     *ReliabilityUseCase
	config RecoveryWorkerConfig
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
	active bool
}

func NewRecoveryWorker(uc *ReliabilityUseCase, cfg RecoveryWorkerConfig) *RecoveryWorker {
	if cfg.Interval == 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	return &RecoveryWorker{
		uc:     uc,
		config: cfg,
	}
}

func (w *RecoveryWorker) Start(parentCtx context.Context) {
	w.mu.Lock()
	if w.active {
		w.mu.Unlock()
		return
	}
	w.active = true

	ctx, cancel := context.WithCancel(parentCtx)
	w.cancel = cancel
	w.wg.Add(1)
	w.mu.Unlock()

	go w.runLoop(ctx)
}

func (w *RecoveryWorker) Stop() {
	w.mu.Lock()
	if !w.active {
		w.mu.Unlock()
		return
	}
	w.active = false
	if w.cancel != nil {
		w.cancel()
	}
	w.mu.Unlock()

	w.wg.Wait()
}

func (w *RecoveryWorker) runLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	log.Printf("[RECOVERY_WORKER] Started background operation recovery daemon (Interval: %v)", w.config.Interval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[RECOVERY_WORKER] Stopped background operation recovery daemon gracefully")
			return
		case <-ticker.C:
			w.performRecovery(ctx)
		}
	}
}

func (w *RecoveryWorker) performRecovery(ctx context.Context) {
	reconciled := w.uc.ReconcileInterruptedOperations(ctx)
	if reconciled > 0 {
		log.Printf("[RECOVERY_WORKER] Successfully reconciled %d interrupted/stale control-plane operations", reconciled)
	}

	timedOut := w.uc.DetectOperationTimeouts(ctx, 5*time.Minute)
	if timedOut > 0 {
		log.Printf("[RECOVERY_WORKER] Marked %d stale running operations as TIMED_OUT", timedOut)
	}
}
