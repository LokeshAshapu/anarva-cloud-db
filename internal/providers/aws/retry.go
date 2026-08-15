package aws

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Timeout    time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   3 * time.Second,
		Timeout:    10 * time.Second,
	}
}

func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Retryable status codes & transient error messages
	if strings.Contains(msg, "429") || strings.Contains(msg, "Throttling") || strings.Contains(msg, "RequestLimitExceeded") {
		return true
	}
	if strings.Contains(msg, "502") || strings.Contains(msg, "503") || strings.Contains(msg, "504") || strings.Contains(msg, "ServiceUnavailable") {
		return true
	}
	if strings.Contains(msg, "connection reset") || strings.Contains(msg, "i/o timeout") || strings.Contains(msg, "temporary network error") {
		return true
	}
	return false
}

func ExecuteWithRetry[T any](ctx context.Context, cfg RetryConfig, opName string, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	var lastErr error

	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.BaseDelay == 0 {
		cfg.BaseDelay = 100 * time.Millisecond
	}

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		opCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		res, err := fn(opCtx)
		cancel()

		if err == nil {
			return res, nil
		}

		lastErr = MapAWSError(err)
		if !IsRetryableError(err) || attempt == cfg.MaxRetries {
			return zero, lastErr
		}

		// Calculate Exponential Backoff with Jitter
		backoff := cfg.BaseDelay * time.Duration(1<<attempt)
		if backoff > cfg.MaxDelay {
			backoff = cfg.MaxDelay
		}

		// Apply randomized full jitter (0.5x to 1.5x)
		jitterMultiplier := 0.5 + rand.Float64()
		jitterDelay := time.Duration(float64(backoff) * jitterMultiplier)

		select {
		case <-ctx.Done():
			return zero, fmt.Errorf("PROVIDER_TIMEOUT: Operation %s timed out: %w", opName, ctx.Err())
		case <-time.After(jitterDelay):
		}
	}

	return zero, fmt.Errorf("PROVIDER_OPERATION_FAILED: Operation %s failed after %d retries: %w", opName, cfg.MaxRetries, lastErr)
}
