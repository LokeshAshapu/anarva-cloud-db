package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

type rateLimiter struct {
	tokens     int
	maxTokens  int
	lastRefill time.Time
	mu         sync.Mutex
}

type RateLimitMiddleware struct {
	limiters map[string]*rateLimiter
	mu       sync.RWMutex
	rate     int // tokens per minute
	eventSvc *security.SecurityEventService
}

func NewRateLimitMiddleware(ratePerMinute int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiters: make(map[string]*rateLimiter),
		rate:     ratePerMinute,
	}
}

func (m *RateLimitMiddleware) SetEventService(eventSvc *security.SecurityEventService) {
	m.eventSvc = eventSvc
}

func (m *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff
		}

		m.mu.Lock()
		limiter, ok := m.limiters[ip]
		if !ok {
			limiter = &rateLimiter{
				tokens:     m.rate,
				maxTokens:  m.rate,
				lastRefill: time.Now(),
			}
			m.limiters[ip] = limiter
		}
		m.mu.Unlock()

		limiter.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(limiter.lastRefill)
		if elapsed >= time.Minute {
			limiter.tokens = limiter.maxTokens
			limiter.lastRefill = now
		}

		if limiter.tokens <= 0 {
			limiter.mu.Unlock()

			if m.eventSvc != nil {
				reqID := r.Header.Get("X-Request-ID")
				m.eventSvc.RecordEvent(
					security.EventRateLimitViolation,
					security.SeverityMedium,
					"DENIED",
					ip,
					"org-default",
					"",
					"",
					reqID,
					"Rate limit threshold exceeded on "+r.URL.Path,
				)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code":    "TOO_MANY_REQUESTS",
					"message": "Rate limit exceeded. Please retry after 60 seconds.",
				},
			})
			return
		}

		limiter.tokens--
		limiter.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
